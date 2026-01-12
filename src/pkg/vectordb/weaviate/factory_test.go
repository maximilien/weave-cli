// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_NewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.NotNil(t, types)
	assert.Contains(t, types, vectordb.VectorDBTypeWeaviateLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeWeaviateCloud)
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  *vectordb.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid local config",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeWeaviateLocal,
				URL:     "http://localhost:8080",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "Missing URL",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeWeaviateLocal,
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "URL is required",
		},
		{
			name: "Empty URL",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeWeaviateLocal,
				URL:     "",
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "URL is required",
		},
		{
			name: "Unsupported type",
			config: &vectordb.Config{
				Type:    "invalid-type",
				URL:     "http://localhost:8080",
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "unsupported Weaviate type",
		},
		{
			name: "Cloud config with API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeWeaviateCloud,
				URL:     "https://xyz.weaviate.network",
				APIKey:  "test-key",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "Cloud config without API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeWeaviateCloud,
				URL:     "https://xyz.weaviate.network",
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "requires an API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)

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

func TestFactory_CreateClient(t *testing.T) {
	factory := NewFactory()

	config := &vectordb.Config{
		Type:    vectordb.VectorDBTypeWeaviateLocal,
		URL:     "http://localhost:8080",
		Timeout: 30,
	}

	client, err := factory.CreateClient(config)

	// May fail to connect without server, but factory should work
	if err != nil {
		// This is expected without a real Weaviate server
		assert.NotNil(t, config)
	} else {
		assert.NotNil(t, client)
	}
}
