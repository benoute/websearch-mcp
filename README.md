# websearch-mcp

MCP server that searches the web using SearxNG and provides AI-generated summaries of the results.

## Features

- Searches the web using a SearxNG instance
- Fetches content from search results
- Generates summaries using any OpenAI-compatible LLM API
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
  -llm-base-url "https://openrouter.ai/api/v1" \
  -llm-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

## CLI Options

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-searxng-url` | Yes | - | Base URL of the SearxNG instance |
| `-llm-base-url` | Yes | - | Base URL of the OpenAI-compatible LLM API |
| `-llm-model` | Yes | - | Model name to use for summarization |
| `-llm-api-key-env` | No | `LLM_API_KEY` | Environment variable name for the LLM API key |
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
        "-llm-base-url", "https://openrouter.ai/api/v1",
        "-llm-model", "nvidia/nemotron-3-nano-30b-a3b:free"
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
  -llm-base-url "https://openrouter.ai/api/v1" \
  -llm-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

Then configure your MCP client:

```json
{
  "url": "http://localhost:8080/mcp"
}
```

## Tool: `websearch`

Searches the web using SearxNG and returns AI-generated summaries of the results.

### Input

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | The search query |
| `limit` | int | No | `8` | Maximum number of search results |
| `max_tokens` | int | No | `5000` | Maximum tokens per summary |

### Example Input

```json
{
  "query": "golang concurrency patterns",
  "limit": 5,
  "max_tokens": 1000
}
```

### Output

JSON array of results:

```json
[
  {
    "title": "Concurrency in Go - A Practical Guide",
    "url": "https://example.com/go-concurrency",
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
  -llm-base-url "https://openrouter.ai/api/v1" \
  -llm-model "nvidia/nemotron-3-nano-30b-a3b:free"
```

### LiteLLM Proxy

```bash
export LITELLM_API_KEY="sk-..."
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -llm-base-url "http://localhost:4000/v1" \
  -llm-model "gpt-4o" \
  -llm-api-key-env "LITELLM_API_KEY"
```

### Ollama (local)

```bash
export LLM_API_KEY="ollama"  # Required but unused
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -llm-base-url "http://localhost:11434/v1" \
  -llm-model "llama3"
```

### OpenAI

```bash
export OPENAI_API_KEY="sk-..."
./websearch-mcp \
  -searxng-url "http://localhost:8080" \
  -llm-base-url "https://api.openai.com/v1" \
  -llm-model "gpt-4o-mini" \
  -llm-api-key-env "OPENAI_API_KEY"
```

## Requirements

- A running SearxNG instance with JSON output enabled
- An OpenAI-compatible LLM API (OpenRouter, LiteLLM, Ollama, OpenAI, etc.)
- API key for the LLM provider (set via environment variable)

## How It Works

1. Receives a search query via MCP
2. Queries SearxNG for search results
3. For each result (up to `limit`, concurrently):
   - Fetches the page content (5 second timeout)
   - Sends content to the LLM for summarization
   - Collects successful results
4. Returns JSON array of titles, URLs, and summaries

Results that fail to fetch or summarize are silently skipped.
