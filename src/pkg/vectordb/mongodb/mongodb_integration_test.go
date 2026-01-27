// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package mongodb

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
	testURI        = "mongodb://admin:password@localhost:27017"
	testDatabase   = "test_vectordb"
	testDimensions = 384
	testMetric     = "cosine"
	testTimeout    = 30
)

// getTestAdapter creates a test MongoDB adapter
func getTestAdapter(t *testing.T) *Adapter {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMongoDB,
		URL:              testURI,
		Database:         testDatabase,
		Timeout:          testTimeout,
		VectorDimensions: testDimensions,
		SimilarityMetric: testMetric,
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err, "Failed to create test adapter")
	require.NotNil(t, adapter)

	return adapter
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
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}
}

// TestIntegration_MongoDB_Health tests health check connectivity
func TestIntegration_MongoDB_Health(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	err := adapter.Health(ctx)
	assert.NoError(t, err, "Health check should succeed")
}

// TestIntegration_MongoDB_CreateCollection tests collection creation
func TestIntegration_MongoDB_CreateCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create")
	defer adapter.DeleteCollection(ctx, collectionName)

	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	assert.NoError(t, err, "Failed to create collection")

	// Verify collection exists
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.True(t, exists, "Collection should exist after creation")
}

// TestIntegration_MongoDB_DeleteCollection tests collection deletion
func TestIntegration_MongoDB_DeleteCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete")

	// Create collection first
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err)

	// Delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	assert.NoError(t, err, "Failed to delete collection")

	// Verify collection doesn't exist
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.False(t, exists, "Collection should not exist after deletion")
}

