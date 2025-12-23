# websearch-mcp

MCP server that searches the web using SearxNG.

## Features

- Searches the web using a SearxNG instance
- Returns search result snippets from SearxNG
- Configurable result limit

## Quick Start

### Build from Source

Requires Go 1.23+

```bash
go build -o websearch-mcp ./cmd/websearch-mcp
```

### Run

```bash
./websearch-mcp -searxng-url "http://localhost:8080"
```

## CLI Options

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-searxng-url` | Yes | - | Base URL of the SearxNG instance |
| `-http` | No | `false` | Run as HTTP server instead of stdio |
| `-port` | No | `8080` | Port for HTTP mode |
| `-debug` | No | `false` | Enable debug output with timing information |

## MCP Configuration

### Stdio Mode (default)

```json
{
  "mcpServers": {
    "websearch": {
      "command": "/path/to/websearch-mcp",
      "args": [
        "-searxng-url", "http://localhost:8080"
      ]
    }
  }
}
```

### HTTP Mode

Start the server:

```bash
./websearch-mcp \
  -http \
  -port 8080 \
  -searxng-url "http://localhost:8081"
```

Then configure your MCP client:

```json
{
  "url": "http://localhost:8080/mcp"
}
```

## Tool: `websearch`

Searches the web using SearxNG.

### Input

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | The search query |
| `limit` | int | No | `10` | Maximum number of search results |

### Example Input

```json
{
  "query": "golang concurrency patterns",
  "limit": 5
}
```

### Output

JSON array of results:

```json
[
  {
    "title": "Concurrency in Go - A Practical Guide",
    "url": "https://example.com/go-concurrency",
    "snippet": "Learn about goroutines, channels, and common concurrency patterns...",
    "engines": ["google", "bing"],
    "score": 2.5
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | The page title |
| `url` | string | The page URL |
| `snippet` | string | A short snippet/description from the search result |
| `engines` | string[] | List of search engines that returned this result |
| `score` | number | SearxNG's relevance score for the result |

## Requirements

- A running SearxNG instance with JSON output enabled

## How It Works

1. Receives a search query via MCP
2. Queries SearxNG for search results
3. Returns titles, URLs, and snippets from SearxNG
