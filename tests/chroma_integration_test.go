// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration && ((linux && amd64) || (darwin && amd64) || (darwin && arm64))
// +build integration
// +build linux,amd64 darwin,amd64 darwin,arm64

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)

func getChromaTestClient(t *testing.T) *chroma.Client {
	// Check for Chroma URL - try cloud first, then local
	chromaURL := os.Getenv("CHROMA_CLOUD_URL")
	if chromaURL == "" {
		chromaURL = os.Getenv("CHROMA_URL")
	}
	if chromaURL == "" {
		chromaURL = "http://localhost:8000"
	}

	// Get API key
	apiKey := os.Getenv("CHROMA_CLOUD_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CHROMA_API_KEY")
	}

	// Get tenant and database
	tenant := os.Getenv("CHROMA_TENANT")
	if tenant == "" {
		tenant = "default_tenant"
	}
	database := os.Getenv("CHROMA_DATABASE")
	if database == "" {
		database = "default_database"
	}

	config := &chroma.Config{
		URL:              chromaURL,
		APIKey:           apiKey,
		Tenant:           tenant,
		Database:         database,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}

	client, err := chroma.NewClient(config)
	if err != nil {
		t.Skipf("Failed to create Chroma client (is Chroma running?): %v", err)
	}

	return client
}

func TestChromaHealth(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	t.Log("Health check passed")
}

func TestChromaCollectionOperations(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Test collection name with timestamp to avoid conflicts
	collectionName := fmt.Sprintf("test_collection_%d", time.Now().UnixNano())

	// Clean up after test
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	// Test CreateCollection
	t.Run("CreateCollection", func(t *testing.T) {
		err := client.CreateCollection(ctx, collectionName, nil)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		t.Logf("Created collection: %s", collectionName)
	})

	// Test CollectionExists
	t.Run("CollectionExists", func(t *testing.T) {
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to check collection exists: %v", err)
		}
		if !exists {
			t.Fatal("Collection should exist")
		}
		t.Log("Collection exists check passed")
	})

	// Test ListCollections
	t.Run("ListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Fatalf("Failed to list collections: %v", err)
		}

		found := false
		for _, col := range collections {
			if col.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Collection %s not found in list", collectionName)
		}
		t.Logf("Found %d collections, including test collection", len(collections))
	})

	// Test GetCollectionCount
	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get collection count: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected count 0, got %d", count)
		}
		t.Log("Collection count check passed")
	})

	// Test DeleteCollection
	t.Run("DeleteCollection", func(t *testing.T) {
		err := client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to delete collection: %v", err)
		}

		// Verify deletion
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to check collection exists after delete: %v", err)
		}
		if exists {
			t.Fatal("Collection should not exist after deletion")
		}
		t.Log("Collection deleted successfully")
	})
}

func TestChromaCollectionNotFound(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Test getting count for non-existent collection
	_, err := client.GetCollectionCount(ctx, "non_existent_collection_xyz")
	if err == nil {
		t.Fatal("Expected error for non-existent collection")
	}
	t.Logf("Got expected error for non-existent collection: %v", err)
}

