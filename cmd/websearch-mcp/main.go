package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/cors"
)

// Config holds the configuration for the MCP server.
type Config struct {
	SearxngURL string
	IsHTTP     bool
	Port       string
	Logger     *slog.Logger
}

func main() {
	// Define CLI flags
	searxngURL := flag.String("searxng-url", "", "Base URL of the SearxNG instance (required)")
	isHTTP := flag.Bool("http", false, "Run as HTTP server instead of stdio")
	port := flag.String("port", "8080", "Port for HTTP mode")
	debug := flag.Bool("debug", false, "Enable debug output with timing information")

	flag.Parse()

	// Validate required flags
	if *searxngURL == "" {
		fmt.Fprintln(os.Stderr, "error: -searxng-url is required")
		os.Exit(1)
	}

	// Build config
	var logLevel slog.Level
	if *debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	var logOutput io.Writer
	if *isHTTP {
		logOutput = os.Stdout
	} else {
		logOutput = os.Stderr
	}

	config := Config{
		SearxngURL: *searxngURL,
		IsHTTP:     *isHTTP,
		Port:       *port,
		Logger: slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String(slog.TimeKey, a.Value.Time().Format("15:04:05"))
				}
				return a
			},
		})),
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
