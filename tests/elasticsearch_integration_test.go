// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/elasticsearch"
)

func TestElasticsearchIntegration(t *testing.T) {
	// Check if Elasticsearch is available
	address := os.Getenv("ELASTICSEARCH_LOCAL_ADDRESS")
	if address == "" {
		address = "http://localhost:9200"
	}

	// Check if Elasticsearch is reachable
	// If not set or not reachable, skip the test
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeElasticsearchLocal,
		Address:          address,
		Timeout:          30,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
	}

	factory := elasticsearch.NewFactory()
	client, err := factory.CreateClient(config)
	if err != nil {
		t.Skipf("Skipping Elasticsearch integration test: failed to create client: %v", err)
	}

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
	collectionName := "test_collection_elasticsearch"

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
		for _, col := range collections {
			if col.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Collection %s not found in list", collectionName)
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:      "test-doc-1",
		Content: "This is a test document for Elasticsearch integration testing.",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": 1,
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
		}
		if doc == nil {
			t.Error("Document should not be nil")
			return
		}
		if doc.ID != testDoc.ID {
			t.Errorf("Expected document ID %s, got %s", testDoc.ID, doc.ID)
		}
		if doc.Content != testDoc.Content {
			t.Errorf("Expected content %s, got %s", testDoc.Content, doc.Content)
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		updatedDoc := &vectordb.Document{
			ID:      testDoc.ID,
			Content: "This is an updated test document.",
			Metadata: map[string]interface{}{
				"category": "test",
				"priority": 2,
				"updated":  true,
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
		}
		if doc.Content != updatedDoc.Content {
			t.Errorf("Document content not updated: expected %s, got %s", updatedDoc.Content, doc.Content)
		}
	})

	t.Run("CreateDocuments", func(t *testing.T) {
		docs := []*vectordb.Document{
			{
				ID:      "test-doc-2",
				Content: "Second test document",
				Metadata: map[string]interface{}{
					"category": "test",
					"batch":    true,
				},
			},
			{
				ID:      "test-doc-3",
				Content: "Third test document",
				Metadata: map[string]interface{}{
					"category": "test",
					"batch":    true,
				},
			},
		}

		err := client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Errorf("Failed to create documents in batch: %v", err)
		}
	})

	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}
		if len(docs) < 3 {
			t.Errorf("Expected at least 3 documents, got %d", len(docs))
		}
	})

	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get collection count: %v", err)
		}
		if count < 3 {
			t.Errorf("Expected at least 3 documents, got %d", count)
		}
	})

	t.Run("SearchSemantic", func(t *testing.T) {
		options := &vectordb.QueryOptions{
			TopK: 5,
		}
		results, err := client.SearchSemantic(ctx, collectionName, "test document", options)
		if err != nil {
			t.Errorf("Failed to search by content: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected search results, got none")
		}
	})

	t.Run("SearchByMetadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"category": "test",
		}
		options := &vectordb.QueryOptions{
			TopK: 10,
		}
		results, err := client.SearchByMetadata(ctx, collectionName, metadata, options)
		if err != nil {
			t.Errorf("Failed to search by metadata: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected search results, got none")
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, collectionName, "test-doc-2")
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, collectionName, "test-doc-2")
		if err == nil {
			t.Error("Document should not exist after deletion")
		}
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		err := client.DeleteDocuments(ctx, collectionName, []string{"test-doc-3"})
		if err != nil {
			t.Errorf("Failed to delete documents: %v", err)
		}
	})

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
