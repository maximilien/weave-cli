// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Stub implementations for VectorDBClient interface
// These will be replaced with actual implementations in subsequent phases

// DocumentOperations stubs

func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	return fmt.Errorf("CreateDocument not yet implemented for Elasticsearch")
}

func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	return fmt.Errorf("CreateDocuments not yet implemented for Elasticsearch")
}

func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	return nil, fmt.Errorf("GetDocument not yet implemented for Elasticsearch")
}

func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	return fmt.Errorf("UpdateDocument not yet implemented for Elasticsearch")
}

func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	return fmt.Errorf("DeleteDocument not yet implemented for Elasticsearch")
}

func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return fmt.Errorf("DeleteDocuments not yet implemented for Elasticsearch")
}

func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	return fmt.Errorf("DeleteDocumentsByMetadata not yet implemented for Elasticsearch")
}

func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	return nil, fmt.Errorf("ListDocuments not yet implemented for Elasticsearch")
}

// QueryOperations stubs

func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("SearchSemantic not yet implemented for Elasticsearch")
}

func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("SearchBM25 not yet implemented for Elasticsearch")
}

func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("SearchHybrid not yet implemented for Elasticsearch")
}

func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("SearchByMetadata not yet implemented for Elasticsearch")
}

// SchemaOperations stubs

func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	return nil, fmt.Errorf("GetSchema not yet implemented for Elasticsearch")
}

func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	return fmt.Errorf("UpdateSchema not yet implemented for Elasticsearch")
}

func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "none", // Elasticsearch uses manual embeddings by default
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
		},
	}
}

func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if schema.Class == "" {
		return fmt.Errorf("schema class (collection name) is required")
	}
	return nil
}
