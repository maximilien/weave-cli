// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mocks

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/mock"
)

// MockVectorDBClient is a mock implementation of vectordb.VectorDBClient for testing
type MockVectorDBClient struct {
	mock.Mock
}

// Health mocks the Health method
func (m *MockVectorDBClient) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Collection Operations

// CreateCollection mocks the CreateCollection method
func (m *MockVectorDBClient) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	args := m.Called(ctx, name, schema)
	return args.Error(0)
}

// DeleteCollection mocks the DeleteCollection method
func (m *MockVectorDBClient) DeleteCollection(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// ListCollections mocks the ListCollections method
func (m *MockVectorDBClient) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]vectordb.CollectionInfo), args.Error(1)
}

// CollectionExists mocks the CollectionExists method
func (m *MockVectorDBClient) CollectionExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// GetCollectionCount mocks the GetCollectionCount method
func (m *MockVectorDBClient) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(int64), args.Error(1)
}

// Document Operations

// CreateDocument mocks the CreateDocument method
func (m *MockVectorDBClient) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	args := m.Called(ctx, collectionName, document)
	return args.Error(0)
}

// CreateDocuments mocks the CreateDocuments method
func (m *MockVectorDBClient) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	args := m.Called(ctx, collectionName, documents)
	return args.Error(0)
}

// GetDocument mocks the GetDocument method
func (m *MockVectorDBClient) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	args := m.Called(ctx, collectionName, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vectordb.Document), args.Error(1)
}

// UpdateDocument mocks the UpdateDocument method
func (m *MockVectorDBClient) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	args := m.Called(ctx, collectionName, document)
	return args.Error(0)
}

// DeleteDocument mocks the DeleteDocument method
func (m *MockVectorDBClient) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	args := m.Called(ctx, collectionName, documentID)
	return args.Error(0)
}

// DeleteDocuments mocks the DeleteDocuments method
func (m *MockVectorDBClient) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	args := m.Called(ctx, collectionName, documentIDs)
	return args.Error(0)
}

// DeleteDocumentsByMetadata mocks the DeleteDocumentsByMetadata method
func (m *MockVectorDBClient) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	args := m.Called(ctx, collectionName, metadata)
	return args.Error(0)
}

// ListDocuments mocks the ListDocuments method
func (m *MockVectorDBClient) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	args := m.Called(ctx, collectionName, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vectordb.Document), args.Error(1)
}

// Query Operations

// SearchSemantic mocks the SearchSemantic method
func (m *MockVectorDBClient) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	args := m.Called(ctx, collectionName, query, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vectordb.QueryResult), args.Error(1)
}

// SearchBM25 mocks the SearchBM25 method
func (m *MockVectorDBClient) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	args := m.Called(ctx, collectionName, query, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vectordb.QueryResult), args.Error(1)
}

// SearchHybrid mocks the SearchHybrid method
func (m *MockVectorDBClient) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	args := m.Called(ctx, collectionName, query, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vectordb.QueryResult), args.Error(1)
}

// SearchByMetadata mocks the SearchByMetadata method
func (m *MockVectorDBClient) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	args := m.Called(ctx, collectionName, metadata, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vectordb.QueryResult), args.Error(1)
}

// Schema Operations

// GetSchema mocks the GetSchema method
func (m *MockVectorDBClient) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	args := m.Called(ctx, collectionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vectordb.CollectionSchema), args.Error(1)
}

// UpdateSchema mocks the UpdateSchema method
func (m *MockVectorDBClient) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	args := m.Called(ctx, collectionName, schema)
	return args.Error(0)
}

// GetDefaultSchema mocks the GetDefaultSchema method
func (m *MockVectorDBClient) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	args := m.Called(schemaType, collectionName)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*vectordb.CollectionSchema)
}

// ValidateSchema mocks the ValidateSchema method
func (m *MockVectorDBClient) ValidateSchema(schema *vectordb.CollectionSchema) error {
	args := m.Called(schema)
	return args.Error(0)
}
