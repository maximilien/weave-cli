// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// GetSchema retrieves the index mappings and converts to CollectionSchema
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Check if index exists
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("index %s not found", collectionName)
	}

	// Return default schema for now
	// Full schema introspection would require complex type checking
	// of Elasticsearch Property union types which is non-trivial
	schema := &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Elasticsearch uses manual embeddings
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
				Name:     "vector_field",
				DataType: []string{"vector"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}

	return schema, nil
}

// UpdateSchema updates the index mappings
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Validate schema first
	if err := a.ValidateSchema(schema); err != nil {
		return err
	}

	// Note: Elasticsearch doesn't allow updating existing field mappings on an existing index
	// You can only add new fields to the mapping
	// For full schema updates, you need to reindex

	// For now, return an informative error
	// In the future, we could implement:
	// 1. Create a new index with the new schema
	// 2. Reindex data from old to new
	// 3. Delete old index
	// 4. Create alias to point to new index

	return fmt.Errorf("UpdateSchema not fully implemented: Elasticsearch requires reindexing to change existing field mappings. You can only add new fields to an existing index")
}

// GetDefaultSchema returns the default schema for a collection
// This is already implemented in stubs.go and will remain there
// as it doesn't require any database interaction

// ValidateSchema validates a schema
// This is already implemented in stubs.go and will remain there
// as it's a simple validation function
