//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// noopEmbeddingFunction is a simple embedding function that doesn't use tokenizers
// This avoids the libtokenizers CGO dependency issues from default_ef
// It generates dummy embeddings of the correct dimensions
type noopEmbeddingFunction struct {
	dimensions int
}

func (n *noopEmbeddingFunction) EmbedDocuments(ctx context.Context, documents []string) ([]embeddings.Embedding, error) {
	// Generate dummy embeddings for each document
	embeds := make([]embeddings.Embedding, len(documents))
	for i := range documents {
		embeds[i] = n.generateDummyEmbedding()
	}
	return embeds, nil
}

func (n *noopEmbeddingFunction) EmbedQuery(ctx context.Context, query string) (embeddings.Embedding, error) {
	// Generate a dummy embedding for the query
	return n.generateDummyEmbedding(), nil
}

func (n *noopEmbeddingFunction) EmbedRecords(ctx context.Context, records []*chroma.Record, force bool) error {
	// EmbedDocuments and EmbedQuery handle embedding generation
	// This method is called by the SDK - we don't need to do anything here
	// The SDK will use EmbedDocuments for batch operations
	return nil
}

// generateDummyEmbedding creates a dummy embedding vector of the correct dimensions
func (n *noopEmbeddingFunction) generateDummyEmbedding() embeddings.Embedding {
	dims := n.dimensions
	if dims == 0 {
		dims = 1536 // Default to OpenAI text-embedding-3-small dimensions
	}
	// Create a vector of zeros (simplest dummy embedding)
	vec := make([]float32, dims)
	return embeddings.NewEmbeddingFromFloat32(vec)
}

// Client wraps the Chroma client with vector database functionality
type Client struct {
	client chroma.Client
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
	// For local client, URL is required. For cloud client (with API key), URL is optional.
	if config.APIKey == "" && config.URL == "" {
		return nil, fmt.Errorf("Chroma URL is required for local client")
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

	// Create Chroma client using v2 API
	var client chroma.Client
	var err error

	// Use cloud client for cloud configurations, HTTP client for local
	if config.APIKey != "" {
		// Cloud client - automatically uses https://api.trychroma.com
		// The cloud client reads tenant/database from CHROMA_TENANT and CHROMA_DATABASE env vars
		// We pass them explicitly if they're set in config and not defaults
		opts := []chroma.ClientOption{
			chroma.WithCloudAPIKey(config.APIKey),
		}

		// Only pass tenant/database if they're non-default values
		if config.Tenant != "" && config.Tenant != "default_tenant" &&
			config.Database != "" && config.Database != "default_database" {
			opts = append(opts, chroma.WithDatabaseAndTenant(config.Database, config.Tenant))
		}

		client, err = chroma.NewCloudClient(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create Chroma cloud client: %w", err)
		}
	} else {
		// Local HTTP client
		opts := []chroma.ClientOption{
			chroma.WithBaseURL(config.URL),
			chroma.WithDatabaseAndTenant(config.Database, config.Tenant),
		}
		client, err = chroma.NewHTTPClient(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create Chroma HTTP client: %w", err)
		}
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

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	// Detect cloud: APIKey indicates Chroma Cloud deployment
	isCloud := c.config.APIKey != ""
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}

// getCollection retrieves a collection by name with the required embedding function
func (c *Client) getCollection(ctx context.Context, name string) (chroma.Collection, error) {
	// Chroma v2 requires an embedding function for GetCollection
	// Use noopEmbeddingFunction to avoid tokenizer CGO dependencies
	// It generates dummy embeddings of the correct dimensions
	return c.client.GetCollection(ctx, name, chroma.WithEmbeddingFunctionGet(&noopEmbeddingFunction{dimensions: c.config.VectorDimensions}))
}

// Health checks the health of the Chroma instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Use heartbeat to check health
	err := c.client.Heartbeat(ctx)
	if err != nil {
		return fmt.Errorf("Chroma health check failed: %w", err)
	}

	return nil
}

// Close closes the Chroma client connection
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetDistanceFunction returns the distance function string for the configured similarity metric
func (c *Client) GetDistanceFunction() string {
	switch c.config.SimilarityMetric {
	case "l2", "euclidean":
		return "l2"
	case "ip", "inner_product":
		return "ip"
	case "cosine":
		return "cosine"
	default:
		return "cosine"
	}
}

// CreateCollection creates a new collection with the given schema
func (c *Client) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Build metadata options
	var opts []chroma.CreateCollectionOption

	// Add distance function metadata
	opts = append(opts, chroma.WithCollectionMetadataCreate(
		chroma.NewMetadata(
			chroma.NewStringAttribute("hnsw:space", c.GetDistanceFunction()),
		),
	))

	// Add noopEmbeddingFunction to avoid tokenizer CGO dependencies
	// It generates dummy embeddings of the correct dimensions
	opts = append(opts, chroma.WithEmbeddingFunctionCreate(&noopEmbeddingFunction{dimensions: c.config.VectorDimensions}))

	// Create collection
	_, err := c.client.GetOrCreateCollection(ctx, name, opts...)
	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", name, err)
	}

	return nil
}

// DeleteCollection deletes a collection and all its documents
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Check if collection exists first
	exists, err := c.CollectionExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", name)
	}

	// Delete the collection
	err = c.client.DeleteCollection(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", name, err)
	}

	return nil
}

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var result []vectordb.CollectionInfo
	for _, col := range collections {
		info := vectordb.CollectionInfo{
			Name: col.Name(),
		}

		// Get count for each collection (ignore errors)
		count, _ := col.Count(ctx)
		info.Count = int64(count)

		// Extract vectorizer from metadata if available
		metadata := col.Metadata()
		if metadata != nil {
			if v, ok := metadata.GetString("vectorizer"); ok {
				info.Vectorizer = v
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check collection exists: %w", err)
	}

	for _, col := range collections {
		if col.Name() == name {
			return true, nil
		}
	}

	return false, nil
}

// GetCollectionCount returns the number of documents in a collection
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	collection, err := c.getCollection(ctx, name)
	if err != nil {
		return 0, vectordb.ErrNotFound("collection", name)
	}

	count, err := collection.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection count: %w", err)
	}

	return int64(count), nil
}
