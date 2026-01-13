// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetDefaultSchema(t *testing.T) {
	adapter := &Adapter{
		client: &Client{
			config: &Config{
				URI:     "bolt://localhost:7687",
				Timeout: 30 * time.Second,
			},
		},
	}

	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "test-collection")

	assert.NotNil(t, schema)
	assert.Equal(t, "test-collection", schema.Class)
}

func TestAdapter_ValidateSchema(t *testing.T) {
	adapter := &Adapter{
		client: &Client{
			config: &Config{
				URI:     "bolt://localhost:7687",
				Timeout: 30 * time.Second,
			},
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
				Class: "test-collection",
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
				Class:      "test-collection",
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
		client: &Client{
			config: &Config{
				URI:     "bolt://localhost:7687",
				Timeout: 30 * time.Second,
			},
		},
	}

	schema := &vectordb.CollectionSchema{
		Class: "test-collection",
		Properties: []vectordb.SchemaProperty{
			{Name: "text", DataType: []string{"text"}},
		},
	}

	err := adapter.UpdateSchema(nil, "test-collection", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema updates are not supported")
}
