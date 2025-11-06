// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
)

// Display helper functions shared across config commands

func displayDatabaseConfig(name string, dbConfig *config.VectorDBConfig) {
	color.New(color.FgCyan, color.Bold).Printf("Type: %s\n", dbConfig.Type)
	fmt.Println()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		displayWeaviateCloudConfig(dbConfig)
	case config.VectorDBTypeLocal:
		displayWeaviateLocalConfig(dbConfig)
	case config.VectorDBTypeMock:
		displayMockConfig(dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func displayWeaviateCloudConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgGreen).Printf("🌐 Weaviate Cloud Configuration\n")
	fmt.Printf("  URL: %s\n", cfg.URL)

	// Mask API key for security
	apiKeyDisplay := "***hidden***"
	if cfg.APIKey == "" {
		apiKeyDisplay = "❌ not set"
		color.New(color.FgRed).Printf("  API Key: %s\n", apiKeyDisplay)
	} else {
		color.New(color.FgGreen).Printf("  API Key: %s\n", apiKeyDisplay)
	}

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s)\n", collection.Name, collection.Type)
		}
	}
}

func displayWeaviateLocalConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgBlue).Printf("🏠 Weaviate Local Configuration\n")
	fmt.Printf("  URL: %s\n", cfg.URL)

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s)\n", collection.Name, collection.Type)
		}
	}
}

func displayMockConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgYellow).Printf("🎭 Mock Database Configuration\n")
	fmt.Printf("  Enabled: %t\n", cfg.Enabled)
	fmt.Printf("  Simulate Embeddings: %t\n", cfg.SimulateEmbeddings)
	fmt.Printf("  Embedding Dimension: %d\n", cfg.EmbeddingDimension)

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s): %s\n", collection.Name, collection.Type, collection.Description)
		}
	}
}

func displayJSONSchemaFields(jsonSchema map[string]interface{}, indent string) {
	for field, fieldType := range jsonSchema {
		fmt.Printf("%s%s: %v\n", indent, field, fieldType)
	}
}

// Output formatting helper functions
func printError(msg string) {
	color.New(color.FgRed, color.Bold).Printf("❌ %s\n", msg)
}

func printWarning(msg string) {
	color.New(color.FgYellow, color.Bold).Printf("⚠️  %s\n", msg)
}

func printHeader(msg string) {
	color.New(color.FgCyan, color.Bold).Printf("═══ %s ═══\n", msg)
}
