// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Qdrant
type Factory struct{}

// NewFactory creates a new Qdrant factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Qdrant client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeQdrantLocal,
		vectordb.VectorDBTypeQdrantCloud,
	}
}

// ValidateConfig validates the configuration for Qdrant
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeQdrantLocal && config.Type != vectordb.VectorDBTypeQdrantCloud {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Qdrant type: %s", config.Type))
	}

	// Validate URL/Address
	if config.URL == "" && config.Address == "" {
		return vectordb.ErrInvalidConfig("Qdrant URL or address is required")
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
			"Cosine":    true,
			"Dot":       true,
			"Euclidean": true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid similarity metric: %s (must be Cosine, Dot, or Euclidean)", config.SimilarityMetric))
		}
	}

	// Cloud-specific validation
	if config.Type == vectordb.VectorDBTypeQdrantCloud {
		if config.APIKey == "" {
			return vectordb.ErrInvalidConfig("Qdrant Cloud requires an API key")
		}
	}

	return nil
}

// init registers the Qdrant factory for both local and cloud types
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeQdrantLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeQdrantCloud, factory)
}
