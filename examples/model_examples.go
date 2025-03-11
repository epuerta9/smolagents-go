// Example of using different model types with smolagents
//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/epuerta9/smolagents-go/pkg/agents"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// GetWeather is a simple tool that returns the weather for a location
func GetWeather(location string) string {
	return fmt.Sprintf("The weather in %s is sunny", location)
}

func main() {
	// Create a simple tool
	weatherTool := tools.CreateTool[func(string) string](
		"get_weather",
		"Get the current weather at the given location",
	)(GetWeather)

	// Example 1: Using HfApiModel
	fmt.Println("Example 1: Using HfApiModel")
	hfModel := models.NewHfApiModel(
		"mistralai/Mistral-7B-Instruct-v0.2",
		models.WithApiKey(os.Getenv("HF_API_KEY")),
		models.WithMaxTokens(1024),
	)
	runExample(weatherTool, hfModel, "What's the weather like in Paris?")

	// Example 2: Using OpenAIModel
	fmt.Println("\nExample 2: Using OpenAIModel")
	openaiModel := models.NewOpenAIModel(
		"gpt-3.5-turbo",
		models.WithApiKey(os.Getenv("OPENAI_API_KEY")),
		models.WithMaxTokens(1024),
	)
	runExample(weatherTool, openaiModel, "What's the weather like in London?")

	// Example 3: Using AzureOpenAIModel
	fmt.Println("\nExample 3: Using AzureOpenAIModel")
	azureModel := models.NewAzureOpenAIModel(
		"gpt-4", // This should be your deployment name
		models.WithApiKey(os.Getenv("AZURE_OPENAI_API_KEY")),
		models.WithAzureEndpoint(os.Getenv("AZURE_OPENAI_ENDPOINT")),
		models.WithMaxTokens(1024),
	)
	runExample(weatherTool, azureModel, "What's the weather like in Tokyo?")
}

func runExample(tool tools.Tool, model models.Model, query string) {
	// Create a ToolCallingAgent with the tool and model
	agent, err := agents.NewToolCallingAgent([]tools.Tool{tool}, model)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Run the agent
	fmt.Printf("Query: %s\n", query)
	result, err := agent.Run(context.Background(), query)
	if err != nil {
		log.Printf("Error running agent: %v", err)
		return
	}

	// Print the result
	fmt.Printf("Result: %v\n", result)
}
