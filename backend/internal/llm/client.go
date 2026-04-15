package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/cwshea/localization/internal/config"
)

type Translator interface {
	Translate(ctx context.Context, text string, targetLanguage string) (string, error)
}

type ClientFactory struct {
	cfg *config.Config
}

func NewClientFactory(cfg *config.Config) *ClientFactory {
	return &ClientFactory{cfg: cfg}
}

func (f *ClientFactory) NewTranslator(provider string) (Translator, error) {
	switch provider {
	case "chatgpt5":
		if f.cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is not available (set OPENAI_API_KEY or configure GCP Secret Manager)")
		}
		model := "gpt-5"
		temperature := float32(1.0)
		timeout := 60 * time.Second
		if c := f.cfg.OpenAI; c != nil {
			model = c.LLM.Model
			temperature = c.LLM.Temperature
			timeout = c.TimeoutDuration()
		}
		return NewOpenAIClient(f.cfg.OpenAIAPIKey, model, temperature, timeout), nil
	case "gemini25":
		model := "gemini-2.5-pro"
		temperature := float32(1.0)
		timeout := 60 * time.Second
		if c := f.cfg.Gemini; c != nil {
			model = c.LLM.Model
			temperature = c.LLM.Temperature
			timeout = c.TimeoutDuration()
		}
		return NewGeminiClient(f.cfg.GCPProject, f.cfg.GCPLocation, model, temperature, timeout), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func buildPrompt(text, targetLanguage string) string {
	return fmt.Sprintf(
		"You are a professional translator. Translate the following American English text into %s. "+
			"Return ONLY the translated text, nothing else. Do not add quotes or explanations.\n\nText: %s",
		targetLanguage, text,
	)
}
