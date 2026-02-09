// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
)

// EmbeddingProvider defines the interface for generating embeddings
type EmbeddingProvider interface {
	// GenerateEmbedding generates an embedding for a single text
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

	// GenerateEmbeddings generates embeddings for multiple texts (batch)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error)

	// GetModelName returns the model name
	GetModelName() string

	// GetProvider returns the provider name (openai, sentence-transformers, ollama)
	GetProvider() string

	// IsAvailable checks if the provider is available/configured
	IsAvailable(ctx context.Context) error
}
