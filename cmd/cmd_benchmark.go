package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
)

// BenchmarkResult holds the metrics for a single model benchmark run.
type BenchmarkResult struct {
	Model            string
	TokensPerSec     float64
	TimeToFirstToken time.Duration
	TotalTime        time.Duration
	Tokens           int
	Error            string
}

type benchmarkFlags struct {
	models string
	runs   int
}

const defaultBenchmarkPrompt = "Explain the key differences between procedural and object-oriented programming in three concise paragraphs."

// benchmarkModel runs a single benchmark against one model with the given prompt
// and returns the measured metrics.
func benchmarkModel(client *api.Client, model, prompt string) BenchmarkResult {
	result := BenchmarkResult{Model: model}

	start := time.Now()
	var ttft time.Duration
	var ttftOnce sync.Once

	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		ttftOnce.Do(func() {
			if resp.Response != "" {
				ttft = time.Since(start)
			}
		})

		if resp.Done {
			result.TotalTime = resp.TotalDuration
			result.Tokens = resp.EvalCount
		}
		return nil
	})

	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.TimeToFirstToken = ttft
	if result.TotalTime.Seconds() > 0 && result.Tokens > 0 {
		result.TokensPerSec = float64(result.Tokens) / result.TotalTime.Seconds()
	}

	return result
}

// averageResults computes the average of multiple benchmark runs.
func averageResults(runs []BenchmarkResult) BenchmarkResult {
	if len(runs) == 0 {
		return BenchmarkResult{}
	}

	avg := BenchmarkResult{Model: runs[0].Model}
	var validRuns int

	for _, r := range runs {
		if r.Error != "" {
			avg.Error = r.Error
			continue
		}
		avg.TokensPerSec += r.TokensPerSec
		avg.TimeToFirstToken += r.TimeToFirstToken
		avg.TotalTime += r.TotalTime
		avg.Tokens += r.Tokens
		validRuns++
	}

	if validRuns == 0 {
		return avg
	}

	n := float64(validRuns)
	avg.TokensPerSec /= n
	avg.TimeToFirstToken = time.Duration(int64(avg.TimeToFirstToken) / int64(validRuns))
	avg.TotalTime = time.Duration(int64(avg.TotalTime) / int64(validRuns))
	avg.Tokens /= validRuns

	return avg
}

// formatBenchDuration formats a time.Duration with appropriate sub-second precision for benchmarks.
func formatBenchDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// runBenchmark executes the benchmark command logic.
func runBenchmark(cmd *cobra.Command, args []string) error {
	flags := benchmarkFlags{
		models: cmd.Flag("models").Value.String(),
		runs:   1,
	}

	if r, _ := cmd.Flags().GetInt("runs"); r > 0 {
		flags.runs = r
	}

	prompt := defaultBenchmarkPrompt
	if len(args) > 0 {
		prompt = args[0]
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("could not create lychee client: %w", err)
	}

	// Resolve models to benchmark
	var models []string
	if flags.models != "" {
		models = strings.Split(flags.models, ",")
		for i := range models {
			models[i] = strings.TrimSpace(models[i])
		}
	} else {
		// Use all installed models
		listResp, err := client.List(cmd.Context())
		if err != nil {
			return fmt.Errorf("could not list models: %w", err)
		}
		for _, m := range listResp.Models {
			models = append(models, m.Name)
		}
		sort.Strings(models)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models found to benchmark; pull models first or specify --models")
	}

	fmt.Fprintf(os.Stderr, "🔬 Benchmarking %d model(s) with %d run(s) each...\n", len(models), flags.runs)
	fmt.Fprintf(os.Stderr, "Prompt: %s\n\n", prompt)

	var allResults []BenchmarkResult

	for _, model := range models {
		fmt.Fprintf(os.Stderr, "⏳ Benchmarking %s...\n", model)

		var modelRuns []BenchmarkResult
		for i := range flags.runs {
			if flags.runs > 1 {
				fmt.Fprintf(os.Stderr, "  Run %d/%d...\n", i+1, flags.runs)
			}
			result := benchmarkModel(client, model, prompt)
			if result.Error != "" {
				fmt.Fprintf(os.Stderr, "  ⚠️  Error: %s\n", result.Error)
			}
			modelRuns = append(modelRuns, result)
		}

		avg := averageResults(modelRuns)
		allResults = append(allResults, avg)

		if avg.Error == "" {
			fmt.Fprintf(os.Stderr, "  ✅ %.1f tok/s | TTFT: %s | Total: %s | Tokens: %d\n\n",
				avg.TokensPerSec,
				formatBenchDuration(avg.TimeToFirstToken),
				formatBenchDuration(avg.TotalTime),
				avg.Tokens,
			)
		} else {
			fmt.Fprintf(os.Stderr, "  ❌ Failed\n\n")
		}
	}

	// Sort by tokens/sec descending for ranking
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].TokensPerSec > allResults[j].TokensPerSec
	})

	// Print comparison table
	printBenchmarkTable(allResults)

	return nil
}

func printBenchmarkTable(results []BenchmarkResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "MODEL", "TOK/S", "TTFT", "TOTAL TIME", "TOKENS", "STATUS"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("  ")

	medal := []string{"🥇", "🥈", "🥉"}

	for i, r := range results {
		rank := fmt.Sprintf("%d", i+1)
		if i < len(medal) {
			rank = medal[i]
		}

		tokPerSec := "-"
		ttft := "-"
		totalTime := "-"
		tokens := "-"
		status := "✅"

		if r.Error != "" {
			status = "❌ " + r.Error
		} else {
			tokPerSec = fmt.Sprintf("%.1f", r.TokensPerSec)
			ttft = formatBenchDuration(r.TimeToFirstToken)
			totalTime = formatBenchDuration(r.TotalTime)
			tokens = fmt.Sprintf("%d", r.Tokens)
		}

		table.Append([]string{rank, r.Model, tokPerSec, ttft, totalTime, tokens, status})
	}

	fmt.Println()
	table.Render()
	fmt.Println("\n💡 Higher TOK/S is better. TTFT = Time To First Token.")
}

// NewBenchmarkCmd creates the benchmark cobra command.
func NewBenchmarkCmd() *cobra.Command {
	benchmarkCmd := &cobra.Command{
		Use:   "benchmark [PROMPT]",
		Short: "Benchmark models against the same prompt and compare performance",
		Long: `Benchmark installed models by running the same prompt through each one and
measuring tokens/sec, time to first token, and total generation time.

If no models are specified with --models, all installed models are benchmarked.
Use --runs to average results over multiple runs for more stable measurements.

Examples:
  lychee benchmark "Explain quantum computing"
  lychee benchmark --models llama3.2:3b,phi-3:mini "What is AI?"
  lychee benchmark --runs 3 "Compare procedural vs OOP"`,
		Args: cobra.MaximumNArgs(1),
		RunE: runBenchmark,
	}

	benchmarkCmd.Flags().String("models", "", "Comma-separated list of models to benchmark (default: all installed)")
	benchmarkCmd.Flags().Int("runs", 1, "Number of runs per model for averaging")

	return benchmarkCmd
}
