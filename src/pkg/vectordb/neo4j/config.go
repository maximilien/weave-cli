// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"time"
)

// Config represents the configuration for a Neo4j client
type Config struct {
	// URI is the Neo4j connection URI (e.g., bolt://localhost:7687 or neo4j://localhost:7687)
	URI string

	// Username for authentication (default: neo4j)
	Username string

	// Password for authentication
	Password string

	// Database name (default: neo4j)
	Database string

	// MaxConnections for connection pool (default: 50)
	MaxConnections int

	// Timeout for operations (default: 30s)
	Timeout time.Duration

	// VectorDimensions for vector index (default: 1536)
	VectorDimensions int

	// SimilarityMetric for vector search (cosine or euclidean, default: cosine)
	SimilarityMetric string
}

// DefaultConfig returns a default Neo4j configuration
func DefaultConfig() *Config {
	return &Config{
		URI:              "bolt://localhost:7687",
		Username:         "neo4j",
		Password:         "",
		Database:         "neo4j",
		MaxConnections:   50,
		Timeout:          30 * time.Second,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
	}
}
