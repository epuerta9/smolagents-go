package main

import (
	"context"
	"fmt"
	"log"

	"github.com/epuerta9/smolagents-go/pkg/models"
)

// weatherTool is an example tool that could be used to get weather information
var weatherTool = map[string]any{
	"name":        "get_weather",
	"description": "Get the current weather for a location",
	"parameters": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "The city and country, e.g., 'London, UK'",
			},
		},
		"required": []string{"location"},
	},
}

func main() {
	// Create a new OpenAI model instance
	model := models.NewOpenAIModel(
		"gpt-3.5-turbo",
		models.WithMaxTokens(1000),
		// API key will be automatically loaded from OPENAI_API_KEY environment variable
	)

	// Example 1: Simple text generation
	messages := []models.Message{
		{
			Role:    models.RoleSystem,
			Content: "You are a helpful assistant.",
		},
		{
			Role:    models.RoleUser,
			Content: "What are the three principles of object-oriented programming?",
		},
	}

	fmt.Println("Example 1: Simple text generation")
	fmt.Println("Question: What are the three principles of object-oriented programming?")

	response, err := model.Generate(context.Background(), messages)
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Printf("\nResponse:\n%s\n\n", response)

	// Example 2: Using tools
	messages = []models.Message{
		{
			Role:    models.RoleSystem,
			Content: "You are a helpful weather assistant. Use the get_weather tool to check the weather.",
		},
		{
			Role:    models.RoleUser,
			Content: "What's the weather like in London?",
		},
	}

	tools := []map[string]any{weatherTool}

	fmt.Println("Example 2: Using tools")
	fmt.Println("Question: What's the weather like in London?")

	toolResponse, err := model.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		log.Fatalf("Failed to generate tool response: %v", err)
	}

	fmt.Printf("\nTool Response:\n%s\n", toolResponse)
}
