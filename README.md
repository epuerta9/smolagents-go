# SmolagentsGo

SmolagentsGo is a Go implementation of the [smolagents](https://github.com/huggingface/smolagents) library, which provides a framework for creating AI agents with tool calling capabilities.

## Features

- **Type Safety with Generics** - Fully leverages Go's generics for type-safe agent and tool development
- **Easy Tool Creation** - Simple API for creating tools from functions
- **Flexible Agents** - Support for different agent types (ToolCallingAgent, CodeAgent)
- **Hugging Face Integration** - Built-in support for Hugging Face models

## Installation

```bash
go get github.com/huggingface/smolagents-go
```

## Usage

### Creating a Tool

Tools are functions that can be called by an agent. You can create a tool using the `tools.CreateTool` function with generics for type safety:

```go
import "github.com/huggingface/smolagents-go/pkg/tools"

// Define a function
func GetWeather(location string, celsius bool) string {
    if celsius {
        return fmt.Sprintf("The current weather in %s is sunny with a temperature of 25 °C.", location)
    }
    return fmt.Sprintf("The current weather in %s is sunny with a temperature of 77 °F.", location)
}

// Create a tool with type-safe generics
getWeather := tools.CreateTool[func(string, bool) string](
    "get_weather",
    "Get the current weather at the given location.",
)(GetWeather)
```

### Creating a Model

Models are used to generate responses. You can create a model using the `models.NewHfApiModel` function:

```go
import "github.com/huggingface/smolagents-go/pkg/models"

// Create a model
model := models.NewHfApiModel(
    "mistralai/Mistral-7B-Instruct-v0.2",
    models.WithApiKey("your-api-key"),
    models.WithMaxTokens(1024),
)
```

### Creating an Agent

Agents use models and tools to solve tasks. You can create an agent using the `agents.NewToolCallingAgent` or `agents.NewCodeAgent` functions:

```go
import "github.com/huggingface/smolagents-go/pkg/agents"

// Create a ToolCallingAgent
agent, err := agents.NewToolCallingAgent(
    []tools.Tool{getWeather, convertCurrency, getJoke},
    model,
    agents.WithMaxSteps(10),
    agents.WithSystemPrompt("You are a helpful assistant that can use tools to help the user."),
)
if err != nil {
    log.Fatalf("Failed to create agent: %v", err)
}

// Run the agent
ctx := context.Background()
task := "What's the weather like in Paris?"
result, err := agent.Run(ctx, task)
if err != nil {
    log.Fatalf("Agent execution failed: %v", err)
}
fmt.Printf("Result: %s\n", result)
```

## Examples

### Multiple Tools

The [multiple_tools.go](examples/multiple_tools.go) example demonstrates how to create and use multiple tools with an agent.

```bash
cd examples
go run multiple_tools.go
```

### Multiple Tools with Generics

The [multiple_tools_generic.go](examples/multiple_tools_generic.go) example shows how to use generics for type-safe tool creation.

```bash
cd examples
go run multiple_tools_generic.go
```

### RAG (Retrieval-Augmented Generation)

The [rag_example](examples/rag_example) example demonstrates how to use the library for RAG.

```bash
cd examples/rag_example
go run main.go
```

## Building the Examples

You can build all the examples with:

```bash
make build
```

This will create executable binaries in the `bin` directory.

## Running Tests

You can run the tests with:

```bash
make test
```

## Implementation Notes

### Generics for Type Safety

This implementation uses Go's generics (introduced in Go 1.18) to provide type safety for tools. When creating a tool, you must specify the function signature as a type parameter:

```go
// Without generics (would cause compile-time errors):
myTool := tools.CreateTool("name", "description")(myFunction)

// With generics (type-safe):
myTool := tools.CreateTool[func(string, int) bool]("name", "description")(myFunction)
```

This ensures that the function signature is checked at compile time, preventing runtime errors.

### Using `any` Instead of `interface{}`

This implementation uses Go's `any` type alias (introduced in Go 1.18) instead of `interface{}` for better readability and modern Go style.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details. 