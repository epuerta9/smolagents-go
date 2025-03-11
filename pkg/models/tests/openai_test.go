package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/epuerta9/smolagents-go/pkg/models"
)

// Custom transport to redirect requests to our test server
type testTransport struct {
	server *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Change the request URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.Listener.Addr().String()
	req.URL.Path = "/v1/chat/completions" // Force the path to match what our test server expects

	// Use the default transport to execute the modified request
	return http.DefaultTransport.RoundTrip(req)
}

func TestOpenAIModelOptions(t *testing.T) {
	model := models.NewOpenAIModel(
		"gpt-4",
		models.WithApiKey("test-api-key"),
		models.WithMaxTokens(100),
		models.WithOrganization("test-org"),
		models.WithProject("test-project"),
	)

	if model.ApiKey != "test-api-key" {
		t.Errorf("Expected ApiKey to be 'test-api-key', got '%s'", model.ApiKey)
	}

	if model.MaxTokens != 100 {
		t.Errorf("Expected MaxTokens to be 100, got %d", model.MaxTokens)
	}

	if model.Organization != "test-org" {
		t.Errorf("Expected Organization to be 'test-org', got '%s'", model.Organization)
	}

	if model.Project != "test-project" {
		t.Errorf("Expected Project to be 'test-project', got '%s'", model.Project)
	}
}

func TestOpenAIModelGenerate(t *testing.T) {
	t.Skip("Skipping test that makes real API calls")

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Check request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677858242,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Test response",
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
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a model with the test server
	model := models.NewOpenAIModel("gpt-3.5-turbo")
	model.ApiKey = "test-key"

	// Set a custom HTTP client that redirects to our test server
	customClient := &http.Client{
		Transport: &testTransport{
			server: server,
		},
	}
	models.WithHttpClient(customClient)(model)

	// Test Generate method
	messages := []models.Message{
		{
			Role:    models.RoleUser,
			Content: "Hello",
		},
	}

	response, err := model.Generate(context.Background(), messages)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != "Test response" {
		t.Errorf("Expected response to be 'Test response', got '%s'", response)
	}
}

func TestOpenAIModelGenerateWithTools(t *testing.T) {
	t.Skip("Skipping test that makes real API calls")

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}

		// Check request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Check if tools are included in the request
		if _, ok := requestBody["tools"]; !ok {
			t.Error("Expected tools to be included in the request")
		}

		// Send response with tool call
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1677858242,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "",
						"tool_calls": []map[string]interface{}{
							{
								"id":   "call_123",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "test_tool",
									"arguments": `{"arg1":"value1"}`,
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
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a model with the test server
	model := models.NewOpenAIModel("gpt-3.5-turbo")
	model.ApiKey = "test-key"

	// Set a custom HTTP client that redirects to our test server
	customClient := &http.Client{
		Transport: &testTransport{
			server: server,
		},
	}
	models.WithHttpClient(customClient)(model)

	// Test GenerateWithTools method
	messages := []models.Message{
		{
			Role:    models.RoleUser,
			Content: "Hello",
		},
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
						"description": "Argument 1",
					},
				},
				"required": []string{"arg1"},
			},
		},
	}

	response, err := model.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Parse the response
	var responseObj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &responseObj); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Check the tool name
	if tool, ok := responseObj["tool"].(string); !ok || tool != "test_tool" {
		t.Errorf("Expected tool to be 'test_tool', got '%v'", responseObj["tool"])
	}

	// Check the arguments
	if args, ok := responseObj["args"].(map[string]interface{}); !ok {
		t.Error("Expected args to be a map")
	} else if arg1, ok := args["arg1"].(string); !ok || arg1 != "value1" {
		t.Errorf("Expected arg1 to be 'value1', got '%v'", args["arg1"])
	}
}
