// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"strconv"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	weaviateClient "github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	weaviateDoc := a.convertDocument(document)
	if err := a.weaveClient.CreateDocument(ctx, collectionName, *weaviateDoc); err != nil {
		return a.wrapError(err, "create document")
	}
	return nil
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	for _, doc := range documents {
		if err := a.CreateDocument(ctx, collectionName, doc); err != nil {
			return err
		}
	}
	return nil
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	doc, err := a.weaveClient.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		return nil, a.wrapError(err, "get document")
	}
	return a.convertDocumentFromWeaviate(doc), nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	content := document.Content
	if content == "" {
		content = document.Text
	}
	if err := a.weaveClient.UpdateDocument(ctx, collectionName, document.ID, content, document.Metadata); err != nil {
		return a.wrapError(err, "update document")
	}
	return nil
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	if err := a.weaveClient.DeleteDocument(ctx, collectionName, documentID); err != nil {
		return a.wrapError(err, "delete document")
	}
	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	for _, id := range documentIDs {
		if err := a.weaveClient.DeleteDocument(ctx, collectionName, id); err != nil {
			return a.wrapError(err, "delete documents")
		}
	}
	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// Convert metadata map to filter strings
	var filters []string
	for key, value := range metadata {
		filters = append(filters, key+"="+toString(value))
	}

	_, err := a.weaveClient.DeleteDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return a.wrapError(err, "delete documents by metadata")
	}
	return nil
}

// ListDocuments returns a list of documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// WeaveClient ListDocuments doesn't support offset, so we'll get more and slice
	totalLimit := limit + offset
	docs, err := a.weaveClient.ListDocuments(ctx, collectionName, totalLimit)
	if err != nil {
		return nil, a.wrapError(err, "list documents")
	}

	// Convert to pointers and apply offset
	docPtrs := make([]*weaviateClient.Document, len(docs))
	for i := range docs {
		docPtrs[i] = &docs[i]
	}

	// Apply offset
	if offset >= len(docPtrs) {
		return []*vectordb.Document{}, nil
	}

	end := offset + limit
	if end > len(docPtrs) {
		end = len(docPtrs)
	}

	slicedDocs := docPtrs[offset:end]
	return a.convertDocumentsFromWeaviate(slicedDocs), nil
}

// toString converts an interface{} to string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}
