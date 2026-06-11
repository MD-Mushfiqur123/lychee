package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateJSONSchema(t *testing.T) {
	t.Run("nil or empty schema", func(t *testing.T) {
		err := ValidateJSONSchema(`{"any":"thing"}`, nil)
		if err != nil {
			t.Errorf("expected nil error for nil schema, got %v", err)
		}

		err = ValidateJSONSchema(`{"any":"thing"}`, json.RawMessage(""))
		if err != nil {
			t.Errorf("expected nil error for empty schema, got %v", err)
		}
	})

	t.Run("literal json format formatStr", func(t *testing.T) {
		schema := json.RawMessage(`"json"`)

		err := ValidateJSONSchema(`{"valid": true}`, schema)
		if err != nil {
			t.Errorf("expected valid JSON to pass, got %v", err)
		}

		err = ValidateJSONSchema(`{"invalid": `, schema)
		if err == nil {
			t.Errorf("expected invalid JSON to fail")
		} else if !strings.Contains(err.Error(), "not valid JSON") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("basic types validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "integer"},
				"score": {"type": "number"},
				"active": {"type": "boolean"}
			},
			"required": ["name", "age"]
		}`)

		// Correct input
		err := ValidateJSONSchema(`{"name": "Alice", "age": 30, "score": 95.5, "active": true}`, schema)
		if err != nil {
			t.Errorf("expected valid schema to pass, got %v", err)
		}

		// Missing required property
		err = ValidateJSONSchema(`{"name": "Alice", "score": 95.5}`, schema)
		if err == nil {
			t.Errorf("expected failure for missing required property")
		} else if !strings.Contains(err.Error(), "missing required property") {
			t.Errorf("unexpected error: %v", err)
		}

		// Incorrect type (string instead of integer)
		err = ValidateJSONSchema(`{"name": "Alice", "age": "thirty"}`, schema)
		if err == nil {
			t.Errorf("expected failure for wrong type")
		} else if !strings.Contains(err.Error(), "expected integer") {
			t.Errorf("unexpected error: %v", err)
		}

		// Incorrect integer format (float value)
		err = ValidateJSONSchema(`{"name": "Alice", "age": 30.5}`, schema)
		if err == nil {
			t.Errorf("expected failure for non-integer float")
		} else if !strings.Contains(err.Error(), "expected integer") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("array validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "array",
			"items": {
				"type": "string"
			}
		}`)

		err := ValidateJSONSchema(`["one", "two", "three"]`, schema)
		if err != nil {
			t.Errorf("expected valid array to pass, got %v", err)
		}

		err = ValidateJSONSchema(`["one", 2, "three"]`, schema)
		if err == nil {
			t.Errorf("expected array validation failure on element type mismatch")
		} else if !strings.Contains(err.Error(), "expected string") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("enum validation", func(t *testing.T) {
		schema := json.RawMessage(`{
			"type": "string",
			"enum": ["red", "green", "blue"]
		}`)

		err := ValidateJSONSchema(`"green"`, schema)
		if err != nil {
			t.Errorf("expected enum match to pass, got %v", err)
		}

		err = ValidateJSONSchema(`"yellow"`, schema)
		if err == nil {
			t.Errorf("expected enum mismatch to fail")
		} else if !strings.Contains(err.Error(), "not in enum") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nesting depth limit validation", func(t *testing.T) {
		// Generate a very deeply nested JSON object and matching schema to trigger depth limit
		var buildDeepSchema func(d int) string
		buildDeepSchema = func(d int) string {
			if d <= 0 {
				return `{"type": "string"}`
			}
			return `{"type": "object", "properties": {"next": ` + buildDeepSchema(d-1) + `}}`
		}

		schema := json.RawMessage(buildDeepSchema(35))

		// Construct corresponding deep JSON input
		var buildDeepJSON func(d int) string
		buildDeepJSON = func(d int) string {
			if d <= 0 {
				return `"end"`
			}
			return `{"next": ` + buildDeepJSON(d-1) + `}`
		}
		jsonData := buildDeepJSON(35)

		err := ValidateJSONSchema(jsonData, schema)
		if err == nil {
			t.Errorf("expected deep nesting depth limit to trigger failure")
		} else if !strings.Contains(err.Error(), "nesting depth limit exceeded") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
