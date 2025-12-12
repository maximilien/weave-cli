// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Schema helper methods
// These are simple functions that don't require database interaction

func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Elasticsearch uses manual embeddings by default
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
		},
	}
}

func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if schema.Class == "" {
		return fmt.Errorf("schema class (collection name) is required")
	}
	return nil
}
