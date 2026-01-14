// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package supabase

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTimeout    = 30
	testDimensions = 1536 // Match OpenAI text-embedding-3-small default
)

// getTestAdapter creates a test Supabase adapter
func getTestAdapter(t *testing.T) vectordb.VectorDBClient {
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	if databaseURL == "" || databaseKey == "" {
		t.Skip("Skipping Supabase integration test: SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeSupabase,
		DatabaseURL:      databaseURL,
		DatabaseKey:      databaseKey,
		Timeout:          testTimeout,
		VectorDimensions: testDimensions,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err, "Failed to create test adapter")
	require.NotNil(t, adapter)

	// Clean up any leftover test collections from previous failed runs
	cleanupTestCollections(t, adapter)

	return adapter
}

// cleanupTestCollections removes all test collections (those containing "test_" prefix)
func cleanupTestCollections(t *testing.T, adapter vectordb.VectorDBClient) {
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
		Class:      collectionName,
		Vectorizer: "none", // Use 'none' to avoid requiring OpenAI API key for basic tests
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "category",
				DataType: []string{"text"},
			},
		},
	}
}

// TestIntegration_Supabase_Health tests health check connectivity
func TestIntegration_Supabase_Health(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	err := adapter.Health(ctx)
	assert.NoError(t, err, "Health check should succeed")
}

