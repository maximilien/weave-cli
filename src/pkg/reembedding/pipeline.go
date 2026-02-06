// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/embeddings"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// EmbeddingPipeline handles re-generating embeddings for documents
type EmbeddingPipeline struct {
	embeddingModel string
	provider       string
	dimensions     int
	llmClient      *llm.OpenAIClient
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

	// Create LLM client for embedding generation (only for OpenAI models for now)
	var llmClient *llm.OpenAIClient
	if info.Provider == "openai" {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set (required for model %s)", modelName)
		}

		llmClient, err = llm.NewOpenAIClient(apiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
		}
	}

	return &EmbeddingPipeline{
		embeddingModel: modelName,
		provider:       info.Provider,
		dimensions:     dims,
		llmClient:      llmClient,
	}, nil
}

// ProcessBatch generates new embeddings for a batch of documents
func (p *EmbeddingPipeline) ProcessBatch(ctx context.Context, docs []*vectordb.Document) error {
	if len(docs) == 0 {
		return nil
	}

	// For now, only support OpenAI embeddings
	// TODO: Add support for sentence-transformers and Ollama
	if p.provider != "openai" {
		return fmt.Errorf("embedding provider '%s' not yet supported for re-embedding (only 'openai' is currently supported)", p.provider)
	}

	for _, doc := range docs {
		if doc == nil {
			continue
		}

		// Generate embedding for document text
		text := doc.Text
		if text == "" {
			text = doc.Content // Fallback to content if text is empty
		}

		if text == "" {
			continue // Skip documents with no text
		}

		embedding, err := p.llmClient.GenerateEmbedding(ctx, text, p.embeddingModel)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for document %s: %w", doc.ID, err)
		}

		// Validate embedding dimensions
		if len(embedding) != p.dimensions {
			return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", p.dimensions, len(embedding))
		}

		// Note: The embedding is generated but we need to store it in the document
		// This is handled by the VectorDB client when creating the document
		// For now, we just validate that we can generate embeddings successfully
	}

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
	// For now, only OpenAI is supported
	// TODO: Add support for sentence-transformers and Ollama
	return p.provider == "openai"
}
