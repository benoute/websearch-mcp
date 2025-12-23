package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/cors"
)

// Config holds the configuration for the MCP server.
type Config struct {
	SearxngURL      string
	OpenAIBaseURL   string
	SummarizerModel string
	OpenAIAPIKey    string
	IsHTTP          bool
	Port            string
}

func main() {
	// Define CLI flags
	searxngURL := flag.String("searxng-url", "", "Base URL of the SearxNG instance (required)")
	openaiBaseURL := flag.String("openai-base-url", "", "Base URL of the OpenAI-compatible provider API (required)")
	summarizerModel := flag.String("summarizer-model", "", "Model name to use for summarization (required)")
	openaiAPIKeyEnv := flag.String("openai-api-key-env", "OPENAI_API_KEY", "Environment variable name for the OpenAI API key")
	isHTTP := flag.Bool("http", false, "Run as HTTP server instead of stdio")
	port := flag.String("port", "8080", "Port for HTTP mode")

	flag.Parse()

	// Validate required flags
	if *searxngURL == "" {
		fmt.Fprintln(os.Stderr, "error: -searxng-url is required")
		os.Exit(1)
	}
	if *openaiBaseURL == "" {
		fmt.Fprintln(os.Stderr, "error: -openai-base-url is required")
		os.Exit(1)
	}
	if *summarizerModel == "" {
		fmt.Fprintln(os.Stderr, "error: -summarizer-model is required")
		os.Exit(1)
	}

	// Get and validate the OpenAI API key from environment
	openaiAPIKey := os.Getenv(*openaiAPIKeyEnv)
	if openaiAPIKey == "" {
		fmt.Fprintf(os.Stderr, "error: environment variable %s must be set and non-empty\n", *openaiAPIKeyEnv)
		os.Exit(1)
	}

	// Build config
	config := Config{
		SearxngURL:      *searxngURL,
		OpenAIBaseURL:   *openaiBaseURL,
		SummarizerModel: *summarizerModel,
		OpenAIAPIKey:    openaiAPIKey,
		IsHTTP:          *isHTTP,
		Port:            *port,
	}

	// Setup MCP server
	server := setupMCPServer(config)

	// Run the server
	if config.IsHTTP {
		// HTTP mode with CORS support
		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return server
		}, nil)

		corsHandler := cors.AllowAll().Handler(handler)

		addr := ":" + config.Port
		log.Printf("Starting HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, corsHandler); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	} else {
		// Stdio mode
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
