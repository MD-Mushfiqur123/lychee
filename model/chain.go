package model

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lychee/lychee/api"
)

// ComposeEvent represents a progress event streamed during composition.
type ComposeEvent struct {
	Event  string               `json:"event"`
	Index  int                  `json:"index,omitempty"`
	Model  string               `json:"model,omitempty"`
	Text   string               `json:"text,omitempty"`
	Output string               `json:"output,omitempty"`
	Result *api.ComposeResponse `json:"result,omitempty"`
}

// ExecuteChain runs a composition request sequentially and in parallel, substituting templates.
func ExecuteChain(
	ctx context.Context,
	req *api.ComposeRequest,
	runStep func(ctx context.Context, modelName string, prompt string, options map[string]any, onChunk func(string)) (string, error),
	onEvent func(ComposeEvent),
) (*api.ComposeResponse, error) {
	results := make([]api.StepResult, 0, len(req.Steps))
	currentInput := req.Input

	for i, step := range req.Steps {
		if onEvent != nil {
			onEvent(ComposeEvent{
				Event: "step_start",
				Index: i,
				Model: step.Model,
			})
		}

		// Helper to replace templates in the prompt
		replaceTemplates := func(prompt string) string {
			prompt = strings.ReplaceAll(prompt, "{{input}}", currentInput)
			for j, res := range results {
				placeholder := fmt.Sprintf("{{step[%d].output}}", j)
				prompt = strings.ReplaceAll(prompt, placeholder, res.Output)
				for k, pRes := range res.ParallelResults {
					placeholder := fmt.Sprintf("{{step[%d].parallel[%d].output}}", j, k)
					prompt = strings.ReplaceAll(prompt, placeholder, pRes.Output)
				}
			}
			return prompt
		}

		// Run a step with optional timeout and fallback
		execWithRetry := func(ctx context.Context, s api.ComposeStep, stepIdx int, isParallel bool, parallelIdx int, onChunk func(string)) (string, error) {
			runCtx := ctx
			if s.TimeoutSec > 0 {
				var cancel context.CancelFunc
				runCtx, cancel = context.WithTimeout(ctx, time.Duration(s.TimeoutSec)*time.Second)
				defer cancel()
			}

			prompt := replaceTemplates(s.Prompt)

			output, err := runStep(runCtx, s.Model, prompt, s.Options, onChunk)
			if err != nil && s.FallbackModel != "" {
				// Retry with fallback model
				if onEvent != nil {
					fallbackMsg := fmt.Sprintf("step failed, retrying with fallback model: %s", s.FallbackModel)
					onEvent(ComposeEvent{
						Event: "step_fallback",
						Index: stepIdx,
						Model: s.FallbackModel,
						Text:  fallbackMsg,
					})
				}
				output, err = runStep(runCtx, s.FallbackModel, prompt, s.Options, onChunk)
			}
			return output, err
		}

		// Channel and variables for concurrency
		var parallelResults []api.StepResult
		var wg sync.WaitGroup
		var parallelErr error
		var mu sync.Mutex

		if len(step.Parallel) > 0 {
			parallelResults = make([]api.StepResult, len(step.Parallel))
			for idx, pStep := range step.Parallel {
				wg.Add(1)
				go func(pIdx int, ps api.ComposeStep) {
					defer wg.Done()
					pOnChunk := func(text string) {
						if onEvent != nil {
							onEvent(ComposeEvent{
								Event: "parallel_progress",
								Index: i,
								Model: ps.Model,
								Text:  text,
							})
						}
					}

					pOut, pErr := execWithRetry(ctx, ps, i, true, pIdx, pOnChunk)
					mu.Lock()
					defer mu.Unlock()
					if pErr != nil {
						parallelErr = fmt.Errorf("parallel step %d (%s) failed: %w", pIdx, ps.Model, pErr)
						return
					}
					parallelResults[pIdx] = api.StepResult{
						Model:  ps.Model,
						Output: pOut,
					}
				}(idx, pStep)
			}
		}

		// Run the main step
		mainOnChunk := func(text string) {
			if onEvent != nil {
				onEvent(ComposeEvent{
					Event: "step_progress",
					Index: i,
					Model: step.Model,
					Text:  text,
				})
			}
		}

		output, err := execWithRetry(ctx, step, i, false, 0, mainOnChunk)
		if err != nil {
			return nil, fmt.Errorf("step %d (%s) failed: %w", i, step.Model, err)
		}

		// Wait for parallel steps to complete
		wg.Wait()
		if parallelErr != nil {
			return nil, parallelErr
		}

		if onEvent != nil {
			onEvent(ComposeEvent{
				Event:  "step_complete",
				Index:  i,
				Model:  step.Model,
				Output: output,
			})
		}

		stepRes := api.StepResult{
			Model:           step.Model,
			Output:          output,
			ParallelResults: parallelResults,
		}
		results = append(results, stepRes)
		currentInput = output
	}

	finalOutput := ""
	if len(results) > 0 {
		finalOutput = results[len(results)-1].Output
	}

	resp := &api.ComposeResponse{
		Output:  finalOutput,
		Results: results,
	}

	if onEvent != nil {
		onEvent(ComposeEvent{
			Event:  "complete",
			Result: resp,
		})
	}

	return resp, nil
}
