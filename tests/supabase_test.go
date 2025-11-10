// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
)

// TestSupabaseFactory tests the Supabase factory implementation
func TestSupabaseFactory(t *testing.T) {
	factory := supabase.NewFactory()

	// Test GetSupportedTypes
	supportedTypes := factory.GetSupportedTypes()
	if len(supportedTypes) != 1 {
		t.Errorf("Expected 1 supported type, got %d", len(supportedTypes))
	}
	if supportedTypes[0] != vectordb.VectorDBTypeSupabase {
		t.Errorf("Expected %s, got %s", vectordb.VectorDBTypeSupabase, supportedTypes[0])
	}

	// Test ValidateConfig with valid config
	validConfig := &vectordb.Config{
		Type:        vectordb.VectorDBTypeSupabase,
		DatabaseURL: "postgresql://user:pass@localhost:5432/dbname",
		DatabaseKey: "test-key",
		Timeout:     30,
	}

	if err := factory.ValidateConfig(validConfig); err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}

	// Test ValidateConfig with invalid configs
	invalidConfigs := []*vectordb.Config{
		nil, // nil config
		{
			Type:        vectordb.VectorDBTypeWeaviateCloud,
			DatabaseURL: "postgresql://user:pass@localhost:5432/dbname",
			DatabaseKey: "test-key",
		}, // wrong type
		{
			Type:        vectordb.VectorDBTypeSupabase,
			DatabaseURL: "",
			DatabaseKey: "test-key",
		}, // missing database URL
		{
			Type:        vectordb.VectorDBTypeSupabase,
			DatabaseURL: "postgresql://user:pass@localhost:5432/dbname",
			DatabaseKey: "",
		}, // missing database key
		{
			Type:        vectordb.VectorDBTypeSupabase,
			DatabaseURL: "http://invalid-url",
			DatabaseKey: "test-key",
		}, // invalid URL format
		{
			Type:        vectordb.VectorDBTypeSupabase,
			DatabaseURL: "postgresql://user:pass@localhost:5432/dbname",
			DatabaseKey: "test-key",
			Timeout:     -1,
		}, // negative timeout
	}

	for i, config := range invalidConfigs {
		if err := factory.ValidateConfig(config); err == nil {
			t.Errorf("Invalid config %d should return error", i)
		}
	}
}

// TestSupabaseAdapter tests the Supabase adapter with mock database
func TestSupabaseAdapter(t *testing.T) {
	// Skip if no Supabase credentials are provided
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	if databaseURL == "" || databaseKey == "" {
		t.Skip("Skipping Supabase adapter test: SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:        vectordb.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Timeout:     30,
	}

	adapter, err := supabase.NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create Supabase adapter: %v", err)
	}

	ctx := context.Background()

	// Test health check
	if err := adapter.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test schema operations
	testCollectionName := "test_collection"

	// Test GetDefaultSchema
	defaultSchema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, testCollectionName)
	if defaultSchema == nil {
		t.Error("GetDefaultSchema returned nil")
	}
	if defaultSchema.Class != testCollectionName {
		t.Errorf("Expected class name %s, got %s", testCollectionName, defaultSchema.Class)
	}

	// Test ValidateSchema
	if err := adapter.ValidateSchema(defaultSchema); err != nil {
		t.Errorf("ValidateSchema failed for default schema: %v", err)
	}

	// Test invalid schema validation
	invalidSchema := &vectordb.CollectionSchema{
		Class:      "", // empty class name
		Properties: []vectordb.SchemaProperty{},
	}
	if err := adapter.ValidateSchema(invalidSchema); err == nil {
		t.Error("ValidateSchema should fail for invalid schema")
	}

	// Test collection operations (if we have a test database)
	t.Run("CollectionOperations", func(t *testing.T) {
		testSupabaseCollectionOperations(t, adapter, testCollectionName)
	})

	// Test document operations (if we have a test database)
	t.Run("DocumentOperations", func(t *testing.T) {
		testSupabaseDocumentOperations(t, adapter, testCollectionName)
	})

	// Test query operations (if we have a test database)
	t.Run("QueryOperations", func(t *testing.T) {
		testSupabaseQueryOperations(t, adapter, testCollectionName)
	})
}

