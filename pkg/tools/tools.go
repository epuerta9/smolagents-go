// Package tools provides tools that can be used by AI agents.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Tool represents a function that can be called by an agent.
type Tool interface {
	// Name returns the name of the tool.
	Name() string

	// Description returns a description of what the tool does.
	Description() string

	// Schema returns the JSON schema of the tool.
	Schema() *ToolSchema

	// Execute executes the tool with the given arguments.
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// ToolSchema represents the JSON schema for a tool.
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required"`
}

// PropertyDef defines a property in a tool schema.
type PropertyDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// FunctionTool is a tool that wraps a Go function.
type FunctionTool[F any] struct {
	name        string
	description string
	fn          F
	schema      *ToolSchema
}

// NewFunctionTool creates a new tool from a function.
func NewFunctionTool[F any](name, description string, fn F) (*FunctionTool[F], error) {
	if name == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	if description == "" {
		return nil, fmt.Errorf("tool description cannot be empty")
	}

	// Validate function
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("fn must be a function, got %s", fnType.Kind())
	}

	// Create tool schema from function signature
	schema, err := createSchemaFromFunction(fnType)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &FunctionTool[F]{
		name:        name,
		description: description,
		fn:          fn,
		schema:      schema,
	}, nil
}

// Name returns the name of the tool.
func (t *FunctionTool[F]) Name() string {
	return t.name
}

// Description returns a description of what the tool does.
func (t *FunctionTool[F]) Description() string {
	return t.description
}

// Schema returns the JSON schema of the tool.
func (t *FunctionTool[F]) Schema() *ToolSchema {
	return t.schema
}

// Execute executes the tool with the given arguments.
func (t *FunctionTool[F]) Execute(ctx context.Context, args map[string]any) (any, error) {
	fnType := reflect.TypeOf(t.fn)
	fnValue := reflect.ValueOf(t.fn)

	// Prepare arguments
	callArgs, err := prepareArguments(fnType, args)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments: %w", err)
	}

	// Call function
	results := fnValue.Call(callArgs)

	// Handle results
	if len(results) == 0 {
		return nil, nil
	}

	// Check for error return
	lastResultIdx := len(results) - 1
	if fnType.NumOut() > 1 && fnType.Out(lastResultIdx).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !results[lastResultIdx].IsNil() {
			return nil, results[lastResultIdx].Interface().(error)
		}

		// Return the first result if there's no error
		return results[0].Interface(), nil
	}

	// Return the first result
	return results[0].Interface(), nil
}

// Helper functions to work with the tool function

func createSchemaFromFunction(fnType reflect.Type) (*ToolSchema, error) {
	properties := make(map[string]PropertyDef)
	required := []string{}

	// Process input parameters
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		paramName := fmt.Sprintf("arg%d", i)

		// Map Go types to JSON schema types
		jsonType, err := goTypeToJSONType(paramType)
		if err != nil {
			return nil, err
		}

		properties[paramName] = PropertyDef{
			Type:        jsonType,
			Description: fmt.Sprintf("Parameter %d of type %s", i, paramType.String()),
		}

		required = append(required, paramName)
	}

	return &ToolSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}, nil
}

func goTypeToJSONType(t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.String:
		return "string", nil
	case reflect.Bool:
		return "boolean", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer", nil
	case reflect.Float32, reflect.Float64:
		return "number", nil
	case reflect.Slice, reflect.Array:
		return "array", nil
	case reflect.Map, reflect.Struct:
		return "object", nil
	default:
		return "", fmt.Errorf("unsupported type: %s", t.String())
	}
}

func prepareArguments(fnType reflect.Type, args map[string]any) ([]reflect.Value, error) {
	callArgs := make([]reflect.Value, fnType.NumIn())

	// For each parameter of the function
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		paramName := fmt.Sprintf("arg%d", i)

		// Find the corresponding argument
		arg, ok := args[paramName]
		if !ok {
			return nil, fmt.Errorf("missing required argument: %s", paramName)
		}

		// Convert argument to the correct type
		value, err := convertArgument(arg, paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert argument %s: %w", paramName, err)
		}

		callArgs[i] = value
	}

	return callArgs, nil
}

func convertArgument(arg any, targetType reflect.Type) (reflect.Value, error) {
	// Handle nil
	if arg == nil {
		return reflect.Zero(targetType), nil
	}

	// Try to directly convert
	argValue := reflect.ValueOf(arg)
	if argValue.Type().ConvertibleTo(targetType) {
		return argValue.Convert(targetType), nil
	}

	// Try using JSON marshaling/unmarshaling for more complex conversions
	jsonData, err := json.Marshal(arg)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to marshal argument: %w", err)
	}

	// Create a new instance of the target type
	newValue := reflect.New(targetType).Interface()

	if err := json.Unmarshal(jsonData, newValue); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to unmarshal argument: %w", err)
	}

	return reflect.ValueOf(newValue).Elem(), nil
}

// DecorateFunction adds metadata to a function and returns a FunctionTool.
func DecorateFunction[F any](fn F, name, description string) (*FunctionTool[F], error) {
	return NewFunctionTool(name, description, fn)
}

// CreateTool is a decorator-style function that creates a new FunctionTool.
// Usage:
//
//	var GetWeather = tools.CreateTool("get_weather", "Get the current weather at the given location.")(func(location string) string {
//		// implementation
//	})
func CreateTool[F any](name, description string) func(F) *FunctionTool[F] {
	return func(fn F) *FunctionTool[F] {
		tool, err := NewFunctionTool(name, description, fn)
		if err != nil {
			panic(fmt.Sprintf("failed to create tool: %v", err))
		}
		return tool
	}
}

// FormatToolDescription formats a tool description for the model prompt.
func FormatToolDescription(tool Tool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Tool Name: %s\n", tool.Name()))
	sb.WriteString(fmt.Sprintf("Description: %s\n", tool.Description()))

	schema := tool.Schema()
	if len(schema.Properties) > 0 {
		sb.WriteString("Parameters:\n")

		for name, prop := range schema.Properties {
			required := ""
			for _, req := range schema.Required {
				if req == name {
					required = " (required)"
					break
				}
			}

			sb.WriteString(fmt.Sprintf("  - %s: %s%s\n    %s\n",
				name, prop.Type, required, prop.Description))
		}
	}

	return sb.String()
}
