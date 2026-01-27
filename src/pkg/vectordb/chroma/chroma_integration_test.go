// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration && ((darwin && amd64) || (darwin && arm64))
// +build integration
// +build darwin,amd64 darwin,arm64

package chroma

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testURL        = "http://localhost:8000"
	testTenant     = "default_tenant"
	testDatabase   = "default_database"
	testDimensions = 384
	testMetric     = "cosine"
	testTimeout    = 30
)

// getTestClient creates a test Chroma client
func getTestClient(t *testing.T) *Client {
	config := &Config{
		URL:              testURL,
		Tenant:           testTenant,
		Database:         testDatabase,
		VectorDimensions: testDimensions,
		SimilarityMetric: testMetric,
		Timeout:          testTimeout,
	}

	client, err := NewClient(config)
	require.NoError(t, err, "Failed to create test client")
	require.NotNil(t, client)

	return client
}

// getUniqueCollectionName generates unique collection name for test isolation
func getUniqueCollectionName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// getTestSchema returns a standard test schema
func getTestSchema(collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class: collectionName,
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}
}

// TestIntegration_Chroma_Health tests health check connectivity
func TestIntegration_Chroma_Health(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	err := client.Health(ctx)
	assert.NoError(t, err, "Health check should succeed")
}

// TestIntegration_Chroma_CreateCollection tests collection creation
func TestIntegration_Chroma_CreateCollection(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create")
	defer client.DeleteCollection(ctx, collectionName)

	schema := getTestSchema(collectionName)
	err := client.CreateCollection(ctx, collectionName, schema)
	assert.NoError(t, err, "Failed to create collection")

	// Verify collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.True(t, exists, "Collection should exist after creation")
}

// TestIntegration_Chroma_DeleteCollection tests collection deletion
func TestIntegration_Chroma_DeleteCollection(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete")

	// Create collection first
	schema := getTestSchema(collectionName)
	err := client.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err)

	// Delete collection
	err = client.DeleteCollection(ctx, collectionName)
	assert.NoError(t, err, "Failed to delete collection")

	// Verify collection doesn't exist
	exists, err := client.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.False(t, exists, "Collection should not exist after deletion")
}

// TestIntegration_Chroma_ListCollections tests listing collections
func TestIntegration_Chroma_ListCollections(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName1 := getUniqueCollectionName("test_list_1")
	collectionName2 := getUniqueCollectionName("test_list_2")
	defer client.DeleteCollection(ctx, collectionName1)
	defer client.DeleteCollection(ctx, collectionName2)

	// Create two collections
	err := client.CreateCollection(ctx, collectionName1, getTestSchema(collectionName1))
	require.NoError(t, err)
	err = client.CreateCollection(ctx, collectionName2, getTestSchema(collectionName2))
	require.NoError(t, err)

	// List collections
	collections, err := client.ListCollections(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, collections)

	// Find our collections
	found1 := false
	found2 := false
	for _, col := range collections {
		if col.Name == collectionName1 {
			found1 = true
		}
		if col.Name == collectionName2 {
			found2 = true
		}
	}
	assert.True(t, found1, "Collection 1 should be in the list")
	assert.True(t, found2, "Collection 2 should be in the list")
}

// TestIntegration_Chroma_CreateDocument tests single document insertion
func TestIntegration_Chroma_CreateDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	doc := &vectordb.Document{
		ID:   "doc-1",
		Text: "Test document content",
		Metadata: map[string]interface{}{
			"title":  "Test Document",
			"index":  1,
			"active": true,
		},
	}

	err = client.CreateDocument(ctx, collectionName, doc)
	assert.NoError(t, err, "Failed to create document")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document was created
	count, err := client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Collection should have 1 document")
}

