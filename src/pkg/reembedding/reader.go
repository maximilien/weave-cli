// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CollectionReader reads documents from a source collection in paginated batches
type CollectionReader struct {
	client         vectordb.VectorDBClient
	collectionName string
	batchSize      int
	offset         int
	totalDocs      int64
}

// NewCollectionReader creates a new CollectionReader for the given collection
func NewCollectionReader(client vectordb.VectorDBClient, collectionName string, batchSize int) (*CollectionReader, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	if collectionName == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Get total document count
	ctx := context.Background()
	totalDocs, err := client.GetCollectionCount(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection count: %w", err)
	}

	return &CollectionReader{
		client:         client,
		collectionName: collectionName,
		batchSize:      batchSize,
		offset:         0,
		totalDocs:      totalDocs,
	}, nil
}

// ReadBatch reads the next batch of documents from the collection
func (r *CollectionReader) ReadBatch(ctx context.Context) ([]*vectordb.Document, error) {
	if !r.HasMore() {
		return nil, nil
	}

	docs, err := r.client.ListDocuments(ctx, r.collectionName, r.batchSize, r.offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch at offset %d: %w", r.offset, err)
	}

	r.offset += len(docs)
	return docs, nil
}

// HasMore returns true if there are more documents to read
func (r *CollectionReader) HasMore() bool {
	return r.offset < int(r.totalDocs)
}

// Progress returns the current progress (documents read, total documents)
func (r *CollectionReader) Progress() (int, int64) {
	return r.offset, r.totalDocs
}

// Reset resets the reader to the beginning of the collection
func (r *CollectionReader) Reset() {
	r.offset = 0
}
