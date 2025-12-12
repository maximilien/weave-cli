// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Elasticsearch
type Factory struct{}

// NewFactory creates a new Elasticsearch factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Elasticsearch client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeElasticsearchLocal,
		vectordb.VectorDBTypeElasticsearchCloud,
	}
}

// ValidateConfig validates the configuration for Elasticsearch
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	validTypes := map[vectordb.VectorDBType]bool{
		vectordb.VectorDBTypeElasticsearchLocal: true,
		vectordb.VectorDBTypeElasticsearchCloud: true,
	}
	if !validTypes[config.Type] {
		return vectordb.ErrInvalidConfig(
			fmt.Sprintf("unsupported Elasticsearch type: %s", config.Type))
	}

	// Validate required fields (URL, Address, or APIKey must be provided)
	if config.URL == "" && config.Address == "" && config.APIKey == "" {
		return vectordb.ErrInvalidConfig(
			"URL, Address, or APIKey is required for Elasticsearch connection")
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
			"cosine":      true,
			"dot_product": true,
			"l2_norm":     true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(
				fmt.Sprintf("invalid similarity metric: %s (must be cosine, dot_product, or l2_norm)",
					config.SimilarityMetric))
		}
	}

	return nil
}

// init registers the Elasticsearch factory globally
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeElasticsearchLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeElasticsearchCloud, factory)
}
