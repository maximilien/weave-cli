// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getNeo4jLocalConfig returns Neo4j Local configuration
func getNeo4jLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Neo4j Local type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeNeo4jLocal {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if neo4jConfig := tryCreateNeo4jLocalConfigFromEnv(); neo4jConfig != nil {
		return neo4jConfig, nil
	}

	return nil, fmt.Errorf("no Neo4j Local configuration found")
}

// getNeo4jCloudConfig returns Neo4j Cloud configuration
func getNeo4jCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	// Check all configured databases for Neo4j Cloud type
	for _, dbConfig := range cfg.Databases.VectorDatabases {
		if dbConfig.Type == config.VectorDBTypeNeo4jCloud {
			return &dbConfig, nil
		}
	}

	// Try to create from environment variables
	if neo4jConfig := tryCreateNeo4jCloudConfigFromEnv(); neo4jConfig != nil {
		return neo4jConfig, nil
	}

	return nil, fmt.Errorf("no Neo4j Cloud configuration found")
}

// tryCreateNeo4jLocalConfigFromEnv attempts to create a Neo4j Local config from environment variables
func tryCreateNeo4jLocalConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("NEO4J_URL")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	// Require password at minimum
	if password == "" {
		return nil
	}

	// Set defaults
	if url == "" {
		url = "bolt://localhost:7687"
	}
	if username == "" {
		username = "neo4j"
	}

	return &config.VectorDBConfig{
		Name:             "neo4j-local",
		Type:             config.VectorDBTypeNeo4jLocal,
		URL:              url,
		Username:         username,
		Password:         password,
		Database:         getEnvOrDefault("NEO4J_DATABASE", "neo4j"),
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          30,
	}
}

// tryCreateNeo4jCloudConfigFromEnv attempts to create a Neo4j Cloud config from environment variables
func tryCreateNeo4jCloudConfigFromEnv() *config.VectorDBConfig {
	url := os.Getenv("NEO4J_CLOUD_URL")
	username := os.Getenv("NEO4J_CLOUD_USERNAME")
	password := os.Getenv("NEO4J_CLOUD_PASSWORD")

	// Require all for cloud
	if url == "" || username == "" || password == "" {
		return nil
	}

	return &config.VectorDBConfig{
		Name:             "neo4j-cloud",
		Type:             config.VectorDBTypeNeo4jCloud,
		URL:              url,
		Username:         username,
		Password:         password,
		Database:         getEnvOrDefault("NEO4J_CLOUD_DATABASE", "neo4j"),
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          60,
	}
}
