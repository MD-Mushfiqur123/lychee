package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	lychee "github.com/MD-Mushfiqur123/lychee/server/internal/client/lychee"
	"github.com/MD-Mushfiqur123/lychee/anthropic"
	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/llm"
)

func TestAnthropicMessagesRoute_Integration(t *testing.T) {
	t.Setenv("LYCHEE_CONTEXT_LENGTH", "4096")
	t.Setenv("LYCHEE_GO_TEMPLATE", "")

	mock := mockRunner{
		ChatFn: func(_ context.Context, req llm.ChatRequest, fn func(llm.ChatResponse)) error {
			fn(llm.ChatResponse{
				Message:            api.Message{Role: "assistant", Content: "Hello from the Anthropic compatibility layer!"},
				Done:               true,
				DoneReason:         llm.DoneReasonStop,
				PromptEvalCount:    10,
				PromptEvalDuration: time.Millisecond,
				EvalCount:          15,
				EvalDuration:       2 * time.Millisecond,
			})
			return nil
		},
	}
	s := newServerWithMockRunner(t, &mock)
	createMinimalGGUFModel(t, s, "anthropic-model", nil, "", nil)

	rc := &lychee.Registry{
		HTTPClient: panicOnRoundTrip,
	}

	router, err := s.GenerateRoutes(rc)
	if err != nil {
		t.Fatalf("failed to generate routes: %v", err)
	}

	httpSrv := httptest.NewServer(router)
	t.Cleanup(httpSrv.Close)

	t.Run("non-streaming basic request", func(t *testing.T) {
		body := `{"model": "anthropic-model", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hi there"}]}`
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, httpSrv.URL+"/v1/messages", strings.NewReader(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpSrv.Client().Do(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var anthropicResp anthropic.MessagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if anthropicResp.Role != "assistant" {
			t.Errorf("expected role 'assistant', got %q", anthropicResp.Role)
		}

		if len(anthropicResp.Content) != 1 {
			t.Fatalf("expected 1 content block, got %d", len(anthropicResp.Content))
		}

		if anthropicResp.Content[0].Type != "text" || anthropicResp.Content[0].Text == nil || *anthropicResp.Content[0].Text != "Hello from the Anthropic compatibility layer!" {
			t.Errorf("unexpected content block: %+v", anthropicResp.Content[0])
		}

		if anthropicResp.Usage.InputTokens != 10 {
			t.Errorf("expected input tokens 10, got %d", anthropicResp.Usage.InputTokens)
		}

		if anthropicResp.Usage.OutputTokens != 15 {
			t.Errorf("expected output tokens 15, got %d", anthropicResp.Usage.OutputTokens)
		}
	})

	t.Run("streaming request", func(t *testing.T) {
		body := `{"model": "anthropic-model", "max_tokens": 1024, "stream": true, "messages": [{"role": "user", "content": "Hi there"}]}`
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, httpSrv.URL+"/v1/messages", strings.NewReader(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpSrv.Client().Do(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected content type text/event-stream, got %q", contentType)
		}
	})

	t.Run("missing model bad request", func(t *testing.T) {
		body := `{"max_tokens": 1024, "messages": [{"role": "user", "content": "Hi there"}]}`
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, httpSrv.URL+"/v1/messages", strings.NewReader(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpSrv.Client().Do(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}

		var errorResp anthropic.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errorResp.Error.Type != "invalid_request_error" {
			t.Errorf("expected error type invalid_request_error, got %q", errorResp.Error.Type)
		}

		if !strings.Contains(errorResp.Error.Message, "model is required") {
			t.Errorf("expected message to contain 'model is required', got %q", errorResp.Error.Message)
		}
	})
}
