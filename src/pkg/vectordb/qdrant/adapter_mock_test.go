// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_DeleteCollection(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test deletion with empty name - will error from server
	err = adapter.DeleteCollection(ctx, "")
	assert.Error(t, err)
	// Error comes from Qdrant gRPC validation
	assert.Contains(t, err.Error(), "Validation error")
}

func TestAdapter_GetCollectionCount(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test with empty collection name - should error before server call
	count, err := adapter.GetCollectionCount(ctx, "")
	assert.Equal(t, int64(0), count)
	assert.Error(t, err)
}

func TestAdapter_CollectionExists(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Test with empty name
	exists, err := adapter.CollectionExists(ctx, "")
	assert.False(t, exists)
	assert.Error(t, err)
}

func TestClient_SetGetCollection(t *testing.T) {
	config := &Config{
		Host:    "localhost",
		Port:    6334,
		Timeout: 30,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)

	// Test set and get collection
	client.SetCollection("test-collection")
	assert.Equal(t, "test-collection", client.GetCollection())

	// Test with empty string
	client.SetCollection("")
	assert.Equal(t, "", client.GetCollection())
}

func TestClient_Close(t *testing.T) {
	config := &Config{
		Host:    "localhost",
		Port:    6334,
		Timeout: 30,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)

	// Close should succeed (may return grpc error if no server, but that's ok)
	_ = client.Close()

	// Verify conn is closed by setting it to nil to test nil conn handling
	client.conn = nil
	err = client.Close()
	assert.NoError(t, err)
}

func TestFactory_NewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.NotNil(t, types)
	assert.Contains(t, types, vectordb.VectorDBTypeQdrantLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeQdrantCloud)
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
				Type:    vectordb.VectorDBTypeQdrantLocal,
				URL:     "http://localhost:6334",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "Missing URL",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeQdrantLocal,
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "URL or address is required",
		},
		{
			name: "Empty URL",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeQdrantLocal,
				URL:     "",
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "URL or address is required",
		},
		{
			name: "Unsupported type",
			config: &vectordb.Config{
				Type:    "invalid-type",
				URL:     "http://localhost:6334",
				Timeout: 30,
			},
			wantErr: true,
			errMsg:  "unsupported Qdrant type",
		},
		{
			name: "Cloud config with API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeQdrantCloud,
				URL:     "https://xyz.cloud.qdrant.io",
				APIKey:  "test-key",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "Cloud config without API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeQdrantCloud,
				URL:     "https://xyz.cloud.qdrant.io",
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

func TestAdapter_GetDefaultSchema(t *testing.T) {
	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		Timeout:          30,
		VectorDimensions: 384,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "test-collection")
	assert.NotNil(t, schema)
	assert.Equal(t, "test-collection", schema.Class)
	assert.Equal(t, "text-embedding-3-small", schema.Vectorizer)
}

func TestAdapter_ValidateSchema(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	tests := []struct {
		name   string
		schema *vectordb.CollectionSchema
	}{
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class:      "test",
				Vectorizer: "text-embedding-3-small",
			},
		},
		{
			name: "Schema with properties",
			schema: &vectordb.CollectionSchema{
				Class:      "test",
				Properties: []vectordb.SchemaProperty{},
			},
		},
		{
			name:   "Nil schema",
			schema: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ValidateSchema always returns nil for Qdrant
			err := adapter.ValidateSchema(tt.schema)
			assert.NoError(t, err)
		})
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected []string
	}{
		{
			name:     "Host with port",
			address:  "localhost:6334",
			expected: []string{"localhost", "6334"},
		},
		{
			name:     "Host without port",
			address:  "localhost",
			expected: []string{"localhost"},
		},
		{
			name:     "IP with port",
			address:  "127.0.0.1:6334",
			expected: []string{"127.0.0.1", "6334"},
		},
		{
			name:     "Domain with port",
			address:  "qdrant.example.com:6334",
			expected: []string{"qdrant.example.com", "6334"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitHostPort(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindProtocol(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "HTTP protocol",
			url:      "http://localhost:6334",
			expected: 7, // Length of "http://"
		},
		{
			name:     "HTTPS protocol",
			url:      "https://cloud.qdrant.io",
			expected: 8, // Length of "https://"
		},
		{
			name:     "No protocol",
			url:      "localhost:6334",
			expected: -1,
		},
		{
			name:     "Only domain",
			url:      "qdrant.example.com",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := findProtocol(tt.url)
			assert.Equal(t, tt.expected, idx)
		})
	}
}
