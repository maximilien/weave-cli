// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Pinecone client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	// TODO: Add Pinecone client once SDK is added
	// client    *pinecone.Client
	config    *vectordb.Config
	llmClient *llm.OpenAIClient
	apiKey    string
	host      string
}

// NewAdapter creates a new Pinecone adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Get API key from config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("PINECONE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY is required")
	}

	// Get host/URL
	host := config.URL
	if host == "" {
		host = os.Getenv("PINECONE_HOST")
	}
	// Host can be empty for serverless - will be determined per-index

	// Create LLM client for embeddings
	var llmClient *llm.OpenAIClient
	var err error
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	// TODO: Initialize Pinecone client once SDK is added
	// client, err := pinecone.NewClient(apiKey)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	// }

	return &Adapter{
		// client:    client,
		config:    config,
		llmClient: llmClient,
		apiKey:    apiKey,
		host:      host,
	}, nil
}

// Health checks the health of the Pinecone connection
func (a *Adapter) Health(ctx context.Context) error {
	// TODO: Implement health check using Pinecone SDK
	// For now, just check if we have API key
	if a.apiKey == "" {
		return fmt.Errorf("no API key configured")
	}
	return nil
}

// Close closes the Pinecone client connection
func (a *Adapter) Close() error {
	// Pinecone client typically doesn't need explicit closing
	return nil
}
