// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
	"testing"
)

// TestIssue1_StderrProgressNotTreatedAsError tests that stderr progress output
// from sentence-transformers model loading doesn't cause query failures
// Regression test for AuctionsMax.ai Issue #1 (HIGH priority)
func TestIssue1_StderrProgressNotTreatedAsError(t *testing.T) {
	// Skip if sentence-transformers not available
	provider, err := NewSentenceTransformersProvider("all-MiniLM-L6-v2")
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	ctx := context.Background()
	if err := provider.IsAvailable(ctx); err != nil {
		t.Skipf("sentence-transformers not available: %v", err)
	}

	// This is the scenario that failed in production:
	// First query after server start loads the model, producing stderr progress
	// The fix ensures we don't treat stderr as error if exit code is 0
	texts := []string{
		"Nikon F2 camera",
		"Vintage Leica M3",
		"Canon AE-1 Program",
	}

	// Generate embeddings - should succeed even with stderr progress
	embeddings, err := provider.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Expected success despite stderr progress, got error: %v", err)
	}

	// Verify we got embeddings for all texts
	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// Verify embeddings have correct dimensions (384 for all-MiniLM-L6-v2)
	expectedDims := 384
	for i, emb := range embeddings {
		if len(emb) != expectedDims {
			t.Errorf("Embedding %d: expected %d dimensions, got %d", i, expectedDims, len(emb))
		}
	}

	// Verify embeddings are not all zeros (model actually worked)
	for i, emb := range embeddings {
		allZeros := true
		for _, val := range emb {
			if val != 0.0 {
				allZeros = false
				break
			}
		}
		if allZeros {
			t.Errorf("Embedding %d is all zeros - model may not have loaded correctly", i)
		}
	}
}

// TestIssue1_ActualErrorsStillReported tests that real errors are still caught
// despite the stderr progress fix
func TestIssue1_ActualErrorsStillReported(t *testing.T) {
	// Use invalid model name to trigger actual error
	provider, err := NewSentenceTransformersProvider("non-existent-model-xyz")
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	ctx := context.Background()

	// This should fail with actual error (not stderr progress)
	texts := []string{"test"}
	_, err = provider.GenerateEmbeddings(ctx, texts)

	// We expect an error for non-existent model
	if err == nil {
		t.Error("Expected error for non-existent model, got nil")
	}

	// Error message should indicate it's a real error, not progress
	if err != nil {
		errStr := err.Error()
		// Should contain actual error indicators
		// (The model will fail to load, triggering a traceback or error message)
		if errStr == "" {
			t.Error("Error message is empty")
		}
	}
}

// TestIssue1_EmptyTextHandling tests edge case with empty texts
func TestIssue1_EmptyTextHandling(t *testing.T) {
	provider, err := NewSentenceTransformersProvider("all-MiniLM-L6-v2")
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	ctx := context.Background()
	if err := provider.IsAvailable(ctx); err != nil {
		t.Skipf("sentence-transformers not available: %v", err)
	}

	// Test with empty and non-empty texts mixed
	texts := []string{
		"",                // Empty text
		"Valid text",      // Normal text
		"",                // Another empty
		"More valid text", // Normal text
	}

	embeddings, err := provider.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Expected success with mixed empty/valid texts, got error: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// All embeddings should have same dimensions (even for empty texts)
	expectedDims := 384
	for i, emb := range embeddings {
		if len(emb) != expectedDims {
			t.Errorf("Embedding %d: expected %d dimensions, got %d", i, expectedDims, len(emb))
		}
	}
}

// TestIssue1_BatchEmbeddingPerformance tests that batch operations work correctly
// and don't fail due to stderr progress (production scenario from Client0)
func TestIssue1_BatchEmbeddingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch test in short mode")
	}

	provider, err := NewSentenceTransformersProvider("all-MiniLM-L6-v2")
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	ctx := context.Background()
	if err := provider.IsAvailable(ctx); err != nil {
		t.Skipf("sentence-transformers not available: %v", err)
	}

	// Simulate a small batch similar to Client0's use case
	texts := make([]string, 50)
	for i := 0; i < 50; i++ {
		texts[i] = "Test document " + string(rune(i))
	}

	embeddings, err := provider.GenerateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Batch embedding failed: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
}
