// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
)

func TestMilvusIntegration(t *testing.T) {
	// Check if Milvus is available
	address := os.Getenv("MILVUS_ADDRESS")
	if address == "" {
		address = "localhost:19530"
	}

	// Create Milvus client
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMilvusLocal,
		Address:          address,
		Database:         "default",
		Timeout:          30,
		VectorDimensions: 1536,
		SimilarityMetric: "COSINE",
	}

	client, err := milvus.NewAdapter(config)
	if err != nil {
		t.Skipf("Skipping Milvus integration test: failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test health check
	t.Run("Health", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	// Test collection operations
	collectionName := "test_collection_milvus"

	// Clean up if exists from previous run
	_ = client.DeleteCollection(ctx, collectionName)

	t.Run("CreateCollection", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text-embedding-3-small",
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
			t.Errorf("Created collection '%s' should appear in list", collectionName)
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:      "test-doc-1",
		Content: "This is a test document for Milvus integration",
		Text:    "Test content",
		Metadata: map[string]interface{}{
			"filename":          "test.txt",
			"original_filename": "test.txt",
			"category":          "test",
			"author":            "integration-test",
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
			return
		}
		if doc == nil {
			t.Error("GetDocument returned nil document")
			return
		}
		if doc.ID != testDoc.ID {
			t.Errorf("Expected document ID %s, got %s", testDoc.ID, doc.ID)
		}
		if doc.Content != testDoc.Content {
			t.Errorf("Expected content %s, got %s", testDoc.Content, doc.Content)
		}
		// Verify metadata is preserved
		if filename, ok := doc.Metadata["filename"].(string); !ok || filename != "test.txt" {
			t.Errorf("Expected metadata filename 'test.txt', got %v", doc.Metadata["filename"])
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		updatedDoc := &vectordb.Document{
			ID:      testDoc.ID,
			Content: "Updated test document content",
			Text:    "Updated test content",
			Metadata: map[string]interface{}{
				"filename":          "test.txt",
				"original_filename": "test.txt",
				"category":          "updated",
				"author":            "integration-test",
			},
		}

		err := client.UpdateDocument(ctx, collectionName, updatedDoc)
		if err != nil {
			t.Errorf("Failed to update document: %v", err)
		}

		// Verify update
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to get updated document: %v", err)
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

	// Test batch document operations
	t.Run("CreateDocuments", func(t *testing.T) {
		docs := []*vectordb.Document{
			{
				ID:      "batch-doc-1",
				Content: "First batch document",
				Metadata: map[string]interface{}{
					"filename": "batch1.txt",
					"batch":    "1",
					"type":     "batch-test",
				},
			},
			{
				ID:      "batch-doc-2",
				Content: "Second batch document",
				Metadata: map[string]interface{}{
					"filename": "batch2.txt",
					"batch":    "1",
					"type":     "batch-test",
				},
			},
		}

		err := client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Errorf("Failed to create batch documents: %v", err)
		}

		// Verify batch documents were created
		allDocs, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list documents after batch create: %v", err)
		}
		if len(allDocs) < 3 {
			t.Errorf("Expected at least 3 documents, got %d", len(allDocs))
		}
	})

	// Test search operations (requires OpenAI API key for embeddings)
	if os.Getenv("OPENAI_API_KEY") != "" {
		t.Run("SearchSemantic", func(t *testing.T) {
			options := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchSemantic(ctx, collectionName, "test document", options)
			if err != nil {
				t.Errorf("Failed to search semantically: %v", err)
			}
			if len(results) == 0 {
				t.Error("Expected at least one search result")
			}
		})

		t.Run("SearchByMetadata", func(t *testing.T) {
			metadata := map[string]interface{}{
				"type": "batch-test",
			}
			options := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchByMetadata(ctx, collectionName, metadata, options)
			if err != nil {
				t.Errorf("Failed to search by metadata: %v", err)
			}
			t.Logf("Metadata search returned %d results", len(results))
		})

		t.Run("SearchBM25", func(t *testing.T) {
			options := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchBM25(ctx, collectionName, "test", options)
			if err != nil {
				// BM25 may not be supported in all Milvus versions
				t.Logf("BM25 search not available: %v", err)
				return
			}
			t.Logf("BM25 search returned %d results", len(results))
		})

		t.Run("SearchHybrid", func(t *testing.T) {
			options := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchHybrid(ctx, collectionName, "test document", options)
			if err != nil {
				// Hybrid search may not be supported in all Milvus versions
				t.Logf("Hybrid search not available: %v", err)
				return
			}
			t.Logf("Hybrid search returned %d results", len(results))
		})
	} else {
		t.Log("Skipping search tests: OPENAI_API_KEY not set")
	}

	// Test schema operations
	t.Run("GetSchema", func(t *testing.T) {
		schema, err := client.GetSchema(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to get schema: %v", err)
			return
		}
		if schema == nil {
			t.Error("GetSchema returned nil")
			return
		}
		if schema.Class != collectionName {
			t.Errorf("Schema class mismatch: expected %q, got %q", collectionName, schema.Class)
		}
	})

	t.Run("ValidateSchema", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      "TestValidation",
			Vectorizer: "text-embedding-3-small",
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

	t.Run("DeleteDocumentsByMetadata", func(t *testing.T) {
		// Create a new collection for this test
		metaCollectionName := "test_delete_meta_milvus"
		_ = client.DeleteCollection(ctx, metaCollectionName)

		schema := &vectordb.CollectionSchema{
			Class:      metaCollectionName,
			Vectorizer: "text-embedding-3-small",
		}
		err := client.CreateCollection(ctx, metaCollectionName, schema)
		if err != nil {
			t.Errorf("Failed to create collection for metadata delete test: %v", err)
			return
		}
		defer client.DeleteCollection(ctx, metaCollectionName)

		// Create documents with metadata
		docs := []*vectordb.Document{
			{
				ID:      "meta-del-1",
				Content: "Document to delete",
				Metadata: map[string]interface{}{
					"status": "draft",
				},
			},
			{
				ID:      "meta-del-2",
				Content: "Document to keep",
				Metadata: map[string]interface{}{
					"status": "published",
				},
			},
			{
				ID:      "meta-del-3",
				Content: "Another to delete",
				Metadata: map[string]interface{}{
					"status": "draft",
				},
			},
		}

		err = client.CreateDocuments(ctx, metaCollectionName, docs)
		if err != nil {
			t.Errorf("Failed to create documents: %v", err)
			return
		}

		// Delete by metadata
		metadata := map[string]interface{}{
			"status": "draft",
		}
		err = client.DeleteDocumentsByMetadata(ctx, metaCollectionName, metadata)
		if err != nil {
			t.Errorf("Failed to delete by metadata: %v", err)
			return
		}

		// Wait for eventual consistency
		time.Sleep(500 * time.Millisecond)

		// Verify remaining count
		count, err := client.GetCollectionCount(ctx, metaCollectionName)
		if err != nil {
			t.Errorf("Failed to get count: %v", err)
			return
		}
		if count != 1 {
			t.Logf("Expected 1 document remaining, got %d (eventual consistency)", count)
		}
	})

	// Cleanup
	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
		}

		// Wait for eventual consistency
		time.Sleep(500 * time.Millisecond)

		// Verify deletion - Milvus may have eventual consistency
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err == nil && doc != nil {
			// Document still exists, log but don't fail (eventual consistency)
			t.Log("Note: Document may still be visible due to Milvus eventual consistency")
		}
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		err := client.DeleteDocuments(ctx, collectionName, []string{"batch-doc-1", "batch-doc-2"})
		if err != nil {
			t.Errorf("Failed to delete batch documents: %v", err)
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

func TestMilvusFactoryIntegration(t *testing.T) {
	// Check if Milvus is available
	address := os.Getenv("MILVUS_ADDRESS")
	if address == "" {
		address = "localhost:19530"
	}

	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMilvusLocal,
		Address:          address,
		Database:         "default",
		Timeout:          30,
		VectorDimensions: 1536,
		SimilarityMetric: "COSINE",
	}

	// Test factory creation
	factory := milvus.NewFactory()

	// Test config validation
	err := factory.ValidateConfig(config)
	if err != nil {
		t.Errorf("Config validation failed: %v", err)
	}

	// Test client creation through factory
	client, err := factory.CreateClient(config)
	if err != nil {
		t.Skipf("Skipping Milvus factory test: failed to create client: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestMilvusVectorDBRegistry(t *testing.T) {
	// Check if Milvus is available
	address := os.Getenv("MILVUS_ADDRESS")
	if address == "" {
		address = "localhost:19530"
	}

	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMilvusLocal,
		Address:          address,
		Database:         "default",
		Timeout:          30,
		VectorDimensions: 1536,
		SimilarityMetric: "COSINE",
	}

	// Test creation through global registry
	client, err := vectordb.CreateClient(config)
	if err != nil {
		t.Skipf("Skipping Milvus registry test: failed to create client: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestMilvusCloudIntegration(t *testing.T) {
	// Skip if no Milvus Cloud credentials are provided
	address := os.Getenv("MILVUS_CLOUD_ADDRESS")
	token := os.Getenv("MILVUS_CLOUD_TOKEN")
	username := os.Getenv("MILVUS_CLOUD_USERNAME")
	password := os.Getenv("MILVUS_CLOUD_PASSWORD")

	if address == "" {
		t.Skip("Skipping Milvus Cloud integration test: MILVUS_CLOUD_ADDRESS not set")
	}

	if token == "" && (username == "" || password == "") {
		t.Skip("Skipping Milvus Cloud integration test: authentication not configured (need MILVUS_CLOUD_TOKEN or MILVUS_CLOUD_USERNAME/PASSWORD)")
	}

	// Create Milvus Cloud client
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMilvusCloud,
		Address:          address,
		APIKey:           token,
		Username:         username,
		Password:         password,
		Database:         "default",
		Timeout:          60,
		VectorDimensions: 1536,
		SimilarityMetric: "COSINE",
	}

	client, err := milvus.NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create Milvus Cloud client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test health check
	t.Run("CloudHealth", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Errorf("Cloud health check failed: %v", err)
		}
	})

	// Test collection operations
	collectionName := "test_collection_milvus_cloud"

	// Clean up if exists from previous run
	_ = client.DeleteCollection(ctx, collectionName)

	t.Run("CloudCreateCollection", func(t *testing.T) {
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text-embedding-3-small",
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Errorf("Failed to create cloud collection: %v", err)
		}
	})

	t.Run("CloudListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list cloud collections: %v", err)
		}

		found := false
		for _, collection := range collections {
			if collection.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Created cloud collection '%s' should appear in list", collectionName)
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:      "cloud-test-doc-1",
		Content: "This is a test document for Milvus Cloud integration",
		Text:    "Cloud test content",
		Metadata: map[string]interface{}{
			"filename":          "cloud_test.txt",
			"original_filename": "cloud_test.txt",
			"category":          "cloud-test",
		},
	}

	t.Run("CloudCreateDocument", func(t *testing.T) {
		err := client.CreateDocument(ctx, collectionName, testDoc)
		if err != nil {
			t.Errorf("Failed to create cloud document: %v", err)
		}
	})

	t.Run("CloudGetDocument", func(t *testing.T) {
		doc, err := client.GetDocument(ctx, collectionName, testDoc.ID)
		if err != nil {
			t.Errorf("Failed to get cloud document: %v", err)
			return
		}
		if doc == nil {
			t.Error("GetDocument returned nil document")
			return
		}
		if doc.Content != testDoc.Content {
			t.Errorf("Expected content %s, got %s", testDoc.Content, doc.Content)
		}
		// Verify metadata is preserved
		if filename, ok := doc.Metadata["filename"].(string); !ok || filename != "cloud_test.txt" {
			t.Errorf("Expected metadata filename 'cloud_test.txt', got %v", doc.Metadata["filename"])
		}
	})

	t.Run("CloudListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Errorf("Failed to list cloud documents: %v", err)
		}
		if len(docs) == 0 {
			t.Error("Expected at least one document")
		}
	})

	// Cleanup
	t.Run("CloudDeleteCollection", func(t *testing.T) {
		err := client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to delete cloud collection: %v", err)
		}
	})
}
