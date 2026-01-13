// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "Default timeout (30s)",
			timeout:  30 * time.Second,
			expected: 30 * time.Second,
		},
		{
			name:     "Custom timeout (60s)",
			timeout:  60 * time.Second,
			expected: 60 * time.Second,
		},
		{
			name:     "Small timeout (10s)",
			timeout:  10 * time.Second,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					Timeout: tt.timeout,
				},
			}
			result := client.getTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_GetTimeoutFor(t *testing.T) {
	tests := []struct {
		name          string
		uri           string
		baseTimeout   time.Duration
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Local health operation with default timeout",
			uri:           "bolt://localhost:7687",
			baseTimeout:   30 * time.Second,
			operationType: vectordb.OperationTypeHealth,
			expected:      30 * time.Second, // Custom timeout overrides
		},
		{
			name:          "Cloud health operation (neo4j+s) with default timeout",
			uri:           "neo4j+s://my-cluster.databases.neo4j.io",
			baseTimeout:   30 * time.Second,
			operationType: vectordb.OperationTypeHealth,
			expected:      30 * time.Second, // Custom timeout overrides
		},
		{
			name:          "Cloud health operation (aura) with default timeout",
			uri:           "neo4j://my-cluster.aura.cloud",
			baseTimeout:   30 * time.Second,
			operationType: vectordb.OperationTypeHealth,
			expected:      30 * time.Second, // Custom timeout overrides
		},
		{
			name:          "Local collection operation with zero timeout",
			uri:           "bolt://localhost:7687",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      20 * time.Second, // Local default for collection
		},
		{
			name:          "Cloud collection operation with zero timeout",
			uri:           "neo4j+s://my-cluster.databases.neo4j.io",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second, // Cloud default for collection
		},
		{
			name:          "Local query operation with zero timeout",
			uri:           "bolt://localhost:7687",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeQuery,
			expected:      20 * time.Second, // Local default for query
		},
		{
			name:          "Cloud query operation with zero timeout",
			uri:           "neo4j+s://my-cluster.databases.neo4j.io",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeQuery,
			expected:      40 * time.Second, // Cloud default for query
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					URI:     tt.uri,
					Timeout: tt.baseTimeout,
				},
			}
			result := client.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
