// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "Default timeout (0)",
			timeout:  0,
			expected: 0,
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
		cloudID       string
		baseTimeout   int
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Local health operation with default timeout",
			cloudID:       "",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeHealth,
			expected:      10 * time.Second, // Local default for health
		},
		{
			name:          "Cloud health operation with default timeout",
			cloudID:       "my-cluster:dXMtY2VudHJhbDEuZ2NwLmNsb3VkLmVzLmlv",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeHealth,
			expected:      20 * time.Second, // Cloud default for health
		},
		{
			name:          "Local collection operation with default timeout",
			cloudID:       "",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      20 * time.Second, // Local default for collection
		},
		{
			name:          "Cloud collection operation with default timeout",
			cloudID:       "my-cluster:dXMtY2VudHJhbDEuZ2NwLmNsb3VkLmVzLmlv",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second, // Cloud default for collection
		},
		{
			name:          "Custom timeout overrides defaults",
			cloudID:       "",
			baseTimeout:   45,
			operationType: vectordb.OperationTypeQuery,
			expected:      45 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					CloudID: tt.cloudID,
					Timeout: tt.baseTimeout,
				},
			}
			result := client.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractCloudID(t *testing.T) {
	tests := []struct {
		name     string
		config   *vectordb.Config
		expected string
	}{
		{
			name: "Valid Cloud ID format",
			config: &vectordb.Config{
				URL: "my-cluster:dXMtY2VudHJhbDEuZ2NwLmNsb3VkLmVzLmlv",
			},
			expected: "my-cluster:dXMtY2VudHJhbDEuZ2NwLmNsb3VkLmVzLmlv",
		},
		{
			name: "HTTP URL",
			config: &vectordb.Config{
				URL: "http://localhost:9200",
			},
			expected: "",
		},
		{
			name: "HTTPS URL",
			config: &vectordb.Config{
				URL: "https://my-cluster.es.io:9200",
			},
			expected: "",
		},
		{
			name: "Empty URL",
			config: &vectordb.Config{
				URL: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCloudID(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAddresses(t *testing.T) {
	tests := []struct {
		name     string
		config   *vectordb.Config
		expected []string
	}{
		{
			name: "URL with http prefix",
			config: &vectordb.Config{
				URL: "http://localhost:9200",
			},
			expected: []string{"http://localhost:9200"},
		},
		{
			name: "URL with https prefix",
			config: &vectordb.Config{
				URL: "https://my-cluster.es.io:9200",
			},
			expected: []string{"https://my-cluster.es.io:9200"},
		},
		{
			name: "Cloud ID (not an address)",
			config: &vectordb.Config{
				URL: "my-cluster:dXMtY2VudHJhbDEuZ2NwLmNsb3VkLmVzLmlv",
			},
			expected: nil,
		},
		{
			name: "Single address",
			config: &vectordb.Config{
				Address: "localhost:9200",
			},
			expected: []string{"localhost:9200"},
		},
		{
			name: "Multiple addresses",
			config: &vectordb.Config{
				Address: "host1:9200,host2:9200,host3:9200",
			},
			expected: []string{"host1:9200", "host2:9200", "host3:9200"},
		},
		{
			name: "Multiple addresses with spaces",
			config: &vectordb.Config{
				Address: "host1:9200, host2:9200 , host3:9200",
			},
			expected: []string{"host1:9200", "host2:9200", "host3:9200"},
		},
		{
			name: "Both URL and Address",
			config: &vectordb.Config{
				URL:     "http://primary:9200",
				Address: "secondary:9200,tertiary:9200",
			},
			expected: []string{"http://primary:9200", "secondary:9200", "tertiary:9200"},
		},
		{
			name: "Empty config",
			config: &vectordb.Config{
				URL:     "",
				Address: "",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAddresses(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}
