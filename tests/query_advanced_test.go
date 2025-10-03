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

// TestQueryBM25Search tests BM25 keyword search functionality
func TestQueryBM25Search(t *testing.T) {
	// Setup mock configuration
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test_bm25", Type: "text", Description: "Test collection for BM25"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Create test documents with distinct keywords
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "Python programming language tutorial for beginners. Learn Python syntax and data structures.",
			Metadata: map[string]interface{}{
				"filename": "python_tutorial.txt",
				"type":     "text",
			},
		},
		{
			ID:      "doc2",
			Content: "JavaScript web development guide. Master JavaScript frameworks and libraries.",
			Metadata: map[string]interface{}{
				"filename": "javascript_guide.txt",
				"type":     "text",
			},
		},
		{
			ID:      "doc3",
			Content: "Go programming language best practices. Building scalable applications with Go.",
			Metadata: map[string]interface{}{
				"filename": "go_practices.txt",
				"type":     "text",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test_bm25", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("BM25SearchPython", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, UseBM25: true}
		results, err := client.Query(ctx, "test_bm25", "Python", options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'Python' BM25 query")
		}

		// Check that we got the Python document with highest score
		if results[0].ID != "doc1" {
			t.Errorf("Expected doc1 (Python document) to be first result, got %s", results[0].ID)
		}
	})

	t.Run("BM25SearchJavaScript", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, UseBM25: true}
		results, err := client.Query(ctx, "test_bm25", "JavaScript", options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'JavaScript' BM25 query")
		}

		// Check that we got the JavaScript document
		if results[0].ID != "doc2" {
			t.Errorf("Expected doc2 (JavaScript document) to be first result, got %s", results[0].ID)
		}
	})

	t.Run("BM25SearchGo", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5, UseBM25: true}
		results, err := client.Query(ctx, "test_bm25", "Go programming", options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for 'Go programming' BM25 query")
		}

		// Check that we got the Go document
		found := false
		for _, result := range results {
			if result.ID == "doc3" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find doc3 (Go document) in BM25 results")
		}
	})

	t.Run("BM25SearchTopK", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 1, UseBM25: true}
		results, err := client.Query(ctx, "test_bm25", "programming", options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		if len(results) > 1 {
			t.Errorf("Expected at most 1 result with TopK=1, got %d", len(results))
		}
	})

	t.Run("BM25VsSemanticSearch", func(t *testing.T) {
		// BM25 search for exact keyword
		bm25Options := weaviate.QueryOptions{TopK: 3, UseBM25: true}
		bm25Results, err := client.Query(ctx, "test_bm25", "Python", bm25Options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		// Semantic search for same keyword
		semanticOptions := weaviate.QueryOptions{TopK: 3, UseBM25: false}
		semanticResults, err := client.Query(ctx, "test_bm25", "Python", semanticOptions)
		if err != nil {
			t.Errorf("Semantic query failed: %v", err)
		}

		// Both should return results
		if len(bm25Results) == 0 {
			t.Error("Expected BM25 results for 'Python'")
		}
		if len(semanticResults) == 0 {
			t.Error("Expected semantic results for 'Python'")
		}

		// BM25 should find the exact keyword match first
		if len(bm25Results) > 0 && bm25Results[0].ID != "doc1" {
			t.Errorf("Expected BM25 to rank doc1 (Python) first, got %s", bm25Results[0].ID)
		}
	})
}

// TestQueryScoreCalculation tests that scores are properly calculated
func TestQueryScoreCalculation(t *testing.T) {
	// Setup mock configuration
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test_scores", Type: "text", Description: "Test collection for scores"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Create test documents with varying similarity
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "machine learning deep neural networks artificial intelligence",
			Metadata: map[string]interface{}{
				"filename": "ml_exact.txt",
			},
		},
		{
			ID:      "doc2",
			Content: "machine learning algorithms and data science techniques",
			Metadata: map[string]interface{}{
				"filename": "ml_related.txt",
			},
		},
		{
			ID:      "doc3",
			Content: "cooking recipes and culinary techniques for beginners",
			Metadata: map[string]interface{}{
				"filename": "cooking.txt",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test_scores", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("ScoresAreNotAllSame", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3}
		results, err := client.Query(ctx, "test_scores", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) < 2 {
			t.Error("Expected at least 2 results to compare scores")
		}

		// Check that not all scores are 1.0
		allScoresOne := true
		for _, result := range results {
			if result.Score != 1.0 {
				allScoresOne = false
				break
			}
		}

		if allScoresOne {
			t.Error("All scores are 1.0, expected variation in scores")
		}
	})

	t.Run("ScoresAreInValidRange", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3}
		results, err := client.Query(ctx, "test_scores", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		for _, result := range results {
			if result.Score < 0.0 || result.Score > 1.0 {
				t.Errorf("Score %f is outside valid range [0.0, 1.0]", result.Score)
			}
		}
	})

	t.Run("ScoresDecreaseWithRelevance", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3}
		results, err := client.Query(ctx, "test_scores", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) < 2 {
			t.Error("Expected at least 2 results to compare scores")
		}

		// Scores should generally decrease (or stay same) as we go down the list
		// Allow for small variations due to semantic similarity
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score+0.1 {
				t.Errorf("Score at position %d (%f) is significantly higher than position %d (%f)",
					i, results[i].Score, i-1, results[i-1].Score)
			}
		}
	})

	t.Run("ExactMatchHasHighScore", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3}
		results, err := client.Query(ctx, "test_scores", "machine learning deep neural networks", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		}

		// First result should have high score (>0.5 for good semantic similarity)
		if results[0].Score < 0.5 {
			t.Errorf("Expected first result to have score > 0.5, got %f", results[0].Score)
		}
	})

	t.Run("UnrelatedContentHasLowerScore", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3}
		results, err := client.Query(ctx, "test_scores", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		// Find the cooking document (unrelated)
		for _, result := range results {
			if result.ID == "doc3" {
				if result.Score > 0.3 {
					t.Errorf("Expected unrelated content to have score < 0.3, got %f", result.Score)
				}
				break
			}
		}
	})

	t.Run("BM25ScoresAreValid", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 3, UseBM25: true}
		results, err := client.Query(ctx, "test_scores", "machine learning", options)
		if err != nil {
			t.Errorf("BM25 query failed: %v", err)
		}

		for _, result := range results {
			if result.Score < 0.0 {
				t.Errorf("BM25 score %f is negative", result.Score)
			}
		}
	})
}

