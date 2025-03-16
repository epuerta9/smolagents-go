package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// anthropicDefaultTimeout is used if no timeout is specified
const anthropicDefaultTimeout = 60 * time.Second

// AnthropicModel is a model that uses the Anthropic API.
type AnthropicModel struct {
	Model      string
	ApiKey     string
	MaxTokens  int
	client     *anthropic.Client
	httpClient *http.Client
}

// NewAnthropicModel creates a new AnthropicModel.
func NewAnthropicModel(model string, options ...Option) *AnthropicModel {
	m := &AnthropicModel{
		Model:     model,
		MaxTokens: 1024,
		httpClient: &http.Client{
			Timeout: anthropicDefaultTimeout,
		},
	}

	// Try to get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		m.ApiKey = apiKey
	}

	// Apply options
	for _, option := range options {
		option(m)
	}

	// Initialize the Anthropic client with options
	m.client = anthropic.NewClient(
		option.WithAPIKey(m.ApiKey),
		option.WithHTTPClient(m.httpClient),
	)

	return m
}

// Generate generates a response for the given messages.
func (m *AnthropicModel) Generate(ctx context.Context, messages []Message) (string, error) {
	return m.generateInternal(ctx, messages, nil)
}

// GenerateWithTools generates a response for the given messages with tools.
func (m *AnthropicModel) GenerateWithTools(ctx context.Context, messages []Message, tools []map[string]any) (string, error) {
	return m.generateInternal(ctx, messages, tools)
}

// generateInternal is the internal implementation of Generate and GenerateWithTools.
func (m *AnthropicModel) generateInternal(ctx context.Context, messages []Message, tools []map[string]any) (string, error) {
	if m.client == nil {
		return "", errors.New("anthropic client not initialized")
	}

	// Convert our Message type to Anthropic's format
	var systemPrompt string
	var anthropicMessages []anthropic.MessageParam

	for _, msg := range messages {
		// Handle system message separately
		if msg.Role == RoleSystem {
			systemPrompt = msg.Content
			continue
		}

		var anthropicMsg anthropic.MessageParam

		switch msg.Role {
		case RoleUser:
			anthropicMsg = anthropic.NewUserMessage(
				anthropic.NewTextBlock(msg.Content),
			)
		case RoleAssistant:
			anthropicMsg = anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(msg.Content),
			)
		case RoleTool:
			anthropicMsg = anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(msg.Content),
			)
		}

		anthropicMessages = append(anthropicMessages, anthropicMsg)
	}

	// Create the message request params
	params := anthropic.MessageNewParams{
		Model:     anthropic.F(m.Model),
		MaxTokens: anthropic.F(int64(m.MaxTokens)),
		Messages:  anthropic.F(anthropicMessages),
	}

	// Set system prompt if provided
	if systemPrompt != "" {
		// Create a text block for the system prompt
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		})
	}

	// Add tools if provided
	if len(tools) > 0 {
		var anthropicTools []anthropic.ToolUnionUnionParam
		for _, tool := range tools {
			// Extract tool properties
			functionData, ok := tool["function"].(map[string]any)
			if !ok {
				continue
			}

			name, ok := functionData["name"].(string)
			if !ok {
				continue
			}

			// Create tool parameter with required fields
			toolParam := anthropic.ToolParam{
				Name: anthropic.F(name),
			}

			// Add description if provided
			if description, ok := functionData["description"].(string); ok {
				toolParam.Description = anthropic.F(description)
			}

			var parameters map[string]any
			if params, ok := functionData["parameters"].(map[string]any); ok {
				parameters = params
			} else {
				parameters = map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				}
			}

			toolParam.InputSchema = anthropic.F(any(parameters))

			anthropicTools = append(anthropicTools, toolParam)
		}
		params.Tools = anthropic.F(anthropicTools)

	}

	// Make the API call
	resp, err := m.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create message: %w", err)
	}

	// Process the response
	if len(resp.Content) == 0 {
		return "", errors.New("empty response from model")
	}

	var content string

	for _, block := range resp.Content {
		switch b := block.AsUnion().(type) {
		case anthropic.TextBlock:
			// Append text from text blocks.
			content += b.Text
		case anthropic.ToolUseBlock:
			// Create a properly formatted tool call response for the agent
			toolResponse := map[string]any{
				"tool": b.Name,
				"args": b.Input,
			}
			toolResponseJSON, err := json.Marshal(toolResponse)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool response: %w", err)
			}
			content += "\r\n"
			content += string(toolResponseJSON)

		}
	}
	return content, nil
}

// WithAnthropicModel sets the model for Anthropic API requests.
func WithAnthropicModel(modelName string) Option {
	return func(model any) {
		switch m := model.(type) {
		case *AnthropicModel:
			m.Model = modelName
		}
	}
}
