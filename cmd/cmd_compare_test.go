package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func TestCompareHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("expected path /api/chat, got %s", r.URL.Path)
		}

		var req api.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/x-ndjson")

		// Stream two chat chunks
		chunk1 := api.ChatResponse{
			Model:   req.Model,
			Message: api.Message{Role: "assistant", Content: "Hello from " + req.Model},
			Done:    false,
		}
		chunk2 := api.ChatResponse{
			Model:         req.Model,
			Message:       api.Message{Role: "assistant", Content: "."},
			Done:          true,
			EvalCount:     5,
			TotalDuration: 500000000, // 0.5s
		}

		json.NewEncoder(w).Encode(chunk1)
		json.NewEncoder(w).Encode(chunk2)
	}))
	defer ts.Close()

	os.Setenv("LYCHEE_HOST", ts.URL)
	os.Setenv("OLLAMA_HOST", ts.URL)
	defer os.Unsetenv("LYCHEE_HOST")
	defer os.Unsetenv("OLLAMA_HOST")

	err := compareHandler(context.Background(), "modelA", "modelB", "test prompt", "system instructions", 100, false)
	if err != nil {
		t.Fatalf("compareHandler failed: %v", err)
	}
}
