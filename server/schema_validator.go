package server

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// ValidateJSONSchema validates a JSON string or raw byte slice against a JSON Schema.
func ValidateJSONSchema(output string, schemaRaw json.RawMessage) error {
	if len(schemaRaw) == 0 {
		return nil
	}

	// If the format is just the string "json", verify it's valid JSON
	var formatStr string
	if err := json.Unmarshal(schemaRaw, &formatStr); err == nil && formatStr == "json" {
		if !json.Valid([]byte(output)) {
			return fmt.Errorf("output is not valid JSON")
		}
		return nil
	}

	// Otherwise, it must be a JSON Schema object
	var schema map[string]any
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return fmt.Errorf("invalid JSON schema definition: %w", err)
	}

	var data any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return fmt.Errorf("failed to parse output as JSON: %w", err)
	}

	return validateValue(data, schema, "", 0, schema)
}

func resolveRef(ref string, root map[string]any) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("external or relative refs not supported: %s", ref)
	}

	parts := strings.Split(ref[2:], "/")
	var current any = root

	for _, part := range parts {
		// Escape JSON pointer characters
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		switch c := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = c[part]
			if !ok {
				return nil, fmt.Errorf("reference %q: property %q not found", ref, part)
			}
		default:
			return nil, fmt.Errorf("reference %q: cannot traverse non-object type %T at %q", ref, current, part)
		}
	}

	resolved, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("reference %q did not resolve to a valid schema object", ref)
	}

	return resolved, nil
}

