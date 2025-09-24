package tests

import (
	"context"
	"fmt"
	"testing"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
)

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