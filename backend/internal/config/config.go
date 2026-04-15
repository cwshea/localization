package config

import (
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Config struct {
	Port         string
	DatabaseURL  string
	OpenAIAPIKey string
	GCPProject   string
	GCPLocation  string
	OpenAI       *LLMConfig
	Gemini       *LLMConfig
}

func Load() *Config {
	cfgDir := getEnv("CONFIG_DIR", ".")

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localization:localization@localhost:5432/localization?sslmode=disable"),
	}

	// Load LLM configs from YAML files
	openaiCfg, err := LoadLLMConfig(cfgDir + "/gpt-5.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load gpt-5.yaml: %v\n", err)
	}
	cfg.OpenAI = openaiCfg

	geminiCfg, err := LoadLLMConfig(cfgDir + "/gemini.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load gemini.yaml: %v\n", err)
	}
	cfg.Gemini = geminiCfg

	// OpenAI key: prefer env var, fall back to GCP Secret Manager using YAML config
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg.OpenAIAPIKey = key
	} else if openaiCfg != nil && openaiCfg.Secrets.Source == "gcp-secret" {
		key, err := fetchGCPSecret(
			openaiCfg.Secrets.GCP.ProjectID,
			openaiCfg.Secrets.GCP.SecretName,
			openaiCfg.Secrets.GCP.Version,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch OpenAI key from GCP Secret Manager: %v\n", err)
		} else {
			cfg.OpenAIAPIKey = key
		}
	}

	// Gemini: GCP project/location for Vertex AI ADC
	defaultProject := "prj-dm-ds-v0-pdigi-dev-5gf"
	if openaiCfg != nil && openaiCfg.Secrets.GCP.ProjectID != "" {
		defaultProject = openaiCfg.Secrets.GCP.ProjectID
	}
	cfg.GCPProject = getEnv("GCP_PROJECT", defaultProject)
	cfg.GCPLocation = getEnv("GCP_LOCATION", "us-central1")

	return cfg
}

func fetchGCPSecret(projectID, secretName, version string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating secret manager client: %w", err)
	}
	defer client.Close()

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretName, version)
	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("accessing secret %s: %w", name, err)
	}

	return string(result.Payload.Data), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
