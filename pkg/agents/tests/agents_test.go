package tests

import (
	"context"
	"testing"

	"github.com/epuerta9/smolagents-go/pkg/agents"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// MockModel implements the models.Model interface for testing
type MockModel struct {
	generateResponse string
	generateError    error
}

func (m *MockModel) Generate(ctx context.Context, messages []models.Message) (string, error) {
	if m.generateError != nil {
		return "", m.generateError
	}
	return m.generateResponse, nil
}

func (m *MockModel) GenerateWithTools(ctx context.Context, messages []models.Message, tools []map[string]any) (string, error) {
	return m.Generate(ctx, messages)
}

// MockTool implements the tools.Tool interface for testing
type MockTool struct {
	name        string
	description string
	output      any
	err         error
}

func (t *MockTool) Name() string        { return t.name }
func (t *MockTool) Description() string { return t.description }
func (t *MockTool) Schema() *tools.ToolSchema {
	return &tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.PropertyDef{
			"arg1": {
				Type:        "string",
				Description: "Test argument",
			},
		},
	}
}
func (t *MockTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.output, nil
}

// TestBaseAgentCreation tests creating a new BaseAgent
func TestBaseAgentCreation(t *testing.T) {
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		output:      "test output",
	}
	mockModel := &MockModel{
		generateResponse: "test response",
	}

	tests := []struct {
		name    string
		tools   []tools.Tool
		model   models.Model
		opts    []agents.Option
		wantErr bool
	}{
		{
			name:    "valid creation",
			tools:   []tools.Tool{mockTool},
			model:   mockModel,
			wantErr: false,
		},
		{
			name:    "no tools",
			tools:   []tools.Tool{},
			model:   mockModel,
			wantErr: true,
		},
		{
			name:    "nil model",
			tools:   []tools.Tool{mockTool},
			model:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := agents.NewBaseAgent(tt.tools, tt.model, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBaseAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && agent == nil {
				t.Error("NewBaseAgent() returned nil agent without error")
			}
		})
	}
}

// TestCodeAgentExecution tests the CodeAgent's execution
func TestCodeAgentExecution(t *testing.T) {
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		output:      "tool output",
	}

	tests := []struct {
		name           string
		modelResponse  string
		task           string
		expectedResult any
		expectToolCall bool
		wantErr        bool
	}{
		{
			name:           "direct response",
			modelResponse:  "This is a final answer",
			task:           "test task",
			expectedResult: "This is a final answer",
			expectToolCall: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockModel := &MockModel{
				generateResponse: tt.modelResponse,
			}

			agent, err := agents.NewCodeAgent([]tools.Tool{mockTool}, mockModel)
			if err != nil {
				t.Fatalf("Failed to create CodeAgent: %v", err)
			}

			// Create a memory step for testing
			messages := []models.Message{
				{Role: models.RoleUser, Content: tt.task},
			}
			step := agent.GetMemory().AddActionStep(tt.task, messages)

			// Test the Step method directly
			result, err := agent.Step(context.Background(), step)
			if (err != nil) != tt.wantErr {
				t.Errorf("CodeAgent.Step() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expectedResult {
				t.Errorf("CodeAgent.Step() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestToolCallingAgentExecution tests the ToolCallingAgent's execution
func TestToolCallingAgentExecution(t *testing.T) {
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		output:      "tool output",
	}

	tests := []struct {
		name           string
		modelResponse  string
		task           string
		expectedResult any
		expectToolCall bool
		wantErr        bool
	}{
		{
			name:           "direct response",
			modelResponse:  "This is a final answer",
			task:           "test task",
			expectedResult: "This is a final answer",
			expectToolCall: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockModel := &MockModel{
				generateResponse: tt.modelResponse,
			}

			agent, err := agents.NewToolCallingAgent([]tools.Tool{mockTool}, mockModel)
			if err != nil {
				t.Fatalf("Failed to create ToolCallingAgent: %v", err)
			}

			// Create a memory step for testing
			messages := []models.Message{
				{Role: models.RoleUser, Content: tt.task},
			}
			step := agent.GetMemory().AddActionStep(tt.task, messages)

			// Test the Step method directly
			result, err := agent.Step(context.Background(), step)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToolCallingAgent.Step() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expectedResult {
				t.Errorf("ToolCallingAgent.Step() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestAgentOptions tests the agent options
func TestAgentOptions(t *testing.T) {
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
	}
	mockModel := &MockModel{}

	tests := []struct {
		name          string
		opts          []agents.Option
		expectedName  string
		expectedSteps int
	}{
		{
			name: "custom name and max steps",
			opts: []agents.Option{
				agents.WithName("CustomAgent"),
				agents.WithMaxSteps(10),
			},
			expectedName:  "CustomAgent",
			expectedSteps: 10,
		},
		{
			name: "custom system prompt",
			opts: []agents.Option{
				agents.WithSystemPrompt("Custom prompt"),
			},
			expectedName: "BaseAgent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := agents.NewBaseAgent([]tools.Tool{mockTool}, mockModel, tt.opts...)
			if err != nil {
				t.Fatalf("Failed to create agent: %v", err)
			}

			if agent.GetName() != tt.expectedName {
				t.Errorf("Agent name = %v, want %v", agent.GetName(), tt.expectedName)
			}
		})
	}
}
