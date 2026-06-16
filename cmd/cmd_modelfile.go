package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MD-Mushfiqur123/lychee/parser"
)

func NewModelfileCmd() *cobra.Command {
	modelfileCmd := &cobra.Command{
		Use:   "modelfile",
		Short: "Manage and analyze Modelfiles",
	}

	lintCmd := &cobra.Command{
		Use:   "lint [filename]",
		Short: "Lint a Modelfile for syntax errors",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "Modelfile"
			if len(args) > 0 {
				filename = args[0]
			}

			f, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("could not open %s: %w", filename, err)
			}
			defer f.Close()

			mf, err := parser.ParseFile(f)
			if err != nil {
				return fmt.Errorf("lint failed for %s: %w", filename, err)
			}

			warnings := 0
			for _, c := range mf.Commands {
				if strings.EqualFold(c.Name, "system") {
					warnings++
					fmt.Fprintf(cmd.OutOrStdout(), "⚠️ Warning: SYSTEM is deprecated. Use FROM and message roles instead.\n")
				}
			}

			if warnings > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠️ Lint passed with %d warnings for %s.\n", warnings, filename)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✅ Lint passed for %s. No issues found.\n", filename)
			}

			return nil
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactively initialize a new Modelfile",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🍒 Welcome to the Lychee Modelfile Builder wizard! Let's build your Modelfile.")
			
			reader := bufio.NewReader(os.Stdin)

			// 1. Ask for base model
			fmt.Print("Enter base model name or path (e.g. llama3.2, /path/to/model.gguf) [default: llama3.2]: ")
			fromModel, _ := reader.ReadString('\n')
			fromModel = strings.TrimSpace(fromModel)
			if fromModel == "" {
				fromModel = "llama3.2"
			}

			// 2. Ask for system prompt
			fmt.Print("Enter system prompt (instruction) [default: None]: ")
			systemPrompt, _ := reader.ReadString('\n')
			systemPrompt = strings.TrimSpace(systemPrompt)

			// 3. Ask for temperature
			fmt.Print("Enter temperature (e.g. 0.7) [default: 0.7]: ")
			tempStr, _ := reader.ReadString('\n')
			tempStr = strings.TrimSpace(tempStr)
			if tempStr == "" {
				tempStr = "0.7"
			}
			tempVal, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				tempVal = 0.7
			}

			// 4. Ask for num_ctx
			fmt.Print("Enter context length (num_ctx, e.g. 2048, 4096) [default: 2048]: ")
			numCtxStr, _ := reader.ReadString('\n')
			numCtxStr = strings.TrimSpace(numCtxStr)
			if numCtxStr == "" {
				numCtxStr = "2048"
			}
			numCtxVal, err := strconv.Atoi(numCtxStr)
			if err != nil {
				numCtxVal = 2048
			}

			// Format Modelfile content
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("FROM %s\n", fromModel))
			builder.WriteString(fmt.Sprintf("PARAMETER temperature %f\n", tempVal))
			builder.WriteString(fmt.Sprintf("PARAMETER num_ctx %d\n", numCtxVal))
			if systemPrompt != "" {
				builder.WriteString(fmt.Sprintf("SYSTEM %q\n", systemPrompt))
			}

			// Save to Modelfile
			filename := "Modelfile"
			err = os.WriteFile(filename, []byte(builder.String()), 0644)
			if err != nil {
				return fmt.Errorf("failed to save Modelfile: %w", err)
			}

			fmt.Printf("✅ Success! Formatted Modelfile saved to ./%s\n", filename)
			fmt.Println("To compile and create your model, run:")
			fmt.Printf("  lychee create your-custom-name -f ./%s\n", filename)

			return nil
		},
	}

	modelfileCmd.AddCommand(lintCmd, initCmd)
	return modelfileCmd
}
