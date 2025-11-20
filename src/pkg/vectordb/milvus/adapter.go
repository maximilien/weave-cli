// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Milvus client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Milvus adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	milvusConfig := &Config{
		Address:          config.Address,
		Username:         config.Username,
		Password:         config.Password,
		APIKey:           config.APIKey,
		Database:         config.Database,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	client, err := NewClient(milvusConfig)
	if err != nil {
		return nil, err
	}

	// Create LLM client for embeddings (optional, only if OpenAI API key is available)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		var err error
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail - embeddings won't work but other operations will
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{
		Client:    client,
		llmClient: llmClient,
	}, nil
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	vectorizer := "text-embedding-3-small"
	return &vectordb.CollectionSchema{
		Vectorizer: vectorizer,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	// Milvus requires explicit schemas, so validation is minimal
	// The actual schema is created in CreateCollection
	return nil
}

// UpdateSchema updates a collection's schema
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Milvus doesn't support dynamic schema updates
	// Schema is defined at collection creation time
	return fmt.Errorf("Milvus does not support schema updates after collection creation")
}
