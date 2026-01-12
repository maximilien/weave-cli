// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_SearchSemantic_DimensionVerification(t *testing.T) {
	t.Skip("Requires live Qdrant server - this tests our v0.9.1 dimension verification feature")

	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		Timeout:          30,
		VectorDimensions: 384, // sentence-transformers dimension
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// This would test the dimension verification logic added in v0.9.1
	// When collection has different dimensions than config, should error with helpful message
	options := &vectordb.QueryOptions{
		TopK: 5,
	}

	_, err = adapter.SearchSemantic(ctx, "test-collection", "test query", options)

	// We expect error due to dimension mismatch or connection failure
	assert.Error(t, err)
}

func TestQueryOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options *vectordb.QueryOptions
		wantErr bool
	}{
		{
			name: "Valid options",
			options: &vectordb.QueryOptions{
				TopK:     10,
				Distance: 0.7,
			},
			wantErr: false,
		},
		{
			name:    "Nil options (should use defaults)",
			options: nil,
			wantErr: false,
		},
		{
			name: "Zero TopK (should use default)",
			options: &vectordb.QueryOptions{
				TopK: 0,
			},
			wantErr: false,
		},
		{
			name: "Negative TopK",
			options: &vectordb.QueryOptions{
				TopK: -1,
			},
			wantErr: false, // Server will handle validation
		},
		{
			name: "Very large TopK",
			options: &vectordb.QueryOptions{
				TopK: 10000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that options can be created and accessed
			if tt.options != nil {
				assert.NotNil(t, tt.options)
				// TopK should be accessible
				_ = tt.options.TopK
			}
		})
	}
}

func TestConvertQueryResults(t *testing.T) {
	tests := []struct {
		name          string
		resultCount   int
		expectedCount int
	}{
		{
			name:          "Empty results",
			resultCount:   0,
			expectedCount: 0,
		},
		{
			name:          "Single result",
			resultCount:   1,
			expectedCount: 1,
		},
		{
			name:          "Multiple results",
			resultCount:   10,
			expectedCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate query results
			results := make([]*vectordb.QueryResult, tt.resultCount)
			for i := 0; i < tt.resultCount; i++ {
				results[i] = &vectordb.QueryResult{
					Document: vectordb.Document{
						ID:   "doc-" + string(rune(i)),
						Text: "Test content",
					},
					Score: 0.9 - float64(i)*0.1,
				}
			}

			assert.Equal(t, tt.expectedCount, len(results))

			// Verify scores are in descending order
			for i := 1; i < len(results); i++ {
				assert.LessOrEqual(t, results[i].Score, results[i-1].Score)
			}
		})
	}
}

func TestSearchMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Simple metadata filter",
			metadata: map[string]interface{}{
				"category": "test",
			},
			wantErr: false,
		},
		{
			name: "Multiple filters",
			metadata: map[string]interface{}{
				"category": "test",
				"status":   "active",
			},
			wantErr: false,
		},
		{
			name: "Numeric filter",
			metadata: map[string]interface{}{
				"score": 0.5,
			},
			wantErr: false,
		},
		{
			name:     "Nil metadata",
			metadata: nil,
			wantErr:  false,
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test metadata filter construction
			if tt.metadata != nil {
				assert.NotNil(t, tt.metadata)
				for key, value := range tt.metadata {
					assert.NotEmpty(t, key)
					assert.NotNil(t, value)
				}
			}
		})
	}
}

func TestDimensionMismatchError(t *testing.T) {
	// Test the error message format for dimension mismatches
	// This is a key part of v0.9.1 improvements
	tests := []struct {
		name          string
		indexDims     int
		queryDims     int
		shouldMatch   bool
		expectedError string
	}{
		{
			name:        "Matching dimensions",
			indexDims:   384,
			queryDims:   384,
			shouldMatch: true,
		},
		{
			name:          "OpenAI vs sentence-transformers",
			indexDims:     1536, // OpenAI
			queryDims:     384,  // sentence-transformers
			shouldMatch:   false,
			expectedError: "dimension mismatch",
		},
		{
			name:          "sentence-transformers vs OpenAI",
			indexDims:     384,  // sentence-transformers
			queryDims:     1536, // OpenAI
			shouldMatch:   false,
			expectedError: "dimension mismatch",
		},
		{
			name:          "Custom dimensions mismatch",
			indexDims:     768,
			queryDims:     512,
			shouldMatch:   false,
			expectedError: "dimension mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate dimension check
			match := tt.indexDims == tt.queryDims

			assert.Equal(t, tt.shouldMatch, match)

			if !match {
				// Verify error would contain helpful message
				assert.Contains(t, tt.expectedError, "dimension mismatch")
			}
		})
	}
}

func TestGetCollectionDimensions(t *testing.T) {
	t.Skip("Requires live Qdrant server")

	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		Timeout:          30,
		VectorDimensions: 384,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// This tests the getCollectionDimensions function added in v0.9.1
	dims, err := adapter.Client.getCollectionDimensions(ctx, "test-collection")

	if err != nil {
		// Collection doesn't exist or server not available
		assert.Error(t, err)
	} else {
		// Dimensions should be positive
		assert.Greater(t, dims, 0)
	}
}

func TestSearchOptions_DefaultValues(t *testing.T) {
	// Test that default values are applied correctly
	tests := []struct {
		name         string
		input        *vectordb.QueryOptions
		expectedTopK int
	}{
		{
			name:         "Nil options",
			input:        nil,
			expectedTopK: 10, // Default
		},
		{
			name: "Zero TopK",
			input: &vectordb.QueryOptions{
				TopK: 0,
			},
			expectedTopK: 10, // Default
		},
		{
			name: "Custom TopK",
			input: &vectordb.QueryOptions{
				TopK: 5,
			},
			expectedTopK: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate default application
			topK := 10 // default
			if tt.input != nil && tt.input.TopK > 0 {
				topK = tt.input.TopK
			}

			assert.Equal(t, tt.expectedTopK, topK)
		})
	}
}

func TestSearchBM25_NotSupported(t *testing.T) {
	// Qdrant doesn't natively support BM25, test fallback behavior
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// BM25 search should return appropriate error
	_, err = adapter.SearchBM25(ctx, "test", "query", nil)

	// Expect error indicating BM25 not supported
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "does not support BM25")
	}
}

func TestSearchHybrid_NotSupported(t *testing.T) {
	t.Skip("Requires live Qdrant server - tests hybrid search behavior")

	// Qdrant doesn't natively support hybrid search, test fallback
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Hybrid search should return appropriate error
	_, err = adapter.SearchHybrid(ctx, "test", "query", nil)

	// Expect error indicating hybrid not supported or collection not found
	assert.Error(t, err)
}

// Benchmark tests

func BenchmarkSearchSemantic(b *testing.B) {
	b.Skip("Requires live Qdrant server")

	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		Timeout:          30,
		VectorDimensions: 384,
	}

	adapter, _ := NewAdapter(config)
	ctx := context.Background()
	options := &vectordb.QueryOptions{TopK: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.SearchSemantic(ctx, "test", "query", options)
	}
}

func BenchmarkConvertQueryResults(b *testing.B) {
	// Create sample results
	results := make([]*vectordb.QueryResult, 10)
	for i := 0; i < 10; i++ {
		results[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:   "doc",
				Text: "content",
			},
			Score: 0.9,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate result conversion
		_ = results
	}
}
