//go:build !(darwin && (amd64 || arm64))

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// ChromaStubFactory implements the ClientFactory interface for unsupported platforms
type ChromaStubFactory struct{}

// CreateClient returns an error indicating Chroma is not supported on this platform
func (f *ChromaStubFactory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return nil, fmt.Errorf("Chroma is only supported on macOS (AMD64/ARM64) - not available on this platform due to CGO dependencies")
}

// ValidateConfig returns an error indicating Chroma is not supported on this platform
func (f *ChromaStubFactory) ValidateConfig(config *vectordb.Config) error {
	return fmt.Errorf("Chroma is only supported on macOS (AMD64/ARM64) - not available on this platform due to CGO dependencies")
}

// GetSupportedTypes returns the list of supported database types
func (f *ChromaStubFactory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeChromaLocal,
		vectordb.VectorDBTypeChromaCloud,
	}
}

// init registers the Chroma stub factory on unsupported platforms
func init() {
	factory := &ChromaStubFactory{}
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaCloud, factory)
}
