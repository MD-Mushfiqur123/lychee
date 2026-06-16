package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/llm"
	"github.com/MD-Mushfiqur123/lychee/progress"
)

type QuantizeProgress struct {
	mu      sync.Mutex
	message string
	current int
	total   int
}

func (q *QuantizeProgress) Set(current, total int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = current
	q.total = total
}

func (q *QuantizeProgress) String() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.total <= 0 {
		return q.message
	}
	percent := float64(q.current) / float64(q.total) * 100
	const barWidth = 30
	completedWidth := int(float64(barWidth) * float64(q.current) / float64(q.total))
	remainingWidth := barWidth - completedWidth
	if completedWidth < 0 {
		completedWidth = 0
	}
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	return fmt.Sprintf("%-25s %3.0f%% ▕%s%s▏ %d/%d",
		q.message, percent,
		strings.Repeat("█", completedWidth), strings.Repeat(" ", remainingWidth),
		q.current, q.total)
}

func NewQuantizeCmd() *cobra.Command {
	quantizeCmd := &cobra.Command{
		Use:   "quantize INPUT OUTPUT TYPE",
		Short: "Quantize a local GGUF model file",
		Long:  `Quantize a local GGUF model file to a specified format (e.g. Q4_K_M, Q8_0, F16).`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]
			outputPath := args[1]
			quantType := args[2]

			// Verify input file exists
			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("input model file not found: %w", err)
			}

			// Locate llama-quantize executable
			quantizeExe, err := llm.FindLlamaCppBinary("llama-quantize")
			if err != nil {
				return fmt.Errorf("llama-quantize binary not found: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Using quantize binary: %s\n", quantizeExe)
			fmt.Fprintf(cmd.OutOrStdout(), "Quantizing %s -> %s (%s)\n", inputPath, outputPath, quantType)

			// Setup execution command
			execCmd := exec.Command(quantizeExe, "--allow-requantize", inputPath, outputPath, quantType)
			
			// Setup progress bar
			p := progress.NewProgress(cmd.ErrOrStderr())
			progState := &QuantizeProgress{message: "Quantizing tensors"}
			p.Add("quantize", progState)
			defer p.Stop()

			// Pipe stdout to scan progress
			stdout, err := execCmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create stdout pipe: %w", err)
			}
			execCmd.Stderr = os.Stderr

			if err := execCmd.Start(); err != nil {
				return fmt.Errorf("failed to start llama-quantize: %w", err)
			}

			// Matches progress output like "[ 42/ 200]"
			progressRegex := regexp.MustCompile(`\[\s*(\d+)/\s*(\d+)\]`)

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if matches := progressRegex.FindStringSubmatch(line); len(matches) == 3 {
					current, _ := strconv.Atoi(matches[1])
					total, _ := strconv.Atoi(matches[2])
					progState.Set(current, total)
				}
			}

			if err := execCmd.Wait(); err != nil {
				return fmt.Errorf("quantization failed: %w", err)
			}

			progState.Set(100, 100) // Ensure 100% at finish
			return nil
		},
	}

	return quantizeCmd
}
