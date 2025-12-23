package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SearchResult represents a single search result from SearxNG.
type SearchResult struct {
	Title   string
	URL     string
	Content string
	Engines []string
	Score   float64
}

// searxngResponse represents the JSON response from SearxNG.
type searxngResponse struct {
	Results []struct {
		Title   string   `json:"title"`
		URL     string   `json:"url"`
		Content string   `json:"content"`
		Engines []string `json:"engines"`
		Score   float64  `json:"score"`
	} `json:"results"`
}

// Search queries a SearxNG instance and returns up to limit results.
// If fewer results are returned than requested, additional pages are fetched.
func Search(ctx context.Context, searxngURL, query string, limit int) ([]SearchResult, error) {
	results := make([]SearchResult, 0, limit)
	seen := make(map[string]bool)
	pageNo := 1

	// Create client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for len(results) < limit {
		pageResults, err := fetchPage(ctx, client, searxngURL, query, pageNo)
		if err != nil {
			// If we already have some results and this page fails, return what we have
			if len(results) > 0 {
				break
			}
			return nil, err
		}

		// No more results available
		if len(pageResults) == 0 {
			break
		}

		// Add results, deduplicating by URL
		for _, r := range pageResults {
			if len(results) >= limit {
				break
			}
			if !seen[r.URL] {
				seen[r.URL] = true
				results = append(results, r)
			}
		}

		pageNo++
	}

	return results, nil
}

// fetchPage fetches a single page of results from SearxNG.
func fetchPage(ctx context.Context, client *http.Client, searxngURL, query string, pageNo int) ([]SearchResult, error) {
	// Build the request URL
	u, err := url.Parse(searxngURL)
	if err != nil {
		return nil, fmt.Errorf("invalid searxng URL: %w", err)
	}
	u.Path = "/search"
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("pageno", strconv.Itoa(pageNo))
	u.RawQuery = q.Encode()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "websearch-mcp/1.0")

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

	// Convert to SearchResult slice
	results := make([]SearchResult, 0, len(searxResp.Results))
	for _, r := range searxResp.Results {
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
			Engines: r.Engines,
			Score:   r.Score,
		})
	}

	return results, nil
}
