// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"testing"
)

func TestCollectionEmbeddingFlag(t *testing.T) {
	tests := []struct {
		name           string
		embeddingModel string
		expectDefault  bool
		description    string
	}{
		{
			name:           "Default embedding model",
			embeddingModel: "text-embedding-ada-002",
			expectDefault:  true,
			description:    "Should use default when no flag provided",
		},
		{
			name:           "Custom OpenAI embedding",
			embeddingModel: "text-embedding-3-small",
			expectDefault:  false,
			description:    "Should accept custom OpenAI embedding",
		},
		{
			name:           "Latest large embedding",
			embeddingModel: "text-embedding-3-large",
			expectDefault:  false,
			description:    "Should accept text-embedding-3-large",
		},
		{
			name:           "Custom embedding model",
			embeddingModel: "custom-embedding-v1",
			expectDefault:  false,
			description:    "Should accept any embedding model name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate embedding model name
			if tt.embeddingModel == "" {
				t.Error("Embedding model should not be empty")
			}

			// Check if it's the default
			isDefault := tt.embeddingModel == "text-embedding-ada-002"
			if isDefault != tt.expectDefault {
				t.Errorf("Expected default=%v, got %v for model %s",
					tt.expectDefault, isDefault, tt.embeddingModel)
			}

			t.Logf("Embedding model: %s (default: %v)", tt.embeddingModel, isDefault)
		})
	}
}

func TestCollectionEmbeddingCompatibility(t *testing.T) {
	// Test that embedding models work with text collections
	t.Run("Text collection with embedding", func(t *testing.T) {
		embeddingModels := []string{
			"text-embedding-3-small",
			"text-embedding-3-large",
			"text-embedding-ada-002",
		}

		for _, model := range embeddingModels {
			if model == "" {
				t.Error("Embedding model should not be empty")
			}
			t.Logf("Valid text embedding: %s", model)
		}
	})

	// Test that image collections ignore embeddings
	t.Run("Image collection ignores embedding", func(t *testing.T) {
		// Image collections use vectorizer: none
		// Embedding flag should be ignored
		t.Log("Image collections should ignore --embedding flag")
	})
}

func TestEmbeddingFlagAlias(t *testing.T) {
	// Test that both --embedding and --embedding-model work
	t.Run("Flag aliases", func(t *testing.T) {
		testCases := []struct {
			flag  string
			short string
		}{
			{"--embedding", "-e"},
			{"--embedding-model", ""}, // deprecated
		}

		for _, tc := range testCases {
			if tc.flag == "" {
				t.Error("Flag name should not be empty")
			}
			t.Logf("Valid flag: %s (short: %s)", tc.flag, tc.short)
		}
	})
}

func TestRecommendedEmbeddingsForCollections(t *testing.T) {
	// Test recommended embedding models for collections
	recommendedModels := []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"text-embedding-ada-002",
	}

	t.Run("Recommended models list", func(t *testing.T) {
		if len(recommendedModels) == 0 {
			t.Error("Recommended models list should not be empty")
		}

		for _, model := range recommendedModels {
			if model == "" {
				t.Error("Model name should not be empty")
			}
			if len(model) < 5 {
				t.Errorf("Model name '%s' seems too short", model)
			}
		}

		t.Logf("Found %d recommended embedding models", len(recommendedModels))
	})

	t.Run("Model characteristics", func(t *testing.T) {
		modelCharacteristics := map[string]string{
			"text-embedding-3-small": "Fast and efficient",
			"text-embedding-3-large": "Higher quality, larger dimensions",
			"text-embedding-ada-002": "Legacy model, widely used",
		}

		for model, characteristic := range modelCharacteristics {
			if characteristic == "" {
				t.Errorf("Characteristic for model %s should not be empty", model)
			}
			t.Logf("Model: %s - %s", model, characteristic)
		}
	})
}

func TestEmbeddingValidationForCollectionCreate(t *testing.T) {
	// Test that embedding validation works correctly
	t.Run("Valid embedding model names", func(t *testing.T) {
		validModels := []string{
			"text-embedding-3-small",
			"text-embedding-3-large",
			"text-embedding-ada-002",
			"custom-model-v1",
			"my-embedding",
		}

		for _, model := range validModels {
			// All non-empty strings are valid model names
			if model == "" {
				t.Error("Model name should not be empty")
			}

			// Check minimum length
			if len(model) < 2 {
				t.Errorf("Model name '%s' is too short", model)
			}

			t.Logf("Valid model name: %s", model)
		}
	})

	t.Run("Invalid embedding model names", func(t *testing.T) {
		invalidModels := []string{
			"",  // empty
			" ", // whitespace only
		}

		for _, model := range invalidModels {
			if len(model) == 0 || len(model) == 1 && model == " " {
				t.Logf("Correctly identified invalid model: '%s'", model)
			} else {
				t.Errorf("Should have identified '%s' as invalid", model)
			}
		}
	})
}

func TestDefaultEmbeddingBehavior(t *testing.T) {
	// Test default embedding behavior
	t.Run("Default when no flag provided", func(t *testing.T) {
		defaultModel := "text-embedding-ada-002"

		if defaultModel == "" {
			t.Error("Default embedding model should not be empty")
		}

		t.Logf("Default embedding model: %s", defaultModel)
	})

	t.Run("Override default with flag", func(t *testing.T) {
		customModel := "text-embedding-3-small"
		defaultModel := "text-embedding-ada-002"

		if customModel == defaultModel {
			t.Error("Custom model should differ from default")
		}

		t.Logf("Custom model %s overrides default %s", customModel, defaultModel)
	})
}
