package api

// ComposeStep represents a single step in a model composition chain.
type ComposeStep struct {
	Model         string         `json:"model"`
	Prompt        string         `json:"prompt"`
	Options       map[string]any `json:"options,omitempty"`
	TimeoutSec    int            `json:"timeout_sec,omitempty"`     // Timeout threshold in seconds
	FallbackModel string         `json:"fallback_model,omitempty"` // Fallback model name on failure
	Parallel      []ComposeStep  `json:"parallel,omitempty"`       // Concurrent sibling steps
}

// ComposeRequest defines the request payload for model chaining.
type ComposeRequest struct {
	Input  string        `json:"input"`
	Steps  []ComposeStep `json:"steps"`
	Stream bool          `json:"stream,omitempty"`
}

// StepResult represents the output of a single composition step.
type StepResult struct {
	Model           string       `json:"model"`
	Output          string       `json:"output"`
	ParallelResults []StepResult `json:"parallel_results,omitempty"`
}

// ComposeResponse defines the response payload for model chaining.
type ComposeResponse struct {
	Output  string       `json:"output"`
	Results []StepResult `json:"results"`
}
