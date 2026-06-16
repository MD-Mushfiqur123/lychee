package cmd

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/MD-Mushfiqur123/lychee/api"
)

type benchFlagOptions struct {
	models       string
	epochs       int
	maxTokens    int
	temperature  float64
	seed         int
	timeout      int
	prompt       string
	imageFile    string
	keepAlive    float64
	format       string
	outputFile   string
	debug        bool
	verbose      bool
	warmup       int
	promptTokens int
	numCtx       int
}

type benchMetrics struct {
	Model    string
	Step     string
	Count    int
	Duration time.Duration
}

type benchModelInfo struct {
	Name              string
	ParameterSize     string
	QuantizationLevel string
	Family            string
	SizeBytes         int64
	VRAMBytes         int64
	NumCtx            int64
}

const DefaultPrompt = `Please write a descriptive story about a llama named Alonso who grows up to be President of the Land of Llamas. Include details about Alonso's childhood, adolescent years, and how he grew up to be a political mover and shaker. Write the story with a sense of whimsy.`

var promptWordList = []string{
	"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
	"a", "bright", "sunny", "day", "in", "the", "meadow", "where",
	"flowers", "bloom", "and", "birds", "sing", "their", "morning",
	"songs", "while", "gentle", "breeze", "carries", "sweet", "scent",
	"of", "pine", "trees", "across", "rolling", "hills", "toward",
	"distant", "mountains", "covered", "with", "fresh", "snow",
	"beneath", "clear", "blue", "sky", "children", "play", "near",
	"old", "stone", "bridge", "that", "crosses", "winding", "river",
}

var tokensPerWord = 1.3

func generatePromptForTokenCount(targetTokens int, epoch int) string {
	targetWords := int(float64(targetTokens) / tokensPerWord)
	if targetWords < 1 {
		targetWords = 1
	}

	offset := epoch * 7
	n := len(promptWordList)
	words := make([]string, targetWords)
	for i := range words {
		words[i] = promptWordList[((i+offset)%n+n)%n]
	}
	return strings.Join(words, " ")
}

func calibratePromptTokens(targetTokens, actualTokens, wordCount int) {
	if actualTokens <= 0 || wordCount <= 0 {
		return
	}
	tokensPerWord = float64(actualTokens) / float64(wordCount)
	newWords := int(float64(targetTokens) / tokensPerWord)
	fmt.Fprintf(os.Stderr, "bench: calibrated %.2f tokens/word (target=%d, got=%d, words=%d → %d)\n",
		tokensPerWord, targetTokens, actualTokens, wordCount, newWords)
}

func buildGenerateRequest(model string, fOpt benchFlagOptions, imgData api.ImageData, epoch int) *api.GenerateRequest {
	options := make(map[string]interface{})
	if fOpt.maxTokens > 0 {
		options["num_predict"] = fOpt.maxTokens
	}
	options["temperature"] = fOpt.temperature
	if fOpt.seed > 0 {
		options["seed"] = fOpt.seed
	}
	if fOpt.numCtx > 0 {
		options["num_ctx"] = fOpt.numCtx
	}

	var keepAliveDuration *api.Duration
	if fOpt.keepAlive > 0 {
		duration := api.Duration{Duration: time.Duration(fOpt.keepAlive * float64(time.Second))}
		keepAliveDuration = &duration
	}

	prompt := fOpt.prompt
	if fOpt.promptTokens > 0 {
		prompt = generatePromptForTokenCount(fOpt.promptTokens, epoch)
	} else {
		prompt = fmt.Sprintf("[%d] %s", epoch, prompt)
	}

	req := &api.GenerateRequest{
		Model:     model,
		Prompt:    prompt,
		Raw:       true,
		Options:   options,
		KeepAlive: keepAliveDuration,
	}

	if imgData != nil {
		req.Images = []api.ImageData{imgData}
	}

	return req
}

func fetchModelInfo(ctx context.Context, client *api.Client, model string) benchModelInfo {
	info := benchModelInfo{Name: model}
	resp, err := client.Show(ctx, &api.ShowRequest{Model: model})
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Could not fetch model info for '%s': %v\n", model, err)
		return info
	}
	info.ParameterSize = resp.Details.ParameterSize
	info.QuantizationLevel = resp.Details.QuantizationLevel
	info.Family = resp.Details.Family
	return info
}

