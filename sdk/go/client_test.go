package lychee

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// NewClient tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNewClient_DefaultURL(t *testing.T) {
	c := NewClient("")
	if c.baseURL != "http://localhost:11434" {
		t.Errorf("expected http://localhost:11434, got %s", c.baseURL)
	}
}

func TestNewClient_CustomURL(t *testing.T) {
	c := NewClient("http://192.168.1.100:11434")
	if c.baseURL != "http://192.168.1.100:11434" {
		t.Errorf("expected http://192.168.1.100:11434, got %s", c.baseURL)
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:11434/")
	if c.baseURL != "http://localhost:11434" {
		t.Errorf("expected http://localhost:11434, got %s", c.baseURL)
	}
}

func TestBaseURL(t *testing.T) {
	c := NewClient("http://example.com:8080")
	if c.BaseURL() != "http://example.com:8080" {
		t.Errorf("expected http://example.com:8080, got %s", c.BaseURL())
	}
}

func TestNewClient_HTTPClientIsNotNil(t *testing.T) {
	c := NewClient("http://localhost:11434")
	if c.http == nil {
		t.Error("expected non-nil HTTP client")
	}
}

func TestNewClientWithHTTP(t *testing.T) {
	custom := &http.Client{}
	c := NewClientWithHTTP("http://localhost:9999", custom)
	if c.http != custom {
		t.Error("expected custom HTTP client to be used")
	}
	if c.baseURL != "http://localhost:9999" {
		t.Errorf("expected http://localhost:9999, got %s", c.baseURL)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GenerateRequest
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateRequest_RoundTrip(t *testing.T) {
	req := GenerateRequest{
		Model:  "gemma3",
		Prompt: "Hello",
	}
	rt(t, req)
}

func TestGenerateRequest_WithOptions(t *testing.T) {
	req := GenerateRequest{
		Model:   "gemma3",
		Prompt:  "Hello",
		Options: map[string]any{"temperature": 0.7, "top_p": 0.9},
	}
	rt(t, req)
}

func TestGenerateRequest_WithSystem(t *testing.T) {
	req := GenerateRequest{
		Model:  "gemma3",
		Prompt: "Hello",
		System: "You are a helpful assistant.",
	}
	rt(t, req)
}

// ─────────────────────────────────────────────────────────────────────────────
// ChatRequest
// ─────────────────────────────────────────────────────────────────────────────

func TestChatRequest_RoundTrip(t *testing.T) {
	req := ChatRequest{
		Model: "gemma3",
		Messages: []Message{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "What is AI?"},
		},
	}
	rt(t, req)
}

func TestChatRequest_WithOptions(t *testing.T) {
	req := ChatRequest{
		Model: "gemma3",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Options: map[string]any{"temperature": 0.5},
	}
	rt(t, req)
}

// ─────────────────────────────────────────────────────────────────────────────
// Message
// ─────────────────────────────────────────────────────────────────────────────

func TestMessage_RoundTrip(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "What is AI?",
	}
	rt(t, msg)
}

func TestMessage_WithImages(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Describe this image",
		Images:  []string{"base64encoded..."},
	}
	rt(t, msg)
}

// ─────────────────────────────────────────────────────────────────────────────
// EmbedRequest
// ─────────────────────────────────────────────────────────────────────────────

func TestEmbedRequest_RoundTrip(t *testing.T) {
	req := EmbedRequest{
		Model: "nomic-embed-text",
		Input: "Hello world",
	}
	rt(t, req)
}

func TestEmbedRequest_WithOptions(t *testing.T) {
	req := EmbedRequest{
		Model:   "nomic-embed-text",
		Input:   []string{"Hello", "World"},
		Options: map[string]any{"temperature": 0},
	}
	rt(t, req)
}

// ─────────────────────────────────────────────────────────────────────────────
// Response types
// ─────────────────────────────────────────────────────────────────────────────

func TestModel_JSON(t *testing.T) {
	jsonStr := `{"name":"gemma3:latest","model":"gemma3:latest","modified_at":"2024-01-01T00:00:00Z","size":1234567890,"digest":"abc123","details":{"parent_model":"","format":"gguf","family":"gemma","families":["gemma"],"parameter_size":"2B","quantization_level":"Q4_0"}}`
	m := unmarshalT[Model](t, jsonStr)

	if m.Name != "gemma3:latest" {
		t.Errorf("expected gemma3:latest, got %s", m.Name)
	}
	if m.Size != 1234567890 {
		t.Errorf("expected 1234567890, got %d", m.Size)
	}
	if m.Digest != "abc123" {
		t.Errorf("expected abc123, got %s", m.Digest)
	}
	if m.Details.Family != "gemma" {
		t.Errorf("expected gemma, got %s", m.Details.Family)
	}
	if m.Details.ParameterSize != "2B" {
		t.Errorf("expected 2B, got %s", m.Details.ParameterSize)
	}
}

func TestGenerateResponse_JSON(t *testing.T) {
	jsonStr := `{"model":"gemma3","created_at":"2024-01-01T00:00:00Z","response":"Hello!","done":true,"done_reason":"stop"}`
	r := unmarshalT[GenerateResponse](t, jsonStr)

	if r.Model != "gemma3" {
		t.Errorf("expected gemma3, got %s", r.Model)
	}
	if r.Response != "Hello!" {
		t.Errorf("expected Hello!, got %s", r.Response)
	}
	if !r.Done {
		t.Error("expected done to be true")
	}
	if r.DoneReason != "stop" {
		t.Errorf("expected stop, got %s", r.DoneReason)
	}
}

