// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// TestQueryE2E tests the end-to-end query functionality
func TestQueryE2E(t *testing.T) {
	// Setup mock configuration
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

	// Create test documents
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "This document is about machine learning algorithms and neural networks",
			Metadata: map[string]interface{}{
				"filename": "ml_guide.txt",
				"type":     "text",
			},
		},
		{
			ID:      "doc2",
			Content: "Artificial intelligence and deep learning concepts for beginners",
			Metadata: map[string]interface{}{
				"filename": "ai_basics.txt",
				"type":     "text",
			},
		},
		{
			ID:      "doc3",
			Content: "Data preprocessing and feature engineering techniques",
			Metadata: map[string]interface{}{
				"filename": "data_prep.txt",
				"type":     "text",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("QueryMachineLearning", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'machine learning' query")
		}

		// Check that we got the machine learning document
		found := false
		for _, result := range results {
			if result.ID == "doc1" {
				found = true
				if result.Score < 0.9 || result.Score > 1.0 {
					t.Errorf("Expected score between 0.9 and 1.0 for exact match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc1 (machine learning document) in results")
		}
	})

	t.Run("QueryArtificialIntelligence", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test", "artificial intelligence", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'artificial intelligence' query")
		}

		// Check that we got the AI document
		found := false
		for _, result := range results {
			if result.ID == "doc2" {
				found = true
				if result.Score < 0.9 || result.Score > 1.0 {
					t.Errorf("Expected score between 0.9 and 1.0 for exact match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc2 (AI document) in results")
		}
	})

	t.Run("QueryDataPreprocessing", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test", "data preprocessing", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'data preprocessing' query")
		}

		// Check that we got the data preprocessing document
		found := false
		for _, result := range results {
			if result.ID == "doc3" {
				found = true
				if result.Score < 0.9 || result.Score > 1.0 {
					t.Errorf("Expected score between 0.9 and 1.0 for exact match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc3 (data preprocessing document) in results")
		}
	})

	t.Run("QueryTopKLimit", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 2}
		results, err := client.Query(ctx, "test", "learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) > 2 {
			t.Errorf("Expected at most 2 results, got %d", len(results))
		}
	})

	t.Run("QueryNoResults", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test", "nonexistent content", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) != 0 {
			t.Error("Expected no results for 'nonexistent content' query")
		}
	})

	t.Run("QueryCaseInsensitive", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test", "MACHINE LEARNING", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for case insensitive 'MACHINE LEARNING' query")
		}

		// Check that we got the machine learning document
		found := false
		for _, result := range results {
			if result.ID == "doc1" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find doc1 (machine learning document) in case insensitive results")
		}
	})
}

// TestQueryMetadataE2E tests the end-to-end query functionality with metadata search
func TestQueryMetadataE2E(t *testing.T) {
	// Setup mock configuration
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

	// Create test documents with different content and metadata
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "This document is about machine learning algorithms and neural networks",
			Metadata: map[string]interface{}{
				"filename": "ml_guide.txt",
				"type":     "text",
				"url":      "https://example.com/ml-guide",
			},
		},
		{
			ID:      "doc2",
			Content: "Artificial intelligence and deep learning concepts for beginners",
			Metadata: map[string]interface{}{
				"filename": "ai_basics.txt",
				"type":     "text",
				"url":      "https://example.com/ai-basics",
			},
		},
		{
			ID:      "doc3",
			Content: "Data preprocessing and feature engineering techniques",
			Metadata: map[string]interface{}{
				"filename": "data_prep.txt",
				"type":     "text",
				"url":      "https://maximilien.org/data-prep",
			},
		},
		{
			ID:      "doc4",
			Content: "General programming concepts and best practices",
			Metadata: map[string]interface{}{
				"filename": "programming.txt",
				"type":     "text",
				"url":      "https://github.com/maximilien/programming-guide",
			},
		},
		{
			ID:      "doc5",
			Content: "Web development frameworks and tools",
			Metadata: map[string]interface{}{
				"filename": "web_dev.txt",
				"type":     "text",
				"url":      "https://example.com/web-dev",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("QueryWithoutMetadataSearch", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: false}
		results, err := client.Query(ctx, "test", "maximilien", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		// Should not find doc3 or doc4 (maximilien in metadata only)
		if len(results) > 0 {
			t.Error("Expected no results for 'maximilien' query without metadata search")
		}
	})

	t.Run("QueryWithMetadataSearch", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "maximilien", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'maximilien' query with metadata search")
		}

		// Check that we found documents with maximilien in metadata
		found := false
		for _, result := range results {
			if result.ID == "doc3" || result.ID == "doc4" {
				found = true
				if result.Score < 0.1 || result.Score > 0.3 {
					t.Errorf("Expected score between 0.1 and 0.3 for metadata-only match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc3 or doc4 (maximilien in metadata) in results")
		}
	})

	t.Run("QuerySpecificURL", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "maximilien.org", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'maximilien.org' query with metadata search")
		}

		// Check that we found doc3 specifically
		found := false
		for _, result := range results {
			if result.ID == "doc3" {
				found = true
				if result.Score < 0.1 || result.Score > 0.3 {
					t.Errorf("Expected score between 0.1 and 0.3 for metadata-only match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc3 (maximilien.org in metadata) in results")
		}
	})

	t.Run("QueryGitHubURL", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "github.com", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'github.com' query with metadata search")
		}

		// Check that we found doc4 specifically
		found := false
		for _, result := range results {
			if result.ID == "doc4" {
				found = true
				if result.Score < 0.1 || result.Score > 0.3 {
					t.Errorf("Expected score between 0.1 and 0.3 for metadata-only match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc4 (github.com in metadata) in results")
		}
	})

	t.Run("QueryFilenameInMetadata", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "ml_guide", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'ml_guide' query with metadata search")
		}

		// Check that we found doc1 specifically
		found := false
		for _, result := range results {
			if result.ID == "doc1" {
				found = true
				if result.Score < 0.1 || result.Score > 0.3 {
					t.Errorf("Expected score between 0.1 and 0.3 for metadata-only match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc1 (ml_guide in metadata) in results")
		}
	})

	t.Run("QueryContentAndMetadata", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'machine learning' query with metadata search")
		}

		// Should find doc1 (machine learning in content)
		found := false
		for _, result := range results {
			if result.ID == "doc1" {
				found = true
				if result.Score < 0.9 || result.Score > 1.0 {
					t.Errorf("Expected score between 0.9 and 1.0 for exact match, got %f", result.Score)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find doc1 (machine learning in content) in results")
		}
	})

	t.Run("QueryTopKLimitWithMetadata", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 2, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "example.com", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) > 2 {
			t.Errorf("Expected at most 2 results, got %d", len(results))
		}
	})

	t.Run("QueryNoResultsWithMetadata", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "nonexistent content", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) != 0 {
			t.Error("Expected no results for 'nonexistent content' query with metadata search")
		}
	})

	t.Run("QueryCaseInsensitiveMetadata", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, SearchMetadata: true}
		results, err := client.Query(ctx, "test", "MAXIMILIEN", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for case insensitive 'MAXIMILIEN' query with metadata search")
		}

		// Check that we found documents with maximilien in metadata
		found := false
		for _, result := range results {
			if result.ID == "doc3" || result.ID == "doc4" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find doc3 or doc4 (maximilien in metadata) in case insensitive results")
		}
	})
}