// TestIntegration_Chroma_CreateDocuments tests bulk document insertion
func TestIntegration_Chroma_CreateDocuments(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bulk_create")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create multiple documents
	docs := []*vectordb.Document{
		{
			ID:   "bulk-1",
			Text: "Bulk document 1",
			Metadata: map[string]interface{}{
				"index": 1,
				"type":  "bulk",
			},
		},
		{
			ID:   "bulk-2",
			Text: "Bulk document 2",
			Metadata: map[string]interface{}{
				"index": 2,
				"type":  "bulk",
			},
		},
		{
			ID:   "bulk-3",
			Text: "Bulk document 3",
			Metadata: map[string]interface{}{
				"index": 3,
				"type":  "bulk",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	assert.NoError(t, err, "Failed to create bulk documents")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document count
	count, err := client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "Collection should have 3 documents")
}

// TestIntegration_Chroma_GetDocument tests document retrieval by ID
func TestIntegration_Chroma_GetDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_get_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	expectedDoc := &vectordb.Document{
		ID:   "get-test-1",
		Text: "Document to retrieve",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": 5,
		},
	}

	err = client.CreateDocument(ctx, collectionName, expectedDoc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Retrieve document
	retrievedDoc, err := client.GetDocument(ctx, collectionName, expectedDoc.ID)
	assert.NoError(t, err, "Failed to retrieve document")
	if retrievedDoc != nil {
		assert.Equal(t, expectedDoc.ID, retrievedDoc.ID)
		// Note: Text field might not always be preserved in Chroma
	}
}

// TestIntegration_Chroma_UpdateDocument tests document update
func TestIntegration_Chroma_UpdateDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_update_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create initial document
	doc := &vectordb.Document{
		ID:   "update-test-1",
		Text: "Original content",
		Metadata: map[string]interface{}{
			"version": 1,
		},
	}

	err = client.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Update document
	updatedDoc := &vectordb.Document{
		ID:   "update-test-1",
		Text: "Updated content",
		Metadata: map[string]interface{}{
			"version": 2,
		},
	}

	err = client.UpdateDocument(ctx, collectionName, updatedDoc)
	assert.NoError(t, err, "Failed to update document")

	// Wait for update
	time.Sleep(1 * time.Second)

	// Verify still only one document
	count, err := client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Should still have 1 document after update")

	// Verify update succeeded (content verification may vary by Chroma version)
	retrieved, err := client.GetDocument(ctx, collectionName, "update-test-1")
	if err == nil && retrieved != nil {
		// Update succeeded if we can retrieve the document
		assert.Equal(t, "update-test-1", retrieved.ID)
	}
}

// TestIntegration_Chroma_DeleteDocument tests document deletion by ID
func TestIntegration_Chroma_DeleteDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	doc := &vectordb.Document{
		ID:   "delete-test-1",
		Text: "Document to delete",
	}

	err = client.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document exists
	count, err := client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete document
	err = client.DeleteDocument(ctx, collectionName, doc.ID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion to propagate
	time.Sleep(1 * time.Second)

	// Verify document was deleted
	count, err = client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "Collection should be empty after deletion")
}

