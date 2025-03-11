package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestMessageRoles ensures message roles are correctly defined
func TestMessageRoles(t *testing.T) {
	if RoleSystem != "system" {
		t.Errorf("Expected RoleSystem to be 'system', got '%s'", RoleSystem)
	}

	if RoleUser != "user" {
		t.Errorf("Expected RoleUser to be 'user', got '%s'", RoleUser)
	}

	if RoleAssistant != "assistant" {
		t.Errorf("Expected RoleAssistant to be 'assistant', got '%s'", RoleAssistant)
	}

	if RoleTool != "tool" {
		t.Errorf("Expected RoleTool to be 'tool', got '%s'", RoleTool)
	}
}

// TestHfApiModelOptions tests the option functions for HfApiModel
func TestHfApiModelOptions(t *testing.T) {
	// Create a model with default options
	model := NewHfApiModel("test-model")

	// Check default values
	if model.MaxTokens != 1024 {
		t.Errorf("Expected default MaxTokens to be 1024, got %d", model.MaxTokens)
	}

	if model.ApiKey != "" {
		t.Errorf("Expected default ApiKey to be empty, got '%s'", model.ApiKey)
	}

	// Apply options
	customClient := &http.Client{Timeout: 30 * time.Second}
	model = NewHfApiModel("test-model",
		WithMaxTokens(2048),
		WithApiKey("test-api-key"),
		WithHttpClient(customClient),
	)

	// Check applied options
	if model.MaxTokens != 2048 {
		t.Errorf("Expected MaxTokens to be 2048, got %d", model.MaxTokens)
	}

	if model.ApiKey != "test-api-key" {
		t.Errorf("Expected ApiKey to be 'test-api-key', got '%s'", model.ApiKey)
	}

	if model.Client != customClient {
		t.Error("Expected Client to be the custom HTTP client")
	}
}

// TestHfApiModelGenerate tests the Generate method
func TestHfApiModelGenerate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Error decoding request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check request structure
		inputs, ok := reqBody["inputs"]
		if !ok {
			t.Error("Expected request to have 'inputs' field")
		}

		// Check that inputs is an array of messages
		inputsArr, ok := inputs.([]interface{})
		if !ok || len(inputsArr) == 0 {
			t.Error("Expected inputs to be a non-empty array")
		}

		// Send response
		response := []map[string]string{
			{"generated_text": "This is a test response"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create model that uses the test server
	model := NewHfApiModel("test-model",
		WithApiKey("test-api-key"),
		WithMaxTokens(100),
	)
	model.ApiURL = server.URL

	// Create test messages
	messages := []Message{
		{Role: RoleSystem, Content: "You are a helpful assistant."},
		{Role: RoleUser, Content: "Hello, world!"},
	}

	// Call Generate
	response, err := model.Generate(context.Background(), messages)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != "This is a test response" {
		t.Errorf("Expected response 'This is a test response', got '%s'", response)
	}
}

// TestHfApiModelGenerateWithTools tests the GenerateWithTools method
func TestHfApiModelGenerateWithTools(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Error decoding request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check if tools are present in parameters
		params, ok := reqBody["parameters"].(map[string]interface{})
		if !ok {
			t.Error("Expected request to have 'parameters' field")
		}

		tools, ok := params["tools"]
		if !ok {
			t.Error("Expected parameters to have 'tools' field")
		}

		// Check if tools is an array
		toolsArr, ok := tools.([]interface{})
		if !ok || len(toolsArr) == 0 {
			t.Error("Expected tools to be a non-empty array")
		}

		// Send response
		response := []map[string]string{
			{"generated_text": "I'll use a tool to help with that."},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create model that uses the test server
	model := NewHfApiModel("test-model", WithApiKey("test-api-key"))
	model.ApiURL = server.URL

	// Create test messages
	messages := []Message{
		{Role: RoleSystem, Content: "You are a helpful assistant."},
		{Role: RoleUser, Content: "What's the weather like?"},
	}

	// Create test tools
	tools := []map[string]any{
		{
			"type": "function",
			"function": map[string]any{
				"name":        "get_weather",
				"description": "Get the current weather",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Call GenerateWithTools
	response, err := model.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != "I'll use a tool to help with that." {
		t.Errorf("Expected specific response, got '%s'", response)
	}
}

// TestModelErrorHandling tests error handling in the model
func TestModelErrorHandling(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create model that uses the test server
	model := NewHfApiModel("test-model")
	model.ApiURL = server.URL

	// Create test messages
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	// Call Generate and expect an error
	_, err := model.Generate(context.Background(), messages)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestEmptyResponseHandling tests handling of empty responses from the API
func TestEmptyResponseHandling(t *testing.T) {
	// Create a test server that returns an empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	// Create model that uses the test server
	model := NewHfApiModel("test-model")
	model.ApiURL = server.URL

	// Create test messages
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	// Call Generate and expect an error about empty response
	_, err := model.Generate(context.Background(), messages)
	if err == nil {
		t.Error("Expected error about empty response, got nil")
	}
}
