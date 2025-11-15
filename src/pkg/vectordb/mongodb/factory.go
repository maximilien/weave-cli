// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for MongoDB
type Factory struct{}

// NewFactory creates a new MongoDB factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new MongoDB client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeMongoDB,
	}
}

// ValidateConfig validates the configuration for MongoDB
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeMongoDB {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported MongoDB type: %s", config.Type))
	}

	// Validate URL (MongoDB URI)
	if config.URL == "" {
		return vectordb.ErrInvalidConfig("MongoDB URI is required")
	}

	// Validate database name
	if config.Database == "" {
		return vectordb.ErrInvalidConfig("MongoDB database name is required")
	}

	// Validate timeout
	if config.Timeout < 0 {
		return vectordb.ErrInvalidConfig("timeout cannot be negative")
	}

	// Validate vector dimensions if specified
	if config.VectorDimensions < 0 {
		return vectordb.ErrInvalidConfig("vector dimensions cannot be negative")
	}

	// Validate similarity metric if specified
	if config.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"cosine":     true,
			"euclidean":  true,
			"dotProduct": true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid similarity metric: %s (must be cosine, euclidean, or dotProduct)", config.SimilarityMetric))
		}
	}

	return nil
}

// init registers the MongoDB factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeMongoDB, factory)
}