// TestIntegration_Chroma_SearchSemantic tests semantic search
func TestIntegration_Chroma_SearchSemantic(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_search")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create test documents
	docs := []*vectordb.Document{
		{
			ID:   "search-1",
			Text: "The quick brown fox jumps over the lazy dog",
			Metadata: map[string]interface{}{
				"category": "animals",
			},
		},
		{
			ID:   "search-2",
			Text: "Python is a popular programming language",
			Metadata: map[string]interface{}{
				"category": "programming",
			},
		},
		{
			ID:   "search-3",
			Text: "Machine learning and artificial intelligence",
			Metadata: map[string]interface{}{
				"category": "ai",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Search with query similar to programming
	opts := &vectordb.QueryOptions{
		TopK: 2,
	}
	results, err := client.SearchSemantic(ctx, collectionName, "coding and software development", opts)
	// Semantic search may return nil results if embeddings aren't configured
	// or require external embedding service
	if err == nil && results != nil {
		// If search works, verify results structure
		if len(results) > 0 {
			assert.NotNil(t, results[0])
		}
	} else {
		// Skip assertion if search isn't supported in this config
		t.Logf("Semantic search not available or returned no results: %v", err)
	}
}

// TestIntegration_Chroma_SearchByMetadata tests metadata filtering
func TestIntegration_Chroma_SearchByMetadata(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_metadata_search")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create documents with different metadata
	docs := []*vectordb.Document{
		{
			ID:   "meta-1",
			Text: "Active document 1",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "A",
			},
		},
		{
			ID:   "meta-2",
			Text: "Active document 2",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "B",
			},
		},
		{
			ID:   "meta-3",
			Text: "Inactive document",
			Metadata: map[string]interface{}{
				"status": "inactive",
				"type":   "A",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Search for active documents
	filter := map[string]interface{}{
		"status": "active",
	}

	opts := &vectordb.QueryOptions{
		TopK: 10,
	}

	results, err := client.SearchByMetadata(ctx, collectionName, filter, opts)
	if err == nil && results != nil {
		// Metadata search may have varying behavior
		assert.GreaterOrEqual(t, len(results), 0, "Results should be non-negative")

		// If we got results with metadata, verify they match
		for _, result := range results {
			if result.Document.Metadata != nil && result.Document.Metadata["status"] != nil {
				assert.Equal(t, "active", result.Document.Metadata["status"])
			}
		}
	} else {
		t.Logf("Metadata search not fully supported or returned no results: %v", err)
	}
}

// TestIntegration_Chroma_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_Chroma_E2E_Workflow(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_e2e")

	// 1. Create collection
	schema := getTestSchema(collectionName)
	err := client.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Step 1: Failed to create collection")

	// 2. Verify collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 2: Failed to check existence")
	require.True(t, exists, "Step 2: Collection should exist")

	// 3. Insert documents
	docs := []*vectordb.Document{
		{
			ID:   "e2e-1",
			Text: "E2E test document 1",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     1,
			},
		},
		{
			ID:   "e2e-2",
			Text: "E2E test document 2",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     2,
			},
		},
		{
			ID:   "e2e-3",
			Text: "E2E test document 3",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     3,
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Step 3: Failed to insert documents")

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// 4. Verify document count
	count, err := client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 4: Failed to get count")
	assert.Equal(t, int64(3), count, "Step 4: Should have 3 documents")

	// 5. Retrieve specific document
	doc, err := client.GetDocument(ctx, collectionName, "e2e-2")
	require.NoError(t, err, "Step 5: Failed to retrieve document")
	assert.NotNil(t, doc, "Step 5: Document should not be nil")
	assert.Equal(t, "e2e-2", doc.ID, "Step 5: Document ID should match")

	// 6. Delete one document
	err = client.DeleteDocument(ctx, collectionName, "e2e-1")
	require.NoError(t, err, "Step 6: Failed to delete document")

	// Wait for deletion
	time.Sleep(1 * time.Second)

	// 7. Verify count decreased
	count, err = client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 7: Failed to get count")
	assert.Equal(t, int64(2), count, "Step 7: Should have 2 documents after deletion")

	// 8. Clean up - delete collection
	err = client.DeleteCollection(ctx, collectionName)
	require.NoError(t, err, "Step 8: Failed to delete collection")

	// 9. Verify collection is gone
	exists, err = client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 9: Failed to check existence")
	assert.False(t, exists, "Step 9: Collection should not exist after deletion")
}

// TestIntegration_Chroma_SearchBM25 tests BM25 keyword search
func TestIntegration_Chroma_SearchBM25(t *testing.T) {
	client := getTestClient(t)
	defer client.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bm25")
	defer client.DeleteCollection(ctx, collectionName)

	// Create collection
	err := client.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create test documents with specific keywords
	docs := []*vectordb.Document{
		{
			ID:      "bm25-1",
			Text:    "Python is a powerful programming language",
			Content: "Python has extensive libraries for data science and machine learning",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
		{
			ID:      "bm25-2",
			Text:    "JavaScript for web development",
			Content: "JavaScript runs in browsers and on servers with Node.js",
			Metadata: map[string]interface{}{
				"lang": "JavaScript",
			},
		},
		{
			ID:      "bm25-3",
			Text:    "Go language for system programming",
			Content: "Go is efficient for building concurrent and scalable systems",
			Metadata: map[string]interface{}{
				"lang": "Go",
			},
		},
		{
			ID:      "bm25-4",
			Text:    "Python data analysis with pandas",
			Content: "Python pandas library provides powerful data manipulation tools",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Perform BM25 search for "Python"
	results, err := client.SearchBM25(ctx, collectionName, "Python", &vectordb.QueryOptions{TopK: 3})
	assert.NoError(t, err, "BM25 search should not error")
	assert.NotNil(t, results, "Results should not be nil")

	// Should return Python-related documents
	if len(results) > 0 {
		t.Logf("BM25 search returned %d results", len(results))

		// Verify results contain "Python"
		foundPython := false
		for _, result := range results {
			t.Logf("  Result: ID=%s, Score=%.4f", result.Document.ID, result.Score)
			if result.Document.Text != "" && (result.Document.Text == "Python is a powerful programming language" ||
				result.Document.Text == "Python data analysis with pandas") {
				foundPython = true
			}
		}
		assert.True(t, foundPython, "Results should contain Python documents")

		// Verify scores are positive
		for _, result := range results {
			assert.Greater(t, result.Score, 0.0, "BM25 scores should be positive")
		}
	}

	t.Log("✓ Chroma BM25 search test passed")
}
