//go:build linux && amd64
// +build linux,amd64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getChromaLocalConfig returns Chroma Local configuration
func getChromaLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Chroma Local type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeChromaLocal {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if chromaConfig := tryCreateChromaLocalConfigFromEnv(); chromaConfig != nil {
		return chromaConfig, nil
	}

	return nil, fmt.Errorf("no Chroma Local configuration found")
}

// getChromaCloudConfig returns Chroma Cloud configuration
func getChromaCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Chroma Cloud type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeChromaCloud {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if chromaConfig := tryCreateChromaCloudConfigFromEnv(); chromaConfig != nil {
		return chromaConfig, nil
	}

	return nil, fmt.Errorf("no Chroma Cloud configuration found")
}

// tryCreateChromaLocalConfigFromEnv attempts to create a Chroma Local config from environment variables
func tryCreateChromaLocalConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("CHROMA_URL")

	// Require URL for local
	if url == "" {
		return nil
	}

	// For local Chroma, use CHROMA_LOCAL_DATABASE to avoid conflict with cloud CHROMA_DATABASE
	// Local Chroma typically uses default_database/default_tenant
	database := os.Getenv("CHROMA_LOCAL_DATABASE")
	if database == "" {
		database = "default_database"
	}

	return &config.VectorDBConfig{
		Name:             "chroma-local",
		Type:             config.VectorDBTypeChromaLocal,
		URL:              url,
		Database:         database,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}
}

// tryCreateChromaCloudConfigFromEnv attempts to create a Chroma Cloud config from environment variables
func tryCreateChromaCloudConfigFromEnv() *config.VectorDBConfig {
	// Try CHROMA_CLOUD_API_KEY first, then fall back to CHROMA_API_KEY
	apiKey := os.Getenv("CHROMA_CLOUD_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CHROMA_API_KEY")
	}

	// Require API key for cloud
	if apiKey == "" {
		return nil
	}

	// For Chroma Cloud, tenant is the team UUID and database is the database name
	tenant := getEnvOrDefault("CHROMA_TENANT", "default_tenant")
	database := getEnvOrDefault("CHROMA_DATABASE", "default_database")

	// URL is not needed for cloud client (it uses api.trychroma.com automatically)
	// but we keep it in config for informational purposes
	url := getEnvOrDefault("CHROMA_CLOUD_URL", "https://api.trychroma.com")

	return &config.VectorDBConfig{
		Name:             "chroma-cloud",
		Type:             config.VectorDBTypeChromaCloud,
		URL:              url,
		APIKey:           apiKey,
		Database:         database,
		Tenant:           tenant,
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          60,
	}
}
