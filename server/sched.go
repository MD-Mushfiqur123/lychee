package server

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/discover"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	"github.com/lychee/lychee/fs/ggml"
	"github.com/lychee/lychee/llm"
	"github.com/lychee/lychee/logutil"
	"github.com/lychee/lychee/ml"
	"github.com/lychee/lychee/types/model"
	"github.com/lychee/lychee/x/mlxrunner"
)

type LlmRequest struct {
	ctx             context.Context //nolint:containedctx
	model           *Model
	opts            api.Options
	sessionDuration *api.Duration
	successCh       chan *runnerRef
	errCh           chan error
	schedAttempts   uint

	// oomRetryAttempted is set after a llama-server load crash triggers an
	// evict-all-and-retry. Prevents infinite retry on persistent load failures.
	oomRetryAttempted bool

	// numCtxAuto is true when NumCtx came from Lychee's automatic VRAM-tier
	// default rather than explicit request, model, or environment config.
	numCtxAuto bool

	// numBatchAuto is true when NumBatch came from Lychee's default options
	// rather than an explicit request or model option.
	numBatchAuto bool

	// useMMapAuto is true when UseMMap was derived by the scheduler rather than
	// explicitly requested.
	useMMapAuto bool

	// contextShift is a llama-server launch attribute resolved from the
	// request-level shift option before scheduling.
	contextShift bool
	shift        *bool
}

type Scheduler struct {
	pendingReqCh  chan *LlmRequest
	finishedReqCh chan *LlmRequest
	expiredCh     chan *runnerRef
	unloadedCh    chan any

	// loadedMu protects loaded and activeLoading
	loadedMu sync.Mutex

	// activeLoading is the model that we are currently working on loading,
	// including by evicting one or more other models. We can only load
	// one model at a time but new requests to models that already loaded can
	// happen in parallel
	activeLoading llm.LlamaServer
	loaded        map[string]*runnerRef

	loadFn          func(req *LlmRequest, systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, requireFull bool) bool
	newServerFn     func(systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, model string, f *ggml.GGML, adapters []string, projectors []string, opts api.Options, numParallel int, config llm.LlamaServerConfig) (llm.LlamaServer, error)
	getGpuFn        func(ctx context.Context, runners []ml.FilteredRunnerDiscovery) []ml.DeviceInfo
	getSystemInfoFn func() ml.SystemInfo
	waitForRecovery time.Duration
	ctx             context.Context
}

// Default automatic value for number of models we allow per GPU
// Model will still need to fit in VRAM, but loading many small models
// on a large GPU can cause stalling
var defaultModelsPerGPU = 3

var ErrMaxQueue = errors.New("server busy, please try again.  maximum pending requests exceeded")

func InitScheduler(ctx context.Context) *Scheduler {
	maxQueue := envconfig.MaxQueue()
	sched := &Scheduler{
		pendingReqCh:    make(chan *LlmRequest, maxQueue),
		finishedReqCh:   make(chan *LlmRequest, maxQueue),
		expiredCh:       make(chan *runnerRef, maxQueue),
		unloadedCh:      make(chan any, maxQueue),
		loaded:          make(map[string]*runnerRef),
		newServerFn:     llm.NewLlamaServer,
		getGpuFn:        discover.GPUDevices,
		getSystemInfoFn: discover.GetSystemInfo,
		waitForRecovery: 5 * time.Second,
	}
	sched.loadFn = sched.load
	return sched
}

// schedulerModelKey returns the scheduler map key for a model.
// GGUF-backed models use ModelPath; safetensors/image models without a
// ModelPath use manifest digest so distinct models don't collide.
func schedulerModelKey(m *Model) string {
	if m == nil {
		return ""
	}
	if m.ModelPath != "" {
		return m.ModelPath
	}
	if m.Digest != "" {
		return "digest:" + m.Digest
	}
	if m.Name != "" {
		return "name:" + m.Name
	}
	if m.ShortName != "" {
		return "short:" + m.ShortName
	}
	return ""
}

// context must be canceled to decrement ref count and release the runner
func (s *Scheduler) GetRunner(c context.Context, m *Model, opts api.Options, sessionDuration *api.Duration) (chan *runnerRef, chan error) {
	return s.getRunner(c, m, opts, sessionDuration, false, false, nil)
}

const contextShiftSmallContextLimit = 8192

func resolveContextShift(shift *bool, numCtx int) bool {
	if shift != nil {
		return *shift
	}

	return numCtx > 0 && numCtx < contextShiftSmallContextLimit
}

