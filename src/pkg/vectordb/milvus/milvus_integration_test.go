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
