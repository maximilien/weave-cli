// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Collection field names (standardized across all collections)
const (
	FieldDocumentID = "document_id"
	FieldText       = "text"
	FieldContent    = "content"
	FieldImage      = "image"
	FieldImageData  = "image_data"
	FieldURL        = "url"
	FieldEmbedding  = "embedding"
	FieldMetadata   = "metadata"
	FieldCreatedAt  = "created_at"
	FieldUpdatedAt  = "updated_at"
)

// CreateCollection creates a new Milvus collection with explicit schema
func (c *Client) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Check if collection already exists
	exists, err := c.client.HasCollection(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if exists {
		return fmt.Errorf("collection already exists: %s", name)
	}

	// Build Milvus schema with explicit fields using builder pattern
	milvusSchema := &entity.Schema{
		CollectionName: name,
		Description:    fmt.Sprintf("Collection %s for vector search", name),
		AutoID:         false, // We provide our own document IDs
		Fields: []*entity.Field{
			// Primary key field
			entity.NewField().
				WithName(FieldDocumentID).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "256").
				WithIsPrimaryKey(true),
			// Text fields
			entity.NewField().
				WithName(FieldText).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "65535"),
			entity.NewField().
				WithName(FieldContent).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "65535"),
			// Image fields
			entity.NewField().
				WithName(FieldImage).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "512"),
			entity.NewField().
				WithName(FieldImageData).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "65535"),
			// URL field
			entity.NewField().
				WithName(FieldURL).
				WithDataType(entity.FieldTypeVarChar).
				WithTypeParams("max_length", "512"),
			// Vector embedding field
			entity.NewField().
				WithName(FieldEmbedding).
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(int64(c.config.VectorDimensions)),
			// Metadata field (JSON)
			entity.NewField().
				WithName(FieldMetadata).
				WithDataType(entity.FieldTypeJSON),
			// Timestamp fields
			entity.NewField().
				WithName(FieldCreatedAt).
				WithDataType(entity.FieldTypeInt64),
			entity.NewField().
				WithName(FieldUpdatedAt).
				WithDataType(entity.FieldTypeInt64),
		},
	}

	// Create collection
	err = c.client.CreateCollection(ctx, milvusSchema, 1) // shardNum=1 for simplicity
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Create IVF_FLAT index on vector field for semantic search
	index, err := entity.NewIndexIvfFlat(c.getMetricType(), 128) // nlist=128
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	err = c.client.CreateIndex(ctx, name, FieldEmbedding, index, false)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}

	// Load collection into memory for searching
	err = c.client.LoadCollection(ctx, name, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	return nil
}

// DeleteCollection deletes a collection and all its documents
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Check if collection exists
	exists, err := c.client.HasCollection(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("collection does not exist: %s", name)
	}

	// Drop collection
	err = c.client.DropCollection(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// List all collections
	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	result := make([]vectordb.CollectionInfo, 0, len(collections))
	for _, coll := range collections {
		// Get collection statistics
		stats, err := c.client.GetCollectionStatistics(ctx, coll.Name)
		var count int64
		if err == nil {
			// Parse row count from stats
			if rowCount, ok := stats["row_count"]; ok {
				_, _ = fmt.Sscanf(rowCount, "%d", &count)
			}
		}

		// Get vectorizer from schema
		vectorizer := ""
		if schema, err := c.GetSchema(ctx, coll.Name); err == nil && schema != nil {
			vectorizer = schema.Vectorizer
		}

		result = append(result, vectordb.CollectionInfo{
			Name:       coll.Name,
			Count:      count,
			Vectorizer: vectorizer,
		})
	}

	return result, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	exists, err := c.client.HasCollection(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return exists, nil
}

// GetCollectionCount returns the number of documents in a collection
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Get collection statistics
	stats, err := c.client.GetCollectionStatistics(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection statistics: %w", err)
	}

	// Parse row count from stats
	var count int64
	if rowCount, ok := stats["row_count"]; ok {
		_, _ = fmt.Sscanf(rowCount, "%d", &count)
	}

	return count, nil
}

// GetSchema returns the schema information for a collection
func (c *Client) GetSchema(ctx context.Context, name string) (*vectordb.CollectionSchema, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeSchema))
	defer cancel()

	// Get collection schema
	coll, err := c.client.DescribeCollection(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}

	// Find embedding field to get vectorizer info
	var vectorizer string
	for _, field := range coll.Schema.Fields {
		if field.Name == FieldEmbedding {
			vectorizer = "text-embedding-3-small" // Default, we don't store this in schema
			break
		}
	}

	return &vectordb.CollectionSchema{
		Vectorizer: vectorizer,
	}, nil
}