func TestChromaDocumentOperations(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Create a test collection
	collectionName := fmt.Sprintf("test_docs_%d", time.Now().UnixNano())

	// Clean up after test
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	// Create collection first
	err := client.CreateCollection(ctx, collectionName, nil)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Test CreateDocument
	t.Run("CreateDocument", func(t *testing.T) {
		doc := &vectordb.Document{
			ID:      "doc1",
			Content: "This is test document content for Chroma",
			URL:     "https://example.com/doc1",
			Metadata: map[string]interface{}{
				"author": "test",
				"type":   "article",
			},
		}

		err := client.CreateDocument(ctx, collectionName, doc)
		if err != nil {
			t.Fatalf("Failed to create document: %v", err)
		}
		t.Log("Document created successfully")
	})

	// Test GetDocument
	t.Run("GetDocument", func(t *testing.T) {
		doc, err := client.GetDocument(ctx, collectionName, "doc1")
		if err != nil {
			t.Fatalf("Failed to get document: %v", err)
		}

		if doc.ID != "doc1" {
			t.Fatalf("Expected ID 'doc1', got '%s'", doc.ID)
		}
		if doc.Content != "This is test document content for Chroma" {
			t.Fatalf("Content mismatch: got '%s'", doc.Content)
		}
		if doc.URL != "https://example.com/doc1" {
			t.Fatalf("URL mismatch: got '%s'", doc.URL)
		}
		t.Logf("Got document: ID=%s, Content=%s", doc.ID, doc.Content[:30]+"...")
	})

	// Test UpdateDocument
	t.Run("UpdateDocument", func(t *testing.T) {
		updatedDoc := &vectordb.Document{
			ID:      "doc1",
			Content: "Updated content for the document",
			URL:     "https://example.com/doc1-updated",
			Metadata: map[string]interface{}{
				"author": "test-updated",
				"type":   "article",
			},
		}

		err := client.UpdateDocument(ctx, collectionName, updatedDoc)
		if err != nil {
			t.Fatalf("Failed to update document: %v", err)
		}

		// Verify update
		doc, err := client.GetDocument(ctx, collectionName, "doc1")
		if err != nil {
			t.Fatalf("Failed to get updated document: %v", err)
		}
		if doc.Content != "Updated content for the document" {
			t.Fatalf("Update failed: got content '%s'", doc.Content)
		}
		t.Log("Document updated successfully")
	})

	// Test DeleteDocument
	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, collectionName, "doc1")
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, collectionName, "doc1")
		if err == nil {
			t.Fatal("Expected error getting deleted document")
		}
		t.Log("Document deleted successfully")
	})
}

func TestChromaDocumentNotFound(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Create a test collection
	collectionName := fmt.Sprintf("test_doc_notfound_%d", time.Now().UnixNano())
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	err := client.CreateCollection(ctx, collectionName, nil)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Test getting non-existent document
	_, err = client.GetDocument(ctx, collectionName, "non_existent_doc")
	if err == nil {
		t.Fatal("Expected error for non-existent document")
	}
	t.Logf("Got expected error for non-existent document: %v", err)
}

func TestChromaBatchOperations(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Create a test collection
	collectionName := fmt.Sprintf("test_batch_%d", time.Now().UnixNano())
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	err := client.CreateCollection(ctx, collectionName, nil)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Test CreateDocuments (batch)
	t.Run("CreateDocuments", func(t *testing.T) {
		docs := []*vectordb.Document{
			{ID: "batch1", Content: "First batch document", Metadata: map[string]interface{}{"batch": "1"}},
			{ID: "batch2", Content: "Second batch document", Metadata: map[string]interface{}{"batch": "1"}},
			{ID: "batch3", Content: "Third batch document", Metadata: map[string]interface{}{"batch": "2"}},
		}

		err := client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Fatalf("Failed to create documents in batch: %v", err)
		}
		t.Logf("Created %d documents in batch", len(docs))
	})

	// Test ListDocuments
	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 0, 0)
		if err != nil {
			t.Fatalf("Failed to list documents: %v", err)
		}
		if len(docs) != 3 {
			t.Fatalf("Expected 3 documents, got %d", len(docs))
		}
		t.Logf("Listed %d documents", len(docs))
	})

	// Test ListDocuments with limit
	t.Run("ListDocumentsWithLimit", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 2, 0)
		if err != nil {
			t.Fatalf("Failed to list documents with limit: %v", err)
		}
		if len(docs) != 2 {
			t.Fatalf("Expected 2 documents (limit), got %d", len(docs))
		}
		t.Logf("Listed %d documents with limit", len(docs))
	})

	// Test DeleteDocuments (batch)
	t.Run("DeleteDocuments", func(t *testing.T) {
		err := client.DeleteDocuments(ctx, collectionName, []string{"batch1", "batch2"})
		if err != nil {
			t.Fatalf("Failed to delete documents in batch: %v", err)
		}

		// Verify deletion
		docs, err := client.ListDocuments(ctx, collectionName, 0, 0)
		if err != nil {
			t.Fatalf("Failed to list documents after delete: %v", err)
		}
		if len(docs) != 1 {
			t.Fatalf("Expected 1 document after batch delete, got %d", len(docs))
		}
		t.Log("Batch delete successful")
	})
}

