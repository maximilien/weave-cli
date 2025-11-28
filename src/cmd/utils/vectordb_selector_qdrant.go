// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getQdrantLocalConfig returns Qdrant Local configuration
func getQdrantLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Qdrant Local type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeQdrantLocal {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if qdrantConfig := tryCreateQdrantLocalConfigFromEnv(); qdrantConfig != nil {
		return qdrantConfig, nil
	}

	return nil, fmt.Errorf("no Qdrant Local configuration found")
}

// getQdrantCloudConfig returns Qdrant Cloud configuration
func getQdrantCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Qdrant Cloud type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeQdrantCloud {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if qdrantConfig := tryCreateQdrantCloudConfigFromEnv(); qdrantConfig != nil {
		return qdrantConfig, nil
	}

	return nil, fmt.Errorf("no Qdrant Cloud configuration found")
}

// tryCreateQdrantLocalConfigFromEnv attempts to create a Qdrant Local config from environment variables
func tryCreateQdrantLocalConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("QDRANT_URL")

	// Require URL for local
	if url == "" {
		url = "localhost:6334" // Default gRPC port
	}

	return &config.VectorDBConfig{
		Name:             "qdrant-local",
		Type:             config.VectorDBTypeQdrantLocal,
		URL:              url,
		VectorDimensions: 1536,
		SimilarityMetric: "Cosine",
		Timeout:          30,
	}
}

// tryCreateQdrantCloudConfigFromEnv attempts to create a Qdrant Cloud config from environment variables
func tryCreateQdrantCloudConfigFromEnv() *config.VectorDBConfig {
	// Try QDRANT_CLOUD_API_KEY first, then fall back to QDRANT_API_KEY
	apiKey := os.Getenv("QDRANT_CLOUD_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("QDRANT_API_KEY")
	}

	// Require API key for cloud
	if apiKey == "" {
		return nil
	}

	// URL for Qdrant Cloud
	url := getEnvOrDefault("QDRANT_CLOUD_URL", "")
	if url == "" {
		url = os.Getenv("QDRANT_URL")
	}

	if url == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:             "qdrant-cloud",
		Type:             config.VectorDBTypeQdrantCloud,
		URL:              url,
		APIKey:           apiKey,
		VectorDimensions: 1536,
		SimilarityMetric: "Cosine",
		Timeout:          60,
	}
}
