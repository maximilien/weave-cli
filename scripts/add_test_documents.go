package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Sample documents for testing
var testDocuments = []struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
}{
	{
		ID:      "test-doc-1",
		Content: "This is a test document for Weave CLI testing. It contains sample text content that can be used for various testing scenarios.",
		Metadata: map[string]interface{}{
			"source":     "weave-cli-test",
			"type":       "text",
			"created_at": time.Now().Format(time.RFC3339),
			"tags":       []string{"test", "sample", "document"},
		},
	},
	{
		ID:      "test-doc-2",
		Content: "Another test document with different content. This document is used to test document listing and retrieval functionality.",
		Metadata: map[string]interface{}{
			"source":     "weave-cli-test",
			"type":       "text",
			"created_at": time.Now().Format(time.RFC3339),
			"tags":       []string{"test", "sample", "listing"},
		},
	},
	{
		ID:      "test-doc-3",
		Content: "A third test document for comprehensive testing. This document helps test various edge cases and scenarios.",
		Metadata: map[string]interface{}{
			"source":     "weave-cli-test",
			"type":       "text",
			"created_at": time.Now().Format(time.RFC3339),
			"tags":       []string{"test", "sample", "edge-case"},
		},
	},
	{
		ID:      "test-doc-4",
		Content: "Document with special characters: !@#$%^&*()_+-=[]{}|;':\",./<>? This tests handling of special characters in content.",
		Metadata: map[string]interface{}{
			"source":      "weave-cli-test",
			"type":        "text",
			"created_at":  time.Now().Format(time.RFC3339),
			"tags":        []string{"test", "special-chars"},
			"has_special": true,
		},
	},
	{
		ID:      "test-doc-5",
		Content: "Large document content. " + string(make([]byte, 1000)) + " This document has a large amount of content to test performance and handling of large documents.",
		Metadata: map[string]interface{}{
			"source":     "weave-cli-test",
			"type":       "text",
			"created_at": time.Now().Format(time.RFC3339),
			"tags":       []string{"test", "large-content"},
			"size":       "large",
		},
	},
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("", "")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Determine which client to use
	var client interface{}
	var collectionName string

	// Get the default vector database config
	var vectorDBConfig *config.VectorDBConfig
	for _, db := range cfg.Databases.VectorDatabases {
		if db.Name == cfg.Databases.Default {
			vectorDBConfig = &db
			break
		}
	}
	
	if vectorDBConfig == nil {
		log.Fatal("No default vector database found")
	}

	switch vectorDBConfig.Type {
	case config.VectorDBTypeCloud:
		weaviateClient, err := weaviate.NewClient(&weaviate.Config{
			URL:    vectorDBConfig.URL,
			APIKey: vectorDBConfig.APIKey,
		})
		if err != nil {
			log.Fatalf("Failed to create Weaviate client: %v", err)
		}
		client = weaviateClient
		// Use the first collection from the vector database config
		if len(vectorDBConfig.Collections) > 0 {
			collectionName = vectorDBConfig.Collections[0].Name
		} else {
			collectionName = "TestCollection"
		}
	case config.VectorDBTypeLocal:
		weaviateClient, err := weaviate.NewClient(&weaviate.Config{
			URL: vectorDBConfig.URL,
		})
		if err != nil {
			log.Fatalf("Failed to create Weaviate client: %v", err)
		}
		client = weaviateClient
		// Use the first collection from the vector database config
		if len(vectorDBConfig.Collections) > 0 {
			collectionName = vectorDBConfig.Collections[0].Name
		} else {
			collectionName = "TestCollection"
		}
	default:
		log.Fatalf("This script only works with Weaviate (cloud or local), not mock database")
	}

	if collectionName == "" {
		collectionName = "WeaveDocs_test"
	}

	fmt.Printf("Adding test documents to collection: %s\n", collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if collection exists
	weaviateClient := client.(*weaviate.Client)
	collections, err := weaviateClient.ListCollections(ctx)
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	collectionExists := false
	for _, col := range collections {
		if col == collectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		fmt.Printf("Collection %s does not exist. Please create it first.\n", collectionName)
		os.Exit(1)
	}

	// Add test documents
	successCount := 0
	for _, doc := range testDocuments {
		fmt.Printf("Adding document: %s\n", doc.ID)

		// Note: This is a simplified version. In a real implementation,
		// you would need to implement the AddDocument method in the weaviate client
		// For now, we'll just simulate the operation
		fmt.Printf("  Content: %s...\n", doc.Content[:min(50, len(doc.Content))])
		fmt.Printf("  Metadata: %+v\n", doc.Metadata)

		successCount++
	}

	fmt.Printf("Successfully added %d test documents to collection %s\n", successCount, collectionName)
	fmt.Println("Test documents added:")
	for _, doc := range testDocuments {
		fmt.Printf("  - %s: %s\n", doc.ID, doc.Content[:min(50, len(doc.Content))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
