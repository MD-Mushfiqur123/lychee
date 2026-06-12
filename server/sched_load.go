package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	"github.com/lychee/lychee/fs/ggml"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/ml"
	"github.com/lychee/lychee/types/model"
	"github.com/lychee/lychee/x/imagegen"
	"github.com/lychee/lychee/x/mlxrunner"
)

// load creates a new model based on req and loads it. If requireFull is true then the model must be loaded fully onto GPUs
// (if any). Returns whether the scheduler needs to evict a model to make this one fit.
func (s *Scheduler) load(req *LlmRequest, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, requireFull bool) bool {
	numParallel := max(int(envconfig.NumParallel()), 1)
	completion := req.model.CheckCapabilities(model.CapabilityCompletion) == nil

	// Embedding models should always be loaded with parallel=1
	if !completion {
		numParallel = 1
	}

	// Some architectures are not safe with num_parallel > 1.
	// ref: https://github.com/lychee/lychee/issues/4165
	if slices.Contains(model.SingleParallelArchitectures, req.model.Config.ModelFamily) && numParallel != 1 {
		numParallel = 1
		slog.Warn("model architecture does not currently support parallel requests", "architecture", req.model.Config.ModelFamily)
	}

	sessionDuration := envconfig.KeepAlive()
	if req.sessionDuration != nil {
		sessionDuration = req.sessionDuration.Duration
	}

	s.loadedMu.Lock()
	llama := s.activeLoading
	var f *ggml.GGML
	loadGpus := gpus
	var launchOpts api.Options
	var err error

	if llama == nil {
		if !req.model.IsMLX() {
			var loadErr error
			f, loadErr = llm.LoadModel(req.model.ModelPath, 1024)
			if loadErr != nil {
				slog.Info("failed to load model metadata", "model", req.model.ModelPath, "error", loadErr)
				req.errCh <- loadErr
				s.loadedMu.Unlock()
				return false
			}

			predictedCtx := effectiveLlamaServerContext(req.opts.NumCtx, f, numParallel)
			predicted := llm.PredictServerVRAM(req.model.ModelPath, f, predictedCtx)
			loadGpus, launchOpts = selectLlamaServerPlacement(systemInfo, gpus, predicted, req.opts)
			availableForBatch, _, _ := availableMemoryForPlacement(systemInfo, loadGpus, launchOpts)
			flashAttention := llm.LlamaServerFlashAttention(loadGpus)
			req.applyAutomaticGenerationBatch(completion, predictedCtx, predicted, availableForBatch, flashAttention, loadGpus)
			launchOpts.NumBatch = req.opts.NumBatch
			predictedForLoad := predicted + generationBatchSurchargeForCompletion(completion, launchOpts.NumBatch)

			// Pre-flight check: estimate whether the model fits in remaining memory.
			// llama-server auto-detects layers based on available VRAM, so if
			// we predict it won't fit, evict before spawning.
			if requireFull && !explicitPartialGPUOffload(launchOpts, f) && len(s.loaded) > 0 && len(loadGpus) > 0 {
				freeMemory, gpuFreeMemory, systemLimited := availableMemoryForPlacement(systemInfo, loadGpus, launchOpts)
				// Use 80% of free memory as threshold to leave headroom.
				if predictedForLoad > freeMemory*80/100 {
					slog.Info("llama-server model predicted to exceed available memory, evicting",
						"predicted", format.HumanBytes2(predictedForLoad),
						"predicted_num_ctx", predictedCtx,
						"num_batch", launchOpts.NumBatch,
						"available", format.HumanBytes2(freeMemory),
						"gpu_free", format.HumanBytes2(gpuFreeMemory),
						"system_free", format.HumanBytes2(systemInfo.FreeMemory),
						"system_limited", systemLimited)
					s.loadedMu.Unlock()
					return true
				}
				slog.Info("llama-server model fits alongside existing models",
					"predicted", format.HumanBytes2(predictedForLoad),
					"predicted_num_ctx", predictedCtx,
					"num_batch", launchOpts.NumBatch,
					"available", format.HumanBytes2(freeMemory),
					"gpu_free", format.HumanBytes2(gpuFreeMemory),
					"system_free", format.HumanBytes2(systemInfo.FreeMemory),
					"system_limited", systemLimited)
			}

			launchOpts = s.applyLlamaServerMmapDefaults(req, launchOpts, systemInfo, loadGpus, f, numParallel)
			req.contextShift = resolveContextShift(req.shift, effectiveModelContext(launchOpts.NumCtx, f))

			config := llamaServerConfigForModel(req.model)
			config.ContextShift = req.contextShift
			config.CacheReuse = req.contextShift || (req.shift != nil && *req.shift)
			if launchOpts.DraftModel != "" {
				if dm, err := GetModel(launchOpts.DraftModel); err == nil && dm.ModelPath != "" {
					config.DraftModelPath = dm.ModelPath
				} else {
					config.DraftModelPath = launchOpts.DraftModel
				}
			}
			llama, err = s.newServerFn(systemInfo, loadGpus, req.model.ModelPath, f, req.model.AdapterPaths, req.model.ProjectorPaths, launchOpts, numParallel, config)
			if err != nil {
				// some older models are not compatible with newer versions of llama.cpp
				// show a generalized compatibility error until there is a better way to
				// check for model compatibility
				if errors.Is(err, ggml.ErrUnsupportedFormat) || strings.Contains(err.Error(), "failed to load model") {
					err = fmt.Errorf("%v: this model may be incompatible with your version of Lychee. If you previously pulled this model, try updating it by running `lychee pull %s`", err, req.model.ShortName)
				}
			}
		} else {
			modelName := req.model.ShortName
			if slices.Contains(req.model.Config.Capabilities, "image") {
				llama, err = imagegen.NewServer(modelName)
			} else {
				llama, err = mlxrunner.NewClient(modelName)
			}
		}
		if err != nil {
			slog.Info("failed to create server", "model", req.model.ShortName, "error", err)
			req.errCh <- err
			s.loadedMu.Unlock()
			return false
		}

		s.activeLoading = llama
	} else {
		wantPath := req.model.ModelPath
		if wantPath == "" {
			wantPath = req.model.ShortName
		}
		if s.activeLoading.ModelPath() != wantPath {
			panic(fmt.Errorf("attempting to load different model after eviction (original %v new %v)", s.activeLoading.ModelPath(), wantPath))
		}
	}

	s.loadedMu.Unlock()

	systemTotalMemory := systemInfo.TotalMemory
	systemFreeMemory := systemInfo.FreeMemory
	systemSwapFreeMemory := systemInfo.FreeSwap
	slog.Info("system memory", "total", format.HumanBytes2(systemTotalMemory), "free", format.HumanBytes2(systemFreeMemory), "free_swap", format.HumanBytes2(systemSwapFreeMemory))

	for _, gpu := range loadGpus {
		available := gpu.FreeMemory - envconfig.GpuOverhead() - gpu.MinimumMemory()
		if gpu.FreeMemory < envconfig.GpuOverhead()+gpu.MinimumMemory() {
			available = 0
		}
		slog.Info("gpu memory", "id", gpu.ID, "library", gpu.Library,
			"available", format.HumanBytes2(available),
			"free", format.HumanBytes2(gpu.FreeMemory),
			"minimum", format.HumanBytes2(gpu.MinimumMemory()),
			"overhead", format.HumanBytes2(envconfig.GpuOverhead()))
	}

	gpuIDs, err := llama.Load(req.ctx, systemInfo, loadGpus, requireFull)
	if err != nil {
		if errors.Is(err, llm.ErrLoadRequiredFull) {
			if !requireFull {
				// No other models loaded, yet we still don't fit, so report an error
				slog.Info("model is too large for system memory", "requireFull", requireFull)
				s.activeLoading.Close()
				s.activeLoading = nil
				req.errCh <- err
				return false
			}
			return true
		}

		slog.Info("Load failed", "model", req.model.ModelPath, "error", err)
		s.activeLoading.Close()
		s.activeLoading = nil

		s.loadedMu.Lock()
		loadedCount := len(s.loaded)
		s.loadedMu.Unlock()
		otherLoaded := loadedCount > 0
		if !req.oomRetryAttempted && llm.IsOutOfMemory(err) {
			if oldNumCtx, effectiveNumCtx, newNumCtx, oldNumBatch, newNumBatch, ok := req.reduceAutoNumCtxForLoadOOM(f, numParallel, completion, systemInfo, loadGpus, launchOpts); ok {
				req.oomRetryAttempted = true
				slog.Warn("llama-server load failed; reducing automatic context and retrying once",
					"model", req.model.ModelPath,
					"old_num_ctx", oldNumCtx,
					"effective_num_ctx", effectiveNumCtx,
					"new_num_ctx", newNumCtx,
					"old_num_batch", oldNumBatch,
					"new_num_batch", newNumBatch,
					"loaded_count", loadedCount,
					"evict_all", otherLoaded,
					"error", err)
				return true
			}
		}
		if otherLoaded && !req.oomRetryAttempted && llm.IsOutOfMemory(err) {
			req.oomRetryAttempted = true
			slog.Warn("llama-server load failed; evicting all other models and retrying once", "model", req.model.ModelPath, "error", err)
			return true
		}

		req.errCh <- err
		return false
	}
	logTemplateSelection(req.model)

	// Determine if we have discrete GPUs which we should monitor VRAM usage on during shutdown
	discreteGPUs := false
iGPUScan:
	for _, devid := range gpuIDs {
		for _, dev := range loadGpus {
			if dev.DeviceID == devid {
				if !dev.Integrated {
					discreteGPUs = true
					break iGPUScan
				}
			}
		}
	}

	totalSize, vramSize := llama.MemorySize()
	trainContext := modelTrainContext(f)
	if effectiveNumCtx := llama.ContextLength(); req.model.ModelPath != "" && effectiveNumCtx > 0 {
		req.opts.NumCtx = effectiveNumCtx
		req.contextShift = resolveContextShift(req.shift, effectiveNumCtx)
	}
	runner := &runnerRef{
		model:           req.model,
		modelPath:       req.model.ModelPath,
		modelKey:        schedulerModelKey(req.model),
		llama:           llama,
		Options:         &req.opts,
		sessionDuration: sessionDuration,
		gpus:            gpuIDs,
		discreteGPUs:    discreteGPUs,
		isImagegen:      slices.Contains(req.model.Config.Capabilities, "image"),
		totalSize:       totalSize,
		vramSize:        vramSize,
		loading:         true,
		pid:             llama.Pid(),
		numCtxAuto:      req.numCtxAuto,
		numBatchAuto:    req.numBatchAuto,
		useMMapAuto:     req.useMMapAuto,
		contextShift:    req.contextShift,
		trainContext:    trainContext,
	}
	runner.numParallel = numParallel
	runner.refMu.Lock() // hold lock until running or aborted

	s.loadedMu.Lock()
	if oldRunner, ok := s.loaded[runner.modelKey]; ok {
		// Shouldn't happen, but safeguard against leaking a runner
		slog.Warn("model was still loaded", "old_runner", oldRunner, "new_runner", runner)
		oldRunner.refMu.Lock()
		oldRunner.unload()
		oldRunner.refMu.Unlock()
	}
	s.activeLoading = nil
	s.loaded[runner.modelKey] = runner
	slog.Info("loaded runners", "count", len(s.loaded))
	s.loadedMu.Unlock()

	go func() {
		defer runner.refMu.Unlock()
		if err = llama.WaitUntilRunning(req.ctx); err != nil {
			slog.Error("error loading llama server", "error", err)
			req.errCh <- err
			slog.Debug("triggering expiration for failed load", "runner", runner)
			s.expiredCh <- runner
			return
		}
		slog.Debug("finished setting up", "runner", runner)
		if runner.pid < 0 {
			runner.pid = llama.Pid()
		}
		runner.refCount++
		runner.loading = false

		// Supervise runner crash recovery
		go func(r *runnerRef) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			watchCtx := s.ctx
			if watchCtx == nil {
				watchCtx = context.Background()
			}
			for {
				select {
				case <-watchCtx.Done():
					return
				case <-ticker.C:
					r.refMu.Lock()
					s.loadedMu.Lock()
					current, exists := s.loaded[r.modelKey]
					s.loadedMu.Unlock()
					if !exists || current != r {
						r.refMu.Unlock()
						return
					}
					if r.HasExited() {
						slog.Error("supervised llama-server process exited unexpectedly", "model", r.modelKey, "pid", r.pid)
						r.refCount = 0
						slog.Warn("supervised watchdog triggering eviction cleanup for crashed model", "model", r.modelKey)
						s.expiredCh <- r
						r.refMu.Unlock()
						return
					}
					r.refMu.Unlock()
				}
			}
		}(runner)

		go func() {
			<-req.ctx.Done()
			slog.Debug("context for request finished")
			s.finishedReqCh <- req
		}()
		req.successCh <- runner
	}()

	return false
}

