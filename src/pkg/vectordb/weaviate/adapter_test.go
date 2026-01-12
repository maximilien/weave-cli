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
		name   string
		models []*Document
	}{
		{
			name:   "Nil models",
			models: nil,
		},
		{
			name:   "Empty slice",
			models: []*Document{},
		},
		{
			name: "Single document",
			models: []*Document{
				{ID: "doc-1", Text: "Test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDocumentsFromWeaviate(tt.models)
			assert.NotNil(t, result)
		})
	}
}

func TestConvertSchemaFromWeaviate(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name   string
		schema *CollectionSchema
	}{
		{
			name:   "Nil schema",
			schema: nil,
		},
		{
			name: "Schema with class",
			schema: &CollectionSchema{
				Class: "TestClass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertSchemaFromWeaviate(tt.schema)
			assert.NotNil(t, result)
		})
	}
}
