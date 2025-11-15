// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the MongoDB client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new MongoDB adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	mongoConfig := &Config{
		URI:              config.URL,
		Database:         config.Database,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	client, err := NewClient(mongoConfig)
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
