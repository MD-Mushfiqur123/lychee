package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	xcmd "github.com/MD-Mushfiqur123/lychee/x/cmd"
)

func NewAgentCmd() *cobra.Command {
	agentCmd := &cobra.Command{
		Use:     "agent MODEL",
		Short:   "Run a model in agentic tool-use mode",
		Args:    cobra.ExactArgs(1),
		PreRunE: checkServerHeartbeat,
		RunE:    AgentHandler,
	}

	agentCmd.Flags().String("keepalive", "", "Duration to keep a model loaded (e.g. 5m)")
	agentCmd.Flags().Bool("verbose", false, "Show timings for response")
	agentCmd.Flags().Bool("nowordwrap", false, "Don't wrap words to the next line automatically")
	agentCmd.Flags().String("think", "", "Enable thinking mode: true/false or high/medium/low for supported models")
	agentCmd.Flags().Lookup("think").NoOptDefVal = "true"
	agentCmd.Flags().Bool("hidethinking", false, "Hide thinking output (if provided)")
	agentCmd.Flags().Bool("yolo", false, "Skip all tool approval prompts (use with caution)")
	agentCmd.Flags().Bool("websearch", false, "Enable web search tool in agent mode")

	return agentCmd
}

func AgentHandler(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	info, err := func() (*api.ShowResponse, error) {
		showReq := &api.ShowRequest{Name: modelName}
		info, err := client.Show(cmd.Context(), showReq)
		var se api.StatusError
		if errors.As(err, &se) && se.StatusCode == http.StatusNotFound {
			if err := PullHandler(cmd, []string{modelName}); err != nil {
				return nil, err
			}
			return client.Show(cmd.Context(), &api.ShowRequest{Name: modelName})
		}
		return info, err
	}()
	if err != nil {
		return err
	}

	ensureCloudStub(cmd.Context(), client, modelName)

	opts := runOptions{
		Model:       modelName,
		WordWrap:    os.Getenv("TERM") == "xterm-256color",
		Options:     map[string]any{},
		ShowConnect: true,
	}

	nowrap, _ := cmd.Flags().GetBool("nowordwrap")
	opts.WordWrap = !nowrap

	thinkFlag := cmd.Flags().Lookup("think")
	if thinkFlag.Changed {
		thinkStr, _ := cmd.Flags().GetString("think")
		switch thinkStr {
		case "", "true":
			opts.Think = &api.ThinkValue{Value: true}
		case "false":
			opts.Think = &api.ThinkValue{Value: false}
		case "high", "medium", "low", "max":
			opts.Think = &api.ThinkValue{Value: thinkStr}
		default:
			return fmt.Errorf("invalid value for --think: %q", thinkStr)
		}
	} else {
		opts.Think, err = inferThinkingOption(&info.Capabilities, &opts, false)
		if err != nil {
			return err
		}
	}

	keepAlive, _ := cmd.Flags().GetString("keepalive")
	if keepAlive != "" {
		d, err := time.ParseDuration(keepAlive)
		if err != nil {
			return err
		}
		opts.KeepAlive = &api.Duration{Duration: d}
	}

	yoloMode, _ := cmd.Flags().GetBool("yolo")
	enableWebsearch, _ := cmd.Flags().GetBool("websearch")
	hideThinking, _ := cmd.Flags().GetBool("hidethinking")

	// Preload the model
	if err := loadOrUnloadModel(cmd, &opts); err != nil {
		return err
	}

	// Always run interactive agent loop for agent mode
	return xcmd.GenerateInteractive(cmd, opts.Model, opts.WordWrap, opts.Options, opts.Think, hideThinking, opts.KeepAlive, yoloMode, enableWebsearch)
}
