package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SearchResult represents a single search result from SearxNG.
type SearchResult struct {
	Title   string
	URL     string
	Content string
}

// searxngResponse represents the JSON response from SearxNG.
type searxngResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

// Search queries a SearxNG instance and returns up to limit results.
func Search(ctx context.Context, searxngURL, query string, limit int) ([]SearchResult, error) {
	// Build the request URL
	u, err := url.Parse(searxngURL)
	if err != nil {
		return nil, fmt.Errorf("invalid searxng URL: %w", err)
	}
	u.Path = "/search"
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "websearch-mcp/1.0")

	// Create client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse JSON response
	var searxResp searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&searxResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to SearchResult slice, respecting limit
	results := make([]SearchResult, 0, limit)
	for i, r := range searxResp.Results {
		if i >= limit {
			break
		}
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
		})
	}

	return results, nil
}
