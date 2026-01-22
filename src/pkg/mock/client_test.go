// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mock

import (
	"context"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestNewClient(t *testing.T) {
	cfg := &config.MockConfig{
		Collections: []config.MockCollection{
			{Name: "TestCollection1"},
			{Name: "TestCollection2"},
		},
		EmbeddingDimension: 384,
		SimulateEmbeddings: true,
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("Expected client, got nil")
	}

	if client.config != cfg {
		t.Error("Expected config to be set")
	}

	if len(client.collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(client.collections))
	}

	// Check that collections were initialized
	if _, exists := client.collections["TestCollection1"]; !exists {
		t.Error("Expected TestCollection1 to exist")
	}
	if _, exists := client.collections["TestCollection2"]; !exists {
		t.Error("Expected TestCollection2 to exist")
	}
}

func TestHealth(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListCollections(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	collections, err := client.ListCollections(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(collections) == 0 {
		t.Error("Expected at least one collection")
	}

	// Should include WeaveDocs from test data
	found := false
	for _, name := range collections {
		if name == "WeaveDocs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected WeaveDocs collection from test data")
	}
}

func TestDeleteCollection(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		shouldError    bool
	}{
		{
			name:           "Delete existing collection",
			collectionName: "WeaveDocs",
			shouldError:    false,
		},
		{
			name:           "Delete non-existent collection",
			collectionName: "NonExistent",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			err := client.DeleteCollection(ctx, tt.collectionName)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify collection was deleted
				if _, exists := client.collections[tt.collectionName]; exists {
					t.Errorf("Collection %s should have been deleted", tt.collectionName)
				}
			}
		})
	}
}

func TestListDocuments(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	tests := []struct {
		name           string
		collectionName string
		limit          int
		shouldError    bool
		checkLimit     bool
	}{
		{
			name:           "List all documents",
			collectionName: "WeaveDocs",
			limit:          0,
			shouldError:    false,
			checkLimit:     false,
		},
		{
			name:           "List with limit",
			collectionName: "WeaveDocs",
			limit:          2,
			shouldError:    false,
			checkLimit:     true,
		},
		{
			name:           "List from non-existent collection",
			collectionName: "NonExistent",
			limit:          0,
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := client.ListDocuments(ctx, tt.collectionName, tt.limit)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkLimit && len(docs) != tt.limit {
				t.Errorf("Expected %d documents, got %d", tt.limit, len(docs))
			}
		})
	}
}

func TestCountDocuments(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	tests := []struct {
		name           string
		collectionName string
		shouldError    bool
		expectCount    bool
	}{
		{
			name:           "Count documents in existing collection",
			collectionName: "WeaveDocs",
			shouldError:    false,
			expectCount:    true,
		},
		{
			name:           "Count in non-existent collection",
			collectionName: "NonExistent",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := client.CountDocuments(ctx, tt.collectionName)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectCount && count == 0 {
				t.Error("Expected non-zero count from test data")
			}
		})
	}
}

func TestGetDocument(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	tests := []struct {
		name           string
		collectionName string
		documentID     string
		shouldError    bool
	}{
		{
			name:           "Get existing document",
			collectionName: "WeaveDocs",
			documentID:     "doc1-chunk1",
			shouldError:    false,
		},
		{
			name:           "Get non-existent document",
			collectionName: "WeaveDocs",
			documentID:     "nonexistent",
			shouldError:    true,
		},
		{
			name:           "Get from non-existent collection",
			collectionName: "NonExistent",
			documentID:     "doc1",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := client.GetDocument(ctx, tt.collectionName, tt.documentID)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if doc == nil {
				t.Error("Expected document, got nil")
				return
			}

			if doc.ID != tt.documentID {
				t.Errorf("Expected document ID %s, got %s", tt.documentID, doc.ID)
			}
		})
	}
}

