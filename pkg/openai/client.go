package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cuichanghe/daily-reminder-bot/pkg/logger"
	"go.uber.org/zap"
)

// Client is an OpenAI-compatible API client
type Client struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

// NewClient creates a new OpenAI-compatible API client
func NewClient(apiKey, baseURL, model string, maxTokens int, temperature float64, timeout time.Duration) *Client {
	return &Client{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		client:      &http.Client{Timeout: timeout},
	}
}

// ChatCompletion sends a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (*ChatCompletionResponse, error) {
	logger.Debug("OpenAI.ChatCompletion called",
		zap.String("model", c.model),
		zap.Int("message_count", len(messages)),
		zap.String("base_url", c.baseURL))
	start := time.Now()

	reqBody := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
	}

	logger.Debug("Request payload",
		zap.Int("max_tokens", c.maxTokens),
		zap.Float64("temperature", c.temperature))

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Failed to marshal request",
			zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create request",
			zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	logger.Debug("Sending HTTP request",
		zap.String("url", url),
		zap.String("method", "POST"))

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error("HTTP request failed",
			zap.String("url", url),
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logger.Debug("HTTP response received",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)))

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		logger.Error("Failed to decode response",
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		logger.Error("API returned error",
			zap.String("error_message", chatResp.Error.Message),
			zap.String("error_type", chatResp.Error.Type))
		return nil, fmt.Errorf("API error: %s (type: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("API returned non-OK status",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Log token usage if available
	if chatResp.Usage.TotalTokens > 0 {
		logger.Debug("Token usage",
			zap.Int("prompt_tokens", chatResp.Usage.PromptTokens),
			zap.Int("completion_tokens", chatResp.Usage.CompletionTokens),
			zap.Int("total_tokens", chatResp.Usage.TotalTokens))
	}

	logger.Info("ChatCompletion successful",
		zap.String("model", c.model),
		zap.Duration("duration", time.Since(start)))

	return &chatResp, nil
}

// GetContent is a convenience method that returns the generated content directly
func (c *Client) GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger.Debug("OpenAI.GetContent called",
		zap.Int("system_prompt_len", len(systemPrompt)),
		zap.Int("user_prompt_len", len(userPrompt)))

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	resp, err := c.ChatCompletion(ctx, messages)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		logger.Warn("No choices in response")
		return "", fmt.Errorf("no choices in response")
	}

	logger.Debug("Content generated",
		zap.Int("content_len", len(resp.Choices[0].Message.Content)))
	return resp.Choices[0].Message.Content, nil
}
