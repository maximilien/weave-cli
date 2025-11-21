// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"
	"time"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Client wraps the Chroma client with vector database functionality
type Client struct {
	client *chroma.Client
	config *Config
}

// Config holds Chroma client configuration
type Config struct {
	// Connection settings
	URL      string // Chroma server URL (e.g., http://localhost:8000)
	APIKey   string // API key for Chroma Cloud
	Tenant   string // Tenant name (default: "default_tenant")
	Database string // Database name (default: "default_database")
	Timeout  int    // Timeout in seconds for operations (default: 10)

	// Vector settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536 for OpenAI)
	SimilarityMetric string // Distance metric: "l2", "ip", or "cosine"
}

// NewClient creates a new Chroma client
func NewClient(config *Config) (*Client, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("Chroma URL is required")
	}

	// Set defaults
	if config.Tenant == "" {
		config.Tenant = "default_tenant"
	}
	if config.Database == "" {
		config.Database = "default_database"
	}
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536 // OpenAI text-embedding-3-small default
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "cosine"
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}

	// Create Chroma client
	var client *chroma.Client
	var err error

	if config.APIKey != "" {
		// Chroma Cloud with API key
		client, err = chroma.NewClient(
			chroma.WithBasePath(config.URL),
			chroma.WithAuth(types.NewTokenAuthCredentialsProvider(config.APIKey, types.AuthorizationTokenHeader)),
			chroma.WithTenant(config.Tenant),
			chroma.WithDatabase(config.Database),
		)
	} else {
		// Local Chroma
		client, err = chroma.NewClient(
			chroma.WithBasePath(config.URL),
			chroma.WithTenant(config.Tenant),
			chroma.WithDatabase(config.Database),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Chroma client: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// getTimeout returns the timeout duration for operations
func (c *Client) getTimeout() time.Duration {
	timeout := c.config.Timeout
	if timeout == 0 {
		timeout = 10
	}
	return time.Duration(timeout) * time.Second
}

// Health checks the health of the Chroma instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Use heartbeat to check health
	_, err := c.client.Heartbeat(ctx)
	if err != nil {
		return fmt.Errorf("Chroma health check failed: %w", err)
	}

	return nil
}

// Close closes the Chroma client connection
func (c *Client) Close(ctx context.Context) error {
	// Chroma Go client doesn't have a close method
	return nil
}

// GetDistanceFunction returns the Chroma distance function for the configured similarity metric
func (c *Client) GetDistanceFunction() types.DistanceFunction {
	switch c.config.SimilarityMetric {
	case "l2", "euclidean":
		return types.L2
	case "ip", "inner_product":
		return types.IP
	case "cosine":
		return types.COSINE
	default:
		return types.COSINE
	}
}

// CreateCollection creates a new collection with the given schema
func (c *Client) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Build metadata from schema
	metadata := make(map[string]interface{})
	if schema != nil && schema.Vectorizer != "" {
		metadata["vectorizer"] = schema.Vectorizer
	}

	// Create collection with distance function
	// Use NewConsistentHashEmbeddingFunction to avoid loading ONNX models
	_, err := c.client.CreateCollection(
		ctx,
		name,
		map[string]interface{}{"hnsw:space": string(c.GetDistanceFunction())},
		true, // createOrGet
		types.NewConsistentHashEmbeddingFunction(), // Placeholder - we'll provide embeddings
		c.GetDistanceFunction(),
	)
	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", name, err)
	}

	return nil
}

// DeleteCollection deletes a collection and all its documents
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection first to check if it exists
	_, err := c.client.GetCollection(ctx, name, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", name)
	}

	// Delete the collection
	_, err = c.client.DeleteCollection(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", name, err)
	}

	return nil
}

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var result []vectordb.CollectionInfo
	for _, col := range collections {
		info := vectordb.CollectionInfo{
			Name: col.Name,
		}

		// Get count for each collection
		collection, err := c.client.GetCollection(ctx, col.Name, nil)
		if err == nil {
			count, err := collection.Count(ctx)
			if err == nil {
				info.Count = int64(count)
			}
		}

		// Extract vectorizer from metadata if available
		if col.Metadata != nil {
			if v, ok := col.Metadata["vectorizer"].(string); ok {
				info.Vectorizer = v
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check collection exists: %w", err)
	}

	for _, col := range collections {
		if col.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetCollectionCount returns the number of documents in a collection
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection, err := c.client.GetCollection(ctx, name, nil)
	if err != nil {
		return 0, vectordb.ErrNotFound("collection", name)
	}

	count, err := collection.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection count: %w", err)
	}

	return int64(count), nil
}
