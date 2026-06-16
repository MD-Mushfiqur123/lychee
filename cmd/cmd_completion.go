package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion scripts for the specified shell",
		Long: `To load completions:

Bash:

  $ source <(lychee completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ lychee completion bash > /etc/bash_completion.d/lychee
  # macOS:
  $ lychee completion bash > /usr/local/etc/bash_completion.d/lychee

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ lychee completion zsh > "${fpath[1]}/_lychee"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ lychee completion fish > ~/.config/fish/completions/lychee.fish

PowerShell:

  PS> lychee completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> lychee completion powershell > lychee.ps1
  # and source this file in your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return completionCmd
}
