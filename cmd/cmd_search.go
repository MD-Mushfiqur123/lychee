package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lychee/lychee/cmd/launch"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewSearchCmd() *cobra.Command {
	var catalogURL string
	var token string

	cmd := &cobra.Command{
		Use:     "search QUERY",
		Short:   "Search for models in the registry/catalog",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: checkServerHeartbeat,
		RunE: func(cmd *cobra.Command, args []string) error {
			return searchHandler(cmd, args, catalogURL, token)
		},
	}

	cmd.Flags().StringVar(&catalogURL, "catalog", "", "Optional custom catalog server URL (e.g. http://192.168.1.100:9090/)")
	cmd.Flags().StringVarP(&token, "token", "t", "", "Authentication token for custom catalog server")
	return cmd
}

func fetchCatalogFromURL(catalogURL, query, token string) [][4]string {
	urlStr := strings.TrimSuffix(catalogURL, "/") + "/catalog"
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var items []struct {
		Name string   `json:"name"`
		ID   string   `json:"id"`
		Tags []string `json:"tags"`
		Size string   `json:"size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil
	}

	var entries [][4]string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), query) || strings.Contains(strings.ToLower(item.ID), query) {
			tags := strings.Join(item.Tags, ",")
			entries = append(entries, [4]string{
				item.Name,
				item.ID,
				tags,
				item.Size,
			})
		}
	}
	return entries
}

func searchHandler(cmd *cobra.Command, args []string, catalogURL, token string) error {
	query := strings.ToLower(strings.Join(args, " "))
	
	if catalogURL != "" {
		fmt.Printf("Searching for %q on custom catalog %s...\n\n", query, catalogURL)
	} else {
		fmt.Printf("Searching for %q on HuggingFace and local catalog...\n\n", query)
	}

	seen := make(map[string]bool)
	var merged [][4]string

	// If custom catalog URL is specified, query it first
	if catalogURL != "" {
		customEntries := fetchCatalogFromURL(catalogURL, query, token)
		for _, m := range customEntries {
			repoLower := strings.ToLower(m[1])
			if !seen[repoLower] {
				merged = append(merged, m)
				seen[repoLower] = true
			}
		}
	} else {
		for _, m := range hfCatalogEntries() {
			if strings.Contains(strings.ToLower(m[0]), query) ||
				strings.Contains(strings.ToLower(m[1]), query) ||
				strings.Contains(strings.ToLower(m[2]), query) {
				merged = append(merged, m)
				seen[strings.ToLower(m[1])] = true
			}
		}

		dynamic := fetchDynamicHFModels(query)
		for _, m := range dynamic {
			repoLower := strings.ToLower(m[1])
			if !seen[repoLower] {
				merged = append(merged, m)
				seen[repoLower] = true
			}
		}
	}

	if len(merged) == 0 {
		fmt.Println("No models found matching your query.")
		return nil
	}

	fmt.Printf("  %-32s %-52s\n", "NAME", "REGISTRY/REPOS ID")
	fmt.Println("  " + strings.Repeat("-", 90))
	for _, m := range merged {
		fmt.Printf("  %-32s %-52s\n", m[0], m[1])
	}
	fmt.Println()

	// If interactive, offer to pull one of the models
	if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
		var selectionItems []launch.SelectionItem
		for _, m := range merged {
			selectionItems = append(selectionItems, launch.SelectionItem{
				Name:        m[1],
				Description: fmt.Sprintf("%s (%s)", m[0], m[1]),
			})
		}
		
		choice, err := runTUISingleSelector("Select a model to pull/download:", selectionItems, "", nil)
		if err != nil {
			if errors.Is(err, launch.ErrCancelled) {
				return nil
			}
			return err
		}
		
		if choice != "" {
			fmt.Printf("Pulling model %q...\n", choice)
			if catalogURL != "" {
				// Pull from custom catalog host directly
				return PullHandler(cmd, []string{choice})
			}
			return HFPullHandler(cmd, []string{choice}, "", false)
		}
	}

	return nil
}
