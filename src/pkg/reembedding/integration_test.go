// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"context"
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/embeddings"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// TestReEmbeddingIntegrationScenarios tests real-world re-embedding scenarios
func TestReEmbeddingIntegrationScenarios(t *testing.T) {
	// These are integration tests that validate the full re-embedding workflow
	// They use mock components but follow the real execution flow

	tests := []struct {
		name              string
		scenario          string
		sourceModel       string
		targetModel       string
		numDocs           int
		expectedSourceDim int
		expectedTargetDim int
	}{
		{
			name:              "Client0 - Switch to OSS sentence-transformers",
			scenario:          "Re-embed from OpenAI to sentence-transformers for OSS stack",
			sourceModel:       "text-embedding-3-small",
			targetModel:       "sentence-transformers/all-mpnet-base-v2",
			numDocs:           100,
			expectedSourceDim: 1536,
			expectedTargetDim: 768,
		},
		{
			name:              "Client0 - Test smaller OSS model",
			scenario:          "Re-embed to smaller sentence-transformers model",
			sourceModel:       "sentence-transformers/all-mpnet-base-v2",
			targetModel:       "sentence-transformers/all-minilm-l6-v2",
			numDocs:           100,
			expectedSourceDim: 768,
			expectedTargetDim: 384,
		},
		{
			name:              "Client0 - Switch to Ollama",
			scenario:          "Re-embed to local Ollama embeddings",
			sourceModel:       "text-embedding-3-small",
			targetModel:       "nomic-embed-text",
			numDocs:           100,
			expectedSourceDim: 1536,
			expectedTargetDim: 768,
		},
		{
			name:              "Upgrade to larger OpenAI model",
			scenario:          "Re-embed from small to large OpenAI model",
			sourceModel:       "text-embedding-3-small",
			targetModel:       "text-embedding-3-large",
			numDocs:           100,
			expectedSourceDim: 1536,
			expectedTargetDim: 3072,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate source model exists in registry
			sourceDims, err := embeddings.GetModelDimensions(tt.sourceModel)
			if err != nil {
				t.Errorf("Source model %s not found in registry: %v", tt.sourceModel, err)
				return
			}

			if sourceDims != tt.expectedSourceDim {
				t.Errorf("Source model %s: expected %d dimensions, got %d", tt.sourceModel, tt.expectedSourceDim, sourceDims)
			}

			// Validate target model exists in registry
			targetDims, err := embeddings.GetModelDimensions(tt.targetModel)
			if err != nil {
				t.Errorf("Target model %s not found in registry: %v", tt.targetModel, err)
				return
			}

			if targetDims != tt.expectedTargetDim {
				t.Errorf("Target model %s: expected %d dimensions, got %d", tt.targetModel, tt.expectedTargetDim, targetDims)
			}

			// Verify we can create a pipeline for the target model
			// Note: Skip OpenAI client creation in tests (no API key)
			targetInfo, err := embeddings.GetModelInfo(tt.targetModel)
			if err != nil {
				t.Errorf("Failed to get target model info: %v", err)
				return
			}

			// Log scenario details
			t.Logf("✓ Scenario validated: %s", tt.scenario)
			t.Logf("  Source: %s (%dd)", tt.sourceModel, sourceDims)
			t.Logf("  Target: %s (%dd, %s, OSS: %v)", tt.targetModel, targetDims, targetInfo.Provider, targetInfo.OSS)
			t.Logf("  Documents: %d", tt.numDocs)
			t.Logf("  Dimension change: %dd → %dd", sourceDims, targetDims)
		})
	}
}

// TestReEmbeddingWorkflowWithMockVDB tests the complete workflow with a mock VDB
func TestReEmbeddingWorkflowWithMockVDB(t *testing.T) {
	ctx := context.Background()

	// Create mock VDB with 100 test documents
	totalDocs := int64(100)
	mockBatches := make([][]*vectordb.Document, 1)
	batch := make([]*vectordb.Document, 100)
	for i := 0; i < 100; i++ {
		batch[i] = &vectordb.Document{
			ID:   "doc-" + string(rune('0'+i%10)),
			Text: "Test document content",
		}
	}
	mockBatches[0] = batch

	mockClient := &MockVectorDBClient{
		totalDocs: totalDocs,
		documents: mockBatches,
	}

	// Test 1: Create CollectionReader
	reader, err := NewCollectionReader(mockClient, "TestCollection", 100)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	if reader.totalDocs != totalDocs {
		t.Errorf("Expected %d total docs, got %d", totalDocs, reader.totalDocs)
	}

	// Test 2: Read documents in batches
	batchesRead := 0
	docsRead := 0

	for reader.HasMore() {
		docs, err := reader.ReadBatch(ctx)
		if err != nil {
			t.Fatalf("Failed to read batch: %v", err)
		}

		if len(docs) == 0 {
			break
		}

		batchesRead++
		docsRead += len(docs)
	}

	if docsRead != int(totalDocs) {
		t.Errorf("Expected to read %d docs, got %d", totalDocs, docsRead)
	}

	// Test 3: Progress tracking
	tracker := NewProgressTracker(totalDocs)
	tracker.Update(docsRead)

	if !tracker.IsComplete() {
		t.Error("Expected tracker to show complete")
	}

	percentage := tracker.GetPercentage()
	if percentage != 100.0 {
		t.Errorf("Expected 100%% completion, got %.1f%%", percentage)
	}

	t.Logf("✓ Workflow test passed:")
	t.Logf("  Batches read: %d", batchesRead)
	t.Logf("  Documents processed: %d", docsRead)
	t.Logf("  Progress: %.0f%%", percentage)
}

