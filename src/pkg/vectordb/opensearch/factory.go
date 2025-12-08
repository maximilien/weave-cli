// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for OpenSearch
type Factory struct{}

// NewFactory creates a new OpenSearch factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new OpenSearch client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeOpenSearchLocal,
		vectordb.VectorDBTypeOpenSearchCloud,
	}
}

// ValidateConfig validates the configuration for OpenSearch
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeOpenSearchLocal && config.Type != vectordb.VectorDBTypeOpenSearchCloud {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported OpenSearch type: %s", config.Type))
	}

	// Validate URL/Address
	if config.URL == "" && config.Address == "" {
		return vectordb.ErrInvalidConfig("OpenSearch URL or address is required")
	}

	// Validate timeout
	if config.Timeout < 0 {
		return vectordb.ErrInvalidConfig("timeout cannot be negative")
	}

	// Validate vector dimensions if specified
	if config.VectorDimensions < 0 {
		return vectordb.ErrInvalidConfig("vector dimensions cannot be negative")
	}

	// Validate similarity metric if specified (OpenSearch k-NN supports specific distance metrics)
	if config.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"l2":           true, // Euclidean distance
			"cosinesimil":  true, // Cosine similarity
			"innerproduct": true, // Inner product
			"l1":           true, // Manhattan distance
			"linf":         true, // Chebyshev distance
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid similarity metric: %s (must be l2, cosinesimil, innerproduct, l1, or linf)", config.SimilarityMetric))
		}
	}

	// Cloud-specific validation (AWS OpenSearch Service)
	if config.Type == vectordb.VectorDBTypeOpenSearchCloud {
		// For AWS OpenSearch, we need either API key or AWS credentials
		// AWS credentials will be picked up from environment/config by SDK
		// So we don't strictly require APIKey here
	}

	return nil
}

// init registers the OpenSearch factory for both local and cloud types
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeOpenSearchLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeOpenSearchCloud, factory)
}
