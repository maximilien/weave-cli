//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (c *Client) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection '%s': %w", collectionName, err)
	}

	// Prepare content - use Content or Text
	content := document.Content
	if content == "" {
		content = document.Text
	}

	// Prepare metadata map - filter out unsupported types
	// Chroma only supports string, int, float, bool in metadata
	metadataMap := make(map[string]interface{})
	if document.Metadata != nil {
		for k, v := range document.Metadata {
			// Only include primitive types that Chroma supports
			switch v.(type) {
			case string, int, int32, int64, float32, float64, bool:
				metadataMap[k] = v
				// Skip arrays, slices, maps and other complex types
			}
		}
	}
	// Add standard fields to metadata
	if document.URL != "" {
		metadataMap["url"] = document.URL
	}
	if document.Image != "" {
		metadataMap["image"] = document.Image
	}

	// Prepare metadata - ALWAYS provide it (even if empty) to keep array lengths consistent
	metadata, err := chroma.NewDocumentMetadataFromMap(metadataMap)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	// Add document to collection - wrap single values in slices and spread them
	err = collection.Add(ctx,
		chroma.WithIDs([]chroma.DocumentID{chroma.DocumentID(document.ID)}...),
		chroma.WithTexts([]string{content}...),
		chroma.WithMetadatas([]chroma.DocumentMetadata{metadata}...),
	)
	if err != nil {
		return fmt.Errorf("failed to create document %s: %w", document.ID, err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Get document by ID
	result, err := collection.Get(ctx,
		chroma.WithIDsGet(chroma.DocumentID(documentID)),
		chroma.WithIncludeGet(chroma.IncludeDocuments, chroma.IncludeMetadatas),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document %s: %w", documentID, err)
	}

	// Check if document was found using raw result data
	ids := result.GetIDs()
	if len(ids) == 0 {
		return nil, vectordb.ErrNotFound("document", documentID)
	}

	// Build document from raw result data
	doc := &vectordb.Document{
		ID: string(ids[0]),
	}

	// Get document content if available
	documents := result.GetDocuments()
	if len(documents) > 0 && documents[0] != nil {
		doc.Content = documents[0].ContentString()
	}

	// Extract metadata - check known keys
	metadatas := result.GetMetadatas()
	if len(metadatas) > 0 && metadatas[0] != nil {
		doc.Metadata = make(map[string]interface{})
		if v, ok := metadatas[0].GetString("url"); ok {
			doc.URL = v
		}
		if v, ok := metadatas[0].GetString("image"); ok {
			doc.Image = v
		}
		// Copy other common metadata fields
		if v, ok := metadatas[0].GetString("filename"); ok {
			doc.Metadata["filename"] = v
		}
		if v, ok := metadatas[0].GetString("type"); ok {
			doc.Metadata["type"] = v
		}
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Prepare content
	content := document.Content
	if content == "" {
		content = document.Text
	}

	// Prepare metadata map
	metadataMap := make(map[string]interface{})
	if document.Metadata != nil {
		for k, v := range document.Metadata {
			metadataMap[k] = v
		}
	}
	if document.URL != "" {
		metadataMap["url"] = document.URL
	}
	if document.Image != "" {
		metadataMap["image"] = document.Image
	}

	// Prepare metadata - ALWAYS provide it (even if empty) to keep array lengths consistent
	metadata, err := chroma.NewDocumentMetadataFromMap(metadataMap)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	// Update document - wrap single values in slices and spread them
	err = collection.Update(ctx,
		chroma.WithIDsUpdate([]chroma.DocumentID{chroma.DocumentID(document.ID)}...),
		chroma.WithTextsUpdate([]string{content}...),
		chroma.WithMetadatasUpdate([]chroma.DocumentMetadata{metadata}...),
	)
	if err != nil {
		return fmt.Errorf("failed to update document %s: %w", document.ID, err)
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Delete document by ID - Chroma expects variadic argument
	err = collection.Delete(ctx, chroma.WithIDsDelete(chroma.DocumentID(documentID)))
	if err != nil {
		return fmt.Errorf("failed to delete document %s: %w", documentID, err)
	}

	return nil
}

// ListDocuments returns all documents in a collection
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Build get options - must include all fields we want
	getOpts := []chroma.CollectionGetOption{
		chroma.WithIncludeGet(chroma.IncludeDocuments, chroma.IncludeMetadatas),
	}
	// Chroma requires a limit to return results - use provided limit or default to high number
	if limit > 0 {
		getOpts = append(getOpts, chroma.WithLimitGet(limit+offset))
	} else {
		// Default to 10000 if no limit specified (Chroma won't return anything without a limit)
		getOpts = append(getOpts, chroma.WithLimitGet(10000))
	}

	// Get all documents
	result, err := collection.Get(ctx, getOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Build document list from raw result data
	var docs []*vectordb.Document

	ids := result.GetIDs()
	documents := result.GetDocuments()
	metadatas := result.GetMetadatas()

	for i, id := range ids {
		// Apply offset
		if i < offset {
			continue
		}

		doc := &vectordb.Document{
			ID: string(id),
		}

		// Get document content if available
		if i < len(documents) && documents[i] != nil {
			doc.Content = documents[i].ContentString()
		}

		// Get metadata if available
		if i < len(metadatas) && metadatas[i] != nil {
			doc.Metadata = make(map[string]interface{})
			if v, ok := metadatas[i].GetString("url"); ok {
				doc.URL = v
			}
			if v, ok := metadatas[i].GetString("image"); ok {
				doc.Image = v
			}
			if v, ok := metadatas[i].GetString("filename"); ok {
				doc.Metadata["filename"] = v
			}
			if v, ok := metadatas[i].GetString("type"); ok {
				doc.Metadata["type"] = v
			}
		}

		docs = append(docs, doc)

		// Apply limit if specified
		if limit > 0 && len(docs) >= limit {
			break
		}
	}

	return docs, nil
}

// CreateDocuments creates multiple documents in batch
func (c *Client) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	if len(documents) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Prepare batch data
	var ids []chroma.DocumentID
	var contents []string
	var metadatas []chroma.DocumentMetadata

	for _, doc := range documents {
		ids = append(ids, chroma.DocumentID(doc.ID))

		// Prepare content
		content := doc.Content
		if content == "" {
			content = doc.Text
		}
		contents = append(contents, content)

		// Prepare metadata map
		metadataMap := make(map[string]interface{})
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadataMap[k] = v
			}
		}
		if doc.URL != "" {
			metadataMap["url"] = doc.URL
		}
		if doc.Image != "" {
			metadataMap["image"] = doc.Image
		}
		metadata, _ := chroma.NewDocumentMetadataFromMap(metadataMap)
		metadatas = append(metadatas, metadata)
	}

	// Add all documents - use variadic expansion (spread with ...)
	err = collection.Add(ctx,
		chroma.WithIDs(ids...),
		chroma.WithTexts(contents...),
		chroma.WithMetadatas(metadatas...),
	)
	if err != nil {
		return fmt.Errorf("failed to create documents: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by ID
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Convert to DocumentIDs
	var ids []chroma.DocumentID
	for _, id := range documentIDs {
		ids = append(ids, chroma.DocumentID(id))
	}

	// Delete documents by IDs
	err = collection.Delete(ctx, chroma.WithIDsDelete(ids...))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (c *Client) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Build where clauses from metadata
	var clauses []chroma.WhereClause
	for k, v := range metadata {
		switch val := v.(type) {
		case string:
			clauses = append(clauses, chroma.EqString(k, val))
		case int:
			clauses = append(clauses, chroma.EqInt(k, val))
		case float64:
			clauses = append(clauses, chroma.EqFloat(k, float32(val)))
		case bool:
			clauses = append(clauses, chroma.EqBool(k, val))
		}
	}

	// Build delete options
	var deleteOpts []chroma.CollectionDeleteOption
	if len(clauses) == 1 {
		deleteOpts = append(deleteOpts, chroma.WithWhereDelete(clauses[0]))
	} else if len(clauses) > 1 {
		deleteOpts = append(deleteOpts, chroma.WithWhereDelete(chroma.And(clauses...)))
	}

	// Delete documents matching metadata filter
	err = collection.Delete(ctx, deleteOpts...)
	if err != nil {
		return fmt.Errorf("failed to delete documents by metadata: %w", err)
	}

	return nil
}
