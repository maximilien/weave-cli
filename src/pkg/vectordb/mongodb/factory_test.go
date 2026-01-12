// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	if len(types) != 1 {
		t.Errorf("Expected 1 supported type, got %d", len(types))
	}

	if types[0] != vectordb.VectorDBTypeMongoDB {
		t.Errorf("Expected type %s, got %s", vectordb.VectorDBTypeMongoDB, types[0])
	}
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory()

	testCases := []struct {
		name        string
		config      *vectordb.Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeWeaviateCloud,
			},
			expectError: true,
		},
		{
			name: "Missing URL",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMongoDB,
				Database: "test",
			},
			expectError: true,
		},
		{
			name: "Missing Database",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMongoDB,
				URL:  "mongodb+srv://test:test@cluster.mongodb.net",
			},
			expectError: true,
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMongoDB,
				URL:      "mongodb+srv://test:test@cluster.mongodb.net",
				Database: "test",
				Timeout:  -1,
			},
			expectError: true,
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				VectorDimensions: -100,
			},
			expectError: true,
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				SimilarityMetric: "invalid",
			},
			expectError: true,
		},
		{
			name: "Valid cosine metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				SimilarityMetric: "cosine",
			},
			expectError: false,
		},
		{
			name: "Valid euclidean metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				SimilarityMetric: "euclidean",
			},
			expectError: false,
		},
		{
			name: "Valid dotProduct metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				SimilarityMetric: "dotProduct",
			},
			expectError: false,
		},
		{
			name: "Valid minimal config",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMongoDB,
				URL:      "mongodb+srv://test:test@cluster.mongodb.net",
				Database: "test",
			},
			expectError: false,
		},
		{
			name: "Valid full config",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "testdb",
				VectorDimensions: 1536,
				SimilarityMetric: "cosine",
				Timeout:          30,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := factory.ValidateConfig(tc.config)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFactory_CreateClient(t *testing.T) {
	t.Skip("Requires MongoDB connection - needs mock-based approach")

	factory := NewFactory()

	// Test with valid config (will fail to connect, but that's expected)
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeMongoDB,
		URL:              "mongodb+srv://test:test@cluster.mongodb.net",
		Database:         "test",
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
	}

	_, err := factory.CreateClient(config)
	// We expect connection to fail without actual MongoDB
	if err == nil {
		t.Skip("Connection succeeded unexpectedly (requires actual MongoDB)")
	}
}

func TestNewAdapter(t *testing.T) {
	t.Skip("Requires MongoDB connection - needs mock-based approach")

	testCases := []struct {
		name        string
		config      *vectordb.Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMongoDB,
				URL:              "mongodb+srv://test:test@cluster.mongodb.net",
				Database:         "test",
				VectorDimensions: 768,
				SimilarityMetric: "euclidean",
			},
			expectError: false, // Will fail to connect, but adapter creation should work
		},
		{
			name: "Minimal config",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMongoDB,
				URL:      "mongodb+srv://test:test@cluster.mongodb.net",
				Database: "test",
			},
			expectError: false, // Will fail to connect, but adapter creation should work
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewAdapter(tc.config)
			// We expect connection failures without actual MongoDB
			if err == nil {
				t.Skip("Connection succeeded unexpectedly (requires actual MongoDB)")
			}
		})
	}
}