func fetchMemoryUsage(ctx context.Context, client *api.Client, model string) (size, vram int64) {
	resp, err := client.ListRunning(ctx)
	if err != nil {
		if debug := os.Getenv("LYCHEE_DEBUG"); debug != "" {
			fmt.Fprintf(os.Stderr, "WARNING: Could not fetch memory usage: %v\n", err)
		}
		return 0, 0
	}
	for _, m := range resp.Models {
		if m.Name == model || m.Model == model {
			return m.Size, m.SizeVRAM
		}
	}
	for _, m := range resp.Models {
		if strings.HasPrefix(m.Name, model) || strings.HasPrefix(m.Model, model) {
			return m.Size, m.SizeVRAM
		}
	}
	return 0, 0
}

func fetchContextLength(ctx context.Context, client *api.Client, model string) int64 {
	resp, err := client.ListRunning(ctx)
	if err != nil {
		return 0
	}
	for _, m := range resp.Models {
		if m.Name == model || m.Model == model || strings.HasPrefix(m.Name, model) || strings.HasPrefix(m.Model, model) {
			return int64(m.ContextLength)
		}
	}
	return 0
}

func outputFormatHeader(w io.Writer, format string, verbose bool) {
	switch format {
	case "benchstat":
		if verbose {
			fmt.Fprintf(w, "goos: %s\n", runtime.GOOS)
			fmt.Fprintf(w, "goarch: %s\n", runtime.GOARCH)
		}
	case "csv":
		headings := []string{"NAME", "STEP", "COUNT", "NS_PER_COUNT", "TOKEN_PER_SEC"}
		fmt.Fprintln(w, strings.Join(headings, ","))
	}
}

func outputModelInfo(w io.Writer, format string, info benchModelInfo) {
	params := cmp.Or(info.ParameterSize, "unknown")
	quant := cmp.Or(info.QuantizationLevel, "unknown")
	family := cmp.Or(info.Family, "unknown")

	memStr := ""
	if info.SizeBytes > 0 {
		memStr = fmt.Sprintf(" | Size: %d | VRAM: %d", info.SizeBytes, info.VRAMBytes)
	}
	ctxStr := ""
	if info.NumCtx > 0 {
		ctxStr = fmt.Sprintf(" | NumCtx: %d", info.NumCtx)
	}
	fmt.Fprintf(w, "# Model: %s | Params: %s | Quant: %s | Family: %s%s%s\n",
		info.Name, params, quant, family, memStr, ctxStr)
}

