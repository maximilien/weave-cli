// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Document represents a document node in Neo4j
type Document struct {
	ID       string                 // Node ID property
	Content  string                 // Text content
	Vector   []float32              // Embedding vector
	Metadata map[string]interface{} // Additional properties
}

// CreateDocument creates a new document node with vector embedding
func (c *Client) CreateDocument(ctx context.Context, collectionName string, doc *Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID is required")
	}

	// Build Cypher query to create node with label and properties
	query := fmt.Sprintf(`
		CREATE (d:%s {
			id: $id,
			content: $content,
			embedding: $embedding
		})
		SET d += $metadata
		RETURN d.id as id
	`, collectionName)

	params := map[string]interface{}{
		"id":        doc.ID,
		"content":   doc.Content,
		"embedding": doc.Vector,
		"metadata":  doc.Metadata,
	}

	_, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, id string) (*Document, error) {
	query := fmt.Sprintf(`
		MATCH (d:%s {id: $id})
		RETURN d.id as id, d.content as content, d.embedding as embedding, properties(d) as props
	`, collectionName)

	params := map[string]interface{}{
		"id": id,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	record := result.Records[0]
	doc := &Document{
		Metadata: make(map[string]interface{}),
	}

	if val, ok := record.Get("id"); ok {
		doc.ID = val.(string)
	}
	if val, ok := record.Get("content"); ok {
		if content, ok := val.(string); ok {
			doc.Content = content
		}
	}
	if val, ok := record.Get("embedding"); ok {
		if vec, ok := val.([]interface{}); ok {
			doc.Vector = make([]float32, len(vec))
			for i, v := range vec {
				if f, ok := v.(float64); ok {
					doc.Vector[i] = float32(f)
				}
			}
		}
	}
	if val, ok := record.Get("props"); ok {
		if props, ok := val.(map[string]interface{}); ok {
			// Copy all properties except the ones we've already extracted
			for k, v := range props {
				if k != "id" && k != "content" && k != "embedding" {
					doc.Metadata[k] = v
				}
			}
		}
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, doc *Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID is required")
	}

	query := fmt.Sprintf(`
		MATCH (d:%s {id: $id})
		SET d.content = $content,
		    d.embedding = $embedding,
		    d += $metadata
		RETURN d.id as id
	`, collectionName)

	params := map[string]interface{}{
		"id":        doc.ID,
		"content":   doc.Content,
		"embedding": doc.Vector,
		"metadata":  doc.Metadata,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if len(result.Records) == 0 {
		return fmt.Errorf("document not found: %s", doc.ID)
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, id string) error {
	query := fmt.Sprintf(`
		MATCH (d:%s {id: $id})
		DETACH DELETE d
	`, collectionName)

	params := map[string]interface{}{
		"id": id,
	}

	_, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// ListDocuments returns all documents with the given label
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int) ([]*Document, error) {
	query := fmt.Sprintf(`
		MATCH (d:%s)
		RETURN d.id as id, d.content as content, d.embedding as embedding, properties(d) as props
		LIMIT $limit
	`, collectionName)

	if limit <= 0 {
		limit = 100 // Default limit
	}

	params := map[string]interface{}{
		"limit": limit,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var docs []*Document
	for _, record := range result.Records {
		doc := &Document{
			Metadata: make(map[string]interface{}),
		}

		if val, ok := record.Get("id"); ok {
			if id, ok := val.(string); ok {
				doc.ID = id
			}
		}
		if val, ok := record.Get("content"); ok {
			if content, ok := val.(string); ok {
				doc.Content = content
			}
		}
		if val, ok := record.Get("embedding"); ok {
			if vec, ok := val.([]interface{}); ok {
				doc.Vector = make([]float32, len(vec))
				for i, v := range vec {
					if f, ok := v.(float64); ok {
						doc.Vector[i] = float32(f)
					}
				}
			}
		}
		if val, ok := record.Get("props"); ok {
			if props, ok := val.(map[string]interface{}); ok {
				for k, v := range props {
					if k != "id" && k != "content" && k != "embedding" {
						doc.Metadata[k] = v
					}
				}
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// DeleteAllDocuments deletes all documents with the given label
func (c *Client) DeleteAllDocuments(ctx context.Context, collectionName string) error {
	query := fmt.Sprintf(`
		MATCH (d:%s)
		DETACH DELETE d
	`, collectionName)

	_, err := c.executeQuery(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to delete all documents: %w", err)
	}

	return nil
}

// BatchCreateDocuments creates multiple documents in a single transaction
func (c *Client) BatchCreateDocuments(ctx context.Context, collectionName string, docs []*Document) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	if len(docs) == 0 {
		return nil
	}

	// Build batch create query using UNWIND
	query := fmt.Sprintf(`
		UNWIND $docs as doc
		CREATE (d:%s {
			id: doc.id,
			content: doc.content,
			embedding: doc.embedding
		})
		SET d += doc.metadata
	`, collectionName)

	// Prepare documents array
	docsArray := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		docsArray[i] = map[string]interface{}{
			"id":        doc.ID,
			"content":   doc.Content,
			"embedding": doc.Vector,
			"metadata":  doc.Metadata,
		}
	}

	params := map[string]interface{}{
		"docs": docsArray,
	}

	_, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to batch create documents: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by ID
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, ids []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	if len(ids) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		MATCH (d:%s)
		WHERE d.id IN $ids
		DETACH DELETE d
	`, collectionName)

	params := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}
