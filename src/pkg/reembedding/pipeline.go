// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package reembedding

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/embeddings"
	"github.com/maximilien/weave-cli/src/pkg/reembedding/providers"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// EmbeddingPipeline handles re-generating embeddings for documents
type EmbeddingPipeline struct {
	embeddingModel string
	provider       string
	dimensions     int
	embProvider    providers.EmbeddingProvider
}

// NewEmbeddingPipeline creates a new embedding pipeline with auto-detected dimensions
func NewEmbeddingPipeline(modelName string) (*EmbeddingPipeline, error) {
	if modelName == "" {
		return nil, fmt.Errorf("embedding model name cannot be empty")
	}

	// Auto-detect dimensions using model registry
	dims, err := embeddings.GetModelDimensions(modelName)
	if err != nil {
		return nil, fmt.Errorf("unknown embedding model %s: %w (run 'weave embeddings list' to see supported models)", modelName, err)
	}

	info, err := embeddings.GetModelInfo(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}

	// Create embedding provider
	ctx := context.Background()
	embProvider, err := providers.CreateProvider(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}

	return &EmbeddingPipeline{
		embeddingModel: modelName,
		provider:       info.Provider,
		dimensions:     dims,
		embProvider:    embProvider,
	}, nil
}

// ProcessBatch generates new embeddings for a batch of documents
func (p *EmbeddingPipeline) ProcessBatch(ctx context.Context, docs []*vectordb.Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Check if provider is available
	if p.embProvider == nil {
		return fmt.Errorf("embedding provider not initialized")
	}

	// Collect texts from documents
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			texts = append(texts, "") // Placeholder for nil docs
			continue
		}

		text := doc.Text
		if text == "" {
			text = doc.Content // Fallback to content if text is empty
		}

		texts = append(texts, text)
	}

	// Generate embeddings in batch
	embeddings, err := p.embProvider.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Validate and store embeddings in documents
	emptyTextCount := 0
	zeroVectorCount := 0
	normalEmbeddingCount := 0

	for i, doc := range docs {
		if doc == nil {
			continue
		}

		embedding := embeddings[i]

		// Handle empty text -> create zero vector of correct dimensions
		if texts[i] == "" || len(embedding) == 0 {
			// Create zero vector for empty/nil embeddings
			doc.Embedding = make([]float64, p.dimensions)
			if texts[i] == "" {
				emptyTextCount++
			}
			zeroVectorCount++
			continue
		}

		// Validate embedding dimensions
		if len(embedding) != p.dimensions {
			return fmt.Errorf("embedding dimension mismatch for document %s: expected %d, got %d", doc.ID, p.dimensions, len(embedding))
		}

		// Store embedding in document (VDB will use this if present)
		doc.Embedding = embedding
		normalEmbeddingCount++
	}

	fmt.Fprintf(os.Stderr, "DEBUG Pipeline: Batch size=%d, Empty text=%d, Zero vectors created=%d, Normal embeddings=%d\n",
		len(docs), emptyTextCount, zeroVectorCount, normalEmbeddingCount)

	return nil
}

// GetDimensions returns the vector dimensions for this pipeline
func (p *EmbeddingPipeline) GetDimensions() int {
	return p.dimensions
}

// GetModelName returns the embedding model name
func (p *EmbeddingPipeline) GetModelName() string {
	return p.embeddingModel
}

// GetProvider returns the embedding provider
func (p *EmbeddingPipeline) GetProvider() string {
	return p.provider
}

// IsSupported returns true if the model is supported for re-embedding
func (p *EmbeddingPipeline) IsSupported() bool {
	// Check if provider is available
	ctx := context.Background()
	return p.embProvider.IsAvailable(ctx) == nil
}