func OutputMetrics(w io.Writer, format string, metrics []benchMetrics, verbose bool) {
	switch format {
	case "benchstat":
		for _, m := range metrics {
			if m.Step == "generate" || m.Step == "prefill" {
				if m.Count > 0 {
					nsPerToken := float64(m.Duration.Nanoseconds()) / float64(m.Count)
					tokensPerSec := float64(m.Count) / (float64(m.Duration.Nanoseconds()) + 1e-12) * 1e9
					fmt.Fprintf(w, "BenchmarkModel/name=%s/step=%s 1 %.2f ns/token %.2f token/sec\n",
						m.Model, m.Step, nsPerToken, tokensPerSec)
				} else {
					fmt.Fprintf(w, "BenchmarkModel/name=%s/step=%s 1 0 ns/token 0 token/sec\n",
						m.Model, m.Step)
				}
			} else if m.Step == "ttft" {
				fmt.Fprintf(w, "BenchmarkModel/name=%s/step=ttft 1 %d ns/op\n",
					m.Model, m.Duration.Nanoseconds())
			} else {
				fmt.Fprintf(w, "BenchmarkModel/name=%s/step=%s 1 %d ns/op\n",
					m.Model, m.Step, m.Duration.Nanoseconds())
			}
		}
	case "csv":
		for _, m := range metrics {
			if m.Step == "generate" || m.Step == "prefill" {
				var nsPerToken float64
				var tokensPerSec float64
				if m.Count > 0 {
					nsPerToken = float64(m.Duration.Nanoseconds()) / float64(m.Count)
					tokensPerSec = float64(m.Count) / (float64(m.Duration.Nanoseconds()) + 1e-12) * 1e9
				}
				fmt.Fprintf(w, "%s,%s,%d,%.2f,%.2f\n", m.Model, m.Step, m.Count, nsPerToken, tokensPerSec)
			} else {
				fmt.Fprintf(w, "%s,%s,1,%d,0\n", m.Model, m.Step, m.Duration.Nanoseconds())
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown output format '%s'\n", format)
	}
}

func readImage(imageFile string) (api.ImageData, error) {
	data, err := os.ReadFile(imageFile)
	if err != nil {
		return nil, err
	}
	return api.ImageData(data), nil
}

func unloadModel(client *api.Client, model string, timeoutSeconds int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	_ = client.Generate(ctx, &api.GenerateRequest{
		Model:     model,
		KeepAlive: &api.Duration{Duration: 0},
	}, func(r api.GenerateResponse) error {
		return nil
	})
}

func BenchmarkModel(fOpt benchFlagOptions) error {
	models := strings.Split(fOpt.models, ",")

	var imgData api.ImageData
	var err error
	if fOpt.imageFile != "" {
		imgData, err = readImage(fOpt.imageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Couldn't read image '%s': %v\n", fOpt.imageFile, err)
			return err
		}
	}

	if fOpt.debug && imgData != nil {
		fmt.Fprintf(os.Stderr, "Read file '%s'\n", fOpt.imageFile)
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Couldn't create lychee client: %v\n", err)
		return err
	}

	var out io.Writer = os.Stdout
	if fOpt.outputFile != "" {
		f, err := os.OpenFile(fOpt.outputFile, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cannot open output file %s: %v\n", fOpt.outputFile, err)
			return err
		}
		defer f.Close()
		out = f
	}

	outputFormatHeader(out, fOpt.format, fOpt.verbose)

	if fOpt.debug && fOpt.promptTokens > 0 {
		prompt := generatePromptForTokenCount(fOpt.promptTokens, 0)
		wordCount := len(strings.Fields(prompt))
		fmt.Fprintf(os.Stderr, "Generated prompt targeting ~%d tokens (%d words, varied per epoch)\n", fOpt.promptTokens, wordCount)
	}

	for _, model := range models {
		infoCtx, infoCancel := context.WithTimeout(context.Background(), 10*time.Second)
		info := fetchModelInfo(infoCtx, client, model)
		infoCancel()

		for i := range fOpt.warmup {
			req := buildGenerateRequest(model, fOpt, imgData, -(i + 1))
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fOpt.timeout)*time.Second)

			var warmupMetrics *api.Metrics
			err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
				if resp.Done {
					warmupMetrics = &resp.Metrics
				}
				return nil
			})
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: Warmup %d/%d for %s failed: %v\n", i+1, fOpt.warmup, model, err)
			} else {
				if fOpt.debug {
					fmt.Fprintf(os.Stderr, "Warmup %d/%d for %s complete\n", i+1, fOpt.warmup, model)
				}
				if i == fOpt.warmup-1 && fOpt.promptTokens > 0 && warmupMetrics != nil {
					prompt := generatePromptForTokenCount(fOpt.promptTokens, -(i + 1))
					wordCount := len(strings.Fields(prompt))
					calibratePromptTokens(fOpt.promptTokens, warmupMetrics.PromptEvalCount, wordCount)
				}
			}
		}

		memCtx, memCancel := context.WithTimeout(context.Background(), 5*time.Second)
		info.SizeBytes, info.VRAMBytes = fetchMemoryUsage(memCtx, client, model)
		if fOpt.numCtx > 0 {
			info.NumCtx = int64(fOpt.numCtx)
		} else {
			info.NumCtx = fetchContextLength(memCtx, client, model)
		}
		memCancel()

		outputModelInfo(out, fOpt.format, info)

		shortCount := 0
		for epoch := range fOpt.epochs {
			var responseMetrics *api.Metrics
			var ttft time.Duration
			short := false

			const maxRetries = 3
			for attempt := range maxRetries + 1 {
				responseMetrics = nil
				ttft = 0
				var ttftOnce sync.Once

				req := buildGenerateRequest(model, fOpt, imgData, epoch+attempt*1000)
				requestStart := time.Now()

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fOpt.timeout)*time.Second)

				err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
					if fOpt.debug {
						fmt.Fprintf(os.Stderr, "%s", cmp.Or(resp.Thinking, resp.Response))
					}

					ttftOnce.Do(func() {
						if resp.Response != "" || resp.Thinking != "" {
							ttft = time.Since(requestStart)
						}
					})

					if resp.Done {
						responseMetrics = &resp.Metrics
					}
					return nil
				})
				cancel()

				if fOpt.debug {
					fmt.Fprintln(os.Stderr)
				}

				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						fmt.Fprintf(os.Stderr, "ERROR: Request timed out with model '%s' after %vs\n", model, fOpt.timeout)
					} else {
						fmt.Fprintf(os.Stderr, "ERROR: Couldn't generate with model '%s': %v\n", model, err)
					}
					break
				}

				if responseMetrics == nil {
					fmt.Fprintf(os.Stderr, "ERROR: No metrics received for model '%s'\n", model)
					break
				}

				short = fOpt.maxTokens > 0 && responseMetrics.EvalCount < fOpt.maxTokens
				if !short || attempt == maxRetries {
					break
				}

				if fOpt.debug {
					fmt.Fprintf(os.Stderr, "Short response (%d/%d tokens), retrying with different prompt (attempt %d/%d)\n",
						responseMetrics.EvalCount, fOpt.maxTokens, attempt+1, maxRetries)
				}
			}

			if err != nil || responseMetrics == nil {
				continue
			}

			if short {
				shortCount++
				if fOpt.debug {
					fmt.Fprintf(os.Stderr, "WARNING: Short response (%d/%d tokens) after %d retries for epoch %d\n",
						responseMetrics.EvalCount, fOpt.maxTokens, maxRetries, epoch+1)
				}
			}

			metrics := []benchMetrics{
				{
					Model:    model,
					Step:     "prefill",
					Count:    responseMetrics.PromptEvalCount,
					Duration: responseMetrics.PromptEvalDuration,
				},
				{
					Model:    model,
					Step:     "generate",
					Count:    responseMetrics.EvalCount,
					Duration: responseMetrics.EvalDuration,
				},
				{
					Model:    model,
					Step:     "ttft",
					Count:    1,
					Duration: ttft,
				},
				{
					Model:    model,
					Step:     "load",
					Count:    1,
					Duration: responseMetrics.LoadDuration,
				},
				{
					Model:    model,
					Step:     "total",
					Count:    1,
					Duration: responseMetrics.TotalDuration,
				},
			}

			OutputMetrics(out, fOpt.format, metrics, fOpt.verbose)

			if fOpt.debug && fOpt.promptTokens > 0 {
				fmt.Fprintf(os.Stderr, "Generated prompt targeting ~%d tokens (actual: %d)\n",
					fOpt.promptTokens, responseMetrics.PromptEvalCount)
			}

			if fOpt.keepAlive > 0 {
				time.Sleep(time.Duration(fOpt.keepAlive*float64(time.Second)) + 200*time.Millisecond)
			}
		}

		if shortCount > 0 {
			fmt.Fprintf(os.Stderr, "WARNING: %d/%d epochs for '%s' had short responses (<%d tokens). Generation metrics may be unreliable.\n",
				shortCount, fOpt.epochs, model, fOpt.maxTokens)
		}

		unloadModel(client, model, fOpt.timeout)
	}

	return nil
}

