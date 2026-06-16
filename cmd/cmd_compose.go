package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/spf13/cobra"
)

func NewComposeCmd() *cobra.Command {
	var stepsFile string
	var inputFlag string
	var streamFlag bool

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Run a model composition pipeline",
		Long: `Run multiple models in sequence or parallel as a composition pipeline.
Steps can be provided in a JSON file.

Example:
  lychee compose --steps pipeline.json --input "Hello World"`,
		Args: cobra.NoArgs,
		PreRunE: checkServerHeartbeat,
		RunE: func(cmd *cobra.Command, args []string) error {
			if stepsFile == "" {
				return fmt.Errorf("steps file path (--steps) is required")
			}

			data, err := os.ReadFile(stepsFile)
			if err != nil {
				return fmt.Errorf("failed to read steps file: %w", err)
			}

			var steps []api.ComposeStep
			if err := json.Unmarshal(data, &steps); err != nil {
				return fmt.Errorf("failed to parse steps JSON: %w", err)
			}

			client, err := api.ClientFromEnvironment()
			if err != nil {
				return err
			}

			req := &api.ComposeRequest{
				Input:  inputFlag,
				Steps:  steps,
				Stream: streamFlag,
			}

			fn := func(ev api.ComposeEvent) error {
				switch ev.Event {
				case "step_start":
					fmt.Printf("\n[Step %d] Executing model '%s'...\n", ev.Index+1, ev.Model)
				case "step_fallback":
					fmt.Printf("[Step %d] Fallback triggered to '%s': %s\n", ev.Index+1, ev.Model, ev.Text)
				case "step_progress":
					fmt.Print(ev.Text)
					os.Stdout.Sync()
				case "parallel_progress":
					// Format concurrently arriving chunks to stdout
					fmt.Printf("\n[Parallel %s]: %s", ev.Model, ev.Text)
				case "step_skipped":
					fmt.Printf("[Step %d] Skipped: %s\n", ev.Index+1, ev.Text)
				case "step_error":
					fmt.Printf("\n[Step %d] Warning (skipped on error): %s\n", ev.Index+1, ev.Text)
				case "step_complete":
					fmt.Printf("\n[Step %d] Complete. Output:\n%s\n", ev.Index+1, ev.Output)
				case "complete":
					fmt.Printf("\nPipeline completed successfully.\nFinal Output:\n%s\n", ev.Result.Output)
				}
				return nil
			}

			return client.Compose(cmd.Context(), req, fn)
		},
	}

	cmd.Flags().StringVarP(&stepsFile, "steps", "s", "", "Path to composition steps JSON file (required)")
	cmd.Flags().StringVarP(&inputFlag, "input", "i", "", "Initial input text for the pipeline")
	cmd.Flags().BoolVar(&streamFlag, "stream", true, "Stream step outputs dynamically")

	return cmd
}