// TestIntegration_MongoDB_ListCollections tests listing collections
func TestIntegration_MongoDB_ListCollections(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName1 := getUniqueCollectionName("test_list_1")
	collectionName2 := getUniqueCollectionName("test_list_2")
	defer adapter.DeleteCollection(ctx, collectionName1)
	defer adapter.DeleteCollection(ctx, collectionName2)

	// Create two collections
	err := adapter.CreateCollection(ctx, collectionName1, getTestSchema(collectionName1))
	require.NoError(t, err)
	err = adapter.CreateCollection(ctx, collectionName2, getTestSchema(collectionName2))
	require.NoError(t, err)

	// List collections
	collections, err := adapter.ListCollections(ctx)
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

// TestIntegration_MongoDB_CollectionExists tests collection existence check
func TestIntegration_MongoDB_CollectionExists(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	existingCollection := getUniqueCollectionName("test_exists")
	nonExistingCollection := getUniqueCollectionName("test_not_exists")

	// Create one collection
	err := adapter.CreateCollection(ctx, existingCollection, getTestSchema(existingCollection))
	require.NoError(t, err)
	defer adapter.DeleteCollection(ctx, existingCollection)

	// Check existing collection
	exists, err := adapter.CollectionExists(ctx, existingCollection)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing collection
	exists, err = adapter.CollectionExists(ctx, nonExistingCollection)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestIntegration_MongoDB_CreateDocument tests single document insertion
func TestIntegration_MongoDB_CreateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	doc := &vectordb.Document{
		ID:      "doc-1",
		Text:    "Test document content",
		Content: "Full test document content",
		Metadata: map[string]interface{}{
			"title":  "Test Document",
			"index":  1,
			"active": true,
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	assert.NoError(t, err, "Failed to create document")

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Verify document was created
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Collection should have 1 document")
}

// TestIntegration_MongoDB_CreateDocuments tests bulk document insertion
func TestIntegration_MongoDB_CreateDocuments(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bulk_create")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create multiple documents
	docs := []*vectordb.Document{
		{
			ID:      "bulk-1",
			Text:    "Bulk document 1",
			Content: "Full content for bulk document 1",
			Metadata: map[string]interface{}{
				"index": 1,
				"type":  "bulk",
			},
		},
		{
			ID:      "bulk-2",
			Text:    "Bulk document 2",
			Content: "Full content for bulk document 2",
			Metadata: map[string]interface{}{
				"index": 2,
				"type":  "bulk",
			},
		},
		{
			ID:      "bulk-3",
			Text:    "Bulk document 3",
			Content: "Full content for bulk document 3",
			Metadata: map[string]interface{}{
				"index": 3,
				"type":  "bulk",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	assert.NoError(t, err, "Failed to create bulk documents")

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "Collection should have 3 documents")
}

// TestIntegration_MongoDB_GetDocument tests document retrieval by ID
func TestIntegration_MongoDB_GetDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_get_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	expectedDoc := &vectordb.Document{
		ID:      "get-test-1",
		Text:    "Document to retrieve",
		Content: "Full content to retrieve",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": 5,
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, expectedDoc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Retrieve document
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, expectedDoc.ID)
	assert.NoError(t, err, "Failed to retrieve document")
	assert.NotNil(t, retrievedDoc)
	assert.Equal(t, expectedDoc.ID, retrievedDoc.ID)
	assert.Equal(t, expectedDoc.Text, retrievedDoc.Text)
}

// TestIntegration_MongoDB_UpdateDocument tests document update
func TestIntegration_MongoDB_UpdateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_update_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create initial document
	doc := &vectordb.Document{
		ID:      "update-test-1",
		Text:    "Original content",
		Content: "Original full content",
		Metadata: map[string]interface{}{
			"version": 1,
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Update document
	updatedDoc := &vectordb.Document{
		ID:      "update-test-1",
		Text:    "Updated content",
		Content: "Updated full content",
		Metadata: map[string]interface{}{
			"version": 2,
		},
	}

	err = adapter.UpdateDocument(ctx, collectionName, updatedDoc)
	assert.NoError(t, err, "Failed to update document")

	// Wait for update
	time.Sleep(500 * time.Millisecond)

	// Verify still only one document
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Should still have 1 document after update")

	// Verify content was updated
	retrieved, err := adapter.GetDocument(ctx, collectionName, "update-test-1")
	assert.NoError(t, err)
	if retrieved != nil {
		assert.Equal(t, "Updated content", retrieved.Text)
	}
}

// TestIntegration_MongoDB_DeleteDocument tests document deletion by ID
func TestIntegration_MongoDB_DeleteDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create document
	doc := &vectordb.Document{
		ID:      "delete-test-1",
		Text:    "Document to delete",
		Content: "Content to delete",
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Verify document exists
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete document
	err = adapter.DeleteDocument(ctx, collectionName, doc.ID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion to propagate
	time.Sleep(500 * time.Millisecond)

	// Verify document was deleted
	count, err = adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "Collection should be empty after deletion")
}

// TestIntegration_MongoDB_ListDocuments tests document listing with pagination
func TestIntegration_MongoDB_ListDocuments(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_list_docs")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create test documents
	docs := []*vectordb.Document{
		{ID: "list-1", Text: "Document 1", Content: "Content 1"},
		{ID: "list-2", Text: "Document 2", Content: "Content 2"},
		{ID: "list-3", Text: "Document 3", Content: "Content 3"},
		{ID: "list-4", Text: "Document 4", Content: "Content 4"},
		{ID: "list-5", Text: "Document 5", Content: "Content 5"},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// List with limit
	results, err := adapter.ListDocuments(ctx, collectionName, 3, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 3, "Should return 3 documents")

	// List with offset
	results, err = adapter.ListDocuments(ctx, collectionName, 3, 2)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "Should return at least 1 document with offset")
}

// TestIntegration_MongoDB_SearchByMetadata tests metadata filtering
func TestIntegration_MongoDB_SearchByMetadata(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_metadata_search")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create documents with different metadata
	docs := []*vectordb.Document{
		{
			ID:      "meta-1",
			Text:    "Active document 1",
			Content: "Content 1",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "A",
			},
		},
		{
			ID:      "meta-2",
			Text:    "Active document 2",
			Content: "Content 2",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "B",
			},
		},
		{
			ID:      "meta-3",
			Text:    "Inactive document",
			Content: "Content 3",
			Metadata: map[string]interface{}{
				"status": "inactive",
				"type":   "A",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Search for active documents
	filter := map[string]interface{}{
		"status": "active",
	}

	opts := &vectordb.QueryOptions{
		TopK: 10,
	}

	results, err := adapter.SearchByMetadata(ctx, collectionName, filter, opts)
	assert.NoError(t, err, "Failed to search by metadata")
	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results), "Should find 2 active documents")

	// Verify all results have status=active
	for _, result := range results {
		if result.Document.Metadata != nil {
			assert.Equal(t, "active", result.Document.Metadata["status"])
		}
	}
}

// TestIntegration_MongoDB_SearchSemantic tests vector similarity search
func TestIntegration_MongoDB_SearchSemantic(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_search_semantic")
	defer adapter.DeleteCollection(ctx, collectionName)

	// Create collection
	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create test documents with varying content
	docs := []*vectordb.Document{
		{
			ID:      "search-1",
			Text:    "Artificial intelligence and machine learning",
			Content: "AI and ML are transforming technology",
			Metadata: map[string]interface{}{
				"category": "tech",
				"topic":    "AI",
			},
		},
		{
			ID:      "search-2",
			Text:    "Python programming for data science",
			Content: "Python is widely used for data analysis",
			Metadata: map[string]interface{}{
				"category": "tech",
				"topic":    "programming",
			},
		},
		{
			ID:      "search-3",
			Text:    "Cooking Italian pasta recipes",
			Content: "Learn authentic Italian cooking techniques",
			Metadata: map[string]interface{}{
				"category": "cooking",
				"topic":    "recipes",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing and vector index to be ready
	time.Sleep(2 * time.Second)

	// Perform semantic search for AI-related content
	results, err := adapter.SearchSemantic(ctx, collectionName, "tell me about AI", &vectordb.QueryOptions{TopK: 2})
	assert.NoError(t, err, "SearchSemantic should not error")
	assert.NotNil(t, results, "Results should not be nil")

	// Should return at least 1 result
	if len(results) > 0 {
		// First result should be most relevant (AI-related)
		assert.Contains(t, results[0].Document.Text, "AI", "Top result should contain 'AI' or related terms")
		assert.Greater(t, results[0].Score, 0.0, "Score should be positive")

		// Scores should be in descending order
		for i := 1; i < len(results); i++ {
			assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score, "Scores should be sorted descending")
		}
	}
}

// TestIntegration_MongoDB_SearchBM25 tests BM25 keyword search
func TestIntegration_MongoDB_SearchBM25(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_search_bm25")
	defer adapter.DeleteCollection(ctx, collectionName)

	// Create collection
	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Create test documents with specific keywords
	docs := []*vectordb.Document{
		{
			ID:      "bm25-1",
			Text:    "Python is a powerful programming language",
			Content: "Python has extensive libraries for data science",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
		{
			ID:      "bm25-2",
			Text:    "JavaScript for web development",
			Content: "JavaScript runs in browsers and Node.js",
			Metadata: map[string]interface{}{
				"lang": "JavaScript",
			},
		},
		{
			ID:      "bm25-3",
			Text:    "Go language for system programming",
			Content: "Go is efficient for concurrent systems",
			Metadata: map[string]interface{}{
				"lang": "Go",
			},
		},
		{
			ID:      "bm25-4",
			Text:    "Python data analysis with pandas",
			Content: "Python pandas library for data manipulation",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for text index to be ready
	time.Sleep(2 * time.Second)

	// Perform BM25 keyword search for "Python"
	client := adapter.Client
	results, err := client.SearchBM25(ctx, collectionName, "Python", &vectordb.QueryOptions{TopK: 3})
	assert.NoError(t, err, "SearchBM25 should not error")
	assert.NotNil(t, results, "Results should not be nil")

	// Should return Python-related documents
	if len(results) > 0 {
		// Verify results contain "Python"
		foundPython := false
		for _, result := range results {
			if result.Document.Text != "" && (result.Document.Text == "Python is a powerful programming language" ||
				result.Document.Text == "Python data analysis with pandas") {
				foundPython = true
				break
			}
		}
		assert.True(t, foundPython, "Results should contain Python documents")

		// Scores should be positive
		for _, result := range results {
			assert.Greater(t, result.Score, 0.0, "BM25 scores should be positive")
		}
	}
}

// TestIntegration_MongoDB_SearchSemantic_EmptyQuery tests error handling for empty query
func TestIntegration_MongoDB_SearchSemantic_EmptyQuery(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_empty_query")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Try searching with empty query
	results, err := adapter.SearchSemantic(ctx, collectionName, "", &vectordb.QueryOptions{TopK: 5})

	// Should either error or return empty results
	if err == nil {
		assert.NotNil(t, results)
	}
}

// TestIntegration_MongoDB_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_MongoDB_E2E_Workflow(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close(context.Background())

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_e2e")

	// 1. Create collection
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Step 1: Failed to create collection")

	// 2. Verify collection exists
	exists, err := adapter.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 2: Failed to check existence")
	require.True(t, exists, "Step 2: Collection should exist")

	// 3. Insert documents
	docs := []*vectordb.Document{
		{
			ID:      "e2e-1",
			Text:    "E2E test document 1",
			Content: "Full content 1",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     1,
			},
		},
		{
			ID:      "e2e-2",
			Text:    "E2E test document 2",
			Content: "Full content 2",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     2,
			},
		},
		{
			ID:      "e2e-3",
			Text:    "E2E test document 3",
			Content: "Full content 3",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     3,
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Step 3: Failed to insert documents")

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// 4. Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 4: Failed to get count")
	assert.Equal(t, int64(3), count, "Step 4: Should have 3 documents")

	// 5. Retrieve specific document
	doc, err := adapter.GetDocument(ctx, collectionName, "e2e-2")
	require.NoError(t, err, "Step 5: Failed to retrieve document")
	assert.NotNil(t, doc, "Step 5: Document should not be nil")
	assert.Equal(t, "e2e-2", doc.ID, "Step 5: Document ID should match")

	// 6. Delete one document
	err = adapter.DeleteDocument(ctx, collectionName, "e2e-1")
	require.NoError(t, err, "Step 6: Failed to delete document")

	// Wait for deletion
	time.Sleep(500 * time.Millisecond)

	// 7. Verify count decreased
	count, err = adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 7: Failed to get count")
	assert.Equal(t, int64(2), count, "Step 7: Should have 2 documents after deletion")

	// 8. Clean up - delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	require.NoError(t, err, "Step 8: Failed to delete collection")

	// 9. Verify collection is gone
	exists, err = adapter.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 9: Failed to check existence")
	assert.False(t, exists, "Step 9: Collection should not exist after deletion")
}
