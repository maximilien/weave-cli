// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Redis
type Factory struct{}

// NewFactory creates a new Redis factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Redis vector database client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeRedisLocal,
		vectordb.VectorDBTypeRedisCloud,
	}
}

// ValidateConfig validates the configuration for Redis
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	if config.Type != vectordb.VectorDBTypeRedisLocal && config.Type != vectordb.VectorDBTypeRedisCloud {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Redis type: %s", config.Type))
	}

	if config.URL == "" && config.Address == "" {
		return vectordb.ErrInvalidConfig("Redis URL or address is required")
	}

	if config.Timeout < 0 {
		return vectordb.ErrInvalidConfig("timeout cannot be negative")
	}

	if config.VectorDimensions < 0 {
		return vectordb.ErrInvalidConfig("vector dimensions cannot be negative")
	}

	if config.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"COSINE": true,
			"L2":     true,
			"IP":     true,
		}
		if !validMetrics[config.SimilarityMetric] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf(
				"invalid similarity metric: %s (must be COSINE, L2, or IP)",
				config.SimilarityMetric))
		}
	}

	if config.Type == vectordb.VectorDBTypeRedisCloud {
		if config.Password == "" && config.APIKey == "" {
			return vectordb.ErrInvalidConfig("Redis Cloud requires a password or API key")
		}
	}

	return nil
}

// init registers the Redis factory for both local and cloud types
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeRedisLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeRedisCloud, factory)
}
