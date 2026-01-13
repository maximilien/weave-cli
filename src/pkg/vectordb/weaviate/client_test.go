// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_HTTP(t *testing.T) {
	config := &Config{
		URL:     "http://localhost:8080",
		Timeout: 10,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, config, client.config)
}

func TestNewClient_HTTPS(t *testing.T) {
	config := &Config{
		URL:     "https://my-cluster.weaviate.network",
		APIKey:  "test-api-key",
		Timeout: 30,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestNewClient_WithOpenAIKey(t *testing.T) {
	config := &Config{
		URL:          "https://my-cluster.weaviate.network",
		APIKey:       "test-api-key",
		OpenAIAPIKey: "openai-key",
		Timeout:      20,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_NoProtocol(t *testing.T) {
	config := &Config{
		URL:     "localhost:8080",
		Timeout: 10,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_LocalNoAuth(t *testing.T) {
	config := &Config{
		URL:     "http://localhost:8080",
		Timeout: 5,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "", client.config.APIKey)
}

func TestNewClient_CloudWithAuth(t *testing.T) {
	config := &Config{
		URL:     "https://cluster.weaviate.network",
		APIKey:  "weaviate-api-key",
		Timeout: 15,
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "weaviate-api-key", client.config.APIKey)
}
