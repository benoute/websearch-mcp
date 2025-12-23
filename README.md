# websearch-mcp

MCP server that searches the web using SearxNG and optionally provides AI-generated summaries of the results.

## Features

- Searches the web using a SearxNG instance
- Returns search result snippets from SearxNG
- Optional AI-powered summarization of page content using any OpenAI-compatible LLM API
- Configurable result limit and summary token count
- Concurrent fetching and summarization (up to 8 parallel requests)

## Quick Start

### Build from Source

Requires Go 1.23+

```bash
go build -o websearch-mcp ./cmd/websearch-mcp
```

### Run

```bash
# Set your LLM API key
export LLM_API_KEY="your-api-key"

# Run with OpenRouter
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -openai-base-url "https://openrouter.ai/api/v1" \
  -summarizer-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

## CLI Options

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-searxng-url` | Yes | - | Base URL of the SearxNG instance |
| `-openai-base-url` | Yes | - | Base URL of the OpenAI-compatible LLM API |
| `-summarizer-model` | Yes | - | Model name to use for summarization |
| `-openai-api-key-env` | No | `LLM_API_KEY` | Environment variable name for the LLM API key |
| `-http` | No | `false` | Run as HTTP server instead of stdio |
| `-port` | No | `8080` | Port for HTTP mode |

## MCP Configuration

### Stdio Mode (default)

```json
{
  "mcpServers": {
    "websearch": {
      "command": "/path/to/websearch-mcp",
      "args": [
        "-searxng-url", "http://localhost:8080",
        "-openai-base-url", "https://openrouter.ai/api/v1",
        "-summarizer-model", "nvidia/nemotron-3-nano-30b-a3b:free"
      ],
      "env": {
        "LLM_API_KEY": "your-api-key"
      }
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
  -searxng-url "http://localhost:8081" \
  -openai-base-url "https://openrouter.ai/api/v1" \
  -summarizer-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

Then configure your MCP client:

```json
{
  "url": "http://localhost:8080/mcp"
}
```

## Tool: `websearch`

Searches the web using SearxNG and optionally returns AI-generated summaries of the results.

### Input

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | The search query |
| `limit` | int | No | `8` | Maximum number of search results |
| `summary` | bool | No | `false` | Enable summarization of each web page from the results |
| `max_summary_tokens` | int | No | `256` | Maximum tokens per summary |

### Example Input (without summarization)

```json
{
  "query": "golang concurrency patterns",
  "limit": 5
}
```

### Example Input (with summarization)

```json
{
  "query": "golang concurrency patterns",
  "limit": 5,
  "summary": true,
  "max_summary_tokens": 512
}
```

### Output

JSON array of results. When `summary` is false (default), only `snippet` is returned:

```json
[
  {
    "title": "Concurrency in Go - A Practical Guide",
    "url": "https://example.com/go-concurrency",
    "snippet": "Learn about goroutines, channels, and common concurrency patterns..."
  }
]
```

When `summary` is true, results include both `snippet` and `summary`:

```json
[
  {
    "title": "Concurrency in Go - A Practical Guide",
    "url": "https://example.com/go-concurrency",
    "snippet": "Learn about goroutines, channels, and common concurrency patterns...",
    "summary": "This article covers goroutines, channels, and common concurrency patterns in Go including worker pools, fan-out/fan-in, and pipeline patterns..."
  }
]
```

## LLM Provider Examples

### OpenRouter

```bash
export LLM_API_KEY="sk-or-..."
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -openai-base-url "https://openrouter.ai/api/v1" \
  -summarizer-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

### LiteLLM Proxy

```bash
export LITELLM_API_KEY="sk-..."
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -openai-base-url "http://localhost:4000/v1" \
  -summarizer-model "gpt-4o" \
  -openai-api-key-env "LITELLM_API_KEY"
```

### Ollama (local)

```bash
export LLM_API_KEY="ollama"  # Required but unused
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -openai-base-url "http://localhost:11434/v1" \
  -summarizer-model "llama3"
```

### OpenAI

```bash
export OPENAI_API_KEY="sk-..."
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -openai-base-url "https://api.openai.com/v1" \
  -summarizer-model "gpt-4o-mini" \
  -openai-api-key-env "OPENAI_API_KEY"
```

## Requirements

- A running SearxNG instance with JSON output enabled
- An OpenAI-compatible LLM API (OpenRouter, LiteLLM, Ollama, OpenAI, etc.)
- API key for the LLM provider (set via environment variable)

## How It Works

1. Receives a search query via MCP
2. Queries SearxNG for search results
3. Returns titles, URLs, and snippets from SearxNG
4. If `summary` is enabled, for each result (up to `limit`, concurrently):
   - Fetches the page content (5 second timeout)
   - Sends content to the LLM for summarization
   - Returns both snippets and summaries

Results that fail to fetch or summarize still return snippets from SearxNG.
