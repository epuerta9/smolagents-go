package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/epuerta9/smolagents-go/pkg/memory"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// CodeAgent is an agent specialized in generating and executing code.
type CodeAgent struct {
	*BaseAgent
}

// NewCodeAgent creates a new CodeAgent with the given tools and model.
func NewCodeAgent(tools []tools.Tool, model models.Model, opts ...Option) (*CodeAgent, error) {
	baseAgent, err := NewBaseAgent(tools, model, opts...)
	if err != nil {
		return nil, err
	}

	agent := &CodeAgent{
		BaseAgent: baseAgent,
	}

	// Set default agent properties if not overridden by options
	if agent.name == "BaseAgent" {
		agent.name = "CodeAgent"
	}

	if agent.description == "A base agent implementation" {
		agent.description = "An agent specialized in generating and executing code"
	}

	if agent.systemPrompt == "You are a helpful assistant that can use tools to help the user." {
		agent.systemPrompt = `You are a helpful assistant that can use tools to help the user.
You will be given a user request, and your job is to create a final answer for the user.
You have access to tools that can help you fulfill the user's request.
If necessary, you can call these tools and use their output to craft a better answer.
You can also generate and execute code to solve complex problems.
When you have the answer to the user's request, respond with the relevant information.`
	}

	return agent, nil
}

func (a *CodeAgent) executeAndAddResToMem(ctx context.Context, step *memory.ActionStep, toolName string,
	args map[string]any) (any, error) {
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

// Step executes a single step of the agent's reasoning.
func (a *CodeAgent) Step(ctx context.Context, step *memory.ActionStep) (any, error) {
	// Generate model response
	response, err := a.model.Generate(ctx, step.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Add assistant response to memory
	step.Messages = append(step.Messages, models.Message{
		Role:    models.RoleAssistant,
		Content: response,
	})

	// Check if the response contains code blocks
	codeBlocks := extractCodeBlocks(response)

	// For simplicity, we'll just execute the first code block that contains a tool call
	for _, codeBlock := range codeBlocks {
		// Check if the code block contains a tool call
		toolName, args, err := a.extractToolCallFromCode(codeBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to extract tool call from code: %w", err)
		}

		if toolName != "" {
			return a.executeAndAddResToMem(ctx, step, toolName, args)
		}
	}

	// Check if the response is a direct tool call (JSON format)
	toolName, args, err := a.extractToolCall(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tool call: %w", err)
	}

	// If no tool call, treat as final answer
	if toolName == "" {
		return response, nil
	}

	return a.executeAndAddResToMem(ctx, step, toolName, args)
}

// extractCodeBlocks extracts code blocks from a string.
func extractCodeBlocks(s string) []string {
	var blocks []string

	// Match code blocks between triple backticks
	re := regexp.MustCompile("```(?:\\w+)?\\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(s, -1)

	for _, match := range matches {
		if len(match) > 1 {
			blocks = append(blocks, match[1])
		}
	}

	return blocks
}

// extractToolCallFromCode extracts a tool call from a code block.
func (a *CodeAgent) extractToolCallFromCode(code string) (string, map[string]any, error) {
	// Look for patterns like: result = tool_name(arg1="value1", arg2="value2")
	re := regexp.MustCompile(`(\w+)\s*\((.*?)\)`)
	match := re.FindStringSubmatch(code)

	if len(match) < 3 {
		return "", nil, nil
	}

	toolName := match[1]
	argsStr := match[2]

	// Check if the tool exists
	var found bool
	for _, tool := range a.tools {
		if tool.Name() == toolName {
			found = true
			break
		}
	}

	if !found {
		return "", nil, nil
	}

	// Parse arguments
	args := make(map[string]any)

	// Split by commas, but handle quoted strings properly
	re = regexp.MustCompile(`(\w+)\s*=\s*(?:"([^"]*)"|'([^']*)'|(\d+(?:\.\d+)?))`)
	argMatches := re.FindAllStringSubmatch(argsStr, -1)

	for _, argMatch := range argMatches {
		if len(argMatch) < 5 {
			continue
		}

		argName := argMatch[1]
		var argValue any

		if argMatch[2] != "" {
			// Double-quoted string
			argValue = argMatch[2]
		} else if argMatch[3] != "" {
			// Single-quoted string
			argValue = argMatch[3]
		} else if argMatch[4] != "" {
			// Number
			if strings.Contains(argMatch[4], ".") {
				// Float
				var f float64
				fmt.Sscanf(argMatch[4], "%f", &f)
				argValue = f
			} else {
				// Integer
				var i int
				fmt.Sscanf(argMatch[4], "%d", &i)
				argValue = i
			}
		}

		args[argName] = argValue
	}

	return toolName, args, nil
}
