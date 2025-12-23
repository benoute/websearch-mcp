package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Fetch performs an HTTP GET request and returns the response body as a string.
// It only accepts text-based content types and returns an error for binary content
// or non-2xx status codes.
func Fetch(ctx context.Context, rawURL string, timeout time.Duration) (string, error) {
	// Validate URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("URL missing scheme: %s", rawURL)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL missing host: %s", rawURL)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "websearch-mcp/1.0")

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Validate content type
	contentType := resp.Header.Get("Content-Type")
	if !isTextContentType(contentType) {
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// isTextContentType checks if the content type is text-based.
func isTextContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	// Extract the media type (ignore parameters like charset)
	mediaType := strings.TrimSpace(strings.Split(contentType, ";")[0])
	mediaType = strings.ToLower(mediaType)

	allowedTypes := []string{
		"text/html",
		"text/plain",
		"application/json",
		"application/xml",
		"text/xml",
	}

	for _, allowed := range allowedTypes {
		if mediaType == allowed {
			return true
		}
	}

	// Also accept variants like application/xhtml+xml, application/rss+xml, etc.
	if strings.HasSuffix(mediaType, "+xml") || strings.HasSuffix(mediaType, "+json") {
		return true
	}

	return false
}