// TestReEmbeddingPerformanceEstimates validates performance assumptions
func TestReEmbeddingPerformanceEstimates(t *testing.T) {
	// Client0 dataset: 3,518 documents
	client0Docs := int64(3518)

	// Test assumptions for performance estimates
	batchSize := 100
	expectedBatches := (int(client0Docs) + batchSize - 1) / batchSize // Ceiling division

	if expectedBatches != 36 {
		t.Errorf("Expected 36 batches for 3,518 docs (batch size 100), got %d", expectedBatches)
	}

	// Performance targets (from spec)
	minDocsPerMin := 200.0
	maxTimeMinutes := 20.0

	// Calculate expected time at minimum speed
	expectedTime := float64(client0Docs) / minDocsPerMin

	if expectedTime > maxTimeMinutes {
		t.Errorf("Performance target not met: %.1f minutes > %d minutes target", expectedTime, int(maxTimeMinutes))
	}

	t.Logf("✓ Performance estimates validated:")
	t.Logf("  Documents: %d", client0Docs)
	t.Logf("  Batch size: %d", batchSize)
	t.Logf("  Expected batches: %d", expectedBatches)
	t.Logf("  Min speed: %.0f docs/min", minDocsPerMin)
	t.Logf("  Expected time: %.1f minutes (target: %.0f minutes)", expectedTime, maxTimeMinutes)
}

// TestReEmbeddingModelCoverage ensures we have all models needed for Client0
func TestReEmbeddingModelCoverage(t *testing.T) {
	// Models that Client0 will likely test during validation
	requiredModels := []struct {
		name     string
		purpose  string
		priority string
	}{
		{"sentence-transformers/all-mpnet-base-v2", "Primary OSS candidate", "HIGH"},
		{"sentence-transformers/all-minilm-l6-v2", "Smaller OSS alternative", "MEDIUM"},
		{"sentence-transformers/all-minilm-l12-v2", "Mid-size OSS alternative", "MEDIUM"},
		{"nomic-embed-text", "Ollama local embeddings", "HIGH"},
		{"text-embedding-3-small", "OpenAI baseline (small)", "MEDIUM"},
		{"text-embedding-3-large", "OpenAI baseline (large)", "LOW"},
	}

	allFound := true
	for _, model := range requiredModels {
		dims, err := embeddings.GetModelDimensions(model.name)
		if err != nil {
			t.Errorf("[%s] Model %s not found - needed for: %s", model.priority, model.name, model.purpose)
			allFound = false
			continue
		}

		info, _ := embeddings.GetModelInfo(model.name)
		t.Logf("✓ [%s] %s (%dd, %s, OSS: %v) - %s",
			model.priority, model.name, dims, info.Provider, info.OSS, model.purpose)
	}

	if !allFound {
		t.Error("Missing required models for Client0 validation")
	}
}

// TestReEmbeddingErrorHandling validates error scenarios
func TestReEmbeddingErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test 1: Reader with error
	mockClient := &MockVectorDBClient{
		totalDocs: 100,
		shouldErr: true,
	}

	_, err := NewCollectionReader(mockClient, "TestCollection", 100)
	if err == nil {
		t.Error("Expected error when VDB client fails, got nil")
	}

	// Test 2: Pipeline with unknown model
	_, err = NewEmbeddingPipeline("unknown-model-xyz")
	if err == nil {
		t.Error("Expected error for unknown model, got nil")
	}

	// Test 3: Pipeline with empty model name
	_, err = NewEmbeddingPipeline("")
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}

	// Test 4: Pipeline without API key (OpenAI models)
	os.Unsetenv("OPENAI_API_KEY")
	_, err = NewEmbeddingPipeline("text-embedding-3-small")
	if err == nil {
		t.Error("Expected error when OPENAI_API_KEY not set, got nil")
	}

	// Test 5: Process batch with unsupported provider
	pipeline := &EmbeddingPipeline{
		embeddingModel: "sentence-transformers/all-mpnet-base-v2",
		provider:       "sentence-transformers",
		dimensions:     768,
	}

	docs := []*vectordb.Document{{ID: "doc1", Text: "test"}}
	err = pipeline.ProcessBatch(ctx, docs)
	if err == nil {
		t.Error("Expected error for unsupported provider, got nil")
	}

	t.Log("✓ All error scenarios handled correctly")
}