// TestIntegration_Supabase_CreateCollection tests collection creation
func TestIntegration_Supabase_CreateCollection(t *testing.T) {
	adapter := getTestAdapter(t)

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

// TestIntegration_Supabase_DeleteCollection tests collection deletion
func TestIntegration_Supabase_DeleteCollection(t *testing.T) {
	adapter := getTestAdapter(t)

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

// TestIntegration_Supabase_ListCollections tests listing collections
func TestIntegration_Supabase_ListCollections(t *testing.T) {
	adapter := getTestAdapter(t)

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
	// Note: Supabase normalizes collection names (underscores -> hyphens)
	normalizedName1 := strings.ReplaceAll(collectionName1, "_", "-")
	normalizedName2 := strings.ReplaceAll(collectionName2, "_", "-")

	found1 := false
	found2 := false
	for _, col := range collections {
		if col.Name == normalizedName1 {
			found1 = true
		}
		if col.Name == normalizedName2 {
			found2 = true
		}
	}

	assert.True(t, found1, "Collection 1 should be in the list")
	assert.True(t, found2, "Collection 2 should be in the list")
}

// TestIntegration_Supabase_CollectionExists tests collection existence check
func TestIntegration_Supabase_CollectionExists(t *testing.T) {
	adapter := getTestAdapter(t)

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

// TestIntegration_Supabase_CreateDocument tests single document insertion
func TestIntegration_Supabase_CreateDocument(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_create_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create document with UUID
	doc := &vectordb.Document{
		ID:      uuid.New().String(),
		Content: "Test document content",
		Metadata: map[string]interface{}{
			"title":    "Test Document",
			"category": "test",
			"index":    int64(1),
			"active":   true,
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

// TestIntegration_Supabase_CreateDocuments tests bulk document insertion
func TestIntegration_Supabase_CreateDocuments(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_bulk_create")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create multiple documents with UUIDs
	docs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Content: "Bulk document 1",
			Metadata: map[string]interface{}{
				"index":    int64(1),
				"category": "bulk",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Bulk document 2",
			Metadata: map[string]interface{}{
				"index":    int64(2),
				"category": "bulk",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Bulk document 3",
			Metadata: map[string]interface{}{
				"index":    int64(3),
				"category": "bulk",
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

// TestIntegration_Supabase_GetDocument tests document retrieval by ID
func TestIntegration_Supabase_GetDocument(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_get_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create document with UUID
	expectedDoc := &vectordb.Document{
		ID:      uuid.New().String(),
		Content: "Document to retrieve",
		Metadata: map[string]interface{}{
			"category": "test",
			"priority": int64(5),
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
	assert.Equal(t, expectedDoc.Content, retrievedDoc.Content)
}

// TestIntegration_Supabase_UpdateDocument tests document update
func TestIntegration_Supabase_UpdateDocument(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_update_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create initial document with UUID
	docID := uuid.New().String()
	doc := &vectordb.Document{
		ID:      docID,
		Content: "Original content",
		Metadata: map[string]interface{}{
			"version":  int64(1),
			"category": "test",
		},
	}

	err = adapter.CreateDocument(ctx, collectionName, doc)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(2 * time.Second)

	// Update document
	updatedDoc := &vectordb.Document{
		ID:      docID,
		Content: "Updated content",
		Metadata: map[string]interface{}{
			"version":  int64(2),
			"category": "updated",
		},
	}

	err = adapter.UpdateDocument(ctx, collectionName, updatedDoc)
	assert.NoError(t, err, "Failed to update document")

	// Wait for update
	time.Sleep(2 * time.Second)

	// Verify update by retrieving the document
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, docID)
	assert.NoError(t, err)
	if retrievedDoc != nil {
		assert.Equal(t, "Updated content", retrievedDoc.Content)
		if category, ok := retrievedDoc.Metadata["category"].(string); ok {
			assert.Equal(t, "updated", category)
		}
	}
}

// TestIntegration_Supabase_DeleteDocument tests document deletion by ID
func TestIntegration_Supabase_DeleteDocument(t *testing.T) {
	adapter := getTestAdapter(t)

	ctx := context.Background()
	collectionName := getUniqueCollectionName("test_delete_doc")
	defer adapter.DeleteCollection(ctx, collectionName)

	err := adapter.CreateCollection(ctx, collectionName, getTestSchema(collectionName))
	require.NoError(t, err)

	// Wait for collection
	time.Sleep(1 * time.Second)

	// Create document with UUID
	docID := uuid.New().String()
	doc := &vectordb.Document{
		ID:      docID,
		Content: "Document to delete",
		Metadata: map[string]interface{}{
			"type":     "test",
			"category": "delete",
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
	err = adapter.DeleteDocument(ctx, collectionName, docID)
	assert.NoError(t, err, "Failed to delete document")

	// Wait for deletion
	time.Sleep(2 * time.Second)

	// Verify deletion by checking count
	count, err = adapter.GetCollectionCount(ctx, collectionName)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count, "Collection should have 0 documents after deletion")
}

// TestIntegration_Supabase_E2E_Workflow tests complete end-to-end workflow
func TestIntegration_Supabase_E2E_Workflow(t *testing.T) {
	adapter := getTestAdapter(t)

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

	// 3. Insert documents with UUIDs
	docID1 := uuid.New().String()
	docID2 := uuid.New().String()
	docID3 := uuid.New().String()

	docs := []*vectordb.Document{
		{
			ID:      docID1,
			Content: "E2E test document 1",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     int64(1),
			},
		},
		{
			ID:      docID2,
			Content: "E2E test document 2",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     int64(2),
			},
		},
		{
			ID:      docID3,
			Content: "E2E test document 3",
			Metadata: map[string]interface{}{
				"workflow": "e2e",
				"step":     int64(3),
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
	doc, err := adapter.GetDocument(ctx, collectionName, docID2)
	require.NoError(t, err, "Step 5: Failed to retrieve document")
	assert.NotNil(t, doc, "Step 5: Document should not be nil")
	assert.Equal(t, docID2, doc.ID, "Step 5: Document ID should match")

	// 6. Delete one document
	err = adapter.DeleteDocument(ctx, collectionName, docID1)
	require.NoError(t, err, "Step 6: Failed to delete document")

	// Wait for deletion
	time.Sleep(2 * time.Second)

	// 7. Verify deletion
	count, err = adapter.GetCollectionCount(ctx, collectionName)
	require.NoError(t, err, "Step 7: Failed to get count")
	assert.Equal(t, int64(2), count, "Step 7: Should have 2 documents after deletion")

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