func TestAddDocument(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	doc := Document{
		ID:      "test-doc-1",
		Content: "Test content",
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	// Add to existing collection
	err := client.AddDocument(ctx, "WeaveDocs", doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify document was added
	retrieved, err := client.GetDocument(ctx, "WeaveDocs", "test-doc-1")
	if err != nil {
		t.Errorf("Failed to retrieve added document: %v", err)
	}
	if retrieved.Content != doc.Content {
		t.Errorf("Expected content %q, got %q", doc.Content, retrieved.Content)
	}

	// Try to add to non-existent collection
	err = client.AddDocument(ctx, "NonExistent", doc)
	if err == nil {
		t.Error("Expected error when adding to non-existent collection, got nil")
	}
}

func TestDeleteDocument(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	// Add a test document first
	doc := Document{
		ID:      "delete-test",
		Content: "To be deleted",
	}
	_ = client.AddDocument(ctx, "WeaveDocs", doc)

	// Delete it
	err := client.DeleteDocument(ctx, "WeaveDocs", "delete-test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify it's gone
	_, err = client.GetDocument(ctx, "WeaveDocs", "delete-test")
	if err == nil {
		t.Error("Expected error when getting deleted document, got nil")
	}

	// Try to delete non-existent document
	err = client.DeleteDocument(ctx, "WeaveDocs", "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent document, got nil")
	}
}

func TestDeleteDocumentsByMetadata(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	// Add test documents with metadata
	docs := []Document{
		{
			ID:      "meta-doc-1",
			Content: "Content 1",
			Metadata: map[string]interface{}{
				"category": "test",
				"tag":      "v1",
			},
		},
		{
			ID:      "meta-doc-2",
			Content: "Content 2",
			Metadata: map[string]interface{}{
				"category": "test",
				"tag":      "v2",
			},
		},
		{
			ID:      "meta-doc-3",
			Content: "Content 3",
			Metadata: map[string]interface{}{
				"category": "prod",
				"tag":      "v1",
			},
		},
	}

	for _, doc := range docs {
		_ = client.AddDocument(ctx, "WeaveDocs", doc)
	}

	// Delete documents with category=test
	count, err := client.DeleteDocumentsByMetadata(ctx, "WeaveDocs", []string{"category=test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 documents deleted, got %d", count)
	}

	// Verify they're gone
	_, err = client.GetDocument(ctx, "WeaveDocs", "meta-doc-1")
	if err == nil {
		t.Error("Expected meta-doc-1 to be deleted")
	}

	// Verify meta-doc-3 still exists
	doc, err := client.GetDocument(ctx, "WeaveDocs", "meta-doc-3")
	if err != nil {
		t.Errorf("Expected meta-doc-3 to exist: %v", err)
	}
	if doc == nil || doc.ID != "meta-doc-3" {
		t.Error("meta-doc-3 should still exist")
	}

	// Test invalid filter format
	_, err = client.DeleteDocumentsByMetadata(ctx, "WeaveDocs", []string{"invalid-filter"})
	if err == nil {
		t.Error("Expected error for invalid filter format, got nil")
	}
}

func TestGetDocumentsByMetadata(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	// Add test documents
	docs := []Document{
		{
			ID:      "filter-doc-1",
			Content: "Content 1",
			Metadata: map[string]interface{}{
				"type": "article",
				"year": "2025",
			},
		},
		{
			ID:      "filter-doc-2",
			Content: "Content 2",
			Metadata: map[string]interface{}{
				"type": "book",
				"year": "2025",
			},
		},
	}

	for _, doc := range docs {
		_ = client.AddDocument(ctx, "WeaveDocs", doc)
	}

	// Filter by type=article
	results, err := client.GetDocumentsByMetadata(ctx, "WeaveDocs", []string{"type=article"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "filter-doc-1" {
		t.Errorf("Expected filter-doc-1, got %s", results[0].ID)
	}

	// Filter by year=2025 (should get both)
	results, err = client.GetDocumentsByMetadata(ctx, "WeaveDocs", []string{"year=2025"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}

	// Invalid filter format
	_, err = client.GetDocumentsByMetadata(ctx, "WeaveDocs", []string{"invalid"})
	if err == nil {
		t.Error("Expected error for invalid filter format, got nil")
	}
}

func TestGetCollectionStats(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	stats, err := client.GetCollectionStats(ctx, "WeaveDocs")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	// Check expected fields
	if name, ok := stats["name"].(string); !ok || name != "WeaveDocs" {
		t.Errorf("Expected name 'WeaveDocs', got %v", stats["name"])
	}

	if count, ok := stats["document_count"].(int); !ok || count == 0 {
		t.Errorf("Expected non-zero document_count, got %v", stats["document_count"])
	}

	if dim, ok := stats["embedding_dimension"].(int); !ok || dim <= 0 {
		t.Errorf("Expected positive embedding_dimension, got %v", stats["embedding_dimension"])
	}

	// Test non-existent collection
	_, err = client.GetCollectionStats(ctx, "NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent collection, got nil")
	}
}

func TestCreateCollection(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	err := client.CreateCollection(ctx, "NewCollection", "text-embedding-ada-002", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify collection exists
	collections, _ := client.ListCollections(ctx)
	found := false
	for _, name := range collections {
		if name == "NewCollection" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected NewCollection to exist after creation")
	}

	// Try to create duplicate
	err = client.CreateCollection(ctx, "NewCollection", "text-embedding-ada-002", nil)
	if err == nil {
		t.Error("Expected error when creating duplicate collection, got nil")
	}
}

func TestDeleteAllDocuments(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	// Add some documents
	for i := 0; i < 5; i++ {
		doc := Document{
			ID:      "bulk-" + string(rune('0'+i)),
			Content: "Content",
		}
		_ = client.AddDocument(ctx, "WeaveDocs", doc)
	}

	// Delete all
	err := client.DeleteAllDocuments(ctx, "WeaveDocs")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify collection is empty
	count, _ := client.CountDocuments(ctx, "WeaveDocs")
	if count != 0 {
		t.Errorf("Expected 0 documents after DeleteAll, got %d", count)
	}

	// Test non-existent collection
	err = client.DeleteAllDocuments(ctx, "NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent collection, got nil")
	}
}

func TestQuery(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	// Add test documents
	docs := []Document{
		{
			ID:      "query-doc-1",
			Content: "Machine learning algorithms are powerful",
		},
		{
			ID:      "query-doc-2",
			Content: "Deep learning neural networks",
		},
		{
			ID:      "query-doc-3",
			Content: "Natural language processing",
		},
	}

	for _, doc := range docs {
		_ = client.AddDocument(ctx, "WeaveDocs", doc)
	}

	options := weaviate.QueryOptions{
		TopK: 2,
	}

	results, err := client.Query(ctx, "WeaveDocs", "machine learning", options)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected query results, got none")
	}

	// Results should have scores
	for _, result := range results {
		if result.Score == 0 {
			t.Error("Expected non-zero score for query result")
		}
	}

	// Test query with non-existent collection
	_, err = client.Query(ctx, "NonExistent", "test", options)
	if err == nil {
		t.Error("Expected error for non-existent collection, got nil")
	}
}

func TestCalculateMockScore(t *testing.T) {
	client := createTestClient()

	doc := Document{
		ID:      "score-test",
		Content: "This is about machine learning and artificial intelligence",
	}

	tests := []struct {
		name       string
		queryWords []string
		expectZero bool
	}{
		{
			name:       "Matching words",
			queryWords: []string{"machine", "learning"},
			expectZero: false,
		},
		{
			name:       "No matching words",
			queryWords: []string{"quantum", "physics"},
			expectZero: true,
		},
		{
			name:       "Partial match",
			queryWords: []string{"machine", "quantum"},
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := client.CalculateMockScore(doc, tt.queryWords)

			if tt.expectZero && score != 0 {
				t.Errorf("Expected zero score, got %.2f", score)
			}
			if !tt.expectZero && score == 0 {
				t.Errorf("Expected non-zero score, got 0")
			}
		})
	}
}

func TestCalculateMockScoreWithMetadata(t *testing.T) {
	client := createTestClient()

	doc := Document{
		ID:      "meta-score-test",
		Content: "Basic content about programming",
		Metadata: map[string]interface{}{
			"tags":     "machine learning, AI, neural networks",
			"category": "technology",
		},
	}

	// Query with words that are ONLY in metadata, not in content
	queryWords := []string{"machine", "learning"}

	// Score with metadata search - should find words in metadata
	scoreWithMeta := client.CalculateMockScoreWithMetadata(doc, queryWords, true)

	// Score without metadata search - should NOT find words (only in content)
	scoreWithoutMeta := client.CalculateMockScoreWithMetadata(doc, queryWords, false)

	// Metadata search should boost score since words are only in metadata
	if scoreWithMeta <= scoreWithoutMeta {
		t.Errorf("Expected metadata search to boost score: with=%f, without=%f", scoreWithMeta, scoreWithoutMeta)
	}

	// Test with no matching words
	noMatchScore := client.CalculateMockScoreWithMetadata(doc, []string{"quantum", "physics"}, true)
	if noMatchScore > 0 {
		t.Errorf("Expected zero score for no match, got %.2f", noMatchScore)
	}
}

func TestContextTimeout(t *testing.T) {
	client := createTestClient()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operations should still work (mocked timeouts don't actually fail)
	// Just verify they don't panic
	_, _ = client.ListCollections(ctx)
	_ = client.Health(ctx)
}

// Helper function to create a test client
func createTestClient() *Client {
	cfg := &config.MockConfig{
		Collections: []config.MockCollection{
			{Name: "WeaveDocs"},
		},
		EmbeddingDimension: 384,
		SimulateEmbeddings: true,
	}
	return NewClient(cfg)
}

// Real-world scenarios
func TestMockClient_RealWorldWorkflow(t *testing.T) {
	t.Run("Full document lifecycle", func(t *testing.T) {
		client := createTestClient()
		ctx := context.Background()

		// 1. Create collection
		err := client.CreateCollection(ctx, "TestWorkflow", "text-embedding-ada-002", nil)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}

		// 2. Add documents
		doc := Document{
			ID:      "workflow-doc-1",
			Content: "Test document content",
			Metadata: map[string]interface{}{
				"source": "test",
				"date":   "2026-01-22",
			},
		}
		err = client.AddDocument(ctx, "TestWorkflow", doc)
		if err != nil {
			t.Fatalf("Failed to add document: %v", err)
		}

		// 3. Query documents
		results, err := client.Query(ctx, "TestWorkflow", "test document", weaviate.QueryOptions{TopK: 10})
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected query results")
		}

		// 4. Get statistics
		stats, err := client.GetCollectionStats(ctx, "TestWorkflow")
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}
		if stats["document_count"].(int) == 0 {
			t.Error("Expected non-zero document count")
		}

		// 5. Delete document
		err = client.DeleteDocument(ctx, "TestWorkflow", "workflow-doc-1")
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		// 6. Verify deletion
		count, _ := client.CountDocuments(ctx, "TestWorkflow")
		if count != 0 {
			t.Errorf("Expected 0 documents after deletion, got %d", count)
		}

		// 7. Delete collection
		err = client.DeleteCollection(ctx, "TestWorkflow")
		if err != nil {
			t.Fatalf("Failed to delete collection: %v", err)
		}
	})

	t.Run("Bulk operations", func(t *testing.T) {
		client := createTestClient()
		ctx := context.Background()

		// Add many documents
		for i := 0; i < 100; i++ {
			doc := Document{
				ID:      "bulk-" + string(rune('0'+i%10)),
				Content: "Bulk document content",
			}
			_ = client.AddDocument(ctx, "WeaveDocs", doc)
		}

		// Count them
		count, err := client.CountDocuments(ctx, "WeaveDocs")
		if err != nil {
			t.Fatalf("Failed to count: %v", err)
		}
		if count == 0 {
			t.Error("Expected documents from bulk add")
		}

		// Delete all
		err = client.DeleteAllDocuments(ctx, "WeaveDocs")
		if err != nil {
			t.Fatalf("Failed to delete all: %v", err)
		}

		// Verify empty
		count, _ = client.CountDocuments(ctx, "WeaveDocs")
		if count != 0 {
			t.Errorf("Expected 0 after bulk delete, got %d", count)
		}
	})
}

