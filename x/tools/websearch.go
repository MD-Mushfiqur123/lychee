package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
)

var ErrWebSearchAuthRequired = errors.New("web search requires authentication")

const (
	webSearchTimeout = 15 * time.Second
)

// WebSearchTool implements web search using local requests.
type WebSearchTool struct{}

// Name returns the tool name.
func (w *WebSearchTool) Name() string {
	return "web_search"
}

// Description returns a description of the tool.
func (w *WebSearchTool) Description() string {
	return "Search the web for current information. Use this when you need up-to-date information that may not be in your training data."
}

// Schema returns the tool's parameter schema.
func (w *WebSearchTool) Schema() api.ToolFunction {
	props := api.NewToolPropertiesMap()
	props.Set("query", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The search query to look up on the web",
	})
	return api.ToolFunction{
		Name:        w.Name(),
		Description: w.Description(),
		Parameters: api.ToolFunctionParameters{
			Type:       "object",
			Properties: props,
			Required:   []string{"query"},
		},
	}
}

// webSearchResult is a single search result.
type webSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// Execute performs the web search.
func (w *WebSearchTool) Execute(args map[string]any) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	ctx := context.Background()

	// Check if a search API key is provided for Brave Search
	apiKey := os.Getenv("LYCHEE_SEARCH_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("BRAVE_SEARCH_API_KEY")
	}

	var results []webSearchResult
	var err error

	if apiKey != "" {
		results, err = braveSearch(ctx, query, apiKey)
	} else {
		results, err = duckDuckGoSearch(ctx, query)
	}

	if err != nil {
		return "", fmt.Errorf("web search failed: %w", err)
	}

	// Format results
	if len(results) == 0 {
		return "No results found for query: " + query, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Content != "" {
			// Truncate long content (UTF-8 safe)
			content := result.Content
			runes := []rune(content)
			if len(runes) > 300 {
				content = string(runes[:300]) + "..."
			}
			sb.WriteString(fmt.Sprintf("   %s\n", content))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func braveSearch(ctx context.Context, query, apiKey string) ([]webSearchResult, error) {
	searchURL := "https://api.search.brave.com/res/v1/web/search?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Subscription-Token", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: webSearchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Brave API returned status %d: %s", resp.StatusCode, string(body))
	}

	var braveResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, err
	}

	var results []webSearchResult
	for _, r := range braveResp.Web.Results {
		results = append(results, webSearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Description,
		})
		if len(results) >= 5 {
			break
		}
	}
	return results, nil
}

func duckDuckGoSearch(ctx context.Context, query string) ([]webSearchResult, error) {
	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: webSearchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo HTML API returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(bodyBytes)

	var results []webSearchResult
	searchPos := 0
	for len(results) < 5 {
		idx := strings.Index(html[searchPos:], "class=\"result__a\"")
		if idx == -1 {
			break
		}
		absoluteIdx := searchPos + idx
		tagStart := strings.LastIndex(html[:absoluteIdx], "<a ")
		if tagStart == -1 {
			searchPos = absoluteIdx + len("class=\"result__a\"")
			continue
		}
		tagEnd := strings.Index(html[tagStart:], ">")
		if tagEnd == -1 {
			searchPos = absoluteIdx + len("class=\"result__a\"")
			continue
		}
		tagContent := html[tagStart : tagStart+tagEnd]

		hrefIdx := strings.Index(tagContent, "href=\"")
		if hrefIdx == -1 {
			searchPos = absoluteIdx + len("class=\"result__a\"")
			continue
		}
		hrefValue := tagContent[hrefIdx+len("href=\""):]
		quoteEnd := strings.Index(hrefValue, "\"")
		if quoteEnd == -1 {
			searchPos = absoluteIdx + len("class=\"result__a\"")
			continue
		}
		rawURL := hrefValue[:quoteEnd]

		actualURL := rawURL
		if strings.Contains(rawURL, "uddg=") {
			u, err := url.Parse(rawURL)
			if err == nil {
				actualURL = u.Query().Get("uddg")
			}
		} else if strings.HasPrefix(rawURL, "//") {
			actualURL = "https:" + rawURL
		}

		titleEnd := strings.Index(html[tagStart+tagEnd:], "</a>")
		if titleEnd == -1 {
			searchPos = absoluteIdx + len("class=\"result__a\"")
			continue
		}
		rawTitle := html[tagStart+tagEnd+1 : tagStart+tagEnd+titleEnd]
		title := cleanHTML(rawTitle)

		snippet := ""
		snippetIdx := strings.Index(html[tagStart+tagEnd+titleEnd:], "class=\"result__snippet\"")
		if snippetIdx != -1 {
			nextResultIdx := strings.Index(html[tagStart+tagEnd+titleEnd:], "class=\"result__a\"")
			if nextResultIdx == -1 || snippetIdx < nextResultIdx {
				snippetStart := tagStart + tagEnd + titleEnd + snippetIdx
				snippetTagEnd := strings.Index(html[snippetStart:], ">")
				if snippetTagEnd != -1 {
					snippetContentEnd := strings.Index(html[snippetStart+snippetTagEnd:], "</a>")
					if snippetContentEnd != -1 {
						rawSnippet := html[snippetStart+snippetTagEnd+1 : snippetStart+snippetTagEnd+snippetContentEnd]
						snippet = cleanHTML(rawSnippet)
					}
				}
			}
		}

		if actualURL != "" && title != "" {
			results = append(results, webSearchResult{
				Title:   title,
				URL:     actualURL,
				Content: snippet,
			})
		}

		searchPos = tagStart + tagEnd + titleEnd + len("</a>")
	}

	return results, nil
}
