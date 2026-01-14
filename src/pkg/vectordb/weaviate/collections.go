// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a new collection with the given schema
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, a.client.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Convert schema to field definitions
	var fields []FieldDefinition
	for _, prop := range schema.Properties {
		fields = append(fields, FieldDefinition{
			Name: prop.Name,
			Type: strings.Join(prop.DataType, ","),
		})
	}

	embeddingModel := "text-embedding-ada-002" // Default model
	if schema.Vectorizer != "" {
		embeddingModel = schema.Vectorizer
	}

	if err := a.client.CreateCollection(ctx, name, embeddingModel, fields); err != nil {
		return a.wrapError(err, "create collection")
	}
	return nil
}

// DeleteCollection deletes a collection and all its documents
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, a.client.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	if err := a.client.DeleteCollection(ctx, name); err != nil {
		return a.wrapError(err, "delete collection")
	}
	return nil
}

// ListCollections returns a list of all collections
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, a.client.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	collections, err := a.client.ListCollections(ctx)
	if err != nil {
		return nil, a.wrapError(err, "list collections")
	}

	result := make([]vectordb.CollectionInfo, len(collections))
	for i, collectionName := range collections {
		// Get collection count - we'll implement a simple count method
		count := int64(0) // Default to 0 since GetCollectionCount doesn't exist

		// Get vectorizer from schema
		vectorizer := ""
		if schema, err := a.GetSchema(ctx, collectionName); err == nil && schema != nil {
			vectorizer = schema.Vectorizer
		}

		result[i] = vectordb.CollectionInfo{
			Name:        collectionName,
			Description: "Weaviate collection",
			Count:       count,
			Vectorizer:  vectorizer,
		}
	}

	return result, nil
}

// CollectionExists checks if a collection exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, a.client.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	collections, err := a.client.ListCollections(ctx)
	if err != nil {
		return false, a.wrapError(err, "check collection exists")
	}

	for _, collection := range collections {
		if collection == name {
			return true, nil
		}
	}

	return false, nil
}

// GetCollectionCount returns the number of documents in a collection
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, a.client.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Use Weaviate's GraphQL Aggregate API to get count without pagination limits
	count, err := a.client.GetCollectionCount(ctx, name)
	if err != nil {
		return 0, a.wrapError(err, "get collection count")
	}
	return count, nil
}
