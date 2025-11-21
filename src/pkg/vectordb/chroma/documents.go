// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (c *Client) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Prepare document data
	ids := []string{document.ID}

	// Prepare content - use Content or Text
	content := document.Content
	if content == "" {
		content = document.Text
	}
	documents := []string{content}

	// Prepare metadata
	metadatas := []map[string]interface{}{}
	metadata := make(map[string]interface{})
	if document.Metadata != nil {
		for k, v := range document.Metadata {
			metadata[k] = v
		}
	}
	// Add standard fields to metadata
	if document.URL != "" {
		metadata["url"] = document.URL
	}
	if document.Image != "" {
		metadata["image"] = document.Image
	}
	metadatas = append(metadatas, metadata)

	// Add document to collection
	// Note: Chroma will use its embedding function to generate embeddings
	_, err = collection.Add(ctx, nil, metadatas, documents, ids)
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Get document by ID using Get with IDs filter
	result, err := collection.Get(ctx, nil, nil, []string{documentID}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get document %s: %w", documentID, err)
	}

	// Check if document was found
	if len(result.Documents) == 0 {
		return nil, vectordb.ErrNotFound("document", documentID)
	}

	// Build document from result
	doc := &vectordb.Document{
		ID:      documentID,
		Content: result.Documents[0],
	}

	// Extract metadata
	if len(result.Metadatas) > 0 && result.Metadatas[0] != nil {
		doc.Metadata = make(map[string]interface{})
		for k, v := range result.Metadatas[0] {
			switch k {
			case "url":
				if s, ok := v.(string); ok {
					doc.URL = s
				}
			case "image":
				if s, ok := v.(string); ok {
					doc.Image = s
				}
			default:
				doc.Metadata[k] = v
			}
		}
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Prepare document data
	ids := []string{document.ID}

	// Prepare content
	content := document.Content
	if content == "" {
		content = document.Text
	}
	documents := []string{content}

	// Prepare metadata
	metadatas := []map[string]interface{}{}
	metadata := make(map[string]interface{})
	if document.Metadata != nil {
		for k, v := range document.Metadata {
			metadata[k] = v
		}
	}
	if document.URL != "" {
		metadata["url"] = document.URL
	}
	if document.Image != "" {
		metadata["image"] = document.Image
	}
	metadatas = append(metadatas, metadata)

	// Update document using Modify (updates existing documents)
	_, err = collection.Modify(ctx, nil, metadatas, documents, ids)
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Delete document by ID
	_, err = collection.Delete(ctx, []string{documentID}, nil, nil)
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Get all documents (Chroma Get without filters returns all)
	result, err := collection.Get(ctx, nil, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Build document list
	var docs []*vectordb.Document
	for i, id := range result.Ids {
		doc := &vectordb.Document{
			ID: id,
		}

		// Add content if available
		if i < len(result.Documents) {
			doc.Content = result.Documents[i]
		}

		// Add metadata if available
		if i < len(result.Metadatas) && result.Metadatas[i] != nil {
			doc.Metadata = make(map[string]interface{})
			for k, v := range result.Metadatas[i] {
				switch k {
				case "url":
					if s, ok := v.(string); ok {
						doc.URL = s
					}
				case "image":
					if s, ok := v.(string); ok {
						doc.Image = s
					}
				default:
					doc.Metadata[k] = v
				}
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Prepare batch data
	var ids []string
	var contents []string
	var metadatas []map[string]interface{}

	for _, doc := range documents {
		ids = append(ids, doc.ID)

		// Prepare content
		content := doc.Content
		if content == "" {
			content = doc.Text
		}
		contents = append(contents, content)

		// Prepare metadata
		metadata := make(map[string]interface{})
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}
		if doc.URL != "" {
			metadata["url"] = doc.URL
		}
		if doc.Image != "" {
			metadata["image"] = doc.Image
		}
		metadatas = append(metadatas, metadata)
	}

	// Add all documents
	_, err = collection.Add(ctx, nil, metadatas, contents, ids)
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Delete documents by IDs
	_, err = collection.Delete(ctx, documentIDs, nil, nil)
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
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	// Delete documents matching metadata filter
	_, err = collection.Delete(ctx, nil, metadata, nil)
	if err != nil {
		return fmt.Errorf("failed to delete documents by metadata: %w", err)
	}

	return nil
}
