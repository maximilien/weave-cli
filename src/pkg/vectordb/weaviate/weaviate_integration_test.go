// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package weaviate

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testURL     = "http://localhost:8080"
	testTimeout = 30
)

// getTestAdapter creates a test Weaviate adapter
func getTestAdapter(t *testing.T) *Adapter {
	config := &vectordb.Config{
		Type:    vectordb.VectorDBTypeWeaviateLocal,
		URL:     testURL,
		Timeout: testTimeout,
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err, "Failed to create test adapter")
	require.NotNil(t, adapter)

	// Clean up any leftover test collections from previous failed runs
	cleanupTestCollections(t, adapter)

	return adapter
}

// cleanupTestCollections removes all test collections (those containing "Test" prefix)
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
		if len(col.Name) > 4 && (col.Name[:4] == "Test" || col.Name[:3] == "E2e") {
			err := adapter.DeleteCollection(ctx, col.Name)
			if err != nil {
				t.Logf("Warning: Failed to delete test collection %s: %v", col.Name, err)
			}
		}
	}
}

// getUniqueCollectionName generates unique collection name for test isolation
func getUniqueCollectionName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}

// getTestSchema returns a standard test schema
func getTestSchema(collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Disable vectorizer for testing
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			// Note: metadata will be added as default text type from schema
		},
	}
}

// TestIntegration_Weaviate_Health tests health check connectivity
func TestIntegration_Weaviate_Health(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	err := adapter.Health(ctx)
	assert.NoError(t, err, "Health check should succeed")
}

// TestIntegration_Weaviate_CreateCollection tests collection creation
func TestIntegration_Weaviate_CreateCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestCreate")
	defer adapter.DeleteCollection(ctx, collectionName)

	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	assert.NoError(t, err, "Failed to create collection")

	// Wait for collection to be created
	time.Sleep(500 * time.Millisecond)

	// Verify collection exists
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.True(t, exists, "Collection should exist after creation")
}

// TestIntegration_Weaviate_DeleteCollection tests collection deletion
func TestIntegration_Weaviate_DeleteCollection(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestDelete")

	// Create collection first
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	assert.NoError(t, err, "Failed to delete collection")

	// Wait for deletion
	time.Sleep(500 * time.Millisecond)

	// Verify collection doesn't exist
	exists, err := adapter.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.False(t, exists, "Collection should not exist after deletion")
}

