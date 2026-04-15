package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
)

type GeminiClient struct {
	project     string
	location    string
	model       string
	temperature float32
	timeout     time.Duration
}

func NewGeminiClient(project, location, model string, temperature float32, timeout time.Duration) *GeminiClient {
	return &GeminiClient{
		project:     project,
		location:    location,
		model:       model,
		temperature: temperature,
		timeout:     timeout,
	}
}

func (c *GeminiClient) Translate(ctx context.Context, text string, targetLanguage string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Uses Application Default Credentials (ADC) via Vertex AI backend
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  c.project,
		Location: c.location,
	})
	if err != nil {
		return "", fmt.Errorf("creating gemini client: %w", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText(buildPrompt(text, targetLanguage), genai.RoleUser),
	}
	temp := c.temperature
	result, err := client.Models.GenerateContent(ctx, c.model, contents, &genai.GenerateContentConfig{
		Temperature: &temp,
	})
	if err != nil {
		return "", fmt.Errorf("gemini generate content: %w", err)
	}

	return strings.TrimSpace(result.Text()), nil
}
