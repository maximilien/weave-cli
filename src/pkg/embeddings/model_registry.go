// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package embeddings

import (
	"fmt"
	"strings"
)

// ModelInfo contains information about an embedding model
type ModelInfo struct {
	Name       string
	Dimensions int
	Provider   string
	MaxTokens  int
	OSS        bool // true for open-source models
}

// Known embedding models and their dimensions
var knownModels = map[string]ModelInfo{
	// sentence-transformers (OSS - HuggingFace)
	"sentence-transformers/all-mpnet-base-v2": {
		Name:       "all-mpnet-base-v2",
		Dimensions: 768,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"sentence-transformers/all-minilm-l6-v2": {
		Name:       "all-MiniLM-L6-v2",
		Dimensions: 384,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"sentence-transformers/all-minilm-l12-v2": {
		Name:       "all-MiniLM-L12-v2",
		Dimensions: 384,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"sentence-transformers/paraphrase-multilingual-mpnet-base-v2": {
		Name:       "paraphrase-multilingual-mpnet-base-v2",
		Dimensions: 768,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"sentence-transformers/multi-qa-mpnet-base-dot-v1": {
		Name:       "multi-qa-mpnet-base-dot-v1",
		Dimensions: 768,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"sentence-transformers/paraphrase-minilm-l6-v2": {
		Name:       "paraphrase-MiniLM-L6-v2",
		Dimensions: 384,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},

	// OpenAI
	"openai/text-embedding-3-small": {
		Name:       "text-embedding-3-small",
		Dimensions: 1536,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},
	"openai/text-embedding-3-large": {
		Name:       "text-embedding-3-large",
		Dimensions: 3072,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},
	"openai/text-embedding-ada-002": {
		Name:       "text-embedding-ada-002",
		Dimensions: 1536,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},

	// Ollama (OSS - local)
	"ollama/nomic-embed-text": {
		Name:       "nomic-embed-text",
		Dimensions: 768,
		Provider:   "ollama",
		MaxTokens:  8192,
		OSS:        true,
	},
	"ollama/mxbai-embed-large": {
		Name:       "mxbai-embed-large",
		Dimensions: 1024,
		Provider:   "ollama",
		MaxTokens:  512,
		OSS:        true,
	},
	"ollama/snowflake-arctic-embed": {
		Name:       "snowflake-arctic-embed",
		Dimensions: 1024,
		Provider:   "ollama",
		MaxTokens:  512,
		OSS:        true,
	},

	// Cohere
	"cohere/embed-english-v3.0": {
		Name:       "embed-english-v3.0",
		Dimensions: 1024,
		Provider:   "cohere",
		MaxTokens:  512,
		OSS:        false,
	},
	"cohere/embed-multilingual-v3.0": {
		Name:       "embed-multilingual-v3.0",
		Dimensions: 1024,
		Provider:   "cohere",
		MaxTokens:  512,
		OSS:        false,
	},

	// Voyage AI
	"voyage/voyage-2": {
		Name:       "voyage-2",
		Dimensions: 1024,
		Provider:   "voyage",
		MaxTokens:  4000,
		OSS:        false,
	},
	"voyage/voyage-large-2": {
		Name:       "voyage-large-2",
		Dimensions: 1536,
		Provider:   "voyage",
		MaxTokens:  16000,
		OSS:        false,
	},

	// Aliases for common usage (without provider prefix)
	"text-embedding-3-small": {
		Name:       "text-embedding-3-small",
		Dimensions: 1536,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},
	"text-embedding-3-large": {
		Name:       "text-embedding-3-large",
		Dimensions: 3072,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},
	"text-embedding-ada-002": {
		Name:       "text-embedding-ada-002",
		Dimensions: 1536,
		Provider:   "openai",
		MaxTokens:  8191,
		OSS:        false,
	},
	"nomic-embed-text": {
		Name:       "nomic-embed-text",
		Dimensions: 768,
		Provider:   "ollama",
		MaxTokens:  8192,
		OSS:        true,
	},
	"mxbai-embed-large": {
		Name:       "mxbai-embed-large",
		Dimensions: 1024,
		Provider:   "ollama",
		MaxTokens:  512,
		OSS:        true,
	},
	"embed-english-v3.0": {
		Name:       "embed-english-v3.0",
		Dimensions: 1024,
		Provider:   "cohere",
		MaxTokens:  512,
		OSS:        false,
	},
	"embed-multilingual-v3.0": {
		Name:       "embed-multilingual-v3.0",
		Dimensions: 1024,
		Provider:   "cohere",
		MaxTokens:  512,
		OSS:        false,
	},

	// Aliases for backward compatibility
	"all-mpnet-base-v2": {
		Name:       "all-mpnet-base-v2",
		Dimensions: 768,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
	"all-minilm-l6-v2": {
		Name:       "all-MiniLM-L6-v2",
		Dimensions: 384,
		Provider:   "sentence-transformers",
		MaxTokens:  512,
		OSS:        true,
	},
}

// GetModelDimensions returns the vector dimensions for a known embedding model
func GetModelDimensions(modelName string) (int, error) {
	info, err := GetModelInfo(modelName)
	if err != nil {
		return 0, err
	}
	return info.Dimensions, nil
}

// GetModelInfo returns detailed information about a known embedding model
func GetModelInfo(modelName string) (*ModelInfo, error) {
	// Normalize model name (case-insensitive, trim spaces)
	normalized := strings.ToLower(strings.TrimSpace(modelName))

	// Direct lookup
	if info, ok := knownModels[normalized]; ok {
		return &info, nil
	}

	// Try without provider prefix (e.g., "all-mpnet-base-v2" instead of "sentence-transformers/all-mpnet-base-v2")
	if strings.Contains(normalized, "/") {
		parts := strings.Split(normalized, "/")
		if len(parts) == 2 {
			if info, ok := knownModels[parts[1]]; ok {
				return &info, nil
			}
		}
	}

	// Not found
	return nil, fmt.Errorf("unknown embedding model: %s", modelName)
}

// IsKnownModel returns true if the model is in the registry
func IsKnownModel(modelName string) bool {
	_, err := GetModelInfo(modelName)
	return err == nil
}

// ListKnownModels returns a list of all known models
func ListKnownModels() []ModelInfo {
	seen := make(map[string]bool)
	var models []ModelInfo

	for key, info := range knownModels {
		// Skip aliases (models without provider prefix)
		if !strings.Contains(key, "/") {
			continue
		}

		// Skip duplicates
		if seen[info.Name] {
			continue
		}
		seen[info.Name] = true

		models = append(models, info)
	}

	return models
}

// ListModelsByProvider returns models from a specific provider
func ListModelsByProvider(provider string) []ModelInfo {
	var models []ModelInfo
	seen := make(map[string]bool)

	providerLower := strings.ToLower(provider)

	for key, info := range knownModels {
		// Skip aliases
		if !strings.Contains(key, "/") {
			continue
		}

		// Skip duplicates
		if seen[info.Name] {
			continue
		}

		if strings.ToLower(info.Provider) == providerLower {
			seen[info.Name] = true
			models = append(models, info)
		}
	}

	return models
}

// GetOSSModels returns all open-source embedding models
func GetOSSModels() []ModelInfo {
	var models []ModelInfo
	seen := make(map[string]bool)

	for key, info := range knownModels {
		// Skip aliases
		if !strings.Contains(key, "/") {
			continue
		}

		// Skip duplicates
		if seen[info.Name] {
			continue
		}

		if info.OSS {
			seen[info.Name] = true
			models = append(models, info)
		}
	}

	return models
}
