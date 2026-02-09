// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"context"
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestNewEmbeddingPipeline(t *testing.T) {
	tests := []struct {
		name        string
		modelName   string
		shouldErr   bool
		errContains string
		expectedDim int
		expectedPrv string
	}{
		{
			name:        "OpenAI text-embedding-3-small",
			modelName:   "text-embedding-3-small",
			shouldErr:   false,
			expectedDim: 1536,
			expectedPrv: "openai",
		},
		{
			name:        "OpenAI text-embedding-3-large",
			modelName:   "text-embedding-3-large",
			shouldErr:   false,
			expectedDim: 3072,
			expectedPrv: "openai",
		},
		// Skip: sentence-transformers test requires Python + library
		// Test would fail in CI unless dependencies installed
		{
			name:        "empty model name",
			modelName:   "",
			shouldErr:   true,
			errContains: "cannot be empty",
		},
		{
			name:        "unknown model",
			modelName:   "unknown-model-xyz",
			shouldErr:   true,
			errContains: "unknown embedding model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a dummy API key for OpenAI tests
			if tt.expectedPrv == "openai" {
				os.Setenv("OPENAI_API_KEY", "sk-test-key-for-testing")
				defer os.Unsetenv("OPENAI_API_KEY")
			}

			pipeline, err := NewEmbeddingPipeline(tt.modelName)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pipeline == nil {
				t.Error("Expected non-nil pipeline")
				return
			}

			if pipeline.GetDimensions() != tt.expectedDim {
				t.Errorf("Expected dimensions %d, got %d", tt.expectedDim, pipeline.GetDimensions())
			}

			if pipeline.GetProvider() != tt.expectedPrv {
				t.Errorf("Expected provider %s, got %s", tt.expectedPrv, pipeline.GetProvider())
			}

			if pipeline.GetModelName() != tt.modelName {
				t.Errorf("Expected model name %s, got %s", tt.modelName, pipeline.GetModelName())
			}
		})
	}
}

func TestNewEmbeddingPipeline_MissingAPIKey(t *testing.T) {
	// Ensure no API key is set
	os.Unsetenv("OPENAI_API_KEY")

	pipeline, err := NewEmbeddingPipeline("text-embedding-3-small")

	if err == nil {
		t.Error("Expected error for missing OPENAI_API_KEY, got nil")
		return
	}

	if !contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("Expected error about OPENAI_API_KEY, got: %s", err.Error())
	}

	if pipeline != nil {
		t.Error("Expected nil pipeline when API key is missing")
	}
}

func TestEmbeddingPipeline_GetDimensions(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	pipeline, err := NewEmbeddingPipeline("text-embedding-3-small")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	dims := pipeline.GetDimensions()
	if dims != 1536 {
		t.Errorf("Expected dimensions 1536, got %d", dims)
	}
}

func TestEmbeddingPipeline_GetModelName(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	modelName := "text-embedding-3-large"
	pipeline, err := NewEmbeddingPipeline(modelName)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if pipeline.GetModelName() != modelName {
		t.Errorf("Expected model name %s, got %s", modelName, pipeline.GetModelName())
	}
}

func TestEmbeddingPipeline_GetProvider(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	tests := []struct {
		modelName        string
		expectedProvider string
	}{
		{"text-embedding-3-small", "openai"},
		{"text-embedding-3-large", "openai"},
		{"text-embedding-ada-002", "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			pipeline, err := NewEmbeddingPipeline(tt.modelName)
			if err != nil {
				t.Fatalf("Failed to create pipeline: %v", err)
			}

			if pipeline.GetProvider() != tt.expectedProvider {
				t.Errorf("Expected provider %s, got %s", tt.expectedProvider, pipeline.GetProvider())
			}
		})
	}
}

func TestEmbeddingPipeline_IsSupported(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	tests := []struct {
		modelName string
		supported bool
	}{
		{"text-embedding-3-small", true}, // OpenAI - supported
		{"text-embedding-3-large", true}, // OpenAI - supported
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			pipeline, err := NewEmbeddingPipeline(tt.modelName)
			if err != nil {
				t.Fatalf("Failed to create pipeline: %v", err)
			}

			if pipeline.IsSupported() != tt.supported {
				t.Errorf("Expected IsSupported %v, got %v", tt.supported, pipeline.IsSupported())
			}
		})
	}
}

func TestEmbeddingPipeline_ProcessBatch_UnsupportedProvider(t *testing.T) {
	// Create pipeline with nil provider (simulates provider creation failure)
	pipeline := &EmbeddingPipeline{
		embeddingModel: "sentence-transformers/all-mpnet-base-v2",
		provider:       "sentence-transformers",
		dimensions:     768,
		embProvider:    nil, // No provider for this test
	}

	docs := []*vectordb.Document{
		{ID: "doc1", Text: "test document"},
	}

	ctx := context.Background()
	err := pipeline.ProcessBatch(ctx, docs)

	if err == nil {
		t.Error("Expected error for nil provider, got nil")
		return
	}

	if !contains(err.Error(), "provider not initialized") {
		t.Errorf("Expected error about provider not initialized, got: %s", err.Error())
	}
}

func TestEmbeddingPipeline_ProcessBatch_EmptyDocs(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	pipeline, err := NewEmbeddingPipeline("text-embedding-3-small")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	ctx := context.Background()
	err = pipeline.ProcessBatch(ctx, []*vectordb.Document{})

	if err != nil {
		t.Errorf("Expected no error for empty docs, got: %v", err)
	}
}

func TestEmbeddingPipeline_ProcessBatch_NilDocs(t *testing.T) {
	// Skip if OPENAI_API_KEY not set (test requires real API or mock)
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	pipeline, err := NewEmbeddingPipeline("text-embedding-3-small")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	ctx := context.Background()

	// Should skip nil documents without error
	docs := []*vectordb.Document{nil, nil}
	err = pipeline.ProcessBatch(ctx, docs)

	if err != nil {
		t.Errorf("Expected no error for nil docs, got: %v", err)
	}
}
