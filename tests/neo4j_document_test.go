// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
)

func TestNeo4jDocumentOperations(t *testing.T) {
	// Create client
	password := os.Getenv("NEO4J_PASSWORD")
	if password == "" {
		password = "weaveneo4j" // Default password
	}

	config := &neo4j.Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: password,
		Database: "neo4j",
	}

	client, err := neo4j.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Neo4j client: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()
	testCollection := "TestDocs"

	// Clean up
	client.DeleteCollection(ctx, testCollection, true)

	// Create collection
	err = client.CreateCollection(ctx, testCollection, 3, "cosine")
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}
	defer client.DeleteCollection(ctx, testCollection, true)

	// Test: Create Document
	t.Run("CreateDocument", func(t *testing.T) {
		doc := &neo4j.Document{
			ID:      "doc1",
			Content: "This is a test document",
			Vector:  []float32{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"author": "test",
				"tags":   "sample",
			},
		}

		err := client.CreateDocument(ctx, testCollection, doc)
		if err != nil {
			t.Fatalf("Failed to create document: %v", err)
		}
		t.Log("✓ Document created")
	})

	// Test: Get Document
	t.Run("GetDocument", func(t *testing.T) {
		doc, err := client.GetDocument(ctx, testCollection, "doc1")
		if err != nil {
			t.Fatalf("Failed to get document: %v", err)
		}
		if doc.ID != "doc1" {
			t.Fatalf("Expected ID doc1, got %s", doc.ID)
		}
		if doc.Content != "This is a test document" {
			t.Fatalf("Expected content 'This is a test document', got '%s'", doc.Content)
		}
		if len(doc.Vector) != 3 {
			t.Fatalf("Expected vector length 3, got %d", len(doc.Vector))
		}
		t.Logf("✓ Document retrieved: %+v", doc)
	})

	// Test: Update Document
	t.Run("UpdateDocument", func(t *testing.T) {
		doc := &neo4j.Document{
			ID:      "doc1",
			Content: "Updated content",
			Vector:  []float32{0.4, 0.5, 0.6},
			Metadata: map[string]interface{}{
				"author":  "test_updated",
				"version": 2,
			},
		}

		err := client.UpdateDocument(ctx, testCollection, doc)
		if err != nil {
			t.Fatalf("Failed to update document: %v", err)
		}

		// Verify update
		updated, err := client.GetDocument(ctx, testCollection, "doc1")
		if err != nil {
			t.Fatalf("Failed to get updated document: %v", err)
		}
		if updated.Content != "Updated content" {
			t.Fatalf("Expected updated content, got '%s'", updated.Content)
		}
		t.Log("✓ Document updated")
	})

	// Test: Batch Create Documents
	t.Run("BatchCreateDocuments", func(t *testing.T) {
		docs := []*neo4j.Document{
			{
				ID:      "doc2",
				Content: "Second document",
				Vector:  []float32{0.7, 0.8, 0.9},
				Metadata: map[string]interface{}{
					"batch": 1,
				},
			},
			{
				ID:      "doc3",
				Content: "Third document",
				Vector:  []float32{0.11, 0.12, 0.13},
				Metadata: map[string]interface{}{
					"batch": 1,
				},
			},
		}

		err := client.BatchCreateDocuments(ctx, testCollection, docs)
		if err != nil {
			t.Fatalf("Failed to batch create documents: %v", err)
		}
		t.Log("✓ Batch documents created")
	})

	// Test: List Documents
	t.Run("ListDocuments", func(t *testing.T) {
		docs, err := client.ListDocuments(ctx, testCollection, 10)
		if err != nil {
			t.Fatalf("Failed to list documents: %v", err)
		}
		if len(docs) != 3 {
			t.Fatalf("Expected 3 documents, got %d", len(docs))
		}
		t.Logf("✓ Listed %d documents", len(docs))
	})

	// Test: Delete Document
	t.Run("DeleteDocument", func(t *testing.T) {
		err := client.DeleteDocument(ctx, testCollection, "doc2")
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = client.GetDocument(ctx, testCollection, "doc2")
		if err == nil {
			t.Fatal("Document should not exist after deletion")
		}
		t.Log("✓ Document deleted")
	})

	// Test: Delete Documents (batch)
	t.Run("DeleteDocuments", func(t *testing.T) {
		err := client.DeleteDocuments(ctx, testCollection, []string{"doc1", "doc3"})
		if err != nil {
			t.Fatalf("Failed to delete documents: %v", err)
		}

		// Verify all deleted
		count, err := client.GetCollectionCount(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected count 0 after delete, got %d", count)
		}
		t.Log("✓ Documents deleted")
	})

	// Test: Delete All Documents
	t.Run("DeleteAllDocuments", func(t *testing.T) {
		// Create some documents first
		client.CreateDocument(ctx, testCollection, &neo4j.Document{
			ID:      "temp1",
			Content: "Temp doc",
			Vector:  []float32{0.1, 0.2, 0.3},
		})

		err := client.DeleteAllDocuments(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to delete all documents: %v", err)
		}

		count, err := client.GetCollectionCount(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected count 0, got %d", count)
		}
		t.Log("✓ All documents deleted")
	})
}