func TestChromaDeleteByMetadata(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Create a test collection
	collectionName := fmt.Sprintf("test_meta_delete_%d", time.Now().UnixNano())
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	err := client.CreateCollection(ctx, collectionName, nil)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Create documents with different metadata
	docs := []*vectordb.Document{
		{ID: "meta1", Content: "Document to delete", Metadata: map[string]interface{}{"category": "delete"}},
		{ID: "meta2", Content: "Document to keep", Metadata: map[string]interface{}{"category": "keep"}},
		{ID: "meta3", Content: "Another to delete", Metadata: map[string]interface{}{"category": "delete"}},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	if err != nil {
		t.Fatalf("Failed to create documents: %v", err)
	}

	// Delete by metadata
	err = client.DeleteDocumentsByMetadata(ctx, collectionName, map[string]interface{}{"category": "delete"})
	if err != nil {
		t.Fatalf("Failed to delete by metadata: %v", err)
	}

	// Verify only "keep" document remains
	remaining, err := client.ListDocuments(ctx, collectionName, 0, 0)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("Expected 1 document after metadata delete, got %d", len(remaining))
	}
	if remaining[0].ID != "meta2" {
		t.Fatalf("Expected 'meta2' to remain, got '%s'", remaining[0].ID)
	}
	t.Log("Delete by metadata successful")
}

func TestChromaFactoryIntegration(t *testing.T) {
	// Check for Chroma URL
	chromaURL := os.Getenv("CHROMA_CLOUD_URL")
	if chromaURL == "" {
		chromaURL = os.Getenv("CHROMA_URL")
	}
	if chromaURL == "" {
		t.Skip("Skipping Chroma factory test: no CHROMA_URL or CHROMA_CLOUD_URL set")
	}

	// Get API key
	apiKey := os.Getenv("CHROMA_CLOUD_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CHROMA_API_KEY")
	}

	// Get tenant and database
	tenant := os.Getenv("CHROMA_TENANT")
	if tenant == "" {
		tenant = "default_tenant"
	}
	database := os.Getenv("CHROMA_DATABASE")
	if database == "" {
		database = "default_database"
	}

	// Determine type
	dbType := vectordb.VectorDBTypeChromaLocal
	if apiKey != "" {
		dbType = vectordb.VectorDBTypeChromaCloud
	}

	config := &vectordb.Config{
		Type:             dbType,
		URL:              chromaURL,
		APIKey:           apiKey,
		Tenant:           tenant,
		Database:         database,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}

	client, err := vectordb.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create Chroma client via factory: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test health
	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	t.Log("Factory-created client health check passed")
}

func TestChromaVectorDBRegistry(t *testing.T) {
	// Test that Chroma types are registered
	supportedTypes := vectordb.GetSupportedTypes()

	chromaLocalFound := false
	chromaCloudFound := false

	for _, dbType := range supportedTypes {
		if dbType == vectordb.VectorDBTypeChromaLocal {
			chromaLocalFound = true
		}
		if dbType == vectordb.VectorDBTypeChromaCloud {
			chromaCloudFound = true
		}
	}

	if !chromaLocalFound {
		t.Error("chroma-local not found in supported types")
	}

	if !chromaCloudFound {
		t.Error("chroma-cloud not found in supported types")
	}

	// Test IsSupported
	if !vectordb.IsSupported(vectordb.VectorDBTypeChromaLocal) {
		t.Error("chroma-local should be supported")
	}

	if !vectordb.IsSupported(vectordb.VectorDBTypeChromaCloud) {
		t.Error("chroma-cloud should be supported")
	}

	t.Log("Chroma types correctly registered in vectordb registry")
}
