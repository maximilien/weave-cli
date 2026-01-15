// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCollectionSpec(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    CollectionSpec
		expectError bool
	}{
		{
			name:  "simple collection name",
			input: "MyCollection",
			expected: CollectionSpec{
				Name:   "MyCollection",
				VDBKey: "",
			},
		},
		{
			name:  "collection with VDB key",
			input: "MyCollection:weaviate-local",
			expected: CollectionSpec{
				Name:   "MyCollection",
				VDBKey: "weaviate-local",
			},
		},
		{
			name:  "collection with cloud VDB",
			input: "AuctionDocs:milvus-cloud",
			expected: CollectionSpec{
				Name:   "AuctionDocs",
				VDBKey: "milvus-cloud",
			},
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:  "collection with colon but no VDB",
			input: "MyCollection:",
			expected: CollectionSpec{
				Name:   "MyCollection",
				VDBKey: "",
			},
		},
		{
			name:  "collection with multiple colons (uses first)",
			input: "My:Collection:weaviate-local",
			expected: CollectionSpec{
				Name:   "My",
				VDBKey: "Collection:weaviate-local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCollectionSpec(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.VDBKey, result.VDBKey)
			}
		})
	}
}

func TestParseCollectionSpecs(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    []CollectionSpec
		expectError bool
	}{
		{
			name:  "multiple collections without VDB",
			input: []string{"Col1", "Col2", "Col3"},
			expected: []CollectionSpec{
				{Name: "Col1", VDBKey: ""},
				{Name: "Col2", VDBKey: ""},
				{Name: "Col3", VDBKey: ""},
			},
		},
		{
			name:  "multiple collections with VDB",
			input: []string{"Col1:weaviate-local", "Col2:milvus-cloud"},
			expected: []CollectionSpec{
				{Name: "Col1", VDBKey: "weaviate-local"},
				{Name: "Col2", VDBKey: "milvus-cloud"},
			},
		},
		{
			name:  "mixed collections",
			input: []string{"Col1", "Col2:milvus-local", "Col3"},
			expected: []CollectionSpec{
				{Name: "Col1", VDBKey: ""},
				{Name: "Col2", VDBKey: "milvus-local"},
				{Name: "Col3", VDBKey: ""},
			},
		},
		{
			name:        "empty input",
			input:       []string{},
			expectError: true,
		},
		{
			name:        "invalid spec in list",
			input:       []string{"Col1", "", "Col3"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCollectionSpecs(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(result))
				for i, expected := range tt.expected {
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.VDBKey, result[i].VDBKey)
				}
			}
		})
	}
}

func TestHasExplicitVDBs(t *testing.T) {
	tests := []struct {
		name     string
		specs    []CollectionSpec
		expected bool
	}{
		{
			name: "no explicit VDBs",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: ""},
				{Name: "Col2", VDBKey: ""},
			},
			expected: false,
		},
		{
			name: "has explicit VDB",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: ""},
				{Name: "Col2", VDBKey: "weaviate-local"},
			},
			expected: true,
		},
		{
			name: "all explicit VDBs",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: "weaviate-local"},
				{Name: "Col2", VDBKey: "milvus-cloud"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasExplicitVDBs(tt.specs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUniqueVDBKeys(t *testing.T) {
	tests := []struct {
		name     string
		specs    []CollectionSpec
		expected []string
	}{
		{
			name: "no VDB keys",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: ""},
				{Name: "Col2", VDBKey: ""},
			},
			expected: []string{},
		},
		{
			name: "one VDB key",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: "weaviate-local"},
				{Name: "Col2", VDBKey: ""},
			},
			expected: []string{"weaviate-local"},
		},
		{
			name: "multiple unique VDB keys",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: "weaviate-local"},
				{Name: "Col2", VDBKey: "milvus-cloud"},
				{Name: "Col3", VDBKey: ""},
			},
			expected: []string{"weaviate-local", "milvus-cloud"},
		},
		{
			name: "duplicate VDB keys",
			specs: []CollectionSpec{
				{Name: "Col1", VDBKey: "weaviate-local"},
				{Name: "Col2", VDBKey: "weaviate-local"},
				{Name: "Col3", VDBKey: "milvus-cloud"},
			},
			expected: []string{"weaviate-local", "milvus-cloud"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUniqueVDBKeys(tt.specs)

			// Convert to map for order-independent comparison
			resultMap := make(map[string]bool)
			for _, key := range result {
				resultMap[key] = true
			}
			expectedMap := make(map[string]bool)
			for _, key := range tt.expected {
				expectedMap[key] = true
			}

			assert.Equal(t, expectedMap, resultMap)
			assert.Equal(t, len(tt.expected), len(result))
		})
	}
}
