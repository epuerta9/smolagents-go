package models

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const defaultTimeout = 60 * time.Second

// OpenAIModel is a model that uses the OpenAI API.
type OpenAIModel struct {
	Model        string
	ApiKey       string
	MaxTokens    int
	Organization string
	Project      string
	client       *openai.Client
	httpClient   *http.Client // Store the HTTP client for use with the SDK
}

// NewOpenAIModel creates a new OpenAIModel.
func NewOpenAIModel(model string, options ...Option) *OpenAIModel {
	m := &OpenAIModel{
		Model:     model,
		MaxTokens: 1024,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	// Try to get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		m.ApiKey = apiKey
	}

	// Apply options
	for _, option := range options {
		option(m)
	}

	// Initialize the OpenAI client with options
	clientOptions := []option.RequestOption{}

	// Set API key if provided
	if m.ApiKey != "" {
		clientOptions = append(clientOptions, option.WithAPIKey(m.ApiKey))
	}

	// Set organization if provided
	if m.Organization != "" {
		clientOptions = append(clientOptions, option.WithHeader("OpenAI-Organization", m.Organization))
	}

	// Set project if provided
	if m.Project != "" {
		clientOptions = append(clientOptions, option.WithHeader("OpenAI-Project", m.Project))
	}

	// Set HTTP client if provided
	if m.httpClient != nil {
		clientOptions = append(clientOptions, option.WithHTTPClient(m.httpClient))
	}

	m.client = openai.NewClient(clientOptions...)

	return m
}

// Generate generates a response for the given messages.
func (m *OpenAIModel) Generate(ctx context.Context, messages []Message) (string, error) {
	return m.generateInternal(ctx, messages, nil)
}

// GenerateWithTools generates a response for the given messages with tools.
func (m *OpenAIModel) GenerateWithTools(ctx context.Context, messages []Message, tools []map[string]any) (string, error) {
	return m.generateInternal(ctx, messages, tools)
}

// generateInternal is the internal implementation of Generate and GenerateWithTools.
func (m *OpenAIModel) generateInternal(ctx context.Context, messages []Message, tools []map[string]any) (string, error) {
	if m.client == nil {
		return "", errors.New("OpenAI client not initialized")
	}

	// Convert our Message type to OpenAI's ChatCompletionMessageParamUnion
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range messages {
		switch msg.Role {
		case RoleSystem:
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		case RoleUser:
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case RoleAssistant:
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case RoleTool:
			chatMessages = append(chatMessages, openai.ToolMessage(msg.Name, msg.Content))
		}
	}

	// Prepare the completion parameters
	params := openai.ChatCompletionNewParams{
		Messages:  openai.F(chatMessages),
		Model:     openai.F(m.Model),
		MaxTokens: openai.F(int64(m.MaxTokens)),
	}

	// Add tools if provided
	if len(tools) > 0 {
		var toolsParam []openai.ChatCompletionToolParam
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

			description, ok := functionData["description"].(string)
			if !ok {
				continue
			}

			parameters, ok := functionData["parameters"].(map[string]any)
			if !ok {
				continue
			}

			// Create tool parameter
			toolsParam = append(toolsParam, openai.ChatCompletionToolParam{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.F(name),
					Description: openai.F(description),
					Parameters:  openai.F(openai.FunctionParameters(parameters)),
				}),
			})
		}
		params.Tools = openai.F(toolsParam)
	}

	// Make the API call with appropriate options
	var completion *openai.ChatCompletion
	var err error

	if len(tools) > 0 {
		// Only set tool_choice when tools are provided
		completion, err = m.client.Chat.Completions.New(
			ctx,
			params,
			option.WithJSONSet("tool_choice", "auto"),
		)
	} else {
		completion, err = m.client.Chat.Completions.New(ctx, params)
	}

	if err != nil {
		return "", err
	}

	// Handle the response
	if len(completion.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	choice := completion.Choices[0]

	// Check if there's a tool call
	if len(choice.Message.ToolCalls) > 0 {
		toolCall := choice.Message.ToolCalls[0]

		// Create a properly formatted tool call response
		toolResponse := map[string]any{
			"tool": toolCall.Function.Name,
			"args": json.RawMessage(toolCall.Function.Arguments),
		}

		toolResponseJSON, err := json.Marshal(toolResponse)
		if err != nil {
			return "", err
		}

		return string(toolResponseJSON), nil
	}

	return choice.Message.Content, nil
}

// WithOrganization sets the organization for OpenAI API requests.
func WithOrganization(org string) Option {
	return func(model any) {
		switch m := model.(type) {
		case *OpenAIModel:
			m.Organization = org
		}
	}
}

// WithProject sets the project for OpenAI API requests.
func WithProject(project string) Option {
	return func(model any) {
		switch m := model.(type) {
		case *OpenAIModel:
			m.Project = project
		}
	}
}
