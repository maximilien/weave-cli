// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 1)
	assert.Contains(t, types, vectordb.VectorDBTypePinecone)
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
			name:    "Nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMilvusLocal,
			},
			wantErr: true,
			errMsg:  "invalid config type",
		},
		{
			name: "Missing API key",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypePinecone,
			},
			wantErr: true,
			errMsg:  "API key is required for Pinecone",
		},
		{
			name: "Valid config with API key",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "Valid config with API key and URL",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "test-api-key",
				URL:    "https://my-index.pinecone.io",
			},
			wantErr: false,
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

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}