func (req *LlmRequest) reduceAutoNumCtxForLoadOOM(f *ggml.GGML, numParallel int, completion bool, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, launchOpts api.Options) (oldNumCtx, effectiveNumCtx, newNumCtx, oldNumBatch, newNumBatch int, ok bool) {
	if !req.numCtxAuto {
		return 0, 0, 0, 0, 0, false
	}

	oldNumCtx = req.opts.NumCtx
	oldNumBatch = req.opts.NumBatch
	effectiveNumCtx = oldNumCtx
	if f != nil {
		if trainCtx := int(f.KV().ContextLength()); trainCtx > 0 && effectiveNumCtx > trainCtx {
			effectiveNumCtx = trainCtx
		}
	}

	newNumCtx, ok = nextLowerAutoNumCtx(effectiveNumCtx)
	if !ok || newNumCtx >= oldNumCtx {
		return 0, 0, 0, 0, 0, false
	}

	req.opts.NumCtx = newNumCtx
	predictedCtx := effectiveLlamaServerContext(req.opts.NumCtx, f, numParallel)
	predictedVRAM := llm.PredictServerVRAM(req.model.ModelPath, f, predictedCtx)
	available, _, _ := availableMemoryForPlacement(systemInfo, gpus, launchOpts)
	req.applyAutomaticGenerationBatch(completion, predictedCtx, predictedVRAM, available, llm.LlamaServerFlashAttention(gpus), gpus)
	newNumBatch = req.opts.NumBatch
	return oldNumCtx, effectiveNumCtx, newNumCtx, oldNumBatch, newNumBatch, true
}

