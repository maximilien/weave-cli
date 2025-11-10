// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Weaviate
type Factory struct{}

// NewFactory creates a new Weaviate factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Weaviate client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeWeaviateCloud,
		vectordb.VectorDBTypeWeaviateLocal,
	}
}

// ValidateConfig validates the configuration for Weaviate
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeWeaviateCloud && config.Type != vectordb.VectorDBTypeWeaviateLocal {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Weaviate type: %s", config.Type))
	}

	// Validate URL
	if config.URL == "" {
		return vectordb.ErrInvalidConfig("URL is required for Weaviate")
	}

	// Validate URL format
	if !strings.HasPrefix(config.URL, "http://") && !strings.HasPrefix(config.URL, "https://") {
		return vectordb.ErrInvalidConfig("URL must start with http:// or https://")
	}

	// Validate cloud-specific requirements
	if config.Type == vectordb.VectorDBTypeWeaviateCloud {
		if config.APIKey == "" {
			return vectordb.ErrInvalidConfig("API key is required for Weaviate Cloud")
		}

		if config.OpenAIAPIKey == "" {
			return vectordb.ErrInvalidConfig("OpenAI API key is required for Weaviate Cloud")
		}

		// Cloud instances should use HTTPS
		if !strings.HasPrefix(config.URL, "https://") {
			return vectordb.ErrInvalidConfig("Weaviate Cloud requires HTTPS URL")
		}
	}

	// Validate local-specific requirements
	if config.Type == vectordb.VectorDBTypeWeaviateLocal {
		// Local instances typically use HTTP and localhost
		// HTTPS with non-localhost is unusual but not an error
		_ = strings.HasPrefix(config.URL, "https://") && !strings.Contains(config.URL, "localhost")
	}

	// Validate timeout
	if config.Timeout < 0 {
		return vectordb.ErrInvalidConfig("timeout cannot be negative")
	}

	return nil
}

// init registers the Weaviate factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeWeaviateCloud, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeWeaviateLocal, factory)
}
