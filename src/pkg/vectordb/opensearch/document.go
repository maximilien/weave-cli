// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument indexes a document with vector
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Prepare document for indexing
	doc := map[string]interface{}{
		"text":     document.Text,
		"content":  document.Content,
		"metadata": document.Metadata,
	}

	// TODO: Generate embedding if LLM client is available
	// embedding, err := a.llmClient.GenerateEmbedding(ctx, content, "text-embedding-3-small")
	// if err == nil && len(embedding) > 0 {
	//     doc["vector"] = embedding
	// }

	// Determine document ID
	docID := document.ID
	if docID == "" {
		docID = fmt.Sprintf("doc_%d", len(document.Text)) // Simple ID generation
	}

	// Encode document as JSON
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to marshal document: %w", err)
	}

	resp, err := a.client.Document.Create(
		ctx,
		opensearchapi.DocumentCreateReq{
			Index:      collectionName,
			DocumentID: docID,
			Body:       bytes.NewReader(body),
		},
	)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to create document: %w", err)
	}

	// Update document ID if it was generated
	if document.ID == "" {
		document.ID = resp.ID
	}

	return nil
}

// CreateDocuments bulk indexes multiple documents
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	// Use extended timeout for bulk operations
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	// Process in parallel batches for better performance
	// Use a semaphore pattern to limit concurrency
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)
	errCh := make(chan error, len(documents))

	for _, doc := range documents {
		sem <- struct{}{} // Acquire
		go func(d *vectordb.Document) {
			defer func() { <-sem }() // Release
			if err := a.CreateDocument(ctx, collectionName, d); err != nil {
				errCh <- fmt.Errorf("failed to create document %s: %w", d.ID, err)
			}
		}(doc)
	}

	// Wait for all goroutines to complete
	for i := 0; i < maxConcurrency; i++ {
		sem <- struct{}{}
	}
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return fmt.Errorf("OpenSearch: bulk create failed: %w", err)
		}
	}

	return nil
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	resp, err := a.client.Document.Get(
		ctx,
		opensearchapi.DocumentGetReq{
			Index:      collectionName,
			DocumentID: documentID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenSearch: failed to get document: %w", err)
	}

	if !resp.Found {
		return nil, fmt.Errorf("OpenSearch: document not found: %s", documentID)
	}

	// Parse source from RawMessage
	var source map[string]interface{}
	if err := json.Unmarshal(resp.Source, &source); err != nil {
		return nil, fmt.Errorf("OpenSearch: failed to parse document source: %w", err)
	}

	// Build document from parsed source
	doc := &vectordb.Document{
		ID: resp.ID,
	}

	// Extract text field
	if text, ok := source["text"].(string); ok {
		doc.Text = text
	}

	// Extract content field
	if content, ok := source["content"].(string); ok {
		doc.Content = content
	}

	// Extract metadata
	if metadata, ok := source["metadata"].(map[string]interface{}); ok {
		doc.Metadata = metadata
	}

	// Extract image data if present
	if imageData, ok := source["image_data"].(string); ok {
		doc.ImageData = imageData
	}
	if image, ok := source["image"].(string); ok {
		doc.Image = image
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Prepare document for update
	doc := map[string]interface{}{
		"text":     document.Text,
		"content":  document.Content,
		"metadata": document.Metadata,
	}

	// Include image data if present
	if document.ImageData != "" {
		doc["image_data"] = document.ImageData
	}
	if document.Image != "" {
		doc["image"] = document.Image
	}

	// Encode document as JSON
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to marshal update document: %w", err)
	}

	// OpenSearch uses Create with existing ID to update/overwrite a document
	resp, err := a.client.Document.Create(
		ctx,
		opensearchapi.DocumentCreateReq{
			Index:      collectionName,
			DocumentID: document.ID,
			Body:       bytes.NewReader(body),
		},
	)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to update document: %w", err)
	}

	if resp.Result != "updated" && resp.Result != "created" {
		return fmt.Errorf("OpenSearch: unexpected update result: %s", resp.Result)
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	resp, err := a.client.Document.Delete(
		ctx,
		opensearchapi.DocumentDeleteReq{
			Index:      collectionName,
			DocumentID: documentID,
		},
	)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to delete document: %w", err)
	}

	if resp.Result != "deleted" && resp.Result != "not_found" {
		return fmt.Errorf("OpenSearch: unexpected delete result: %s", resp.Result)
	}

	return nil
}

// DeleteDocuments deletes multiple documents
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	// Use extended timeout for bulk operations
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	// Process in parallel batches for better performance
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)
	errCh := make(chan error, len(documentIDs))

	for _, docID := range documentIDs {
		sem <- struct{}{} // Acquire
		go func(id string) {
			defer func() { <-sem }() // Release
			if err := a.DeleteDocument(ctx, collectionName, id); err != nil {
				errCh <- fmt.Errorf("failed to delete document %s: %w", id, err)
			}
		}(docID)
	}

	// Wait for all goroutines to complete
	for i := 0; i < maxConcurrency; i++ {
		sem <- struct{}{}
	}
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return fmt.Errorf("OpenSearch: bulk delete failed: %w", err)
		}
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata filters
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// Build query with metadata filters
	mustClauses := []map[string]interface{}{}
	for key, value := range metadata {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				fmt.Sprintf("metadata.%s", key): value,
			},
		})
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to marshal delete query: %w", err)
	}

	resp, err := a.client.Document.DeleteByQuery(
		ctx,
		opensearchapi.DocumentDeleteByQueryReq{
			Indices: []string{collectionName},
			Body:    bytes.NewReader(body),
		},
	)
	if err != nil {
		return fmt.Errorf("OpenSearch: failed to delete by query: %w", err)
	}

	// resp.Deleted is an int, not a pointer
	if resp.Deleted < 0 {
		return fmt.Errorf("OpenSearch: invalid deleted count: %d", resp.Deleted)
	}

	return nil
}

// ListDocuments lists all documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// TODO: Implement proper search with pagination
	return nil, fmt.Errorf("OpenSearch: ListDocuments not yet fully implemented")
}
