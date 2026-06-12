package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/envconfig"
	"github.com/lychee/lychee/format"
	"github.com/lychee/lychee/logutil"
	"github.com/lychee/lychee/ml"
)

func availableMemoryForLoad(systemInfo ml.SystemInfo, gpus []ml.DeviceInfo) (available, gpuFree uint64, systemLimited bool) {
	var sharedGPUFree uint64
	var discreteGPUFree uint64
	for _, gpu := range gpus {
		gpuFree += gpu.FreeMemory
		if gpu.Integrated {
			sharedGPUFree += gpu.FreeMemory
		} else {
			discreteGPUFree += gpu.FreeMemory
		}
	}

	// On iGPUs, GPU free memory can be a static or slowly refreshed device
	// baseline. updateFreeSpace has already subtracted known Lychee runner
	// allocations from that baseline. Current system free memory is a separate
	// live measurement that already includes those loaded runners, so use the
	// smaller value for shared-memory GPUs without discounting discrete VRAM.
	if systemInfo.FreeMemory > 0 && sharedGPUFree > 0 && systemInfo.FreeMemory < sharedGPUFree {
		return discreteGPUFree + systemInfo.FreeMemory, gpuFree, true
	}

	return gpuFree, gpuFree, false
}

func availableMemoryForPlacement(systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, opts api.Options) (available, gpuFree uint64, systemLimited bool) {
	placementGpus := gpusForPlacement(gpus, opts)
	if len(placementGpus) == 1 && opts.MainGPU != nil {
		gpuFree = placementGpus[0].FreeMemory
		available = availableMemoryForGPU(systemInfo, placementGpus[0])
		systemLimited = available < gpuFree
		return available, gpuFree, systemLimited
	}

	return availableMemoryForLoad(systemInfo, placementGpus)
}

func gpusForPlacement(gpus []ml.DeviceInfo, opts api.Options) []ml.DeviceInfo {
	if opts.MainGPU != nil && *opts.MainGPU >= 0 && *opts.MainGPU < len(gpus) {
		return []ml.DeviceInfo{gpus[*opts.MainGPU]}
	}

	return gpus
}

func selectLlamaServerPlacement(systemInfo ml.SystemInfo, gpus []ml.DeviceInfo, predictedVRAM uint64, opts api.Options) ([]ml.DeviceInfo, api.Options) {
	launchOpts := opts
	if len(gpus) <= 1 || opts.NumGPU == 0 {
		return gpus, launchOpts
	}

	groups := ml.ByLibrary(gpus)
	if len(groups) == 0 {
		return gpus, launchOpts
	}

	if opts.MainGPU != nil {
		gpu, available, ok := bestExplicitMainGPU(systemInfo, groups, *opts.MainGPU)
		if !ok {
			selected := bestGPUGroupByAvailableMemory(systemInfo, groups)
			slog.Warn("requested main_gpu is outside the selected GPU group; passing value through to llama-server",
				"main_gpu", *opts.MainGPU,
				"gpu_count", len(selected))
			logSelectedGPUGroup(gpus, selected)
			return selected, launchOpts
		}

		selected, launchOpts := singleLlamaServerGPUPlacement(gpu, launchOpts)
		slog.Info("selecting requested single GPU for llama-server model",
			"requested_main_gpu", *opts.MainGPU,
			"main_gpu", *launchOpts.MainGPU,
			"id", gpu.ID,
			"filter_id", gpu.FilterID,
			"library", gpu.Library,
			"name", gpu.Name,
			"description", gpu.Description,
			"integrated", gpu.Integrated,
			"available", format.HumanBytes2(available))
		logSelectedGPUGroup(gpus, selected)
		return selected, launchOpts
	}

	if !envconfig.SchedSpread() && predictedVRAM > 0 {
		gpu, available, ok := bestSingleGPUFit(systemInfo, groups, predictedVRAM)
		if ok {
			selected, launchOpts := singleLlamaServerGPUPlacement(gpu, launchOpts)
			slog.Info("selecting single GPU for llama-server model",
				"main_gpu", *launchOpts.MainGPU,
				"id", gpu.ID,
				"filter_id", gpu.FilterID,
				"library", gpu.Library,
				"name", gpu.Name,
				"description", gpu.Description,
				"integrated", gpu.Integrated,
				"predicted", format.HumanBytes2(predictedVRAM),
				"available", format.HumanBytes2(available))
			logSelectedGPUGroup(gpus, selected)
			return selected, launchOpts
		}
	}

	selected := bestGPUGroupByAvailableMemory(systemInfo, groups)
	logSelectedGPUGroup(gpus, selected)
	return selected, launchOpts
}

func singleLlamaServerGPUPlacement(gpu ml.DeviceInfo, opts api.Options) ([]ml.DeviceInfo, api.Options) {
	mainGPU := 0
	opts.MainGPU = &mainGPU
	return []ml.DeviceInfo{gpu}, opts
}