func explicitPartialGPUOffload(opts api.Options, f *ggml.GGML) bool {
	if opts.NumGPU <= 0 || f == nil {
		return false
	}

	return uint64(opts.NumGPU) < f.KV().BlockCount()+1
}

func effectiveLlamaServerContext(numCtx int, f *ggml.GGML, numParallel int) int {
	return effectiveModelContext(numCtx, f) * max(numParallel, 1)
}

const (
	llamaServerGenerationBatchDefault     = 512
	llamaServerGenerationBatchConstrained = 256
	llamaServerGenerationBatchMedium      = 1024
	llamaServerGenerationBatchLarge       = 2048

	llamaServerGenerationBatchMediumHeadroomPercent = 75
	llamaServerGenerationBatchLargeHeadroomPercent  = 60
)

func (req *LlmRequest) applyAutomaticGenerationBatch(completion bool, effectiveCtx int, predictedVRAM, availableMemory uint64, flashAttention ml.FlashAttentionType, gpus []ml.DeviceInfo) {
	if !completion || !req.numBatchAuto {
		return
	}

	req.opts.NumBatch = automaticGenerationBatch(effectiveCtx, predictedVRAM, availableMemory, flashAttention, gpus)
}

func generationBatchSurchargeForCompletion(completion bool, batch int) uint64 {
	if !completion {
		return 0
	}
	return generationBatchSurcharge(batch)
}

