package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/x/agent"
)

const (
	bashTimeout   = 60 * time.Second
	maxOutputSize = 50000
)

type BashTool struct{}

func (b *BashTool) Name() string {
	return "bash"
}

func (b *BashTool) Description() string {
	return "Execute a shell command on the system. Use this to run shell commands, check files, run programs, etc."
}

func (b *BashTool) Schema() api.ToolFunction {
	props := api.NewToolPropertiesMap()
	props.Set("command", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The command to execute",
	})
	return api.ToolFunction{
		Name:        b.Name(),
		Description: b.Description(),
		Parameters: api.ToolFunctionParameters{
			Type:       "object",
			Properties: props,
			Required:   []string{"command"},
		},
	}
}

func (b *BashTool) Execute(args map[string]any) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command parameter is required")
	}

	// Run sandbox command checks
	if safe, pattern := agent.IsSafeCommand(command); !safe {
		return fmt.Sprintf("Command blocked: this command matches a dangerous pattern (%s) and cannot be executed.", pattern), nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), bashTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-Command", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var sb strings.Builder
	if stdout.Len() > 0 {
		output := stdout.String()
		if len(output) > maxOutputSize {
			output = output[:maxOutputSize] + "\n... (output truncated)"
		}
		sb.WriteString(output)
	}

	if stderr.Len() > 0 {
		stderrOutput := stderr.String()
		if len(stderrOutput) > maxOutputSize {
			stderrOutput = stderrOutput[:maxOutputSize] + "\n... (stderr truncated)"
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("stderr:\n")
		sb.WriteString(stderrOutput)
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return sb.String() + "\n\nError: command timed out after 60 seconds", nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return sb.String() + fmt.Sprintf("\n\nExit code: %d", exitErr.ExitCode()), nil
		}
		return sb.String(), fmt.Errorf("executing command: %w", err)
	}

	if sb.Len() == 0 {
		return "(no output)", nil
	}

	return sb.String(), nil
}
