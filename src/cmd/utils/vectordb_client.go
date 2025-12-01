// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"

	// Import adapters to register their factories
	// Note: Chroma is excluded on Windows due to CGO dependencies
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/mock"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/mongodb"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/qdrant"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

// CreateVectorDBClient creates a vector database client using the new abstraction
func CreateVectorDBClient(cfg *config.VectorDBConfig) (vectordb.VectorDBClient, error) {
	client, err := vectordb.CreateClientFromVectorDBConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector database client: %w", err)
	}
	return client, nil
}

// GetSupportedVectorDBTypes returns all supported vector database types
func GetSupportedVectorDBTypes() []vectordb.VectorDBType {
	return vectordb.GetSupportedTypes()
}

// IsVectorDBTypeSupported checks if a vector database type is supported
func IsVectorDBTypeSupported(dbType string) bool {
	return vectordb.IsSupported(vectordb.VectorDBType(dbType))
}
