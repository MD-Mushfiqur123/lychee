package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MD-Mushfiqur123/lychee/api"
)

func NewGenerateClientCmd() *cobra.Command {
	var lang string
	var fileOut string

	cmd := &cobra.Command{
		Use:   "generate-client MODEL",
		Short: "Generate functional API client code snippets for a model",
		Long:  `Queries the running Lychee server for model metadata and outputs fully formed boilerplate code for integration.`,
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: autocompleteInstalledModels,
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			lang = strings.ToLower(lang)

			client, err := api.ClientFromEnvironment()
			if err != nil {
				return fmt.Errorf("failed to create API client: %v", err)
			}

			req := api.ShowRequest{Name: modelName}
			resp, err := client.Show(cmd.Context(), &req)
			if err != nil {
				return fmt.Errorf("failed to retrieve model %q details from server: %v", modelName, err)
			}

			// Extract context length
			contextLength := 0
			if resp.ModelInfo != nil {
				if arch, ok := resp.ModelInfo["general.architecture"].(string); ok && arch != "" {
					if v, ok := resp.ModelInfo[fmt.Sprintf("%s.context_length", arch)]; ok {
						if f, ok := v.(float64); ok {
							contextLength = int(f)
						}
					}
				}
				if contextLength == 0 {
					for _, key := range []string{"context_length", "general.context_length"} {
						if v, ok := resp.ModelInfo[key]; ok {
							if f, ok := v.(float64); ok {
								contextLength = int(f)
								break
							}
						}
					}
				}
			}

			contextLenStr := fmt.Sprintf("%d", contextLength)
			if contextLength <= 0 {
				contextLenStr = "not detected (using server default)"
			}

			// Formulate option blocks based on detected context length
			var pyOptions, jsOptions, rsOptions, goOptions string
			if contextLength > 0 {
				pyOptions = fmt.Sprintf(",\n    options={\n        \"num_ctx\": %d\n    }", contextLength)
				jsOptions = fmt.Sprintf(",\n  options: {\n    num_ctx: %d\n  }", contextLength)
				rsOptions = fmt.Sprintf("\n    let options = Options::default().num_ctx(%d);", contextLength)
				goOptions = fmt.Sprintf(",\n\t\tOptions: map[string]interface{}{\n\t\t\t\"num_ctx\": %d,\n\t\t}", contextLength)
			} else {
				pyOptions = "\n    # Note: could not detect context length from server\n    # options={\n    #     \"num_ctx\": 2048\n    # }"
				jsOptions = "\n  // Note: could not detect context length from server\n  // options: {\n  //   num_ctx: 2048\n  // }"
				rsOptions = "\n    // Note: could not detect context length from server\n    let options = Options::default();"
				goOptions = "\n\t\t// Note: could not detect context length from server\n\t\t// Options: map[string]interface{}{\n\t\t// \t\"num_ctx\": 2048,\n\t\t// }"
			}

			// Extract capabilities
			var caps []string
			for _, c := range resp.Capabilities {
				caps = append(caps, c.String())
			}
			capsStr := strings.Join(caps, ", ")
			if capsStr == "" {
				capsStr = "chat"
			}

			// Handle system prompt injection
			var pySystem, jsSystem, rsSystem, goSystem string
			if resp.System != "" {
				escapedPrompt := escapePrompt(resp.System)
				pySystem = fmt.Sprintf("        {\"role\": \"system\", \"content\": \"%s\"},\n", escapedPrompt)
				jsSystem = fmt.Sprintf("    { role: 'system', content: '%s' },\n", escapedPrompt)
				rsSystem = fmt.Sprintf("            Message {\n                role: \"system\".to_string(),\n                content: \"%s\".to_string(),\n            },\n", escapedPrompt)
				goSystem = fmt.Sprintf("\t\t\t{Role: \"system\", Content: \"%s\"},\n", escapedPrompt)
			}

			var outputBuilder strings.Builder

			switch lang {
			case "python", "py":
				outputBuilder.WriteString(fmt.Sprintf(`# Python Client Integration
# Install: pip install lychee-python
# Model Capabilities: %s
# Context Length: %s

import lychee

client = lychee.Client()
response = client.chat(
    model="%s",
    messages=[
%s        {"role": "user", "content": "Hello! Describe what you can do."}
    ]%s
)
print(response['message']['content'])
`, capsStr, contextLenStr, modelName, pySystem, pyOptions))

			case "javascript", "js", "node":
				outputBuilder.WriteString(fmt.Sprintf(`// JavaScript / Node.js Client Integration
// Install: npm install lychee-js (or pnpm add lychee-js / yarn add lychee-js)
// Model Capabilities: %s
// Context Length: %s

import lychee from 'lychee-js';

const response = await lychee.chat({
  model: '%s',
  messages: [
%s    { role: 'user', content: 'Hello! Describe what you can do.' }
  ]%s
});
console.log(response.message.content);
`, capsStr, contextLenStr, modelName, jsSystem, jsOptions))

			case "rust", "rs":
				outputBuilder.WriteString(fmt.Sprintf(`// Rust SDK Client Integration
// Cargo.toml:
// lychee-rs = { git = "https://github.com/MD-Mushfiqur123/lychee" }
// tokio = { version = "1.0", features = ["full"] }
// Model Capabilities: %s
// Context Length: %s

use lychee_rs::{Client, Message, Options};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new("http://localhost:11434".to_string());
    %s
    let response = client.chat(
        "%s".to_string(),
        vec![
%s            Message {
                role: "user".to_string(),
                content: "Hello! Describe what you can do.".to_string(),
            }
        ],
        options
    ).await?;
    
    println!("{}", response.message.content);
    Ok(())
}
`, capsStr, contextLenStr, rsOptions, modelName, rsSystem))

			case "go", "golang":
				outputBuilder.WriteString(fmt.Sprintf(`// Go Client Integration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	req := &api.ChatRequest{
		Model: "%s",
		Messages: []api.Message{
%s			{Role: "user", Content: "Hello! Describe what you can do."},
		}%s,
	}

	err = client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		return nil
	})
	if err != nil {
		log.Fatalf("chat failed: %v", err)
	}
	fmt.Println()
}
`, modelName, goSystem, goOptions))

			default:
				return fmt.Errorf("unsupported client language: %s (choose python, javascript, rust, or go)", lang)
			}

			generatedCode := outputBuilder.String()

			if fileOut != "" {
				err := os.WriteFile(fileOut, []byte(generatedCode), 0644)
				if err != nil {
					return fmt.Errorf("failed to write generated client code to %s: %v", fileOut, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "✅ Generated client code written to %s\n", fileOut)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), generatedCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "python", "Language for client integration (python, javascript, rust, go)")
	cmd.Flags().StringVarP(&fileOut, "file", "f", "", "Output file path (prints to stdout if empty)")

	return cmd
}

func escapePrompt(prompt string) string {
	escaped := strings.ReplaceAll(prompt, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	return escaped
}
