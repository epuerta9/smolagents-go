package tools

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestCreateTool tests the CreateTool function with generic type parameters
func TestCreateTool(t *testing.T) {
	// Test with a simple string function
	greet := func(name string) string {
		return "Hello, " + name
	}

	// Create a tool with explicit type parameter
	greetTool := CreateTool[func(string) string](
		"greet",
		"Greets a person by name",
	)(greet)

	// Verify the tool properties
	if greetTool.Name() != "greet" {
		t.Errorf("Expected tool name to be 'greet', got '%s'", greetTool.Name())
	}

	if greetTool.Description() != "Greets a person by name" {
		t.Errorf("Expected tool description to be 'Greets a person by name', got '%s'", greetTool.Description())
	}

	// Verify schema
	schema := greetTool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type to be 'object', got '%s'", schema.Type)
	}

	if len(schema.Required) != 1 {
		t.Errorf("Expected schema to have 1 required parameter, got %d", len(schema.Required))
	}

	if len(schema.Properties) != 1 {
		t.Errorf("Expected schema to have 1 property, got %d", len(schema.Properties))
	}

	if _, ok := schema.Properties["arg0"]; !ok {
		t.Error("Expected schema to have property 'arg0'")
	}

	// Test with multiple parameters
	add := func(a, b int) int {
		return a + b
	}

	addTool := CreateTool[func(int, int) int](
		"add",
		"Adds two numbers",
	)(add)

	// Verify multiple parameters
	schema = addTool.Schema()
	if len(schema.Required) != 2 {
		t.Errorf("Expected schema to have 2 required parameters, got %d", len(schema.Required))
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected schema to have 2 properties, got %d", len(schema.Properties))
	}
}

// TestToolExecution tests executing a tool with arguments
func TestToolExecution(t *testing.T) {
	// Create a tool that adds two numbers
	add := func(a, b int) int {
		return a + b
	}

	addTool := CreateTool[func(int, int) int](
		"add",
		"Adds two numbers",
	)(add)

	// Create arguments map
	args := map[string]any{
		"arg0": 5,
		"arg1": 7,
	}

	// Execute the tool
	result, err := addTool.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify result
	if result != 12 {
		t.Errorf("Expected result to be 12, got %v", result)
	}

	// Test with wrong argument type
	args = map[string]any{
		"arg0": "not a number",
		"arg1": 7,
	}

	_, err = addTool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error due to wrong argument type, got nil")
	}

	// Test missing arguments
	args = map[string]any{
		"arg0": 5,
		// Missing arg1
	}

	_, err = addTool.Execute(context.Background(), args)
	if err == nil {
		t.Error("Expected error due to missing argument, got nil")
	}
}

// TestComplexToolExecution tests tools with more complex argument types
func TestComplexToolExecution(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	// Create a tool that processes a more complex struct
	processPerson := func(person Person) string {
		return fmt.Sprintf("%s is %d years old", person.Name, person.Age)
	}

	personTool := CreateTool[func(Person) string](
		"process_person",
		"Processes a person",
	)(processPerson)

	// Try executing with an argument that can be converted via JSON
	args := map[string]any{
		"arg0": map[string]any{
			"Name": "Alice",
			"Age":  30,
		},
	}

	_, err := personTool.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Expected no error with convertible struct, got %v", err)
	}
}

// TestTypeInference tests type inference in the context of tools
func TestTypeInference(t *testing.T) {
	goType := reflect.TypeOf("")
	jsonType, err := goTypeToJSONType(goType)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if jsonType != "string" {
		t.Errorf("Expected JSON type 'string', got '%s'", jsonType)
	}

	// Test with integer
	goType = reflect.TypeOf(int(0))
	jsonType, err = goTypeToJSONType(goType)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if jsonType != "integer" {
		t.Errorf("Expected JSON type 'integer', got '%s'", jsonType)
	}

	// Test with slice
	goType = reflect.TypeOf([]string{})
	jsonType, err = goTypeToJSONType(goType)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if jsonType != "array" {
		t.Errorf("Expected JSON type 'array', got '%s'", jsonType)
	}
}

// TestFormatToolDescription tests the tool description formatting
func TestFormatToolDescription(t *testing.T) {
	// Create a simple tool
	greet := func(name string) string {
		return "Hello, " + name
	}

	greetTool := CreateTool[func(string) string](
		"greet",
		"Greets a person by name",
	)(greet)

	// Format the description
	description := FormatToolDescription(greetTool)

	// Check that the description contains expected strings
	if !strings.Contains(description, "Tool Name: greet") {
		t.Error("Expected description to contain tool name")
	}

	if !strings.Contains(description, "Description: Greets a person by name") {
		t.Error("Expected description to contain tool description")
	}

	if !strings.Contains(description, "Parameters:") {
		t.Error("Expected description to list parameters")
	}
}
