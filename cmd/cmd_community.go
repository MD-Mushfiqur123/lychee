package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func NewCommunityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Display Lychee community links and contribution pipelines",
		Long:  `Outputs links to documentation, chat rooms, and pulls beginner-friendly open-source issues directly from GitHub.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "🍒 Welcome to the Lychee Community Hub!")
			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────────────────────")
			fmt.Fprintln(cmd.OutOrStdout(), "🌐 Website & Docs: https://lychee.github.io/")
			fmt.Fprintln(cmd.OutOrStdout(), "💬 Discord Invite: https://discord.gg/lychee-ai")
			fmt.Fprintln(cmd.OutOrStdout(), "🐙 GitHub Repo:    https://github.com/MD-Mushfiqur123/lychee")
			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────────────────────")
			fmt.Fprintln(cmd.OutOrStdout(), "\nLoading active 'good first issues' from GitHub...")

			// Fetch issues from GitHub API
			client := &http.Client{Timeout: 3 * time.Second}
			resp, err := client.Get("https://api.github.com/repos/lychee-ai/lychee/issues?labels=good-first-issue&state=open&per_page=5")
			if err != nil {
				// Fallback to static list if offline
				printStaticIssues(cmd)
				return nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				printStaticIssues(cmd)
				return nil
			}

			var issues []struct {
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				Number  int    `json:"number"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil || len(issues) == 0 {
				printStaticIssues(cmd)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\n🔥 Good First Issues for Contributors:")
			for _, issue := range issues {
				fmt.Fprintf(cmd.OutOrStdout(), "  [#%d] %s\n      Link: %s\n", issue.Number, issue.Title, issue.HTMLURL)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nTo claim an issue, run: lychee community claim <issue_id>")
			return nil
		},
	}

	cmd.AddCommand(NewCommunityClaimCmd())
	cmd.AddCommand(NewCommunityFeedbackCmd())
	WireCommunityPublishCommands(cmd)
	return cmd
}

func NewCommunityClaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claim ISSUE_ID",
		Short: "Claim a community issue and print local setup guidelines",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueID := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "🍒 Claiming Lychee issue #%s...\n", issueID)
			fmt.Fprintln(cmd.OutOrStdout(), "To start working on this issue, follow these steps:")
			fmt.Fprintln(cmd.OutOrStdout(), "  1. Fork the repository on GitHub: https://github.com/lychee-ai/lychee")
			fmt.Fprintln(cmd.OutOrStdout(), "  2. Clone your fork locally:")
			fmt.Fprintln(cmd.OutOrStdout(), "     git clone https://github.com/<your-username>/lychee.git")
			fmt.Fprintln(cmd.OutOrStdout(), "     cd lychee")
			fmt.Fprintln(cmd.OutOrStdout(), "  3. Create a new branch for the issue:")
			fmt.Fprintf(cmd.OutOrStdout(), "     git checkout -b feature/claim-issue-%s\n", issueID)
			fmt.Fprintln(cmd.OutOrStdout(), "  4. Make your changes and run tests to verify.")
			fmt.Fprintln(cmd.OutOrStdout(), "  5. Push the branch and open a Pull Request!")
			fmt.Fprintln(cmd.OutOrStdout(), "\nGood luck! Reach out on Discord (https://discord.gg/lychee-ai) if you need help.")
			return nil
		},
	}
}

func NewCommunityFeedbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "feedback MODEL_ID",
		Short: "View peer reviews and feedback for a community model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelID := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "💬 Fetching feedback for model %q...\n", modelID)

			comments, err := fetchModelFeedback(modelID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "\n❌ Failed to retrieve feedback: %v\n", err)
				fmt.Fprintln(cmd.OutOrStdout(), "Ensure the model ID matches a community submission or check your internet connection.")
				return nil
			}

			if len(comments) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n💬 Model %q has no reviews or feedback yet.\n", modelID)
				fmt.Fprintln(cmd.OutOrStdout(), "Be the first to leave a review on GitHub!")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n⭐ Community Reviews - %d feedback items:\n", len(comments))
			fmt.Fprintln(cmd.OutOrStdout(), "─────────────────────────────────────────────────────────────────")
			for _, comment := range comments {
				timeStr := comment.CreatedAt.Format("2006-01-02 15:04:05")
				fmt.Fprintf(cmd.OutOrStdout(), "👤 @%s — %s\n", comment.User.Login, timeStr)
				fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", comment.Body)
				fmt.Fprintln(cmd.OutOrStdout(), "─────────────────────────────────────────────────────────────────")
			}
			return nil
		},
	}
}

func printStaticIssues(cmd *cobra.Command) {
	fmt.Fprintln(cmd.OutOrStdout(), "\n🔥 Curated Issues for Beginners:")
	fmt.Fprintln(cmd.OutOrStdout(), "  [#12] Add shell auto-completions for fish/powershell")
	fmt.Fprintln(cmd.OutOrStdout(), "      Link: https://github.com/MD-Mushfiqur123/lychee/issues/12")
	fmt.Fprintln(cmd.OutOrStdout(), "  [#18] Implement custom model exporter to portable tarballs")
	fmt.Fprintln(cmd.OutOrStdout(), "      Link: https://github.com/MD-Mushfiqur123/lychee/issues/18")
	fmt.Fprintln(cmd.OutOrStdout(), "\nJoin our Discord or visit GitHub to browse more open issues!")
}
