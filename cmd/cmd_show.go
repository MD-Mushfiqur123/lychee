package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/format"
)

func ShowHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	license, errLicense := cmd.Flags().GetBool("license")
	modelfile, errModelfile := cmd.Flags().GetBool("modelfile")
	parameters, errParams := cmd.Flags().GetBool("parameters")
	system, errSystem := cmd.Flags().GetBool("system")
	template, errTemplate := cmd.Flags().GetBool("template")
	verbose, errVerbose := cmd.Flags().GetBool("verbose")

	for _, boolErr := range []error{errLicense, errModelfile, errParams, errSystem, errTemplate, errVerbose} {
		if boolErr != nil {
			return errors.New("error retrieving flags")
		}
	}

	flagsSet := 0
	showType := ""

	if license {
		flagsSet++
		showType = "license"
	}

	if modelfile {
		flagsSet++
		showType = "modelfile"
	}

	if parameters {
		flagsSet++
		showType = "parameters"
	}

	if system {
		flagsSet++
		showType = "system"
	}

	if template {
		flagsSet++
		showType = "template"
	}

	if flagsSet > 1 {
		return errors.New("only one of '--license', '--modelfile', '--parameters', '--system', or '--template' can be specified")
	}

	req := api.ShowRequest{Name: args[0], Verbose: verbose}
	resp, err := client.Show(cmd.Context(), &req)
	if err != nil {
		return err
	}

	if flagsSet == 1 {
		switch showType {
		case "license":
			fmt.Println(resp.License)
		case "modelfile":
			fmt.Println(resp.Modelfile)
		case "parameters":
			fmt.Println(resp.Parameters)
		case "system":
			fmt.Print(resp.System)
		case "template":
			fmt.Print(resp.Template)
		}

		return nil
	}

	return showInfo(resp, verbose, os.Stdout)
}

