// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package embeddings

import (
	"testing"
)

// TestIntegrationScenarios tests real-world usage scenarios
func TestIntegrationScenarios(t *testing.T) {
	tests := []struct {
		name           string
		scenario       string
		modelName      string
		expectedDims   int
		expectedOSS    bool
		shouldAutoWork bool
	}{
		{
			name:           "Client0 - sentence-transformers primary",
			scenario:       "OSS AI stack with sentence-transformers",
			modelName:      "sentence-transformers/all-mpnet-base-v2",
			expectedDims:   768,
			expectedOSS:    true,
			shouldAutoWork: true,
		},
		{
			name:           "Client0 - alternative ST model",
			scenario:       "Testing different sentence-transformers model",
			modelName:      "sentence-transformers/all-minilm-l6-v2",
			expectedDims:   384,
			expectedOSS:    true,
			shouldAutoWork: true,
		},
		{
			name:           "Client0 - Ollama nomic",
			scenario:       "OSS stack with Ollama embeddings",
			modelName:      "nomic-embed-text",
			expectedDims:   768,
			expectedOSS:    true,
			shouldAutoWork: true,
		},
		{
			name:           "OpenAI baseline",
			scenario:       "Compare against OpenAI embeddings",
			modelName:      "text-embedding-3-large",
			expectedDims:   3072,
			expectedOSS:    false,
			shouldAutoWork: true,
		},
		{
			name:           "OpenAI small",
			scenario:       "Cost-effective OpenAI option",
			modelName:      "text-embedding-3-small",
			expectedDims:   1536,
			expectedOSS:    false,
			shouldAutoWork: true,
		},
		{
			name:           "Legacy OpenAI",
			scenario:       "Existing projects using ada-002",
			modelName:      "text-embedding-ada-002",
			expectedDims:   1536,
			expectedOSS:    false,
			shouldAutoWork: true,
		},
		{
			name:           "Custom model",
			scenario:       "User's fine-tuned model",
			modelName:      "my-custom-fine-tuned-model",
			expectedDims:   0,
			expectedOSS:    false,
			shouldAutoWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test GetModelDimensions
			dims, err := GetModelDimensions(tt.modelName)

			if tt.shouldAutoWork {
				if err != nil {
					t.Errorf("Expected auto-detection to work for %s, got error: %v", tt.modelName, err)
					return
				}

				if dims != tt.expectedDims {
					t.Errorf("GetModelDimensions(%s) = %d, want %d", tt.modelName, dims, tt.expectedDims)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for unknown model %s, got dims: %d", tt.modelName, dims)
					return
				}
			}

			// Test GetModelInfo
			info, err := GetModelInfo(tt.modelName)

			if tt.shouldAutoWork {
				if err != nil {
					t.Errorf("Expected model info for %s, got error: %v", tt.modelName, err)
					return
				}

				if info.OSS != tt.expectedOSS {
					t.Errorf("Model %s: OSS = %v, want %v", tt.modelName, info.OSS, tt.expectedOSS)
				}

				if info.Dimensions != tt.expectedDims {
					t.Errorf("Model %s: Dimensions = %d, want %d", tt.modelName, info.Dimensions, tt.expectedDims)
				}
			}
		})
	}
}

// TestBatchReEmbeddingPrep tests that we have all models needed for re-embedding feature
func TestBatchReEmbeddingPrep(t *testing.T) {
	// Models that Client0 will likely test during 3-week validation
	requiredModels := []string{
		"sentence-transformers/all-mpnet-base-v2",     // Primary candidate
		"sentence-transformers/all-minilm-l6-v2",      // Smaller alternative
		"sentence-transformers/all-minilm-l12-v2",     // Mid-size alternative
		"nomic-embed-text",                            // Ollama option
		"text-embedding-3-small",                      // OpenAI baseline (small)
		"text-embedding-3-large",                      // OpenAI baseline (large)
	}

	for _, modelName := range requiredModels {
		t.Run(modelName, func(t *testing.T) {
			dims, err := GetModelDimensions(modelName)
			if err != nil {
				t.Errorf("Model %s not found in registry - needed for batch re-embedding", modelName)
				return
			}

			if dims <= 0 {
				t.Errorf("Model %s has invalid dimensions: %d", modelName, dims)
			}

			// Verify we can get full info
			info, err := GetModelInfo(modelName)
			if err != nil {
				t.Errorf("Could not get info for %s: %v", modelName, err)
				return
			}

			// Log for visibility
			t.Logf("✓ %s: %dd (Provider: %s, OSS: %v)",
				modelName, info.Dimensions, info.Provider, info.OSS)
		})
	}
}

// TestProviderCoverage ensures we support main OSS providers
func TestProviderCoverage(t *testing.T) {
	requiredProviders := map[string]int{
		"sentence-transformers": 5, // At least 5 ST models
		"ollama":                2, // At least 2 Ollama models
		"openai":                3, // At least 3 OpenAI models
	}

	for provider, minCount := range requiredProviders {
		t.Run(provider, func(t *testing.T) {
			models := ListModelsByProvider(provider)
			if len(models) < minCount {
				t.Errorf("Provider %s has %d models, want at least %d",
					provider, len(models), minCount)
			}

			// Log for visibility
			t.Logf("Provider %s: %d models", provider, len(models))
			for _, model := range models {
				t.Logf("  - %s (%dd)", model.Name, model.Dimensions)
			}
		})
	}
}

// TestOSSStackSupport tests that we adequately support OSS AI stacks
func TestOSSStackSupport(t *testing.T) {
	ossModels := GetOSSModels()

	if len(ossModels) < 8 {
		t.Errorf("Found %d OSS models, want at least 8 for good OSS support", len(ossModels))
	}

	// Should include both sentence-transformers and ollama
	hasSTModels := false
	hasOllamaModels := false

	for _, model := range ossModels {
		if model.Provider == "sentence-transformers" {
			hasSTModels = true
		}
		if model.Provider == "ollama" {
			hasOllamaModels = true
		}
	}

	if !hasSTModels {
		t.Error("OSS models should include sentence-transformers")
	}

	if !hasOllamaModels {
		t.Error("OSS models should include Ollama models")
	}

	// Log OSS options
	t.Logf("OSS Models available: %d", len(ossModels))
	for _, model := range ossModels {
		t.Logf("  - %s/%s (%dd)",
			model.Provider, model.Name, model.Dimensions)
	}
}
