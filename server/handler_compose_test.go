package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/api"
	"github.com/lychee/lychee/llm"
)

func TestComposeHandler(t *testing.T) {
	t.Setenv("LYCHEE_MODELS", t.TempDir())
	gin.SetMode(gin.TestMode)

	mock := mockRunner{
		CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
			content := "output-default"
			if strings.Contains(req.Prompt, "step 1") {
				content = "result-one"
			} else if strings.Contains(req.Prompt, "result-one") {
				content = "result-two"
			} else if strings.Contains(req.Prompt, "parallel 1") {
				content = "parallel-one"
			} else if strings.Contains(req.Prompt, "fallback") {
				content = "fallback-success"
			}
			fn(llm.CompletionResponse{
				Content: content,
				Done:    true,
			})
			return nil
		},
	}

	s := newServerWithMockRunner(t, &mock)
	createMinimalGGUFModel(t, s, "test-compose-model", nil, "", nil)
	createMinimalGGUFModel(t, s, "test-fallback-model", nil, "", nil)

	r := gin.New()
	r.POST("/api/compose", s.ComposeHandler)

	t.Run("successful sequential composition", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.ComposeStep{
				{
					Model:  "test-compose-model",
					Prompt: "step 1: {{input}}",
				},
				{
					Model:  "test-compose-model",
					Prompt: "step 2: {{step[0].output}}",
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/compose", bytes.NewReader(jsonData))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.ComposeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Output != "result-two" {
			t.Errorf("expected final output 'result-two', got %q", resp.Output)
		}

		if len(resp.Results) != 2 {
			t.Fatalf("expected 2 step results, got %d", len(resp.Results))
		}
	})

	t.Run("concurrency and parallel executions", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.ComposeStep{
				{
					Model:  "test-compose-model",
					Prompt: "step 1: {{input}}",
					Parallel: []api.ComposeStep{
						{
							Model:  "test-compose-model",
							Prompt: "parallel 1: {{input}}",
						},
					},
				},
				{
					Model:  "test-compose-model",
					Prompt: "step 2: {{step[0].output}} and {{step[0].parallel[0].output}}",
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/compose", bytes.NewReader(jsonData))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.ComposeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(resp.Results) != 2 {
			t.Fatalf("expected 2 step results, got %d", len(resp.Results))
		}

		if len(resp.Results[0].ParallelResults) != 1 {
			t.Fatalf("expected 1 parallel result, got %d", len(resp.Results[0].ParallelResults))
		}

		if resp.Results[0].ParallelResults[0].Output != "parallel-one" {
			t.Errorf("expected parallel output 'parallel-one', got %q", resp.Results[0].ParallelResults[0].Output)
		}
	})

	t.Run("timeouts and fallbacks", func(t *testing.T) {
		// Mock failing model behavior that triggers fallback
		failMock := mockRunner{
			CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
				if strings.Contains(req.Prompt, "fail") {
					return context.DeadlineExceeded // simulate timeout/error
				}
				fn(llm.CompletionResponse{
					Content: "fallback-worked",
					Done:    true,
				})
				return nil
			},
		}

		sFail := newServerWithMockRunner(t, &failMock)
		createMinimalGGUFModel(t, sFail, "failing-model", nil, "", nil)
		createMinimalGGUFModel(t, sFail, "working-fallback-model", nil, "", nil)

		rFail := gin.New()
		rFail.POST("/api/compose", sFail.ComposeHandler)

		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.ComposeStep{
				{
					Model:         "failing-model",
					Prompt:        "fail prompt",
					FallbackModel: "working-fallback-model",
					TimeoutSec:    2,
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/compose", bytes.NewReader(jsonData))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		rFail.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 via fallback, got %d: %s", w.Code, w.Body.String())
		}

		var resp api.ComposeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Output != "fallback-worked" {
			t.Errorf("expected output to be 'fallback-worked' via retry, got %q", resp.Output)
		}
	})

	t.Run("streaming SSE composition", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.ComposeStep{
				{
					Model:  "test-compose-model",
					Prompt: "step 1: {{input}}",
				},
			},
			Stream: true,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/compose", bytes.NewReader(jsonData))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "event: message") {
			t.Errorf("expected response to contain SSE events, got: %s", body)
		}
		if !strings.Contains(body, "step_start") {
			t.Errorf("expected response to contain 'step_start' event, got: %s", body)
		}
		if !strings.Contains(body, "complete") {
			t.Errorf("expected response to contain 'complete' event, got: %s", body)
		}
	})

	t.Run("empty steps validation", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.ComposeStep{},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/api/compose", bytes.NewReader(jsonData))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}
