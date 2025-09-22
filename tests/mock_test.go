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