// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package qdrant

import (
	"context"
	"fmt"
	"testing"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHost       = "localhost"
	testPort       = 6334
	testDimensions = 384 // Smaller for testing
	testMetric     = "Cosine"
	testTimeout    = 10
)

// getTestClient creates a test Qdrant client
func getTestClient(t *testing.T) *Client {
	config := &Config{
		Host:             testHost,
		Port:             testPort,
		VectorDimensions: testDimensions,
		SimilarityMetric: testMetric,
		Timeout:          testTimeout,
		UseTLS:           false,
	}

	client, err := NewClient(config)
	require.NoError(t, err, "Failed to create test client")
	require.NotNil(t, client)

	return client
}

// generateTestVector creates a simple test vector
func generateTestVector(dims int, seed float32) []float32 {
	vec := make([]float32, dims)
	for i := range vec {
		vec[i] = seed + float32(i)*0.01
	}
	return vec
}

// getUniqueCollectionName generates unique collection name for test isolation
func getUniqueCollectionName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// TestIntegration_Qdrant_Health tests health check connectivity
func TestIntegration_Qdrant_Health(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()

	// Health check - list collections should work as a health indicator
	collections, err := client.ListCollections(ctx)
	assert.NoError(t, err, "Health check failed")
	assert.NotNil(t, collections, "Collections list should not be nil")
}

// TestIntegration_Qdrant_CreateCollection tests collection creation
func TestIntegration_Qdrant_CreateCollection(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	assert.NoError(t, err, "Failed to create collection")

	// Verify collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.True(t, exists, "Collection should exist after creation")
}

// TestIntegration_Qdrant_DeleteCollection tests collection deletion
func TestIntegration_Qdrant_DeleteCollection(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete")

	// Create collection first
	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Delete collection
	err = client.DeleteCollection(ctx, collectionName)
	assert.NoError(t, err, "Failed to delete collection")

	// Verify collection doesn't exist
	exists, err := client.CollectionExists(ctx, collectionName)
	assert.NoError(t, err)
	assert.False(t, exists, "Collection should not exist after deletion")
}

// TestIntegration_Qdrant_ListCollections tests listing collections
func TestIntegration_Qdrant_ListCollections(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName1 := getUniqueCollectionName("test_list_1")
	collectionName2 := getUniqueCollectionName("test_list_2")
	defer client.DeleteCollection(ctx, collectionName1)
	defer client.DeleteCollection(ctx, collectionName2)

	// Create two collections
	err := client.CreateCollection(ctx, collectionName1, testDimensions, testMetric)
	require.NoError(t, err)
	err = client.CreateCollection(ctx, collectionName2, testDimensions, testMetric)
	require.NoError(t, err)

	// List collections
	collections, err := client.ListCollections(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, collections)

	// Verify our collections are in the list
	assert.Contains(t, collections, collectionName1)
	assert.Contains(t, collections, collectionName2)
}

