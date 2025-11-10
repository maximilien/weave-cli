// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// GetSchema retrieves the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	schema, err := a.client.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, a.wrapError(err, "get schema")
	}
	return a.convertSchemaFromWeaviate(schema), nil
}

// UpdateSchema updates the schema for a collection
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Weaviate doesn't support schema updates after creation, so we'll return an error
	return vectordb.ErrUnsupported("schema updates are not supported in Weaviate")
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	// Create a default schema based on the type
	vectorizer := "text2vec-openai"
	properties := []vectordb.SchemaProperty{
		{
			Name:     "content",
			DataType: []string{"text"},
		},
		{
			Name:     "text",
			DataType: []string{"text"},
		},
		{
			Name:     "url",
			DataType: []string{"text"},
		},
		{
			Name:     "metadata",
			DataType: []string{"object"},
		},
	}

	if schemaType == vectordb.SchemaTypeImage {
		vectorizer = "none"
		properties = append(properties, vectordb.SchemaProperty{
			Name:     "image",
			DataType: []string{"text"},
		})
		properties = append(properties, vectordb.SchemaProperty{
			Name:     "image_data",
			DataType: []string{"text"},
		})
	}

	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: vectorizer,
		Properties: properties,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return vectordb.ErrInvalidSchema("schema cannot be nil")
	}

	if schema.Class == "" {
		return vectordb.ErrInvalidSchema("schema class name cannot be empty")
	}

	if len(schema.Properties) == 0 {
		return vectordb.ErrInvalidSchema("schema must have at least one property")
	}

	// Validate each property
	for _, prop := range schema.Properties {
		if prop.Name == "" {
			return vectordb.ErrInvalidSchema("property name cannot be empty")
		}

		if len(prop.DataType) == 0 {
			return vectordb.ErrInvalidSchema("property must have at least one data type")
		}

		// Validate nested properties recursively
		if err := a.validateNestedProperties(prop.NestedProperties); err != nil {
			return err
		}
	}

	return nil
}

// validateNestedProperties validates nested properties recursively
func (a *Adapter) validateNestedProperties(props []vectordb.SchemaProperty) error {
	for _, prop := range props {
		if prop.Name == "" {
			return vectordb.ErrInvalidSchema("nested property name cannot be empty")
		}

		if len(prop.DataType) == 0 {
			return vectordb.ErrInvalidSchema("nested property must have at least one data type")
		}

		// Recursively validate nested properties
		if err := a.validateNestedProperties(prop.NestedProperties); err != nil {
			return err
		}
	}

	return nil
}
