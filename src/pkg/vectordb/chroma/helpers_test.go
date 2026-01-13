//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

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
			expected: 10 * time.Second,
		},
		{
			name:     "Custom timeout",
			timeout:  30,
			expected: 30 * time.Second,
		},
		{
			name:     "Small timeout",
			timeout:  5,
			expected: 5 * time.Second,
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
		url           string
		apiKey        string
		baseTimeout   int
		operationType vectordb.OperationType
		expected      time.Duration
	}{
		{
			name:          "Local collection operation with default timeout",
			url:           "http://localhost:8000",
			apiKey:        "",
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      20 * time.Second,
		},
		{
			name:          "Cloud collection operation with default timeout",
			url:           "https://api.trychroma.com",
			apiKey:        "test-api-key", // APIKey indicates cloud
			baseTimeout:   0,
			operationType: vectordb.OperationTypeCollection,
			expected:      40 * time.Second,
		},
		{
			name:          "Custom timeout overrides defaults",
			url:           "http://localhost:8000",
			apiKey:        "",
			baseTimeout:   30,
			operationType: vectordb.OperationTypeQuery,
			expected:      30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					URL:     tt.url,
					APIKey:  tt.apiKey,
					Timeout: tt.baseTimeout,
				},
			}
			result := client.getTimeoutFor(tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
