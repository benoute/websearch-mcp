package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	websearch "github.com/benoute/websearch-mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type websearchToolInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"The search query. Optionally add 'site:website.com' in the search query to search within a specific website (website.com in this example)`
	Limit int    `json:"limit,omitempty" jsonschema_description:"Maximum number of search results (default: 10)"`
}

type SearchResultOutput struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Snippet string   `json:"snippet"`
	Engines []string `json:"engines"`
	Score   float64  `json:"score"`
}

func setupMCPServer(config Config) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "websearch", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "websearch",
		Description: "Search the web using SearxNG.",
	}, func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input websearchToolInput,
	) (*mcp.CallToolResult, any, error) {
		return handleWebsearch(ctx, config, input)
	})

	return server
}

func handleWebsearch(
	ctx context.Context,
	config Config,
	input websearchToolInput,
) (*mcp.CallToolResult, any, error) {
	// Validate required input
	if input.Query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "query is required"},
			},
			IsError: true,
		}, nil, nil
	}

	// Apply defaults
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	// Search using SearxNG
	searchStart := time.Now()
	results, err := websearch.Search(ctx, config.SearxngURL, input.Query, limit)
	searchDuration := time.Since(searchStart)
	config.Logger.LogAttrs(ctx, slog.LevelDebug,
		"SearxNG search",
		slog.String("query", input.Query),
		slog.Int("results", len(results)),
		slog.Int64("duration_ms", searchDuration.Milliseconds()),
	)

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "search failed: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	var outputs []SearchResultOutput
	for _, result := range results {
		outputs = append(outputs, SearchResultOutput{
			Title:   result.Title,
			URL:     result.URL,
			Snippet: result.Content,
			Engines: result.Engines,
			Score:   result.Score,
		})
	}

	// Marshal results to JSON
	jsonBytes, err := json.Marshal(outputs)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "failed to marshal results: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}
