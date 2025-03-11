package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/epuerta9/smolagents-go/pkg/agents"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// GetWeatherTool is a tool that gets the current weather for a location
func GetWeatherTool(location string, celsius bool) string {
	// This is a mock implementation
	// In a real application, you would call a weather API
	if celsius {
		return fmt.Sprintf("The current weather in %s is sunny with a temperature of 25 °C.", location)
	}
	return fmt.Sprintf("The current weather in %s is sunny with a temperature of 77 °F.", location)
}

// ConvertCurrencyTool is a tool that converts between currencies
func ConvertCurrencyTool(amount float64, fromCurrency, toCurrency string) string {
	// This is a mock implementation
	// In a real application, you would call a currency conversion API
	var rate float64
	switch {
	case fromCurrency == "USD" && toCurrency == "EUR":
		rate = 0.85
	case fromCurrency == "EUR" && toCurrency == "USD":
		rate = 1.18
	case fromCurrency == "USD" && toCurrency == "GBP":
		rate = 0.75
	case fromCurrency == "GBP" && toCurrency == "USD":
		rate = 1.33
	default:
		return fmt.Sprintf("Conversion from %s to %s is not supported.", fromCurrency, toCurrency)
	}

	converted := amount * rate
	return fmt.Sprintf("%.2f %s is equal to %.2f %s.", amount, fromCurrency, converted, toCurrency)
}

// GetJokeTool is a tool that returns a random joke
func GetJokeTool() string {
	// This is a mock implementation
	// In a real application, you would call a joke API
	jokes := []string{
		"Why don't scientists trust atoms? Because they make up everything!",
		"Why did the scarecrow win an award? Because he was outstanding in his field!",
		"I told my wife she was drawing her eyebrows too high. She looked surprised.",
		"What do you call a fake noodle? An impasta!",
		"Why did the bicycle fall over? Because it was two tired!",
	}

	// Return a "random" joke
	return jokes[time.Now().Unix()%int64(len(jokes))]
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create a model
	model := models.NewOpenAIModel(
		"gpt-4",
		models.WithApiKey(apiKey),
		models.WithMaxTokens(1024),
		models.WithHttpClient(&http.Client{Timeout: 60 * time.Second}),
	)

	// Create tools with explicit type parameters
	getWeather := tools.CreateTool[func(string, bool) string](
		"get_weather",
		"Get the current weather at the given location.",
	)(GetWeatherTool)

	convertCurrency := tools.CreateTool[func(float64, string, string) string](
		"convert_currency",
		"Converts a specified amount from one currency to another.",
	)(ConvertCurrencyTool)

	getJoke := tools.CreateTool[func() string](
		"get_joke",
		"Returns a random joke.",
	)(GetJokeTool)

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
	task := "What's the weather like in Paris? Also, can you convert 100 USD to EUR? And tell me a joke to brighten my day."

	fmt.Printf("Task: %s\n\n", task)

	result, err := agent.Run(ctx, task)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	fmt.Printf("Result: %s\n", result)
}