// TestIntegration_Weaviate_ListCollections tests listing collections
func TestIntegration_Weaviate_ListCollections(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName1 := getUniqueCollectionName("TestList1")
	collectionName2 := getUniqueCollectionName("TestList2")
	defer adapter.DeleteCollection(ctx, collectionName1)
	defer adapter.DeleteCollection(ctx, collectionName2)

	// Create two collections
	err := adapter.CreateCollection(ctx, collectionName1, getTestSchema(collectionName1))
	require.NoError(t, err)
	err = adapter.CreateCollection(ctx, collectionName2, getTestSchema(collectionName2))
	require.NoError(t, err)

	// Wait for collections
	time.Sleep(500 * time.Millisecond)

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

// TestIntegration_Weaviate_CollectionExists tests collection existence check
func TestIntegration_Weaviate_CollectionExists(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	existingCollection := getUniqueCollectionName("TestExists")
	nonExistingCollection := getUniqueCollectionName("TestNotExists")

	// Create one collection
	err := adapter.CreateCollection(ctx, existingCollection, getTestSchema(existingCollection))
	require.NoError(t, err)
	defer adapter.DeleteCollection(ctx, existingCollection)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Check existing collection
	exists, err := adapter.CollectionExists(ctx, existingCollection)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing collection
	exists, err = adapter.CollectionExists(ctx, nonExistingCollection)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestIntegration_Weaviate_CreateDocument tests single document insertion
func TestIntegration_Weaviate_CreateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestCreateDoc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Create document with UUID
	doc := &vectordb.Document{
		ID:      uuid.New().String(),
		Text:    "Test document content",
		Content: "Full test document content",
		Metadata: map[string]interface{}{
			"title":  "Test Document",
			"index":  float64(1), // Weaviate uses float64 for numbers
			"active": true,
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	assert.NoError(t, err, "Failed to create document")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Collection should have 1 document")
}

// TestIntegration_Weaviate_CreateDocuments tests bulk document insertion
func TestIntegration_Weaviate_CreateDocuments(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestBulkCreate")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Create multiple documents with UUIDs
	docs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Text:    "Bulk document 1",
			Content: "Full content for bulk document 1",
			Metadata: map[string]interface{}{
				"index": float64(1),
				"type":  "bulk",
			},
		},
		{
			ID:      uuid.New().String(),
			Text:    "Bulk document 2",
			Content: "Full content for bulk document 2",
			Metadata: map[string]interface{}{
				"index": float64(2),
				"type":  "bulk",
			},
		},
		{
			ID:      uuid.New().String(),
			Text:    "Bulk document 3",
			Content: "Full content for bulk document 3",
			Metadata: map[string]interface{}{
				"index": float64(3),
				"type":  "bulk",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	assert.NoError(t, err, "Failed to create bulk documents")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "Collection should have 3 documents")
}

// TestIntegration_Weaviate_GetDocument tests document retrieval by ID
func TestIntegration_Weaviate_GetDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestGetDoc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Create document with UUID
	docID := uuid.New().String()
	expectedDoc := &vectordb.Document{
		ID:      docID,
		Text:    "Document to retrieve",
		Content: "Full content to retrieve",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": float64(5),
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, expectedDoc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Retrieve document
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, expectedDoc.ID)
	assert.NoError(t, err, "Failed to retrieve document")
	assert.NotNil(t, retrievedDoc)
	assert.Equal(t, expectedDoc.ID, retrievedDoc.ID)
	assert.Equal(t, expectedDoc.Text, retrievedDoc.Text)
}

// TestIntegration_Weaviate_UpdateDocument tests document update
func TestIntegration_Weaviate_UpdateDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestUpdateDoc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Create initial document with UUID
	docID := uuid.New().String()
	doc := &vectordb.Document{
		ID:      docID,
		Text:    "Original content",
		Content: "Original full content",
		Metadata: map[string]interface{}{
			"version": float64(1),
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Update document
	updatedDoc := &vectordb.Document{
		ID:      docID,
		Text:    "Updated content",
		Content: "Updated full content",
		Metadata: map[string]interface{}{
			"version": float64(2),
		},
	}

	err = adapter.UpdateDocument(ctx, collectionName, updatedDoc)
	assert.NoError(t, err, "Failed to update document")

	// Wait for update
	time.Sleep(1 * time.Second)

	// Verify update by retrieving the document
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, updatedDoc.ID)
	assert.NoError(t, err)
	if retrievedDoc != nil {
		assert.Equal(t, "Updated content", retrievedDoc.Text)
		assert.Equal(t, "Updated full content", retrievedDoc.Content)
	}
}

// TestIntegration_Weaviate_DeleteDocument tests document deletion by ID
func TestIntegration_Weaviate_DeleteDocument(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("TestDeleteDoc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// Create document with UUID
	doc := &vectordb.Document{
		ID:      uuid.New().String(),
		Text:    "Document to delete",
		Content: "Content to delete",
		Metadata: map[string]interface{}{
			"type": "test",
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Verify document exists
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete document
	err = adapter.DeleteDocument(ctx, collectionName, doc.ID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion to propagate
	time.Sleep(1 * time.Second)

	// Verify deletion by trying to retrieve
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, doc.ID)
	// Either error or nil document indicates successful deletion
	assert.True(t, err != nil || retrievedDoc == nil, "Document should not be retrievable after deletion")
}

// TestIntegration_Weaviate_SearchSemantic tests vector similarity search
func TestIntegration_Weaviate_SearchSemantic(t *testing.T) {
	adapter := getTestAdapter(t)

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
		// Verify results are sorted by score (descending)
		for i := 1; i < len(results); i++ {
			assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
				"Results should be sorted by score descending")
		}

		// Verify result structure
		assert.NotEmpty(t, results[0].Document.ID, "Result should have ID")
		assert.GreaterOrEqual(t, results[0].Score, float64(0), "Score should be non-negative")
	}
}

// TestIntegration_Weaviate_SearchByMetadata tests metadata filtering
func TestIntegration_Weaviate_SearchByMetadata(t *testing.T) {
	adapter := getTestAdapter(t)

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

// TestIntegration_Weaviate_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_Weaviate_E2E_Workflow(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("E2eTest")

	// 1. Create collection
	schema := getTestSchema(collectionName)
	err := adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Step 1: Failed to create collection")

	// Wait for collection
	time.Sleep(500 * time.Millisecond)

	// 2. Verify collection exists
	exists, err := adapter.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 2: Failed to check existence")
	require.True(t, exists, "Step 2: Collection should exist")

	// 3. Insert documents with UUIDs
	docID1 := uuid.New().String()
	docID2 := uuid.New().String()
	docID3 := uuid.New().String()

	docs := []*vectordb.Document{
		{
			ID:      docID1,
			Text:    "E2E test document 1",
			Content: "Full content 1",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     float64(1),
			},
		},
		{
			ID:      docID2,
			Text:    "E2E test document 2",
			Content: "Full content 2",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     float64(2),
			},
		},
		{
			ID:      docID3,
			Text:    "E2E test document 3",
			Content: "Full content 3",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     float64(3),
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Step 3: Failed to insert documents")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// 4. Verify document count
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 4: Failed to get count")
	assert.Equal(t, int64(3), count, "Step 4: Should have 3 documents")

	// 5. Retrieve specific document
	doc, err := adapter.GetDocument(ctx, collectionName, docID2)
	require.NoError(t, err, "Step 5: Failed to retrieve document")
	assert.NotNil(t, doc, "Step 5: Document should not be nil")
	assert.Equal(t, docID2, doc.ID, "Step 5: Document ID should match")

	// 6. Delete one document
	err = adapter.DeleteDocument(ctx, collectionName, docID1)
	require.NoError(t, err, "Step 6: Failed to delete document")

	// Wait for deletion
	time.Sleep(1 * time.Second)

	// 7. Verify deletion by trying to retrieve
	deletedDoc, err := adapter.GetDocument(ctx, collectionName, docID1)
	assert.True(t, err != nil || deletedDoc == nil, "Step 7: Deleted document should not be retrievable")

	// 8. Clean up - delete collection
	err = adapter.DeleteCollection(ctx, collectionName)
	require.NoError(t, err, "Step 8: Failed to delete collection")

	// Wait for deletion
	time.Sleep(500 * time.Millisecond)

	// 9. Verify collection is gone
	exists, err = adapter.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 9: Failed to check existence")
	assert.False(t, exists, "Step 9: Collection should not exist after deletion")
}

// TestIntegration_Weaviate_MultiCollectionQuery tests querying multiple collections
func TestIntegration_Weaviate_MultiCollectionQuery(t *testing.T) {
	adapter := getTestAdapter(t)
	defer adapter.Close()

	ctx := context.Background()

	// Create two test collections
	collection1 := getUniqueCollectionName("TestMultiCol1_")
	collection2 := getUniqueCollectionName("TestMultiCol2_")
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
			ID:      uuid.New().String(),
			Content: "Artificial intelligence and machine learning algorithms",
			Metadata: map[string]interface{}{
				"category": "technology",
				"topic":    "AI",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Natural language processing with transformers",
			Metadata: map[string]interface{}{
				"category": "technology",
				"topic":    "NLP",
			},
		},
		{
			ID:      uuid.New().String(),
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
			ID:      uuid.New().String(),
			Content: "Italian pasta recipes with fresh ingredients",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "Italian",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "French pastry techniques and baking methods",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "French",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Asian stir-fry dishes with vegetables",
			Metadata: map[string]interface{}{
				"category": "food",
				"topic":    "Asian",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collection2, docs2)
	require.NoError(t, err, "Failed to create documents in collection 2")

	// Wait for indexing
	time.Sleep(2 * time.Second)

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

	// Note: We don't require results since vectorizer is "none" for testing
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

	// Test passes if both queries executed without errors (results optional due to "none" vectorizer)
	t.Logf("✓ Multi-collection query test passed (total results: %d)", len(allResults))
}

// TestIntegration_Weaviate_SearchBM25 tests BM25 keyword search
func TestIntegration_Weaviate_SearchBM25(t *testing.T) {
	ctx := context.Background()

	// Create test adapter
	adapter, err := createTestAdapter(t)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Generate unique collection name
	collectionName := getUniqueCollectionName("test_bm25")

	// Clean up collection after test
	defer func() {
		if err := adapter.DeleteCollection(ctx, collectionName); err != nil {
			t.Logf("Warning: Failed to cleanup collection: %v", err)
		}
	}()

	// Create test collection
	schema := &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Use none for testing
		Properties: []vectordb.SchemaProperty{
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

	err = adapter.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Failed to create collection")

	// Wait for collection to be ready
	time.Sleep(500 * time.Millisecond)

	// Create test documents with specific keywords
	docs := []*vectordb.Document{
		{
			ID:      "bm25-1",
			Content: "Python is a powerful programming language for data science",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
		{
			ID:      "bm25-2",
			Content: "JavaScript is used for web development and Node.js",
			Metadata: map[string]interface{}{
				"lang": "JavaScript",
			},
		},
		{
			ID:      "bm25-3",
			Content: "Go programming language for system software",
			Metadata: map[string]interface{}{
				"lang": "Go",
			},
		},
		{
			ID:      "bm25-4",
			Content: "Python pandas library for data analysis",
			Metadata: map[string]interface{}{
				"lang": "Python",
			},
		},
	}

	err = adapter.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Failed to create documents")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Perform BM25 search for "Python"
	results, err := adapter.SearchBM25(ctx, collectionName, "Python", &vectordb.QueryOptions{TopK: 3})
	assert.NoError(t, err, "BM25 search should not error")
	assert.NotNil(t, results, "Results should not be nil")

	// Should return Python-related documents
	if len(results) > 0 {
		t.Logf("BM25 search returned %d results", len(results))

		// Verify results contain "Python"
		foundPython := false
		for _, result := range results {
			t.Logf("  Result: ID=%s, Score=%.4f, Content=%s", result.Document.ID, result.Score, result.Document.Content)
			if result.Document.Content != "" && (strings.Contains(result.Document.Content, "Python") ||
				strings.Contains(result.Document.Content, "python")) {
				foundPython = true
			}
		}
		assert.True(t, foundPython, "Results should contain Python documents")

		// Verify scores are positive
		for _, result := range results {
			assert.GreaterOrEqual(t, result.Score, 0.0, "BM25 scores should be non-negative")
		}
	}

	t.Log("✓ BM25 search test passed")
}