func validateValue(val any, schema map[string]any, path string, depth int, root map[string]any) error {
	if depth > 32 {
		return fmt.Errorf("schema nesting depth limit exceeded")
	}

	if schema == nil {
		return nil
	}

	// Handle $ref
	if refVal, ok := schema["$ref"].(string); ok {
		resolvedSchema, err := resolveRef(refVal, root)
		if err != nil {
			return err
		}
		return validateValue(val, resolvedSchema, path, depth+1, root)
	}

	// Validate type
	if typeVal, ok := schema["type"].(string); ok {
		switch typeVal {
		case "object":
			obj, ok := val.(map[string]any)
			if !ok {
				return fmt.Errorf("path %q: expected object, got %T", path, val)
			}
			// Validate required properties
			if reqs, ok := schema["required"].([]any); ok {
				for _, r := range reqs {
					reqKey, ok := r.(string)
					if ok {
						if _, exists := obj[reqKey]; !exists {
							return fmt.Errorf("path %q: missing required property %q", path, reqKey)
						}
					}
				}
			}
			// Validate properties
			if props, ok := schema["properties"].(map[string]any); ok {
				for propKey, propSchemaVal := range props {
					propSchema, ok := propSchemaVal.(map[string]any)
					if ok {
						if propVal, exists := obj[propKey]; exists {
							newPath := propKey
							if path != "" {
								newPath = path + "." + propKey
							}
							if err := validateValue(propVal, propSchema, newPath, depth+1, root); err != nil {
								return err
							}
						}
					}
				}
			}
			// Validate additionalProperties
			if addProps, exists := schema["additionalProperties"]; exists {
				if addPropsBool, ok := addProps.(bool); ok && !addPropsBool {
					allowedProps := make(map[string]bool)
					if props, ok := schema["properties"].(map[string]any); ok {
						for k := range props {
							allowedProps[k] = true
						}
					}
					for k := range obj {
						if !allowedProps[k] {
							return fmt.Errorf("path %q: additional property %q is not allowed", path, k)
						}
					}
				}
			}
		case "array":
			arr, ok := val.([]any)
			if !ok {
				return fmt.Errorf("path %q: expected array, got %T", path, val)
			}
			if itemsSchemaVal, ok := schema["items"].(map[string]any); ok {
				for i, item := range arr {
					newPath := fmt.Sprintf("%s[%d]", path, i)
					if err := validateValue(item, itemsSchemaVal, newPath, depth+1, root); err != nil {
						return err
					}
				}
			}
		case "string":
			sVal, ok := val.(string)
			if !ok {
				return fmt.Errorf("path %q: expected string, got %T", path, val)
			}
			if minLength, ok := schema["minLength"].(float64); ok && float64(len(sVal)) < minLength {
				return fmt.Errorf("path %q: string length %d is less than minLength %g", path, len(sVal), minLength)
			}
			if maxLength, ok := schema["maxLength"].(float64); ok && float64(len(sVal)) > maxLength {
				return fmt.Errorf("path %q: string length %d is greater than maxLength %g", path, len(sVal), maxLength)
			}
			if patternStr, ok := schema["pattern"].(string); ok {
				matched, err := regexp.MatchString(patternStr, sVal)
				if err != nil {
					return fmt.Errorf("path %q: invalid pattern regex %q: %w", path, patternStr, err)
				}
				if !matched {
					return fmt.Errorf("path %q: string %q does not match pattern %q", path, sVal, patternStr)
				}
			}
		case "number", "integer":
			f, ok := val.(float64)
			if !ok {
				return fmt.Errorf("path %q: expected number, got %T", path, val)
			}
			if typeVal == "integer" && f != float64(int64(f)) {
				return fmt.Errorf("path %q: expected integer, got float %f", path, f)
			}
			if minimum, ok := schema["minimum"].(float64); ok && f < minimum {
				return fmt.Errorf("path %q: value %g is less than minimum %g", path, f, minimum)
			}
			if maximum, ok := schema["maximum"].(float64); ok && f > maximum {
				return fmt.Errorf("path %q: value %g is greater than maximum %g", path, f, maximum)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("path %q: expected boolean, got %T", path, val)
			}
		}
	}

	// Validate enum
	if enumVals, ok := schema["enum"].([]any); ok {
		matched := false
		for _, e := range enumVals {
			if reflect.DeepEqual(e, val) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("path %q: value %v not in enum %v", path, val, enumVals)
		}
	}

	// Validate allOf
	if allOf, ok := schema["allOf"].([]any); ok {
		for i, subSchemaVal := range allOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid allOf schema at index %d", i)
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err != nil {
				return fmt.Errorf("allOf schema validation failed at index %d: %w", i, err)
			}
		}
	}

	// Validate anyOf
	if anyOf, ok := schema["anyOf"].([]any); ok {
		matched := false
		var lastErr error
		for _, subSchemaVal := range anyOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				continue
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err == nil {
				matched = true
				break
			} else {
				lastErr = err
			}
		}
		if !matched {
			return fmt.Errorf("anyOf schema validation failed: value does not match any allowed schema (last error: %v)", lastErr)
		}
	}

	// Validate oneOf
	if oneOf, ok := schema["oneOf"].([]any); ok {
		matches := 0
		var lastErr error
		for _, subSchemaVal := range oneOf {
			subSchema, ok := subSchemaVal.(map[string]any)
			if !ok {
				continue
			}
			if err := validateValue(val, subSchema, path, depth+1, root); err == nil {
				matches++
			} else {
				lastErr = err
			}
		}
		if matches != 1 {
			return fmt.Errorf("oneOf schema validation failed: expected exactly 1 match, found %d matches (last error: %v)", matches, lastErr)
		}
	}

	// Validate not
	if notSchemaVal, ok := schema["not"]; ok {
		if notSchema, ok := notSchemaVal.(map[string]any); ok {
			if err := validateValue(val, notSchema, path, depth+1, root); err == nil {
				return fmt.Errorf("path %q: value must NOT match the 'not' schema", path)
			}
		}
	}

	// Validate const
	if constVal, ok := schema["const"]; ok {
		if !reflect.DeepEqual(constVal, val) {
			return fmt.Errorf("path %q: value must equal const %v, got %v", path, constVal, val)
		}
	}

	// Validate format (string only)
	if sVal, ok := val.(string); ok {
		if formatStr, ok := schema["format"].(string); ok {
			if err := validateFormat(sVal, formatStr, path); err != nil {
				return err
			}
		}
	}

	// Validate if/then/else
	if ifSchemaVal, ok := schema["if"]; ok {
		if ifSchema, ok := ifSchemaVal.(map[string]any); ok {
			ifPasses := validateValue(val, ifSchema, path, depth+1, root) == nil
			if ifPasses {
				if thenSchemaVal, ok := schema["then"]; ok {
					if thenSchema, ok := thenSchemaVal.(map[string]any); ok {
						if err := validateValue(val, thenSchema, path, depth+1, root); err != nil {
							return fmt.Errorf("path %q: 'then' schema failed: %w", path, err)
						}
					}
				}
			} else {
				if elseSchemaVal, ok := schema["else"]; ok {
					if elseSchema, ok := elseSchemaVal.(map[string]any); ok {
						if err := validateValue(val, elseSchema, path, depth+1, root); err != nil {
							return fmt.Errorf("path %q: 'else' schema failed: %w", path, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// validateFormat checks a string value against a JSON Schema format keyword.
func validateFormat(s, format, path string) error {
	switch format {
	case "email":
		if _, err := mail.ParseAddress(s); err != nil {
			return fmt.Errorf("path %q: value %q is not a valid email address", path, s)
		}
	case "uri", "url":
		u, err := url.ParseRequestURI(s)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("path %q: value %q is not a valid URI", path, s)
		}
	case "date-time":
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
		}
		parsed := false
		for _, f := range formats {
			if _, err := time.Parse(f, s); err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return fmt.Errorf("path %q: value %q is not a valid date-time (RFC3339)", path, s)
		}
	case "date":
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("path %q: value %q is not a valid date (YYYY-MM-DD)", path, s)
		}
	case "time":
		if _, err := time.Parse("15:04:05", s); err != nil {
			if _, err2 := time.Parse("15:04:05Z07:00", s); err2 != nil {
				return fmt.Errorf("path %q: value %q is not a valid time (HH:MM:SS)", path, s)
			}
		}
	case "uuid":
		uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		if !uuidPattern.MatchString(s) {
			return fmt.Errorf("path %q: value %q is not a valid UUID", path, s)
		}
	case "ipv4":
		ipv4Pattern := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
		if !ipv4Pattern.MatchString(s) {
			return fmt.Errorf("path %q: value %q is not a valid IPv4 address", path, s)
		}
		parts := strings.Split(s, ".")
		for _, part := range parts {
			var n int
			fmt.Sscanf(part, "%d", &n)
			if n < 0 || n > 255 {
				return fmt.Errorf("path %q: value %q is not a valid IPv4 address", path, s)
			}
		}
	// Silently ignore unknown formats (per JSON Schema spec)
	}
	return nil
}
