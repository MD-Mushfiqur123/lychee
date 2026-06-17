package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/discover"
	"github.com/MD-Mushfiqur123/lychee/format"
	"github.com/MD-Mushfiqur123/lychee/version"
)

func NewInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show system and Lychee information",
		Long: `Display combined system and Lychee information including version,
Go version, OS/Arch/CPUs, RAM, GPU, server status, and installed models count.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// ── Lychee ─────────────────────────────────────────────
			fmt.Printf("🧠  Lychee version    %s\n", version.Version)

			// ── Go runtime ─────────────────────────────────────────
			fmt.Printf("🔧  Go version        %s\n", runtime.Version())

			// ── OS / Arch / CPUs ───────────────────────────────────
			fmt.Printf("💻  OS / Arch         %s / %s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("🔄  CPUs              %d\n", runtime.NumCPU())

			// ── RAM ────────────────────────────────────────────────
			mem, err := discover.GetCPUMem()
			if err == nil {
				fmt.Printf("🧮  RAM total         %s\n", format.HumanBytes2(mem.TotalMemory))
				fmt.Printf("📊  RAM used          %s\n", format.HumanBytes2(mem.TotalMemory-mem.FreeMemory))
				fmt.Printf("🆓  RAM free          %s\n", format.HumanBytes2(mem.FreeMemory))
			} else {
				fmt.Println("🧮  RAM               unable to detect")
			}

			// ── GPU ────────────────────────────────────────────────
			gpuLine := detectInfoGPU()
			fmt.Printf("🎮  GPU               %s\n", gpuLine)

			// ── Server status ──────────────────────────────────────
			client, clientErr := api.ClientFromEnvironment()
			if clientErr != nil {
				fmt.Println("🌐  Server status     ⚠️  client error")
			} else {
				if err := client.Heartbeat(cmd.Context()); err == nil {
					fmt.Println("🌐  Server status     ✅ running")
				} else {
					fmt.Println("🌐  Server status     ⬜ not running")
				}
			}

			// ── Installed models ───────────────────────────────────
			if client != nil {
				resp, err := client.List(cmd.Context())
				if err == nil {
					fmt.Printf("📦  Installed models  %d\n", len(resp.Models))
				} else {
					fmt.Println("📦  Installed models  unable to query (server may not be running)")
				}
			}

			return nil
		},
	}
}

// detectInfoGPU tries nvidia-smi and returns a human-readable GPU description.
func detectInfoGPU() string {
	out, err := exec.Command("nvidia-smi",
		"--query-gpu=name",
		"--format=csv,noheader,nounits",
	).Output()
	if err == nil && len(out) > 0 {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		// Filter empty lines
		var names []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				names = append(names, l)
			}
		}
		if len(names) == 0 {
			return "not detected"
		}
		if len(names) == 1 {
			return names[0]
		}
		return fmt.Sprintf("%d GPUs (%s + %d more)",
			len(names), names[0], len(names)-1)
	}
	return "not detected"
}
