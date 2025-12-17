// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	opensearch "github.com/opensearch-project/opensearch-go/v4"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the OpenSearch client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	client    *opensearchapi.Client
	config    *vectordb.Config
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new OpenSearch adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Determine URL/address
	url := config.URL
	if url == "" {
		url = config.Address
	}
	if url == "" {
		url = "http://localhost:9200"
	}

	// Parse addresses (support comma-separated)
	addresses := strings.Split(url, ",")
	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}

	// Create HTTP transport with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	_ = timeout // Will be used when we implement context timeouts

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Type == vectordb.VectorDBTypeOpenSearchLocal,
		},
	}

	// Create OpenSearch client config
	osConfig := opensearch.Config{
		Addresses: addresses,
		Transport: transport,
	}

	// Add authentication for cloud
	if config.Type == vectordb.VectorDBTypeOpenSearchCloud {
		if config.Username != "" && config.Password != "" {
			osConfig.Username = config.Username
			osConfig.Password = config.Password
		} else if config.APIKey != "" {
			// API key authentication
			osConfig.Header = http.Header{
				"Authorization": []string{"ApiKey " + config.APIKey},
			}
		}
		// TODO: Add AWS Signature V4 support for AWS OpenSearch Service
	}

	// Create the OpenSearch API client
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: osConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Create LLM client for embeddings (optional)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{
		client:    client,
		config:    config,
		llmClient: llmClient,
	}, nil
}

// Health checks the health of the OpenSearch cluster
func (a *Adapter) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	resp, err := a.client.Cluster.Health(ctx, nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check cluster status
	if resp.Status == "red" {
		return fmt.Errorf("cluster status is red")
	}

	return nil
}

// getTimeout returns the configured timeout as a time.Duration
func (a *Adapter) getTimeout() time.Duration {
	if a.config.Timeout > 0 {
		return time.Duration(a.config.Timeout) * time.Second
	}
	return 30 * time.Second // Default timeout
}

// GetSchema returns the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// Check if collection exists first
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("collection not found: %s", collectionName)
	}

	// Return basic schema
	// OpenSearch schemas are defined at index creation time
	// For simplicity, return the default schema structure
	schema := &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "text-embedding-3-small", // Default vectorizer
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "vector",
				DataType: []string{"knn_vector"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}

	return schema, nil
}

// UpdateSchema updates the schema for a collection
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// OpenSearch allows adding new fields via Put Mapping API
	// However, existing fields cannot be modified (similar to Elasticsearch)
	// For simplicity, return error suggesting recreation
	return fmt.Errorf("OpenSearch does not support schema updates for existing fields - delete and recreate the collection instead")
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class: collectionName,
	}
}

// ValidateSchema validates if a schema is compatible
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema.Class == "" {
		return fmt.Errorf("collection name is required")
	}
	return nil
}

// Close closes the OpenSearch client connection
func (a *Adapter) Close() error {
	// TODO: Implement cleanup if needed
	return nil
}
