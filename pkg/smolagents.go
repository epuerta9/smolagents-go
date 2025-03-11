// Package smolagents provides a framework for creating AI agents with tool calling capabilities.
// It is inspired by and aims to be compatible with Python's smolagents library.
package smolagents

import (
	"github.com/epuerta9/smolagents-go/pkg/agents"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// Re-export main types for easier access
type (
	// Agent is the main interface for agents
	Agent = agents.Agent

	// CodeAgent specializes in executing code
	CodeAgent = agents.CodeAgent

	// ToolCallingAgent specializes in tool calling
	ToolCallingAgent = agents.ToolCallingAgent

	// Tool represents a function that can be called by an agent
	Tool = tools.Tool

	// Model represents an LLM that can be used by agents
	Model = models.Model

	// HfApiModel is a model that uses the Hugging Face API
	HfApiModel = models.HfApiModel

	// OpenAIModel is a model that uses the OpenAI API
	OpenAIModel = models.OpenAIModel
)

// Version of the package
const Version = "0.1.0"

// CreateCodeAgent creates a new CodeAgent with the given tools and model
func CreateCodeAgent(tools []tools.Tool, model models.Model, opts ...agents.Option) (*agents.CodeAgent, error) {
	return agents.NewCodeAgent(tools, model, opts...)
}

// CreateToolCallingAgent creates a new ToolCallingAgent with the given tools and model
func CreateToolCallingAgent(tools []tools.Tool, model models.Model, opts ...agents.Option) (*agents.ToolCallingAgent, error) {
	return agents.NewToolCallingAgent(tools, model, opts...)
}

// Functions for creating and configuring tools and models
// Re-export these for easier access

// CreateTool is a decorator-style function that creates a new Tool
// Example:
//
//	GetWeather := smolagents.CreateTool("get_weather", "Get the current weather")(func(location string) string {
//	    // implementation
//	    return fmt.Sprintf("The weather in %s is sunny", location)
//	})
func CreateTool[F any](name, description string) func(F) *tools.FunctionTool[F] {
	return tools.CreateTool[F](name, description)
}

// WithMaxTokens is an option to configure the maximum number of tokens to generate
func WithMaxTokens(maxTokens int) models.Option {
	return models.WithMaxTokens(maxTokens)
}

// WithApiKey is an option to configure the API key for models
func WithApiKey(apiKey string) models.Option {
	return models.WithApiKey(apiKey)
}

// WithOrganization is an option to configure the organization for OpenAI models
func WithOrganization(org string) models.Option {
	return models.WithOrganization(org)
}

// WithProject is an option to configure the project for OpenAI models
func WithProject(project string) models.Option {
	return models.WithProject(project)
}

// WithMaxSteps is an option to configure the maximum number of steps for agents
func WithMaxSteps(maxSteps int) agents.Option {
	return agents.WithMaxSteps(maxSteps)
}

// WithSystemPrompt is an option to configure the system prompt for agents
func WithSystemPrompt(systemPrompt string) agents.Option {
	return agents.WithSystemPrompt(systemPrompt)
}
