// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/pinecone"
)

// TestPineconeIntegration runs integration tests with Pinecone Cloud
func TestPineconeIntegration(t *testing.T) {
	// Skip if no Pinecone configuration
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("PINECONE_CLOUD_API_KEY")
	}
	if apiKey == "" {
		t.Skip("Skipping Pinecone integration tests - missing PINECONE_API_KEY or PINECONE_CLOUD_API_KEY")
	}

	// Require OpenAI API key for embeddings
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping Pinecone integration tests - missing OPENAI_API_KEY (required for embeddings)")
	}

	// Create adapter
	adapter, err := pinecone.NewAdapter(&vectordb.Config{
		Type:   vectordb.VectorDBTypePinecone,
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Pinecone adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Generate unique collection name
	collectionName := fmt.Sprintf("weave_test_%d", time.Now().UnixNano())

	t.Run("Health", func(t *testing.T) {
		err := adapter.Health(ctx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		} else {
			t.Logf("✓ Health check passed")
		}
	})

	t.Run("ListCollections", func(t *testing.T) {
		collections, err := adapter.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
			return
		}
		t.Logf("✓ Found %d existing indexes", len(collections))
	})

	t.Run("CreateCollection", func(t *testing.T) {
		schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
		err := adapter.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Errorf("Failed to create collection: %v", err)
			return
		}
		t.Logf("✓ Created index: %s", collectionName)

		// Wait for index to be ready (Pinecone serverless indexes take time to initialize)
		time.Sleep(60 * time.Second)
	})

	t.Run("CollectionExists", func(t *testing.T) {
		exists, err := adapter.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to check collection existence: %v", err)
			return
		}
		if !exists {
			t.Errorf("Collection should exist but doesn't")
			return
		}
		t.Logf("✓ Collection exists check passed")
	})

	t.Run("CreateDocument", func(t *testing.T) {
		doc := &vectordb.Document{
			ID:      "test-doc-1",
			Content: "This is a test document about artificial intelligence and machine learning.",
			Metadata: map[string]interface{}{
				"category": "technology",
				"author":   "test-user",
			},
		}

		err := adapter.CreateDocument(ctx, collectionName, doc)
		if err != nil {
			t.Errorf("Failed to create document: %v", err)
			return
		}
		t.Logf("✓ Document created successfully")

		// Wait for vector to be indexed
		time.Sleep(5 * time.Second)
	})

	t.Run("CreateDocuments", func(t *testing.T) {
		docs := []*vectordb.Document{
			{
				ID:      "test-doc-2",
				Content: "Deep learning neural networks are transforming computer vision.",
				Metadata: map[string]interface{}{
					"category": "ai",
					"author":   "test-user",
				},
			},
			{
				ID:      "test-doc-3",
				Content: "Natural language processing enables machines to understand human language.",
				Metadata: map[string]interface{}{
					"category": "nlp",
					"author":   "test-user",
				},
			},
		}

		err := adapter.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Errorf("Failed to create documents: %v", err)
			return
		}
		t.Logf("✓ Created %d documents successfully", len(docs))

		// Wait for vectors to be indexed
		time.Sleep(5 * time.Second)
	})

	t.Run("GetDocument", func(t *testing.T) {
		doc, err := adapter.GetDocument(ctx, collectionName, "test-doc-1")
		if err != nil {
			t.Errorf("Failed to get document: %v", err)
			return
		}
		if doc.ID != "test-doc-1" {
			t.Errorf("Expected ID test-doc-1, got %s", doc.ID)
			return
		}
		if doc.Content == "" {
			t.Errorf("Document content is empty")
			return
		}
		t.Logf("✓ Retrieved document: ID=%s, Content=%s...", doc.ID, doc.Content[:50])
	})

	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := adapter.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
			return
		}
		if len(docs) < 3 {
			t.Errorf("Expected at least 3 documents, got %d", len(docs))
			return
		}
		t.Logf("✓ Listed %d documents", len(docs))
	})

	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := adapter.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get collection count: %v", err)
			return
		}
		if count < 3 {
			t.Errorf("Expected at least 3 documents, got %d", count)
			return
		}
		t.Logf("✓ Collection has %d documents", count)
	})

	t.Run("SearchSemantic", func(t *testing.T) {
		results, err := adapter.SearchSemantic(ctx, collectionName, "machine learning", &vectordb.QueryOptions{
			TopK: 5,
		})
		if err != nil {
			t.Errorf("Failed to search: %v", err)
			return
		}
		if len(results) == 0 {
			t.Errorf("Expected search results, got none")
			return
		}
		t.Logf("✓ Semantic search returned %d results", len(results))
		for i, result := range results {
			t.Logf("  Result %d: ID=%s, Score=%.3f, Content=%s...",
				i+1, result.Document.ID, result.Score, result.Document.Content[:50])
		}
	})

	t.Run("SearchByMetadata", func(t *testing.T) {
		results, err := adapter.SearchByMetadata(ctx, collectionName, map[string]interface{}{
			"category": "technology",
		}, &vectordb.QueryOptions{
			TopK: 10,
		})
		if err != nil {
			t.Errorf("Failed to search by metadata: %v", err)
			return
		}
		if len(results) == 0 {
			t.Errorf("Expected search results, got none")
			return
		}
		t.Logf("✓ Metadata search returned %d results", len(results))
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		doc := &vectordb.Document{
			ID:      "test-doc-1",
			Content: "Updated content about AI and ML technologies.",
			Metadata: map[string]interface{}{
				"category": "technology",
				"author":   "updated-user",
				"updated":  true,
			},
		}

		err := adapter.UpdateDocument(ctx, collectionName, doc)
		if err != nil {
			t.Errorf("Failed to update document: %v", err)
			return
		}
		t.Logf("✓ Document updated successfully")

		// Wait for update to be indexed
		time.Sleep(5 * time.Second)

		// Verify update
		updated, err := adapter.GetDocument(ctx, collectionName, "test-doc-1")
		if err != nil {
			t.Errorf("Failed to get updated document: %v", err)
			return
		}
		if updated.Content != doc.Content {
			t.Errorf("Content not updated correctly")
		}
		t.Logf("✓ Update verified")
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := adapter.DeleteDocument(ctx, collectionName, "test-doc-3")
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
			return
		}
		t.Logf("✓ Document deleted successfully")

		// Wait for deletion to be processed
		time.Sleep(5 * time.Second)
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		err := adapter.DeleteDocuments(ctx, collectionName, []string{"test-doc-2"})
		if err != nil {
			t.Errorf("Failed to delete documents: %v", err)
			return
		}
		t.Logf("✓ Documents deleted successfully")

		// Wait for deletion to be processed
		time.Sleep(5 * time.Second)
	})

	t.Run("DeleteCollection", func(t *testing.T) {
		err := adapter.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to delete collection: %v", err)
			return
		}
		t.Logf("✓ Collection deleted successfully")
	})
}

// TestPineconeErrorHandling tests error scenarios
func TestPineconeErrorHandling(t *testing.T) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("PINECONE_CLOUD_API_KEY")
	}
	if apiKey == "" {
		t.Skip("Skipping Pinecone error handling tests - missing PINECONE_API_KEY")
	}

	adapter, err := pinecone.NewAdapter(&vectordb.Config{
		Type:   vectordb.VectorDBTypePinecone,
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create Pinecone adapter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("NonExistentCollection", func(t *testing.T) {
		_, err := adapter.GetDocument(ctx, "non_existent_index_xyz", "doc-1")
		if err == nil {
			t.Error("Expected error for non-existent collection, got nil")
		} else {
			t.Logf("✓ Correctly handled non-existent collection: %v", err)
		}
	})

	t.Run("NonExistentDocument", func(t *testing.T) {
		// Try to get a document from a collection that doesn't exist
		_, err := adapter.GetDocument(ctx, "test-collection", "non-existent-doc-id")
		if err == nil {
			t.Error("Expected error for non-existent document, got nil")
		} else {
			t.Logf("✓ Correctly handled non-existent document: %v", err)
		}
	})
}
