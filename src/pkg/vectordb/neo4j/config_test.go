// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "bolt://localhost:7687", config.URI)
	assert.Equal(t, "neo4j", config.Username)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, "neo4j", config.Database)
	assert.Equal(t, 50, config.MaxConnections)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 1536, config.VectorDimensions)
	assert.Equal(t, "cosine", config.SimilarityMetric)
}
