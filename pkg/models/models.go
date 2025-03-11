// Package models provides interfaces and implementations for language models.
package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MessageRole represents the role of a message.
type MessageRole string

const (
	// RoleSystem is the role for system messages.
	RoleSystem MessageRole = "system"
	// RoleUser is the role for user messages.
	RoleUser MessageRole = "user"
	// RoleAssistant is the role for assistant messages.
	RoleAssistant MessageRole = "assistant"
	// RoleTool is the role for tool messages.
	RoleTool MessageRole = "tool"
)

// Message represents a chat message.
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
	Name    string      `json:"name,omitempty"`
}

// Model represents a language model that can generate responses.
type Model interface {
	// Generate generates a response for the given messages.
	Generate(ctx context.Context, messages []Message) (string, error)

	// GenerateWithTools generates a response for the given messages,
	// with the tools provided as JSON schema.
	GenerateWithTools(ctx context.Context, messages []Message, tools []map[string]any) (string, error)
}

// HfApiModel is a model that uses the Hugging Face Inference API.
type HfApiModel struct {
	Model     string
	ApiKey    string
	ApiURL    string
	MaxTokens int
	Client    *http.Client
}

// Option is a functional option for configuring a model.
type Option func(model any)

// WithMaxTokens sets the maximum number of tokens to generate.
func WithMaxTokens(maxTokens int) Option {
	return func(model any) {
		switch m := model.(type) {
		case *HfApiModel:
			m.MaxTokens = maxTokens
		case *OpenAIModel:
			m.MaxTokens = maxTokens
		}
	}
}

// WithApiKey sets the API key to use for authentication.
func WithApiKey(apiKey string) Option {
	return func(model any) {
		switch m := model.(type) {
		case *HfApiModel:
			m.ApiKey = apiKey
		case *OpenAIModel:
			m.ApiKey = apiKey
		}
	}
}

// WithHttpClient sets the HTTP client to use for API requests.
func WithHttpClient(client *http.Client) Option {
	return func(model any) {
		switch m := model.(type) {
		case *HfApiModel:
			m.Client = client
		case *OpenAIModel:
			m.httpClient = client
		}
	}
}

// NewHfApiModel creates a new HfApiModel.
func NewHfApiModel(model string, options ...Option) *HfApiModel {
	m := &HfApiModel{
		Model:     model,
		ApiURL:    "https://api-inference.huggingface.co/models",
		MaxTokens: 1024,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, option := range options {
		option(m)
	}

	return m
}

// Generate generates a response for the given messages.
func (m *HfApiModel) Generate(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to the format expected by the API
	payload := map[string]any{
		"inputs": messages,
		"parameters": map[string]any{
			"max_new_tokens":   m.MaxTokens,
			"return_full_text": false,
		},
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s", m.ApiURL, m.Model),
		strings.NewReader(string(jsonPayload)),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if m.ApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.ApiKey))
	}

	// Send request
	resp, err := m.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var result []struct {
		GeneratedText string `json:"generated_text"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	if len(result) == 0 {
		return "", errors.New("empty response from model")
	}

	return result[0].GeneratedText, nil
}

// GenerateWithTools generates a response for the given messages,
// with the tools provided as JSON schema.
func (m *HfApiModel) GenerateWithTools(
	ctx context.Context,
	messages []Message,
	tools []map[string]any,
) (string, error) {
	// Convert messages to the format expected by the API
	payload := map[string]any{
		"inputs": messages,
		"parameters": map[string]any{
			"max_new_tokens":   m.MaxTokens,
			"return_full_text": false,
			"tools":            tools,
		},
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s", m.ApiURL, m.Model),
		strings.NewReader(string(jsonPayload)),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if m.ApiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.ApiKey))
	}

	// Send request
	resp, err := m.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var result []struct {
		GeneratedText string `json:"generated_text"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	if len(result) == 0 {
		return "", errors.New("empty response from model")
	}

	return result[0].GeneratedText, nil
}
