// Package agents provides the core agent implementations.
package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/epuerta9/smolagents-go/pkg/memory"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// Option is a functional option for configuring an agent.
type Option func(a *BaseAgent) error

// WithMaxSteps sets the maximum number of steps for the agent.
func WithMaxSteps(maxSteps int) Option {
	return func(a *BaseAgent) error {
		if maxSteps <= 0 {
			return errors.New("maxSteps must be greater than 0")
		}
		a.maxSteps = maxSteps
		return nil
	}
}

// WithSystemPrompt sets the system prompt for the agent.
func WithSystemPrompt(systemPrompt string) Option {
	return func(a *BaseAgent) error {
		a.systemPrompt = systemPrompt
		return nil
	}
}

// WithName sets the name of the agent.
func WithName(name string) Option {
	return func(a *BaseAgent) error {
		a.name = name
		return nil
	}
}

// WithDescription sets the description of the agent.
func WithDescription(description string) Option {
	return func(a *BaseAgent) error {
		a.description = description
		return nil
	}
}

// Agent is the interface that all agents must implement.
type Agent interface {
	// Run runs the agent on the given task.
	Run(ctx context.Context, task string) (any, error)

	// Step executes a single step of the agent's reasoning.
	Step(ctx context.Context, memory *memory.ActionStep) (any, error)

	// GetTools returns the tools available to the agent.
	GetTools() []tools.Tool

	// GetMemory returns the agent's memory.
	GetMemory() *memory.Memory

	// GetModel returns the agent's model.
	GetModel() models.Model

	// GetName returns the agent's name.
	GetName() string

	// GetDescription returns the agent's description.
	GetDescription() string
}

// BaseAgent provides a base implementation of the Agent interface.
type BaseAgent struct {
	tools        []tools.Tool
	model        models.Model
	memory       *memory.Memory
	maxSteps     int
	systemPrompt string
	name         string
	description  string
	stepper      Stepper
}

// Stepper is an interface for executing agent steps.
type Stepper interface {
	Step(ctx context.Context, step *memory.ActionStep) (any, error)
}

