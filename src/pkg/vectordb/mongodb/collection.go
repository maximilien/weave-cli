// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a new collection with Atlas Vector Search index
func (c *Client) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Create the collection if it doesn't exist
	err := c.database.CreateCollection(ctx, name)
	if err != nil {
		// Ignore error if collection already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	// Create indexes
	collection := c.getCollection(name)

	// Create unique index on document_id
	unique := true
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "document_id", Value: 1}},
		Options: &options.IndexOptions{Unique: &unique},
	})
	if err != nil {
		return fmt.Errorf("failed to create document_id index: %w", err)
	}

	// Create text index for BM25 search if needed
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "text", Value: "text"},
			{Key: "content", Value: "text"},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create text index: %w", err)
	}

	// Store embedding model metadata as a special document
	// MongoDB doesn't have collection-level metadata, so we use a reserved document
	metadataDoc := bson.M{
		"document_id":       "_weave_metadata",
		"vector_dimensions": c.config.VectorDimensions,
		"similarity_metric": c.config.SimilarityMetric,
		"is_metadata":       true,
	}

	// Insert or update metadata document
	filter := bson.M{"document_id": "_weave_metadata"}
	update := bson.M{"$set": metadataDoc}
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		// Don't fail collection creation if metadata insert fails
		fmt.Printf("Warning: Failed to store embedding metadata: %v\n", err)
	}

	// Create vector search index
	// Note: This requires MongoDB Atlas and must be done via Atlas UI or Admin API
	// We'll document this requirement and provide instructions
	return nil
}

// DeleteCollection deletes a collection and all its documents
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	err := c.getCollection(name).Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	names, err := c.database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	collections := make([]vectordb.CollectionInfo, 0, len(names))
	for _, name := range names {
		count, err := c.GetCollectionCount(ctx, name)
		if err != nil {
			// Log error but continue
			count = 0
		}

		// Get vectorizer from schema
		vectorizer := ""
		if schema, err := c.GetSchema(ctx, name); err == nil && schema != nil {
			vectorizer = schema.Vectorizer
		}

		collections = append(collections, vectordb.CollectionInfo{
			Name:       name,
			Count:      count,
			Vectorizer: vectorizer,
		})
	}

	return collections, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	names, err := c.database.ListCollectionNames(ctx, bson.D{{Key: "name", Value: name}})
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return len(names) > 0, nil
}

// GetCollectionCount returns the number of documents in a collection
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeCollection))
	defer cancel()

	// Exclude metadata document from count
	filter := bson.M{"document_id": bson.M{"$ne": "_weave_metadata"}}
	count, err := c.getCollection(name).CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection count: %w", err)
	}

	return count, nil
}

// getCollectionDimensions retrieves the vector dimensions from collection metadata
func (c *Client) getCollectionDimensions(ctx context.Context, name string) (int, error) {
	collection := c.getCollection(name)

	// Try to find the metadata document
	filter := bson.M{"document_id": "_weave_metadata"}
	var result struct {
		VectorDimensions int `bson:"vector_dimensions"`
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		// If metadata document doesn't exist, fall back to config default
		return c.config.VectorDimensions, nil
	}

	if result.VectorDimensions > 0 {
		return result.VectorDimensions, nil
	}

	return c.config.VectorDimensions, nil
}
