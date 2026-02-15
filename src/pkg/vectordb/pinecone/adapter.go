// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/reembedding/providers"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

// Adapter wraps the Pinecone client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	client    *pinecone.Client
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

	// Initialize Pinecone client
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	return &Adapter{
		client:    pc,
		config:    config,
		llmClient: llmClient,
		apiKey:    apiKey,
		host:      host,
	}, nil
}

// Health checks the health of the Pinecone connection
func (a *Adapter) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	if a.client == nil {
		return fmt.Errorf("Pinecone client not initialized")
	}

	// Try to list indexes as a health check
	_, err := a.client.ListIndexes(ctx)
	if err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") || strings.Contains(errMsg, "authentication") {
			return fmt.Errorf("Pinecone health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid or missing PINECONE_API_KEY\n"+
				"  2. API key not active in Pinecone console\n"+
				"  3. Wrong environment specified\n"+
				"  → Verify API key at https://app.pinecone.io", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Pinecone health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. Network connectivity issues\n"+
				"  2. Firewall blocking outbound HTTPS\n"+
				"  3. Pinecone service temporarily unavailable\n"+
				"  → Check status at https://status.pinecone.io", err)
		}
		return fmt.Errorf("Pinecone health check failed: %w", err)
	}

	return nil
}

// Close closes the Pinecone client connection
func (a *Adapter) Close() error {
	// Pinecone client typically doesn't need explicit closing
	return nil
}

// getTimeout returns the configured timeout as a time.Duration
func (a *Adapter) getTimeout() time.Duration {
	if a.config.Timeout > 0 {
		return time.Duration(a.config.Timeout) * time.Second
	}
	return 30 * time.Second // Default timeout
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (a *Adapter) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	isCloud := true // Pinecone is cloud-only
	return vectordb.GetTimeoutForOperation(opType, isCloud, a.config.Timeout)
}

// createEmbeddingProvider creates an embedding provider for OSS models
func (a *Adapter) createEmbeddingProvider(ctx context.Context, modelName string) (providers.EmbeddingProvider, error) {
	return providers.CreateProvider(ctx, modelName)
}
