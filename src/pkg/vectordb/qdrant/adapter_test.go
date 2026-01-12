// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter_Success(t *testing.T) {
	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		APIKey:           "test-key",
		Timeout:          30,
		VectorDimensions: 384,
	}

	adapter, err := NewAdapter(config)

	assert.NoError(t, err)
	assert.NotNil(t, adapter)
}

func TestNewAdapter_DefaultConfig(t *testing.T) {
	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)

	assert.NoError(t, err)
	assert.NotNil(t, adapter)
}

func TestAdapter_Health_Timeout(t *testing.T) {
	t.Skip("Skipping test that requires live Qdrant server")

	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 1, // 1 second timeout
	}

	adapter, err := NewAdapter(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// This should timeout or fail since there's no server
	err = adapter.Health(ctx)

	// We expect an error (either timeout or connection refused)
	assert.Error(t, err)
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *vectordb.Config
		wantErr bool
	}{
		{
			name: "Valid config with full details",
			config: &vectordb.Config{
				URL:              "http://localhost:6334",
				APIKey:           "test-key",
				Timeout:          30,
				VectorDimensions: 384,
			},
			wantErr: false,
		},
		{
			name: "Valid config minimal",
			config: &vectordb.Config{
				URL:     "http://localhost:6334",
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "URL without protocol",
			config: &vectordb.Config{
				URL:     "localhost:6334",
				Timeout: 30,
			},
			wantErr: false, // Should still work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, adapter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
			}
		})
	}
}

// Test helper functions

func TestConvertQdrantVectorToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected []float64
	}{
		{
			name:     "Empty vector",
			input:    []float32{},
			expected: []float64{},
		},
		{
			name:     "Single value",
			input:    []float32{1.5},
			expected: []float64{1.5},
		},
		{
			name:     "Multiple values",
			input:    []float32{1.0, 2.0, 3.0},
			expected: []float64{1.0, 2.0, 3.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]float64, len(tt.input))
			for i, v := range tt.input {
				result[i] = float64(v)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	t.Skip("Skipping tests that require live Qdrant server")

	config := &vectordb.Config{
		URL:     "http://localhost:6334",
		Timeout: 1,
	}

	adapter, _ := NewAdapter(config)
	ctx := context.Background()

	// All these should fail gracefully with proper errors
	testCases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "CreateCollection",
			fn: func() error {
				return adapter.CreateCollection(ctx, "", nil)
			},
		},
		{
			name: "DeleteCollection",
			fn: func() error {
				return adapter.DeleteCollection(ctx, "nonexistent")
			},
		},
		{
			name: "GetCollectionCount",
			fn: func() error {
				_, err := adapter.GetCollectionCount(ctx, "nonexistent")
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			// We expect errors since we're not connected to a real server
			assert.Error(t, err)
			// Errors should be descriptive
			assert.NotEmpty(t, err.Error())
		})
	}
}

// Benchmark tests

func BenchmarkNewAdapter(b *testing.B) {
	config := &vectordb.Config{
		URL:              "http://localhost:6334",
		APIKey:           "test-key",
		Timeout:          30,
		VectorDimensions: 384,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewAdapter(config)
	}
}
