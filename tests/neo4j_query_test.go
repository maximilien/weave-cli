// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
)

func TestNeo4jVectorSearch(t *testing.T) {
	// Create client
	config := &neo4j.Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "testpassword",
		Database: "neo4j",
	}

	client, err := neo4j.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Neo4j client: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()
	testCollection := "TestSearchCollection"

	// Clean up and setup
	client.DeleteCollection(ctx, testCollection, true)

	err = client.CreateCollection(ctx, testCollection, 3, "cosine")
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}
	defer client.DeleteCollection(ctx, testCollection, true)

	// Create test documents with different vectors
	docs := []*neo4j.Document{
		{
			ID:      "doc1",
			Content: "First document about cats",
			Vector:  []float32{1.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"category": "animals",
				"type":     "cat",
			},
		},
		{
			ID:      "doc2",
			Content: "Second document about dogs",
			Vector:  []float32{0.0, 1.0, 0.0},
			Metadata: map[string]interface{}{
				"category": "animals",
				"type":     "dog",
			},
		},
		{
			ID:      "doc3",
			Content: "Third document about birds",
			Vector:  []float32{0.0, 0.0, 1.0},
			Metadata: map[string]interface{}{
				"category": "animals",
				"type":     "bird",
			},
		},
		{
			ID:      "doc4",
			Content: "Fourth document similar to cats",
			Vector:  []float32{0.9, 0.1, 0.0},
			Metadata: map[string]interface{}{
				"category": "animals",
				"type":     "cat",
			},
		},
	}

	err = client.BatchCreateDocuments(ctx, testCollection, docs)
	if err != nil {
		t.Fatalf("Failed to create test documents: %v", err)
	}

	// Test: Basic Vector Search
	t.Run("VectorSearch", func(t *testing.T) {
		// Search for vectors similar to [1.0, 0.0, 0.0] (should find doc1 and doc4)
		queryVector := []float32{1.0, 0.0, 0.0}
		results, err := client.VectorSearch(ctx, testCollection, queryVector, 2)
		if err != nil {
			t.Fatalf("Failed to perform vector search: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		// First result should be doc1 (exact match)
		if results[0].Document.ID != "doc1" {
			t.Errorf("Expected first result to be doc1, got %s", results[0].Document.ID)
		}

		// Second result should be doc4 (similar vector)
		if results[1].Document.ID != "doc4" {
			t.Errorf("Expected second result to be doc4, got %s", results[1].Document.ID)
		}

		// Scores should be in descending order
		if results[0].Score < results[1].Score {
			t.Errorf("Scores should be in descending order: %f < %f", results[0].Score, results[1].Score)
		}

		t.Logf("✓ Vector search returned %d results", len(results))
		for i, result := range results {
			t.Logf("  [%d] %s (score: %.4f): %s", i+1, result.Document.ID, result.Score, result.Document.Content)
		}
	})

	// Test: Vector Search with More Results
	t.Run("VectorSearchAllResults", func(t *testing.T) {
		queryVector := []float32{0.0, 1.0, 0.0}
		results, err := client.VectorSearch(ctx, testCollection, queryVector, 10)
		if err != nil {
			t.Fatalf("Failed to perform vector search: %v", err)
		}

		if len(results) != 4 {
			t.Fatalf("Expected 4 results (all docs), got %d", len(results))
		}

		// First result should be doc2 (exact match to query vector)
		if results[0].Document.ID != "doc2" {
			t.Errorf("Expected first result to be doc2, got %s", results[0].Document.ID)
		}

		t.Logf("✓ Found all %d documents", len(results))
	})

	// Test: Vector Search with Filter
	t.Run("VectorSearchWithFilter", func(t *testing.T) {
		queryVector := []float32{1.0, 0.0, 0.0}
		filter := map[string]interface{}{
			"type": "cat",
		}

		results, err := client.VectorSearchWithFilter(ctx, testCollection, queryVector, 10, filter)
		if err != nil {
			t.Fatalf("Failed to perform filtered vector search: %v", err)
		}

		// Should only return cat documents (doc1 and doc4)
		if len(results) != 2 {
			t.Fatalf("Expected 2 results (cat docs), got %d", len(results))
		}

		for _, result := range results {
			if result.Document.Metadata["type"] != "cat" {
				t.Errorf("Expected type=cat, got %v", result.Document.Metadata["type"])
			}
		}

		t.Logf("✓ Filtered search returned %d cat documents", len(results))
	})

	// Test: Vector Search with Multiple Filters
	t.Run("VectorSearchWithMultipleFilters", func(t *testing.T) {
		queryVector := []float32{0.5, 0.5, 0.0}
		filter := map[string]interface{}{
			"category": "animals",
			"type":     "dog",
		}

		results, err := client.VectorSearchWithFilter(ctx, testCollection, queryVector, 10, filter)
		if err != nil {
			t.Fatalf("Failed to perform filtered vector search: %v", err)
		}

		// Should only return dog documents (doc2)
		if len(results) != 1 {
			t.Fatalf("Expected 1 result (dog doc), got %d", len(results))
		}

		if results[0].Document.ID != "doc2" {
			t.Errorf("Expected doc2, got %s", results[0].Document.ID)
		}

		t.Logf("✓ Multi-filter search returned correct document")
	})
}
