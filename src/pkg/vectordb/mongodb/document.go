// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (c *Client) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)
	mdoc := c.toMongoDocument(document)

	_, err := collection.InsertOne(ctx, mdoc)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// CreateDocuments creates multiple documents in batch
func (c *Client) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	if len(documents) == 0 {
		return nil
	}

	collection := c.getCollection(collectionName)
	mdocs := make([]interface{}, len(documents))
	for i, doc := range documents {
		mdocs[i] = c.toMongoDocument(doc)
	}

	_, err := collection.InsertMany(ctx, mdocs)
	if err != nil {
		return fmt.Errorf("failed to create documents: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)
	filter := bson.D{{Key: "document_id", Value: documentID}}

	var mdoc mongoDocument
	err := collection.FindOne(ctx, filter).Decode(&mdoc)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return c.fromMongoDocument(&mdoc), nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)
	filter := bson.D{{Key: "document_id", Value: document.ID}}

	mdoc := c.toMongoDocument(document)
	update := bson.D{{Key: "$set", Value: mdoc}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)
	filter := bson.D{{Key: "document_id", Value: documentID}}

	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	if len(documentIDs) == 0 {
		return nil
	}

	collection := c.getCollection(collectionName)
	filter := bson.D{{Key: "document_id", Value: bson.D{{Key: "$in", Value: documentIDs}}}}

	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (c *Client) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)

	// Build filter for metadata fields
	filter := bson.D{}
	for key, value := range metadata {
		filter = append(filter, bson.E{Key: "metadata." + key, Value: value})
	}

	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete documents by metadata: %w", err)
	}

	return nil
}

// DeleteAllDocuments deletes all documents in a collection
func (c *Client) DeleteAllDocuments(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)

	// Delete all documents using empty filter
	_, err := collection.DeleteMany(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to delete all documents: %w", err)
	}

	return nil
}

// ListDocuments returns a list of documents in a collection
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []*vectordb.Document
	for cursor.Next(ctx) {
		var mdoc mongoDocument
		if err := cursor.Decode(&mdoc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		documents = append(documents, c.fromMongoDocument(&mdoc))
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return documents, nil
}
