// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"context"
	"strconv"
	"strings"

	mockClient "github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Collection operations

// CreateCollection creates a new collection with the given schema
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	// Convert schema to field definitions for mock client
	var fields []weaviate.FieldDefinition
	for _, prop := range schema.Properties {
		fields = append(fields, weaviate.FieldDefinition{
			Name: prop.Name,
			Type: strings.Join(prop.DataType, ","),
		})
	}

	embeddingModel := "text-embedding-ada-002" // Default model
	if schema.Vectorizer != "" {
		embeddingModel = schema.Vectorizer
	}

	return a.client.CreateCollection(ctx, name, embeddingModel, fields)
}

// DeleteCollection deletes a collection and all its documents
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	return a.client.DeleteCollection(ctx, name)
}

// ListCollections returns a list of all collections
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	collections, err := a.client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]vectordb.CollectionInfo, len(collections))
	for i, name := range collections {
		// Get collection count
		count, countErr := a.client.CountDocuments(ctx, name)
		if countErr != nil {
			count = 0
		}

		result[i] = vectordb.CollectionInfo{
			Name:        name,
			Description: "Mock collection",
			Count:       int64(count),
		}
	}

	return result, nil
}

// CollectionExists checks if a collection exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	collections, err := a.client.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, collection := range collections {
		if collection == name {
			return true, nil
		}
	}

	return false, nil
}

// GetCollectionCount returns the number of documents in a collection
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	count, err := a.client.CountDocuments(ctx, name)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// Document operations

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	mockDoc := a.convertDocument(document)
	return a.client.CreateDocument(ctx, collectionName, *mockDoc)
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
	doc, err := a.client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		return nil, err
	}
	return a.convertDocumentFromMock(doc), nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	content := document.Content
	if content == "" {
		content = document.Text
	}
	return a.client.UpdateDocument(ctx, collectionName, document.ID, content, document.Metadata)
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	return a.client.DeleteDocument(ctx, collectionName, documentID)
}

// DeleteDocuments deletes multiple documents by IDs
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	for _, id := range documentIDs {
		if err := a.client.DeleteDocument(ctx, collectionName, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// Convert metadata map to filter strings for mock client
	var filters []string
	for key, value := range metadata {
		filters = append(filters, key+"="+toString(value))
	}

	_, err := a.client.DeleteDocumentsByMetadata(ctx, collectionName, filters)
	return err
}

// ListDocuments returns a list of documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// Mock client doesn't support offset, so we'll get all and slice
	docs, err := a.client.ListDocuments(ctx, collectionName, limit+offset)
	if err != nil {
		return nil, err
	}

	// Apply offset
	if offset >= len(docs) {
		return []*vectordb.Document{}, nil
	}

	end := offset + limit
	if end > len(docs) {
		end = len(docs)
	}

	slicedDocs := docs[offset:end]
	return a.convertDocumentsFromMock(convertToPointers(slicedDocs)), nil
}

// Query operations

// SearchSemantic performs semantic search using vector embeddings
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	topK, _ := a.convertQueryOptions(options)

	weaviateOptions := weaviate.QueryOptions{
		TopK:           topK,
		SearchMetadata: options != nil && options.SearchMetadata,
		NoTruncate:     options != nil && options.NoTruncate,
		UseBM25:        false, // Semantic search
	}

	results, err := a.client.Query(ctx, collectionName, query, weaviateOptions)
	if err != nil {
		return nil, err
	}

	return a.convertMockQueryResults(results), nil
}

// SearchBM25 performs keyword-based search using BM25
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	topK, _ := a.convertQueryOptions(options)

	weaviateOptions := weaviate.QueryOptions{
		TopK:           topK,
		SearchMetadata: options != nil && options.SearchMetadata,
		NoTruncate:     options != nil && options.NoTruncate,
		UseBM25:        true, // BM25 search
	}

	results, err := a.client.Query(ctx, collectionName, query, weaviateOptions)
	if err != nil {
		return nil, err
	}

	return a.convertMockQueryResults(results), nil
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// For mock, just use semantic search
	return a.SearchSemantic(ctx, collectionName, query, options)
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Convert metadata map to filter strings for mock client
	var filters []string
	for key, value := range metadata {
		filters = append(filters, key+"="+toString(value))
	}

	docs, err := a.client.GetDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return nil, err
	}

	// Convert to query results with default scores
	results := make([]*vectordb.QueryResult, len(docs))
	for i, doc := range docs {
		results[i] = &vectordb.QueryResult{
			Document: *a.convertDocumentFromMock(&doc),
			Score:    1.0, // Default score for metadata matches
		}
	}

	return results, nil
}

// Schema operations

// GetSchema retrieves the schema for a collection (mock implementation)
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// Mock implementation - return a basic schema
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "text-embedding-ada-002",
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}, nil
}

// UpdateSchema updates the schema for a collection (mock implementation)
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Mock implementation - always succeeds
	return nil
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	vectorizer := "text-embedding-ada-002"
	properties := []vectordb.SchemaProperty{
		{
			Name:     "content",
			DataType: []string{"text"},
		},
		{
			Name:     "metadata",
			DataType: []string{"object"},
		},
	}

	if schemaType == vectordb.SchemaTypeImage {
		vectorizer = "none"
		properties = append(properties, vectordb.SchemaProperty{
			Name:     "image",
			DataType: []string{"text"},
		})
	}

	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: vectorizer,
		Properties: properties,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return vectordb.ErrInvalidSchema("schema cannot be nil")
	}

	if schema.Class == "" {
		return vectordb.ErrInvalidSchema("schema class name cannot be empty")
	}

	return nil
}

// Helper functions

// convertToPointers converts a slice of Documents to a slice of *Document
func convertToPointers(docs []mockClient.Document) []*mockClient.Document {
	result := make([]*mockClient.Document, len(docs))
	for i := range docs {
		result[i] = &docs[i]
	}
	return result
}

// convertMockQueryResults converts mock query results to vectordb query results
func (a *Adapter) convertMockQueryResults(results []weaviate.QueryResult) []*vectordb.QueryResult {
	queryResults := make([]*vectordb.QueryResult, len(results))
	for i, result := range results {
		queryResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       result.ID,
				Content:  result.Content,
				Metadata: result.Metadata,
			},
			Score: result.Score,
		}
	}
	return queryResults
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
