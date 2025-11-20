// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Milvus
type Factory struct{}

// NewFactory creates a new Milvus factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Milvus client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeMilvusLocal,
		vectordb.VectorDBTypeMilvusCloud,
	}
}

// ValidateConfig validates the configuration for Milvus
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeMilvusLocal && config.Type != vectordb.VectorDBTypeMilvusCloud {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Milvus type: %s", config.Type))
	}

	// Validate address
	if config.Address == "" {
		return vectordb.ErrInvalidConfig("Milvus address is required")
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
			"L2":     true,
			"IP":     true,
			"COSINE": true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid similarity metric: %s (must be L2, IP, or COSINE)", config.SimilarityMetric))
		}
	}

	// Cloud-specific validation
	if config.Type == vectordb.VectorDBTypeMilvusCloud {
		// Cloud instances require authentication (either APIKey or username/password)
		hasAPIKey := config.APIKey != ""
		hasUserPass := config.Username != "" && config.Password != ""
		if !hasAPIKey && !hasUserPass {
			return vectordb.ErrInvalidConfig("Milvus cloud requires either APIKey or username and password")
		}
	}

	return nil
}

// init registers the Milvus factory for both local and cloud types
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusCloud, factory)
}