func bestExplicitMainGPU(systemInfo ml.SystemInfo, groups [][]ml.DeviceInfo, mainGPU int) (gpu ml.DeviceInfo, available uint64, ok bool) {
	if mainGPU < 0 {
		return ml.DeviceInfo{}, 0, false
	}

	for _, group := range groups {
		if mainGPU >= len(group) {
			continue
		}
		candidate := group[mainGPU]
		candidateAvailable := availableMemoryForGPU(systemInfo, candidate)
		if !ok || betterPlacementGPU(candidate, candidateAvailable, gpu, available) {
			gpu = candidate
			available = candidateAvailable
			ok = true
		}
	}

	return gpu, available, ok
}

func bestSingleGPUFit(systemInfo ml.SystemInfo, groups [][]ml.DeviceInfo, predictedVRAM uint64) (gpu ml.DeviceInfo, available uint64, ok bool) {
	for _, group := range groups {
		for _, candidate := range group {
			candidateAvailable := availableMemoryForGPU(systemInfo, candidate)
			if predictedVRAM > candidateAvailable*80/100 {
				continue
			}
			if !ok || betterPlacementGPU(candidate, candidateAvailable, gpu, available) {
				gpu = candidate
				available = candidateAvailable
				ok = true
			}
		}
	}

	return gpu, available, ok
}

func betterPlacementGPU(candidate ml.DeviceInfo, candidateAvailable uint64, current ml.DeviceInfo, currentAvailable uint64) bool {
	if candidate.Integrated != current.Integrated {
		return !candidate.Integrated
	}

	return candidateAvailable > currentAvailable
}

func bestGPUGroupByAvailableMemory(systemInfo ml.SystemInfo, groups [][]ml.DeviceInfo) []ml.DeviceInfo {
	var best []ml.DeviceInfo
	var bestAvailable uint64
	for _, group := range groups {
		available, _, _ := availableMemoryForLoad(systemInfo, group)
		if best == nil || betterPlacementGroup(group, available, best, bestAvailable) {
			best = group
			bestAvailable = available
		}
	}

	return best
}

func betterPlacementGroup(candidate []ml.DeviceInfo, candidateAvailable uint64, current []ml.DeviceInfo, currentAvailable uint64) bool {
	candidateDiscrete := hasDiscreteGPU(candidate)
	currentDiscrete := hasDiscreteGPU(current)
	if candidateDiscrete != currentDiscrete {
		return candidateDiscrete
	}

	return candidateAvailable > currentAvailable
}

func hasDiscreteGPU(gpus []ml.DeviceInfo) bool {
	for _, gpu := range gpus {
		if !gpu.Integrated {
			return true
		}
	}
	return false
}

func availableMemoryForGPU(systemInfo ml.SystemInfo, gpu ml.DeviceInfo) uint64 {
	if gpu.Integrated && systemInfo.FreeMemory > 0 && systemInfo.FreeMemory < gpu.FreeMemory {
		return systemInfo.FreeMemory
	}

	return gpu.FreeMemory
}

func logSelectedGPUGroup(all, selected []ml.DeviceInfo) {
	if len(selected) == 0 || len(selected) == len(all) {
		return
	}

	slog.Info("selecting GPU backend for llama-server model",
		"library", selected[0].Library,
		"gpu_count", len(selected),
		"available_gpu_count", len(all))
}

func hasDeviceLibrary(gpus []ml.DeviceInfo, library string) bool {
	for _, gpu := range gpus {
		if strings.EqualFold(gpu.Library, library) {
			return true
		}
	}
	return false
}

func allDevicesLibrary(gpus []ml.DeviceInfo, library string) bool {
	if len(gpus) == 0 {
		return false
	}
	for _, gpu := range gpus {
		if !strings.EqualFold(gpu.Library, library) {
			return false
		}
	}
	return true
}

func allDiscreteGPUs(gpus []ml.DeviceInfo) bool {
	if len(gpus) == 0 {
		return false
	}
	for _, gpu := range gpus {
		if gpu.Integrated {
			return false
		}
	}
	return true
}

