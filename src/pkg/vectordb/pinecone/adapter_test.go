// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypePinecone,
		},
	}

	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "test-collection")

	assert.NotNil(t, schema)
	assert.Equal(t, "test-collection", schema.Class)
	assert.Equal(t, "none", schema.Vectorizer)
	assert.Len(t, schema.Properties, 2)
	assert.Equal(t, "vector", schema.Properties[0].Name)
	assert.Equal(t, "metadata", schema.Properties[1].Name)
}

func TestAdapter_ValidateSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypePinecone,
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
				Properties: []vectordb.SchemaProperty{
					{Name: "vector", DataType: []string{"number[]"}},
				},
			},
			wantErr: true,
			errMsg:  "schema class cannot be empty",
		},
		{
			name: "No properties",
			schema: &vectordb.CollectionSchema{
				Class:      "test-collection",
				Properties: []vectordb.SchemaProperty{},
			},
			wantErr: true,
			errMsg:  "schema must have at least one property",
		},
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class: "test-collection",
				Properties: []vectordb.SchemaProperty{
					{Name: "vector", DataType: []string{"number[]"}},
					{Name: "metadata", DataType: []string{"object"}},
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

func TestAdapter_UpdateSchema(t *testing.T) {
	adapter := &Adapter{
		config: &vectordb.Config{
			Type: vectordb.VectorDBTypePinecone,
		},
	}

	schema := &vectordb.CollectionSchema{
		Class: "test-collection",
		Properties: []vectordb.SchemaProperty{
			{Name: "vector", DataType: []string{"number[]"}},
		},
	}

	err := adapter.UpdateSchema(context.Background(), "test-collection", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema updates not supported")
}
