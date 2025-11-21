// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Chroma
type Factory struct{}

// NewFactory creates a new Chroma factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Chroma client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	chromaConfig := &Config{
		URL:              config.URL,
		APIKey:           config.APIKey,
		Tenant:           "default_tenant",
		Database:         config.Database,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	// Set defaults
	if chromaConfig.Database == "" {
		chromaConfig.Database = "default_database"
	}

	return NewClient(chromaConfig)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeChromaLocal,
		vectordb.VectorDBTypeChromaCloud,
	}
}

// ValidateConfig validates the configuration for Chroma
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeChromaLocal && config.Type != vectordb.VectorDBTypeChromaCloud {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Chroma type: %s", config.Type))
	}

	// Validate URL
	if config.URL == "" {
		return vectordb.ErrInvalidConfig("Chroma URL is required")
	}

	// Validate API key for cloud
	if config.Type == vectordb.VectorDBTypeChromaCloud && config.APIKey == "" {
		return vectordb.ErrInvalidConfig("API key is required for Chroma Cloud")
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
			"cosine": true,
			"l2":     true,
			"ip":     true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid similarity metric: %s (must be cosine, l2, or ip)", config.SimilarityMetric))
		}
	}

	return nil
}

// init registers the Chroma factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaCloud, factory)
}