// NewBaseAgent creates a new BaseAgent with the given tools and model.
func NewBaseAgent(tools []tools.Tool, model models.Model, opts ...Option) (*BaseAgent, error) {
	if len(tools) == 0 {
		return nil, errors.New("at least one tool is required")
	}

	if model == nil {
		return nil, errors.New("model is required")
	}

	agent := &BaseAgent{
		tools:        tools,
		model:        model,
		memory:       memory.NewMemory(),
		maxSteps:     20, // Default max steps
		systemPrompt: "You are a helpful assistant that can use tools to help the user.",
		name:         "BaseAgent",
		description:  "A base agent implementation",
	}

	for _, opt := range opts {
		if err := opt(agent); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	return agent, nil
}

// SetStepper sets the stepper for the agent.
func (a *BaseAgent) SetStepper(s Stepper) {
	a.stepper = s
}

// GetTools returns the tools available to the agent.
func (a *BaseAgent) GetTools() []tools.Tool {
	return a.tools
}

// GetMemory returns the agent's memory.
func (a *BaseAgent) GetMemory() *memory.Memory {
	return a.memory
}

// GetModel returns the agent's model.
func (a *BaseAgent) GetModel() models.Model {
	return a.model
}

// GetName returns the agent's name.
func (a *BaseAgent) GetName() string {
	return a.name
}

// GetDescription returns the agent's description.
func (a *BaseAgent) GetDescription() string {
	return a.description
}

// Run runs the agent on the given task.
func (a *BaseAgent) Run(ctx context.Context, task string) (any, error) {
	// Initialize the memory
	a.memory = memory.NewMemory()

	// Add the system prompt to memory
	systemMessages := []models.Message{
		{
			Role:    models.RoleSystem,
			Content: a.systemPrompt,
		},
	}
	a.memory.AddSystemPromptStep(a.systemPrompt, systemMessages)
	a.memory.CompleteCurrentStep()

	// Add the task to memory
	taskMessages := []models.Message{
		{
			Role:    models.RoleUser,
			Content: task,
		},
	}
	a.memory.AddTaskStep(task, taskMessages)
	a.memory.CompleteCurrentStep()

	// Execute steps until completion or max steps reached
	var finalAnswer any
	var lastError error

	for step := 0; step < a.maxSteps; step++ {
		// Create action step
		messages := a.buildMessages()
		actionStep := a.memory.AddActionStep(task, messages)

		// Execute step
		var result any
		var err error
		if a.stepper != nil {
			result, err = a.stepper.Step(ctx, actionStep)
		} else {
			result, err = a.Step(ctx, actionStep)
		}
		if err != nil {
			a.memory.CompleteCurrentStep()
			lastError = err
			break
		}

		// Check if we have a final answer
		if result != nil {
			finalAnswer = result
			a.memory.CompleteCurrentStep()
			break
		}

		a.memory.CompleteCurrentStep()
	}

	if finalAnswer == nil && lastError == nil {
		lastError = fmt.Errorf("agent reached maximum number of steps (%d) without finding an answer", a.maxSteps)
	}

	return finalAnswer, lastError
}

// buildMessages constructs the message history for the model.
func (a *BaseAgent) buildMessages() []models.Message {
	var messages []models.Message

	// Add system prompt
	messages = append(messages, models.Message{
		Role:    models.RoleSystem,
		Content: a.systemPrompt,
	})

	// Add tool definitions to system prompt
	if len(a.tools) > 0 {
		toolsDesc := a.buildToolsDescription()
		messages = append(messages, models.Message{
			Role:    models.RoleSystem,
			Content: toolsDesc,
		})
	}

	// Add messages from memory
	memMessages := a.memory.GetMessages()
	for _, msg := range memMessages {
		// Skip system messages as we've already added them
		if msg.Role == models.RoleSystem {
			continue
		}
		messages = append(messages, msg)
	}

	return messages
}

// buildToolsDescription constructs a description of all available tools.
func (a *BaseAgent) buildToolsDescription() string {
	var builder strings.Builder

	builder.WriteString("You have access to the following tools:\n\n")

	for _, tool := range a.tools {
		builder.WriteString(tools.FormatToolDescription(tool))
		builder.WriteString("\n")
	}

	builder.WriteString("To use a tool, respond with a message formatted as follows:\n")
	builder.WriteString("```json\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"tool\": \"tool_name\",\n")
	builder.WriteString("  \"args\": {\n")
	builder.WriteString("    \"arg1\": \"value1\",\n")
	builder.WriteString("    \"arg2\": \"value2\"\n")
	builder.WriteString("  }\n")
	builder.WriteString("}\n")
	builder.WriteString("```\n")
	builder.WriteString("If you want to provide a final answer, just respond with text instead.\n")

	return builder.String()
}

// Step executes a single step of the agent's reasoning.
// This is a placeholder implementation that should be overridden by derived agents.
func (a *BaseAgent) Step(ctx context.Context, step *memory.ActionStep) (any, error) {
	return nil, errors.New("Step method must be implemented by derived agents")
}

// extractToolCall extracts a tool call from the model's response.
func (a *BaseAgent) extractToolCall(response string) (string, map[string]any, error) {
	// Extract JSON from the response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return "", nil, nil // No tool call, just a regular message
	}

	var call struct {
		Tool string         `json:"tool"`
		Args map[string]any `json:"args"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &call); err != nil {
		return "", nil, fmt.Errorf("failed to parse tool call: %w", err)
	}

	if call.Tool == "" {
		return "", nil, nil // No tool call
	}

	return call.Tool, call.Args, nil
}

// findTool finds a tool by name.
func (a *BaseAgent) findTool(name string) (tools.Tool, error) {
	for _, tool := range a.tools {
		if tool.Name() == name {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("tool not found: %s", name)
}

// executeToolCall executes a tool call.
func (a *BaseAgent) executeToolCall(
	ctx context.Context,
	step *memory.ActionStep,
	toolName string,
	args map[string]any,
) (any, error) {
	// Find the tool
	tool, err := a.findTool(toolName)
	if err != nil {
		return nil, err
	}

	// Execute the tool
	result, err := tool.Execute(ctx, args)

	// Record the tool call in memory
	a.memory.AddToolCall(toolName, args, result, err)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// extractJSON extracts JSON from a string.
func extractJSON(s string) string {
	// Look for JSON between triple backticks
	start := strings.Index(s, "```json")
	if start == -1 {
		start = strings.Index(s, "```")
	}
	if start == -1 {
		return "" // No JSON found
	}

	// Find the end of the first line
	start = strings.Index(s[start:], "\n")
	if start == -1 {
		return "" // Invalid format
	}
	start += 1 // Skip the newline

	// Find the end marker
	end := strings.Index(s[start:], "```")
	if end == -1 {
		return "" // Invalid format
	}

	// Extract the JSON
	return strings.TrimSpace(s[start : start+end])
}
