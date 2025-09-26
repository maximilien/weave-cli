package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestMockClient(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)

	// Test health check
	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test listing collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		t.Errorf("Failed to list collections: %v", err)
	}

	if len(collections) != 1 {
		t.Errorf("Expected 1 collection, got %d", len(collections))
	}

	if collections[0] != "test" {
		t.Errorf("Expected collection 'test', got %s", collections[0])
	}
}

func TestMockClientDocumentOperations(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Add a document
	doc := mock.Document{
		ID:      "test-doc-1",
		Content: "This is a test document",
		Metadata: map[string]interface{}{
			"source": "test",
			"type":   "text",
		},
	}

	if err := client.AddDocument(ctx, "test", doc); err != nil {
		t.Errorf("Failed to add document: %v", err)
	}

	// List documents
	documents, err := client.ListDocuments(ctx, "test", 10)
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}

	if len(documents) != 1 {
		t.Errorf("Expected 1 document, got %d", len(documents))
	}

	if documents[0].ID != "test-doc-1" {
		t.Errorf("Expected document ID 'test-doc-1', got %s", documents[0].ID)
	}

	// Get specific document
	retrievedDoc, err := client.GetDocument(ctx, "test", "test-doc-1")
	if err != nil {
		t.Errorf("Failed to get document: %v", err)
	}

	if retrievedDoc.ID != "test-doc-1" {
		t.Errorf("Expected document ID 'test-doc-1', got %s", retrievedDoc.ID)
	}

	// Delete document
	if err := client.DeleteDocument(ctx, "test", "test-doc-1"); err != nil {
		t.Errorf("Failed to delete document: %v", err)
	}

	// Verify document is deleted
	documents, err = client.ListDocuments(ctx, "test", 10)
	if err != nil {
		t.Errorf("Failed to list documents after deletion: %v", err)
	}

	if len(documents) != 0 {
		t.Errorf("Expected 0 documents after deletion, got %d", len(documents))
	}
}

func TestMockClientDeleteDocumentsByMetadata(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Add test documents with metadata
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "Content 1",
			Metadata: map[string]interface{}{
				"filename": "file1.png",
				"type":     "image",
				"size":     "1024",
			},
		},
		{
			ID:      "doc2",
			Content: "Content 2",
			Metadata: map[string]interface{}{
				"filename": "file2.jpg",
				"type":     "image",
				"size":     "2048",
			},
		},
		{
			ID:      "doc3",
			Content: "Content 3",
			Metadata: map[string]interface{}{
				"filename": "file1.png",
				"type":     "document",
				"size":     "512",
			},
		},
		{
			ID:      "doc4",
			Content: "Content 4",
			Metadata: map[string]interface{}{
				"filename": "file3.pdf",
				"type":     "document",
				"size":     "1024",
			},
		},
	}

	// Add documents to collection
	for _, doc := range testDocs {
		if err := client.AddDocument(ctx, "test", doc); err != nil {
			t.Errorf("Failed to add document: %v", err)
		}
	}

	// Test 1: Delete by single metadata filter
	deletedCount, err := client.DeleteDocumentsByMetadata(ctx, "test", []string{"filename=file1.png"})
	if err != nil {
		t.Errorf("Failed to delete documents by metadata: %v", err)
	}
	if deletedCount != 2 {
		t.Errorf("Expected 2 documents deleted, got %d", deletedCount)
	}

	// Verify remaining documents
	documents, err := client.ListDocuments(ctx, "test", 10)
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}
	if len(documents) != 2 {
		t.Errorf("Expected 2 documents remaining, got %d", len(documents))
	}

	// Test 2: Delete by multiple metadata filters
	deletedCount, err = client.DeleteDocumentsByMetadata(ctx, "test", []string{"type=document", "size=1024"})
	if err != nil {
		t.Errorf("Failed to delete documents by metadata: %v", err)
	}
	if deletedCount != 1 {
		t.Errorf("Expected 1 document deleted, got %d", deletedCount)
	}

	// Verify remaining documents
	documents, err = client.ListDocuments(ctx, "test", 10)
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}
	if len(documents) != 1 {
		t.Errorf("Expected 1 document remaining, got %d", len(documents))
	}

	// Test 3: Delete with no matching documents
	deletedCount, err = client.DeleteDocumentsByMetadata(ctx, "test", []string{"filename=nonexistent.png"})
	if err != nil {
		t.Errorf("Failed to delete documents by metadata: %v", err)
	}
	if deletedCount != 0 {
		t.Errorf("Expected 0 documents deleted, got %d", deletedCount)
	}

	// Test 4: Invalid metadata filter format
	_, err = client.DeleteDocumentsByMetadata(ctx, "test", []string{"invalid-filter"})
	if err == nil {
		t.Error("Expected error for invalid metadata filter format")
	}

	// Test 5: Non-existent collection
	_, err = client.DeleteDocumentsByMetadata(ctx, "nonexistent", []string{"filename=test.png"})
	if err == nil {
		t.Error("Expected error for non-existent collection")
	}
}