func TestMockClient_EdgeCases(t *testing.T) {
	t.Run("Empty query string", func(t *testing.T) {
		client := createTestClient()
		ctx := context.Background()

		results, err := client.Query(ctx, "WeaveDocs", "", weaviate.QueryOptions{})
		if err != nil {
			t.Errorf("Unexpected error for empty query: %v", err)
		}
		// Empty query with no words means no scoring, so results might be empty
		// Just verify no error occurred
		t.Logf("Empty query returned %d results", len(results))
	})

	t.Run("Document with nil metadata", func(t *testing.T) {
		client := createTestClient()
		ctx := context.Background()

		doc := Document{
			ID:       "nil-meta",
			Content:  "Content without metadata",
			Metadata: nil,
		}

		err := client.AddDocument(ctx, "WeaveDocs", doc)
		if err != nil {
			t.Errorf("Should handle nil metadata: %v", err)
		}

		// Query should still work
		retrieved, err := client.GetDocument(ctx, "WeaveDocs", "nil-meta")
		if err != nil {
			t.Errorf("Failed to retrieve doc with nil metadata: %v", err)
		}
		if retrieved.Metadata != nil {
			t.Error("Expected nil metadata to be preserved")
		}
	})

	t.Run("Very long content", func(t *testing.T) {
		client := createTestClient()
		ctx := context.Background()

		longContent := strings.Repeat("word ", 10000)
		doc := Document{
			ID:      "long-content",
			Content: longContent,
		}

		err := client.AddDocument(ctx, "WeaveDocs", doc)
		if err != nil {
			t.Errorf("Should handle long content: %v", err)
		}

		retrieved, _ := client.GetDocument(ctx, "WeaveDocs", "long-content")
		if len(retrieved.Content) != len(longContent) {
			t.Error("Long content should be preserved")
		}
	})
}
