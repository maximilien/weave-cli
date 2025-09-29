// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max


package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// TestWeaviateIntegration runs fast integration tests with Weaviate
func TestWeaviateIntegration(t *testing.T) {
	// Skip if no Weaviate configuration
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping Weaviate integration tests - missing WEAVIATE_URL or WEAVIATE_API_KEY")
	}

	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping Weaviate integration tests - invalid URL format")
	}

	// Create test configuration
	cfg := &config.VectorDBConfig{
		Name:   "test-cloud",
		Type:   config.VectorDBTypeCloud,
		URL:    os.Getenv("WEAVIATE_URL"),
		APIKey: os.Getenv("WEAVIATE_API_KEY"),
		Collections: []config.Collection{
			{Name: os.Getenv("WEAVIATE_COLLECTION_TEST"), Type: "text"},
		},
	}

	if cfg.Collections[0].Name == "" {
		cfg.Collections[0].Name = "WeaveCLITest"
	}

	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	// Use very short timeout for fast tests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCollection := cfg.Collections[0].Name

	t.Run("FastHealthCheck", func(t *testing.T) {
		// Quick health check with minimal timeout
		healthCtx, healthCancel := context.WithTimeout(ctx, 2*time.Second)
		defer healthCancel()

		collections, err := client.ListCollections(healthCtx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
		t.Logf("Found %d collections", len(collections))
	})

	t.Run("FastCollectionOperations", func(t *testing.T) {
		// Test collection listing with short timeout
		listCtx, listCancel := context.WithTimeout(ctx, 2*time.Second)
		defer listCancel()

		collections, err := client.ListCollections(listCtx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
			return
		}

		// Check if test collection exists
		collectionExists := false
		for _, col := range collections {
			if col == testCollection {
				collectionExists = true
				break
			}
		}

		if !collectionExists {
			t.Logf("Test collection %s does not exist, skipping collection tests", testCollection)
			return
		}

		// Test document listing with very small limit for speed
		docCtx, docCancel := context.WithTimeout(ctx, 2*time.Second)
		defer docCancel()

		documents, err := client.ListDocuments(docCtx, testCollection, 1) // Only 1 document for speed
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
			return
		}

		t.Logf("Found %d documents in test collection", len(documents))

		// Test getting a specific document if any exist
		if len(documents) > 0 {
			getCtx, getCancel := context.WithTimeout(ctx, 1*time.Second)
			defer getCancel()

			docID := documents[0].ID
			doc, err := client.GetDocument(getCtx, testCollection, docID)
			if err != nil {
				t.Errorf("Failed to get document %s: %v", docID, err)
			} else {
				t.Logf("Successfully retrieved document: %s", doc.ID)
			}
		}
	})
}

// TestWeaviateConnectionSpeed tests connection speed and basic operations
func TestWeaviateConnectionSpeed(t *testing.T) {
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping connection speed tests - missing Weaviate configuration")
	}

	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping connection speed tests - invalid URL format")
	}

	cfg := &config.VectorDBConfig{
		Name:   "test-cloud",
		Type:   config.VectorDBTypeCloud,
		URL:    os.Getenv("WEAVIATE_URL"),
		APIKey: os.Getenv("WEAVIATE_API_KEY"),
		Collections: []config.Collection{
			{Name: os.Getenv("WEAVIATE_COLLECTION_TEST"), Type: "text"},
		},
	}

	if cfg.Collections[0].Name == "" {
		cfg.Collections[0].Name = "WeaveCLITest"
	}

	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Measure connection speed
	start := time.Now()
	collections, err := client.ListCollections(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Connection test failed: %v", err)
		return
	}

	t.Logf("Connection established in %v, found %d collections", duration, len(collections))

	// Test should complete quickly
	if duration > 5*time.Second {
		t.Errorf("Connection too slow: %v", duration)
	}
}

// TestWeaviateErrorHandling tests error handling scenarios
func TestWeaviateErrorHandling(t *testing.T) {
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping error handling tests - missing Weaviate configuration")
	}

	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping error handling tests - invalid URL format")
	}

	cfg := &config.VectorDBConfig{
		Name:   "test-cloud",
		Type:   config.VectorDBTypeCloud,
		URL:    os.Getenv("WEAVIATE_URL"),
		APIKey: os.Getenv("WEAVIATE_API_KEY"),
		Collections: []config.Collection{
			{Name: os.Getenv("WEAVIATE_COLLECTION_TEST"), Type: "text"},
		},
	}

	if cfg.Collections[0].Name == "" {
		cfg.Collections[0].Name = "WeaveCLITest"
	}

	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test non-existent collection
	_, err = client.ListDocuments(ctx, "NonExistentCollection", 5)
	if err == nil {
		t.Error("Expected error for non-existent collection, got nil")
	} else {
		t.Logf("Correctly handled non-existent collection: %v", err)
	}

	// Test non-existent document
	_, err = client.GetDocument(ctx, cfg.Collections[0].Name, "non-existent-doc-id")
	if err == nil {
		t.Error("Expected error for non-existent document, got nil")
	} else {
		t.Logf("Correctly handled non-existent document: %v", err)
	}
}

// BenchmarkWeaviateOperations benchmarks key operations
func BenchmarkWeaviateOperations(b *testing.B) {
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		b.Skip("Skipping benchmarks - missing Weaviate configuration")
	}

	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		b.Skip("Skipping benchmarks - invalid URL format")
	}

	cfg := &config.VectorDBConfig{
		Name:   "test-cloud",
		Type:   config.VectorDBTypeCloud,
		URL:    os.Getenv("WEAVIATE_URL"),
		APIKey: os.Getenv("WEAVIATE_API_KEY"),
		Collections: []config.Collection{
			{Name: os.Getenv("WEAVIATE_COLLECTION_TEST"), Type: "text"},
		},
	}

	if cfg.Collections[0].Name == "" {
		cfg.Collections[0].Name = "WeaveCLITest"
	}

	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		b.Fatalf("Failed to create Weaviate client: %v", err)
	}

	ctx := context.Background()

	b.Run("ListCollections", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.ListCollections(ctx)
			if err != nil {
				b.Errorf("ListCollections failed: %v", err)
			}
		}
	})

	b.Run("ListDocuments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.ListDocuments(ctx, cfg.Collections[0].Name, 5)
			if err != nil {
				b.Errorf("ListDocuments failed: %v", err)
			}
		}
	})
}
