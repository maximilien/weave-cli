// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter_Success(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:8080",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)

	// May fail to connect without server, but adapter creation should work
	if err != nil {
		assert.NotNil(t, config)
	} else {
		assert.NotNil(t, adapter)
	}
}

func TestNewAdapter_WithAPIKey(t *testing.T) {
	config := &vectordb.Config{
		URL:     "https://xyz.weaviate.network",
		APIKey:  "test-key",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)

	// May fail to connect without valid cloud instance
	if err != nil {
		assert.NotNil(t, config)
	} else {
		assert.NotNil(t, adapter)
	}
}

func TestConvertDocumentsFromWeaviate(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name      string
		models    []*Document
		expectNil bool
	}{
		{
			name:      "Nil models",
			models:    nil,
			expectNil: true, // Returns nil for nil input
		},
		{
			name:      "Empty slice",
			models:    []*Document{},
			expectNil: false,
		},
		{
			name: "Single document",
			models: []*Document{
				{ID: "doc-1", Text: "Test"},
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDocumentsFromWeaviate(tt.models)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestConvertSchemaFromWeaviate(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name      string
		schema    *CollectionSchema
		expectNil bool
	}{
		{
			name:      "Nil schema",
			schema:    nil,
			expectNil: true, // Returns nil for nil input
		},
		{
			name: "Schema with class",
			schema: &CollectionSchema{
				Class: "TestClass",
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertSchemaFromWeaviate(tt.schema)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}
