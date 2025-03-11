package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func TestOpenAISDKIntegration(t *testing.T) {
	// Create a test server that mocks the OpenAI API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a chat completion request
		if r.URL.Path == "/chat/completions" {
			// Parse the request body
			var requestBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&requestBody)

			// Check if this is a tool call request
			var response map[string]interface{}
			if _, hasTools := requestBody["tools"]; hasTools {
				// Return a tool call response
				response = map[string]interface{}{
					"id":      "chatcmpl-123",
					"object":  "chat.completion",
					"created": 1677858242,
					"model":   "gpt-4",
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"message": map[string]interface{}{
								"role": "assistant",
								"tool_calls": []map[string]interface{}{
									{
										"id":   "call_123",
										"type": "function",
										"function": map[string]interface{}{
											"name":      "get_weather",
											"arguments": `{"location":"London, UK"}`,
										},
									},
								},
							},
							"finish_reason": "tool_calls",
						},
					},
					"usage": map[string]interface{}{
						"prompt_tokens":     10,
						"completion_tokens": 20,
						"total_tokens":      30,
					},
				}
			} else {
				// Return a regular text response
				response = map[string]interface{}{
					"id":      "chatcmpl-123",
					"object":  "chat.completion",
					"created": 1677858242,
					"model":   "gpt-4",
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": "The three principles of object-oriented programming are encapsulation, inheritance, and polymorphism.",
							},
							"finish_reason": "stop",
						},
					},
					"usage": map[string]interface{}{
						"prompt_tokens":     10,
						"completion_tokens": 20,
						"total_tokens":      30,
					},
				}
			}

			// Send the response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return 404 for any other request
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a new OpenAI client that points to our test server
	client := openai.NewClient(
		option.WithAPIKey("test-api-key"),
		option.WithBaseURL(server.URL),
	)

	// Test simple text generation
	t.Run("Simple Text Generation", func(t *testing.T) {
		chatCompletion, err := client.Chat.Completions.New(
			context.Background(),
			openai.ChatCompletionNewParams{
				Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
					openai.UserMessage("What are the three principles of object-oriented programming?"),
				}),
				Model: openai.F("gpt-4"),
			},
		)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
			return
		}

		if len(chatCompletion.Choices) == 0 {
			t.Error("Expected at least one choice in response")
			return
		}

		expectedContent := "The three principles of object-oriented programming are encapsulation, inheritance, and polymorphism."
		if chatCompletion.Choices[0].Message.Content != expectedContent {
			t.Errorf("Expected content to be '%s', got '%s'", expectedContent, chatCompletion.Choices[0].Message.Content)
		}
	})

	// Test tool usage
	t.Run("Tool Usage", func(t *testing.T) {
		chatCompletion, err := client.Chat.Completions.New(
			context.Background(),
			openai.ChatCompletionNewParams{
				Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
					openai.UserMessage("What's the weather like in London?"),
				}),
				Model: openai.F("gpt-4"),
				Tools: openai.F([]openai.ChatCompletionToolParam{
					{
						Type: openai.F(openai.ChatCompletionToolTypeFunction),
						Function: openai.F(openai.FunctionDefinitionParam{
							Name:        openai.F("get_weather"),
							Description: openai.F("Get the current weather for a location"),
							Parameters: openai.F(openai.FunctionParameters{
								"type": "object",
								"properties": map[string]interface{}{
									"location": map[string]string{
										"type":        "string",
										"description": "The city and country, e.g., 'London, UK'",
									},
								},
								"required": []string{"location"},
							}),
						}),
					},
				}),
			},
		)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
			return
		}

		if len(chatCompletion.Choices) == 0 {
			t.Error("Expected at least one choice in response")
			return
		}

		if len(chatCompletion.Choices[0].Message.ToolCalls) == 0 {
			t.Error("Expected at least one tool call in response")
			return
		}

		toolCall := chatCompletion.Choices[0].Message.ToolCalls[0]
		if toolCall.Function.Name != "get_weather" {
			t.Errorf("Expected tool name to be 'get_weather', got '%s'", toolCall.Function.Name)
		}

		// Parse the arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			t.Errorf("Failed to parse tool arguments: %v", err)
			return
		}

		if location, ok := args["location"].(string); !ok || location != "London, UK" {
			t.Errorf("Expected location to be 'London, UK', got '%v'", args["location"])
		}
	})
}
