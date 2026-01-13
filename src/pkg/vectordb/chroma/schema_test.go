// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build (darwin && amd64) || (darwin && arm64)

package chroma

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestClient_ValidateSchema(t *testing.T) {
	client := &Client{
		config: &Config{},
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
			errMsg:  "schema class name is required",
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
					{Name: "text", DataType: []string{"text"}},
					{Name: "metadata", DataType: []string{"object"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid schema with single property",
			schema: &vectordb.CollectionSchema{
				Class: "minimal-collection",
				Properties: []vectordb.SchemaProperty{
					{Name: "content", DataType: []string{"text"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateSchema(tt.schema)

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
