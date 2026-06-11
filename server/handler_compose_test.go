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

	r := gin.New()
	r.POST("/api/compose", s.ComposeHandler)

	t.Run("successful sequential composition", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.Step{
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

		if resp.Results[0].Output != "result-one" {
			t.Errorf("expected step 0 output 'result-one', got %q", resp.Results[0].Output)
		}

		if resp.Results[1].Output != "result-two" {
			t.Errorf("expected step 1 output 'result-two', got %q", resp.Results[1].Output)
		}
	})

	t.Run("empty steps validation", func(t *testing.T) {
		reqBody := api.ComposeRequest{
			Input: "start-input",
			Steps: []api.Step{},
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
