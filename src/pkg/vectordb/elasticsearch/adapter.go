// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Elasticsearch client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Elasticsearch adapter from vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Map vectordb.Config to Elasticsearch Config
	esConfig := &Config{
		CloudID:          extractCloudID(config),
		APIKey:           config.APIKey,
		Addresses:        extractAddresses(config),
		Username:         config.Username,
		Password:         config.Password,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	// Create Elasticsearch client
	client, err := NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
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

// extractCloudID extracts the Elastic Cloud ID from the config
// Cloud ID is typically provided via a dedicated field or URL pattern
func extractCloudID(config *vectordb.Config) string {
	// Check if URL looks like a Cloud ID (contains ':' and base64-like string)
	if config.URL != "" && strings.Contains(config.URL, ":") &&
		!strings.HasPrefix(config.URL, "http://") &&
		!strings.HasPrefix(config.URL, "https://") {
		// This looks like a Cloud ID (format: "cluster-name:base64-encoded-data")
		return config.URL
	}
	return ""
}

// extractAddresses extracts Elasticsearch addresses from the config
// Supports both URL and Address fields
func extractAddresses(config *vectordb.Config) []string {
	var addresses []string

	// Check URL field first
	if config.URL != "" {
		// Only use as address if it looks like a URL (http:// or https://)
		if strings.HasPrefix(config.URL, "http://") || strings.HasPrefix(config.URL, "https://") {
			addresses = append(addresses, config.URL)
		}
	}

	// Check Address field
	if config.Address != "" {
		// Split by comma in case multiple addresses are provided
		for _, addr := range strings.Split(config.Address, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				addresses = append(addresses, addr)
			}
		}
	}

	return addresses
}

// Close closes the Elasticsearch adapter and cleans up resources
func (a *Adapter) Close(ctx context.Context) error {
	if a.Client != nil {
		return a.Client.Close(ctx)
	}
	return nil
}
