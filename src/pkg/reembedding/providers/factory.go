// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/embeddings"
)

// CreateProvider creates an embedding provider based on the model name
func CreateProvider(ctx context.Context, modelName string) (EmbeddingProvider, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	// Get model info from registry
	info, err := embeddings.GetModelInfo(modelName)
	if err != nil {
		return nil, fmt.Errorf("unknown embedding model %s: %w", modelName, err)
	}

	var provider EmbeddingProvider

	// Create provider based on model provider
	switch info.Provider {
	case "sentence-transformers":
		provider, err = NewSentenceTransformersProvider(modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to create sentence-transformers provider: %w", err)
		}

	case "openai":
		// Strip openai/ prefix if present
		modelName = strings.TrimPrefix(modelName, "openai/")
		provider, err = NewOpenAIProvider(modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}

	case "ollama":
		return nil, fmt.Errorf("Ollama provider not yet implemented (coming in v0.9.19)")

	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", info.Provider)
	}

	// Check if provider is available
	if err := provider.IsAvailable(ctx); err != nil {
		return nil, fmt.Errorf("provider %s not available: %w", info.Provider, err)
	}

	return provider, nil
}