func effectiveModelContext(numCtx int, f *ggml.GGML) int {
	return effectiveContext(numCtx, modelTrainContext(f))
}

func modelTrainContext(f *ggml.GGML) int {
	if f == nil {
		return 0
	}

	return int(f.KV().ContextLength())
}

func effectiveContext(numCtx, trainCtx int) int {
	if trainCtx > 0 && numCtx > trainCtx {
		return trainCtx
	}

	return numCtx
}

func (s *Scheduler) getRunner(c context.Context, m *Model, opts api.Options, sessionDuration *api.Duration, numCtxAuto bool, numBatchAuto bool, shift *bool) (chan *runnerRef, chan error) {
	if opts.NumCtx < 4 {
		opts.NumCtx = 4
	}

	if m.CheckCapabilities(model.CapabilityVision) == nil {
		// multimodal models require at least 2048 context
		opts.NumCtx = max(opts.NumCtx, 2048)
	}

	contextShift := false
	if m.ModelPath != "" {
		contextShift = resolveContextShift(shift, opts.NumCtx)
	}

	req := &LlmRequest{
		ctx:             c,
		model:           m,
		opts:            opts,
		sessionDuration: sessionDuration,
		successCh:       make(chan *runnerRef, 1),
		errCh:           make(chan error, 1),
		numCtxAuto:      numCtxAuto,
		numBatchAuto:    numBatchAuto,
		contextShift:    contextShift,
		shift:           shift,
	}

	key := schedulerModelKey(req.model)
	s.loadedMu.Lock()
	runner := s.loaded[key]
	s.loadedMu.Unlock()
	if runner != nil && !runner.needsReload(c, req) {
		req.useLoadedRunner(runner, s.finishedReqCh)
	} else {
		select {
		case s.pendingReqCh <- req:
		default:
			req.errCh <- ErrMaxQueue
		}
	}
	return req.successCh, req.errCh
}

// Returns immediately, spawns go routines for the scheduler which will shutdown when ctx is done
func (s *Scheduler) Run(ctx context.Context) {
	slog.Debug("starting llm scheduler")
	s.ctx = ctx
	go func() {
		s.processPending(ctx)
	}()

	go func() {
		s.processCompleted(ctx)
	}()
}

