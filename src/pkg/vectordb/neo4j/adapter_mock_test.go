// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package neo4j

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapter_DeleteCollection_EmptyName tests delete with empty collection name
func TestAdapter_DeleteCollection_EmptyName(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test deletion with empty name - should error
	err = adapter.DeleteCollection(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection name")
}

// TestAdapter_CollectionExists_EmptyName tests exists check with empty name
func TestAdapter_CollectionExists_EmptyName(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with empty name - should error
	exists, err := adapter.CollectionExists(ctx, "")
	assert.False(t, exists)
	assert.Error(t, err)
}

// TestAdapter_GetCollectionCount_EmptyName tests count with empty collection name
func TestAdapter_GetCollectionCount_EmptyName(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with empty collection name
	count, err := adapter.GetCollectionCount(ctx, "")
	assert.Equal(t, int64(0), count)
	assert.Error(t, err)
}

// TestAdapter_CreateDocument_InvalidInput tests document creation with invalid input
func TestAdapter_CreateDocument_InvalidInput(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		doc        *vectordb.Document
		wantErr    bool
	}{
		{
			name:       "Empty collection name",
			collection: "",
			doc: &vectordb.Document{
				ID:      "test-1",
				Content: "test content",
				Text:    "test text",
			},
			wantErr: true,
		},
		{
			name:       "Nil document",
			collection: "test-collection",
			doc:        nil,
			wantErr:    true,
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			doc: &vectordb.Document{
				ID:      "",
				Content: "test content",
				Text:    "test text",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.CreateDocument(ctx, tt.collection, tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestAdapter_GetDocument_InvalidInput tests document retrieval with invalid input
func TestAdapter_GetDocument_InvalidInput(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		docID      string
		wantErr    bool
	}{
		{
			name:       "Empty collection name",
			collection: "",
			docID:      "test-1",
			wantErr:    true,
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			docID:      "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := adapter.GetDocument(ctx, tt.collection, tt.docID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, doc)
			}
		})
	}
}

// TestAdapter_UpdateDocument_InvalidInput tests document update with invalid input
func TestAdapter_UpdateDocument_InvalidInput(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		doc        *vectordb.Document
		wantErr    bool
	}{
		{
			name:       "Empty collection name",
			collection: "",
			doc: &vectordb.Document{
				ID:      "test-1",
				Content: "updated content",
			},
			wantErr: true,
		},
		{
			name:       "Nil document",
			collection: "test-collection",
			doc:        nil,
			wantErr:    true,
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			doc: &vectordb.Document{
				ID:      "",
				Content: "updated content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.UpdateDocument(ctx, tt.collection, tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestAdapter_DeleteDocument_InvalidInput tests document deletion with invalid input
func TestAdapter_DeleteDocument_InvalidInput(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		docID      string
		wantErr    bool
	}{
		{
			name:       "Empty collection name",
			collection: "",
			docID:      "test-1",
			wantErr:    true,
		},
		{
			name:       "Empty document ID",
			collection: "test-collection",
			docID:      "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.DeleteDocument(ctx, tt.collection, tt.docID)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// TestAdapter_SearchSemantic_InvalidInput tests semantic search with invalid input
func TestAdapter_SearchSemantic_InvalidInput(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		collection string
		queryText  string
		options    *vectordb.QueryOptions
		wantErr    bool
	}{
		{
			name:       "Empty collection name",
			collection: "",
			queryText:  "test query",
			options: &vectordb.QueryOptions{
				TopK: 5,
			},
			wantErr: true,
		},
		{
			name:       "Empty query text",
			collection: "test-collection",
			queryText:  "",
			options: &vectordb.QueryOptions{
				TopK: 5,
			},
			wantErr: true,
		},
		{
			name:       "Nil options",
			collection: "test-collection",
			queryText:  "test query",
			options:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := adapter.SearchSemantic(ctx, tt.collection, tt.queryText, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			}
		})
	}
}

// TestClient_Close tests client close
func TestClient_Close(t *testing.T) {
	config := &Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "password",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Close should not error even if not connected
	err = client.Close(ctx)
	assert.NoError(t, err)

	// Second close should also not error
	err = client.Close(ctx)
	assert.NoError(t, err)
}

// TestAdapter_CreateCollection_EmptyName tests collection creation with empty name
func TestAdapter_CreateCollection_EmptyName(t *testing.T) {
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeNeo4jLocal,
		URL:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "password",
		Timeout:          30,
		VectorDimensions: 1536,
	}

	factory := NewFactory()
	adapter, err := factory.CreateClient(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test creation with empty name - should error
	err = adapter.CreateCollection(ctx, "", &vectordb.CollectionSchema{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection name")
}
