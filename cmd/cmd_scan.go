package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/spf13/cobra"
)

// NewScanCmd creates the cmd command for lychee scan.
func NewScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan system hardware and recommend models",
		Long: `Detects your GPU, VRAM, and RAM then recommends the best models
that will fit and run well on your hardware.

Examples:
  lychee scan
  lychee scan --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, _ := cmd.Flags().GetBool("json")
			return scanHandler(cmd.Context(), jsonOut)
		},
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

type systemInfo struct {
	OS       string    `json:"os"`
	Arch     string    `json:"arch"`
	CPUS     int       `json:"cpus"`
	RAMBytes int64     `json:"ram_bytes"`
	GPUs     []gpuInfo `json:"gpus,omitempty"`
}

type gpuInfo struct {
	Name      string `json:"name"`
	VRAMBytes int64  `json:"vram_bytes"`
}

func scanHandler(ctx context.Context, jsonOut bool) error {
	fmt.Println("  Scanning system hardware...")
	fmt.Println()

	// Get basic system info
	info := systemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		CPUS: runtime.NumCPU(),
	}

	// Try to get RAM
	info.RAMBytes = getSystemRAM()

	// Try to get GPU info from server
	var totalVRAM int64
	client, serverErr := api.ClientFromEnvironment()
	if serverErr == nil {
		psCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if running, err := client.ListRunning(psCtx); err == nil {
			for _, m := range running.Models {
				if m.SizeVRAM > 0 {
					totalVRAM += m.SizeVRAM
				}
			}
			if totalVRAM > 0 {
				info.GPUs = append(info.GPUs, gpuInfo{
					Name:      "Active GPU Memory",
					VRAMBytes: totalVRAM,
				})
			}
		}
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	// Print system summary
	fmt.Printf("  %-18s %s/%s (%d CPUs)\n", "Platform:", info.OS, info.Arch, info.CPUS)
	if info.RAMBytes > 0 {
		fmt.Printf("  %-18s %s\n", "System RAM:", formatBytes(info.RAMBytes))
	}
	if totalVRAM > 0 {
		fmt.Printf("  %-18s %s\n", "GPU VRAM:", formatBytes(totalVRAM))
	}
	fmt.Println()

	// Recommend models based on available memory
	availMem := totalVRAM
	if availMem == 0 {
		availMem = info.RAMBytes / 2 // Use half RAM for CPU inference estimate
	}

	fmt.Println("  Recommended models for your hardware:")
	fmt.Println()
	fmt.Printf("  %-28s %-10s %-8s %s\n", "MODEL", "SIZE", "FITS", "NOTES")
	fmt.Println("  " + strings.Repeat("─", 75))

	type rec struct {
		name  string
		sizeB int64
		notes string
		repo  string
	}

	recommendations := []rec{
		{"qwen3:0.6b", 400 * 1024 * 1024, "any hardware, fastest", "bartowski/Qwen3-0.6B-GGUF"},
		{"llama3.2:1b", 800 * 1024 * 1024, "edge device / low RAM", "bartowski/Llama-3.2-1B-Instruct-GGUF"},
		{"qwen3:1.7b", 1100 * 1024 * 1024, "good quality, very fast", "bartowski/Qwen3-1.7B-GGUF"},
		{"llama3.2:3b", 2000 * 1024 * 1024, "great balance", "bartowski/Llama-3.2-3B-Instruct-GGUF"},
		{"qwen3:4b", 2600 * 1024 * 1024, "solid quality", "bartowski/Qwen3-4B-GGUF"},
		{"mistral:7b", 4400 * 1024 * 1024, "workhorse model", "bartowski/Mistral-7B-Instruct-v0.3-GGUF"},
		{"llama3.1:8b", 4900 * 1024 * 1024, "best 8B model", "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"},
		{"qwen3:8b", 5200 * 1024 * 1024, "reasoning capable", "bartowski/Qwen3-8B-GGUF"},
		{"phi4:14b", 8600 * 1024 * 1024, "Microsoft, very capable", "bartowski/phi-4-GGUF"},
		{"qwen3:14b", 9300 * 1024 * 1024, "high quality reasoning", "bartowski/Qwen3-14B-GGUF"},
		{"mistral-small:22b", 13500 * 1024 * 1024, "near-frontier quality", "bartowski/Mistral-Small-Instruct-2409-GGUF"},
		{"qwen3:32b", 20400 * 1024 * 1024, "top open model", "bartowski/Qwen3-32B-GGUF"},
		{"llama3.3:70b", 43000 * 1024 * 1024, "frontier quality", "bartowski/Llama-3.3-70B-Instruct-GGUF"},
	}

	shown := 0
	for _, r := range recommendations {
		fits := "✓"
		if availMem > 0 && r.sizeB > availMem {
			fits = "✗"
		}
		sizeStr := formatBytes(r.sizeB)
		fmt.Printf("  %-28s %-10s %-8s %s\n", r.name, sizeStr, fits, r.notes)
		shown++
	}

	fmt.Println()
	fmt.Printf("  Pull any model with:  lychee hf pull <repo>\n")
	fmt.Printf("  Browse all models:    lychee hf list\n")
	fmt.Printf("  Search by type:       lychee hf search code\n")
	_ = shown
	return nil
}

// getSystemRAM returns total system RAM in bytes, best effort
func getSystemRAM() int64 {
	// Try reading /proc/meminfo on Linux
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					var kb int64
					fmt.Sscanf(strings.TrimPrefix(line, "MemTotal:"), "%d", &kb)
					return kb * 1024
				}
			}
		}
	}
	return 0
}
