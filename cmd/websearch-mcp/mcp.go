package main

import (
	"context"
	"encoding/json"
	"time"

	websearch "github.com/benoute/websearch-mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/sync/errgroup"
)

type websearchToolInput struct {
	Query            string `json:"query" jsonschema:"required" jsonschema_description:"The search query"`
	Limit            int    `json:"limit,omitempty" jsonschema_description:"Maximum number of search results (default: 8)"`
	Summary          bool   `json:"summary,omitempty" jsonschema_description:"Enable summarization of each web page from the results (default: false)"`
	MaxSummaryTokens int    `json:"max_summary_tokens,omitempty" jsonschema_description:"Maximum tokens per summary (default: 256)"`
}

type SearchResultOutput struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Summary string `json:"summary,omitempty"`
}

func setupMCPServer(config Config) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "websearch", Version: "v1.0.0"}, nil)

	summarizer := websearch.NewSummarizer(
		config.OpenAIBaseURL,
		config.OpenAIAPIKey,
		config.SummarizerModel,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "websearch",
		Description: "Search the web and optionally get summaries of each web page from the results.",
	}, func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		input websearchToolInput,
	) (*mcp.CallToolResult, any, error) {
		return handleWebsearch(ctx, config, summarizer, input)
	})

	return server
}

func handleWebsearch(
	ctx context.Context,
	config Config,
	summarizer *websearch.Summarizer,
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
		limit = 8
	}
	maxSummaryTokens := input.MaxSummaryTokens
	if maxSummaryTokens <= 0 {
		maxSummaryTokens = 256
	}

	// Search using SearxNG
	searchStart := time.Now()
	results, err := websearch.Search(ctx, config.SearxngURL, input.Query, limit)
	searchDuration := time.Since(searchStart)
	config.Logger.Debug("SearxNG search", "query", input.Query, "duration_ms", searchDuration.Milliseconds())

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "search failed: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	var outputs []SearchResultOutput

	if input.Summary {
		// Process results concurrently with max 8 workers for summarization
		outputs = make([]SearchResultOutput, len(results))

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(8)

		for i, result := range results {
			i, result := i, result // capture loop variables
			g.Go(func() error {
				// Fetch content with 5 second timeout
				fetchCtx, cancel := context.WithTimeout(gctx, 5*time.Second)
				defer cancel()

				fetchStart := time.Now()
				content, err := websearch.Fetch(fetchCtx, result.URL, 5*time.Second)
				fetchDuration := time.Since(fetchStart)

				if err != nil {
					config.Logger.Debug("result fetch failed",
						"index", i+1,
						"url", result.URL,
						"fetch_ms", fetchDuration.Milliseconds(),
						"error", err)
					// Skip failed fetches, still return snippet
					outputs[i] = SearchResultOutput{
						Title:   result.Title,
						URL:     result.URL,
						Snippet: result.Content,
					}
					return nil
				}

				// Summarize the content
				summarizeStart := time.Now()
				summary, err := summarizer.Summarize(gctx, content, input.Query, result.URL, maxSummaryTokens)
				summarizeDuration := time.Since(summarizeStart)

				if err != nil {
					config.Logger.Debug("result summarize failed",
						"index", i+1,
						"url", result.URL,
						"fetch_ms", fetchDuration.Milliseconds(),
						"summarize_ms", summarizeDuration.Milliseconds(),
						"error", err)
					// Skip failed summarizations, still return snippet
					outputs[i] = SearchResultOutput{
						Title:   result.Title,
						URL:     result.URL,
						Snippet: result.Content,
					}
					return nil
				}

				config.Logger.Debug("result processed",
					"index", i+1,
					"url", result.URL,
					"fetch_ms", fetchDuration.Milliseconds(),
					"summarize_ms", summarizeDuration.Milliseconds(),
					"summary_len", len(summary))

				outputs[i] = SearchResultOutput{
					Title:   result.Title,
					URL:     result.URL,
					Snippet: result.Content,
					Summary: summary,
				}

				return nil
			})
		}

		// Wait for all goroutines to complete
		_ = g.Wait()
	} else {
		// No summarization - just return snippets from SearxNG
		for _, result := range results {
			outputs = append(outputs, SearchResultOutput{
				Title:   result.Title,
				URL:     result.URL,
				Snippet: result.Content,
			})
		}
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
