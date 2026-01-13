// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		Client: &Client{
			config: &Config{},
		},
	}

	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "test-index")

	assert.NotNil(t, schema)
	assert.Equal(t, "test-index", schema.Class)
	assert.Equal(t, "none", schema.Vectorizer)
	assert.Len(t, schema.Properties, 2)
	assert.Equal(t, "text", schema.Properties[0].Name)
	assert.Equal(t, "content", schema.Properties[1].Name)
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
					{Name: "text", DataType: []string{"text"}},
				},
			},
			wantErr: true,
			errMsg:  "schema class (collection name) is required",
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
