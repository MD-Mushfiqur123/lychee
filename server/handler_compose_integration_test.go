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
	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/llm"
)

func TestComposeIntegration(t *testing.T) {
	t.Setenv("LYCHEE_MODELS", t.TempDir())
	gin.SetMode(gin.TestMode)

	mock := mockRunner{
		CompletionFn: func(ctx context.Context, req llm.CompletionRequest, fn func(llm.CompletionResponse)) error {
			content := "default"
			if strings.Contains(req.Prompt, "translate") {
				content = "bonjour"
			}
			fn(llm.CompletionResponse{
				Content: content,
				Done:    true,
			})
			return nil
		},
	}

	s := newServerWithMockRunner(t, &mock)
	createMinimalGGUFModel(t, s, "test-model", nil, "", nil)

	r := gin.New()
	r.POST("/api/compose", s.ComposeHandler)

	reqBody := api.ComposeRequest{
		Input: "hello",
		Steps: []api.ComposeStep{
			{
				Model:  "test-model",
				Prompt: "translate to French: {{input}}",
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

	if resp.Output != "bonjour" {
		t.Errorf("expected output 'bonjour', got %q", resp.Output)
	}
}