// Free memory reporting on GPUs can lag for a while even after the runner
// exits, so we have to keep checking until we see the available memory recover,
// otherwise subsequent model loads will get far less layers loaded or worse
// case, may completely fall back to CPU mode.
// This routine must be called before the runner unloads so it can establish
// a before and after GPU memory allocation.  The returned channel
// will be notified when we're done waiting, or have timed out and should
// proceed anyway
func (s *Scheduler) waitForVRAMRecovery(runner *runnerRef, runners []ml.FilteredRunnerDiscovery) chan any {
	finished := make(chan any, 1)

	// CPU, Metal and iGPUs don't need checking, so no waiting required
	if len(runner.gpus) == 0 || !runner.discreteGPUs ||
		(len(runner.gpus) == 1 && runner.gpus[0].Library == "Metal") {
		finished <- struct{}{}
		slog.Debug("no need to wait for VRAM recovery", "runner", runner)
		return finished
	}
	start := time.Now()

	// Establish a baseline before we unload
	gpusBefore := s.getGpuFn(context.Background(), runners)
	var totalMemoryBefore, freeMemoryBefore uint64
	for _, gpu := range gpusBefore {
		totalMemoryBefore += gpu.TotalMemory
		freeMemoryBefore += gpu.FreeMemory
	}
	totalMemoryNow := totalMemoryBefore
	freeMemoryNow := freeMemoryBefore

	go func() {
		// typical convergence is 0.5-1.5s - If it takes too long to discover and converge, let the scheduler estimate VRAM usage
		ctx, cancel := context.WithTimeout(context.Background(), s.waitForRecovery)
		defer cancel()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Query GPUs, look for free to go back up
				gpusNow := s.getGpuFn(ctx, runners)
				totalMemoryNow = 0
				freeMemoryNow = 0
				for _, gpu := range gpusNow {
					totalMemoryNow += gpu.TotalMemory
					freeMemoryNow += gpu.FreeMemory
				}
				if freeMemoryNow > freeMemoryBefore {
					logutil.Trace("gpu VRAM convergence", "percent", int(float32(freeMemoryNow-freeMemoryBefore)/float32(runner.vramSize)*100))
				} else {
					logutil.Trace("gpu VRAM convergence", "percent", 0)
				}
				// If we're within ~75% of the estimated memory usage recovered, bail out
				if float32(freeMemoryNow-freeMemoryBefore) > float32(runner.vramSize)*0.75 {
					slog.Debug(fmt.Sprintf("gpu VRAM free memory converged after %0.2f seconds", time.Since(start).Seconds()), "free_before", format.HumanBytes2(freeMemoryBefore), "free_now", format.HumanBytes2(freeMemoryNow), "runner", runner)
					finished <- struct{}{}
					return
				}
			case <-ctx.Done():
				slog.Debug("gpu VRAM usage didn't recover within timeout", "seconds", time.Since(start).Seconds(), "free_before", format.HumanBytes2(freeMemoryBefore), "free_now", format.HumanBytes2(freeMemoryNow), "runner", runner)
				finished <- struct{}{}
				return
			}
		}
	}()
	return finished
}

func (s *Scheduler) updateFreeSpace(allGpus []ml.DeviceInfo) {
	if len(allGpus) == 0 {
		return
	}
	predMap := map[ml.DeviceID]uint64{} // Sum up the total predicted usage per GPU for all runners
	s.loadedMu.Lock()
	runners := make([]*runnerRef, 0, len(s.loaded))
	for _, r := range s.loaded {
		runners = append(runners, r)
	}
	s.loadedMu.Unlock()
	for _, r := range runners {
		r.refMu.Lock()
		if r.llama != nil {
			for _, gpu := range allGpus {
				predMap[gpu.DeviceID] += r.llama.VRAMByGPU(gpu.DeviceID)
			}
		} else {
			slog.Warn("unexpected nil runner reference, memory prediction may be incorrect")
		}
		r.refMu.Unlock()
	}

	// Now that we've summed up all the GPU usage predictions across all the loaded runners, update the gpu list
	for i := range allGpus {
		if p, ok := predMap[allGpus[i].DeviceID]; ok {
			slog.Debug("gpu reported", "gpu", allGpus[i].ID, "library", allGpus[i].Library, "available", format.HumanBytes2(allGpus[i].FreeMemory))
			if p > allGpus[i].TotalMemory {
				// Shouldn't happen
				slog.Warn("predicted usage exceeds VRAM", "gpu", allGpus[i].ID, "totalMemory", allGpus[i].TotalMemory, "predicted", p)
				allGpus[i].FreeMemory = 0
			} else if (allGpus[i].TotalMemory - p) < allGpus[i].FreeMemory { // predicted free is smaller than reported free, use it
				// TODO maybe we should just always trust our numbers, since cuda's free memory reporting is laggy
				// and we might unload models we didn't actually need to.  The risk is if some other GPU intensive app is loaded
				// after we start our first runner, then we'll never account for that, so picking the smallest free value seems prudent.
				allGpus[i].FreeMemory = allGpus[i].TotalMemory - p
			}
			slog.Info("updated VRAM based on existing loaded models", "gpu", allGpus[i].ID, "library", allGpus[i].Library, "total", format.HumanBytes2(allGpus[i].TotalMemory), "available", format.HumanBytes2(allGpus[i].FreeMemory))
		}
	}
}
