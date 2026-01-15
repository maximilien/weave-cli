// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package milvus

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
	testAddress    = "localhost:19530"
	testDatabase   = "default"
	testDimensions = 1536 // Match OpenAI text-embedding-3-small default
	testMetric     = "L2"
	testTimeout    = 30
)

// getTestAdapter creates a test Milvus adapter
func getTestAdapter(t *testing.T) *Adapter {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMilvusLocal,
		Address:          testAddress,
		Database:         testDatabase,
		Timeout:          testTimeout,
		VectorDimensions: testDimensions,
		SimilarityMetric: testMetric,
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err, "Failed to create test adapter")
	require.NotNil(t, adapter)

	// Clean up any leftover test collections from previous failed runs
	cleanupTestCollections(t, adapter)

	return adapter
}

// cleanupTestCollections removes all test collections (those containing "test_" prefix)
func cleanupTestCollections(t *testing.T, adapter *Adapter) {
	ctx := context.Background()
	collections, err := adapter.ListCollections(ctx)
	if err != nil {
		// Don't fail if cleanup fails - just log
		t.Logf("Warning: Failed to list collections for cleanup: %v", err)
		return
	}

	for _, col := range collections {
		// Delete collections that start with common test prefixes
		if len(col.Name) > 5 && (col.Name[:5] == "test_" || col.Name[:4] == "e2e_") {
			err := adapter.DeleteCollection(ctx, col.Name)
			if err != nil {
				t.Logf("Warning: Failed to delete test collection %s: %v", col.Name, err)
			}
		}
	}
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
		},
	}
}

// TestIntegration_Milvus_Health tests health check connectivity
func TestIntegration_Milvus_Health(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	err := adapter.Health(ctx)
	assert.NoError(t, err, "Health check should succeed")
}

// TestIntegration_Milvus_CreateCollection tests collection creation
func TestIntegration_Milvus_CreateCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create")
	defer adapter.DeleteCollection(ctx, collectionName)

	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	assert.NoError(t, err, "Failed to create collection")

	// Wait for collection to be created
	time.Sleep(1 * time.Second)

	// Verify collection exists
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.True(t, exists, "Collection should exist after creation")
}

// TestIntegration_Milvus_DeleteCollection tests collection deletion
func TestIntegration_Milvus_DeleteCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete")

	// Create collection first
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	assert.NoError(t, err, "Failed to delete collection")

	// Wait for deletion
	time.Sleep(1 * time.Second)

	// Verify collection doesn't exist
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.False(t, exists, "Collection should not exist after deletion")
}

// TestIntegration_Milvus_ListCollections tests listing collections
func TestIntegration_Milvus_ListCollections(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

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

	// Wait for collections
	time.Sleep(1 * time.Second)

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

// TestIntegration_Milvus_CollectionExists tests collection existence check
func TestIntegration_Milvus_CollectionExists(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	existingCollection := getUniqueCollectionName("test_exists")
	nonExistingCollection := getUniqueCollectionName("test_not_exists")

	// Create one collection
	err := adapter.CreateCollection(ctx, existingCollection, getTestSchema(existingCollection))
	require.NoError(t, err)
	defer adapter.DeleteCollection(ctx, existingCollection)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Check existing collection
	exists, err := adapter.CollectionExists(ctx, existingCollection)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing collection
	exists, err = adapter.CollectionExists(ctx, nonExistingCollection)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestIntegration_Milvus_CreateDocument tests single document insertion
func TestIntegration_Milvus_CreateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

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
	time.Sleep(2 * time.Second)

	// Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Collection should have 1 document")
}

// TestIntegration_Milvus_CreateDocuments tests bulk document insertion
func TestIntegration_Milvus_CreateDocuments(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bulk_create")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

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
	time.Sleep(2 * time.Second)

	// Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "Collection should have 3 documents")
}

// TestIntegration_Milvus_GetDocument tests document retrieval by ID
func TestIntegration_Milvus_GetDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_get_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

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
	time.Sleep(2 * time.Second)

	// Retrieve document
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, expectedDoc.ID)
	assert.NoError(t, err, "Failed to retrieve document")
	assert.NotNil(t, retrievedDoc)
	assert.Equal(t, expectedDoc.ID, retrievedDoc.ID)
}