func TestMockClientVirtualDocumentDeletion(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "VirtualTestCollection", Type: "text", Description: "Test collection for virtual deletion"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Clear the collection first to avoid interference
	client.DeleteCollection(ctx, "VirtualTestCollection")

	// Add test documents with chunked metadata structure
	testDocs := []mock.Document{
		{
			ID:      "ragme-io-chunk-1",
			Content: "This is the first chunk of ragme-io.pdf",
			Metadata: map[string]interface{}{
				"metadata": `{"original_filename": "ragme-io.pdf", "is_chunked": true, "chunk_index": 0, "total_chunks": 3}`,
			},
		},
		{
			ID:      "ragme-io-chunk-2",
			Content: "This is the second chunk of ragme-io.pdf",
			Metadata: map[string]interface{}{
				"metadata": `{"original_filename": "ragme-io.pdf", "is_chunked": true, "chunk_index": 1, "total_chunks": 3}`,
			},
		},
		{
			ID:      "ragme-io-chunk-3",
			Content: "This is the third chunk of ragme-io.pdf",
			Metadata: map[string]interface{}{
				"metadata": `{"original_filename": "ragme-io.pdf", "is_chunked": true, "chunk_index": 2, "total_chunks": 3}`,
			},
		},
		{
			ID:      "other-doc-chunk-1",
			Content: "This is a chunk from a different document",
			Metadata: map[string]interface{}{
				"metadata": `{"original_filename": "other-doc.pdf", "is_chunked": true, "chunk_index": 0, "total_chunks": 2}`,
			},
		},
	}

	// Add documents to collection
	for _, doc := range testDocs {
		if err := client.AddDocument(ctx, "VirtualTestCollection", doc); err != nil {
			t.Errorf("Failed to add document: %v", err)
		}
	}

	// Verify initial state
	documents, err := client.ListDocuments(ctx, "VirtualTestCollection", 10)
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}
	if len(documents) != 4 {
		t.Errorf("Expected 4 documents initially, got %d", len(documents))
	}

	// Test virtual deletion by original filename
	// This should delete all chunks associated with "ragme-io.pdf"
	deletedCount := 0
	expectedRagmeDocs := []string{"ragme-io-chunk-1", "ragme-io-chunk-2", "ragme-io-chunk-3"}

	// First, collect all documents that should be deleted
	var docsToDelete []string
	for _, doc := range documents {
		if metadata, ok := doc.Metadata["metadata"]; ok {
			if metadataStr, ok := metadata.(string); ok {
				// Parse the JSON metadata to extract original filename
				var metadataObj map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
					if docOriginalFilename, ok := metadataObj["original_filename"].(string); ok {
						if docOriginalFilename == "ragme-io.pdf" {
							docsToDelete = append(docsToDelete, doc.ID)
						}
					}
				}
			}
		}
	}

	// Then delete them
	for _, docID := range docsToDelete {
		if err := client.DeleteDocument(ctx, "VirtualTestCollection", docID); err != nil {
			// Only log error if the document was expected to exist
			if contains(expectedRagmeDocs, docID) {
				t.Errorf("Failed to delete document %s: %v", docID, err)
			}
		} else {
			deletedCount++
		}
	}

	// Should have deleted 3 documents (3 chunks)
	if deletedCount != 3 {
		t.Errorf("Expected 3 documents deleted for ragme-io.pdf, got %d", deletedCount)
	}

	// Verify remaining documents
	documents, err = client.ListDocuments(ctx, "VirtualTestCollection", 10)
	if err != nil {
		t.Errorf("Failed to list documents after virtual deletion: %v", err)
	}
	if len(documents) != 1 {
		t.Errorf("Expected 1 document remaining after virtual deletion, got %d", len(documents))
	}

	// Verify the remaining document is from the other file
	if documents[0].ID != "other-doc-chunk-1" {
		t.Errorf("Expected remaining document to be 'other-doc-chunk-1', got %s", documents[0].ID)
	}
}

