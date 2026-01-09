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
	openAIGenerateURL = "https://api.openai.com/v1/chat/completions"
	defaultOpenAIModel = "gpt-3.5-turbo"
)

type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIProvider(apiKey string, model string) *OpenAIProvider {
	if model == "" {
		model = defaultOpenAIModel
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GenerateCommand(ctx context.Context, prompt string, sysCtx string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a command-line expert helper.
Your task is to provide a single, valid shell command that performs the user's request.
Context: %s
Rules:
1. Return ONLY the command. No markdown, no explanations, no code blocks.
2. If multiple steps are needed, combine them with && or | as appropriate.
3. If the request is dangerous, return a comment starting with # explaining why.`, sysCtx)

	return p.chatCompletion(ctx, systemPrompt, prompt)
}

func (p *OpenAIProvider) ExplainCommand(ctx context.Context, command string) (string, error) {
	systemPrompt := `You are a command-line expert.
Your task is to explain the given shell command clearly and concisely.
Break down the flags and arguments.
Keep the explanation under 5 lines if possible.`

	return p.chatCompletion(ctx, systemPrompt, command)
}

func (p *OpenAIProvider) chatCompletion(ctx context.Context, systemMsg, userMsg string) (string, error) {
	reqBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemMsg},
			{"role": "user", "content": userMsg},
		},
		"temperature": 0.2, // Low temperature for deterministic code
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIGenerateURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from provider")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
