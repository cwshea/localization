package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LLMConfig struct {
	LLM struct {
		Provider    string  `yaml:"provider"`
		Model       string  `yaml:"model"`
		Temperature float32 `yaml:"temperature"`
		Timeout     float64 `yaml:"timeout"`
	} `yaml:"llm"`
	Secrets struct {
		Source string `yaml:"source"`
		UseADC bool   `yaml:"use_adc"`
		GCP    struct {
			ProjectID  string `yaml:"project_id"`
			SecretName string `yaml:"secret_name"`
			Version    string `yaml:"version"`
		} `yaml:"gcp"`
	} `yaml:"secrets"`
}

func (c *LLMConfig) TimeoutDuration() time.Duration {
	if c.LLM.Timeout > 0 {
		return time.Duration(c.LLM.Timeout * float64(time.Second))
	}
	return 60 * time.Second
}

func LoadLLMConfig(path string) (*LLMConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg LLMConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &cfg, nil
}
