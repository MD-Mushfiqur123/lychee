package server

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/MD-Mushfiqur123/lychee/llm"
)

type ByDurationAndName []*runnerRef

func (a ByDurationAndName) Len() int      { return len(a) }
func (a ByDurationAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDurationAndName) Less(i, j int) bool {
	// Primary sort by session duration (uint64 to handle negatives)
	d1 := uint64(a[i].sessionDuration)
	d2 := uint64(a[j].sessionDuration)
	if d1 != d2 {
		return d1 < d2
	}
	// Secondary sort by model key/path lex order
	n1 := a[i].modelPath
	if n1 == "" {
		n1 = a[i].modelKey
	}
	n2 := a[j].modelPath
	if n2 == "" {
		n2 = a[j].modelKey
	}
	return n1 < n2
}

// evictAllAndWait synchronously expires every currently loaded runner except
// the one being loaded (matched by modelKey) and waits for all unload events
// to drain. Returns false if the context was cancelled mid-wait so the caller
// can exit the scheduling loop. Used by the OOM retry path in processPending.
func (s *Scheduler) evictAllAndWait(ctx context.Context, keepKey string) bool {
	s.loadedMu.Lock()
	runnersToExpire := make([]*runnerRef, 0, len(s.loaded))
	for key, r := range s.loaded {
		if key == keepKey {
			continue
		}
		runnersToExpire = append(runnersToExpire, r)
	}
	s.loadedMu.Unlock()

	if len(runnersToExpire) == 0 {
		return true
	}

	slog.Info("evicting all other loaded models for OOM retry", "count", len(runnersToExpire))
	for _, runner := range runnersToExpire {
		runner.refMu.Lock()
		if runner.expireTimer != nil {
			runner.expireTimer.Stop()
			runner.expireTimer = nil
		}
		runner.sessionDuration = 0
		if runner.refCount <= 0 {
			s.expiredCh <- runner
		}
		runner.refMu.Unlock()
	}

	// Wait for every unload event. Each runner produces exactly one
	// unloadedCh signal when its cleanup finishes.
	for range runnersToExpire {
		select {
		case <-ctx.Done():
			slog.Debug("shutting down scheduler during evict-all wait")
			return false
		case <-s.unloadedCh:
		}
	}
	return true
}

func (s *Scheduler) expireRunnersForRuntimeOOM(model *Model, err error) {
	if !llm.IsOutOfMemory(err) {
		return
	}

	s.loadedMu.Lock()
	runners := make([]*runnerRef, 0, len(s.loaded))
	for _, runner := range s.loaded {
		runners = append(runners, runner)
	}
	s.loadedMu.Unlock()

	if len(runners) == 0 {
		return
	}

	slog.Warn("runtime OOM detected; expiring loaded models to clear memory before next request", "model", schedulerModelKey(model), "error", err)
	for _, runner := range runners {
		runner.refMu.Lock()
		if runner.expireTimer != nil {
			runner.expireTimer.Stop()
			runner.expireTimer = nil
		}
		runner.sessionDuration = 0
		if runner.refCount <= 0 {
			s.expiredCh <- runner
		}
		runner.refMu.Unlock()
	}
}

// findRunnerToUnload finds a runner to unload to make room for a new model
func (s *Scheduler) findRunnerToUnload() *runnerRef {
	s.loadedMu.Lock()
	runnerList := make([]*runnerRef, 0, len(s.loaded))
	for _, r := range s.loaded {
		runnerList = append(runnerList, r)
	}
	s.loadedMu.Unlock()
	if len(runnerList) == 0 {
		slog.Debug("no loaded runner to unload")
		return nil
	}

	// In the future we can enhance the algorithm to be smarter about picking the optimal runner to unload
	// e.g., if we have multiple options, will one make room for the request?
	sort.Sort(ByDurationAndName(runnerList))

	// First try to find a runner that's already idle
	for _, runner := range runnerList {
		runner.refMu.Lock()
		rc := runner.refCount
		runner.refMu.Unlock()
		if rc == 0 {
			slog.Debug("found an idle runner to unload", "runner", runner)
			return runner
		}
	}
	// None appear idle, just wait for the one with the shortest duration
	slog.Debug("no idle runners, picking the shortest duration", "runner_count", len(runnerList), "runner", runnerList[0])
	return runnerList[0]
}

func (s *Scheduler) unloadAllRunners() {
	s.loadedMu.Lock()
	defer s.loadedMu.Unlock()

	if s.activeLoading != nil {
		slog.Debug("shutting down currently loading runner")
		s.activeLoading.Close()
		s.activeLoading = nil
	}

	for model, runner := range s.loaded {
		if runner.llama != nil {
			slog.Debug("shutting down runner", "model", model)
			runner.llama.Close()
		}
	}
}

func (s *Scheduler) expireRunner(model *Model) {
	modelKey := schedulerModelKey(model)
	s.loadedMu.Lock()
	runner, ok := s.loaded[modelKey]
	s.loadedMu.Unlock()
	if ok {
		runner.refMu.Lock()
		runner.expiresAt = time.Now()
		if runner.expireTimer != nil {
			runner.expireTimer.Stop()
			runner.expireTimer = nil
		}
		runner.sessionDuration = 0
		if runner.refCount <= 0 {
			s.expiredCh <- runner
		}
		runner.refMu.Unlock()
	}
}
