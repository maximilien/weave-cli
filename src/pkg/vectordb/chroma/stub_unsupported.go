//go:build !(linux && amd64) && !(darwin && amd64) && !(darwin && arm64)
// +build !linux !amd64
// +build !darwin !amd64
// +build !darwin !arm64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

// Package chroma stub for unsupported platforms
// Chroma is only supported on Linux AMD64, macOS AMD64, and macOS ARM64
package chroma

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Chroma (unsupported platforms stub)
type Factory struct{}

// NewFactory creates a new Chroma factory (stub)
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient returns an error indicating Chroma is not supported on this platform
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return nil, fmt.Errorf("Chroma is only supported on Linux AMD64, macOS AMD64, and macOS ARM64")
}

// ValidateConfig returns an error indicating Chroma is not supported on this platform
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	return fmt.Errorf("Chroma is only supported on Linux AMD64, macOS AMD64, and macOS ARM64")
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeChromaLocal,
		vectordb.VectorDBTypeChromaCloud,
	}
}

// init registers the Chroma factory (stub that returns errors)
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeChromaCloud, factory)
}
