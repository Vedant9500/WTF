package ai

import "context"

// Provider defines the interface for Generative AI providers
type Provider interface {
	// Name returns the name of the provider (e.g., "openai", "gemini", "ollama")
	Name() string

	// GenerateCommand generates a shell command based on a natural language prompt
	// prompt: The user's query (e.g., "how to resize an image")
	// sysCtx: System context (OS, shell, installed tools) to guide the AI
	GenerateCommand(ctx context.Context, prompt string, sysCtx string) (string, error)

	// ExplainCommand provides a detailed explanation of a shell command
	// command: The command to explain (e.g., "tar -czf archive.tar.gz folder")
	ExplainCommand(ctx context.Context, command string) (string, error)
}

// Config holds configuration for AI providers
type Config struct {
	Provider     string // "openai", "gemini", "ollama"
	APIKey       string // API Key for cloud providers
	Model        string // Model name (optional, uses defaults if empty)
	EndpointURL  string // Custom endpoint (mainly for Ollama)
}
