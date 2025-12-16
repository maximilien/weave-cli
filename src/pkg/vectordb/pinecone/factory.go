// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements vectordb.ClientFactory for Pinecone
type Factory struct{}

// NewFactory creates a new Pinecone factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Pinecone client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	if config.Type != vectordb.VectorDBTypePinecone {
		return nil, fmt.Errorf("invalid config type for Pinecone: %s", config.Type)
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone adapter: %w", err)
	}

	return adapter, nil
}

// GetSupportedTypes returns the database types supported by this factory
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypePinecone,
	}
}

// ValidateConfig validates the Pinecone configuration
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type != vectordb.VectorDBTypePinecone {
		return fmt.Errorf("invalid config type: expected Pinecone, got %s", config.Type)
	}

	// API key is required
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for Pinecone")
	}

	// URL/host is optional for serverless Pinecone
	// It will be determined per-index

	return nil
}

// init registers the Pinecone factory with the global registry
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypePinecone, factory)
}
