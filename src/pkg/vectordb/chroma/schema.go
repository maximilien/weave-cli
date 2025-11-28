//go:build (darwin && amd64) || (darwin && arm64)
// +build darwin,amd64 darwin,arm64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// GetSchema retrieves the schema for a collection
// Note: Chroma is schema-less, so we return a default schema
func (c *Client) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// Chroma doesn't have explicit schemas, return a default schema
	return &vectordb.CollectionSchema{
		Class: collectionName,
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
// Note: Chroma is schema-less, so this is a no-op
func (c *Client) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Chroma doesn't require explicit schema updates
	return nil
}

// GetDefaultSchema returns a default schema for the given type
func (c *Client) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	switch schemaType {
	case vectordb.SchemaTypeText:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "none", // Chroma uses embedding functions
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
			Vectorizer: "none", // Chroma uses embedding functions
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

	// Chroma schemas are flexible
	return nil
}
