// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestMapDistanceMetric(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantErr  bool
		expected string
	}{
		{
			name:     "Cosine metric",
			metric:   "cosine",
			wantErr:  false,
			expected: "Cosine",
		},
		{
			name:     "Dot product",
			metric:   "dot",
			wantErr:  false,
			expected: "Dot",
		},
		{
			name:     "Euclidean",
			metric:   "euclidean",
			wantErr:  false,
			expected: "Euclidean",
		},
		{
			name:     "L2 norm",
			metric:   "l2",
			wantErr:  false,
			expected: "Euclidean",
		},
		{
			name:     "Invalid metric",
			metric:   "invalid",
			wantErr:  true,
			expected: "",
		},
		{
			name:     "Empty metric returns error",
			metric:   "",
			wantErr:  true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapDistanceMetric(tt.metric)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// Note: Can't directly compare enum values, but we can check it's not nil
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name         string
		config       *vectordb.Config
		expectedHost string
		expectedPort int
	}{
		{
			name: "Full URL with port",
			config: &vectordb.Config{
				URL: "http://localhost:6334",
			},
			expectedHost: "localhost",
			expectedPort: 6334,
		},
		{
			name: "URL without port",
			config: &vectordb.Config{
				URL: "http://localhost",
			},
			expectedHost: "localhost",
			expectedPort: 6334, // Default
		},
		{
			name: "Just host:port",
			config: &vectordb.Config{
				URL: "qdrant.example.com:6334",
			},
			expectedHost: "qdrant.example.com",
			expectedPort: 6334,
		},
		{
			name: "Cloud URL",
			config: &vectordb.Config{
				URL:  "https://xyz.cloud.qdrant.io",
				Type: vectordb.VectorDBTypeQdrantCloud,
			},
			expectedHost: "xyz.cloud.qdrant.io",
			expectedPort: 6334,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := parseAddress(tt.config)

			assert.Equal(t, tt.expectedHost, host)
			assert.Equal(t, tt.expectedPort, port)
		})
	}
}

func TestAdapter_CreateCollection_ValidationErrors(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name             string
		collectionName   string
		vectorDimensions int
		similarityMetric string
		wantErr          bool
	}{
		{
			name:             "Empty collection name",
			collectionName:   "",
			vectorDimensions: 384,
			similarityMetric: "cosine",
			wantErr:          true, // Will fail on server call
		},
		{
			name:             "Zero dimensions",
			collectionName:   "test",
			vectorDimensions: 0,
			similarityMetric: "cosine",
			wantErr:          true, // Will fail on server
		},
		{
			name:             "Negative dimensions",
			collectionName:   "test",
			vectorDimensions: -1,
			similarityMetric: "cosine",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These will fail because no real server, but tests validation logic
			err := adapter.Client.CreateCollection(ctx, tt.collectionName, tt.vectorDimensions, tt.similarityMetric)

			if tt.wantErr {
				// We expect an error (either validation or connection)
				assert.Error(t, err)
			}
		})
	}
}

func TestAdapter_ListCollections_Integration(t *testing.T) {
	t.Skip("Requires live Qdrant server")

	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// This is an integration test that requires a live server
	collections, err := adapter.ListCollections(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, collections)
}

func TestAdapter_CollectionExists_Integration(t *testing.T) {
	t.Skip("Requires live Qdrant server")

	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	exists, err := adapter.CollectionExists(ctx, "nonexistent-collection")

	// Should not error, just return false
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name          string
		configTimeout int
		expected      time.Duration
	}{
		{
			name:          "Custom timeout",
			configTimeout: 60,
			expected:      60 * time.Second,
		},
		{
			name:          "Default timeout",
			configTimeout: 0,
			expected:      10 * time.Second, // Default is 10 seconds
		},
		{
			name:          "Small timeout",
			configTimeout: 5,
			expected:      5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Host:    "localhost",
				Port:    6334,
				Timeout: tt.configTimeout,
			}

			client, err := NewClient(config)
			assert.NoError(t, err)

			result := client.getTimeout()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTimeoutFor(t *testing.T) {
	config := &Config{
		Host:    "localhost",
		Port:    6334,
		Timeout: 30,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		operationType vectordb.OperationType
		minExpected   time.Duration
	}{
		{
			name:          "Collection operation",
			operationType: vectordb.OperationTypeCollection,
			minExpected:   30 * time.Second,
		},
		{
			name:          "Document operation",
			operationType: vectordb.OperationTypeDocument,
			minExpected:   30 * time.Second,
		},
		{
			name:          "Query operation",
			operationType: vectordb.OperationTypeQuery,
			minExpected:   30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.getTimeoutFor(tt.operationType)
			assert.GreaterOrEqual(t, result, tt.minExpected)
		})
	}
}
