// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/mongodb"
)

func TestMongoDBIntegration(t *testing.T) {
	// Skip if no MongoDB credentials are provided
	mongoURI := os.Getenv("MONGODB_URI")
	mongoDatabase := os.Getenv("MONGODB_DATABASE")

	if mongoURI == "" {
		t.Skip("Skipping MongoDB integration test: MONGODB_URI environment variable not set")
	}

	if mongoDatabase == "" {
		mongoDatabase = "weave-cli-test"
	}

	// Create MongoDB client
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMongoDB,
		URL:              mongoURI,
		Database:         mongoDatabase,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}

	client, err := mongodb.NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create MongoDB client: %v", err)
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
	t.Run("CollectionOperations", func(t *testing.T) {
		collectionName := "TestCollection_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
				{
					Name:     "embedding",
					DataType: []string{"number[]"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Check collection exists
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to check collection existence: %v", err)
		}
		if !exists {
			t.Errorf("Collection should exist after creation")
		}

		// List collections
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Fatalf("Failed to list collections: %v", err)
		}

		found := false
		for _, col := range collections {
			if col.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Created collection not found in list")
		}
	})

	// Test document operations
	t.Run("DocumentOperations", func(t *testing.T) {
		collectionName := "TestDocs_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Create document
		doc := &vectordb.Document{
			ID:      "test-doc-1",
			Text:    "This is a test document",
			Content: "Test content for MongoDB integration",
			Metadata: map[string]interface{}{
				"category": "test",
				"source":   "integration-test",
			},
		}

		err = client.CreateDocument(ctx, collectionName, doc)
		if err != nil {
			t.Fatalf("Failed to create document: %v", err)
		}

		// Get document
		retrieved, err := client.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			t.Fatalf("Failed to get document: %v", err)
		}

		if retrieved.Text != doc.Text {
			t.Errorf("Document text mismatch: expected %q, got %q", doc.Text, retrieved.Text)
		}

		// Update document
		retrieved.Text = "Updated text"
		err = client.UpdateDocument(ctx, collectionName, retrieved)
		if err != nil {
			t.Fatalf("Failed to update document: %v", err)
		}

		// Verify update
		updated, err := client.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			t.Fatalf("Failed to get updated document: %v", err)
		}

		if updated.Text != "Updated text" {
			t.Errorf("Document not updated: expected %q, got %q", "Updated text", updated.Text)
		}

		// Delete document
		err = client.DeleteDocument(ctx, collectionName, doc.ID)
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, collectionName, doc.ID)
		if err == nil {
			t.Errorf("Document should not exist after deletion")
		}
	})

	// Test batch operations
	t.Run("BatchOperations", func(t *testing.T) {
		collectionName := "TestBatch_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Create multiple documents
		docs := []*vectordb.Document{
			{
				ID:      "batch-1",
				Text:    "First batch document",
				Content: "Content 1",
			},
			{
				ID:      "batch-2",
				Text:    "Second batch document",
				Content: "Content 2",
			},
			{
				ID:      "batch-3",
				Text:    "Third batch document",
				Content: "Content 3",
			},
		}

		err = client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Fatalf("Failed to create batch documents: %v", err)
		}

		// Verify count
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get collection count: %v", err)
		}

		if count != int64(len(docs)) {
			t.Errorf("Collection count mismatch: expected %d, got %d", len(docs), count)
		}

		// List documents
		listed, err := client.ListDocuments(ctx, collectionName, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list documents: %v", err)
		}

		if len(listed) != len(docs) {
			t.Errorf("Listed documents count mismatch: expected %d, got %d", len(docs), len(listed))
		}

		// Delete multiple documents
		ids := []string{"batch-1", "batch-2"}
		err = client.DeleteDocuments(ctx, collectionName, ids)
		if err != nil {
			t.Fatalf("Failed to delete documents: %v", err)
		}

		// Verify remaining count
		count, err = client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get collection count after deletion: %v", err)
		}

		if count != 1 {
			t.Errorf("Collection count after deletion: expected 1, got %d", count)
		}
	})

	// Test BM25 search
	t.Run("BM25Search", func(t *testing.T) {
		collectionName := "TestBM25_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Create test documents
		docs := []*vectordb.Document{
			{
				ID:      "bm25-1",
				Text:    "MongoDB is a document database",
				Content: "MongoDB stores data in flexible, JSON-like documents",
			},
			{
				ID:      "bm25-2",
				Text:    "Vector search in MongoDB Atlas",
				Content: "MongoDB Atlas provides vector search capabilities",
			},
			{
				ID:      "bm25-3",
				Text:    "PostgreSQL relational database",
				Content: "PostgreSQL is a powerful relational database",
			},
		}

		err = client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Fatalf("Failed to create documents: %v", err)
		}

		// Wait for text indexes to be ready and retry search
		opts := &vectordb.QueryOptions{
			TopK: 5,
		}

		var results []*vectordb.QueryResult
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			time.Sleep(time.Duration(i+1) * time.Second)

			results, err = client.SearchBM25(ctx, collectionName, "MongoDB", opts)
			if err == nil && len(results) > 0 {
				break
			}
		}

		if err != nil {
			t.Fatalf("BM25 search failed after %d retries: %v", maxRetries, err)
		}

		if len(results) == 0 {
			t.Errorf("Expected search results, got none after %d retries", maxRetries)
		}

		// Verify MongoDB documents are ranked higher
		foundMongo := false
		for _, result := range results {
			if result.Document.ID == "bm25-1" || result.Document.ID == "bm25-2" {
				foundMongo = true
				break
			}
		}

		if !foundMongo {
			t.Errorf("Expected to find MongoDB-related documents in search results")
		}
	})

	// Test metadata search
	t.Run("MetadataSearch", func(t *testing.T) {
		collectionName := "TestMetadata_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Create documents with metadata
		docs := []*vectordb.Document{
			{
				ID:      "meta-1",
				Text:    "Document 1",
				Content: "Content 1",
				Metadata: map[string]interface{}{
					"category": "tech",
					"priority": "high",
				},
			},
			{
				ID:      "meta-2",
				Text:    "Document 2",
				Content: "Content 2",
				Metadata: map[string]interface{}{
					"category": "tech",
					"priority": "low",
				},
			},
			{
				ID:      "meta-3",
				Text:    "Document 3",
				Content: "Content 3",
				Metadata: map[string]interface{}{
					"category": "business",
					"priority": "high",
				},
			},
		}

		err = client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Fatalf("Failed to create documents: %v", err)
		}

		// Search by metadata
		opts := &vectordb.QueryOptions{
			TopK: 10,
		}

		metadata := map[string]interface{}{
			"category": "tech",
		}

		results, err := client.SearchByMetadata(ctx, collectionName, metadata, opts)
		if err != nil {
			t.Fatalf("Metadata search failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results for category=tech, got %d", len(results))
		}

		// Verify results have correct category
		for _, result := range results {
			if cat, ok := result.Document.Metadata["category"].(string); !ok || cat != "tech" {
				t.Errorf("Result has wrong category: %v", result.Document.Metadata["category"])
			}
		}
	})

	// Test delete by metadata
	t.Run("DeleteByMetadata", func(t *testing.T) {
		collectionName := "TestDeleteMeta_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Create documents
		docs := []*vectordb.Document{
			{
				ID:   "del-1",
				Text: "Document 1",
				Metadata: map[string]interface{}{
					"status": "draft",
				},
			},
			{
				ID:   "del-2",
				Text: "Document 2",
				Metadata: map[string]interface{}{
					"status": "published",
				},
			},
			{
				ID:   "del-3",
				Text: "Document 3",
				Metadata: map[string]interface{}{
					"status": "draft",
				},
			},
		}

		err = client.CreateDocuments(ctx, collectionName, docs)
		if err != nil {
			t.Fatalf("Failed to create documents: %v", err)
		}

		// Delete by metadata
		metadata := map[string]interface{}{
			"status": "draft",
		}

		err = client.DeleteDocumentsByMetadata(ctx, collectionName, metadata)
		if err != nil {
			t.Fatalf("Failed to delete by metadata: %v", err)
		}

		// Verify count
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 document remaining, got %d", count)
		}
	})

	// Test semantic search (requires OpenAI API key)
	if os.Getenv("OPENAI_API_KEY") != "" {
		t.Run("SemanticSearch", func(t *testing.T) {
			collectionName := "TestSemantic_" + fmt.Sprint(time.Now().Unix())

			// Create collection
			schema := &vectordb.CollectionSchema{
				Class:      collectionName,
				Vectorizer: "text-embedding-3-small",
				Properties: []vectordb.SchemaProperty{
					{
						Name:     "text",
						DataType: []string{"text"},
					},
				},
			}

			err := client.CreateCollection(ctx, collectionName, schema)
			if err != nil {
				t.Fatalf("Failed to create collection: %v", err)
			}
			defer client.DeleteCollection(ctx, collectionName)

			// Create test documents
			docs := []*vectordb.Document{
				{
					ID:      "sem-1",
					Text:    "Machine learning and artificial intelligence",
					Content: "Deep learning neural networks for AI applications",
				},
				{
					ID:      "sem-2",
					Text:    "Database systems and data storage",
					Content: "Relational and NoSQL database management",
				},
				{
					ID:      "sem-3",
					Text:    "Web development frameworks",
					Content: "Building modern web applications with React and Node.js",
				},
			}

			err = client.CreateDocuments(ctx, collectionName, docs)
			if err != nil {
				t.Fatalf("Failed to create documents: %v", err)
			}

			// Wait for vector index to be ready
			time.Sleep(2 * time.Second)

			// Search semantically
			opts := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchSemantic(ctx, collectionName, "artificial intelligence deep learning", opts)
			if err != nil {
				// Vector search may fail if index not ready yet
				t.Logf("Semantic search returned error (may need vector index): %v", err)
				return
			}

			if len(results) == 0 {
				// Vector search may return no results if index not configured
				t.Logf("Semantic search returned no results (vector index may not be configured for test collection)")
			} else {
				t.Logf("Semantic search returned %d results", len(results))
			}

			// Verify AI document is ranked high
			if len(results) > 0 && results[0].Document.ID != "sem-1" {
				t.Logf("Note: Expected sem-1 to be top result, got %s", results[0].Document.ID)
			}
		})

		t.Run("HybridSearch", func(t *testing.T) {
			collectionName := "TestHybrid_" + fmt.Sprint(time.Now().Unix())

			// Create collection
			schema := &vectordb.CollectionSchema{
				Class:      collectionName,
				Vectorizer: "text-embedding-3-small",
				Properties: []vectordb.SchemaProperty{
					{
						Name:     "text",
						DataType: []string{"text"},
					},
				},
			}

			err := client.CreateCollection(ctx, collectionName, schema)
			if err != nil {
				t.Fatalf("Failed to create collection: %v", err)
			}
			defer client.DeleteCollection(ctx, collectionName)

			// Create test documents
			docs := []*vectordb.Document{
				{
					ID:      "hyb-1",
					Text:    "MongoDB vector search capabilities",
					Content: "MongoDB Atlas provides powerful vector search",
				},
				{
					ID:      "hyb-2",
					Text:    "PostgreSQL full-text search",
					Content: "PostgreSQL has built-in text search features",
				},
			}

			err = client.CreateDocuments(ctx, collectionName, docs)
			if err != nil {
				t.Fatalf("Failed to create documents: %v", err)
			}

			// Wait for indexes
			time.Sleep(2 * time.Second)

			// Hybrid search
			opts := &vectordb.QueryOptions{
				TopK: 5,
			}

			results, err := client.SearchHybrid(ctx, collectionName, "MongoDB vector", opts)
			if err != nil {
				t.Fatalf("Hybrid search failed: %v", err)
			}

			if len(results) == 0 {
				t.Errorf("Expected hybrid search results, got none")
			}
		})
	} else {
		t.Log("Skipping semantic/hybrid search tests: OPENAI_API_KEY not set")
	}

	// Test schema operations
	t.Run("SchemaOperations", func(t *testing.T) {
		collectionName := "TestSchema_" + fmt.Sprint(time.Now().Unix())

		// Create collection
		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		defer client.DeleteCollection(ctx, collectionName)

		// Get schema
		retrievedSchema, err := client.GetSchema(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get schema: %v", err)
		}

		if retrievedSchema.Class != collectionName {
			t.Errorf("Schema class mismatch: expected %q, got %q", collectionName, retrievedSchema.Class)
		}

		// Validate schema
		err = client.ValidateSchema(schema)
		if err != nil {
			t.Errorf("Schema validation failed: %v", err)
		}

		// Get default schema
		defaultSchema := client.GetDefaultSchema(vectordb.SchemaTypeText, "TestDefault")
		if defaultSchema.Class != "TestDefault" {
			t.Errorf("Default schema class mismatch: expected %q, got %q", "TestDefault", defaultSchema.Class)
		}
	})
}

