//go:build windows
// +build windows

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Chroma (Windows stub)
type Factory struct{}

// NewFactory creates a new Chroma factory (Windows stub)
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Chroma client (Windows stub - returns error)
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return nil, fmt.Errorf("Chroma is not supported on Windows due to CGO dependencies")
}

// GetSupportedTypes returns empty list on Windows
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{}
}

// ValidateConfig validates the configuration for Chroma (Windows stub)
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	return fmt.Errorf("Chroma is not supported on Windows due to CGO dependencies")
}

// init does not register Chroma on Windows
func init() {
	// Chroma is not available on Windows, so we don't register it
}
