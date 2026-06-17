package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/progress"
)

func PullHandler(cmd *cobra.Command, args []string) error {
	modelName := args[0]

	// Resolve --run flag and auto-detect terminal
	shouldRun, _ := cmd.Flags().GetBool("run")
	if !shouldRun && term.IsTerminal(int(os.Stdout.Fd())) {
		shouldRun = true
	}

	// Auto-route: if model contains "/", treat as HuggingFace model reference
	if strings.Contains(modelName, "/") {
		fmt.Printf("Detected HuggingFace model: %s\n", modelName)
		fmt.Printf("Routing to HuggingFace pull...\n\n")

		quant, _ := cmd.Flags().GetString("quant")
		list, _ := cmd.Flags().GetBool("list")

		// For --list, just show quants and exit (no download)
		if list {
			return HFPullHandler(cmd, []string{modelName}, quant, list)
		}

		if err := HFPullHandler(cmd, []string{modelName}, quant, false); err != nil {
			return fmt.Errorf("failed to pull from HuggingFace: %w", err)
		}

		// After successful HF pull: auto-create and run if requested
		if shouldRun {
			// Determine model name and modelfile path
			parts := strings.SplitN(modelName, "/", 2)
			org, repo := parts[0], parts[1]

			modelDir := resolveHFModelDir()
			mfPath := filepath.Join(modelDir, "hf-models", org, repo, "Modelfile")

			if _, err := os.Stat(mfPath); os.IsNotExist(err) {
				fmt.Printf("\n  No Modelfile found at %s — skipping create/run\n", mfPath)
				return nil
			}

			fmt.Printf("\n  Creating model and starting chat...\n\n")

			// 1. Create the model from the Modelfile
			createCmd := &cobra.Command{Use: "create"}
			createCmd.Flags().StringP("file", "f", mfPath, "Name of the Modelfile")
			createCmd.Flags().StringP("quantize", "q", "", "Quantize model to this level")
			createCmd.Flags().String("draft-quantize", "", "Quantize draft model to this level")
			createCmd.Flags().Bool("experimental", false, "Enable experimental safetensors model creation")
			createCmd.SetContext(cmd.Context())
			if err := CreateHandler(createCmd, []string{repo}); err != nil {
				return fmt.Errorf("creating model: %w", err)
			}

			// 2. Start interactive chat
			runCmd := &cobra.Command{Use: "run"}
			runCmd.SetContext(cmd.Context())
			if err := launchInteractiveModel(runCmd, repo); err != nil {
				return fmt.Errorf("starting chat: %w", err)
			}
		}

		return nil
	}

	// Standard Lychee/Ollama pull path
	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return fmt.Errorf("failed to parse --insecure flag: %w", err)
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return wrapServerError("connect to Lychee", err)
	}

	p := progress.NewProgress(os.Stderr)
	defer p.Stop()

	bars := make(map[string]*progress.Bar)

	var status string
	var spinner *progress.Spinner

	fn := func(resp api.ProgressResponse) error {
		if resp.Digest != "" {
			if resp.Completed == 0 {
				// This is the initial status update for the
				// layer, which the server sends before
				// beginning the download, for clients to
				// compute total size and prepare for
				// downloads, if needed.
				//
				// Skipping this here to avoid showing a 0%
				// progress bar, which *should* clue the user
				// into the fact that many things are being
				// downloaded and that the current active
				// download is not that last. However, in rare
				// cases it seems to be triggering to some, and
				// it isn't worth explaining, so just ignore
				// and regress to the old UI that keeps giving
				// you the "But wait, there is more!" after
				// each "100% done" bar, which is "better."
				return nil
			}

			if spinner != nil {
				spinner.Stop()
			}

			bar, ok := bars[resp.Digest]
			if !ok {
				name, isDigest := strings.CutPrefix(resp.Digest, "sha256:")
				name = strings.TrimSpace(name)
				if isDigest {
					name = name[:min(12, len(name))]
				}
				bar = progress.NewBar(fmt.Sprintf("pulling %s:", name), resp.Total, resp.Completed)
				bars[resp.Digest] = bar
				p.Add(resp.Digest, bar)
			}

			bar.Set(resp.Completed)
		} else if status != resp.Status {
			if spinner != nil {
				spinner.Stop()
			}

			status = resp.Status
			spinner = progress.NewSpinner(status)
			p.Add(status, spinner)
		}

		return nil
	}

	request := api.PullRequest{Name: args[0], Insecure: insecure}
	if err := client.Pull(cmd.Context(), &request, fn); err != nil {
		return wrapServerError("pull model", err)
	}

	// After successful Ollama pull: auto-run if requested
	if shouldRun {
		// Strip tag from model name (e.g. "llama3.2:latest" → "llama3.2")
		runName := strings.SplitN(modelName, ":", 2)[0]

		fmt.Printf("\n  Starting chat with %s...\n\n", runName)

		runCmd := &cobra.Command{Use: "run"}
		runCmd.SetContext(cmd.Context())
		if err := launchInteractiveModel(runCmd, runName); err != nil {
			return fmt.Errorf("starting chat: %w", err)
		}
	}

	return nil
}

// resolveHFModelDir returns the directory where HF models are stored.
func resolveHFModelDir() string {
	modelDir := filepath.Join(os.Getenv("USERPROFILE"), ".lychee", "models")
	if h := os.Getenv("HOME"); h != "" && modelDir == filepath.Join("", ".lychee", "models") {
		modelDir = filepath.Join(h, ".lychee", "models")
	}
	if lycheeModels := os.Getenv("LYCHEE_MODELS"); lycheeModels != "" {
		modelDir = lycheeModels
	}
	return modelDir
}