func (s *Scheduler) processPending(ctx context.Context) {
	maxRunners := envconfig.MaxRunners()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("shutting down scheduler pending loop")
			return
		case pending := <-s.pendingReqCh:
			// Block other requests until we get this pending request running
			pending.schedAttempts++

			if pending.ctx.Err() != nil {
				slog.Debug("pending request cancelled or timed out, skipping scheduling")
				continue
			}
			logutil.Trace("processing incoming request", "model", pending.model.ModelPath)

			for {
				var runnerToExpire *runnerRef
				pendingKey := schedulerModelKey(pending.model)
				s.loadedMu.Lock()
				runner := s.loaded[pendingKey]
				loadedCount := len(s.loaded)
				runnersSnapshot := make([]ml.FilteredRunnerDiscovery, 0, len(s.loaded))
				for _, r := range s.loaded {
					runnersSnapshot = append(runnersSnapshot, r)
				}
				s.loadedMu.Unlock()

				if runner != nil {
					if runner.needsReload(ctx, pending) {
						slog.Debug("reloading", "runner", runner)
						runnerToExpire = runner
					} else {
						// Runner is usable, return it
						logutil.Trace("using existing loaded runner", "model", pendingKey)
						pending.useLoadedRunner(runner, s.finishedReqCh)
						break
					}
				} else if maxRunners > 0 && loadedCount >= int(maxRunners) {
					slog.Debug("max runners achieved, unloading one to make room", "runner_count", loadedCount)
					runnerToExpire = s.findRunnerToUnload()
				} else {
					// Either no models are loaded or below envconfig.MaxRunners
					// Get a refreshed GPU list
					var gpus []ml.DeviceInfo
					if pending.opts.NumGPU == 0 {
						gpus = []ml.DeviceInfo{}
					} else {
						logutil.Trace("refreshing GPU list", "model", pending.model.ModelPath)
						gpus = s.getGpuFn(ctx, runnersSnapshot)
					}
					logutil.Trace("refreshing system information", "model", pending.model.ModelPath)
					systemInfo := s.getSystemInfoFn()
					if maxRunners <= 0 {
						// No user specified MaxRunners, so figure out what automatic setting to use for the next load attempt
						if pending.opts.NumGPU == 0 {
							// Need to get actual GPU list to set the correct default max models
							logutil.Trace("refreshing GPU list", "model", pending.model.ModelPath)
							g := s.getGpuFn(ctx, runnersSnapshot)
							maxRunners = uint(defaultModelsPerGPU * max(len(g), 1))
						} else {
							maxRunners = uint(defaultModelsPerGPU * max(len(gpus), 1))
						}
						slog.Debug("updating default concurrency", "LYCHEE_MAX_LOADED_MODELS", maxRunners, "gpu_count", len(gpus))
					}

					// Update free memory from currently loaded models
					logutil.Trace("updating free space", "gpu_count", len(gpus), "model", pending.model.ModelPath)
					s.updateFreeSpace(gpus)

					if loadedCount == 0 {
						// No models loaded. Load the model but prefer the best fit.
						slog.Debug("loading first model", "model", pending.model.ModelPath)
						if s.loadFn(pending, systemInfo, gpus, false) {
							slog.Debug("first model load requested retry", "model", pending.model.ModelPath)
							continue
						}
						break
					}

					// More than one loaded model, so we have to see if the
					// new one fits
					logutil.Trace("loading additional model", "model", pending.model.ModelPath)
					needEvict := s.loadFn(pending, systemInfo, gpus, true)
					if !needEvict {
						slog.Debug("new model fits with existing models, loading")
						break
					}

					// OOM retry path: load() crashed post-spawn and we still
					// have other models resident. Evict all of them, wait for
					// every unload, then loop back to retry the load once.
					// load() has already set oomRetryAttempted so a second
					// crash falls through to the fail-fast path.
					if pending.oomRetryAttempted {
						if !s.evictAllAndWait(ctx, pendingKey) {
							return
						}
						continue
					}

					runnerToExpire = s.findRunnerToUnload()
				}

				if runnerToExpire == nil {
					// While we were performing load calculations, the loaded runner(s) unloaded in parallel
					// so findRunnerToUnload returned no runners.  We'll try again and the loadedCount should be zero
					slog.Debug("runner to expire was nil, retrying")
					continue
				}
				// Trigger an expiration to unload once it's done
				runnerToExpire.refMu.Lock()
				slog.Debug("resetting model to expire immediately to make room", "runner", runnerToExpire, "refCount", runnerToExpire.refCount)
				if runnerToExpire.expireTimer != nil {
					runnerToExpire.expireTimer.Stop()
					runnerToExpire.expireTimer = nil
				}
				runnerToExpire.sessionDuration = 0
				if runnerToExpire.refCount <= 0 {
					s.expiredCh <- runnerToExpire
				}
				runnerToExpire.refMu.Unlock()
				// Wait for the unload to happen
				slog.Debug("waiting for pending requests to complete and unload to occur", "runner", runnerToExpire)
				select {
				case <-ctx.Done():
					slog.Debug("shutting down scheduler pending loop")
					return
				case <-s.unloadedCh:
					slog.Debug("unload completed", "runner", runnerToExpire)
					continue
				}
			}
		case <-s.unloadedCh:
			// An unload request when there are no pending request can be ignored
			slog.Debug("ignoring unload event with no pending requests")
		}
	}
}

