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

func TestEmbedHandler(t *testing.T) {
	// Start a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("expected path /api/embed, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Decode request to ensure it's correct
		var req api.EmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}

		// Respond with mock embeddings
		resp := api.EmbedResponse{
			Model:      req.Model,
			Embeddings: [][]float32{{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Set host env var so ClientFromEnvironment uses the mock server
	os.Setenv("LYCHEE_HOST", ts.URL)
	os.Setenv("OLLAMA_HOST", ts.URL)
	defer os.Unsetenv("LYCHEE_HOST")
	defer os.Unsetenv("OLLAMA_HOST")

	// Call embedHandler
	err := embedHandler(context.Background(), []string{"test input"}, "test-model", "text", true)
	if err != nil {
		t.Fatalf("embedHandler failed: %v", err)
	}
}

func TestEmbedHandlerJSON(t *testing.T) {
	// Start mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.EmbedResponse{
			Model:      "test-model",
			Embeddings: [][]float32{{0.1, 0.2}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	os.Setenv("LYCHEE_HOST", ts.URL)
	os.Setenv("OLLAMA_HOST", ts.URL)
	defer os.Unsetenv("LYCHEE_HOST")
	defer os.Unsetenv("OLLAMA_HOST")

	// Call embedHandler with JSON format
	err := embedHandler(context.Background(), []string{"test input"}, "test-model", "json", true)
	if err != nil {
		t.Fatalf("embedHandler failed: %v", err)
	}
}
