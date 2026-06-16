package cmd

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/format"
)

func NewInspectCmd() *cobra.Command {
	inspectCmd := &cobra.Command{
		Use:               "inspect MODEL",
		Short:             "Inspect model metadata and predict VRAM usage profiles",
		Long:              `Retrieve internal GGUF parameters and calculate expected VRAM footprints for different context lengths.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: autocompleteInstalledModels,
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			client, err := api.ClientFromEnvironment()
			if err != nil {
				return err
			}

			// 1. Get size from client List
			listResp, err := client.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			var modelSize int64
			var found bool
			for _, m := range listResp.Models {
				if m.Name == modelName || m.Model == modelName {
					modelSize = m.Size
					found = true
					break
				}
			}

			if !found {
				// Try standard substring match just in case
				for _, m := range listResp.Models {
					if strings.Contains(strings.ToLower(m.Name), strings.ToLower(modelName)) {
						modelSize = m.Size
						modelName = m.Name
						found = true
						break
					}
				}
			}

			// 2. Query Show info
			showResp, err := client.Show(cmd.Context(), &api.ShowRequest{Name: modelName})
			if err != nil {
				return fmt.Errorf("failed to inspect model properties: %w", err)
			}

			arch, _ := showResp.ModelInfo["general.architecture"].(string)
			if arch == "" {
				arch = showResp.Details.Family
			}
			if arch == "" {
				arch = "llama" // default fallback
			}

			// GGUF parameter extractors
			var layers uint64
			if v, ok := showResp.ModelInfo[fmt.Sprintf("%s.block_count", arch)]; ok {
				if val, ok := v.(float64); ok {
					layers = uint64(val)
				}
			} else if v, ok := showResp.ModelInfo["general.block_count"]; ok {
				if val, ok := v.(float64); ok {
					layers = uint64(val)
				}
			}

			var kvHeads uint64
			if v, ok := showResp.ModelInfo[fmt.Sprintf("%s.attention.head_count_kv", arch)]; ok {
				if val, ok := v.(float64); ok {
					kvHeads = uint64(val)
				}
			} else if v, ok := showResp.ModelInfo["general.attention.head_count_kv"]; ok {
				if val, ok := v.(float64); ok {
					kvHeads = uint64(val)
				}
			}
			if kvHeads == 0 {
				kvHeads = 1
			}

			var embeddingLength uint64
			if v, ok := showResp.ModelInfo[fmt.Sprintf("%s.embedding_length", arch)]; ok {
				if val, ok := v.(float64); ok {
					embeddingLength = uint64(val)
				}
			} else if v, ok := showResp.ModelInfo["general.embedding_length"]; ok {
				if val, ok := v.(float64); ok {
					embeddingLength = uint64(val)
				}
			}

			var headCount uint64
			if v, ok := showResp.ModelInfo[fmt.Sprintf("%s.attention.head_count", arch)]; ok {
				if val, ok := v.(float64); ok {
					headCount = uint64(val)
				}
			} else if v, ok := showResp.ModelInfo["general.attention.head_count"]; ok {
				if val, ok := v.(float64); ok {
					headCount = uint64(val)
				}
			}

			var headDim uint64
			if headCount > 0 {
				headDim = embeddingLength / headCount
			} else {
				headDim = 128
			}

			fmt.Fprintf(cmd.OutOrStdout(), "🍒 Lychee Model Inspector: %s\n\n", modelName)
			fmt.Fprintln(cmd.OutOrStdout(), "Model Properties:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Architecture:   %s\n", arch)
			fmt.Fprintf(cmd.OutOrStdout(), "  Parameters:     %s\n", showResp.Details.ParameterSize)
			fmt.Fprintf(cmd.OutOrStdout(), "  Quantization:   %s\n", showResp.Details.QuantizationLevel)
			fmt.Fprintf(cmd.OutOrStdout(), "  File Format:    %s\n", showResp.Details.Format)
			if modelSize > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  File Size:      %s\n\n", format.HumanBytes(modelSize))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  File Size:      Unknown\n\n")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Estimated VRAM Memory Layout Profiles:")

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"CONTEXT WINDOW", "WEIGHTS SIZE", "KV CACHE SIZE", "ESTIMATED TOTAL VRAM"})
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetBorder(true)
			table.SetAutoWrapText(false)

			contextSizes := []int{2048, 4096, 8192, 16384, 32768}
			var data [][]string

			weights := uint64(modelSize)
			baseComputeOverhead := uint64(300 * 1024 * 1024) // 300 MiB

			for _, numCtx := range contextSizes {
				// KV cache calculation
				kvCache := 2 * layers * kvHeads * headDim * uint64(numCtx) * 2
				contextComputeScaling := uint64(numCtx) * 10240
				totalVram := weights + kvCache + baseComputeOverhead + contextComputeScaling

				data = append(data, []string{
					fmt.Sprintf("%d tokens", numCtx),
					format.HumanBytes(int64(weights)),
					format.HumanBytes(int64(kvCache)),
					format.HumanBytes(int64(totalVram)),
				})
			}

			table.AppendBulk(data)
			table.Render()

			fmt.Fprintln(cmd.OutOrStdout(), "\n* Note: Predictions include base model weights, KV Cache context scaling, and a 300 MiB compute graph overhead.")
			return nil
		},
	}

	return inspectCmd
}
