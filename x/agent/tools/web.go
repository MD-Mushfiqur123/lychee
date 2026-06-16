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
	"github.com/MD-Mushfiqur123/lychee/x/agent"
)

// ErrWebSearchAuthRequired is returned when web search requires authentication.
var ErrWebSearchAuthRequired = errors.New("web search requires authentication")

const (
	webSearchTimeout = 15 * time.Second
	webFetchTimeout  = 30 * time.Second
)

// WebSearchTool implements web search using local requests.
type WebSearchTool struct{}

func (w *WebSearchTool) Name() string {
	return "web_search"
}

func (w *WebSearchTool) Description() string {
	return "Search the web for current information. Use this when you need up-to-date information that may not be in your training data."
}

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

type webSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func (w *WebSearchTool) Execute(args map[string]any) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	// Safety: check query against deny patterns
	if safe, pattern := agent.IsSafeCommand(query); !safe {
		return fmt.Sprintf("Search query blocked: matches dangerous pattern (%s).", pattern), nil
	}

	ctx := context.Background()
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

	if len(results) == 0 {
		return "No results found for query: " + query, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Content != "" {
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

// WebFetchTool implements web page fetching locally.
type WebFetchTool struct{}

func (w *WebFetchTool) Name() string {
	return "web_fetch"
}

func (w *WebFetchTool) Description() string {
	return "Fetch and extract text content from a web page. Use this to read the full content of a URL found in search results or provided by the user."
}

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

	// Security: prevent loopback or private network calls
	hostname := parsedURL.Hostname()
	if hostname == "localhost" || hostname == "127.0.0.1" || strings.HasPrefix(hostname, "192.168.") || strings.HasPrefix(hostname, "10.") {
		return "Fetch blocked: intranet or loopback URLs are restricted.", nil
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

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

func cleanHTML(html string) string {
	for {
		startIdx := strings.Index(strings.ToLower(html), "<script")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(strings.ToLower(html[startIdx:]), "</script>")
		if endIdx == -1 {
			break
		}
		html = html[:startIdx] + html[startIdx+endIdx+len("</script>"):]
	}

	for {
		startIdx := strings.Index(strings.ToLower(html), "<style")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(strings.ToLower(html[startIdx:]), "</style>")
		if endIdx == -1 {
			break
		}
		html = html[:startIdx] + html[startIdx+endIdx+len("</style>"):]
	}

	var sb strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
			sb.WriteRune(' ')
		} else if !inTag {
			sb.WriteRune(r)
		}
	}

	text := sb.String()
	replacer := strings.NewReplacer(
		"&quot;", "\"",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&nbsp;", " ",
	)
	text = replacer.Replace(text)

	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

func extractTitle(html string) string {
	lower := strings.ToLower(html)
	start := strings.Index(lower, "<title>")
	if start == -1 {
		return ""
	}
	end := strings.Index(lower[start:], "</title>")
	if end == -1 {
		return ""
	}
	title := html[start+len("<title>") : start+end]
	return strings.TrimSpace(title)
}
