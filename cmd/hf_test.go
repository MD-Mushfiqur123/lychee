package cmd

import (
	"testing"
)

func TestExtractQuantKey(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"Q4_K_M suffix", "Meta-Llama-3-8B-Instruct-Q4_K_M.gguf", "q4_k_m"},
		{"Q8_0 suffix", "phi-3-mini-Q8_0.gguf", "q8_0"},
		{"f16 suffix", "gemma-2b-f16.gguf", "f16"},
		{"sharded format", "model-Q5_K_M-00001-of-00002.gguf", "q5_k_m"},
		{"unknown format", "my_custom_model.gguf", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractQuantKey(tt.fileName)
			if got != tt.expected {
				t.Errorf("extractQuantKey(%q) = %q; want %q", tt.fileName, got, tt.expected)
			}
		})
	}
}

func TestGroupByQuant(t *testing.T) {
	files := []hfFile{
		{Name: "model-Q4_K_M.gguf", quantKey: "q4_k_m", Size: 100},
		{Name: "model-Q4_K_M-00001-of-00002.gguf", quantKey: "q4_k_m", Size: 50},
		{Name: "model-Q8_0.gguf", quantKey: "q8_0", Size: 200},
	}

	grouped := groupByQuant(files)

	if len(grouped) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(grouped))
	}

	if len(grouped["q4_k_m"]) != 2 {
		t.Errorf("expected 2 files in q4_k_m group, got %d", len(grouped["q4_k_m"]))
	}

	if len(grouped["q8_0"]) != 1 {
		t.Errorf("expected 1 file in q8_0 group, got %d", len(grouped["q8_0"]))
	}
}

func TestSelectBestQuant(t *testing.T) {
	quants := map[string][]hfFile{
		"q8_0":   {{Name: "model-Q8_0.gguf", Size: 200}},
		"q4_k_m": {{Name: "model-Q4_K_M.gguf", Size: 100}},
		"other":  {{Name: "model.gguf", Size: 150}},
	}

	// Should select recommended q4_k_m
	selected := selectBestQuant(quants)
	if len(selected) != 1 || selected[0].Name != "model-Q4_K_M.gguf" {
		t.Errorf("expected selection of model-Q4_K_M.gguf, got %+v", selected)
	}

	// Should select q8_0 if q4_k_m is missing
	delete(quants, "q4_k_m")
	selected = selectBestQuant(quants)
	if len(selected) != 1 || selected[0].Name != "model-Q8_0.gguf" {
		t.Errorf("expected selection of model-Q8_0.gguf, got %+v", selected)
	}
}
