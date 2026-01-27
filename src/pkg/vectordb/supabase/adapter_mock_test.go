// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package supabase

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

// TestAdapter_CreateCollection_InvalidInput tests collection creation with invalid input
func TestAdapter_CreateCollection_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL:              "postgresql://user:pass@localhost:5432/db",
			VectorDimensions: 1536,
		},
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		colName string
		schema  *vectordb.CollectionSchema
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty collection name",
			colName: "",
			schema:  &vectordb.CollectionSchema{},
			wantErr: true,
			errMsg:  "collection name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.CreateCollection(ctx, tt.colName, tt.schema)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestAdapter_DeleteCollection_EmptyName tests collection deletion with empty name
func TestAdapter_DeleteCollection_EmptyName(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	// Test deletion with empty name - should error
	err := adapter.DeleteCollection(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection name")
}

// TestAdapter_CollectionExists_EmptyName tests collection existence check with empty name
func TestAdapter_CollectionExists_EmptyName(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	// Test with empty name - should error
	exists, err := adapter.CollectionExists(ctx, "")
	assert.False(t, exists)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection name")
}

// TestAdapter_GetCollectionCount_EmptyName tests collection count with empty name
func TestAdapter_GetCollectionCount_EmptyName(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	// Test with empty name - should error
	count, err := adapter.GetCollectionCount(ctx, "")
	assert.Equal(t, int64(0), count)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection name")
}

// TestAdapter_CreateDocument_InvalidInput tests document creation with invalid input
func TestAdapter_CreateDocument_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL:              "postgresql://user:pass@localhost:5432/db",
			VectorDimensions: 1536,
		},
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		doc        *vectordb.Document
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Empty collection name",
			collection: "",
			doc: &vectordb.Document{
				ID:      "test-1",
				Content: "test content",
			},
			wantErr: true,
			errMsg:  "collection name",
		},
		{
			name:       "Nil document",
			collection: "test-collection",
			doc:        nil,
			wantErr:    true,
			errMsg:     "document cannot be nil",
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			doc: &vectordb.Document{
				ID:      "",
				Content: "test content",
			},
			wantErr: true,
			errMsg:  "document ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.CreateDocument(ctx, tt.collection, tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAdapter_GetDocument_InvalidInput tests document retrieval with invalid input
func TestAdapter_GetDocument_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		docID      string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Empty collection name",
			collection: "",
			docID:      "test-1",
			wantErr:    true,
			errMsg:     "collection name",
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			docID:      "",
			wantErr:    true,
			errMsg:     "document ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := adapter.GetDocument(ctx, tt.collection, tt.docID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, doc)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAdapter_UpdateDocument_InvalidInput tests document update with invalid input
func TestAdapter_UpdateDocument_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL:              "postgresql://user:pass@localhost:5432/db",
			VectorDimensions: 1536,
		},
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		doc        *vectordb.Document
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Empty collection name",
			collection: "",
			doc: &vectordb.Document{
				ID:      "test-1",
				Content: "updated content",
			},
			wantErr: true,
			errMsg:  "collection name",
		},
		{
			name:       "Nil document",
			collection: "test-collection",
			doc:        nil,
			wantErr:    true,
			errMsg:     "document cannot be nil",
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			doc: &vectordb.Document{
				ID:      "",
				Content: "updated content",
			},
			wantErr: true,
			errMsg:  "document ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.UpdateDocument(ctx, tt.collection, tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAdapter_DeleteDocument_InvalidInput tests document deletion with invalid input
func TestAdapter_DeleteDocument_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		docID      string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Empty collection name",
			collection: "",
			docID:      "test-1",
			wantErr:    true,
			errMsg:     "collection name",
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			docID:      "",
			wantErr:    true,
			errMsg:     "document ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.DeleteDocument(ctx, tt.collection, tt.docID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAdapter_SearchSemantic_InvalidInput tests semantic search with invalid input
func TestAdapter_SearchSemantic_InvalidInput(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL:              "postgresql://user:pass@localhost:5432/db",
			VectorDimensions: 1536,
		},
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		query      string
		options    *vectordb.QueryOptions
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Empty collection name",
			collection: "",
			query:      "test query",
			options: &vectordb.QueryOptions{
				TopK: 5,
			},
			wantErr: true,
			errMsg:  "collection name",
		},
		{
			name:       "Empty query",
			collection: "test-collection",
			query:      "",
			options: &vectordb.QueryOptions{
				TopK: 5,
			},
			wantErr: true,
			errMsg:  "query",
		},
		{
			name:       "Nil options",
			collection: "test-collection",
			query:      "test query",
			options:    nil,
			wantErr:    true,
			errMsg:     "options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := adapter.SearchSemantic(ctx, tt.collection, tt.query, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAdapter_ListDocuments_EmptyName tests listing documents with empty collection name
func TestAdapter_ListDocuments_EmptyName(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	ctx := context.Background()

	// Test with empty collection name
	docs, err := adapter.ListDocuments(ctx, "", 10, 0)
	assert.Error(t, err)
	assert.Nil(t, docs)
	assert.Contains(t, err.Error(), "collection name")
}

// TestAdapter_GetDefaultSchema tests default schema generation
func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	tests := []struct {
		name           string
		schemaType     vectordb.SchemaType
		collectionName string
		wantVectorizer string
	}{
		{
			name:           "Text schema",
			schemaType:     vectordb.SchemaTypeText,
			collectionName: "test-collection",
			wantVectorizer: "text2vec-openai",
		},
		{
			name:           "Image schema",
			schemaType:     vectordb.SchemaTypeImage,
			collectionName: "test-collection",
			wantVectorizer: "img2vec-neural",
		},
		{
			name:           "Default schema",
			schemaType:     "unknown",
			collectionName: "test-collection",
			wantVectorizer: "text2vec-openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := adapter.GetDefaultSchema(tt.schemaType, tt.collectionName)
			assert.NotNil(t, schema)
			assert.Equal(t, tt.collectionName, schema.Class)
			assert.Equal(t, tt.wantVectorizer, schema.Vectorizer)
			assert.NotEmpty(t, schema.Properties)
		})
	}
}

// TestAdapter_ValidateSchema tests schema validation
func TestAdapter_ValidateSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			URL: "postgresql://user:pass@localhost:5432/db",
		},
	}

	tests := []struct {
		name    string
		schema  *vectordb.CollectionSchema
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Nil schema",
			schema:  nil,
			wantErr: true,
			errMsg:  "schema cannot be nil",
		},
		{
			name: "Empty class name",
			schema: &vectordb.CollectionSchema{
				Class: "",
			},
			wantErr: true,
			errMsg:  "class name cannot be empty",
		},
		{
			name: "No properties",
			schema: &vectordb.CollectionSchema{
				Class:      "test",
				Properties: []vectordb.SchemaProperty{},
			},
			wantErr: true,
			errMsg:  "at least one property",
		},
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class: "test",
				Properties: []vectordb.SchemaProperty{
					{
						Name:     "content",
						DataType: []string{"text"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateSchema(tt.schema)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
