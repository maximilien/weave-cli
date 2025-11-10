// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Mock
type Factory struct{}

// NewFactory creates a new Mock factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Mock client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeMock,
	}
}

// ValidateConfig validates the configuration for Mock
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeMock {
		return vectordb.ErrInvalidConfig("unsupported database type for mock factory")
	}

	// Validate embedding dimension
	if config.EmbeddingDimension < 0 {
		return vectordb.ErrInvalidConfig("embedding dimension cannot be negative")
	}

	return nil
}

// init registers the Mock factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeMock, factory)
}
