package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/spf13/cobra"
)

// NewStatsCmd creates the cmd command for lychee stats.
func NewStatsCmd() *cobra.Command {
	var intervalFlag int
	var tuiFlag bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show live model performance stats",
		Long: `Displays a live dashboard of all running models showing:
  - tokens/second generation speed
  - VRAM and RAM usage
  - context length in use
  - how long the model has been loaded

Press Ctrl+C to exit.

Examples:
  lychee stats
  lychee stats --interval 2
  lychee stats --tui`,
		Args:    cobra.NoArgs,
		PreRunE: checkServerHeartbeat,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tuiFlag {
				return statsHandlerTUI(time.Duration(intervalFlag) * time.Second)
			}
			return statsHandler(cmd.Context(), time.Duration(intervalFlag)*time.Second)
		},
	}
	cmd.Flags().IntVar(&intervalFlag, "interval", 1, "Refresh interval in seconds")
	cmd.Flags().BoolVar(&tuiFlag, "tui", false, "Use rich TUI dashboard")
	return cmd
}

func statsHandler(ctx context.Context, interval time.Duration) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}

	fmt.Print("\033[?25l")          // hide cursor
	defer fmt.Print("\033[?25h\n") // restore cursor

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	first := true
	lines := 0

	render := func() {
		resp, err := client.ListRunning(ctx)
		if err != nil {
			return
		}

		// Clear previous render
		if !first {
			fmt.Printf("\033[%dA\033[J", lines)
		}
		first = false
		lines = 0

		now := time.Now().Format("15:04:05")
		header := fmt.Sprintf("  Lychee Stats  [%s]  Ctrl+C to exit", now)
		fmt.Println(header)
		fmt.Println("  " + strings.Repeat("─", 76))
		lines += 2

		if len(resp.Models) == 0 {
			fmt.Println("  No models currently loaded.")
			fmt.Println("  Run a model with: lychee run <model>")
			lines += 2
			return
		}

		fmt.Printf("  %-28s %-10s %-10s %-10s %s\n",
			"MODEL", "VRAM", "RAM", "CTX", "EXPIRES")
		fmt.Println("  " + strings.Repeat("─", 76))
		lines += 2

		for _, m := range resp.Models {
			name := m.Name
			if len(name) > 27 {
				name = name[:24] + "..."
			}
			vram := formatBytes(m.SizeVRAM)
			ram := formatBytes(m.Size - m.SizeVRAM)
			if m.Size <= m.SizeVRAM {
				ram = "─"
			}
			ctxVal := "─"
			if m.ContextLength > 0 {
				ctxVal = fmt.Sprintf("%dk", m.ContextLength/1000)
			}
			expires := "─"
			if !m.ExpiresAt.IsZero() {
				remaining := time.Until(m.ExpiresAt)
				if remaining > 0 {
					expires = formatDuration(remaining)
				} else {
					expires = "unloading"
				}
			}
			fmt.Printf("  %-28s %-10s %-10s %-10s %s\n",
				name, vram, ram, ctxVal, expires)
			lines++
		}

		fmt.Println()
		lines++
	}

	render()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			render()
		}
	}
}

type statsMsg struct {
	running *api.ProcessResponse
	err     error
}

type statsModel struct {
	client   *api.Client
	running  *api.ProcessResponse
	err      error
	interval time.Duration
	quitting bool
	width    int
	height   int
}

func (m statsModel) Init() tea.Cmd {
	return m.tick()
}

func (m statsModel) tick() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		resp, err := m.client.ListRunning(context.Background())
		return statsMsg{running: resp, err: err}
	})
}

func (m statsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "r":
			resp, err := m.client.ListRunning(context.Background())
			m.running = resp
			m.err = err
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case statsMsg:
		m.running = msg.running
		m.err = msg.err
		if m.quitting {
			return m, nil
		}
		return m, m.tick()
	}
	return m, nil
}

func (m statsModel) View() string {
	if m.quitting {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		MarginBottom(1)

	headerText := "🍒 Lychee Premium TUI Monitor  |  Press 'q' to quit, 'r' to refresh"
	if m.width > 0 && m.width < 70 {
		headerText = "🍒 Lychee Stats  |  'q' to quit"
	}
	header := headerStyle.Render(headerText)

	if m.err != nil {
		return header + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error connecting to server: %v", m.err))
	}

	if m.running == nil {
		return header + "\n Loading server stats..."
	}

	s := header + "\n"

	if len(m.running.Models) == 0 {
		s += lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240")).Render("  No models currently loaded. Start one with `lychee run <model>`\n")
		return s
	}

	if m.width > 0 && m.width < 70 {
		// Compact view for narrow terminals
		tableHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("248"))
		s += tableHeaderStyle.Render(fmt.Sprintf("  %-24s %-10s %s\n", "MODEL", "VRAM", "EXPIRES"))
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  " + strings.Repeat("─", 50) + "\n")

		rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		for _, loaded := range m.running.Models {
			name := loaded.Name
			if len(name) > 23 {
				name = name[:20] + "..."
			}
			vram := formatBytes(loaded.SizeVRAM)
			expires := "─"
			if !loaded.ExpiresAt.IsZero() {
				remaining := time.Until(loaded.ExpiresAt)
				if remaining > 0 {
					expires = formatDuration(remaining)
				} else {
					expires = "unloading"
				}
			}
			s += rowStyle.Render(fmt.Sprintf("  %-24s %-10s %s\n", name, vram, expires))
		}
		return s
	}

	// Normal full view
	tableHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("248"))
	s += tableHeaderStyle.Render(fmt.Sprintf("  %-30s %-12s %-12s %-10s %s\n", "MODEL", "VRAM", "RAM", "CTX LIMIT", "EXPIRES"))
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  " + strings.Repeat("─", 80) + "\n")

	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	for _, loaded := range m.running.Models {
		name := loaded.Name
		if len(name) > 29 {
			name = name[:26] + "..."
		}
		vram := formatBytes(loaded.SizeVRAM)
		ram := formatBytes(loaded.Size - loaded.SizeVRAM)
		if loaded.Size <= loaded.SizeVRAM {
			ram = "─"
		}
		ctxVal := "─"
		if loaded.ContextLength > 0 {
			ctxVal = fmt.Sprintf("%dk", loaded.ContextLength/1000)
		}
		expires := "─"
		if !loaded.ExpiresAt.IsZero() {
			remaining := time.Until(loaded.ExpiresAt)
			if remaining > 0 {
				expires = formatDuration(remaining)
			} else {
				expires = "unloading"
			}
		}
		s += rowStyle.Render(fmt.Sprintf("  %-30s %-12s %-12s %-10s %s\n", name, vram, ram, ctxVal, expires))
	}

	return s
}

func statsHandlerTUI(interval time.Duration) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}

	running, err := client.ListRunning(context.Background())

	m := statsModel{
		client:   client,
		running:  running,
		err:      err,
		interval: interval,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI dashboard: %w", err)
	}

	return nil
}
