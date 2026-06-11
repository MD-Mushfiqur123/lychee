package api

// ComposeCondition defines a conditional branch in the composition DAG.
// If the condition matches the previous step output, the step is executed.
type ComposeCondition struct {
	// Contains checks if the previous step output contains this substring (case-insensitive)
	Contains string `json:"contains,omitempty"`
	// NotContains checks if the previous step output does NOT contain this substring
	NotContains string `json:"not_contains,omitempty"`
	// MinLength checks if the previous step output is at least N characters
	MinLength int `json:"min_length,omitempty"`
	// MaxLength checks if the previous step output is at most N characters
	MaxLength int `json:"max_length,omitempty"`
	// Always is true if this step should always run regardless of previous output
	Always bool `json:"always,omitempty"`
}

// ComposeStep represents a single step in a model composition chain.
type ComposeStep struct {
	Model         string            `json:"model"`
	Prompt        string            `json:"prompt"`
	Options       map[string]any    `json:"options,omitempty"`
	TimeoutSec    int               `json:"timeout_sec,omitempty"`    // Timeout threshold in seconds
	FallbackModel string            `json:"fallback_model,omitempty"` // Fallback model name on failure
	Parallel      []ComposeStep     `json:"parallel,omitempty"`       // Concurrent sibling steps
	Condition     *ComposeCondition `json:"condition,omitempty"`      // DAG: conditional execution
	SkipOnError   bool              `json:"skip_on_error,omitempty"`  // Continue chain if this step fails
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
	Skipped         bool         `json:"skipped,omitempty"`          // True if step was skipped by condition
	Error           string       `json:"error,omitempty"`            // Set if step failed but SkipOnError=true
	ParallelResults []StepResult `json:"parallel_results,omitempty"` // Results from parallel branches
}

// ComposeResponse defines the response payload for model chaining.
type ComposeResponse struct {
	Output  string       `json:"output"`
	Results []StepResult `json:"results"`
}
