// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// GetSchema retrieves the schema for a collection
// Note: MongoDB is schema-less, so we return a default schema
func (c *Client) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// Get collection dimensions to infer embedding model
	dims, err := c.getCollectionDimensions(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection dimensions: %w", err)
	}

	// Infer embedding model from dimensions
	// This is a heuristic since MongoDB doesn't store model metadata
	vectorizer := inferEmbeddingModelFromDimensions(dims)

	// MongoDB doesn't have explicit schemas, return a default schema
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: vectorizer,
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "document_id",
				DataType: []string{"text"},
			},
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "embedding",
				DataType: []string{"number[]"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}, nil
}

// UpdateSchema updates the schema for a collection
// Note: MongoDB is schema-less, so this is a no-op
func (c *Client) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// MongoDB doesn't require explicit schema updates
	return nil
}

// GetDefaultSchema returns a default schema for the given type
func (c *Client) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	switch schemaType {
	case vectordb.SchemaTypeText:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none", // MongoDB requires external vectorization
			Properties: []vectordb.SchemaProperty{
				{
					Name:        "document_id",
					DataType:    []string{"text"},
					Description: "Unique document identifier",
				},
				{
					Name:        "text",
					DataType:    []string{"text"},
					Description: "Document text content",
				},
				{
					Name:        "content",
					DataType:    []string{"text"},
					Description: "Full document content",
				},
				{
					Name:        "url",
					DataType:    []string{"text"},
					Description: "Source URL",
				},
				{
					Name:        "embedding",
					DataType:    []string{"number[]"},
					Description: "Vector embedding",
				},
				{
					Name:        "metadata",
					DataType:    []string{"object"},
					Description: "Document metadata",
				},
			},
		}
	case vectordb.SchemaTypeImage:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none", // MongoDB requires external vectorization
			Properties: []vectordb.SchemaProperty{
				{
					Name:        "document_id",
					DataType:    []string{"text"},
					Description: "Unique document identifier",
				},
				{
					Name:        "image",
					DataType:    []string{"text"},
					Description: "Image URL",
				},
				{
					Name:        "image_data",
					DataType:    []string{"text"},
					Description: "Base64 encoded image data",
				},
				{
					Name:        "text",
					DataType:    []string{"text"},
					Description: "Image description or caption",
				},
				{
					Name:        "embedding",
					DataType:    []string{"number[]"},
					Description: "Vector embedding",
				},
				{
					Name:        "metadata",
					DataType:    []string{"object"},
					Description: "Image metadata",
				},
			},
		}
	default:
		return c.GetDefaultSchema(vectordb.SchemaTypeText, collectionName)
	}
}

// ValidateSchema validates a schema definition
func (c *Client) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	if schema.Class == "" {
		return fmt.Errorf("schema class name is required")
	}

	if len(schema.Properties) == 0 {
		return fmt.Errorf("schema must have at least one property")
	}

	// MongoDB schemas are flexible - embedding is only required for vector search
	// Text-only collections (BM25 search) don't need embeddings
	return nil
}

// inferEmbeddingModelFromDimensions infers the embedding model from vector dimensions
// This is a heuristic since MongoDB doesn't store model metadata
func inferEmbeddingModelFromDimensions(dims int) string {
	switch dims {
	case 768:
		// sentence-transformers/all-mpnet-base-v2 or all-MiniLM-L12-v2
		return "sentence-transformers/all-mpnet-base-v2"
	case 384:
		// sentence-transformers/all-MiniLM-L6-v2
		return "sentence-transformers/all-MiniLM-L6-v2"
	case 1536:
		// OpenAI text-embedding-3-small or text-embedding-ada-002
		return "text-embedding-3-small"
	case 3072:
		// OpenAI text-embedding-3-large
		return "text-embedding-3-large"
	case 1024:
		// Ollama nomic-embed-text
		return "nomic-embed-text"
	default:
		// Conservative fallback: OpenAI
		return "text-embedding-3-small"
	}
}
