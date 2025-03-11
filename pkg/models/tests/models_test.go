package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/epuerta9/smolagents-go/pkg/models"
)

// TestHfApiModelOptions tests the options for HfApiModel
func TestHfApiModelOptions(t *testing.T) {
	// Test default values
	model := models.NewHfApiModel("test-model")
	if model.Model != "test-model" {
		t.Errorf("Expected model name to be 'test-model', got '%s'", model.Model)
	}
	if model.MaxTokens != 1024 {
		t.Errorf("Expected max tokens to be 1024, got %d", model.MaxTokens)
	}
	if model.ApiKey != "" {
		t.Errorf("Expected API key to be empty, got '%s'", model.ApiKey)
	}
	if model.Client == nil {
		t.Error("Expected client to be non-nil")
	}

	// Test with options
	customClient := &http.Client{Timeout: 30 * 1000000000} // 30 seconds
	model = models.NewHfApiModel(
		"test-model",
		models.WithMaxTokens(2048),
		models.WithApiKey("test-api-key"),
		models.WithHttpClient(customClient),
	)
	if model.MaxTokens != 2048 {
		t.Errorf("Expected max tokens to be 2048, got %d", model.MaxTokens)
	}
	if model.ApiKey != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got '%s'", model.ApiKey)
	}
	if model.Client != customClient {
		t.Error("Expected client to be the custom client")
	}
}

// TestHfApiModelGenerate tests the Generate method of HfApiModel
func TestHfApiModelGenerate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-api-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Check request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := []map[string]interface{}{
			{
				"generated_text": "This is a test response",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a model with the test server URL
	model := models.NewHfApiModel(
		"test-model",
		models.WithApiKey("test-api-key"),
		models.WithHttpClient(server.Client()),
	)
	model.ApiURL = server.URL

	// Test Generate method
	messages := []models.Message{
		{Role: models.RoleUser, Content: "Hello"},
	}
	response, err := model.Generate(context.Background(), messages)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response != "This is a test response" {
		t.Errorf("Expected response to be 'This is a test response', got '%s'", response)
	}
}

// TestHfApiModelGenerateWithTools tests the GenerateWithTools method of HfApiModel
func TestHfApiModelGenerateWithTools(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-api-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Check request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Check if tools are included in the request
		parameters, ok := requestBody["parameters"].(map[string]interface{})
		if !ok {
			t.Error("Expected parameters in the request")
		} else {
			tools, ok := parameters["tools"].([]interface{})
			if !ok || len(tools) == 0 {
				t.Error("Expected tools to be included in the request parameters")
			}
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := []map[string]interface{}{
			{
				"generated_text": `{"tool": "test_tool", "args": {"arg1": "value1"}}`,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a model with the test server URL
	model := models.NewHfApiModel(
		"test-model",
		models.WithApiKey("test-api-key"),
		models.WithHttpClient(server.Client()),
	)
	model.ApiURL = server.URL

	// Test GenerateWithTools method
	messages := []models.Message{
		{Role: models.RoleUser, Content: "Hello"},
	}
	tools := []map[string]any{
		{
			"name":        "test_tool",
			"description": "A test tool",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"arg1": map[string]any{
						"type":        "string",
						"description": "Test argument",
					},
				},
			},
		},
	}
	response, err := model.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response != `{"tool": "test_tool", "args": {"arg1": "value1"}}` {
		t.Errorf("Expected tool call response, got '%s'", response)
	}
}

// TestHfApiModelGenerateError tests error handling in the Generate method
func TestHfApiModelGenerateError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create a model with the test server URL
	model := models.NewHfApiModel(
		"test-model",
		models.WithApiKey("test-api-key"),
		models.WithHttpClient(server.Client()),
	)
	model.ApiURL = server.URL

	// Test Generate method with error
	messages := []models.Message{
		{Role: models.RoleUser, Content: "Hello"},
	}
	_, err := model.Generate(context.Background(), messages)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