func showInfo(resp *api.ShowResponse, verbose bool, w io.Writer) error {
	tableRender := func(header string, rows func() [][]string) {
		fmt.Fprintln(w, " ", header)
		table := tablewriter.NewWriter(w)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorder(false)
		table.SetNoWhiteSpace(true)
		table.SetTablePadding("    ")

		switch header {
		case "Template", "System", "License":
			table.SetColWidth(100)
		}

		table.AppendBulk(rows())
		table.Render()
		fmt.Fprintln(w)
	}

	tableRender("Model", func() (rows [][]string) {
		if resp.RemoteHost != "" {
			rows = append(rows, []string{"", "Remote model", resp.RemoteModel})
			rows = append(rows, []string{"", "Remote URL", resp.RemoteHost})
		}

		if resp.ModelInfo != nil {
			arch, _ := resp.ModelInfo["general.architecture"].(string)
			if arch != "" {
				rows = append(rows, []string{"", "architecture", arch})
			}

			var paramStr string
			if resp.Details.ParameterSize != "" {
				paramStr = resp.Details.ParameterSize
			} else if v, ok := resp.ModelInfo["general.parameter_count"]; ok {
				if f, ok := v.(float64); ok {
					paramStr = format.HumanNumber(uint64(f))
				}
			}
			if paramStr != "" {
				rows = append(rows, []string{"", "parameters", paramStr})
			}

			if v, ok := resp.ModelInfo[fmt.Sprintf("%s.context_length", arch)]; ok {
				if f, ok := v.(float64); ok {
					rows = append(rows, []string{"", "context length", strconv.FormatFloat(f, 'f', -1, 64)})
				}
			}

			if v, ok := resp.ModelInfo[fmt.Sprintf("%s.embedding_length", arch)]; ok {
				if f, ok := v.(float64); ok {
					rows = append(rows, []string{"", "embedding length", strconv.FormatFloat(f, 'f', -1, 64)})
				}
			}
		} else {
			rows = append(rows, []string{"", "architecture", resp.Details.Family})
			rows = append(rows, []string{"", "parameters", resp.Details.ParameterSize})
		}
		rows = append(rows, []string{"", "quantization", resp.Details.QuantizationLevel})
		if resp.Requires != "" {
			rows = append(rows, []string{"", "requires", resp.Requires})
		}
		return
	})

	if len(resp.Capabilities) > 0 {
		tableRender("Capabilities", func() (rows [][]string) {
			for _, capability := range resp.Capabilities {
				rows = append(rows, []string{"", capability.String()})
			}
			return
		})
	}

	if resp.ProjectorInfo != nil {
		tableRender("Projector", func() (rows [][]string) {
			arch, _ := resp.ProjectorInfo["general.architecture"].(string)
			if arch != "" {
				rows = append(rows, []string{"", "architecture", arch})
			}
			if v, ok := resp.ProjectorInfo["general.parameter_count"].(float64); ok {
				rows = append(rows, []string{"", "parameters", format.HumanNumber(uint64(v))})
			}

			projectorValue := func(suffix string) (float64, bool) {
				for _, modality := range []string{"vision", "audio"} {
					if v, ok := resp.ProjectorInfo[fmt.Sprintf("%s.%s.%s", arch, modality, suffix)].(float64); ok {
						return v, true
					}
				}
				return 0, false
			}
			if v, ok := projectorValue("embedding_length"); ok {
				rows = append(rows, []string{"", "embedding length", strconv.FormatFloat(v, 'f', -1, 64)})
			}
			if v, ok := projectorValue("projection_dim"); ok {
				rows = append(rows, []string{"", "dimensions", strconv.FormatFloat(v, 'f', -1, 64)})
			}
			return
		})
	}

	if resp.Parameters != "" {
		tableRender("Parameters", func() (rows [][]string) {
			scanner := bufio.NewScanner(strings.NewReader(resp.Parameters))
			for scanner.Scan() {
				if text := scanner.Text(); text != "" {
					rows = append(rows, append([]string{""}, strings.Fields(text)...))
				}
			}
			return
		})
	}

	if resp.ModelInfo != nil && verbose {
		tableRender("Metadata", func() (rows [][]string) {
			keys := make([]string, 0, len(resp.ModelInfo))
			for k := range resp.ModelInfo {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				var v string
				switch vData := resp.ModelInfo[k].(type) {
				case bool:
					v = fmt.Sprintf("%t", vData)
				case string:
					v = vData
				case float64:
					v = fmt.Sprintf("%g", vData)
				case []any:
					targetWidth := 10 // Small width where we are displaying the data in a column

					var itemsToShow int
					totalWidth := 1 // Start with 1 for opening bracket

					// Find how many we can fit
					for i := range vData {
						itemStr := fmt.Sprintf("%v", vData[i])
						width := runewidth.StringWidth(itemStr)

						// Add separator width (", ") for all items except the first
						if i > 0 {
							width += 2
						}

						// Check if adding this item would exceed our width limit
						if totalWidth+width > targetWidth && i > 0 {
							break
						}

						totalWidth += width
						itemsToShow++
					}

					// Format the output
					if itemsToShow < len(vData) {
						v = fmt.Sprintf("%v", vData[:itemsToShow])
						v = strings.TrimSuffix(v, "]")
						v += fmt.Sprintf(" ...+%d more]", len(vData)-itemsToShow)
					} else {
						v = fmt.Sprintf("%v", vData)
					}
				default:
					v = fmt.Sprintf("%T", vData)
				}
				rows = append(rows, []string{"", k, v})
			}
			return
		})
	}

	if len(resp.Tensors) > 0 && verbose {
		tableRender("Tensors", func() (rows [][]string) {
			for _, t := range resp.Tensors {
				rows = append(rows, []string{"", t.Name, t.Type, fmt.Sprint(t.Shape)})
			}
			return
		})
	}

	head := func(s string, n int) (rows [][]string) {
		scanner := bufio.NewScanner(strings.NewReader(s))
		count := 0
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}
			count++
			if n < 0 || count <= n {
				rows = append(rows, []string{"", text})
			}
		}
		if n >= 0 && count > n {
			rows = append(rows, []string{"", "..."})
		}
		return
	}

	if resp.System != "" {
		tableRender("System", func() [][]string {
			return head(resp.System, 2)
		})
	}

	if resp.License != "" {
		tableRender("License", func() [][]string {
			return head(resp.License, 2)
		})
	}

	return nil
}
