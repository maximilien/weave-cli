// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vectordb

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// VectorDBType represents the type of vector database
type VectorDBType string

const (
	VectorDBTypeWeaviateCloud VectorDBType = "weaviate-cloud"
	VectorDBTypeWeaviateLocal VectorDBType = "weaviate-local"
	VectorDBTypeMock          VectorDBType = "mock"
	VectorDBTypeSupabase      VectorDBType = "supabase"
	VectorDBTypeMilvusLocal   VectorDBType = "milvus-local"
	VectorDBTypeMilvusCloud   VectorDBType = "milvus-cloud"
	VectorDBTypeMongoDB       VectorDBType = "mongodb"
	VectorDBTypeChromaLocal   VectorDBType = "chroma-local"
	VectorDBTypeChromaCloud   VectorDBType = "chroma-cloud"
	VectorDBTypeQdrantLocal   VectorDBType = "qdrant-local"
	VectorDBTypeQdrantCloud   VectorDBType = "qdrant-cloud"
)

// Config represents the configuration for a vector database client
type Config struct {
	Type         VectorDBType `yaml:"type"`
	URL          string       `yaml:"url,omitempty"`
	APIKey       string       `yaml:"api_key,omitempty"`
	OpenAIAPIKey string       `yaml:"openai_api_key,omitempty"`
	Timeout      int          `yaml:"timeout,omitempty"`

	// Mock-specific configuration
	Enabled            bool `yaml:"enabled,omitempty"`
	SimulateEmbeddings bool `yaml:"simulate_embeddings,omitempty"`
	EmbeddingDimension int  `yaml:"embedding_dimension,omitempty"`

	// Supabase-specific configuration
	DatabaseURL string `yaml:"database_url,omitempty"`
	DatabaseKey string `yaml:"database_key,omitempty"`

	// Milvus-specific configuration
	Address  string `yaml:"address,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
	Tenant   string `yaml:"tenant,omitempty"`

	// MongoDB-specific configuration
	VectorDimensions int    `yaml:"vector_dimensions,omitempty"`
	SimilarityMetric string `yaml:"similarity_metric,omitempty"`
}

// ClientFactory defines the interface for creating vector database clients
type ClientFactory interface {
	// CreateClient creates a new vector database client
	CreateClient(config *Config) (VectorDBClient, error)

	// GetSupportedTypes returns the list of supported database types
	GetSupportedTypes() []VectorDBType

	// ValidateConfig validates the configuration for the specific database type
	ValidateConfig(config *Config) error
}

// Registry holds all registered client factories
type Registry struct {
	factories map[VectorDBType]ClientFactory
}

// NewRegistry creates a new factory registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[VectorDBType]ClientFactory),
	}
}

// Register registers a client factory for a specific database type
func (r *Registry) Register(dbType VectorDBType, factory ClientFactory) {
	r.factories[dbType] = factory
}

// CreateClient creates a client using the appropriate factory
func (r *Registry) CreateClient(config *Config) (VectorDBClient, error) {
	factory, exists := r.factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported vector database type: %s", config.Type)
	}

	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration for %s: %w", config.Type, err)
	}

	return factory.CreateClient(config)
}

// GetSupportedTypes returns all supported database types
func (r *Registry) GetSupportedTypes() []VectorDBType {
	types := make([]VectorDBType, 0, len(r.factories))
	for dbType := range r.factories {
		types = append(types, dbType)
	}
	return types
}

// IsSupported checks if a database type is supported
func (r *Registry) IsSupported(dbType VectorDBType) bool {
	_, exists := r.factories[dbType]
	return exists
}

// Global registry instance
var globalRegistry = NewRegistry()

// RegisterFactory registers a client factory globally
func RegisterFactory(dbType VectorDBType, factory ClientFactory) {
	globalRegistry.Register(dbType, factory)
}

// CreateClient creates a client using the global registry
func CreateClient(config *Config) (VectorDBClient, error) {
	return globalRegistry.CreateClient(config)
}

// CreateClientFromVectorDBConfig creates a client from the legacy VectorDBConfig
func CreateClientFromVectorDBConfig(cfg *config.VectorDBConfig) (VectorDBClient, error) {
	// Convert legacy config to new config format
	newConfig := &Config{
		Type:               VectorDBType(cfg.Type),
		URL:                cfg.URL,
		APIKey:             cfg.APIKey,
		DatabaseURL:        cfg.DatabaseURL,      // For Supabase
		DatabaseKey:        cfg.DatabaseKey,      // For Supabase
		Database:           cfg.Database,         // For MongoDB, Milvus, Chroma
		Tenant:             cfg.Tenant,           // For Chroma Cloud
		VectorDimensions:   cfg.VectorDimensions, // For MongoDB, Milvus
		SimilarityMetric:   cfg.SimilarityMetric, // For MongoDB, Milvus
		Address:            cfg.Address,          // For Milvus
		Username:           cfg.Username,         // For Milvus Cloud
		Password:           cfg.Password,         // For Milvus Cloud
		OpenAIAPIKey:       cfg.OpenAIAPIKey,
		Timeout:            cfg.Timeout,
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
	}

	return CreateClient(newConfig)
}

// GetSupportedTypes returns all supported database types from the global registry
func GetSupportedTypes() []VectorDBType {
	return globalRegistry.GetSupportedTypes()
}

// IsSupported checks if a database type is supported in the global registry
func IsSupported(dbType VectorDBType) bool {
	return globalRegistry.IsSupported(dbType)
}
