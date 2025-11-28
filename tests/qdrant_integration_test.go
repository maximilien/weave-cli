// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/qdrant"
)

func TestQdrantIntegration(t *testing.T) {
	// Check if Qdrant is available
	url := os.Getenv("QDRANT_URL")
	if url == "" {
		url = "localhost:6334"
	}

	apiKey := os.Getenv("QDRANT_API_KEY")

	// Create Qdrant client
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeQdrantLocal,
		URL:              url,
		APIKey:           apiKey,
		Timeout:          30,
		VectorDimensions: 1536,
		SimilarityMetric: "Cosine",
	}

	client, err := qdrant.NewAdapter(config)
	if err != nil {
		t.Skipf("Skipping Qdrant integration test: failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test health check
	t.Run("Health", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	// Test collection operations
	collectionName := "test_collection_qdrant"

	// Clean up if exists from previous run
	_ = client.DeleteCollection(ctx, collectionName)

	t.Run("CreateCollection", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text-embedding-3-small",
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Errorf("Failed to create collection: %v", err)
		}
	})

	t.Run("CollectionExists", func(t *testing.T) {
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to check collection existence: %v", err)
		}
		if !exists {
			t.Error("Collection should exist after creation")
		}
	})

	t.Run("ListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		found := false
		for _, collection := range collections {
			if collection.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Created collection '%s' should appear in list", collectionName)
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:   "test-doc-1",
		Text: "This is a test document for Qdrant integration",
		Metadata: map[string]interface{}{
			"filename": "test.txt",
			"category": "test",
			"author":   "integration-test",
		},
	}

	t.Run("CreateDocument", func(t *testing.T) {
		err := client.CreateDocument(ctx, collectionName, testDoc)
		if err != nil {
			t.Errorf("Failed to create document: %v", err)
		}
	})

	t.Run("GetDocument", func(t *testing.T) {
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to get document: %v", err)
			return
		}
		if doc == nil {
			t.Error("GetDocument returned nil document")
			return
		}
		if doc.ID != testDoc.ID {
			t.Errorf("Expected document ID %s, got %s", testDoc.ID, doc.ID)
		}
		if doc.Text != testDoc.Text {
			t.Errorf("Expected text %s, got %s", testDoc.Text, doc.Text)
		}
		// Verify metadata is preserved
		if filename, ok := doc.Metadata["filename"].(string); !ok || filename != "test.txt" {
			t.Errorf("Expected metadata filename 'test.txt', got %v", doc.Metadata["filename"])
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		updatedDoc := &vectordb.Document{
			ID:   testDoc.ID,
			Text: "Updated test document content for Qdrant",
			Metadata: map[string]interface{}{
				"filename": "updated.txt",
				"category": "updated",
			},
		}

		err := client.UpdateDocument(ctx, collectionName, updatedDoc)
		if err != nil {
			t.Errorf("Failed to update document: %v", err)
		}

		// Verify update
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to get updated document: %v", err)
			return
		}
		if doc.Text != updatedDoc.Text {
			t.Errorf("Expected updated text %s, got %s", updatedDoc.Text, doc.Text)
		}
	})

	// Test batch operations
	t.Run("BatchCreateDocuments", func(t *testing.T) {
		batchDocs := []*vectordb.Document{
			{
				ID:   "batch-doc-1",
				Text: "First batch document",
				Metadata: map[string]interface{}{
					"batch": "1",
					"index": int64(1),
				},
			},
			{
				ID:   "batch-doc-2",
				Text: "Second batch document",
				Metadata: map[string]interface{}{
					"batch": "1",
					"index": int64(2),
				},
			},
			{
				ID:   "batch-doc-3",
				Text: "Third batch document",
				Metadata: map[string]interface{}{
					"batch": "1",
					"index": int64(3),
				},
			},
		}

		err := client.CreateDocuments(ctx, collectionName, batchDocs)
		if err != nil {
			t.Errorf("Failed to create batch documents: %v", err)
		}

		// Verify all documents were created
		for _, doc := range batchDocs {
			retrieved, err := client.GetDocument(ctx, collectionName, doc.ID)
			if err != nil {
				t.Errorf("Failed to get batch document %s: %v", doc.ID, err)
			}
			if retrieved == nil || retrieved.ID != doc.ID {
				t.Errorf("Batch document %s not found or has incorrect ID", doc.ID)
			}
		}
	})

	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get collection count: %v", err)
		}
		// We should have at least 4 documents (1 original + 3 batch)
		if count < 4 {
			t.Errorf("Expected at least 4 documents, got %d", count)
		}
	})

	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}
		if len(docs) < 4 {
			t.Errorf("Expected at least 4 documents in list, got %d", len(docs))
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, collectionName, "batch-doc-1")
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, collectionName, "batch-doc-1")
		if err == nil {
			t.Error("Document should not exist after deletion")
		}
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		err := client.DeleteDocuments(ctx, collectionName, []string{"batch-doc-2", "batch-doc-3"})
		if err != nil {
			t.Errorf("Failed to delete batch documents: %v", err)
		}

		// Verify deletions
		_, err = client.GetDocument(ctx, collectionName, "batch-doc-2")
		if err == nil {
			t.Error("batch-doc-2 should not exist after deletion")
		}

		_, err = client.GetDocument(ctx, collectionName, "batch-doc-3")
		if err == nil {
			t.Error("batch-doc-3 should not exist after deletion")
		}
	})

	// Clean up
	t.Run("DeleteCollection", func(t *testing.T) {
		err := client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to delete collection: %v", err)
		}

		// Verify deletion
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to check collection existence after deletion: %v", err)
		}
		if exists {
			t.Error("Collection should not exist after deletion")
		}
	})
}
