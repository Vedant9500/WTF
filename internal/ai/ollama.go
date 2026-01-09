package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOllamaURL   = "http://localhost:11434/api/generate"
	defaultOllamaModel = "llama3"
)

type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

func NewOllamaProvider(endpoint string, model string) *OllamaProvider {
	if endpoint == "" {
		endpoint = defaultOllamaURL
	}
	if model == "" {
		model = defaultOllamaModel
	}
	return &OllamaProvider{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 60 * time.Second}, // Longer timeout for local inference
	}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) GenerateCommand(ctx context.Context, prompt string, sysCtx string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a command-line expert helper.
Your task is to provide a single, valid shell command that performs the user's request.
Context: %s
Rules:
1. Return ONLY the command. No markdown, no explanations, no code blocks.
2. If multiple steps are needed, combine them with && or | as appropriate.
3. If the request is dangerous, return a comment starting with # explaining why.`, sysCtx)

	return p.generate(ctx, systemPrompt, prompt)
}

func (p *OllamaProvider) ExplainCommand(ctx context.Context, command string) (string, error) {
	systemPrompt := `You are a command-line expert.
Your task is to explain the given shell command clearly and concisely.
Break down the flags and arguments.
Keep the explanation under 5 lines if possible.`

	return p.generate(ctx, systemPrompt, command)
}

func (p *OllamaProvider) generate(ctx context.Context, systemMsg, userMsg string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  p.model,
		"prompt": userMsg,
		"system": systemMsg,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("provider error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Response), nil
}