func automaticGenerationBatch(effectiveCtx int, predictedVRAM, availableMemory uint64, flashAttention ml.FlashAttentionType, gpus []ml.DeviceInfo) int {
	if flashAttention == ml.FlashAttentionDisabled && hasCUDADevice(gpus) {
		if constrainedCUDAWithoutFlashAttention(effectiveCtx, gpus) {
			return llamaServerGenerationBatchConstrained
		}
		return llamaServerGenerationBatchDefault
	}

	batch := generationBatchForContext(effectiveCtx)
	for batch > llamaServerGenerationBatchDefault && !generationBatchFits(batch, predictedVRAM, availableMemory) {
		batch = nextLowerGenerationBatch(batch)
	}
	return batch
}

func hasCUDADevice(gpus []ml.DeviceInfo) bool {
	return slices.ContainsFunc(gpus, func(gpu ml.DeviceInfo) bool {
		return gpu.Library == "CUDA"
	})
}

func constrainedCUDAWithoutFlashAttention(effectiveCtx int, gpus []ml.DeviceInfo) bool {
	if effectiveCtx <= 4096 {
		return false
	}
	return slices.ContainsFunc(gpus, func(gpu ml.DeviceInfo) bool {
		if gpu.Library != "CUDA" {
			return false
		}
		memory := gpu.FreeMemory
		if memory == 0 || (gpu.TotalMemory > 0 && gpu.TotalMemory < memory) {
			memory = gpu.TotalMemory
		}
		return memory > 0 && memory <= 8*format.GibiByte
	})
}

