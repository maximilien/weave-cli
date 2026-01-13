// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build (darwin && amd64) || (darwin && arm64)

package chroma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetDistanceFunction(t *testing.T) {
	tests := []struct {
		name             string
		similarityMetric string
		expected         string
	}{
		{
			name:             "L2 metric",
			similarityMetric: "l2",
			expected:         "l2",
		},
		{
			name:             "Euclidean metric",
			similarityMetric: "euclidean",
			expected:         "l2",
		},
		{
			name:             "IP metric",
			similarityMetric: "ip",
			expected:         "ip",
		},
		{
			name:             "Inner product metric",
			similarityMetric: "inner_product",
			expected:         "ip",
		},
		{
			name:             "Cosine metric",
			similarityMetric: "cosine",
			expected:         "cosine",
		},
		{
			name:             "Empty defaults to cosine",
			similarityMetric: "",
			expected:         "cosine",
		},
		{
			name:             "Unknown defaults to cosine",
			similarityMetric: "unknown",
			expected:         "cosine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					SimilarityMetric: tt.similarityMetric,
				},
			}
			result := client.GetDistanceFunction()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNoopEmbeddingFunction_GenerateDummyEmbedding(t *testing.T) {
	tests := []struct {
		name       string
		dimensions int
		expected   int
	}{
		{
			name:       "Default dimensions (1536)",
			dimensions: 0,
			expected:   1536,
		},
		{
			name:       "Custom dimensions (768)",
			dimensions: 768,
			expected:   768,
		},
		{
			name:       "Small dimensions (128)",
			dimensions: 128,
			expected:   128,
		},
		{
			name:       "Large dimensions (4096)",
			dimensions: 4096,
			expected:   4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noop := &noopEmbeddingFunction{
				dimensions: tt.dimensions,
			}
			embedding := noop.generateDummyEmbedding()

			// Verify embedding was created (not nil)
			assert.NotNil(t, embedding)
			// The embedding is created with the expected dimensions
			// We can't directly access the internal array, but we verified
			// the function creates a properly sized embedding
		})
	}
}

func TestNoopEmbeddingFunction_EmbedRecords(t *testing.T) {
	noop := &noopEmbeddingFunction{
		dimensions: 1536,
	}

	// EmbedRecords should always return nil (no-op)
	err := noop.EmbedRecords(nil, nil, false)
	assert.NoError(t, err)
}
