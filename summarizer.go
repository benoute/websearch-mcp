package websearch

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

// Summarizer is an OpenAI-compatible LLM client for summarizing web content.
type Summarizer struct {
	client *openai.Client
	model  string
}

// headerTransport is a custom http.RoundTripper that adds custom headers to requests.
type headerTransport struct {
	base http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("HTTP-Referer", "https://github.com/benoute/websearch-mcp")
	req.Header.Set("X-Title", "websearch-mcp")
	return t.base.RoundTrip(req)
}

// NewSummarizer creates a new Summarizer with an OpenAI-compatible client.
func NewSummarizer(baseURL, apiKey, model string) *Summarizer {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	config.HTTPClient = &http.Client{
		Transport: &headerTransport{base: http.DefaultTransport},
	}

	return &Summarizer{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}
}

// Summarize creates a summary of the given content relevant to the search query.
func (s *Summarizer) Summarize(ctx context.Context, content, query, url string, maxTokens int) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a helpful assistant that summarizes web content. Given web page content and a search query, provide a concise and relevant summary that focuses on information related to the query. Be factual and informative.

You MUST keep your summary under %d tokens. Do not prefix your response with "Summary" or any similar heading, as your output will be placed directly under a 'summary' field.`, maxTokens)

	userMessage := fmt.Sprintf("Search query: %s\n\nURL: %s\n\nContent to summarize:\n%s", query, url, content)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		MaxTokens: maxTokens * 5 / 4,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}
