package tests

import (
	"context"
	"testing"

	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// TestCreateSimpleTool tests creating a simple tool
func TestCreateSimpleTool(t *testing.T) {
	// Create a simple function
	hello := func(name string) string {
		return "Hello, " + name
	}

	// Create a tool with type parameter
	tool := tools.CreateTool[func(string) string](
		"hello",
		"Greets a person by name",
	)(hello)

	// Check that the tool has the expected properties
	if tool.Name() != "hello" {
		t.Errorf("Expected tool name to be 'hello', got '%s'", tool.Name())
	}

	if tool.Description() != "Greets a person by name" {
		t.Errorf("Expected tool description to be 'Greets a person by name', got '%s'", tool.Description())
	}

	// Test executing the tool
	result, err := tool.Execute(context.Background(), map[string]any{
		"arg0": "World",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("Expected result to be 'Hello, World', got '%v'", result)
	}
}
