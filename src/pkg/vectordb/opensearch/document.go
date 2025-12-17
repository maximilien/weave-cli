// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"context"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"

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

	resp, err := a.client.Document.Create(
		ctx,
		opensearchapi.DocumentCreateReq{
			Index:      collectionName,
			DocumentID: docID,
			Body:       opensearchutil.NewJSONReader(doc),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
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

	// For now, implement as sequential creates
	// TODO: Implement proper bulk API usage
	for _, doc := range documents {
		if err := a.CreateDocument(ctx, collectionName, doc); err != nil {
			return fmt.Errorf("failed to create document %s: %w", doc.ID, err)
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
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if !resp.Found {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	// TODO: Parse source properly (resp.Source is RawMessage, needs unmarshaling)
	doc := &vectordb.Document{
		ID: resp.ID,
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// TODO: UpdateDocument API needs investigation
	// Will need to prepare doc with text, content, metadata and use proper update API
	return fmt.Errorf("UpdateDocument not yet fully implemented")
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
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if resp.Result != "deleted" && resp.Result != "not_found" {
		return fmt.Errorf("unexpected delete result: %s", resp.Result)
	}

	return nil
}

// DeleteDocuments deletes multiple documents
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	// Use extended timeout for bulk operations
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	// For now, implement as sequential deletes
	// TODO: Implement proper bulk delete
	for _, docID := range documentIDs {
		if err := a.DeleteDocument(ctx, collectionName, docID); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", docID, err)
		}
	}
	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata filters
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// TODO: Implement delete by query with metadata filters
	return fmt.Errorf("DeleteDocumentsByMetadata not yet fully implemented")
}

// ListDocuments lists all documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// TODO: Implement proper search with pagination
	return nil, fmt.Errorf("ListDocuments not yet fully implemented")
}
