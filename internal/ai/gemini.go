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

// Gemini API URL pattern
const (
	geminiGenerateURLPattern = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
	defaultGeminiModel       = "gemini-1.5-flash"
)

type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiProvider(apiKey string, model string) *GeminiProvider {
	if model == "" {
		model = defaultGeminiModel
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) GenerateCommand(ctx context.Context, prompt string, sysCtx string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a command-line expert helper.
Your task is to provide a single, valid shell command that performs the user's request.
Context: %s
Rules:
1. Return ONLY the command. No markdown, no explanations, no code blocks.
2. If multiple steps are needed, combine them with && or | as appropriate.
3. If the request is dangerous, return a comment starting with # explaining why.`, sysCtx)

	return p.generateContent(ctx, systemPrompt, prompt)
}

func (p *GeminiProvider) ExplainCommand(ctx context.Context, command string) (string, error) {
	systemPrompt := `You are a command-line expert.
Your task is to explain the given shell command clearly and concisely.
Break down the flags and arguments.
Keep the explanation under 5 lines if possible.`

	return p.generateContent(ctx, systemPrompt, command)
}

func (p *GeminiProvider) generateContent(ctx context.Context, systemMsg, userMsg string) (string, error) {
	url := fmt.Sprintf(geminiGenerateURLPattern, p.model, p.apiKey)

	// Construct request body for Gemini API
	// Note: Gemini doesn't always support "system" role in the same way as OpenAI in struct,
	// but 1.5 models support system instructions or we can prepending it.
	// For simplicity and compatibility, we'll prepend system message to user message logic or use parts.
	
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": systemMsg + "\n\nUser Request: " + userMsg},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
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
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from provider")
	}

	return strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text), nil
}
