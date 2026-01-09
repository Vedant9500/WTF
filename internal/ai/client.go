package ai

import (
	"fmt"
	"os"
	"strings"
)

// NewClient creates a new AI provider based on configuration
func NewClient(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("API key is required for OpenAI provider")
		}
		return NewOpenAIProvider(cfg.APIKey, cfg.Model), nil
	case "gemini":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("API key is required for Gemini provider")
		}
		return NewGeminiProvider(cfg.APIKey, cfg.Model), nil
	case "ollama":
		return NewOllamaProvider(cfg.EndpointURL, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}

// GetConfigFromEnv attempts to load AI configuration from environment variables
func GetConfigFromEnv() Config {
	return Config{
		Provider:    os.Getenv("WTF_AI_PROVIDER"),
		APIKey:      os.Getenv("WTF_AI_API_KEY"),
		Model:       os.Getenv("WTF_AI_MODEL"),
		EndpointURL: os.Getenv("WTF_AI_ENDPOINT"),
	}
}
