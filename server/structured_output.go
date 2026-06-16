package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/llm"
	"github.com/MD-Mushfiqur123/lychee/types/model"
)

// StructuredOpts configures structured output generation with retries.
type StructuredOpts struct {
	Model      string
	Prompt     string
	Schema     json.RawMessage
	MaxRetries int // default: 3
	Options    map[string]any
	TimeoutSec int // default: 60
	OnEvent    func(api.StructuredEvent)
}

// StructuredResult contains the generation result and retry metadata.
type StructuredResult struct {
	Output   string   `json:"output"`
	Valid    bool     `json:"valid"`
	Attempts int      `json:"attempts"`
	Errors   []string `json:"errors,omitempty"`
}

// generateStructured runs generation with schema validation and auto-retry.
// On each failed attempt, it constructs a correction prompt containing
// the previous output and validation error, asking the model to fix it.
func (s *Server) generateStructured(ctx context.Context, opts StructuredOpts) (*StructuredResult, error) {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}

	modelRef, err := parseAndValidateModelRef(opts.Model)
	if err != nil {
		return nil, err
	}
	name := modelRef.Name
	name, err = getExistingName(name)
	if err != nil {
		return nil, err
	}
	m, err := GetModel(name.String())
	if err != nil {
		return nil, err
	}

	caps := []model.Capability{model.CapabilityCompletion}
	r, m, runOpts, err := s.scheduleRunner(ctx, name.String(), caps, opts.Options, nil, nil)
	if err != nil {
		return nil, err
	}

	emit := func(event api.StructuredEvent) {
		if opts.OnEvent != nil {
			opts.OnEvent(event)
		}
	}

	currentPrompt := opts.Prompt
	var errorsList []string
	var lastOutput string

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		emit(api.StructuredEvent{Event: "attempt_start", Attempt: attempt})

		var responseSB strings.Builder
		leadingBOS := leadingBOSForModel(m)

		var attemptOpts *api.Options
		if runOpts != nil {
			cop := *runOpts
			attemptOpts = &cop
		}
		if attempt > 1 && attemptOpts != nil {
			if attemptOpts.Temperature == 0 {
				attemptOpts.Temperature = 0.8
			}
			attemptOpts.Temperature += 0.1 * float32(attempt-1)
			if attemptOpts.Temperature > 1.2 {
				attemptOpts.Temperature = 1.2
			}
		}

		timeoutSec := opts.TimeoutSec
		if timeoutSec <= 0 {
			timeoutSec = 60
		}
		attemptCtx, cancelAttempt := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)

		err = r.Completion(attemptCtx, llm.CompletionRequest{
			Prompt:     currentPrompt,
			Options:    attemptOpts,
			LeadingBOS: leadingBOS,
		}, func(cr llm.CompletionResponse) {
			responseSB.WriteString(cr.Content)
		})
		cancelAttempt()

		if err != nil {
			errStr := err.Error()
			errorsList = append(errorsList, errStr)
			emit(api.StructuredEvent{Event: "attempt_fail", Attempt: attempt, Error: errStr})

			currentPrompt = fmt.Sprintf(
				"%s\n\nYour previous attempt failed with error: %s.\nPlease try again.",
				opts.Prompt, errStr,
			)
			continue
		}

		lastOutput = responseSB.String()
		emit(api.StructuredEvent{Event: "attempt_output", Attempt: attempt, Output: lastOutput})

		// If Schema is empty/nil/empty object/null, bypass validation and return success
		if len(opts.Schema) == 0 || string(opts.Schema) == "null" || string(opts.Schema) == "{}" || string(opts.Schema) == "" {
			res := &StructuredResult{
				Output:   lastOutput,
				Valid:    true,
				Attempts: attempt,
			}
			emit(api.StructuredEvent{
				Event: "complete",
				Response: &api.StructuredResponse{
					Output:   res.Output,
					Valid:    res.Valid,
					Attempts: res.Attempts,
				},
			})
			return res, nil
		}

		// Validate schema
		valErr := ValidateJSONSchema(lastOutput, opts.Schema)
		if valErr == nil {
			res := &StructuredResult{
				Output:   lastOutput,
				Valid:    true,
				Attempts: attempt,
			}
			emit(api.StructuredEvent{
				Event: "complete",
				Response: &api.StructuredResponse{
					Output:   res.Output,
					Valid:    res.Valid,
					Attempts: res.Attempts,
				},
			})
			return res, nil
		}

		errorsList = append(errorsList, valErr.Error())
		emit(api.StructuredEvent{Event: "attempt_fail", Attempt: attempt, Error: valErr.Error()})

		// Construct correction prompt
		currentPrompt = fmt.Sprintf(
			"%s\n\nYour previous response was invalid JSON or did not conform to the schema.\nError: %s\nPrevious response: %s\nPlease output corrected JSON matching the schema.",
			opts.Prompt, valErr.Error(), lastOutput,
		)
	}

	res := &StructuredResult{
		Output:   lastOutput,
		Valid:    false,
		Attempts: opts.MaxRetries,
		Errors:   errorsList,
	}
	emit(api.StructuredEvent{
		Event: "complete",
		Response: &api.StructuredResponse{
			Output:   res.Output,
			Valid:    res.Valid,
			Attempts: res.Attempts,
			Errors:   res.Errors,
		},
	})
	return res, nil
}
