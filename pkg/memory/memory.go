// Package memory provides structures for storing agent interactions.
package memory

import (
	"fmt"
	"time"

	"github.com/epuerta9/smolagents-go/pkg/models"
)

// ToolCall represents a call to a tool.
type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Output    any            `json:"output"`
	Error     string         `json:"error,omitempty"`
}

// Step represents a single step in the agent's execution.
type Step struct {
	Type           string           `json:"type"`
	Messages       []models.Message `json:"messages"`
	StartTimestamp time.Time        `json:"start_timestamp"`
	EndTimestamp   time.Time        `json:"end_timestamp"`
	ToolCalls      []ToolCall       `json:"tool_calls,omitempty"`
}

// TaskStep represents the initial task given to the agent.
type TaskStep struct {
	Step
	Task string `json:"task"`
}

// SystemPromptStep represents the system prompt.
type SystemPromptStep struct {
	Step
	SystemPrompt string `json:"system_prompt"`
}

// ActionStep represents an agent's action.
type ActionStep struct {
	Step
	Input  string `json:"input"`
	Output any    `json:"output"`
}

// PlanningStep represents a planning step.
type PlanningStep struct {
	Step
	Facts string `json:"facts"`
	Plan  string `json:"plan"`
}

// Memory stores the agent's execution history.
type Memory struct {
	Steps   []Step `json:"steps"`
	curStep *Step
}

// NewMemory creates a new memory.
func NewMemory() *Memory {
	return &Memory{
		Steps: []Step{},
	}
}

// AddTaskStep adds a task step to the memory.
func (m *Memory) AddTaskStep(task string, messages []models.Message) *TaskStep {
	taskStep := &TaskStep{
		Step: Step{
			Type:           "task",
			Messages:       messages,
			StartTimestamp: time.Now(),
		},
		Task: task,
	}

	m.curStep = &taskStep.Step
	m.Steps = append(m.Steps, taskStep.Step)
	return taskStep
}

// AddSystemPromptStep adds a system prompt step to the memory.
func (m *Memory) AddSystemPromptStep(systemPrompt string, messages []models.Message) *SystemPromptStep {
	systemStep := &SystemPromptStep{
		Step: Step{
			Type:           "system_prompt",
			Messages:       messages,
			StartTimestamp: time.Now(),
		},
		SystemPrompt: systemPrompt,
	}

	m.curStep = &systemStep.Step
	m.Steps = append(m.Steps, systemStep.Step)
	return systemStep
}

// AddActionStep adds an action step to the memory.
func (m *Memory) AddActionStep(input string, messages []models.Message) *ActionStep {
	actionStep := &ActionStep{
		Step: Step{
			Type:           "action",
			Messages:       messages,
			StartTimestamp: time.Now(),
		},
		Input: input,
	}

	m.curStep = &actionStep.Step
	m.Steps = append(m.Steps, actionStep.Step)
	return actionStep
}

// AddPlanningStep adds a planning step to the memory.
func (m *Memory) AddPlanningStep(facts string, plan string, messages []models.Message) *PlanningStep {
	planningStep := &PlanningStep{
		Step: Step{
			Type:           "planning",
			Messages:       messages,
			StartTimestamp: time.Now(),
		},
		Facts: facts,
		Plan:  plan,
	}

	m.curStep = &planningStep.Step
	m.Steps = append(m.Steps, planningStep.Step)
	return planningStep
}

// AddToolCall adds a tool call to the current step.
func (m *Memory) AddToolCall(name string, args map[string]any, output any, err error) *ToolCall {
	if m.curStep == nil {
		return nil
	}

	toolCall := ToolCall{
		Name:      name,
		Arguments: args,
		Output:    output,
	}

	if err != nil {
		toolCall.Error = err.Error()
	}

	m.curStep.ToolCalls = append(m.curStep.ToolCalls, toolCall)
	return &toolCall
}

// CompleteCurrentStep completes the current step.
func (m *Memory) CompleteCurrentStep() {
	if m.curStep == nil {
		return
	}

	m.curStep.EndTimestamp = time.Now()
	m.curStep = nil
}

// GetSteps returns all steps in the memory.
func (m *Memory) GetSteps() []Step {
	return m.Steps
}

// GetToolCalls returns all tool calls from all steps.
func (m *Memory) GetToolCalls() []ToolCall {
	var toolCalls []ToolCall

	for _, step := range m.Steps {
		toolCalls = append(toolCalls, step.ToolCalls...)
	}

	return toolCalls
}

// GetMessages returns all messages from all steps.
func (m *Memory) GetMessages() []models.Message {
	var messages []models.Message

	for _, step := range m.Steps {
		messages = append(messages, step.Messages...)
	}

	return messages
}

// String returns a string representation of the memory.
func (m *Memory) String() string {
	var s string

	for i, step := range m.Steps {
		s += fmt.Sprintf("Step %d: %s\n", i+1, step.Type)

		for j, msg := range step.Messages {
			s += fmt.Sprintf("  Message %d: [%s] %s\n", j+1, msg.Role, msg.Content)
		}

		for j, toolCall := range step.ToolCalls {
			s += fmt.Sprintf("  Tool Call %d: %s\n", j+1, toolCall.Name)
			s += fmt.Sprintf("    Arguments: %v\n", toolCall.Arguments)

			if toolCall.Error != "" {
				s += fmt.Sprintf("    Error: %s\n", toolCall.Error)
			} else {
				s += fmt.Sprintf("    Output: %v\n", toolCall.Output)
			}
		}

		s += "\n"
	}

	return s
}