func generationBatchForContext(effectiveCtx int) int {
	switch {
	case effectiveCtx > 32768:
		return llamaServerGenerationBatchLarge
	case effectiveCtx > 4096:
		return llamaServerGenerationBatchMedium
	default:
		return llamaServerGenerationBatchDefault
	}
}

func generationBatchFits(batch int, predictedVRAM, availableMemory uint64) bool {
	if predictedVRAM == 0 || availableMemory == 0 {
		return true
	}

	threshold := availableMemory * 80 / 100
	if predictedVRAM > threshold {
		return false
	}
	if !generationBatchHasHeadroom(batch, predictedVRAM, availableMemory) {
		return false
	}

	return generationBatchSurcharge(batch) <= threshold-predictedVRAM
}

func generationBatchHasHeadroom(batch int, predictedVRAM, availableMemory uint64) bool {
	switch {
	case batch >= llamaServerGenerationBatchLarge:
		return predictedVRAM <= availableMemory*llamaServerGenerationBatchLargeHeadroomPercent/100
	case batch >= llamaServerGenerationBatchMedium:
		return predictedVRAM <= availableMemory*llamaServerGenerationBatchMediumHeadroomPercent/100
	default:
		return true
	}
}

func nextLowerGenerationBatch(batch int) int {
	switch {
	case batch > llamaServerGenerationBatchMedium:
		return llamaServerGenerationBatchMedium
	default:
		return llamaServerGenerationBatchDefault
	}
}

func generationBatchSurcharge(batch int) uint64 {
	switch {
	case batch >= llamaServerGenerationBatchLarge:
		return 2 * format.GibiByte
	case batch >= llamaServerGenerationBatchMedium:
		return 768 * format.MebiByte
	default:
		return 0
	}
}

func nextLowerAutoNumCtx(numCtx int) (int, bool) {
	switch {
	case numCtx > 32768:
		return 32768, true
	case numCtx > 4096:
		return 4096, true
	default:
		return 0, false
	}
}

func (s *Scheduler) applyLlamaServerMmapDefaults(req *LlmRequest, launchOpts api.Options, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, f *ggml.GGML, numParallel int) api.Options {
	predictedCtx := effectiveLlamaServerContext(req.opts.NumCtx, f, numParallel)
	predictedVRAM := llm.PredictServerVRAM(req.model.ModelPath, f, predictedCtx)
	availableVRAM, _, _ := availableMemoryForPlacement(systemInfo, gpus, launchOpts)

	if reason := disableMmapDefaultReason(runtime.GOOS, req.opts, gpus, f.KV().BlockCount(), predictedVRAM, availableVRAM); reason != "" {
		useMmap := false
		req.opts.UseMMap = &useMmap
		req.useMMapAuto = true
		slog.Info("disabling mmap for llama-server load by default",
			"model", req.model.ModelPath,
			"reason", reason)
	} else {
		s.maybeDisableMmapForHostPressure(req, launchOpts, systemInfo, gpus, f, numParallel)
	}

	launchOpts.UseMMap = req.opts.UseMMap
	return launchOpts
}