func testSupabaseCollectionOperations(t *testing.T, adapter *supabase.Adapter, collectionName string) {
	ctx := context.Background()

	// Clean up any existing collection
	adapter.DeleteCollection(ctx, collectionName)

	// Test CollectionExists (should be false initially)
	exists, err := adapter.CollectionExists(ctx, collectionName)
	if err != nil {
		t.Errorf("CollectionExists failed: %v", err)
	}
	if exists {
		t.Error("Collection should not exist initially")
	}

	// Test CreateCollection
	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
	if err := adapter.CreateCollection(ctx, collectionName, schema); err != nil {
		t.Errorf("CreateCollection failed: %v", err)
		return
	}

	// Test CollectionExists (should be true now)
	exists, err = adapter.CollectionExists(ctx, collectionName)
	if err != nil {
		t.Errorf("CollectionExists failed: %v", err)
	}
	if !exists {
		t.Error("Collection should exist after creation")
	}

	// Test ListCollections
	collections, err := adapter.ListCollections(ctx)
	if err != nil {
		t.Errorf("ListCollections failed: %v", err)
	}

	// Supabase normalizes collection names (underscores -> hyphens)
	expectedName := strings.ReplaceAll(collectionName, "_", "-")
	found := false
	for _, collection := range collections {
		if collection.Name == expectedName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Created collection '%s' (normalized: '%s') not found in list", collectionName, expectedName)
	}

	// Test GetCollectionCount
	count, err := adapter.GetCollectionCount(ctx, collectionName)
	if err != nil {
		t.Errorf("GetCollectionCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Test GetSchema
	retrievedSchema, err := adapter.GetSchema(ctx, collectionName)
	if err != nil {
		t.Errorf("GetSchema failed: %v", err)
	}
	if retrievedSchema.Class != collectionName {
		t.Errorf("Expected class name %s, got %s", collectionName, retrievedSchema.Class)
	}

	// Clean up
	if err := adapter.DeleteCollection(ctx, collectionName); err != nil {
		t.Errorf("DeleteCollection failed: %v", err)
	}
}

func testSupabaseDocumentOperations(t *testing.T, adapter *supabase.Adapter, collectionName string) {
	ctx := context.Background()

	// Create collection for testing
	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
	if err := adapter.CreateCollection(ctx, collectionName, schema); err != nil {
		t.Errorf("CreateCollection failed: %v", err)
		return
	}
	defer adapter.DeleteCollection(ctx, collectionName)

	// Test CreateDocument
	document := &vectordb.Document{
		ID:      "test-doc-1",
		Content: "This is a test document",
		Text:    "Test content",
		Metadata: map[string]interface{}{
			"author": "test-author",
			"type":   "test",
		},
	}

	if err := adapter.CreateDocument(ctx, collectionName, document); err != nil {
		t.Errorf("CreateDocument failed: %v", err)
		return
	}

	// Test GetDocument
	retrievedDoc, err := adapter.GetDocument(ctx, collectionName, document.ID)
	if err != nil {
		t.Errorf("GetDocument failed: %v", err)
		return
	}

	if retrievedDoc.ID != document.ID {
		t.Errorf("Expected ID %s, got %s", document.ID, retrievedDoc.ID)
	}
	if retrievedDoc.Content != document.Content {
		t.Errorf("Expected content %s, got %s", document.Content, retrievedDoc.Content)
	}

	// Test UpdateDocument
	document.Content = "Updated content"
	if err := adapter.UpdateDocument(ctx, collectionName, document); err != nil {
		t.Errorf("UpdateDocument failed: %v", err)
	}

	// Verify update
	updatedDoc, err := adapter.GetDocument(ctx, collectionName, document.ID)
	if err != nil {
		t.Errorf("GetDocument after update failed: %v", err)
	}
	if updatedDoc.Content != "Updated content" {
		t.Errorf("Expected updated content, got %s", updatedDoc.Content)
	}

	// Test CreateDocuments (batch)
	documents := []*vectordb.Document{
		{
			ID:       "test-doc-2",
			Content:  "Second test document",
			Metadata: map[string]interface{}{"batch": true},
		},
		{
			ID:       "test-doc-3",
			Content:  "Third test document",
			Metadata: map[string]interface{}{"batch": true},
		},
	}

	if err := adapter.CreateDocuments(ctx, collectionName, documents); err != nil {
		t.Errorf("CreateDocuments failed: %v", err)
	}

	// Test ListDocuments
	listedDocs, err := adapter.ListDocuments(ctx, collectionName, 10, 0)
	if err != nil {
		t.Errorf("ListDocuments failed: %v", err)
	}
	if len(listedDocs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(listedDocs))
	}

	// Test DeleteDocument
	if err := adapter.DeleteDocument(ctx, collectionName, "test-doc-1"); err != nil {
		t.Errorf("DeleteDocument failed: %v", err)
	}

	// Verify deletion
	_, err = adapter.GetDocument(ctx, collectionName, "test-doc-1")
	if err == nil {
		t.Error("Document should not exist after deletion")
	}

	// Test DeleteDocuments (batch)
	if err := adapter.DeleteDocuments(ctx, collectionName, []string{"test-doc-2", "test-doc-3"}); err != nil {
		t.Errorf("DeleteDocuments failed: %v", err)
	}

	// Verify batch deletion
	finalDocs, err := adapter.ListDocuments(ctx, collectionName, 10, 0)
	if err != nil {
		t.Errorf("ListDocuments after deletion failed: %v", err)
	}
	if len(finalDocs) != 0 {
		t.Errorf("Expected 0 documents after deletion, got %d", len(finalDocs))
	}
}

func testSupabaseQueryOperations(t *testing.T, adapter *supabase.Adapter, collectionName string) {
	ctx := context.Background()

	// Create collection for testing
	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
	if err := adapter.CreateCollection(ctx, collectionName, schema); err != nil {
		t.Errorf("CreateCollection failed: %v", err)
		return
	}
	defer adapter.DeleteCollection(ctx, collectionName)

	// Add test documents
	documents := []*vectordb.Document{
		{
			ID:      "query-doc-1",
			Content: "The quick brown fox jumps over the lazy dog",
			Text:    "Animals and movement",
			Metadata: map[string]interface{}{
				"category": "animals",
				"type":     "sentence",
			},
		},
		{
			ID:      "query-doc-2",
			Content: "Machine learning is a subset of artificial intelligence",
			Text:    "Technology and AI",
			Metadata: map[string]interface{}{
				"category": "technology",
				"type":     "definition",
			},
		},
		{
			ID:      "query-doc-3",
			Content: "The weather is sunny and warm today",
			Text:    "Weather description",
			Metadata: map[string]interface{}{
				"category": "weather",
				"type":     "observation",
			},
		},
	}

	if err := adapter.CreateDocuments(ctx, collectionName, documents); err != nil {
		t.Errorf("CreateDocuments failed: %v", err)
		return
	}

	// Test SearchSemantic
	options := &vectordb.QueryOptions{TopK: 5}
	results, err := adapter.SearchSemantic(ctx, collectionName, "fox", options)
	if err != nil {
		t.Errorf("SearchSemantic failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("SearchSemantic should return at least one result")
	}

	// Test SearchBM25
	results, err = adapter.SearchBM25(ctx, collectionName, "machine learning", options)
	if err != nil {
		t.Errorf("SearchBM25 failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("SearchBM25 should return at least one result")
	}

	// Test SearchHybrid
	results, err = adapter.SearchHybrid(ctx, collectionName, "artificial intelligence", options)
	if err != nil {
		t.Errorf("SearchHybrid failed: %v", err)
	}

	// Test SearchByMetadata
	metadata := map[string]interface{}{
		"category": "animals",
	}
	results, err = adapter.SearchByMetadata(ctx, collectionName, metadata, options)
	if err != nil {
		t.Errorf("SearchByMetadata failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for metadata search, got %d", len(results))
	}
	if len(results) > 0 && results[0].Document.ID != "query-doc-1" {
		t.Errorf("Expected document ID query-doc-1, got %s", results[0].Document.ID)
	}
}
