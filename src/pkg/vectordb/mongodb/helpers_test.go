// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestGetTimeoutForOperations(t *testing.T) {
	config := &Config{
		URI:      "mongodb://localhost:27017",
		Database: "test",
		Timeout:  30,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Skip("Skipping test that requires MongoDB connection")
	}

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

func TestContainsAnyHelper(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substrs  []string
		expected bool
	}{
		{
			name:     "Contains +srv",
			str:      "mongodb+srv://cluster.mongodb.net",
			substrs:  []string{"+srv"},
			expected: true,
		},
		{
			name:     "Contains atlas",
			str:      "mongodb+srv://atlas.mongodb.net",
			substrs:  []string{"atlas"},
			expected: true,
		},
		{
			name:     "Contains none",
			str:      "mongodb://localhost:27017",
			substrs:  []string{"+srv", "atlas"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert slice to variadic args
			result := containsAny(tt.str, tt.substrs...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdapter_CloseMethod(t *testing.T) {
	config := &vectordb.Config{
		URL:     "mongodb://localhost:27017",
		Timeout: 30,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Skip("Skipping test that requires MongoDB connection")
	}

	ctx := context.Background()

	// Close should work (may fail if not connected, but shouldn't panic)
	err = adapter.Close(ctx)
	// We don't assert NoError because connection might not be established
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		_ = adapter.Close(ctx)
	})
}
