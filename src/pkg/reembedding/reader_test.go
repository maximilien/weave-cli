// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// MockVectorDBClient for testing
type MockVectorDBClient struct {
	totalDocs int64
	documents [][]*vectordb.Document
	batchIdx  int
	shouldErr bool
}

func (m *MockVectorDBClient) GetCollectionCount(ctx context.Context, collectionName string) (int64, error) {
	if m.shouldErr {
		return 0, errors.New("mock error")
	}
	return m.totalDocs, nil
}

func (m *MockVectorDBClient) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	if m.shouldErr {
		return nil, errors.New("mock error")
	}

	if m.batchIdx >= len(m.documents) {
		return []*vectordb.Document{}, nil
	}

	batch := m.documents[m.batchIdx]
	m.batchIdx++
	return batch, nil
}

// Implement other required interface methods as no-ops
func (m *MockVectorDBClient) Health(ctx context.Context) error { return nil }
func (m *MockVectorDBClient) CollectionExists(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (m *MockVectorDBClient) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	return nil
}
func (m *MockVectorDBClient) DeleteCollection(ctx context.Context, name string) error { return nil }
func (m *MockVectorDBClient) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	return nil, nil
}
func (m *MockVectorDBClient) CreateDocument(ctx context.Context, collectionName string, doc *vectordb.Document) error {
	return nil
}
func (m *MockVectorDBClient) CreateDocuments(ctx context.Context, collectionName string, docs []*vectordb.Document) error {
	return nil
}
func (m *MockVectorDBClient) GetDocument(ctx context.Context, collectionName string, id string) (*vectordb.Document, error) {
	return nil, nil
}
func (m *MockVectorDBClient) UpdateDocument(ctx context.Context, collectionName string, doc *vectordb.Document) error {
	return nil
}
func (m *MockVectorDBClient) DeleteDocument(ctx context.Context, collectionName string, id string) error {
	return nil
}
func (m *MockVectorDBClient) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return nil
}
func (m *MockVectorDBClient) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockVectorDBClient) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}
func (m *MockVectorDBClient) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}
func (m *MockVectorDBClient) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}
func (m *MockVectorDBClient) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}
func (m *MockVectorDBClient) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	return nil, nil
}
func (m *MockVectorDBClient) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	return nil
}
func (m *MockVectorDBClient) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return nil
}
func (m *MockVectorDBClient) ValidateSchema(schema *vectordb.CollectionSchema) error { return nil }

func TestNewCollectionReader(t *testing.T) {
	tests := []struct {
		name           string
		client         vectordb.VectorDBClient
		collectionName string
		batchSize      int
		totalDocs      int64
		shouldErr      bool
		errContains    string
	}{
		{
			name:           "valid reader with custom batch size",
			client:         &MockVectorDBClient{totalDocs: 100},
			collectionName: "TestCollection",
			batchSize:      50,
			totalDocs:      100,
			shouldErr:      false,
		},
		{
			name:           "valid reader with default batch size",
			client:         &MockVectorDBClient{totalDocs: 200},
			collectionName: "TestCollection",
			batchSize:      0, // Should default to 100
			totalDocs:      200,
			shouldErr:      false,
		},
		{
			name:           "nil client",
			client:         nil,
			collectionName: "TestCollection",
			batchSize:      100,
			shouldErr:      true,
			errContains:    "client cannot be nil",
		},
		{
			name:           "empty collection name",
			client:         &MockVectorDBClient{totalDocs: 100},
			collectionName: "",
			batchSize:      100,
			shouldErr:      true,
			errContains:    "collection name cannot be empty",
		},
		{
			name:           "error getting collection count",
			client:         &MockVectorDBClient{totalDocs: 100, shouldErr: true},
			collectionName: "TestCollection",
			batchSize:      100,
			shouldErr:      true,
			errContains:    "failed to get collection count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewCollectionReader(tt.client, tt.collectionName, tt.batchSize)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if reader == nil {
				t.Error("Expected non-nil reader")
				return
			}

			if reader.totalDocs != tt.totalDocs {
				t.Errorf("Expected totalDocs %d, got %d", tt.totalDocs, reader.totalDocs)
			}

			if tt.batchSize > 0 && reader.batchSize != tt.batchSize {
				t.Errorf("Expected batchSize %d, got %d", tt.batchSize, reader.batchSize)
			}

			if tt.batchSize <= 0 && reader.batchSize != 100 {
				t.Errorf("Expected default batchSize 100, got %d", reader.batchSize)
			}
		})
	}
}

