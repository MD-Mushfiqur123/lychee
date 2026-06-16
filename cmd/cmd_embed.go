package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/spf13/cobra"
)

// NewEmbedCmd creates the cmd command for lychee embed.
func NewEmbedCmd() *cobra.Command {
	var modelFlag string
	var outputFlag string
	var truncateFlag bool

	cmd := &cobra.Command{
		Use:   "embed [text...]",
		Short: "Generate embeddings for text",
		Long: `Generate vector embeddings for one or more inputs.

Examples:
  lychee embed "hello world" --model nomic-embed-text
  lychee embed "first" "second" "third" --model nomic-embed-text
  echo "some text" | lychee embed --model nomic-embed-text
  lychee embed "hello" --model nomic-embed-text --output json`,
		Args:    cobra.ArbitraryArgs,
		PreRunE: checkServerHeartbeat,
		RunE: func(cmd *cobra.Command, args []string) error {
			return embedHandler(cmd.Context(), args, modelFlag, outputFlag, truncateFlag)
		},
	}

	cmd.Flags().StringVarP(&modelFlag, "model", "m", "nomic-embed-text", "Model to use for embeddings")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "text", "Output format: text, json, or csv")
	cmd.Flags().BoolVar(&truncateFlag, "truncate", true, "Truncate input if it exceeds context length")
	return cmd
}

func embedHandler(ctx context.Context, inputs []string, model, outputFmt string, truncate bool) error {
	// Read from stdin if no args
	if len(inputs) == 0 {
		buf := new(strings.Builder)
		tmp := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(tmp)
			if n > 0 {
				buf.Write(tmp[:n])
			}
			if err != nil {
				break
			}
		}
		text := strings.TrimSpace(buf.String())
		if text == "" {
			return fmt.Errorf("no input provided — pass text as arguments or via stdin")
		}
		inputs = []string{text}
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("connecting to Lychee server: %w\nIs lychee running? Try: lychee serve", err)
	}

	req := &api.EmbedRequest{
		Model:    model,
		Input:    inputs,
		Truncate: &truncate,
	}

	resp, err := client.Embed(ctx, req)
	if err != nil {
		return fmt.Errorf("embedding failed: %w", err)
	}

	switch outputFmt {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"model":      resp.Model,
			"embeddings": resp.Embeddings,
		})
	case "csv":
		for i, emb := range resp.Embeddings {
			parts := make([]string, len(emb))
			for j, v := range emb {
				parts[j] = fmt.Sprintf("%.6f", v)
			}
			fmt.Printf("# input %d\n%s\n", i+1, strings.Join(parts, ","))
		}
	default: // text
		for i, emb := range resp.Embeddings {
			if len(inputs) > 1 {
				fmt.Printf("Input %d: %q\n", i+1, truncateStr(inputs[i], 60))
			}
			fmt.Printf("Dimensions: %d\n", len(emb))
			// Print first 8 values as preview
			preview := emb
			if len(preview) > 8 {
				preview = preview[:8]
			}
			parts := make([]string, len(preview))
			for j, v := range preview {
				parts[j] = fmt.Sprintf("%.4f", v)
			}
			suffix := ""
			if len(emb) > 8 {
				suffix = fmt.Sprintf(" ... (%d more)", len(emb)-8)
			}
			fmt.Printf("Vector:     [%s%s]\n", strings.Join(parts, ", "), suffix)
			if len(inputs) > 1 {
				fmt.Println()
			}
		}
		if resp.PromptEvalCount > 0 {
			fmt.Printf("Tokens: %d\n", resp.PromptEvalCount)
		}
	}
	return nil
}
