// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestWeaviateComprehensiveIntegration(t *testing.T) {
	// Skip if no Weaviate credentials are provided
	weaviateURL := os.Getenv("WEAVIATE_URL")
	weaviateAPIKey := os.Getenv("WEAVIATE_API_KEY")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")

	if weaviateURL == "" || weaviateAPIKey == "" {
		t.Skip("Skipping Weaviate integration test: WEAVIATE_URL and WEAVIATE_API_KEY environment variables not set")
	}

	if openAIAPIKey == "" {
		t.Skip("Skipping Weaviate integration test: OPENAI_API_KEY environment variable not set")
	}

	// Create Weaviate client through vectordb abstraction
	config := &vectordb.Config{
		Type:         vectordb.VectorDBTypeWeaviateCloud,
		URL:          weaviateURL,
		APIKey:       weaviateAPIKey,
		OpenAIAPIKey: openAIAPIKey,
		Timeout:      30,
	}

	client, err := vectordb.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test health check
	t.Run("Health", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	// Test collection operations
	collectionName := "TestCollectionWeaviate"

	// Clean up any existing test collection first
	_ = client.DeleteCollection(ctx, collectionName) // Ignore error if doesn't exist

	t.Run("CreateCollection", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "title",
					DataType: []string{"text"},
				},
				{
					Name:     "category",
					DataType: []string{"text"},
				},
			},
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
		for _, collection := range collections {
			if collection.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created collection should appear in list")
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:      "test-doc-1",
		Content: "This is a test document for Weaviate integration",
		Text:    "Test content",
		Metadata: map[string]interface{}{
			"title":    "Test Document",
			"category": "test",
			"author":   "integration-test",
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
			Content: "Updated test document content",
			Text:    "Updated test content",
			Metadata: map[string]interface{}{
				"title":    "Updated Test Document",
				"category": "updated",
				"author":   "integration-test",
			},
		}

		err := client.UpdateDocument(ctx, collectionName, updatedDoc)
		if err != nil {
			t.Errorf("Failed to update document: %v", err)
			return
		}

		// Verify update
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to get updated document: %v", err)
			return
		}
		if doc == nil {
			t.Error("GetDocument returned nil document")
			return
		}
		if doc.Content != updatedDoc.Content {
			t.Errorf("Expected updated content %s, got %s", updatedDoc.Content, doc.Content)
		}
	})

	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}
		if len(docs) == 0 {
			t.Error("Expected at least one document")
		}

		found := false
		for _, doc := range docs {
			if doc.ID == testDoc.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created document should appear in list")
		}
	})

	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get collection count: %v", err)
		}
		if count == 0 {
			t.Error("Expected at least one document in collection")
		}
	})

	// Test search operations
	t.Run("SearchByContent", func(t *testing.T) {
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		results, err := client.SearchSemantic(ctx, collectionName, "test document", options)
		if err != nil {
			t.Errorf("Failed to search by content: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected at least one search result")
		}
	})

	t.Run("SearchByMetadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"category": "updated",
		}
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		results, err := client.SearchByMetadata(ctx, collectionName, metadata, options)
		if err != nil {
			t.Errorf("Failed to search by metadata: %v", err)
		}
		if len(results) == 0 {
			t.Error("Expected at least one search result")
		}
	})

	t.Run("SearchBM25", func(t *testing.T) {
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		results, err := client.SearchBM25(ctx, collectionName, "updated", options)
		if err != nil {
			t.Errorf("Failed to search with BM25: %v", err)
		}
		t.Logf("BM25 search returned %d results", len(results))
	})

	t.Run("SearchHybrid", func(t *testing.T) {
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		results, err := client.SearchHybrid(ctx, collectionName, "test", options)
		if err != nil {
			t.Errorf("Failed to search with hybrid: %v", err)
		}
		t.Logf("Hybrid search returned %d results", len(results))
	})

	// Test schema operations
	t.Run("GetSchema", func(t *testing.T) {
		schema, err := client.GetSchema(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get schema: %v", err)
		}
		if schema.Class != collectionName {
			t.Errorf("Expected schema class %s, got %s", collectionName, schema.Class)
		}
	})

	t.Run("ValidateSchema", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      "ValidTestCollection",
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
			},
		}

		err := client.ValidateSchema(schema)
		if err != nil {
			t.Errorf("Schema validation failed: %v", err)
		}
	})

	// Test batch document operations
	t.Run("CreateDocuments", func(t *testing.T) {
		docs := []*vectordb.Document{
			{
				ID:      "batch-doc-1",
				Content: "First batch document",
				Metadata: map[string]interface{}{
					"batch": "1",
					"type":  "batch-test",
				},
			},
			{
				ID:      "batch-doc-2",
				Content: "Second batch document",
				Metadata: map[string]interface{}{
					"batch": "1",
					"type":  "batch-test",
				},
			},
		}

		err := client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Errorf("Failed to create batch documents: %v", err)
		}
	})

	t.Run("DeleteDocumentsByMetadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"type": "batch-test",
		}

		err := client.DeleteDocumentsByMetadata(ctx, collectionName, metadata)
		if err != nil {
			t.Errorf("Failed to delete documents by metadata: %v", err)
		}
	})

	// Test with different embeddings/vectorizers
	t.Run("DocumentsWithDifferentEmbeddings", func(t *testing.T) {
		// Test 1: Collection with OpenAI embeddings (text2vec-openai)
		openAICollection := "TestOpenAIEmbeddings"
		_ = client.DeleteCollection(ctx, openAICollection) // Clean up if exists

		openAISchema := &vectordb.CollectionSchema{
			Class:      openAICollection,
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, openAICollection, openAISchema)
		if err != nil {
			t.Errorf("Failed to create OpenAI collection: %v", err)
		}

		// Add document with OpenAI embeddings
		openAIDoc := &vectordb.Document{
			ID:      "openai-doc-1",
			Content: "Artificial intelligence and machine learning are transforming technology",
			Metadata: map[string]interface{}{
				"topic":     "AI",
				"embedding": "openai",
			},
		}

		err = client.CreateDocument(ctx, openAICollection, openAIDoc)
		if err != nil {
			t.Errorf("Failed to create document with OpenAI embeddings: %v", err)
		}

		// Verify semantic search works with OpenAI embeddings
		searchOpts := &vectordb.QueryOptions{TopK: 5}
		results, err := client.SearchSemantic(ctx, openAICollection, "tell me about AI and ML", searchOpts)
		if err != nil {
			t.Errorf("Failed to search with OpenAI embeddings: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected results from OpenAI embedding search")
		} else {
			t.Logf("OpenAI embedding search returned %d results with top score: %.3f",
				len(results), results[0].Score)
		}

		// Test 2: Collection with no vectorizer (manual embeddings)
		noneCollection := "TestNoVectorizer"
		_ = client.DeleteCollection(ctx, noneCollection) // Clean up if exists

		noneSchema := &vectordb.CollectionSchema{
			Class:      noneCollection,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
			},
		}

		err = client.CreateCollection(ctx, noneCollection, noneSchema)
		if err != nil {
			t.Errorf("Failed to create collection with no vectorizer: %v", err)
		}

		// Add document to non-vectorized collection
		noneDoc := &vectordb.Document{
			ID:      "none-doc-1",
			Content: "This document has no automatic embeddings",
			Metadata: map[string]interface{}{
				"topic":     "test",
				"embedding": "none",
			},
		}

		err = client.CreateDocument(ctx, noneCollection, noneDoc)
		if err != nil {
			t.Errorf("Failed to create document in non-vectorized collection: %v", err)
		}

		// Verify document was created
		retrievedDoc, err := client.GetDocument(ctx, noneCollection, noneDoc.ID)
		if err != nil {
			t.Errorf("Failed to retrieve document from non-vectorized collection: %v", err)
		}
		if retrievedDoc.Content != noneDoc.Content {
			t.Errorf("Expected content %s, got %s", noneDoc.Content, retrievedDoc.Content)
		}

		t.Logf("Successfully created documents with different embedding strategies:")
		t.Logf("  - OpenAI embeddings (text2vec-openai): %s", openAIDoc.ID)
		t.Logf("  - No vectorizer (manual): %s", noneDoc.ID)

		// Cleanup test collections
		_ = client.DeleteCollection(ctx, openAICollection)
		_ = client.DeleteCollection(ctx, noneCollection)
	})

	// Cleanup
	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, collectionName, testDoc.ID)
		if err == nil {
			t.Error("Document should not exist after deletion")
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

func TestWeaviateFactoryIntegration(t *testing.T) {
	// Test factory with real configuration
	weaviateURL := os.Getenv("WEAVIATE_URL")
	weaviateAPIKey := os.Getenv("WEAVIATE_API_KEY")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")

	if weaviateURL == "" || weaviateAPIKey == "" || openAIAPIKey == "" {
		t.Skip("Skipping Weaviate factory integration test: WEAVIATE_URL, WEAVIATE_API_KEY, or OPENAI_API_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:         vectordb.VectorDBTypeWeaviateCloud,
		URL:          weaviateURL,
		APIKey:       weaviateAPIKey,
		OpenAIAPIKey: openAIAPIKey,
		Timeout:      30,
	}

	// Test factory creation
	factory := weaviate.NewFactory()

	// Test config validation
	err := factory.ValidateConfig(config)
	if err != nil {
		t.Errorf("Config validation failed: %v", err)
	}

	// Test client creation through factory
	client, err := factory.CreateClient(config)
	if err != nil {
		t.Errorf("Failed to create client through factory: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestWeaviateVectorDBRegistry(t *testing.T) {
	// Test that Weaviate can be created through the global registry
	weaviateURL := os.Getenv("WEAVIATE_URL")
	weaviateAPIKey := os.Getenv("WEAVIATE_API_KEY")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")

	if weaviateURL == "" || weaviateAPIKey == "" || openAIAPIKey == "" {
		t.Skip("Skipping Weaviate registry test: WEAVIATE_URL, WEAVIATE_API_KEY, or OPENAI_API_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:         vectordb.VectorDBTypeWeaviateCloud,
		URL:          weaviateURL,
		APIKey:       weaviateAPIKey,
		OpenAIAPIKey: openAIAPIKey,
		Timeout:      30,
	}

	// Test creation through global registry
	client, err := vectordb.CreateClient(config)
	if err != nil {
		t.Errorf("Failed to create client through global registry: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}