func TestCollectionReader_ReadBatch(t *testing.T) {
	tests := []struct {
		name          string
		totalDocs     int64
		batchSize     int
		numBatches    int
		expectedReads int
	}{
		{
			name:          "read 3 batches of 100",
			totalDocs:     300,
			batchSize:     100,
			numBatches:    3,
			expectedReads: 3,
		},
		{
			name:          "read partial last batch",
			totalDocs:     250,
			batchSize:     100,
			numBatches:    3,
			expectedReads: 3,
		},
		{
			name:          "read single batch",
			totalDocs:     50,
			batchSize:     100,
			numBatches:    1,
			expectedReads: 1,
		},
		{
			name:          "empty collection",
			totalDocs:     0,
			batchSize:     100,
			numBatches:    0,
			expectedReads: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock batches
			mockBatches := make([][]*vectordb.Document, tt.numBatches)
			for i := 0; i < tt.numBatches; i++ {
				batchSize := tt.batchSize
				if i == tt.numBatches-1 {
					// Last batch might be smaller
					remaining := int(tt.totalDocs) - (i * tt.batchSize)
					if remaining < batchSize {
						batchSize = remaining
					}
				}

				batch := make([]*vectordb.Document, batchSize)
				for j := 0; j < batchSize; j++ {
					batch[j] = &vectordb.Document{
						ID:   fmt.Sprintf("doc-%d", i*tt.batchSize+j),
						Text: "test content",
					}
				}
				mockBatches[i] = batch
			}

			mockClient := &MockVectorDBClient{
				totalDocs: tt.totalDocs,
				documents: mockBatches,
			}

			reader, err := NewCollectionReader(mockClient, "TestCollection", tt.batchSize)
			if err != nil {
				t.Fatalf("Failed to create reader: %v", err)
			}

			ctx := context.Background()
			batchesRead := 0

			for reader.HasMore() {
				docs, err := reader.ReadBatch(ctx)
				if err != nil {
					t.Fatalf("ReadBatch failed: %v", err)
				}

				if docs == nil && tt.totalDocs > 0 {
					t.Error("Expected non-nil documents")
					break
				}

				batchesRead++
			}

			if batchesRead != tt.expectedReads {
				t.Errorf("Expected %d batches read, got %d", tt.expectedReads, batchesRead)
			}

			// Verify progress
			read, total := reader.Progress()
			if int64(read) != tt.totalDocs {
				t.Errorf("Expected %d docs read, got %d", tt.totalDocs, read)
			}
			if total != tt.totalDocs {
				t.Errorf("Expected total %d, got %d", tt.totalDocs, total)
			}
		})
	}
}

func TestCollectionReader_HasMore(t *testing.T) {
	mockClient := &MockVectorDBClient{
		totalDocs: 100,
		documents: [][]*vectordb.Document{
			make([]*vectordb.Document, 50),
			make([]*vectordb.Document, 50),
		},
	}

	reader, err := NewCollectionReader(mockClient, "TestCollection", 50)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	if !reader.HasMore() {
		t.Error("Expected HasMore to return true initially")
	}

	ctx := context.Background()

	// Read first batch
	_, err = reader.ReadBatch(ctx)
	if err != nil {
		t.Fatalf("ReadBatch failed: %v", err)
	}

	if !reader.HasMore() {
		t.Error("Expected HasMore to return true after first batch")
	}

	// Read second batch
	_, err = reader.ReadBatch(ctx)
	if err != nil {
		t.Fatalf("ReadBatch failed: %v", err)
	}

	if reader.HasMore() {
		t.Error("Expected HasMore to return false after all batches read")
	}
}

func TestCollectionReader_Progress(t *testing.T) {
	mockClient := &MockVectorDBClient{
		totalDocs: 200,
		documents: [][]*vectordb.Document{
			make([]*vectordb.Document, 100),
			make([]*vectordb.Document, 100),
		},
	}

	reader, err := NewCollectionReader(mockClient, "TestCollection", 100)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ctx := context.Background()

	// Initial progress
	read, total := reader.Progress()
	if read != 0 || total != 200 {
		t.Errorf("Expected progress (0, 200), got (%d, %d)", read, total)
	}

	// After first batch
	_, err = reader.ReadBatch(ctx)
	if err != nil {
		t.Fatalf("ReadBatch failed: %v", err)
	}

	read, total = reader.Progress()
	if read != 100 || total != 200 {
		t.Errorf("Expected progress (100, 200), got (%d, %d)", read, total)
	}

	// After second batch
	_, err = reader.ReadBatch(ctx)
	if err != nil {
		t.Fatalf("ReadBatch failed: %v", err)
	}

	read, total = reader.Progress()
	if read != 200 || total != 200 {
		t.Errorf("Expected progress (200, 200), got (%d, %d)", read, total)
	}
}

func TestCollectionReader_Reset(t *testing.T) {
	mockClient := &MockVectorDBClient{
		totalDocs: 100,
		documents: [][]*vectordb.Document{
			make([]*vectordb.Document, 50),
			make([]*vectordb.Document, 50),
		},
	}

	reader, err := NewCollectionReader(mockClient, "TestCollection", 50)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ctx := context.Background()

	// Read first batch
	_, err = reader.ReadBatch(ctx)
	if err != nil {
		t.Fatalf("ReadBatch failed: %v", err)
	}

	read, _ := reader.Progress()
	if read != 50 {
		t.Errorf("Expected 50 docs read, got %d", read)
	}

	// Reset
	reader.Reset()

	read, _ = reader.Progress()
	if read != 0 {
		t.Errorf("Expected 0 docs read after reset, got %d", read)
	}

	if !reader.HasMore() {
		t.Error("Expected HasMore to return true after reset")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
