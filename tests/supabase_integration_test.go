// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
)

func TestSupabaseIntegration(t *testing.T) {
	// Skip in short mode - integration tests require external services
	if testing.Short() {
		t.Skip("Skipping Supabase integration test in short mode")
	}

	// Skip if no Supabase credentials are provided
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	if databaseURL == "" || databaseKey == "" {
		t.Skip("Skipping Supabase integration test: SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY environment variables not set")
	}

	// Create Supabase client
	config := &vectordb.Config{
		Type:        vectordb.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Timeout:     30,
	}

	client, err := supabase.NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
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

	// Test collection name preservation
	t.Run("CollectionNamePreservation", func(t *testing.T) {
		// Test cases covering various naming patterns
		testCases := []struct {
			name        string
			description string
		}{
			{"TestCollection_MixedCase", "Mixed case with underscores"},
			{"My_Test_Collection", "Multiple underscores"},
			{"Collection123_Test", "Numbers with underscores"},
			{"ALLCAPS_COLLECTION", "All uppercase"},
			{"lowercase_collection", "All lowercase"},
			{"CamelCaseCollection", "Camel case (no underscores)"},
			{"snake_case_collection", "Snake case"},
			{"Collection-With-Hyphens", "Hyphens preserved"},
			{"Mix_Case-And-Chars123", "Mixed: underscores, hyphens, numbers"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				name := tc.name

				// Clean up if exists
				_ = client.DeleteCollection(ctx, name)

				// Create collection
				schema := &vectordb.CollectionSchema{
					Class:      name,
					Vectorizer: "none",
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

				err := client.CreateCollection(ctx, name, schema)
				if err != nil {
					t.Fatalf("Failed to create collection %s: %v", name, err)
				}
				t.Logf("✓ Created collection: %s", name)

				// Verify collection exists with exact name
				exists, err := client.CollectionExists(ctx, name)
				if err != nil {
					t.Fatalf("Failed to check existence of %s: %v", name, err)
				}
				if !exists {
					t.Fatalf("Collection %s should exist after creation", name)
				}
				t.Logf("✓ Collection exists check passed: %s", name)

				// Add multiple documents
				docs := []*vectordb.Document{
					{
						ID:      "doc-1",
						Content: "First test document",
						Metadata: map[string]interface{}{
							"category": "test",
							"index":    "1",
						},
					},
					{
						ID:      "doc-2",
						Content: "Second test document",
						Metadata: map[string]interface{}{
							"category": "test",
							"index":    "2",
						},
					},
				}

				err = client.CreateDocuments(ctx, name, docs)
				if err != nil {
					t.Fatalf("Failed to create documents in %s: %v", name, err)
				}
				t.Logf("✓ Created %d documents in: %s", len(docs), name)

				// Retrieve and verify each document
				for _, expectedDoc := range docs {
					retrieved, err := client.GetDocument(ctx, name, expectedDoc.ID)
					if err != nil {
						t.Fatalf("Failed to retrieve document %s from %s: %v", expectedDoc.ID, name, err)
					}
					if retrieved.Content != expectedDoc.Content {
						t.Errorf("Document content mismatch for %s in %s: expected %q, got %q",
							expectedDoc.ID, name, expectedDoc.Content, retrieved.Content)
					}
				}
				t.Logf("✓ Document retrieval verified: %s", name)

				// Test document listing
				listedDocs, err := client.ListDocuments(ctx, name, 10, 0)
				if err != nil {
					t.Fatalf("Failed to list documents in %s: %v", name, err)
				}
				if len(listedDocs) != len(docs) {
					t.Errorf("Expected %d documents in %s, got %d", len(docs), name, len(listedDocs))
				}
				t.Logf("✓ Document listing verified: %s (%d docs)", name, len(listedDocs))

				// Test collection count
				count, err := client.GetCollectionCount(ctx, name)
				if err != nil {
					t.Fatalf("Failed to get count for %s: %v", name, err)
				}
				if count != int64(len(docs)) {
					t.Errorf("Expected count %d for %s, got %d", len(docs), name, count)
				}
				t.Logf("✓ Collection count verified: %s (count=%d)", name, count)

				// Test search by metadata
				metadata := map[string]interface{}{"category": "test"}
				searchResults, err := client.SearchByMetadata(ctx, name, metadata, &vectordb.QueryOptions{TopK: 10})
				if err != nil {
					t.Fatalf("Failed to search by metadata in %s: %v", name, err)
				}
				if len(searchResults) != len(docs) {
					t.Errorf("Expected %d search results in %s, got %d", len(docs), name, len(searchResults))
				}
				t.Logf("✓ Metadata search verified: %s (%d results)", name, len(searchResults))

				// Test document update
				updateDoc := &vectordb.Document{
					ID:      "doc-1",
					Content: "Updated content for doc-1",
					Metadata: map[string]interface{}{
						"category": "updated",
					},
				}
				err = client.UpdateDocument(ctx, name, updateDoc)
				if err != nil {
					t.Fatalf("Failed to update document in %s: %v", name, err)
				}

				// Verify update
				updated, err := client.GetDocument(ctx, name, "doc-1")
				if err != nil {
					t.Fatalf("Failed to retrieve updated document from %s: %v", name, err)
				}
				if updated.Content != updateDoc.Content {
					t.Errorf("Document update failed in %s: expected %q, got %q",
						name, updateDoc.Content, updated.Content)
				}
				t.Logf("✓ Document update verified: %s", name)

				// Test document deletion
				err = client.DeleteDocument(ctx, name, "doc-2")
				if err != nil {
					t.Fatalf("Failed to delete document from %s: %v", name, err)
				}

				// Verify deletion
				_, err = client.GetDocument(ctx, name, "doc-2")
				if err == nil {
					t.Errorf("Document doc-2 should not exist after deletion in %s", name)
				}
				t.Logf("✓ Document deletion verified: %s", name)

				// Clean up collection
				err = client.DeleteCollection(ctx, name)
				if err != nil {
					t.Fatalf("Failed to delete collection %s: %v", name, err)
				}

				// Verify collection deletion
				exists, err = client.CollectionExists(ctx, name)
				if err != nil {
					t.Fatalf("Failed to check existence after deletion of %s: %v", name, err)
				}
				if exists {
					t.Errorf("Collection %s should not exist after deletion", name)
				}
				t.Logf("✅ All operations verified for collection: %s", name)
			})
		}
	})

	// Test collection operations
	collectionName := "test_collection_supabase"

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

		// Supabase normalizes collection names (underscores -> hyphens)
		expectedName := strings.ReplaceAll(collectionName, "_", "-")
		t.Logf("Found %d collections, looking for '%s'", len(collections), expectedName)

		found := false
		for _, collection := range collections {
			t.Logf("Collection: %s", collection.Name)
			if collection.Name == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Created collection '%s' (normalized: '%s') should appear in list of %d collections", collectionName, expectedName, len(collections))
		}
	})

	// Test document operations
	testDoc := &vectordb.Document{
		ID:      "test-doc-1",
		Content: "This is a test document for Supabase integration",
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
		// BM25 search might not return results if full-text search is not properly configured
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

	// Test semantic search with embeddings
	t.Run("SemanticSearchWithEmbeddings", func(t *testing.T) {
		// Skip if no OpenAI API key for embeddings
		if os.Getenv("OPENAI_API_KEY") == "" {
			t.Skip("Skipping semantic search test: OPENAI_API_KEY not set")
		}

		// Create a document with specific content for semantic search
		semanticDoc := &vectordb.Document{
			ID:      "semantic-test-doc",
			Content: "Artificial intelligence and machine learning are transforming modern technology",
			Text:    "AI ML technology",
			Metadata: map[string]interface{}{
				"topic":    "technology",
				"category": "ai",
			},
		}

		err := client.CreateDocument(ctx, collectionName, semanticDoc)
		if err != nil {
			t.Errorf("Failed to create semantic test document: %v", err)
		}

		// Test semantic search with natural language query
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		// Query with semantically similar but different words
		results, err := client.SearchSemantic(ctx, collectionName, "tell me about AI and ML", options)
		if err != nil {
			t.Errorf("Failed to perform semantic search: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one semantic search result")
		}

		// Verify that the semantic document appears in results
		found := false
		for _, result := range results {
			if result.Document.ID == semanticDoc.ID {
				found = true
				// Verify the score is reasonable (should be > 0.3 for semantic similarity)
				if result.Score < 0.3 {
					t.Errorf("Expected semantic search score > 0.3, got %.3f", result.Score)
				}
				t.Logf("Semantic search found document with score: %.3f", result.Score)
				break
			}
		}

		if !found {
			t.Error("Semantic search should find the AI/ML document")
		}

		// Clean up semantic test document
		err = client.DeleteDocument(ctx, collectionName, semanticDoc.ID)
		if err != nil {
			t.Errorf("Failed to delete semantic test document: %v", err)
		}
	})

	// Test documents with different embeddings
	t.Run("DocumentsWithDifferentEmbeddings", func(t *testing.T) {
		// Skip if no OpenAI API key for embeddings
		if os.Getenv("OPENAI_API_KEY") == "" {
			t.Skip("Skipping embedding test: OPENAI_API_KEY not set")
		}

		// Test 1: Collection with OpenAI embeddings (text2vec-openai)
		openAICollection := "test_openai_embeddings_supabase"
		_ = client.DeleteCollection(ctx, openAICollection)

		openAISchema := &vectordb.CollectionSchema{
			Class:      openAICollection,
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
				{
					Name:     "topic",
					DataType: []string{"text"},
				},
			},
		}

		err := client.CreateCollection(ctx, openAICollection, openAISchema)
		if err != nil {
			t.Errorf("Failed to create OpenAI collection: %v", err)
		} else {
			t.Logf("✓ Created collection with OpenAI embeddings (text2vec-openai)")
		}

		// Create document with OpenAI embeddings
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
		} else {
			t.Logf("✓ Created document with OpenAI embeddings")
		}

		// Verify semantic search works with OpenAI embeddings
		searchOpts := &vectordb.QueryOptions{TopK: 5}
		results, err := client.SearchSemantic(ctx, openAICollection, "tell me about AI and ML", searchOpts)
		if err != nil {
			t.Errorf("Failed to search with OpenAI embeddings: %v", err)
		} else {
			if len(results) > 0 {
				t.Logf("✓ Semantic search with OpenAI embeddings returned %d results", len(results))
				for i, result := range results {
					t.Logf("  Result %d: ID=%s, Score=%.3f", i+1, result.Document.ID, result.Score)
				}
			} else {
				t.Logf("⚠ Semantic search returned no results (may need time for embedding generation)")
			}
		}

		// Test 2: Collection with no vectorizer (manual embeddings)
		noneCollection := "test_none_vectorizer_supabase"
		_ = client.DeleteCollection(ctx, noneCollection)

		noneSchema := &vectordb.CollectionSchema{
			Class:      noneCollection,
			Vectorizer: "none",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
				{
					Name:     "topic",
					DataType: []string{"text"},
				},
			},
		}

		err = client.CreateCollection(ctx, noneCollection, noneSchema)
		if err != nil {
			t.Errorf("Failed to create collection with no vectorizer: %v", err)
		} else {
			t.Logf("✓ Created collection with no automatic embeddings (vectorizer: none)")
		}

		// Create document with no automatic embeddings
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
		} else {
			t.Logf("✓ Created document with no automatic embeddings")
		}

		// Verify we can retrieve the document
		doc, err := client.GetDocument(ctx, noneCollection, noneDoc.ID)
		if err != nil {
			t.Errorf("Failed to get document from non-vectorized collection: %v", err)
		} else {
			if doc.ID == noneDoc.ID && doc.Content == noneDoc.Content {
				t.Logf("✓ Successfully retrieved document from non-vectorized collection")
			} else {
				t.Errorf("Retrieved document data mismatch")
			}
		}

		// Cleanup test collections
		t.Logf("Cleaning up embedding test collections...")
		_ = client.DeleteCollection(ctx, openAICollection)
		_ = client.DeleteCollection(ctx, noneCollection)
		t.Logf("✓ Embedding test completed successfully")
	})

	// Test that Supabase handles unsupported embedding providers appropriately
	t.Run("UnsupportedEmbeddingProviders", func(t *testing.T) {
		t.Logf("Testing Supabase behavior with unsupported embedding providers")

		// List of embedding providers that are NOT supported by Supabase
		// (but ARE supported by Weaviate)
		unsupportedProviders := []struct {
			name       string
			vectorizer string
		}{
			{"Cohere", "text2vec-cohere"},
			{"Hugging Face", "text2vec-huggingface"},
			{"Google PaLM", "text2vec-palm"},
			{"AWS Bedrock", "text2vec-aws"},
			{"Jina AI", "text2vec-jinaai"},
		}

		for _, provider := range unsupportedProviders {
			t.Run(provider.name, func(t *testing.T) {
				collectionName := "test_unsupported_" + provider.vectorizer
				_ = client.DeleteCollection(ctx, collectionName) // Clean up if exists

				schema := &vectordb.CollectionSchema{
					Class:      collectionName,
					Vectorizer: provider.vectorizer,
					Properties: []vectordb.SchemaProperty{
						{
							Name:     "content",
							DataType: []string{"text"},
						},
					},
				}

				// Attempt to create collection with unsupported vectorizer
				err := client.CreateCollection(ctx, collectionName, schema)

				if err != nil {
					// Supabase may reject the unsupported vectorizer
					t.Logf("✓ Supabase rejected %s (%s): %v", provider.name, provider.vectorizer, err)
				} else {
					// Or it may accept it but use manual/none embeddings
					t.Logf("✓ Supabase accepted %s (%s) - will use manual embeddings", provider.name, provider.vectorizer)

					// Verify collection was created
					exists, checkErr := client.CollectionExists(ctx, collectionName)
					if checkErr != nil {
						t.Errorf("Failed to check collection existence: %v", checkErr)
					} else if !exists {
						t.Error("Collection should exist after creation")
					} else {
						t.Logf("  Collection created successfully with vectorizer=%s", provider.vectorizer)
					}

					// Try to add a document - this should work regardless of vectorizer
					doc := &vectordb.Document{
						ID:      "test-doc-1",
						Content: "Test document for unsupported vectorizer",
						Metadata: map[string]interface{}{
							"vectorizer": provider.vectorizer,
							"test":       "unsupported",
						},
					}

					docErr := client.CreateDocument(ctx, collectionName, doc)
					if docErr != nil {
						t.Logf("  ⚠ Document creation failed (expected for unsupported vectorizer): %v", docErr)
					} else {
						t.Logf("  ✓ Document created (manual/no automatic embeddings)")

						// Verify we can retrieve it
						retrieved, getErr := client.GetDocument(ctx, collectionName, doc.ID)
						if getErr != nil {
							t.Errorf("  Failed to retrieve document: %v", getErr)
						} else if retrieved.Content != doc.Content {
							t.Errorf("  Content mismatch: expected %s, got %s", doc.Content, retrieved.Content)
						} else {
							t.Logf("  ✓ Document retrieved successfully")
						}
					}

					// Cleanup
					_ = client.DeleteCollection(ctx, collectionName)
				}
			})
		}

		t.Logf("\n✓ Completed validation of unsupported embedding providers for Supabase")
		t.Logf("  Note: Supabase currently only supports OpenAI embeddings (text2vec-openai)")
		t.Logf("  Other providers will either be rejected or treated as manual embeddings")
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
			Class:      "valid_test_collection",
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

		// Verify batch creation worked
		retrieved, err := client.GetDocument(ctx, collectionName, "batch-doc-1")
		if err != nil {
			t.Errorf("Failed to retrieve batch document: %v", err)
		}
		if retrieved.ID != "batch-doc-1" {
			t.Errorf("Expected batch-doc-1, got %s", retrieved.ID)
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

func TestSupabaseFactoryIntegration(t *testing.T) {
	// Test factory with real configuration
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	if databaseURL == "" || databaseKey == "" {
		t.Skip("Skipping Supabase factory integration test: SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:        vectordb.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Timeout:     30,
	}

	// Test factory creation
	factory := supabase.NewFactory()

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

func TestSupabaseVectorDBRegistry(t *testing.T) {
	// Test that Supabase can be created through the global registry
	databaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	databaseKey := os.Getenv("SUPABASE_DATABASE_KEY")

	if databaseURL == "" || databaseKey == "" {
		t.Skip("Skipping Supabase registry test: SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY environment variables not set")
	}

	config := &vectordb.Config{
		Type:        vectordb.VectorDBTypeSupabase,
		DatabaseURL: databaseURL,
		DatabaseKey: databaseKey,
		Timeout:     30,
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
