// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getOpenSearchLocalConfig returns OpenSearch Local configuration
func getOpenSearchLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for OpenSearch Local type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeOpenSearchLocal {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if opensearchConfig := tryCreateOpenSearchLocalConfigFromEnv(); opensearchConfig != nil {
		return opensearchConfig, nil
	}

	return nil, fmt.Errorf("no OpenSearch Local configuration found")
}

// getOpenSearchCloudConfig returns OpenSearch Cloud configuration
func getOpenSearchCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for OpenSearch Cloud type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeOpenSearchCloud {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if opensearchConfig := tryCreateOpenSearchCloudConfigFromEnv(); opensearchConfig != nil {
		return opensearchConfig, nil
	}

	return nil, fmt.Errorf("no OpenSearch Cloud configuration found")
}

// tryCreateOpenSearchLocalConfigFromEnv attempts to create an OpenSearch Local config from environment variables
func tryCreateOpenSearchLocalConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("OPENSEARCH_LOCAL_ADDRESS")

	// Set default
	if url == "" {
		url = "http://localhost:9200"
	}

	return &config.VectorDBConfig{
		Name:             "opensearch-local",
		Type:             config.VectorDBTypeOpenSearchLocal,
		URL:              url,
		VectorDimensions: 1536,
		SimilarityMetric: "l2",
		Timeout:          30,
	}
}

// tryCreateOpenSearchCloudConfigFromEnv attempts to create an OpenSearch Cloud config from environment variables
func tryCreateOpenSearchCloudConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("OPENSEARCH_CLOUD_ADDRESS")
	username := os.Getenv("OPENSEARCH_CLOUD_USERNAME")
	password := os.Getenv("OPENSEARCH_CLOUD_PASSWORD")
	apiKey := os.Getenv("OPENSEARCH_CLOUD_API_KEY")

	// Require URL and either username/password or API key
	if url == "" {
		return nil
	}
	if (username == "" || password == "") && apiKey == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:             "opensearch-cloud",
		Type:             config.VectorDBTypeOpenSearchCloud,
		URL:              url,
		Username:         username,
		Password:         password,
		APIKey:           apiKey,
		VectorDimensions: 1536,
		SimilarityMetric: "cosinesimil",
		Timeout:          60,
	}
}