func NewBenchCmd() *cobra.Command {
	var opts benchFlagOptions

	benchCmd := &cobra.Command{
		Use:   "bench",
		Short: "Benchmark models",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.format != "benchstat" && opts.format != "csv" {
				return fmt.Errorf("invalid output format %q", opts.format)
			}
			if opts.models == "" {
				return fmt.Errorf("no model(s) specified to benchmark; use --model flag")
			}
			return BenchmarkModel(opts)
		},
	}

	benchCmd.Flags().StringVar(&opts.models, "model", "", "Model to benchmark")
	benchCmd.Flags().IntVar(&opts.epochs, "epochs", 6, "Number of epochs (iterations) per model")
	benchCmd.Flags().IntVar(&opts.maxTokens, "max-tokens", 200, "Maximum tokens for model response")
	benchCmd.Flags().Float64Var(&opts.temperature, "temperature", 0, "Temperature parameter")
	benchCmd.Flags().IntVar(&opts.seed, "seed", 0, "Random seed")
	benchCmd.Flags().IntVar(&opts.timeout, "timeout", 60*5, "Timeout in seconds")
	benchCmd.Flags().StringVarP(&opts.prompt, "prompt", "p", DefaultPrompt, "Prompt to use")
	benchCmd.Flags().StringVar(&opts.imageFile, "image", "", "Filename for an image to include")
	benchCmd.Flags().Float64VarP(&opts.keepAlive, "keepalive", "k", 0, "Keep alive duration in seconds")
	benchCmd.Flags().StringVar(&opts.format, "format", "benchstat", "Output format [benchstat|csv]")
	benchCmd.Flags().StringVarP(&opts.outputFile, "output", "o", "", "Output file for results (stdout if empty)")
	benchCmd.Flags().BoolVar(&opts.verbose, "verbose", false, "Show system information")
	benchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Show debug information")
	benchCmd.Flags().IntVar(&opts.warmup, "warmup", 1, "Number of warmup requests before timing")
	benchCmd.Flags().IntVar(&opts.promptTokens, "prompt-tokens", 0, "Generate prompt targeting ~N tokens (0 = use -p prompt)")
	benchCmd.Flags().IntVar(&opts.numCtx, "num-ctx", 0, "Context size (0 = server default)")

	return benchCmd
}