func TestChatResponse_JSON(t *testing.T) {
	jsonStr := `{"model":"gemma3","created_at":"2024-01-01T00:00:00Z","message":{"role":"assistant","content":"Hello, how can I help?"},"done":true}`
	r := unmarshalT[ChatResponse](t, jsonStr)

	if r.Message.Role != "assistant" {
		t.Errorf("expected assistant, got %s", r.Message.Role)
	}
	if r.Message.Content != "Hello, how can I help?" {
		t.Errorf("expected greeting, got %s", r.Message.Content)
	}
	if !r.Done {
		t.Error("expected done to be true")
	}
}

func TestVersionResponse_JSON(t *testing.T) {
	jsonStr := `{"version":"0.5.0"}`
	v := unmarshalT[VersionResponse](t, jsonStr)
	if v.Version != "0.5.0" {
		t.Errorf("expected 0.5.0, got %s", v.Version)
	}
}

func TestPullProgress_JSON(t *testing.T) {
	jsonStr := `{"status":"downloading","digest":"sha256:abc123","total":1000,"completed":500}`
	p := unmarshalT[PullProgress](t, jsonStr)

	if p.Status != "downloading" {
		t.Errorf("expected downloading, got %s", p.Status)
	}
	if p.Total != 1000 {
		t.Errorf("expected 1000, got %d", p.Total)
	}
	if p.Completed != 500 {
		t.Errorf("expected 500, got %d", p.Completed)
	}
}

func TestPullProgress_Success(t *testing.T) {
	jsonStr := `{"status":"success"}`
	p := unmarshalT[PullProgress](t, jsonStr)
	if p.Status != "success" {
		t.Errorf("expected success, got %s", p.Status)
	}
}

func TestListModelsResponse_JSON(t *testing.T) {
	jsonStr := `{"models":[{"name":"gemma3:latest","model":"gemma3:latest","modified_at":"2024-01-01T00:00:00Z","size":1234,"digest":"abc"}]}`
	lr := unmarshalT[listModelsResponse](t, jsonStr)

	if len(lr.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(lr.Models))
	}
	if lr.Models[0].Name != "gemma3:latest" {
		t.Errorf("expected gemma3:latest, got %s", lr.Models[0].Name)
	}
	if lr.Models[0].Size != 1234 {
		t.Errorf("expected 1234, got %d", lr.Models[0].Size)
	}
}

func TestEmbedResponse_JSON(t *testing.T) {
	jsonStr := `{"model":"nomic-embed-text","embeddings":[[0.1,0.2,0.3]]}`
	r := unmarshalT[EmbedResponse](t, jsonStr)

	if r.Model != "nomic-embed-text" {
		t.Errorf("expected nomic-embed-text, got %s", r.Model)
	}
	if len(r.Embeddings) != 1 {
		t.Fatalf("expected 1 embedding array, got %d", len(r.Embeddings))
	}
	if len(r.Embeddings[0]) != 3 {
		t.Fatalf("expected 3 values, got %d", len(r.Embeddings[0]))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Edge cases
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateRequest_EmptyJSON(t *testing.T) {
	// Minimal valid JSON
	jsonStr := `{"model":"m","prompt":"p"}`
	r := unmarshalT[GenerateRequest](t, jsonStr)
	if r.Model != "m" || r.Prompt != "p" {
		t.Error("failed to parse minimal generate request")
	}
}

func TestChatRequest_EmptyMessagesJSON(t *testing.T) {
	jsonStr := `{"model":"m","messages":[]}`
	r := unmarshalT[ChatRequest](t, jsonStr)
	if r.Model != "m" {
		t.Errorf("expected m, got %s", r.Model)
	}
	if len(r.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(r.Messages))
	}
}

func TestModel_EmptyDetailsJSON(t *testing.T) {
	jsonStr := `{"name":"test","model":"test","modified_at":"2024-01-01T00:00:00Z","size":0,"digest":"x"}`
	m := unmarshalT[Model](t, jsonStr)
	if m.Name != "test" {
		t.Errorf("expected test, got %s", m.Name)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// rt performs a JSON marshal/unmarshal round-trip test.
func rt[T any](t *testing.T, v T) {
	t.Helper()
	data, err := jsonMarshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	got, err := unmarshalTFromBytes[T](t, data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	// Compare by re-marshaling both
	reData, _ := jsonMarshal(got)
	origData, _ := jsonMarshal(v)
	if !bytes.Equal(reData, origData) {
		t.Errorf("round-trip mismatch\n  original: %s\n  got:      %s", string(origData), string(reData))
	}
}

func unmarshalT[T any](t *testing.T, jsonStr string) T {
	t.Helper()
	v, err := unmarshalTFromBytes[T](t, []byte(jsonStr))
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	return v
}

func unmarshalTFromBytes[T any](_ *testing.T, data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}

func jsonMarshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	// Trim trailing newline added by Encode
	return bytes.TrimRight(buf.Bytes(), "\n"), err
}

