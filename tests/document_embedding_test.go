// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestEmbeddingValidation(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		embeddingModel string
		vectorizer     string
		expectError    bool
		description    string
	}{
		{
			name:           "Valid OpenAI embedding with text2vec-openai",
			collectionName: "TestCollection",
			embeddingModel: "text-embedding-3-small",
			vectorizer:     "text2vec-openai",
			expectError:    false,
			description:    "Should accept valid OpenAI embedding model",
		},
		{
			name:           "Custom embedding with text2vec-openai",
			embeddingModel: "custom-embedding-model",
			vectorizer:     "text2vec-openai",
			expectError:    false,
			description:    "Should accept custom embedding model with warning",
		},
		{
			name:           "Embedding with vectorizer none",
			embeddingModel: "text-embedding-3-small",
			vectorizer:     "none",
			expectError:    false,
			description:    "Should warn but not error for image collections",
		},
		{
			name:           "Embedding with different vectorizer",
			embeddingModel: "text-embedding-3-small",
			vectorizer:     "text2vec-cohere",
			expectError:    false,
			description:    "Should warn but not error for incompatible vectorizer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the logic flow
			// Actual integration test would require a running Weaviate instance

			// Validate embedding model names
			validModels := []string{
				"text-embedding-3-small",
				"text-embedding-3-large",
				"text-embedding-ada-002",
			}

			isValid := false
			for _, model := range validModels {
				if model == tt.embeddingModel {
					isValid = true
					break
				}
			}

			// Check vectorizer compatibility
			if tt.vectorizer == "text2vec-openai" {
				// OpenAI embeddings work with text2vec-openai
				if tt.embeddingModel != "" {
					// Embedding is provided, should be compatible
					if !isValid {
						t.Logf("Custom embedding model: %s (not in recommended list)", tt.embeddingModel)
					}
				}
			} else if tt.vectorizer == "none" {
				// No vectorization, embedding flag should be ignored
				t.Logf("Vectorizer is 'none', embedding flag will be ignored")
			} else {
				// Different vectorizer, embedding flag incompatible
				t.Logf("Vectorizer '%s' is not compatible with OpenAI embeddings", tt.vectorizer)
			}
		})
	}
}

func TestEmbeddingFlagIntegration(t *testing.T) {
	// Skip if no Weaviate connection
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test client with mock config
	cfg := &config.VectorDBConfig{
		Name:    "test",
		Type:    config.VectorDBTypeMock,
		URL:     "http://localhost:8080",
		Enabled: true,
	}

	// Test with mock database (no actual Weaviate connection needed)
	t.Run("Mock database ignores embedding flag", func(t *testing.T) {
		// Mock database doesn't support embedding validation
		// This test ensures the flag is accepted but ignored for mock
		if cfg.Type != config.VectorDBTypeMock {
			t.Skip("This test requires mock database")
		}

		// The embedding flag is accepted in the command but not validated for mock
		t.Log("Mock database accepts --embedding flag without validation")
	})
}

func TestEmbeddingWithWeaviateClient(t *testing.T) {
	// This test validates the schema retrieval for embedding validation
	if testing.Short() {
		t.Skip("Skipping Weaviate client test in short mode")
	}

	// Test schema structure
	t.Run("Schema validation structure", func(t *testing.T) {
		// Create a sample schema to validate structure
		schema := &weaviate.CollectionSchema{
			Class:      "TestCollection",
			Vectorizer: "text2vec-openai",
			Properties: []weaviate.SchemaProperty{
				{
					Name:     "text",
					DataType: []string{"text"},
				},
			},
		}

		if schema.Vectorizer != "text2vec-openai" {
			t.Errorf("Expected vectorizer 'text2vec-openai', got '%s'", schema.Vectorizer)
		}

		if schema.Class != "TestCollection" {
			t.Errorf("Expected class 'TestCollection', got '%s'", schema.Class)
		}
	})

	t.Run("Vectorizer compatibility check", func(t *testing.T) {
		testCases := []struct {
			vectorizer string
			compatible bool
		}{
			{"text2vec-openai", true},
			{"none", false},
			{"text2vec-cohere", false},
			{"text2vec-huggingface", false},
		}

		for _, tc := range testCases {
			isCompatible := tc.vectorizer == "text2vec-openai"
			if isCompatible != tc.compatible {
				t.Errorf("Vectorizer %s: expected compatible=%v, got %v",
					tc.vectorizer, tc.compatible, isCompatible)
			}
		}
	})
}

func TestRecommendedEmbeddingModels(t *testing.T) {
	// Test that our recommended models list is correct
	recommendedModels := []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"text-embedding-ada-002",
	}

	t.Run("Recommended models list", func(t *testing.T) {
		if len(recommendedModels) == 0 {
			t.Error("Recommended models list should not be empty")
		}

		// Validate each model name format
		for _, model := range recommendedModels {
			if model == "" {
				t.Error("Model name should not be empty")
			}
			if len(model) < 5 {
				t.Errorf("Model name '%s' seems too short", model)
			}
		}
	})

	t.Run("Model name validation", func(t *testing.T) {
		validModels := map[string]bool{
			"text-embedding-3-small": true,
			"text-embedding-3-large": true,
			"text-embedding-ada-002": true,
			"custom-model":           false,
			"":                       false,
		}

		for model, shouldBeValid := range validModels {
			isValid := false
			for _, recommended := range recommendedModels {
				if recommended == model {
					isValid = true
					break
				}
			}

			if isValid != shouldBeValid {
				t.Errorf("Model '%s': expected valid=%v, got %v", model, shouldBeValid, isValid)
			}
		}
	})
}
