package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/spf13/cobra"
)

// NewCompareCmd creates the cmd command for lychee compare.
func NewCompareCmd() *cobra.Command {
	var promptFlag string
	var systemFlag string
	var maxTokens int
	var sideBy bool

	cmd := &cobra.Command{
		Use:   "compare <model1> <model2>",
		Short: "Run two models on the same prompt and compare outputs",
		Long: `Runs the same prompt through two models simultaneously and displays
the outputs side by side for easy comparison.

Examples:
  lychee compare llama3.2:3b mistral:7b
  lychee compare llama3.1:8b qwen3:8b --prompt "Explain quantum computing"
  lychee compare model1 model2 --side-by-side`,
		Args:    cobra.ExactArgs(2),
		PreRunE: checkServerHeartbeat,
		RunE: func(cmd *cobra.Command, args []string) error {
			return compareHandler(cmd.Context(), args[0], args[1], promptFlag, systemFlag, maxTokens, sideBy)
		},
	}

	cmd.Flags().StringVarP(&promptFlag, "prompt", "p", "What are the three laws of robotics? Answer concisely.", "Prompt to send to both models")
	cmd.Flags().StringVar(&systemFlag, "system", "", "System message to use")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 512, "Max tokens per response")
	cmd.Flags().BoolVar(&sideBy, "side-by-side", false, "Show outputs side by side (requires wide terminal)")
	return cmd
}

type modelResponse struct {
	model    string
	response strings.Builder
	tokens   int
	duration time.Duration
	err      error
}

func compareHandler(ctx context.Context, model1, model2, prompt, system string, maxTokens int, sideBySide bool) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}

	fmt.Printf("\n  Prompt: %s\n\n", truncateStr(prompt, 80))
	fmt.Printf("  %-30s  %s\n", model1, model2)
	fmt.Println("  " + strings.Repeat("─", 70))

	r1 := &modelResponse{model: model1}
	r2 := &modelResponse{model: model2}

	var wg sync.WaitGroup
	wg.Add(2)

	runModel := func(r *modelResponse) {
		defer wg.Done()
		start := time.Now()
		_ = start // keeping compiler happy just in case
		msgs := []api.Message{{Role: "user", Content: prompt}}
		if system != "" {
			msgs = append([]api.Message{{Role: "system", Content: system}}, msgs...)
		}
		opts := map[string]any{"num_predict": maxTokens}
		req := &api.ChatRequest{
			Model:    r.model,
			Messages: msgs,
			Options:  opts,
		}
		r.err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
			r.response.WriteString(resp.Message.Content)
			if resp.Done {
				r.tokens = resp.EvalCount
				r.duration = resp.TotalDuration
			}
			return nil
		})
	}

	go runModel(r1)
	go runModel(r2)
	wg.Wait()

	// Print results
	fmt.Println()
	for _, r := range []*modelResponse{r1, r2} {
		header := fmt.Sprintf("  ── %s ", r.model)
		header += strings.Repeat("─", max(0, 70-len(header)+2))
		fmt.Println(header)
		if r.err != nil {
			fmt.Printf("  ERROR: %v\n", r.err)
		} else {
			// Word-wrap at 72 chars
			text := r.response.String()
			for _, line := range strings.Split(text, "\n") {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println()
			tps := float64(0)
			if r.duration > 0 {
				tps = float64(r.tokens) / r.duration.Seconds()
			}
			fmt.Printf("  %d tokens  %.1f tok/s  %s total\n",
				r.tokens, tps, r.duration.Round(time.Millisecond))
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("  ── Summary " + strings.Repeat("─", 59))
	fmt.Printf("  %-20s  %-12s  %-12s  %s\n", "Metric", model1, model2, "Winner")
	fmt.Println("  " + strings.Repeat("─", 70))

	tps1 := tokensPerSec(r1)
	tps2 := tokensPerSec(r2)
	speedWinner := "─"
	if tps1 > tps2*1.05 {
		speedWinner = model1 + " (faster)"
	} else if tps2 > tps1*1.05 {
		speedWinner = model2 + " (faster)"
	}
	fmt.Printf("  %-20s  %-12s  %-12s  %s\n", "Speed",
		fmt.Sprintf("%.1f tok/s", tps1),
		fmt.Sprintf("%.1f tok/s", tps2),
		speedWinner)

	lenWinner := "─"
	l1, l2 := r1.response.Len(), r2.response.Len()
	if l1 > l2 {
		lenWinner = model1 + " (longer)"
	} else if l2 > l1 {
		lenWinner = model2 + " (longer)"
	}
	fmt.Printf("  %-20s  %-12s  %-12s  %s\n", "Response length",
		fmt.Sprintf("%d chars", l1),
		fmt.Sprintf("%d chars", l2),
		lenWinner)

	fmt.Println()
	return nil
}

func tokensPerSec(r *modelResponse) float64 {
	if r.err != nil || r.duration == 0 {
		return 0
	}
	return float64(r.tokens) / r.duration.Seconds()
}