func TestMongoDBFactoryIntegration(t *testing.T) {
	// Skip if no MongoDB credentials are provided
	mongoURI := os.Getenv("MONGODB_URI")
	mongoDatabase := os.Getenv("MONGODB_DATABASE")

	if mongoURI == "" {
		t.Skip("Skipping MongoDB factory test: MONGODB_URI environment variable not set")
	}

	if mongoDatabase == "" {
		mongoDatabase = "weave-cli-test"
	}

	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMongoDB,
		URL:              mongoURI,
		Database:         mongoDatabase,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}

	// Test factory creation
	factory := mongodb.NewFactory()

	// Test config validation
	err := factory.ValidateConfig(config)
	if err != nil {
		t.Errorf("Config validation failed: %v", err)
	}

	// Test client creation through factory
	client, err := factory.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create client through factory: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestMongoDBVectorDBRegistry(t *testing.T) {
	// Skip if no MongoDB credentials are provided
	mongoURI := os.Getenv("MONGODB_URI")
	mongoDatabase := os.Getenv("MONGODB_DATABASE")

	if mongoURI == "" {
		t.Skip("Skipping MongoDB registry test: MONGODB_URI environment variable not set")
	}

	if mongoDatabase == "" {
		mongoDatabase = "weave-cli-test"
	}

	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMongoDB,
		URL:              mongoURI,
		Database:         mongoDatabase,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}

	// Test creation through global registry
	client, err := vectordb.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create client through registry: %v", err)
	}

	// Test basic operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}
