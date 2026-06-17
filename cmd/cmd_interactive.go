package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
)

// ANSI color codes for dark terminals.
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
)

const interactiveDefaultModel = "qwen2.5:3b"

func NewInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive [model]",
		Short: "Start an interactive chat REPL",
		Long:  `Start an interactive chat session with a model. Type /help for commands, /exit to quit.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  InteractiveHandler,
	}
}

func InteractiveHandler(cmd *cobra.Command, args []string) error {
	model := interactiveDefaultModel
	if len(args) > 0 {
		model = args[0]
	}

	printWelcome(model)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(colorCyan + "> " + colorReset)
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch {
		case input == "/exit" || input == "/quit":
			fmt.Println(colorYellow + "\n  Goodbye! 👋" + colorReset)
			return nil
		case input == "/help":
			printHelp()
		case strings.HasPrefix(input, "/model "):
			model = strings.TrimPrefix(input, "/model ")
			fmt.Printf(colorGreen+"  ✓ Switched to: %s\n"+colorReset, model)
		case input == "/models":
			listModels(cmd)
		case input == "/clear":
			fmt.Print("\033[2J\033[H")
			printWelcome(model)
		case input == "/system":
			printSystemInfo()
		default:
			resp, err := chatWithModel(cmd, model, input)
			if err != nil {
				fmt.Printf(colorRed+"  ✗ Error: %v\n"+colorReset, err)
			} else {
				fmt.Println(colorBlue + formatResponse(resp) + colorReset)
			}
		}
		fmt.Println()
	}
	return nil
}

func printWelcome(model string) {
	fmt.Printf(colorBold+colorCyan+"  🤖 Lychee REPL"+colorReset+colorDim+" — model: %s\n"+colorReset, model)
	fmt.Println(colorDim + "  Type /help for commands, /exit to quit, /model to switch" + colorReset)
	fmt.Println()
}

func printHelp() {
	fmt.Println(colorBold + "\n  ── REPL Commands ──" + colorReset)
	fmt.Println()
	fmt.Printf("  %s/exit, /quit%s        Exit the REPL\n", colorYellow, colorReset)
	fmt.Printf("  %s/help%s              Show this help\n", colorYellow, colorReset)
	fmt.Printf("  %s/model <name>%s      Switch to a different model\n", colorYellow, colorReset)
	fmt.Printf("  %s/models%s            List available models\n", colorYellow, colorReset)
	fmt.Printf("  %s/clear%s             Clear the screen\n", colorYellow, colorReset)
	fmt.Printf("  %s/system%s            Show system info (GPU, RAM, CPU)\n", colorYellow, colorReset)
	fmt.Println()
	fmt.Printf("  %sAnything else%s is sent to the model as a prompt.\n", colorDim, colorReset)
	fmt.Println()
}

func printSystemInfo() {
	fmt.Println(colorBold + "\n  ── System Information ──" + colorReset)
	fmt.Println()

	// CPU
	fmt.Printf("  %sCPU:%s  %s (%d cores)\n", colorMagenta, colorReset, runtime.GOARCH, runtime.NumCPU())

	// OS
	fmt.Printf("  %sOS:%s   %s\n", colorMagenta, colorReset, runtime.GOOS)

	// Memory
	mi := readMemInfo()
	fmt.Printf("  %sMem:%s  %s total / %s used / %s free\n",
		colorMagenta, colorReset,
		formatBytes(mi.total),
		formatBytes(mi.used),
		formatBytes(mi.free),
	)

	// Go runtime
	fmt.Printf("  %sGo:%s   %s\n", colorMagenta, colorReset, runtime.Version())

	// GPU
	fmt.Println()
	fmt.Printf("  %sGPU:%s  ", colorMagenta, colorReset)
	gpuInfo := detectGPU()
	if gpuInfo != "" {
		fmt.Print(gpuInfo)
	} else {
		fmt.Print(colorDim + "No GPU detected or nvidia-smi not available" + colorReset)
	}
	fmt.Println()
	fmt.Println()
}

func detectGPU() string {
	// Try nvidia-smi first
	out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,memory.used", "--format=csv,noheader,nounits").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var parts []string
		for _, line := range lines {
			fields := strings.Split(line, ", ")
			if len(fields) >= 3 {
				name := strings.TrimSpace(fields[0])
				total := strings.TrimSpace(fields[1])
				used := strings.TrimSpace(fields[2])
				totalMB, _ := strconv.ParseFloat(total, 64)
				usedMB, _ := strconv.ParseFloat(used, 64)
				parts = append(parts, fmt.Sprintf("%s (%s / %s)",
					name,
					formatBytes(int64(usedMB*1024*1024)),
					formatBytes(int64(totalMB*1024*1024)),
				))
			}
		}
		return strings.Join(parts, "\n         ")
	}
	return ""
}

func listModels(cmd *cobra.Command) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		fmt.Printf(colorRed+"  ✗ Cannot connect to Lychee: %v\n"+colorReset, err)
		return
	}

	models, err := client.List(cmd.Context())
	if err != nil {
		fmt.Printf(colorRed+"  ✗ Cannot list models: %v\n"+colorReset, err)
		return
	}

	if len(models.Models) == 0 {
		fmt.Println(colorDim + "  No models installed. Use 'lychee pull <model>' to download one." + colorReset)
		return
	}

	fmt.Println(colorBold + "\n  ── Available Models ──" + colorReset)
	fmt.Println()
	for _, m := range models.Models {
		fmt.Printf("  %s•%s %s%s%s  %s%s%s\n",
			colorGreen, colorReset,
			colorBold, m.Name, colorReset,
			colorDim, formatBytes(m.Size), colorReset,
		)
	}
	fmt.Println()
}

func chatWithModel(cmd *cobra.Command, model, prompt string) (string, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "", fmt.Errorf("failed to create API client: %w", err)
	}

	var response strings.Builder
	stream := false

	err = client.Generate(cmd.Context(), &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: &stream,
	}, func(resp api.GenerateResponse) error {
		response.WriteString(resp.Response)
		return nil
	})

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.String()), nil
}

func formatResponse(resp string) string {
	// Indent each line of the response with 2 spaces for visual clarity.
	lines := strings.Split(resp, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return "\n" + strings.Join(lines, "\n")
}

// ── Memory helpers (cross-platform) ──

type memInfo struct{ total, used, free int64 }

func readMemInfo() memInfo {
	// Try /proc/meminfo (Linux)
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		return parseProcMeminfo(string(data))
	}
	// Fallback for macOS: use sysctl + vm_stat
	if runtime.GOOS == "darwin" {
		return darwinMemInfo()
	}
	// Windows: use wmic
	if runtime.GOOS == "windows" {
		return windowsMemInfo()
	}
	return memInfo{}
}

func parseProcMeminfo(s string) memInfo {
	var total, avail int64
	for _, line := range strings.Split(s, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseInt(fields[1], 10, 64)
		valBytes := val * 1024
		switch fields[0] {
		case "MemTotal:":
			total = valBytes
		case "MemAvailable:":
			avail = valBytes
		}
	}
	used := total - avail
	return memInfo{total: total, used: used, free: avail}
}

func darwinMemInfo() memInfo {
	// Use sysctl for total
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return memInfo{}
	}
	total, _ := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)

	// Use vm_stat for free pages
	out, err = exec.Command("bash", "-c", "vm_stat | head -7").Output()
	if err != nil {
		return memInfo{total: total}
	}
	freePages := extractVMStatValue(string(out), "free")
	pageSizeStr, err := exec.Command("pagesize").Output()
	if err != nil {
		return memInfo{total: total}
	}
	pageSize, err := strconv.ParseInt(strings.TrimSpace(string(pageSizeStr)), 10, 64)
	if err != nil {
		return memInfo{total: total}
	}
	free := freePages * pageSize
	used := total - free
	return memInfo{total: total, used: used, free: free}
}

func extractVMStatValue(text, key string) int64 {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, key) {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				valStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".")
				val, _ := strconv.ParseInt(valStr, 10, 64)
				return val
			}
		}
	}
	return 0
}

func windowsMemInfo() memInfo {
	// Use wmic to get total physical memory
	out, err := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory").Output()
	if err != nil {
		return memInfo{}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 2 {
		totalStr := strings.TrimSpace(lines[1])
		total, _ := strconv.ParseInt(totalStr, 10, 64)
		// Get available memory
		out2, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory").Output()
		if err == nil {
			lines2 := strings.Split(strings.TrimSpace(string(out2)), "\n")
			if len(lines2) >= 2 {
				freeKB, _ := strconv.ParseInt(strings.TrimSpace(lines2[1]), 10, 64)
				free := freeKB * 1024
				return memInfo{total: total, used: total - free, free: free}
			}
		}
		return memInfo{total: total}
	}
	return memInfo{}
}
