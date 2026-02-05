// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package embeddings

import (
	"testing"
)

func TestGetModelDimensions(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		wantDims      int
		wantErr       bool
		errorContains string
	}{
		// sentence-transformers (full names)
		{"sentence-transformers all-mpnet", "sentence-transformers/all-mpnet-base-v2", 768, false, ""},
		{"sentence-transformers minilm-l6", "sentence-transformers/all-minilm-l6-v2", 384, false, ""},
		{"sentence-transformers minilm-l12", "sentence-transformers/all-minilm-l12-v2", 384, false, ""},

		// sentence-transformers (aliases without prefix)
		{"alias all-mpnet", "all-mpnet-base-v2", 768, false, ""},
		{"alias minilm-l6", "all-minilm-l6-v2", 384, false, ""},

		// OpenAI
		{"openai embedding-3-small", "text-embedding-3-small", 1536, false, ""},
		{"openai embedding-3-large", "text-embedding-3-large", 3072, false, ""},
		{"openai ada-002", "text-embedding-ada-002", 1536, false, ""},

		// Ollama
		{"ollama nomic", "nomic-embed-text", 768, false, ""},
		{"ollama mxbai", "mxbai-embed-large", 1024, false, ""},

		// Cohere
		{"cohere english", "embed-english-v3.0", 1024, false, ""},
		{"cohere multilingual", "embed-multilingual-v3.0", 1024, false, ""},

		// Case insensitivity
		{"uppercase", "SENTENCE-TRANSFORMERS/ALL-MPNET-BASE-V2", 768, false, ""},
		{"mixed case", "Text-Embedding-3-Small", 1536, false, ""},

		// Whitespace handling
		{"with spaces", "  sentence-transformers/all-mpnet-base-v2  ", 768, false, ""},

		// Unknown models
		{"unknown model", "unknown-model-xyz", 0, true, "unknown embedding model"},
		{"custom model", "my-custom-embeddings", 0, true, "unknown embedding model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := GetModelDimensions(tt.modelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && dims != tt.wantDims {
				t.Errorf("GetModelDimensions() = %d, want %d", dims, tt.wantDims)
			}
		})
	}
}

func TestGetModelInfo(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		wantProvider string
		wantDims     int
		wantOSS      bool
		wantErr      bool
	}{
		{"sentence-transformers", "sentence-transformers/all-mpnet-base-v2", "sentence-transformers", 768, true, false},
		{"openai", "text-embedding-3-small", "openai", 1536, false, false},
		{"ollama", "nomic-embed-text", "ollama", 768, true, false},
		{"cohere", "embed-english-v3.0", "cohere", 1024, false, false},
		{"unknown", "unknown-model", "", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if info.Provider != tt.wantProvider {
				t.Errorf("GetModelInfo() provider = %s, want %s", info.Provider, tt.wantProvider)
			}

			if info.Dimensions != tt.wantDims {
				t.Errorf("GetModelInfo() dimensions = %d, want %d", info.Dimensions, tt.wantDims)
			}

			if info.OSS != tt.wantOSS {
				t.Errorf("GetModelInfo() OSS = %v, want %v", info.OSS, tt.wantOSS)
			}
		})
	}
}

func TestIsKnownModel(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		want      bool
	}{
		{"known sentence-transformers", "sentence-transformers/all-mpnet-base-v2", true},
		{"known openai", "text-embedding-3-small", true},
		{"known ollama", "nomic-embed-text", true},
		{"known alias", "all-mpnet-base-v2", true},
		{"unknown", "unknown-model", false},
		{"custom", "my-custom-model", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKnownModel(tt.modelName); got != tt.want {
				t.Errorf("IsKnownModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListKnownModels(t *testing.T) {
	models := ListKnownModels()

	if len(models) == 0 {
		t.Error("ListKnownModels() returned empty list")
	}

	// Should have at least 15+ models (excluding aliases)
	if len(models) < 15 {
		t.Errorf("ListKnownModels() returned %d models, want at least 15", len(models))
	}

	// Check that we have models from different providers
	providers := make(map[string]bool)
	for _, model := range models {
		providers[model.Provider] = true
	}

	expectedProviders := []string{"sentence-transformers", "openai", "ollama", "cohere"}
	for _, provider := range expectedProviders {
		if !providers[provider] {
			t.Errorf("ListKnownModels() missing provider: %s", provider)
		}
	}
}

func TestListModelsByProvider(t *testing.T) {
	tests := []struct {
		provider    string
		minExpected int
	}{
		{"sentence-transformers", 5}, // Should have at least 5 sentence-transformers models
		{"openai", 3},                // Should have at least 3 OpenAI models
		{"ollama", 2},                // Should have at least 2 Ollama models
		{"cohere", 2},                // Should have at least 2 Cohere models
		{"unknown", 0},               // Unknown provider should return empty
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			models := ListModelsByProvider(tt.provider)

			if len(models) < tt.minExpected {
				t.Errorf("ListModelsByProvider(%s) returned %d models, want at least %d",
					tt.provider, len(models), tt.minExpected)
			}

			// Verify all returned models are from the correct provider
			for _, model := range models {
				if model.Provider != tt.provider {
					t.Errorf("ListModelsByProvider(%s) returned model from provider %s",
						tt.provider, model.Provider)
				}
			}
		})
	}
}

func TestGetOSSModels(t *testing.T) {
	models := GetOSSModels()

	if len(models) == 0 {
		t.Error("GetOSSModels() returned empty list")
	}

	// Should have at least 8+ OSS models (sentence-transformers + ollama)
	if len(models) < 8 {
		t.Errorf("GetOSSModels() returned %d models, want at least 8", len(models))
	}

	// Verify all returned models are OSS
	for _, model := range models {
		if !model.OSS {
			t.Errorf("GetOSSModels() returned non-OSS model: %s", model.Name)
		}
	}

	// Should include sentence-transformers and ollama
	hasSTModels := false
	hasOllamaModels := false
	for _, model := range models {
		if model.Provider == "sentence-transformers" {
			hasSTModels = true
		}
		if model.Provider == "ollama" {
			hasOllamaModels = true
		}
	}

	if !hasSTModels {
		t.Error("GetOSSModels() should include sentence-transformers models")
	}
	if !hasOllamaModels {
		t.Error("GetOSSModels() should include Ollama models")
	}
}

func TestModelInfoValues(t *testing.T) {
	// Test specific models have correct attributes
	tests := []struct {
		modelName    string
		wantDims     int
		wantProvider string
		wantOSS      bool
	}{
		{
			"sentence-transformers/all-mpnet-base-v2",
			768,
			"sentence-transformers",
			true,
		},
		{
			"text-embedding-3-large",
			3072,
			"openai",
			false,
		},
		{
			"nomic-embed-text",
			768,
			"ollama",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)
			if err != nil {
				t.Fatalf("GetModelInfo() error = %v", err)
			}

			if info.Dimensions != tt.wantDims {
				t.Errorf("dimensions = %d, want %d", info.Dimensions, tt.wantDims)
			}

			if info.Provider != tt.wantProvider {
				t.Errorf("provider = %s, want %s", info.Provider, tt.wantProvider)
			}

			if info.OSS != tt.wantOSS {
				t.Errorf("OSS = %v, want %v", info.OSS, tt.wantOSS)
			}
		})
	}
}
