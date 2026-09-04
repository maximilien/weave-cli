// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter(t *testing.T) {
	t.Skip("Requires Milvus connection - needs mock-based approach")

	tests := []struct {
		name    string
		config  *vectordb.Config
		wantErr bool
	}{
		{
			name: "Valid local config",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				VectorDimensions: 768,
				SimilarityMetric: "L2",
			},
			wantErr: false,
		},
		{
			name: "Valid cloud config with API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeMilvusCloud,
				Address: "https://cloud.zilliz.com:19530",
				APIKey:  "test-api-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.config)
			if err == nil {
				defer adapter.Close()
				t.Skip("Connection succeeded unexpectedly (requires actual Milvus)")
			}
		})
	}
}

func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		Client: &Client{
			config: &Config{
				VectorDimensions: 1536,
			},
		},
	}

	tests := []struct {
		name           string
		schemaType     vectordb.SchemaType
		collectionName string
	}{
		{
			name:           "Text schema",
			schemaType:     vectordb.SchemaTypeText,
			collectionName: "TestDocs",
		},
		{
			name:           "Image schema",
			schemaType:     vectordb.SchemaTypeImage,
			collectionName: "TestImages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := adapter.GetDefaultSchema(tt.schemaType, tt.collectionName)
			assert.NotNil(t, schema)
			assert.Equal(t, "text-embedding-3-small", schema.Vectorizer)
		})
	}
}

func TestAdapter_ValidateSchema(t *testing.T) {
	adapter := &Adapter{
		Client: &Client{
			config: &Config{},
		},
	}

	tests := []struct {
		name    string
		schema  *vectordb.CollectionSchema
		wantErr bool
	}{
		{
			name:    "Nil schema (Milvus allows minimal validation)",
			schema:  nil,
			wantErr: false,
		},
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class: "TestCollection",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"string"}},
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdapter_UpdateSchema(t *testing.T) {
	adapter := &Adapter{
		Client: &Client{
			config: &Config{},
		},
	}

	ctx := context.Background()
	schema := &vectordb.CollectionSchema{
		Class: "TestCollection",
	}

	err := adapter.UpdateSchema(ctx, "test_collection", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support schema updates")
}

func TestAdapter_Close(t *testing.T) {
	t.Skip("Requires Milvus connection - needs mock-based approach")

	config := &vectordb.Config{
		Type:    vectordb.VectorDBTypeMilvusLocal,
		Address: "localhost:19530",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Skip("Skipping test that requires Milvus connection")
	}

	// Close should work (may fail if not connected, but shouldn't panic)
	assert.NotPanics(t, func() {
		_ = adapter.Close()
	})
}
