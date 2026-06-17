package cmd

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/format"
)

const (
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiReset  = "\033[0m"
)

// colorSize returns a color-coded human-readable size string.
// Small (< 2 GB) → green, medium (2–10 GB) → yellow, large (> 10 GB) → red.
func colorSize(size int64) string {
	s := format.HumanBytes(size)
	switch {
	case size < 2_000_000_000:
		return ansiGreen + s + ansiReset
	case size < 10_000_000_000:
		return ansiYellow + s + ansiReset
	default:
		return ansiRed + s + ansiReset
	}
}

// ModelsHandler handles the "lychee models" command (and its "list" / "ls" aliases).
func ModelsHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	models, err := client.List(cmd.Context())
	if err != nil {
		return err
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	sortBySize, _ := cmd.Flags().GetBool("size")

	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(models)
	}

	// Filter by name prefix when an argument is supplied.
	var filtered []api.ListModelResponse
	for _, m := range models.Models {
		if len(args) == 0 || strings.HasPrefix(strings.ToLower(m.Name), strings.ToLower(args[0])) {
			filtered = append(filtered, m)
		}
	}

	// Sort: by name (default) or by size (descending).
	if sortBySize {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Size > filtered[j].Size
		})
	} else {
		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		})
	}

	var data [][]string
	for _, m := range filtered {
		// Size: color-coded for local models, "-" for remote.
		var size string
		if m.RemoteModel != "" {
			size = "-"
		} else {
			size = colorSize(m.Size)
		}

		// Quantization level from model details.
		quant := "-"
		if m.Details.QuantizationLevel != "" {
			quant = m.Details.QuantizationLevel
		}

		modified := format.HumanTime(m.ModifiedAt, "Never")

		data = append(data, []string{m.Name, size, modified, quant})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "SIZE", "MODIFIED", "QUANTIZATION"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)
	table.SetTablePadding("    ")
	table.AppendBulk(data)
	table.Render()

	return nil
}
