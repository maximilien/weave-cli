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
	address := os.Getenv("OPENSEARCH_LOCAL_ADDRESS")

	// Require address at minimum
	if address == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:             "opensearch-local",
		Type:             config.VectorDBTypeOpenSearchLocal,
		Address:          address,
		URL:              address,
		VectorDimensions: 384,
		SimilarityMetric: "l2",
		Timeout:          30,
	}
}

// tryCreateOpenSearchCloudConfigFromEnv attempts to create an OpenSearch Cloud config from environment variables
func tryCreateOpenSearchCloudConfigFromEnv() *config.VectorDBConfig {
	address := os.Getenv("OPENSEARCH_CLOUD_ADDRESS")
	username := os.Getenv("OPENSEARCH_CLOUD_USERNAME")
	password := os.Getenv("OPENSEARCH_CLOUD_PASSWORD")
	apiKey := os.Getenv("OPENSEARCH_CLOUD_API_KEY")

	// Require address and either credentials or API key
	if address == "" || (username == "" && password == "" && apiKey == "") {
		return nil
	}

	cfg := &config.VectorDBConfig{
		Name:             "opensearch-cloud",
		Type:             config.VectorDBTypeOpenSearchCloud,
		Address:          address,
		URL:              address,
		VectorDimensions: 384,
		SimilarityMetric: "l2",
		Timeout:          60,
	}

	// Prefer API key if available
	if apiKey != "" {
		cfg.APIKey = apiKey
	} else {
		cfg.Username = username
		cfg.Password = password
	}

	return cfg
}