func disableMmapDefaultReason(goos string, opts api.Options, gpus []ml.DeviceInfo, blockCount, predictedVRAM, availableVRAM uint64) string {
	if opts.UseMMap != nil {
		return ""
	}
	if opts.NumGPU == 0 || len(gpus) == 0 || allDevicesLibrary(gpus, "cpu") {
		return "cpu"
	}
	if goos == "windows" && hasDeviceLibrary(gpus, "cuda") {
		return "windows_cuda"
	}
	if hasDeviceLibrary(gpus, "metal") {
		if opts.NumGPU > 0 && blockCount > 0 && uint64(opts.NumGPU) < blockCount+1 {
			return "metal_partial_offload"
		}
		if opts.NumGPU < 0 && predictedVRAM > 0 && availableVRAM > 0 && predictedVRAM > availableVRAM {
			return "metal_partial_offload"
		}
	}
	return ""
}

func (s *Scheduler) maybeDisableMmapForHostPressure(req *LlmRequest, launchOpts api.Options, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, f *ggml.GGML, numParallel int) {
	modelSize := modelFileSize(req.model.ModelPath)
	loadedMmapSize := s.loadedMmapModelSizeLocked()
	predictedCtx := effectiveLlamaServerContext(req.opts.NumCtx, f, numParallel)
	predictedVRAM := llm.PredictServerVRAM(req.model.ModelPath, f, predictedCtx)
	availableVRAM, _, _ := availableMemoryForPlacement(systemInfo, gpus, launchOpts)
	placementGpus := gpusForPlacement(gpus, launchOpts)

	if !disableMmapForHostPressure(runtime.GOOS, req.opts, systemInfo, placementGpus, modelSize, loadedMmapSize, predictedVRAM, availableVRAM) {
		return
	}

	useMmap := false
	req.opts.UseMMap = &useMmap
	req.useMMapAuto = true
	slog.Info("disabling mmap for llama-server load due to host memory pressure",
		"model", req.model.ModelPath,
		"model_size", format.HumanBytes2(modelSize),
		"loaded_mmap_size", format.HumanBytes2(loadedMmapSize),
		"headroom", format.HumanBytes2(mmapHostPressureHeadroom(systemInfo.TotalMemory)),
		"system_free", format.HumanBytes2(systemInfo.FreeMemory),
		"system_total", format.HumanBytes2(systemInfo.TotalMemory),
		"predicted_vram", format.HumanBytes2(predictedVRAM),
		"available_vram", format.HumanBytes2(availableVRAM),
	)
}

func disableMmapForHostPressure(goos string, opts api.Options, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, modelSize, loadedMmapSize, predictedVRAM, availableVRAM uint64) bool {
	if opts.UseMMap != nil || goos != "linux" || modelSize == 0 || systemInfo.FreeMemory == 0 || !allDiscreteGPUs(gpus) {
		return false
	}

	// Only back off mmap when we still expect the model to fit on discrete GPU.
	// If VRAM is already tight, disabling mmap can make partial CPU offload
	// worse by turning file-backed mappings into anonymous memory.
	if predictedVRAM == 0 || availableVRAM == 0 || predictedVRAM > availableVRAM*80/100 {
		return false
	}

	pressure := modelSize + loadedMmapSize + mmapHostPressureHeadroom(systemInfo.TotalMemory)
	return systemInfo.FreeMemory < pressure
}

func mmapHostPressureHeadroom(totalMemory uint64) uint64 {
	if totalMemory == 0 {
		return 8 * format.GigaByte
	}
	return max(8*format.GigaByte, totalMemory/10)
}

func modelFileSize(path string) uint64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint64(info.Size())
}

func (s *Scheduler) loadedMmapModelSizeLocked() uint64 {
	var total uint64
	for _, r := range s.loaded {
		if !runnerUsesMmap(r) {
			continue
		}
		if size := modelFileSize(r.modelPath); size > 0 {
			total += size
		} else {
			total += r.totalSize
		}
	}
	return total
}

func runnerUsesMmap(r *runnerRef) bool {
	if r == nil || r.Options == nil || r.Options.UseMMap == nil {
		return true
	}
	return *r.Options.UseMMap
}
