// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the ClientFactory interface for Supabase
type Factory struct{}

// NewFactory creates a new Supabase factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Supabase client
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	return NewAdapter(config)
}

// GetSupportedTypes returns the list of supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeSupabase,
	}
}

// ValidateConfig validates the configuration for Supabase
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config == nil {
		return vectordb.ErrInvalidConfig("config cannot be nil")
	}

	// Validate database type
	if config.Type != vectordb.VectorDBTypeSupabase {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("unsupported Supabase type: %s", config.Type))
	}

	// Validate database URL
	if config.DatabaseURL == "" {
		return vectordb.ErrInvalidConfig("database URL is required for Supabase")
	}

	// Validate URL format - should be a PostgreSQL connection string
	if !strings.Contains(config.DatabaseURL, "postgresql://") && !strings.Contains(config.DatabaseURL, "postgres://") {
		return vectordb.ErrInvalidConfig("database URL must be a valid PostgreSQL connection string")
	}

	// Validate database key (anon key)
	if config.DatabaseKey == "" {
		return vectordb.ErrInvalidConfig("database key (anon key) is required for Supabase")
	}

	// Validate timeout
	if config.Timeout < 0 {
		return vectordb.ErrInvalidConfig("timeout cannot be negative")
	}

	// Validate URL contains required components
	if !strings.Contains(config.DatabaseURL, "@") {
		return vectordb.ErrInvalidConfig("database URL must contain authentication information")
	}

	// Check if URL contains supabase.co (for hosted Supabase instances)
	if strings.Contains(config.DatabaseURL, "supabase.co") {
		// Additional validation for hosted Supabase
		if !strings.Contains(config.DatabaseURL, "db.") {
			return vectordb.ErrInvalidConfig("Supabase hosted database URL should contain 'db.' subdomain")
		}
	}

	return nil
}

// init registers the Supabase factory
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, factory)
}
