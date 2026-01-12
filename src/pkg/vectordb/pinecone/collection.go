// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

// CreateCollection creates a new Pinecone index
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	if a.client == nil {
		return fmt.Errorf("Pinecone client not initialized")
	}

	// Validate schema
	if err := a.ValidateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Get dimension from config (default 1536 for OpenAI)
	dimension := int32(a.config.VectorDimensions)
	if dimension == 0 {
		dimension = 1536 // OpenAI text-embedding-3-small default
	}

	// Map similarity metric from config
	metric := pinecone.Cosine // Default
	if a.config.SimilarityMetric != "" {
		switch a.config.SimilarityMetric {
		case "cosine":
			metric = pinecone.Cosine
		case "euclidean", "l2":
			metric = pinecone.Euclidean
			// Pinecone SDK may not have DotProduct in older versions
			// Default to Cosine for unsupported metrics
		}
	}

	// Create serverless index (default for Pinecone)
	_, err := a.client.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      name,
		Dimension: dimension,
		Metric:    metric,
		Region:    "us-east-1", // Default region
		Cloud:     "aws",       // Default cloud
	})

	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// DeleteCollection deletes a Pinecone index
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	if a.client == nil {
		return fmt.Errorf("Pinecone client not initialized")
	}

	err := a.client.DeleteIndex(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}

	return nil
}

// CollectionExists checks if a Pinecone index exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	if a.client == nil {
		return false, fmt.Errorf("pinecone client not initialized")
	}

	// Try to describe the index - if it exists, this will succeed
	_, err := a.client.DescribeIndex(ctx, name)
	if err != nil {
		// If error contains "not found", index doesn't exist
		if fmt.Sprint(err) != "" && (fmt.Sprint(err) == "index not found" || fmt.Sprint(err) == "Index not found") {
			return false, nil
		}
		// Other errors are actual failures
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}

	return true, nil
}

// ListCollections lists all Pinecone indexes
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	// List all indexes
	indexes, err := a.client.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}

	// Convert Pinecone indexes to CollectionInfo
	collections := make([]vectordb.CollectionInfo, 0, len(indexes))
	for _, idx := range indexes {
		collections = append(collections, vectordb.CollectionInfo{
			Name: idx.Name,
			// Note: vectordb.CollectionInfo doesn't have VectorDimension or Metric fields
			// These are Pinecone-specific and would need to be added if needed
		})
	}

	return collections, nil
}

// GetCollectionCount returns the number of documents in a Pinecone index
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	if a.client == nil {
		return 0, fmt.Errorf("pinecone client not initialized")
	}

	// Describe the index to get its host
	idx, err := a.client.DescribeIndex(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to describe index: %w", err)
	}

	// Connect to the index
	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return 0, fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Get index stats
	stats, err := idxConn.DescribeIndexStats(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get index stats: %w", err)
	}

	return int64(stats.TotalVectorCount), nil
}

// GetCollectionInfo returns information about a specific Pinecone index
func (a *Adapter) GetCollectionInfo(ctx context.Context, name string) (*vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	// Describe the index
	idx, err := a.client.DescribeIndex(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	// Get vector count
	count, err := a.GetCollectionCount(ctx, name)
	if err != nil {
		// Don't fail if we can't get count, just set to 0
		count = 0
	}

	return &vectordb.CollectionInfo{
		Name:  idx.Name,
		Count: count,
	}, nil
}

// GetSchema returns the schema for a Pinecone index
func (a *Adapter) GetSchema(ctx context.Context, name string) (*vectordb.CollectionSchema, error) {
	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	// Pinecone doesn't have explicit schemas, return minimal schema with vector info
	// We don't need to describe the index here since we're just returning a generic schema
	return &vectordb.CollectionSchema{
		Class:      name,
		Vectorizer: "none", // Pinecone expects pre-computed vectors
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "vector",
				DataType: []string{"number[]"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
		// Note: vectordb.CollectionSchema doesn't have VectorIndexConfig
		// Pinecone-specific config (dimension, metric) is managed internally
	}, nil
}

// GetDefaultSchema returns a default schema for Pinecone
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Pinecone expects pre-computed vectors
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "vector",
				DataType: []string{"number[]"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}
}

// UpdateSchema updates a collection's schema
func (a *Adapter) UpdateSchema(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	// TODO: Implement schema update
	// Pinecone indexes are immutable once created - schema cannot be updated
	return fmt.Errorf("schema updates not supported by Pinecone (indexes are immutable)")
}

// ValidateSchema validates a schema for Pinecone
func (a *Adapter) ValidateSchema(s *vectordb.CollectionSchema) error {
	// Basic validation
	if s == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if s.Class == "" {
		return fmt.Errorf("schema class cannot be empty")
	}

	// Pinecone-specific validation
	// - Must have vector properties with dimensions
	// TODO: Add more specific validation when implementing
	if len(s.Properties) == 0 {
		return fmt.Errorf("schema must have at least one property")
	}

	return nil
}

// getIndexDimensions retrieves the vector dimensions from a Pinecone index
func (a *Adapter) getIndexDimensions(ctx context.Context, name string) (int, error) {
	idx, err := a.client.DescribeIndex(ctx, name)
	if err != nil {
		// Fall back to config default
		if a.config.VectorDimensions > 0 {
			return a.config.VectorDimensions, nil
		}
		return 1536, nil // OpenAI default
	}

	return int(idx.Dimension), nil
}
