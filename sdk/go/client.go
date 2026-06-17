// Package lychee provides a Go client for the Lychee local LLM runtime.
//
// Lychee is a universal local LLM runtime and orchestration layer that
// supports HuggingFace model pulls, multi-model DAG pipelines, structured
// output, persistent conversation memory, and load-balanced routing.
//
// Usage:
//
//	client := lychee.NewClient("http://localhost:11434")
//
//	resp, err := client.Generate(ctx, &lychee.GenerateRequest{
//	    Model:  "gemma3",
//	    Prompt: "Explain quantum computing in one sentence.",
//	})
//	fmt.Println(resp.Response)
//
//	models, err := client.ListModels(ctx)
//	for _, m := range models {
//	    fmt.Println(m.Name, m.Size)
//	}
//
//	if err := client.Health(ctx); err == nil {
//	    fmt.Println("Server is running!")
//	}
package lychee

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Client
// ─────────────────────────────────────────────────────────────────────────────

// Client is an HTTP client for the Lychee server API.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a new Client. If baseURL is empty, defaults to
// "http://localhost:11434".
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 0},
	}
}

// NewClientWithHTTP creates a new Client with a custom http.Client.
func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		http:    httpClient,
	}
}

// BaseURL returns the base URL the client is configured for.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// do performs an HTTP request and unmarshals the JSON response into respData.
func (c *Client) do(ctx context.Context, method, path string, reqData, respData any) error {
	var reqBody io.Reader
	if reqData != nil {
		data, err := json.Marshal(reqData)
		if err != nil {
			return fmt.Errorf("lychee: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("lychee: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("lychee: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lychee: read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("lychee: HTTP %d: %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("lychee: HTTP %d: %s", resp.StatusCode, string(body))
	}

	if len(body) > 0 && respData != nil {
		if err := json.Unmarshal(body, respData); err != nil {
			return fmt.Errorf("lychee: unmarshal response: %w", err)
		}
	}
	return nil
}

// stream performs a streaming request, calling fn for each JSON line.
func (c *Client) stream(ctx context.Context, method, path string, reqData any, fn func([]byte) error) error {
	var buf io.Reader
	if reqData != nil {
		data, err := json.Marshal(reqData)
		if err != nil {
			return fmt.Errorf("lychee: marshal request: %w", err)
		}
		buf = bytes.NewBuffer(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return fmt.Errorf("lychee: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("lychee: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("lychee: HTTP %d: %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("lychee: HTTP %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanBuf := make([]byte, 0, 8*1024*1024) // 8 MB buffer
	scanner.Buffer(scanBuf, 8*1024*1024)
	for scanner.Scan() {
		if err := fn(scanner.Bytes()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// apiError represents an error returned by the Lychee API.
type apiError struct {
	Error string `json:"error"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// GenerateRequest describes a request for text generation.
type GenerateRequest struct {
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt"`
	Suffix   string          `json:"suffix,omitempty"`
	System   string          `json:"system,omitempty"`
	Template string          `json:"template,omitempty"`
	Context  []int           `json:"context,omitempty"`
	Stream   *bool           `json:"stream,omitempty"`
	Raw      bool            `json:"raw,omitempty"`
	Format   json.RawMessage `json:"format,omitempty"`
	Images   []string        `json:"images,omitempty"`
	Options  map[string]any  `json:"options,omitempty"`
}

// GenerateResponse is the response from a generate request.
type GenerateResponse struct {
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	Response   string    `json:"response"`
	Thinking   string    `json:"thinking,omitempty"`
	Done       bool      `json:"done"`
	DoneReason string    `json:"done_reason,omitempty"`
	Context    []int     `json:"context,omitempty"`

	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
}

// ChatRequest describes a request for chat completion.
type ChatRequest struct {
	Model    string          `json:"model"`
	Messages []Message       `json:"messages"`
	Stream   *bool           `json:"stream,omitempty"`
	Format   json.RawMessage `json:"format,omitempty"`
	Options  map[string]any  `json:"options,omitempty"`
}

// ChatResponse is the response from a chat request.
type ChatResponse struct {
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	Message    Message   `json:"message"`
	Done       bool      `json:"done"`
	DoneReason string    `json:"done_reason,omitempty"`

	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
}

// Message represents a single chat message.
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

// Model represents a locally available model.
type Model struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details,omitempty"`
}

// ModelDetails provides details about a model.
type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
	ContextLength     int      `json:"context_length,omitempty"`
	EmbeddingLength   int      `json:"embedding_length,omitempty"`
}

// EmbedRequest describes a request for embeddings.
type EmbedRequest struct {
	Model   string         `json:"model"`
	Input   any            `json:"input"`
	Options map[string]any `json:"options,omitempty"`
}

// EmbedResponse is the response from an embed request.
type EmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`

	TotalDuration   time.Duration `json:"total_duration,omitempty"`
	LoadDuration    time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount int           `json:"prompt_eval_count,omitempty"`
}

// VersionResponse is the response from the version endpoint.
type VersionResponse struct {
	Version string `json:"version"`
}

// PullProgress represents download progress for a model pull.
type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// listModelsResponse is the raw API response from /api/tags.
type listModelsResponse struct {
	Models []Model `json:"models"`
}

// ─────────────────────────────────────────────────────────────────────────────
// API Methods
// ─────────────────────────────────────────────────────────────────────────────

// Generate sends a generate request and returns the response.
// This is the non-streaming variant — it collects all chunks and returns
// the final complete response.
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("lychee: GenerateRequest is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("lychee: model is required")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("lychee: prompt is required")
	}

	// Force non-streaming.
	reqCopy := *req
	reqCopy.Stream = nil // setting to nil is equivalent to false for the API

	var last GenerateResponse
	err := c.stream(ctx, http.MethodPost, "/api/generate", &reqCopy, func(bts []byte) error {
		if err := json.Unmarshal(bts, &last); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &last, nil
}

// GenerateStream sends a generate request with streaming enabled.
// fn is called for each response chunk.
func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest, fn func(GenerateResponse) error) error {
	if req == nil {
		return fmt.Errorf("lychee: GenerateRequest is nil")
	}
	if req.Model == "" {
		return fmt.Errorf("lychee: model is required")
	}
	if req.Prompt == "" {
		return fmt.Errorf("lychee: prompt is required")
	}
	if fn == nil {
		return fmt.Errorf("lychee: callback function is nil")
	}

	reqCopy := *req
	stream := true
	reqCopy.Stream = &stream

	return c.stream(ctx, http.MethodPost, "/api/generate", &reqCopy, func(bts []byte) error {
		var resp GenerateResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}
		return fn(resp)
	})
}

// Chat sends a chat request and returns the response.
// This is the non-streaming variant — it collects all chunks and returns
// the final complete response.
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("lychee: ChatRequest is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("lychee: model is required")
	}

	reqCopy := *req
	reqCopy.Stream = nil

	var last ChatResponse
	err := c.stream(ctx, http.MethodPost, "/api/chat", &reqCopy, func(bts []byte) error {
		if err := json.Unmarshal(bts, &last); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &last, nil
}

// ChatStream sends a chat request with streaming enabled.
// fn is called for each response chunk.
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest, fn func(ChatResponse) error) error {
	if req == nil {
		return fmt.Errorf("lychee: ChatRequest is nil")
	}
	if req.Model == "" {
		return fmt.Errorf("lychee: model is required")
	}
	if fn == nil {
		return fmt.Errorf("lychee: callback function is nil")
	}

	reqCopy := *req
	stream := true
	reqCopy.Stream = &stream

	return c.stream(ctx, http.MethodPost, "/api/chat", &reqCopy, func(bts []byte) error {
		var resp ChatResponse
		if err := json.Unmarshal(bts, &resp); err != nil {
			return err
		}
		return fn(resp)
	})
}

// ListModels returns all locally available models.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	var resp listModelsResponse
	if err := c.do(ctx, http.MethodGet, "/api/tags", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

// PullModel pulls a model from HuggingFace.
// fn is called for each progress update during the download.
func (c *Client) PullModel(ctx context.Context, name string, fn func(PullProgress)) error {
	if name == "" {
		return fmt.Errorf("lychee: model name is required")
	}

	req := struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}{Model: name, Stream: true}

	return c.stream(ctx, http.MethodPost, "/api/pull", req, func(bts []byte) error {
		var progress PullProgress
		if err := json.Unmarshal(bts, &progress); err != nil {
			return err
		}
		if fn != nil {
			fn(progress)
		}
		return nil
	})
}

// Embeddings returns embedding vectors for the given input.
func (c *Client) Embeddings(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("lychee: EmbedRequest is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("lychee: model is required")
	}

	var resp EmbedResponse
	if err := c.do(ctx, http.MethodPost, "/api/embed", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteModel deletes a locally installed model.
func (c *Client) DeleteModel(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("lychee: model name is required")
	}

	req := struct {
		Model string `json:"model"`
	}{Model: name}

	return c.do(ctx, http.MethodDelete, "/api/delete", req, nil)
}

// Version returns the Lychee server version.
func (c *Client) Version(ctx context.Context) (*VersionResponse, error) {
	var resp VersionResponse
	if err := c.do(ctx, http.MethodGet, "/api/version", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Health checks if the Lychee server is reachable via HEAD on "/".
// Returns nil if healthy, error otherwise.
func (c *Client) Health(ctx context.Context) error {
	return c.do(ctx, http.MethodHead, "/", nil, nil)
}
