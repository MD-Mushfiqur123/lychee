package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/format"
)

type CatalogItem struct {
	Name string   `json:"name"`
	ID   string   `json:"id"`
	Tags []string `json:"tags"`
	Size string   `json:"size"`
}

func NewCatalogCmd() *cobra.Command {
	var port int
	var token string

	catalogCmd := &cobra.Command{
		Use:   "serve-catalog",
		Short: "Serve local model list as a shareable JSON catalog index",
		Long:  `Launches a lightweight HTTP server on the specified port. Exposes local models in registry format.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := api.ClientFromEnvironment()
			if err != nil {
				return err
			}

			http.HandleFunc("/catalog", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")

				// Bearer token check
				if token != "" {
					authHeader := r.Header.Get("Authorization")
					expectedHeader := "Bearer " + token
					if authHeader != expectedHeader {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"error": "Unauthorized"}`))
						return
					}
				}

				// Query local installed models
				modelsResp, err := client.List(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
					return
				}

				hostAddr := r.Host
				if strings.HasPrefix(hostAddr, "localhost") || strings.HasPrefix(hostAddr, "127.0.0.1") {
					// Fallback to local default if default hostname is resolved
					hostAddr = "127.0.0.1:11434"
				}

				var catalog []CatalogItem
				for _, m := range modelsResp.Models {
					tags := []string{"chat"}
					for _, c := range m.Capabilities {
						tags = append(tags, c.String())
					}

					catalog = append(catalog, CatalogItem{
						Name: m.Name,
						ID:   fmt.Sprintf("%s/%s", hostAddr, m.Name),
						Tags: tags,
						Size: format.HumanBytes(m.Size),
					})
				}

				_ = json.NewEncoder(w).Encode(catalog)
			})

			addr := ":" + strconv.Itoa(port)
			fmt.Printf("🍒 Lychee Registry Catalog Server listening on http://localhost:%d/catalog\n", port)
			if token != "" {
				fmt.Println("🔒 Security token enabled.")
			}
			fmt.Println("Other users on your network can query this using:")
			if token != "" {
				fmt.Printf("  lychee search --catalog http://<your-ip>:%d/ --token %s <query>\n", port, token)
			} else {
				fmt.Printf("  lychee search --catalog http://<your-ip>:%d/ <query>\n", port)
			}
			
			err = http.ListenAndServe(addr, nil)
			if err != nil {
				return fmt.Errorf("failed to start catalog server on port %d: %w", port, err)
			}
			return nil
		},
	}

	catalogCmd.Flags().IntVarP(&port, "port", "p", 9090, "Port to run catalog server on")
	catalogCmd.Flags().StringVarP(&token, "token", "t", "", "Optional authorization token required for security")
	return catalogCmd
}