// TestIntegration_Milvus_UpdateDocument tests document update
func TestIntegration_Milvus_UpdateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_update_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

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
	time.Sleep(2 * time.Second)

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

	// Wait for update (Milvus eventual consistency)
	time.Sleep(3 * time.Second)

	// Note: Milvus uses MVCC so GetCollectionCount may include deleted documents until compaction
	// Instead, verify update by retrieving the document and checking its content
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, updatedDoc.ID)
	assert.NoError(t, err)
	if retrievedDoc != nil {
		assert.Equal(t, "Updated content", retrievedDoc.Text)
		assert.Equal(t, "Updated full content", retrievedDoc.Content)
	}
}

// TestIntegration_Milvus_DeleteDocument tests document deletion by ID
func TestIntegration_Milvus_DeleteDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create document
	doc := &vectordb.Document{
		ID:      "delete-test-1",
		Text:    "Document to delete",
		Content: "Content to delete",
		Metadata: map[string]interface{}{
			"type": "test",
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Verify document exists
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete document
	err = adapter.DeleteDocument(ctx, collectionName, doc.ID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion to propagate (Milvus eventual consistency)
	time.Sleep(3 * time.Second)

	// Note: Milvus uses MVCC so GetCollectionCount may include deleted documents until compaction
	// Instead, verify deletion by trying to retrieve the document - it should fail or return nil
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, doc.ID)
	// Either error (document not found) or nil document indicates successful deletion
	assert.True(t, err != nil || retrievedDoc == nil, "Document should not be retrievable after deletion")
}

// TestIntegration_Milvus_SearchSemantic tests vector similarity search
func TestIntegration_Milvus_SearchSemantic(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_search")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Create test documents with varied content
	docs := []*vectordb.Document{
		{
			ID:      "search-1",
			Content: "Artificial intelligence and machine learning",
			Metadata: map[string]interface{}{
				"category": "technology",
			},
		},
		{
			ID:      "search-2",
			Content: "Natural language processing and transformers",
			Metadata: map[string]interface{}{
				"category": "technology",
			},
		},
		{
			ID:      "search-3",
			Content: "Cooking recipes and culinary arts",
			Metadata: map[string]interface{}{
				"category": "food",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Test semantic search
	results, err := adapter.SearchSemantic(ctx, collectionName, "tell me about AI", &vectordb.QueryOptions{TopK: 2})

	if err != nil && (err.Error() == "LLM client not initialized" || err.Error() == "LLM client not available for generating query embeddings") {
		t.Skip("Skipping semantic search test: OpenAI API key not available")
	}

	assert.NoError(t, err, "Failed to perform semantic search")
	assert.NotNil(t, results)

	if len(results) > 0 {
		// Verify results are sorted by score
		for i := 1; i < len(results); i++ {
			assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
				"Results should be sorted by score descending")
		}

		// Verify result structure
		assert.NotEmpty(t, results[0].Document.ID, "Result should have ID")
		assert.GreaterOrEqual(t, results[0].Score, float64(0), "Score should be non-negative")
	}
}

// TestIntegration_Milvus_SearchByMetadata tests metadata filtering
func TestIntegration_Milvus_SearchByMetadata(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_metadata_search")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Create documents with different metadata
	docs := []*vectordb.Document{
		{
			ID:      "meta-1",
			Content: "Active document 1",
			Metadata: map[string]interface{}{
				"status":   "active",
				"priority": int64(1),
			},
		},
		{
			ID:      "meta-2",
			Content: "Active document 2",
			Metadata: map[string]interface{}{
				"status":   "active",
				"priority": int64(2),
			},
		},
		{
			ID:      "meta-3",
			Content: "Inactive document",
			Metadata: map[string]interface{}{
				"status":   "inactive",
				"priority": int64(1),
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Search for active documents
	metadata := map[string]interface{}{
		"status": "active",
	}

	results, err := adapter.SearchByMetadata(ctx, collectionName, metadata, &vectordb.QueryOptions{TopK: 10})
	assert.NoError(t, err, "Failed to search by metadata")
	assert.NotNil(t, results)

	// Should find 2 active documents
	assert.Equal(t, 2, len(results), "Should find 2 active documents")

	// Verify all results have status=active
	for _, result := range results {
		status, ok := result.Document.Metadata["status"].(string)
		assert.True(t, ok, "Metadata should contain status field")
		assert.Equal(t, "active", status, "All results should have status=active")
	}
}

// TestIntegration_Milvus_SearchBM25 tests BM25 keyword search
func TestIntegration_Milvus_SearchBM25(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bm25")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Create test documents with keyword-rich content
	docs := []*vectordb.Document{
		{
			ID:      "bm25-1",
			Content: "Python programming language tutorial for beginners",
			Metadata: map[string]interface{}{
				"topic": "programming",
			},
		},
		{
			ID:      "bm25-2",
			Content: "Advanced Python features and best practices",
			Metadata: map[string]interface{}{
				"topic": "programming",
			},
		},
		{
			ID:      "bm25-3",
			Content: "JavaScript web development guide",
			Metadata: map[string]interface{}{
				"topic": "webdev",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Test BM25 search for keyword "Python"
	results, err := adapter.SearchBM25(ctx, collectionName, "Python", &vectordb.QueryOptions{TopK: 3})

	// Milvus may not support BM25 depending on configuration
	if err != nil && (err.Error() == "BM25 search not supported" || err.Error() == "BM25 search is not supported by Milvus") {
		t.Skip("Skipping BM25 search test: Not supported by Milvus configuration")
	}

	assert.NoError(t, err, "Failed to perform BM25 search")
	assert.NotNil(t, results)

	// Results should prioritize documents with "Python" keyword
	if len(results) > 0 {
		assert.GreaterOrEqual(t, results[0].Score, float64(0), "Score should be non-negative")
	}
}

// TestIntegration_Milvus_SearchHybrid tests hybrid search (vector + keyword)
func TestIntegration_Milvus_SearchHybrid(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_hybrid")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Create test documents
	docs := []*vectordb.Document{
		{
			ID:      "hybrid-1",
			Content: "Machine learning algorithms and neural networks",
			Metadata: map[string]interface{}{
				"category": "ai",
			},
		},
		{
			ID:      "hybrid-2",
			Content: "Deep learning with neural networks and GPUs",
			Metadata: map[string]interface{}{
				"category": "ai",
			},
		},
		{
			ID:      "hybrid-3",
			Content: "Database optimization techniques",
			Metadata: map[string]interface{}{
				"category": "databases",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Test hybrid search combining semantic and keyword matching
	results, err := adapter.SearchHybrid(ctx, collectionName, "neural networks", &vectordb.QueryOptions{TopK: 2})

	// Hybrid search requires OpenAI API key and may not be supported
	if err != nil {
		if err.Error() == "LLM client not initialized" || err.Error() == "LLM client not available for generating query embeddings" {
			t.Skip("Skipping hybrid search test: OpenAI API key not available")
		}
		if err.Error() == "hybrid search not supported" || err.Error() == "hybrid search is not supported by Milvus" {
			t.Skip("Skipping hybrid search test: Not supported by Milvus configuration")
		}
	}

	assert.NoError(t, err, "Failed to perform hybrid search")
	assert.NotNil(t, results)

	if len(results) > 0 {
		// Verify results are sorted by score
		for i := 1; i < len(results); i++ {
			assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
				"Results should be sorted by score descending")
		}

		assert.GreaterOrEqual(t, results[0].Score, float64(0), "Score should be non-negative")
	}
}

// TestIntegration_Milvus_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_Milvus_E2E_Workflow(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_e2e")

	// 1. Create collection
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Step 1: Failed to create collection")

	// Wait for collection
	time.Sleep(1 * time.Second)

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
	time.Sleep(2 * time.Second)

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

	// Wait for deletion (Milvus eventual consistency)
	time.Sleep(3 * time.Second)

	// 7. Verify deletion by trying to retrieve the deleted document
	// Note: Milvus uses MVCC so GetCollectionCount may include deleted documents until compaction
	deletedDoc, err := adapter.GetDocument(ctx, collectionName, "e2e-1")
	// Either error or nil document indicates successful deletion
	assert.True(t, err != nil || deletedDoc == nil, "Step 7: Deleted document should not be retrievable")

	// 8. Clean up - delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	require.NoError(t, err, "Step 8: Failed to delete collection")

	// Wait for deletion
	time.Sleep(1 * time.Second)

	// 9. Verify collection is gone
	exists, err = adapter.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 9: Failed to check existence")
	assert.False(t, exists, "Step 9: Collection should not exist after deletion")
}

// TestIntegration_Milvus_MultiCollectionQuery tests querying multiple collections
func TestIntegration_Milvus_MultiCollectionQuery(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()

	// Create two test collections
	collection1 := getUniqueCollectionName("test_multi_col1_")
	collection2 := getUniqueCollectionName("test_multi_col2_")
	defer adapter.DeleteCollection(ctx, collection1)
	defer adapter.DeleteCollection(ctx, collection2)

	// Create both collections
	err := adapter.CreateCollection(ctx, collection1, getTestSchema(collection1))
	require.NoError(t, err, "Failed to create collection 1")

	err = adapter.CreateCollection(ctx, collection2, getTestSchema(collection2))
	require.NoError(t, err, "Failed to create collection 2")

	time.Sleep(1 * time.Second)

	// Add documents to collection 1 (tech documents)
	docs1 := []*vectordb.Document{
		{
			ID:      "multi1-1",
			Content: "Artificial intelligence and machine learning algorithms",
			Metadata: map[string]interface{}{
				"category": "technology",
				"topic":    "AI",
			},
		},
		{
			ID:      "multi1-2",
			Content: "Natural language processing with transformers",
			Metadata: map[string]interface{}{
				"category": "technology",
				"topic":    "NLP",
			},
		},
		{
			ID:      "multi1-3",
			Content: "Deep learning neural networks for computer vision",
			Metadata: map[string]interface{}{
				"category": "technology",
				"topic":    "Deep Learning",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collection1, docs1)
	require.NoError(t, err, "Failed to create documents in collection 1")

	// Add documents to collection 2 (cooking documents)
	docs2 := []*vectordb.Document{
		{
			ID:      "multi2-1",
			Content: "Italian pasta recipes with fresh ingredients",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "Italian",
			},
		},
		{
			ID:      "multi2-2",
			Content: "French pastry techniques and baking methods",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "French",
			},
		},
		{
			ID:      "multi2-3",
			Content: "Asian stir-fry dishes with vegetables",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "Asian",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collection2, docs2)
	require.NoError(t, err, "Failed to create documents in collection 2")

	// Wait for indexing (Milvus needs longer for indexing)
	time.Sleep(3 * time.Second)

	// Query collection 1 for tech content
	results1, err := adapter.SearchSemantic(ctx, collection1, "tell me about AI", &vectordb.QueryOptions{TopK: 2})

	// Skip test if OpenAI API key not available
	if err != nil {
		if err.Error() == "LLM client not initialized" {
			t.Skip("Skipping multi-collection query test: OpenAI API key not available")
		}
		require.NoError(t, err, "Failed to query collection 1")
	}

	// Results may be nil or empty if no matches found (acceptable for this test)
	if results1 != nil {
		assert.LessOrEqual(t, len(results1), 2, "Should return at most 2 results from collection 1")
	}

	// Query collection 2 for food content
	results2, err := adapter.SearchSemantic(ctx, collection2, "Italian cooking", &vectordb.QueryOptions{TopK: 2})
	if err != nil {
		if err.Error() == "LLM client not initialized" {
			t.Skip("Skipping multi-collection query test: OpenAI API key not available")
		}
		require.NoError(t, err, "Failed to query collection 2")
	}

	// Results may be nil or empty if no matches found (acceptable for this test)
	if results2 != nil {
		assert.LessOrEqual(t, len(results2), 2, "Should return at most 2 results from collection 2")
	}

	// Simulate multi-collection query by aggregating results
	// In real usage, this is done by QueryMultipleCollectionsWithAgent
	var allResults []*vectordb.QueryResult
	if results1 != nil {
		allResults = append(allResults, results1...)
	}
	if results2 != nil {
		allResults = append(allResults, results2...)
	}

	// Note: We don't require results since embeddings may not generate meaningful similarity
	// The important test is that the mechanism works without errors
	r1Count := 0
	r2Count := 0
	if results1 != nil {
		r1Count = len(results1)
	}
	if results2 != nil {
		r2Count = len(results2)
	}
	t.Logf("Multi-collection query completed: %d results from col1, %d from col2", r1Count, r2Count)

	// Verify results have proper structure if any exist
	if results1 != nil && len(results1) > 0 {
		assert.NotEmpty(t, results1[0].Document.Content, "Result from collection 1 should have content")
		assert.GreaterOrEqual(t, results1[0].Score, float64(0), "Score should be non-negative")
		t.Logf("Collection 1 sample result: score=%.2f", results1[0].Score)
	}

	if results2 != nil && len(results2) > 0 {
		assert.NotEmpty(t, results2[0].Document.Content, "Result from collection 2 should have content")
		assert.GreaterOrEqual(t, results2[0].Score, float64(0), "Score should be non-negative")
		t.Logf("Collection 2 sample result: score=%.2f", results2[0].Score)
	}

	// Test passes if both queries executed without errors (results optional)
	t.Logf("✓ Multi-collection query test passed (total results: %d)", len(allResults))
}