func (s *Scheduler) processCompleted(ctx context.Context) {
	// Process completed requests, expired timers, and unloading models
	for {
		select {
		case <-ctx.Done():
			slog.Debug("shutting down scheduler completed loop")
			return
		case finished := <-s.finishedReqCh:
			finishedKey := schedulerModelKey(finished.model)
			s.loadedMu.Lock()
			runner := s.loaded[finishedKey]
			s.loadedMu.Unlock()
			if runner == nil {
				slog.Error("finished request signal received after model unloaded", "modelPath", finishedKey)
				continue
			}
			runner.refMu.Lock()
			runner.refCount--
			if runner.refCount <= 0 {
				if runner.sessionDuration <= 0 {
					slog.Debug("runner with zero duration has gone idle, expiring to unload", "runner", runner)
					if runner.expireTimer != nil {
						runner.expireTimer.Stop()
						runner.expireTimer = nil
					}
					s.expiredCh <- runner
				} else if runner.expireTimer == nil {
					slog.Debug("runner with non-zero duration has gone idle, adding timer", "runner", runner, "duration", runner.sessionDuration)
					runner.expireTimer = time.AfterFunc(runner.sessionDuration, func() {
						slog.Debug("timer expired, expiring to unload", "runner", runner)
						runner.refMu.Lock()
						defer runner.refMu.Unlock()
						if runner.expireTimer != nil {
							runner.expireTimer.Stop()
							runner.expireTimer = nil
						}
						s.expiredCh <- runner
					})
					runner.expiresAt = time.Now().Add(runner.sessionDuration)
				} else {
					slog.Debug("runner with non-zero duration has gone idle, resetting timer", "runner", runner, "duration", runner.sessionDuration)
					runner.expireTimer.Reset(runner.sessionDuration)
					runner.expiresAt = time.Now().Add(runner.sessionDuration)
				}
			}
			slog.Debug("after processing request finished event", "runner", runner, "refCount", runner.refCount)
			runner.refMu.Unlock()
		case runner := <-s.expiredCh:
			slog.Debug("runner expired event received", "runner", runner)
			runner.refMu.Lock()
			if runner.refCount > 0 {
				slog.Debug("expired event with positive ref count, retrying", "runner", runner, "refCount", runner.refCount)
				go func(runner *runnerRef) {
					// We can't unload yet, but want to as soon as the current request completes
					// So queue up another expired event
					time.Sleep(10 * time.Millisecond)
					s.expiredCh <- runner
				}(runner)
				runner.refMu.Unlock()
				continue
			}

			s.loadedMu.Lock()
			slog.Debug("got lock to unload expired event", "runner", runner)
			runnerToUnload := s.loaded[runner.modelKey]
			if runnerToUnload == nil {
				// If runnerToUnload is nil, we already processed an event and
				// unloaded it. This double unload can happen if the initial
				// request is canceled and we're trying to load another model
				// that requires this one to be evicted, or the settings change
				// and require a reload
				s.loadedMu.Unlock()
				runner.refMu.Unlock()
				slog.Debug("duplicate expired event, ignoring", "runner", runner)
			} else if runner.pid != runnerToUnload.pid {
				// If the pids do not match, we likely had multiple load
				// failures for the same model in quick succession due to
				// request context canceled and are draining the queue of
				// events. Ensure the orphaned runner is properly shut down, but
				// do not delete the mismatched loaded runner, or wait for VRAM
				// convergence.
				slog.Debug("orphaned runner shutting down", "orphan", runner, "loaded", runnerToUnload)
				runner.unload()
				s.loadedMu.Unlock()
				runner.refMu.Unlock()
			} else {
				slog.Debug("starting background wait for VRAM recovery", "runner", runner)
				runnersSnapshot := make([]ml.FilteredRunnerDiscovery, 0, len(s.loaded))
				for _, r := range s.loaded {
					runnersSnapshot = append(runnersSnapshot, r)
				}
				finished := s.waitForVRAMRecovery(runner, runnersSnapshot)
				runner.unload()
				delete(s.loaded, runner.modelKey)
				s.loadedMu.Unlock()
				slog.Debug("runner terminated and removed from list, blocking for VRAM recovery", "runner", runner)
				<-finished
				runner.refMu.Unlock()
				slog.Debug("sending an unloaded event", "runner", runner)
				s.unloadedCh <- struct{}{}
			}
		}
	}
}

// Complete the pending request and send the runner back to the requester
// Wires up a finished event after the request context is completed
// Updates session duration, and resets expiration timer
func (pending *LlmRequest) useLoadedRunner(runner *runnerRef, finished chan *LlmRequest) {
	runner.refMu.Lock()
	defer runner.refMu.Unlock()
	runner.refCount++
	if runner.expireTimer != nil {
		runner.expireTimer.Stop()
		runner.expireTimer = nil
	}
	if pending.sessionDuration != nil {
		runner.sessionDuration = pending.sessionDuration.Duration
	}
	pending.successCh <- runner
	go func() {
		<-pending.ctx.Done()
		slog.Debug("context for request finished", "runner", runner)
		finished <- pending
	}()
}



// TODO consolidate sched_types.go
type runnerRef struct {
	refMu    sync.Mutex
	refCount uint // prevent unloading if > 0

	llama        llm.LlamaServer
	pid          int
	loading      bool          // True only during initial load, then false forever
	gpus         []ml.DeviceID // Recorded at time of provisioning
	discreteGPUs bool          // True if all devices are discrete GPUs - used to skip VRAM recovery check for iGPUs
	isImagegen   bool          // True if loaded via imagegen runner (vs mlxrunner)
	vramSize     uint64
	totalSize    uint64

	sessionDuration time.Duration
	expireTimer     *time.Timer
	expiresAt       time.Time

	model        *Model
	modelPath    string
	modelKey     string
	numParallel  int
	numCtxAuto   bool
	numBatchAuto bool
	useMMapAuto  bool
	contextShift bool
	trainContext int
	*api.Options
}

