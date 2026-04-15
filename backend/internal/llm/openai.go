package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client      *openai.Client
	model       string
	temperature float32
	timeout     time.Duration
}

func NewOpenAIClient(apiKey, model string, temperature float32, timeout time.Duration) *OpenAIClient {
	return &OpenAIClient{
		client:      openai.NewClient(apiKey),
		model:       model,
		temperature: temperature,
		timeout:     timeout,
	}
}

func (c *OpenAIClient) Translate(ctx context.Context, text string, targetLanguage string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: buildPrompt(text, targetLanguage),
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		return "", fmt.Errorf("openai chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
