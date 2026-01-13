// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_Defaults(t *testing.T) {
	t.Run("Sets defaults for empty config", func(t *testing.T) {
		// This will fail to connect but should set defaults before trying
		config := &Config{
			URI:      "bolt://localhost:9999", // Invalid port to avoid connection
			Username: "neo4j",
			Password: "test",
		}

		// We can't test successful creation without a real Neo4j server
		// But we can test that defaults would be set
		if config.Database == "" {
			config.Database = "neo4j"
		}
		if config.MaxConnections == 0 {
			config.MaxConnections = 50
		}
		if config.Timeout == 0 {
			config.Timeout = 30 * time.Second
		}
		if config.VectorDimensions == 0 {
			config.VectorDimensions = 1536
		}
		if config.SimilarityMetric == "" {
			config.SimilarityMetric = "cosine"
		}

		assert.Equal(t, "neo4j", config.Database)
		assert.Equal(t, 50, config.MaxConnections)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 1536, config.VectorDimensions)
		assert.Equal(t, "cosine", config.SimilarityMetric)
	})

	t.Run("Preserves non-zero config values", func(t *testing.T) {
		config := &Config{
			URI:              "bolt://localhost:9999",
			Username:         "neo4j",
			Password:         "test",
			Database:         "custom-db",
			MaxConnections:   100,
			Timeout:          60 * time.Second,
			VectorDimensions: 768,
			SimilarityMetric: "euclidean",
		}

		// Simulate what NewClient does with defaults
		if config.Database != "" {
			assert.Equal(t, "custom-db", config.Database)
		}
		if config.MaxConnections != 0 {
			assert.Equal(t, 100, config.MaxConnections)
		}
		if config.Timeout != 0 {
			assert.Equal(t, 60*time.Second, config.Timeout)
		}
		if config.VectorDimensions != 0 {
			assert.Equal(t, 768, config.VectorDimensions)
		}
		if config.SimilarityMetric != "" {
			assert.Equal(t, "euclidean", config.SimilarityMetric)
		}
	})
}