func TestMockClientCollectionOperations(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Add some documents
	for i := 0; i < 3; i++ {
		doc := mock.Document{
			ID:      fmt.Sprintf("test-doc-%d", i),
			Content: fmt.Sprintf("This is test document %d", i),
			Metadata: map[string]interface{}{
				"index": i,
			},
		}

		if err := client.AddDocument(ctx, "test", doc); err != nil {
			t.Errorf("Failed to add document %d: %v", i, err)
		}
	}

	// Test collection stats
	stats, err := client.GetCollectionStats(ctx, "test")
	if err != nil {
		t.Errorf("Failed to get collection stats: %v", err)
	}

	if stats["document_count"] != 3 {
		t.Errorf("Expected 3 documents, got %v", stats["document_count"])
	}

	// Delete collection
	if err := client.DeleteCollection(ctx, "test"); err != nil {
		t.Errorf("Failed to delete collection: %v", err)
	}

	// Verify collection is empty
	documents, err := client.ListDocuments(ctx, "test", 10)
	if err != nil {
		t.Errorf("Failed to list documents after collection deletion: %v", err)
	}

	if len(documents) != 0 {
		t.Errorf("Expected 0 documents after collection deletion, got %d", len(documents))
	}
}

func TestMockClientGetDocumentsByMetadata(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Add test documents with metadata
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "Test document 1",
			Metadata: map[string]interface{}{
				"filename": "file1.png",
				"type":     "image",
			},
		},
		{
			ID:      "doc2",
			Content: "Test document 2",
			Metadata: map[string]interface{}{
				"filename": "file2.jpg",
				"type":     "image",
			},
		},
		{
			ID:      "doc3",
			Content: "Test document 3",
			Metadata: map[string]interface{}{
				"filename": "file1.png",
				"type":     "document",
			},
		},
	}

	// Add documents to collection
	for _, doc := range testDocs {
		if err := client.AddDocument(ctx, "test", doc); err != nil {
			t.Errorf("Failed to add document: %v", err)
		}
	}

	t.Run("Get documents by filename", func(t *testing.T) {
		documents, err := client.GetDocumentsByMetadata(ctx, "test", []string{"filename=file1.png"})
		if err != nil {
			t.Errorf("Failed to get documents by metadata: %v", err)
		}

		if len(documents) != 2 {
			t.Errorf("Expected 2 documents, got %d", len(documents))
		}

		// Check that we got the right documents
		ids := make(map[string]bool)
		for _, doc := range documents {
			ids[doc.ID] = true
		}

		if !ids["doc1"] || !ids["doc3"] {
			t.Errorf("Expected documents doc1 and doc3, got: %v", ids)
		}
	})

	t.Run("Get documents by type", func(t *testing.T) {
		documents, err := client.GetDocumentsByMetadata(ctx, "test", []string{"type=image"})
		if err != nil {
			t.Errorf("Failed to get documents by metadata: %v", err)
		}

		if len(documents) != 2 {
			t.Errorf("Expected 2 documents, got %d", len(documents))
		}

		// Check that we got the right documents
		ids := make(map[string]bool)
		for _, doc := range documents {
			ids[doc.ID] = true
		}

		if !ids["doc1"] || !ids["doc2"] {
			t.Errorf("Expected documents doc1 and doc2, got: %v", ids)
		}
	})

	t.Run("Get documents by multiple filters", func(t *testing.T) {
		documents, err := client.GetDocumentsByMetadata(ctx, "test", []string{"filename=file1.png", "type=image"})
		if err != nil {
			t.Errorf("Failed to get documents by metadata: %v", err)
		}

		if len(documents) != 1 {
			t.Errorf("Expected 1 document, got %d", len(documents))
		}

		if documents[0].ID != "doc1" {
			t.Errorf("Expected document doc1, got %s", documents[0].ID)
		}
	})

	t.Run("Get documents with no matches", func(t *testing.T) {
		documents, err := client.GetDocumentsByMetadata(ctx, "test", []string{"filename=nonexistent.png"})
		if err != nil {
			t.Errorf("Failed to get documents by metadata: %v", err)
		}

		if len(documents) != 0 {
			t.Errorf("Expected 0 documents, got %d", len(documents))
		}
	})

	t.Run("Invalid metadata filter format", func(t *testing.T) {
		_, err := client.GetDocumentsByMetadata(ctx, "test", []string{"invalid-filter"})
		if err == nil {
			t.Error("Expected error for invalid filter format")
		}
	})

	t.Run("Non-existent collection", func(t *testing.T) {
		_, err := client.GetDocumentsByMetadata(ctx, "nonexistent", []string{"filename=test.png"})
		if err == nil {
			t.Error("Expected error for non-existent collection")
		}
	})
}

