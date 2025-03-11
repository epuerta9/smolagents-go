package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/epuerta9/smolagents-go/pkg/agents"
	"github.com/epuerta9/smolagents-go/pkg/models"
	"github.com/epuerta9/smolagents-go/pkg/tools"
)

// Document represents a simple document with an ID, title, and content.
type Document struct {
	ID      string
	Title   string
	Content string
}

// SimpleVectorDB is a mock vector database for demonstration purposes.
type SimpleVectorDB struct {
	documents []Document
}

// NewSimpleVectorDB creates a new SimpleVectorDB.
func NewSimpleVectorDB() *SimpleVectorDB {
	return &SimpleVectorDB{
		documents: []Document{
			{
				ID:    "doc1",
				Title: "Introduction to Go",
				Content: `Go is an open source programming language that makes it easy to build simple, 
reliable, and efficient software. Go was designed at Google in 2007 by Robert Griesemer, 
Rob Pike, and Ken Thompson. It is a statically typed language with syntax loosely derived 
from C, but with additional features such as garbage collection, type safety, and 
concurrent programming features.`,
			},
			{
				ID:    "doc2",
				Title: "Go Concurrency",
				Content: `Go provides excellent support for concurrent programming with goroutines and channels. 
Goroutines are lightweight threads managed by the Go runtime. Channels are the pipes that 
connect concurrent goroutines. You can send values into channels from one goroutine and 
receive those values into another goroutine.`,
			},
			{
				ID:    "doc3",
				Title: "Go Interfaces",
				Content: `Interfaces in Go provide a way to specify the behavior of an object: if something can do this, 
then it can be used here. Interfaces are named collections of method signatures. A type implements 
an interface by implementing its methods. Unlike in many other languages, there is no explicit 
declaration of intent, no "implements" keyword.`,
			},
		},
	}
}

// Search performs a simple keyword search on the documents.
func (db *SimpleVectorDB) Search(query string) []Document {
	var results []Document

	// Convert query to lowercase for case-insensitive search
	query = strings.ToLower(query)

	// Split query into keywords
	keywords := strings.Fields(query)

	// Search for documents containing all keywords
	for _, doc := range db.documents {
		content := strings.ToLower(doc.Content)
		title := strings.ToLower(doc.Title)

		// Check if all keywords are present in the document
		allPresent := true
		for _, keyword := range keywords {
			if !strings.Contains(content, keyword) && !strings.Contains(title, keyword) {
				allPresent = false
				break
			}
		}

		if allPresent {
			results = append(results, doc)
		}
	}

	return results
}

// SearchTool is a tool that searches the vector database.
func SearchTool(db *SimpleVectorDB) func(query string) string {
	return func(query string) string {
		results := db.Search(query)

		if len(results) == 0 {
			return "No documents found matching the query."
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d documents:\n\n", len(results)))

		for i, doc := range results {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.Title))
			sb.WriteString(fmt.Sprintf("   %s\n\n", doc.Content))
		}

		return sb.String()
	}
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("HF_API_KEY")
	if apiKey == "" {
		log.Fatal("HF_API_KEY environment variable is required")
	}

	// Create a model
	model := models.NewHfApiModel(
		"mistralai/Mistral-7B-Instruct-v0.2",
		models.WithApiKey(apiKey),
		models.WithMaxTokens(1024),
	)

	// Create a vector database
	db := NewSimpleVectorDB()

	// Create a search tool with explicit type parameter
	search := tools.CreateTool[func(string) string](
		"search",
		"Search for documents in the database.",
	)(SearchTool(db))

	// Create a ToolCallingAgent
	agent, err := agents.NewToolCallingAgent(
		[]tools.Tool{search},
		model,
		agents.WithMaxSteps(5),
		agents.WithSystemPrompt(`You are a helpful assistant that can search a knowledge base to answer questions.
When asked a question, you should:
1. Search the knowledge base for relevant information
2. Use the information to formulate a comprehensive answer
3. If the knowledge base doesn't contain relevant information, say so
4. Always cite your sources`),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Run the agent
	ctx := context.Background()
	task := "What are goroutines in Go and how do they relate to concurrency?"

	fmt.Printf("Question: %s\n\n", task)

	result, err := agent.Run(ctx, task)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}

	fmt.Printf("Answer: %s\n", result)
}