// The refMu must already be held when calling unload
func (runner *runnerRef) unload() {
	if runner.expireTimer != nil {
		runner.expireTimer.Stop()
		runner.expireTimer = nil
	}
	if runner.llama != nil {
		runner.llama.Close()
	}
	runner.model = nil
	runner.Options = nil
	runner.gpus = nil
	runner.contextShift = false
}

func (runner *runnerRef) needsReload(ctx context.Context, req *LlmRequest) bool {
	slog.Debug("evaluating already loaded", "model", schedulerModelKey(req.model))
	runner.refMu.Lock()
	defer runner.refMu.Unlock()

	// Check if runner type (imagegen vs mlxrunner) matches what's requested.
	wantImagegen := slices.Contains(req.model.Config.Capabilities, "image")
	if runner.isImagegen != wantImagegen {
		return true
	}

	timeout := 10 * time.Second
	if runner.loading {
		timeout = 2 * time.Minute // Initial load can take a long time for big models on slow systems...
	}

	if runner.Options == nil {
		return true
	}

	// Don't reload runner if num_gpu=-1 was provided
	optsExisting := runner.Options.Runner
	optsNew := req.opts.Runner
	optsNew.NumCtx = effectiveContext(optsNew.NumCtx, runner.trainContext)
	if runner.numCtxAuto && req.numCtxAuto {
		optsNew.NumCtx = optsExisting.NumCtx
	}
	if runner.numBatchAuto && req.numBatchAuto {
		optsNew.NumBatch = optsExisting.NumBatch
	}
	if runner.useMMapAuto && optsNew.UseMMap == nil {
		optsNew.UseMMap = optsExisting.UseMMap
	}
	if optsNew.NumGPU < 0 {
		optsExisting.NumGPU = -1
		optsNew.NumGPU = -1
	}

	contextShift := req.contextShift
	if req.model.ModelPath != "" {
		contextShift = resolveContextShift(req.shift, optsNew.NumCtx)
	}
	if runner.contextShift != contextShift {
		return true
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if !reflect.DeepEqual(runner.model.AdapterPaths, req.model.AdapterPaths) || // have the adapters changed?
		!reflect.DeepEqual(runner.model.ProjectorPaths, req.model.ProjectorPaths) || // have the projectors changed?
		(!runner.model.IsMLX() && !reflect.DeepEqual(optsExisting, optsNew)) || // have the runner options changed?
		runner.llama.Ping(ctx) != nil {
		return true
	}

	return false
}



func (runner *runnerRef) LogValue() slog.Value {
	if runner == nil {
		return slog.StringValue("nil")
	}
	modelID := runner.modelPath
	if modelID == "" {
		modelID = runner.modelKey
	}
	attrs := []slog.Attr{}
	if runner.model != nil {
		attrs = append(attrs, slog.String("name", runner.model.Name))
	}
	if len(runner.gpus) > 0 {
		attrs = append(attrs,
			slog.Any("inference", runner.gpus),
		)
	}
	attrs = append(attrs,
		slog.String("size", format.HumanBytes2(runner.totalSize)),
		slog.String("vram", format.HumanBytes2(runner.vramSize)),
		slog.Int("parallel", runner.numParallel),
		slog.Int("pid", runner.pid),
		slog.String("model", modelID),
	)
	if runner.Options != nil {
		attrs = append(attrs, slog.Int("num_ctx", runner.Options.NumCtx))
	}
	return slog.GroupValue(attrs...)
}

// Implements discover.RunnerDiscovery
func (runner *runnerRef) GetPort() int {
	if runner.llama != nil {
		return runner.llama.GetPort()
	}
	return -1
}

func (runner *runnerRef) GetDeviceInfos(ctx context.Context) []ml.DeviceInfo {
	if runner.llama != nil {
		return runner.llama.GetDeviceInfos(ctx)
	}
	return nil
}

func (runner *runnerRef) GetActiveDeviceIDs() []ml.DeviceID {
	return runner.gpus
}

func (runner *runnerRef) HasExited() bool {
	if runner.llama != nil {
		return runner.llama.HasExited()
	}
	return true
}