// TestMultipleCollectionCreationIntegration tests multiple collection creation with mock client
func TestMultipleCollectionCreationIntegration(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections:        []config.MockCollection{},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	t.Run("Create Multiple Collections", func(t *testing.T) {
		collectionNames := []string{"MultiCol1", "MultiCol2", "MultiCol3"}
		
		// Create multiple collections
		for _, name := range collectionNames {
			err := client.CreateCollection(ctx, name, "text-embedding-3-small", []weaviate.FieldDefinition{})
			if err != nil {
				t.Errorf("Failed to create collection %s: %v", name, err)
			}
		}

		// Verify all collections were created
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		for _, name := range collectionNames {
			if !contains(collections, name) {
				t.Errorf("Collection %s was not created", name)
			}
		}
	})

	t.Run("Create Multiple Collections with Custom Fields", func(t *testing.T) {
		collectionNames := []string{"CustomCol1", "CustomCol2"}
		customFields := []weaviate.FieldDefinition{
			{Name: "title", Type: "text"},
			{Name: "author", Type: "text"},
		}
		
		// Create multiple collections with custom fields
		for _, name := range collectionNames {
			err := client.CreateCollection(ctx, name, "text-embedding-3-large", customFields)
			if err != nil {
				t.Errorf("Failed to create collection %s: %v", name, err)
			}
		}

		// Verify collections were created
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		for _, name := range collectionNames {
			if !contains(collections, name) {
				t.Errorf("Collection %s was not created", name)
			}
		}
	})

	t.Run("Create Multiple Collections with Error Handling", func(t *testing.T) {
		// Try to create a collection that already exists
		err := client.CreateCollection(ctx, "MultiCol1", "text-embedding-3-small", []weaviate.FieldDefinition{})
		if err == nil {
			t.Error("Expected error when creating duplicate collection")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})
}

// TestCollectionSchemaDeletionIntegration tests collection schema deletion with mock client
func TestCollectionSchemaDeletionIntegration(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections:        []config.MockCollection{},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	t.Run("Delete Collection Schema", func(t *testing.T) {
		collectionName := "SchemaTestCollection"
		
		// Create a collection first
		err := client.CreateCollection(ctx, collectionName, "text-embedding-3-small", []weaviate.FieldDefinition{})
		if err != nil {
			t.Errorf("Failed to create collection: %v", err)
		}

		// Verify collection exists
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		if !contains(collections, collectionName) {
			t.Error("Collection was not created")
		}

		// Delete the collection schema
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to delete collection: %v", err)
		}

		// Verify collection is gone
		collections, err = client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections after deletion: %v", err)
		}

		if contains(collections, collectionName) {
			t.Error("Collection should have been deleted")
		}
	})

	t.Run("Delete Non-existent Collection Schema", func(t *testing.T) {
		// Try to delete a collection that doesn't exist
		err := client.DeleteCollection(ctx, "NonExistentCollection")
		if err == nil {
			t.Error("Expected error when deleting non-existent collection")
		}

		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %v", err)
		}
	})

	t.Run("Delete Collection with Documents", func(t *testing.T) {
		collectionName := "CollectionWithDocs"
		
		// Create a collection
		err := client.CreateCollection(ctx, collectionName, "text-embedding-3-small", []weaviate.FieldDefinition{})
		if err != nil {
			t.Errorf("Failed to create collection: %v", err)
		}

		// Add a document
		document := mock.Document{
			ID:      "test-doc-1",
			Content: "Test content",
			Metadata: map[string]interface{}{
				"title": "Test Document",
			},
		}

		err = client.CreateDocument(ctx, collectionName, document)
		if err != nil {
			t.Errorf("Failed to create document: %v", err)
		}

		// Verify document exists
		documents, err := client.ListDocuments(ctx, collectionName, 10)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}

		if len(documents) != 1 {
			t.Errorf("Expected 1 document, got %d", len(documents))
		}

		// Delete the collection (should remove both schema and documents)
		err = client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("Failed to delete collection: %v", err)
		}

		// Verify collection is gone
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections after deletion: %v", err)
		}

		if contains(collections, collectionName) {
			t.Error("Collection should have been deleted")
		}
	})
}
