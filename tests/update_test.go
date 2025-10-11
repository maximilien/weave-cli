// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
)

func TestMockUpdateDocument(t *testing.T) {
	// Create mock client
	mockConfig := &config.MockConfig{
		Collections: []config.MockCollection{
			{Name: "TestCollection", Type: "text"},
		},
	}
	client := mock.NewClient(mockConfig)

	ctx := context.Background()
	collectionName := "TestCollection"

	// Create a test document
	testDoc := mock.Document{
		ID:      "test-doc-1",
		Content: "Initial test content",
		Metadata: map[string]interface{}{
			"filename": "test.txt",
			"type":     "test",
		},
	}

	// Add document directly to collection (mock client test data setup)
	err := client.CreateDocument(ctx, collectionName, testDoc)
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	// Get initial documents
	docs, err := client.ListDocuments(ctx, collectionName, 100)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	if len(docs) == 0 {
		t.Fatal("No initial documents found in test collection")
	}

	// Test updating content
	t.Run("UpdateContent", func(t *testing.T) {
		docID := docs[0].ID
		newContent := "Updated content for testing"

		err := client.UpdateDocument(ctx, collectionName, docID, newContent, nil)
		if err != nil {
			t.Fatalf("Failed to update document content: %v", err)
		}

		// Verify update
		updatedDocs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			t.Fatalf("Failed to list documents after update: %v", err)
		}

		found := false
		for _, doc := range updatedDocs {
			if doc.ID == docID {
				found = true
				if doc.Content != newContent {
					t.Errorf("Content not updated. Got: %s, Want: %s", doc.Content, newContent)
				}
				if doc.Metadata["updated_at"] == nil {
					t.Error("updated_at timestamp not set")
				}
				break
			}
		}

		if !found {
			t.Error("Document not found after update")
		}
	})

	// Test updating metadata
	t.Run("UpdateMetadata", func(t *testing.T) {
		docID := docs[0].ID
		metadata := map[string]interface{}{
			"version": "2.0",
			"author":  "Test Author",
		}

		err := client.UpdateDocument(ctx, collectionName, docID, "", metadata)
		if err != nil {
			t.Fatalf("Failed to update document metadata: %v", err)
		}

		// Verify update
		updatedDocs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			t.Fatalf("Failed to list documents after update: %v", err)
		}

		found := false
		for _, doc := range updatedDocs {
			if doc.ID == docID {
				found = true
				if doc.Metadata["version"] != "2.0" {
					t.Errorf("Version not updated. Got: %v, Want: 2.0", doc.Metadata["version"])
				}
				if doc.Metadata["author"] != "Test Author" {
					t.Errorf("Author not updated. Got: %v, Want: Test Author", doc.Metadata["author"])
				}
				break
			}
		}

		if !found {
			t.Error("Document not found after metadata update")
		}
	})

	// Test updating both content and metadata
	t.Run("UpdateContentAndMetadata", func(t *testing.T) {
		docID := docs[0].ID
		newContent := "Content and metadata update"
		metadata := map[string]interface{}{
			"status": "reviewed",
		}

		err := client.UpdateDocument(ctx, collectionName, docID, newContent, metadata)
		if err != nil {
			t.Fatalf("Failed to update document: %v", err)
		}

		// Verify update
		updatedDocs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			t.Fatalf("Failed to list documents after update: %v", err)
		}

		found := false
		for _, doc := range updatedDocs {
			if doc.ID == docID {
				found = true
				if doc.Content != newContent {
					t.Errorf("Content not updated. Got: %s, Want: %s", doc.Content, newContent)
				}
				if doc.Metadata["status"] != "reviewed" {
					t.Errorf("Status not updated. Got: %v, Want: reviewed", doc.Metadata["status"])
				}
				break
			}
		}

		if !found {
			t.Error("Document not found after combined update")
		}
	})

	// Test updating non-existent document
	t.Run("UpdateNonExistentDocument", func(t *testing.T) {
		err := client.UpdateDocument(ctx, collectionName, "non-existent-id", "content", nil)
		if err == nil {
			t.Error("Expected error when updating non-existent document, got nil")
		}
	})

	// Test updating in non-existent collection
	t.Run("UpdateInNonExistentCollection", func(t *testing.T) {
		err := client.UpdateDocument(ctx, "NonExistentCollection", docs[0].ID, "content", nil)
		if err == nil {
			t.Error("Expected error when updating in non-existent collection, got nil")
		}
	})
}
