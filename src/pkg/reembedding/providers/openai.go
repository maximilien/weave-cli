// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// OpenAIProvider implements EmbeddingProvider for OpenAI models
type OpenAIProvider struct {
	modelName string
	client    *llm.OpenAIClient
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(modelName string) (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	return &OpenAIProvider{
		modelName: modelName,
		client:    client,
	}, nil
}

// GenerateEmbedding generates an embedding for a single text
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	return p.client.GenerateEmbedding(ctx, text, p.modelName)
}

// GenerateEmbeddings generates embeddings for multiple texts (batch)
func (p *OpenAIProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		embedding, err := p.client.GenerateEmbedding(ctx, text, p.modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetModelName returns the model name
func (p *OpenAIProvider) GetModelName() string {
	return p.modelName
}

// GetProvider returns the provider name
func (p *OpenAIProvider) GetProvider() string {
	return "openai"
}

// IsAvailable checks if OpenAI is available (API key set)
func (p *OpenAIProvider) IsAvailable(ctx context.Context) error {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	return nil
}
