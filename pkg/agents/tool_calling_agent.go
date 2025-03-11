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

// ToolCallingAgent is an agent specialized in calling tools and handling their output.
type ToolCallingAgent struct {
	tools        []tools.Tool
	model        models.Model
	memory       *memory.Memory
	maxSteps     int
	systemPrompt string
	name         string
	description  string
}

// NewToolCallingAgent creates a new ToolCallingAgent with the given tools and model.
func NewToolCallingAgent(tools []tools.Tool, model models.Model, opts ...Option) (*ToolCallingAgent, error) {
	if len(tools) == 0 {
		return nil, errors.New("at least one tool is required")
	}

	if model == nil {
		return nil, errors.New("model is required")
	}

	agent := &ToolCallingAgent{
		tools:        tools,
		model:        model,
		memory:       memory.NewMemory(),
		maxSteps:     20, // Default max steps
		systemPrompt: "You are a helpful assistant that can use tools to help the user.",
		name:         "ToolCallingAgent",
		description:  "An agent specialized in calling tools and handling their output",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(&BaseAgent{
			tools:        agent.tools,
			model:        agent.model,
			memory:       agent.memory,
			maxSteps:     agent.maxSteps,
			systemPrompt: agent.systemPrompt,
			name:         agent.name,
			description:  agent.description,
		}); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	return agent, nil
}

// Run runs the agent on the given task.
func (a *ToolCallingAgent) Run(ctx context.Context, task string) (any, error) {
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
		result, err := a.Step(ctx, actionStep)
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

// Step executes a single step of the agent's reasoning.
func (a *ToolCallingAgent) Step(ctx context.Context, step *memory.ActionStep) (any, error) {
	// Generate model response
	response, err := a.model.GenerateWithTools(
		ctx,
		step.Messages,
		a.buildToolsSchema(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Add assistant response to memory
	step.Messages = append(step.Messages, models.Message{
		Role:    models.RoleAssistant,
		Content: response,
	})

	// Check if the response is a tool call
	toolName, args, err := a.extractToolCall(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tool call: %w", err)
	}

	// If no tool call, treat as final answer
	if toolName == "" {
		return response, nil
	}

	// Execute the tool call
	result, err := a.executeToolCall(ctx, step, toolName, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool call: %w", err)
	}

	// Add tool result to memory
	resultStr := fmt.Sprintf("%v", result)
	step.Messages = append(step.Messages, models.Message{
		Role:    models.RoleTool,
		Name:    toolName,
		Content: resultStr,
	})

	// No final answer yet, continue to next step
	return nil, nil
}

// GetTools returns the tools available to the agent.
func (a *ToolCallingAgent) GetTools() []tools.Tool {
	return a.tools
}

// GetMemory returns the agent's memory.
func (a *ToolCallingAgent) GetMemory() *memory.Memory {
	return a.memory
}

// GetModel returns the agent's model.
func (a *ToolCallingAgent) GetModel() models.Model {
	return a.model
}

// GetName returns the agent's name.
func (a *ToolCallingAgent) GetName() string {
	return a.name
}

// GetDescription returns the agent's description.
func (a *ToolCallingAgent) GetDescription() string {
	return a.description
}

// buildMessages constructs the message history for the model.
func (a *ToolCallingAgent) buildMessages() []models.Message {
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
func (a *ToolCallingAgent) buildToolsDescription() string {
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

// buildToolsSchema builds the JSON schema for the tools.
func (a *ToolCallingAgent) buildToolsSchema() []map[string]any {
	schemas := make([]map[string]any, 0, len(a.tools))

	for _, tool := range a.tools {
		schema := tool.Schema()

		toolSchema := map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters":  schema,
			},
		}

		schemas = append(schemas, toolSchema)
	}

	return schemas
}

// extractToolCall extracts a tool call from the model's response.
func (a *ToolCallingAgent) extractToolCall(response string) (string, map[string]any, error) {
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
func (a *ToolCallingAgent) findTool(name string) (tools.Tool, error) {
	for _, tool := range a.tools {
		if tool.Name() == name {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("tool not found: %s", name)
}

// executeToolCall executes a tool call.
func (a *ToolCallingAgent) executeToolCall(
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