// TestQueryWithDifferentVectorizers tests that query works with different vectorizer configurations
func TestQueryWithDifferentVectorizers(t *testing.T) {
	// Setup mock configuration
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test_vectorizer", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Create test documents
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "Artificial intelligence and machine learning concepts",
			Metadata: map[string]interface{}{
				"filename": "ai_concepts.txt",
			},
		},
		{
			ID:      "doc2",
			Content: "Natural language processing and text analysis",
			Metadata: map[string]interface{}{
				"filename": "nlp_guide.txt",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test_vectorizer", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("SemanticSearchWorksWithVectorizer", func(t *testing.T) {
		// This test verifies that semantic search uses the collection's vectorizer
		// The mock client simulates embeddings, but in real usage, this would use
		// the vectorizer configured in the collection schema (e.g., text2vec-openai)
		options := weaviate.QueryOptions{TopK: 2}
		results, err := client.Query(ctx, "test_vectorizer", "artificial intelligence", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result from semantic search")
		}

		// First result should be doc1 (contains "artificial intelligence")
		if results[0].ID != "doc1" {
			t.Errorf("Expected doc1 to be first result, got %s", results[0].ID)
		}
	})

	t.Run("QueryReturnsValidScoresWithVectorizer", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 2}
		results, err := client.Query(ctx, "test_vectorizer", "machine learning", options)
		if err != nil {
			t.Errorf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		}

		// Verify scores are in valid range
		for _, result := range results {
			if result.Score < 0.0 || result.Score > 1.0 {
				t.Errorf("Score %f is outside valid range [0.0, 1.0]", result.Score)
			}
		}
	})
}

// TestQueryEdgeCases tests edge cases in query functionality
func TestQueryEdgeCases(t *testing.T) {
	// Setup mock configuration
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test_edge", Type: "text", Description: "Test collection"},
		},
	}

	client := mock.NewClient(cfg)
	ctx := context.Background()

	// Create test documents
	testDocs := []mock.Document{
		{
			ID:      "doc1",
			Content: "Test document with content",
			Metadata: map[string]interface{}{
				"filename": "test.txt",
			},
		},
	}

	// Create documents
	for _, doc := range testDocs {
		err := client.CreateDocument(ctx, "test_edge", doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	t.Run("EmptyQueryString", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test_edge", "", options)
		if err != nil {
			t.Errorf("Query with empty string should not error: %v", err)
		}
		// Empty query may or may not return results, but shouldn't crash
		_ = results
	})

	t.Run("VeryLongQueryString", func(t *testing.T) {
		longQuery := "machine learning " + string(make([]byte, 10000))
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test_edge", longQuery, options)
		// Should handle long queries gracefully (may error or return results)
		_ = err
		_ = results
	})

	t.Run("SpecialCharactersInQuery", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 5}
		results, err := client.Query(ctx, "test_edge", "test \"quoted\" text", options)
		if err != nil {
			t.Errorf("Query with special characters should not error: %v", err)
		}
		_ = results
	})

	t.Run("ZeroTopK", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: 0}
		results, err := client.Query(ctx, "test_edge", "test", options)
		if err != nil {
			t.Errorf("Query with TopK=0 should not error (should use default): %v", err)
		}
		// Should return default number of results (5)
		if len(results) == 0 {
			t.Error("Expected default results when TopK=0")
		}
	})

	t.Run("NegativeTopK", func(t *testing.T) {
		options := weaviate.QueryOptions{TopK: -1}
		results, err := client.Query(ctx, "test_edge", "test", options)
		if err != nil {
			t.Errorf("Query with negative TopK should not error (should use default): %v", err)
		}
		// Should return default number of results (5)
		if len(results) == 0 {
			t.Error("Expected default results when TopK is negative")
		}
	})
}
