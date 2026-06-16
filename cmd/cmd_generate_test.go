package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func TestGenerateClientCmd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("expected path /api/show, got %s", r.URL.Path)
		}

		var req api.ShowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		resp := map[string]any{
			"system":       "You are a helpful assistant.",
			"capabilities": []string{"chat"},
			"model_info": map[string]any{
				"general.architecture": "llama",
				"llama.context_length": float64(8192),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	os.Setenv("LYCHEE_HOST", ts.URL)
	os.Setenv("OLLAMA_HOST", ts.URL)
	defer os.Unsetenv("LYCHEE_HOST")
	defer os.Unsetenv("OLLAMA_HOST")

	languages := []string{"python", "js", "rs", "go"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			cmd := NewGenerateClientCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			cmd.SetArgs([]string{"--lang", lang, "test-model"})

			err := cmd.ExecuteContext(context.Background())
			if err != nil {
				t.Fatalf("command execute failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "test-model") {
				t.Errorf("expected output to contain model name, got %s", output)
			}
			if lang == "python" && !strings.Contains(output, "pip install ollama") {
				t.Errorf("expected Python installer reference to use ollama, got %s", output)
			}
			if lang == "js" && !strings.Contains(output, "npm install ollama") {
				t.Errorf("expected JS installer reference to use ollama, got %s", output)
			}
		})
	}
}

func TestGenerateClientCmdNoContextLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"system":       "You are a helpful assistant.",
			"capabilities": []string{"chat"},
			// No model info or context length
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	os.Setenv("LYCHEE_HOST", ts.URL)
	os.Setenv("OLLAMA_HOST", ts.URL)
	defer os.Unsetenv("LYCHEE_HOST")
	defer os.Unsetenv("OLLAMA_HOST")

	cmd := NewGenerateClientCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"--lang", "python", "test-model"})

	err := cmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatalf("command execute failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "could not detect context length from server") {
		t.Errorf("expected fallback comment in output, got %s", output)
	}
}
