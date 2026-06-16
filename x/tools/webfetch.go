package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
)

const (
	webFetchTimeout = 30 * time.Second
)

// WebFetchTool implements web page fetching locally.
type WebFetchTool struct{}

// Name returns the tool name.
func (w *WebFetchTool) Name() string {
	return "web_fetch"
}

// Description returns a description of the tool.
func (w *WebFetchTool) Description() string {
	return "Fetch and extract text content from a web page. Use this to read the full content of a URL found in search results or provided by the user."
}

// Schema returns the tool's parameter schema.
func (w *WebFetchTool) Schema() api.ToolFunction {
	props := api.NewToolPropertiesMap()
	props.Set("url", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The URL to fetch and extract content from",
	})
	return api.ToolFunction{
		Name:        w.Name(),
		Description: w.Description(),
		Parameters: api.ToolFunctionParameters{
			Type:       "object",
			Properties: props,
			Required:   []string{"url"},
		},
	}
}

// Execute fetches content from a web page.
func (w *WebFetchTool) Execute(args map[string]any) (string, error) {
	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		return "", fmt.Errorf("url parameter is required")
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Send request directly
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Allow proxy configuration via environment standard proxy support in default client
	client := &http.Client{Timeout: webFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("web page returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	htmlContent := string(bodyBytes)
	title := extractTitle(htmlContent)
	content := cleanHTML(htmlContent)

	var sb strings.Builder
	if title != "" {
		sb.WriteString(fmt.Sprintf("Title: %s\n\n", title))
	}

	if content != "" {
		sb.WriteString("Content:\n")
		sb.WriteString(content)
	} else {
		sb.WriteString("No content could be extracted from the page.")
	}

	return sb.String(), nil
}
