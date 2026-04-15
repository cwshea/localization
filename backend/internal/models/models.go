package models

import (
	"time"

	"github.com/google/uuid"
)

type SourceString struct {
	ID           uuid.UUID     `json:"id"`
	Text         string        `json:"text"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Translations []Translation `json:"translations"`
}

type Translation struct {
	ID             uuid.UUID `json:"id"`
	SourceID       uuid.UUID `json:"source_id"`
	Locale         string    `json:"locale"`
	TranslatedText string    `json:"translated_text"`
	LLMProvider    string    `json:"llm_provider"`
	TranslatedAt   time.Time `json:"translated_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Text         string   `json:"text"`
	Locales      []string `json:"locales"`
	LLMProviders []string `json:"llm_providers"`
}

type UpdateSourceRequest struct {
	Text string `json:"text"`
}

type UpdateTranslationRequest struct {
	TranslatedText string `json:"translated_text"`
}

type RetranslateRequest struct {
	Locales      []string `json:"locales"`
	LLMProviders []string `json:"llm_providers"`
}

var ValidLocales = map[string]string{
	"en-GB":   "British English",
	"es":      "Spanish",
	"zh-Hant": "Traditional Chinese",
	"zh-Hans": "Simplified Chinese",
	"hi":      "Hindi",
}

var ValidProviders = map[string]bool{
	"chatgpt5": true,
	"gemini25": true,
}
