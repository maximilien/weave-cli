// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "Default timeout (0)",
			timeout:  0,
			expected: 30 * time.Second,
		},
		{
			name:     "Custom timeout",
			timeout:  60,
			expected: 60 * time.Second,
		},
		{
			name:     "Small timeout",
			timeout:  10,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{
				config: &vectordb.Config{
					Timeout: tt.timeout,
				},
			}
			result := adapter.getTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdapter_GetTimeoutFor(t *testing.T) {
	tests := []struct {
		name          string
		baseTimeout   int
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Health operation with default timeout",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeHealth,
			expected:      20 * time.Second, // Cloud default for health
		},
		{
			name:          "Collection operation with default timeout",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second, // Cloud default for collection
		},
		{
			name:          "Query operation with default timeout",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeQuery,
			expected:      40 * time.Second, // Cloud default for query
		},
		{
			name:          "Custom timeout overrides defaults",
			baseTimeout:   45,
			operationType: vectordb.OperationTypeQuery,
			expected:      45 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{
				config: &vectordb.Config{
					Timeout: tt.baseTimeout,
				},
			}
			result := adapter.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
