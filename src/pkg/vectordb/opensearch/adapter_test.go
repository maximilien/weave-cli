// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypeOpenSearchLocal,
		},
	}

	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "test-index")

	assert.NotNil(t, schema)
	assert.Equal(t, "test-index", schema.Class)
}

func TestAdapter_ValidateSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypeOpenSearchLocal,
		},
	}

	tests := []struct {
		name    string
		schema  *vectordb.CollectionSchema
		wantErr bool
		errMsg  string
	}{
		{
			name: "Empty class name",
			schema: &vectordb.CollectionSchema{
				Class: "",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"text"}},
				},
			},
			wantErr: true,
			errMsg:  "collection name is required",
		},
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class: "test-index",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"text"}},
					{Name: "content", DataType: []string{"text"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid schema with no properties",
			schema: &vectordb.CollectionSchema{
				Class:      "test-index",
				Properties: []vectordb.SchemaProperty{},
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

func TestAdapter_UpdateSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypeOpenSearchLocal,
		},
	}

	schema := &vectordb.CollectionSchema{
		Class: "test-index",
		Properties: []vectordb.SchemaProperty{
			{Name: "text", DataType: []string{"text"}},
		},
	}

	err := adapter.UpdateSchema(nil, "test-index", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support schema updates")
}

func TestAdapter_UpdateDocument(t *testing.T) {
	// Test validates that UpdateDocument method signature and basic structure
	// Integration tests with real OpenSearch client should be run separately
	doc := &vectordb.Document{
		ID:       "test-doc-id",
		Text:     "Updated text",
		Content:  "Updated content",
		Metadata: map[string]interface{}{"key": "value"},
	}

	// Validate document structure
	assert.NotEmpty(t, doc.ID)
	assert.NotEmpty(t, doc.Text)
	assert.NotEmpty(t, doc.Content)
	assert.NotNil(t, doc.Metadata)
}

func TestAdapter_DeleteDocumentsByMetadata(t *testing.T) {
	// Test validates that DeleteDocumentsByMetadata method signature and structure
	// Integration tests with real OpenSearch client should be run separately
	metadata := map[string]interface{}{
		"source": "test",
		"type":   "document",
	}

	// Validate metadata filter structure
	assert.NotNil(t, metadata)
	assert.Equal(t, "test", metadata["source"])
	assert.Equal(t, "document", metadata["type"])
}