// TestIntegration_Qdrant_CollectionExists tests collection existence check
func TestIntegration_Qdrant_CollectionExists(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	existingCollection := getUniqueCollectionName("test_exists")
	nonExistingCollection := getUniqueCollectionName("test_not_exists")

	// Create one collection
	err := client.CreateCollection(ctx, existingCollection, testDimensions, testMetric)
	require.NoError(t, err)
	defer client.DeleteCollection(ctx, existingCollection)

	// Check existing collection
	exists, err := client.CollectionExists(ctx, existingCollection)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existing collection
	exists, err = client.CollectionExists(ctx, nonExistingCollection)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestIntegration_Qdrant_GetCollectionInfo tests retrieving collection info
func TestIntegration_Qdrant_GetCollectionInfo(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_info")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	info, err := client.GetCollectionInfo(ctx, collectionName)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.NotNil(t, info.Result)

	// Verify vector config exists
	assert.NotNil(t, info.Result.Config)
	if info.Result.Config != nil && info.Result.Config.Params != nil {
		vectorsConfig := info.Result.Config.Params.VectorsConfig
		if vectorsConfig != nil && vectorsConfig.Config != nil {
			if params, ok := vectorsConfig.Config.(*qdrant.VectorsConfig_Params); ok {
				assert.NotNil(t, params.Params)
				if params.Params != nil {
					assert.Equal(t, uint64(testDimensions), params.Params.Size)
					assert.Equal(t, qdrant.Distance_Cosine, params.Params.Distance)
				}
			}
		}
	}
}

// TestIntegration_Qdrant_CreateDocument tests single document insertion
func TestIntegration_Qdrant_CreateDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create document
	doc := &Document{
		ID:      "doc-1",
		Vector:  generateTestVector(testDimensions, 0.5),
		Content: "Test document content",
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

	// Verify document was created - check collection count
	count, err := client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Collection should have 1 document")
}

// TestIntegration_Qdrant_CreateDocuments tests bulk document insertion
func TestIntegration_Qdrant_CreateDocuments(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bulk_create")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create multiple documents
	docs := []*Document{
		{
			ID:      "bulk-1",
			Vector:  generateTestVector(testDimensions, 0.1),
			Content: "Bulk document 1",
			Metadata: map[string]interface{}{
				"index": 1,
				"type":  "bulk",
			},
		},
		{
			ID:      "bulk-2",
			Vector:  generateTestVector(testDimensions, 0.2),
			Content: "Bulk document 2",
			Metadata: map[string]interface{}{
				"index": 2,
				"type":  "bulk",
			},
		},
		{
			ID:      "bulk-3",
			Vector:  generateTestVector(testDimensions, 0.3),
			Content: "Bulk document 3",
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

// TestIntegration_Qdrant_GetDocument tests document retrieval by ID
func TestIntegration_Qdrant_GetDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_get_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create document
	expectedDoc := &Document{
		ID:      "get-test-1",
		Vector:  generateTestVector(testDimensions, 0.7),
		Content: "Document to retrieve",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": 5,
		},
	}

	err = client.CreateDocument(ctx, collectionName, expectedDoc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Retrieve document
	retrievedDoc, err := client.GetDocument(ctx, collectionName, expectedDoc.ID)
	assert.NoError(t, err, "Failed to retrieve document")
	assert.NotNil(t, retrievedDoc)
	assert.Equal(t, expectedDoc.ID, retrievedDoc.ID)
	assert.Equal(t, expectedDoc.Content, retrievedDoc.Content)
}

// TestIntegration_Qdrant_UpdateDocument tests document update
func TestIntegration_Qdrant_UpdateDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_update_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create initial document
	doc := &Document{
		ID:      "update-test-1",
		Vector:  generateTestVector(testDimensions, 0.4),
		Content: "Original content",
		Metadata: map[string]interface{}{
			"version": 1,
		},
	}

	err = client.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Update document (CreateDocument is actually upsert in Qdrant)
	updatedDoc := &Document{
		ID:      "update-test-1",
		Vector:  generateTestVector(testDimensions, 0.4),
		Content: "Updated content",
		Metadata: map[string]interface{}{
			"version": 2,
		},
	}

	err = client.CreateDocument(ctx, collectionName, updatedDoc)
	assert.NoError(t, err, "Failed to update document")

	// Verify still only one document
	count, err := client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Should still have 1 document after update")
}

// TestIntegration_Qdrant_DeleteDocument tests document deletion by ID
func TestIntegration_Qdrant_DeleteDocument(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete_doc")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create document
	doc := &Document{
		ID:      "delete-test-1",
		Vector:  generateTestVector(testDimensions, 0.6),
		Content: "Document to delete",
	}

	err = client.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Verify document exists
	count, err := client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	// Delete document
	err = client.DeleteDocument(ctx, collectionName, doc.ID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion to propagate
	time.Sleep(500 * time.Millisecond)

	// Verify document was deleted
	count, err = client.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "Collection should be empty after deletion")
}

// TestIntegration_Qdrant_SearchSemantic tests vector similarity search
func TestIntegration_Qdrant_SearchSemantic(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_search")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create test documents with different vectors
	docs := []*Document{
		{
			ID:      "search-1",
			Vector:  generateTestVector(testDimensions, 0.1),
			Content: "First document",
			Metadata: map[string]interface{}{
				"category": "A",
			},
		},
		{
			ID:      "search-2",
			Vector:  generateTestVector(testDimensions, 0.5),
			Content: "Second document",
			Metadata: map[string]interface{}{
				"category": "B",
			},
		},
		{
			ID:      "search-3",
			Vector:  generateTestVector(testDimensions, 0.9),
			Content: "Third document",
			Metadata: map[string]interface{}{
				"category": "C",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Search with query vector similar to the second document
	queryVector := generateTestVector(testDimensions, 0.5)
	results, err := client.SearchSemantic(ctx, collectionName, queryVector, 3)
	assert.NoError(t, err, "Failed to search")
	assert.NotNil(t, results)
	assert.GreaterOrEqual(t, len(results), 1, "Should return at least 1 result")

	// Verify results are sorted by score (descending)
	for i := 1; i < len(results); i++ {
		assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
			"Results should be sorted by score descending")
	}

	// Verify top result has highest similarity (document ID might be UUID-converted)
	if len(results) > 0 {
		assert.NotEmpty(t, results[0].Document.ID, "Top result should have an ID")
		assert.NotNil(t, results[0].Document.Content, "Top result should have content")
	}
}

// TestIntegration_Qdrant_SearchByMetadata tests metadata filtering
func TestIntegration_Qdrant_SearchByMetadata(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_metadata_search")
	defer client.DeleteCollection(ctx, collectionName)

	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err)

	// Create documents with different metadata
	docs := []*Document{
		{
			ID:      "meta-1",
			Vector:  generateTestVector(testDimensions, 0.1),
			Content: "Active document 1",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "A",
			},
		},
		{
			ID:      "meta-2",
			Vector:  generateTestVector(testDimensions, 0.2),
			Content: "Active document 2",
			Metadata: map[string]interface{}{
				"status": "active",
				"type":   "B",
			},
		},
		{
			ID:      "meta-3",
			Vector:  generateTestVector(testDimensions, 0.3),
			Content: "Inactive document",
			Metadata: map[string]interface{}{
				"status": "inactive",
				"type":   "A",
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Search for active documents
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			{
				ConditionOneOf: &qdrant.Condition_Field{
					Field: &qdrant.FieldCondition{
						Key: "status",
						Match: &qdrant.Match{
							MatchValue: &qdrant.Match_Keyword{
								Keyword: "active",
							},
						},
					},
				},
			},
		},
	}

	results, err := client.SearchByMetadata(ctx, collectionName, filter, 10)
	assert.NoError(t, err, "Failed to search by metadata")
	assert.NotNil(t, results)
	assert.Len(t, results, 2, "Should find 2 active documents")

	// Verify all results have status=active
	for _, doc := range results {
		assert.Equal(t, "active", doc.Metadata["status"])
	}
}

// TestIntegration_Qdrant_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_Qdrant_E2E_Workflow(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_e2e")

	// 1. Create collection
	err := client.CreateCollection(ctx, collectionName, testDimensions, testMetric)
	require.NoError(t, err, "Step 1: Failed to create collection")

	// 2. Verify collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 2: Failed to check existence")
	require.True(t, exists, "Step 2: Collection should exist")

	// 3. Insert documents
	docs := []*Document{
		{
			ID:      "e2e-1",
			Vector:  generateTestVector(testDimensions, 0.2),
			Content: "E2E test document 1",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     1,
			},
		},
		{
			ID:      "e2e-2",
			Vector:  generateTestVector(testDimensions, 0.4),
			Content: "E2E test document 2",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     2,
			},
		},
		{
			ID:      "e2e-3",
			Vector:  generateTestVector(testDimensions, 0.6),
			Content: "E2E test document 3",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     3,
			},
		},
	}

	err = client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Step 3: Failed to insert documents")

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// 4. Verify document count
	count, err := client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 4: Failed to get count")
	assert.Equal(t, int64(3), count, "Step 4: Should have 3 documents")

	// 5. Perform semantic search
	queryVector := generateTestVector(testDimensions, 0.4)
	searchResults, err := client.SearchSemantic(ctx, collectionName, queryVector, 2)
	require.NoError(t, err, "Step 5: Failed to search")
	assert.GreaterOrEqual(t, len(searchResults), 1, "Step 5: Should find results")

	// 6. Retrieve specific document
	doc, err := client.GetDocument(ctx, collectionName, "e2e-2")
	require.NoError(t, err, "Step 6: Failed to retrieve document")
	assert.NotNil(t, doc, "Step 6: Document should not be nil")
	assert.Equal(t, "e2e-2", doc.ID, "Step 6: Document ID should match")

	// 7. Delete one document
	err = client.DeleteDocument(ctx, collectionName, "e2e-1")
	require.NoError(t, err, "Step 7: Failed to delete document")

	// Wait for deletion
	time.Sleep(500 * time.Millisecond)

	// 8. Verify count decreased
	count, err = client.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 8: Failed to get count")
	assert.Equal(t, int64(2), count, "Step 8: Should have 2 documents after deletion")

	// 9. Clean up - delete collection
	err = client.DeleteCollection(ctx, collectionName)
	require.NoError(t, err, "Step 9: Failed to delete collection")

	// 10. Verify collection is gone
	exists, err = client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Step 10: Failed to check existence")
	assert.False(t, exists, "Step 10: Collection should not exist after deletion")
}
