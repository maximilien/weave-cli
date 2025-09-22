package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// TestFastMockIntegration runs fast integration tests with mock client
func TestFastMockIntegration(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "FastTestCollection", Type: "text", Description: "Fast test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("FastHealthCheck", func(t *testing.T) {
		healthCtx, healthCancel := context.WithTimeout(ctx, 1*time.Second)
		defer healthCancel()

		if err := client.Health(healthCtx); err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	t.Run("FastCollectionList", func(t *testing.T) {
		listCtx, listCancel := context.WithTimeout(ctx, 1*time.Second)
		defer listCancel()

		collections, err := client.ListCollections(listCtx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		if len(collections) != 1 {
			t.Errorf("Expected 1 collection, got %d", len(collections))
		}

		if collections[0] != "FastTestCollection" {
			t.Errorf("Expected 'FastTestCollection', got %s", collections[0])
		}
	})

	t.Run("FastDocumentOperations", func(t *testing.T) {
		docCtx, docCancel := context.WithTimeout(ctx, 2*time.Second)
		defer docCancel()

		// Add a document
		doc := mock.Document{
			ID:      "fast-test-doc-1",
			Content: "This is a fast test document",
			Metadata: map[string]interface{}{
				"test_type": "fast",
				"timestamp": time.Now().Unix(),
			},
		}

		if err := client.AddDocument(docCtx, "FastTestCollection", doc); err != nil {
			t.Errorf("Failed to add document: %v", err)
		}

		// List documents
		documents, err := client.ListDocuments(docCtx, "FastTestCollection", 5)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}

		if len(documents) != 1 {
			t.Errorf("Expected 1 document, got %d", len(documents))
		}

		// Get specific document
		retrievedDoc, err := client.GetDocument(docCtx, "FastTestCollection", "fast-test-doc-1")
		if err != nil {
			t.Errorf("Failed to get document: %v", err)
		}

		if retrievedDoc.ID != "fast-test-doc-1" {
			t.Errorf("Expected document ID 'fast-test-doc-1', got %s", retrievedDoc.ID)
		}

		// Delete document
		if err := client.DeleteDocument(docCtx, "FastTestCollection", "fast-test-doc-1"); err != nil {
			t.Errorf("Failed to delete document: %v", err)
		}

		// Verify deletion
		documents, err = client.ListDocuments(docCtx, "FastTestCollection", 5)
		if err != nil {
			t.Errorf("Failed to list documents after deletion: %v", err)
		}

		if len(documents) != 0 {
			t.Errorf("Expected 0 documents after deletion, got %d", len(documents))
		}
	})
}

// TestFastWeaviateIntegration runs fast integration tests with Weaviate (if configured)
func TestFastWeaviateIntegration(t *testing.T) {
	// Skip if no Weaviate configuration
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping Weaviate integration tests - missing WEAVIATE_URL or WEAVIATE_API_KEY")
	}
	
	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping Weaviate integration tests - invalid URL format")
	}

	cfg := &config.WeaviateCloudConfig{
		URL:                os.Getenv("WEAVIATE_URL"),
		APIKey:             os.Getenv("WEAVIATE_API_KEY"),
		CollectionNameTest: os.Getenv("WEAVIATE_COLLECTION_TEST"),
	}

	if cfg.CollectionNameTest == "" {
		cfg.CollectionNameTest = "WeaveCLITest"
	}

	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	// Use very short timeout for fast tests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("FastWeaviateHealthCheck", func(t *testing.T) {
		healthCtx, healthCancel := context.WithTimeout(ctx, 3*time.Second)
		defer healthCancel()

		collections, err := client.ListCollections(healthCtx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
		t.Logf("Found %d collections", len(collections))
	})

	t.Run("FastWeaviateCollectionCheck", func(t *testing.T) {
		listCtx, listCancel := context.WithTimeout(ctx, 3*time.Second)
		defer listCancel()

		collections, err := client.ListCollections(listCtx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
			return
		}

		// Check if test collection exists
		collectionExists := false
		for _, col := range collections {
			if col == cfg.CollectionNameTest {
				collectionExists = true
				break
			}
		}

		if !collectionExists {
			t.Logf("Test collection %s does not exist, skipping collection tests", cfg.CollectionNameTest)
			return
		}

		// Test document listing with very small limit for speed
		docCtx, docCancel := context.WithTimeout(ctx, 3*time.Second)
		defer docCancel()

		documents, err := client.ListDocuments(docCtx, cfg.CollectionNameTest, 1) // Only 1 document for speed
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
			return
		}

		t.Logf("Found %d documents in test collection", len(documents))

		// Test getting a specific document if any exist
		if len(documents) > 0 {
			getCtx, getCancel := context.WithTimeout(ctx, 2*time.Second)
			defer getCancel()

			docID := documents[0].ID
			doc, err := client.GetDocument(getCtx, cfg.CollectionNameTest, docID)
			if err != nil {
				t.Errorf("Failed to get document %s: %v", docID, err)
			} else {
				t.Logf("Successfully retrieved document: %s", doc.ID)
			}
		}
	})
}

// TestFastConfigIntegration tests config loading with fast operations
func TestFastConfigIntegration(t *testing.T) {
	t.Run("FastConfigLoad", func(t *testing.T) {
		// Test loading config with default files
		cfg, err := config.LoadConfig("", "")
		if err != nil {
			t.Logf("Config loading failed (expected if no config files): %v", err)
			return
		}

		if cfg == nil {
			t.Error("Config should not be nil")
		}
	})

	t.Run("FastEnvInterpolation", func(t *testing.T) {
		// Set test environment variable
		os.Setenv("FAST_TEST_VAR", "fast_test_value")
		defer os.Unsetenv("FAST_TEST_VAR")

		testCases := []struct {
			input    string
			expected string
		}{
			{"${FAST_TEST_VAR}", "fast_test_value"},
			{"${NONEXISTENT_VAR:-default}", "default"},
			{"simple string", "simple string"},
		}

		for _, tc := range testCases {
			result := config.InterpolateString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		}
	})
}

// BenchmarkFastOperations benchmarks fast operations
func BenchmarkFastOperations(b *testing.B) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "BenchmarkCollection", Type: "text", Description: "Benchmark collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.Health(ctx)
		}
	})

	b.Run("ListCollections", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.ListCollections(ctx)
		}
	})

	b.Run("AddDocument", func(b *testing.B) {
		doc := mock.Document{
			ID:      fmt.Sprintf("bench-doc-%d", b.N),
			Content: "Benchmark document content",
			Metadata: map[string]interface{}{
				"benchmark": true,
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.AddDocument(ctx, "BenchmarkCollection", doc)
		}
	})
}