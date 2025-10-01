// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Manage Weave CLI configuration.

This command provides subcommands to view and manage configuration settings.`,
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show [database-name]",
	Short: "Show currently configured databases",
	Long: `Show the currently configured vector database settings.

This command displays:
- All configured databases or a specific database
- Vector database type (weaviate-cloud, weaviate-local, mock)
- Connection details (URL, API key status)
- Collection names
- Configuration source files

If no database name is provided, it shows the default database.
Use 'weave config list' to see all available databases.`,
	Run: runConfigShow,
}

// configListCmd represents the config list command
var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all configured databases",
	Long: `List all configured vector databases.

This command displays:
- All configured database names
- Database types
- Which database is the default`,
	Run: runConfigList,
}

// configListSchemasCmd represents the config list-schemas command
var configListSchemasCmd = &cobra.Command{
	Use:     "list-schemas",
	Aliases: []string{"ls-schemas"},
	Short:   "List all configured schemas",
	Long: `List all configured schemas defined in config.yaml.

This command displays:
- All configured schema names
- Schema class names
- Schema vectorizer types`,
	Run: runConfigListSchemas,
}

// configShowSchemaCmd represents the config show-schema command
var configShowSchemaCmd = &cobra.Command{
	Use:   "show-schema SCHEMA_NAME",
	Short: "Show a specific schema configuration",
	Long: `Show detailed information about a specific schema.

This command displays:
- Schema name and class
- Vectorizer configuration
- All properties and their types
- JSON schema structures if present
- Metadata field definitions`,
	Args: cobra.ExactArgs(1),
	Run:  runConfigShowSchema,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configListSchemasCmd)
	configCmd.AddCommand(configShowSchemaCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// If a specific database name is provided, show only that database
	if len(args) > 0 {
		dbName := args[0]
		dbConfig, err := cfg.GetDatabase(dbName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", dbName, err))
			os.Exit(1)
		}

		printHeader(fmt.Sprintf("Database Configuration: %s", dbName))
		fmt.Println()
		displayDatabaseConfig(dbName, dbConfig)
	} else {
		// Show default database
		dbConfig, err := cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}

		printHeader("Default Database Configuration")
		fmt.Println()
		displayDatabaseConfig("default", dbConfig)
	}

	// Display configuration sources
	fmt.Println()
	printHeader("Configuration Sources")
	fmt.Printf("Config file: %s\n", config.GetConfigFile())
	fmt.Printf("Env file: %s\n", config.GetEnvFile())

	// Display schema configuration
	fmt.Println()
	printHeader("Schema Configuration")
	if cfg.SchemasDir != "" {
		fmt.Printf("Schemas directory: %s\n", cfg.SchemasDir)
	} else {
		fmt.Printf("Schemas directory: not configured\n")
	}

	schemaCount := len(cfg.GetAllSchemas())
	if schemaCount > 0 {
		fmt.Printf("Configured schemas: %d\n", schemaCount)
		fmt.Println()
		schemaNames := cfg.ListSchemas()
		for i, name := range schemaNames {
			fmt.Printf("  %d. %s\n", i+1, name)
		}
		fmt.Println()
		color.New(color.FgCyan).Printf("💡 Use 'weave config show-schema <name>' to view schema details\n")
	} else {
		fmt.Printf("Configured schemas: 0\n")
	}
}

func runConfigList(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Configured Databases")
	fmt.Println()

	databaseNames := cfg.GetDatabaseNames()
	if len(databaseNames) == 0 {
		printWarning("No databases configured")
		return
	}

	for name, dbType := range databaseNames {
		isDefault := name == "default"
		if isDefault {
			color.New(color.FgGreen, color.Bold).Printf("• %s (default)\n", name)
		} else {
			fmt.Printf("• %s\n", name)
		}
		fmt.Printf("  Type: %s\n", dbType)
		fmt.Println()
	}
}

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

func runConfigListSchemas(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Configured Schemas")
	fmt.Println()

	schemas := cfg.GetAllSchemas()
	if len(schemas) == 0 {
		printWarning("No schemas configured in config.yaml")
		fmt.Println()
		fmt.Println("Add schemas to the 'databases.schemas' section in config.yaml")
		return
	}

	for i, schemaDef := range schemas {
		// Extract schema class and vectorizer
		schemaClass := "unknown"
		vectorizer := "unknown"

		// Schema can be either directly in Schema map or under Schema["schema"]
		var schemaMap map[string]interface{}
		if innerSchema, ok := schemaDef.Schema["schema"].(map[string]interface{}); ok {
			schemaMap = innerSchema
		} else {
			schemaMap = schemaDef.Schema
		}

		if class, ok := schemaMap["class"].(string); ok {
			schemaClass = class
		}
		if vec, ok := schemaMap["vectorizer"].(string); ok {
			vectorizer = vec
		}

		color.New(color.FgCyan, color.Bold).Printf("%d. %s\n", i+1, schemaDef.Name)
		fmt.Printf("   Class: %s\n", schemaClass)
		fmt.Printf("   Vectorizer: %s\n", vectorizer)
		fmt.Println()
	}

	color.New(color.FgGreen).Printf("✅ Found %d schema(s)\n", len(schemas))
}

func runConfigShowSchema(cmd *cobra.Command, args []string) {
	schemaName := args[0]

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get the schema
	schema, err := cfg.GetSchema(schemaName)
	if err != nil {
		printError(fmt.Sprintf("Failed to get schema '%s': %v", schemaName, err))
		os.Exit(1)
	}

	printHeader(fmt.Sprintf("Schema: %s", schemaName))
	fmt.Println()

	// Extract schema details
	// Schema can be either directly in Schema map or under Schema["schema"]
	var schemaMap map[string]interface{}
	if innerSchema, ok := schema.Schema["schema"].(map[string]interface{}); ok {
		schemaMap = innerSchema
	} else {
		schemaMap = schema.Schema
	}

	// Display class and vectorizer
	if class, ok := schemaMap["class"].(string); ok {
		color.New(color.FgCyan).Printf("📋 Class: ")
		fmt.Printf("%s\n", class)
	}

	if vectorizer, ok := schemaMap["vectorizer"].(string); ok {
		color.New(color.FgCyan).Printf("🔧 Vectorizer: ")
		fmt.Printf("%s\n", vectorizer)
	}

	fmt.Println()

	// Display properties
	if properties, ok := schemaMap["properties"].([]interface{}); ok && len(properties) > 0 {
		color.New(color.FgGreen, color.Bold).Printf("🏗️  Properties:\n")
		fmt.Println()

		for i, prop := range properties {
			if propMap, ok := prop.(map[string]interface{}); ok {
				name := propMap["name"]
				datatype := propMap["datatype"]
				description := propMap["description"]

				fmt.Printf("  %d. ", i+1)
				color.New(color.FgYellow).Printf("%v", name)
				fmt.Printf("\n")

				if datatype != nil {
					fmt.Printf("     Type: %v\n", datatype)
				}

				if description != nil && description != "" {
					fmt.Printf("     Description: %v\n", description)
				}

				// Display JSON schema if present
				if jsonSchema, ok := propMap["json_schema"].(map[string]interface{}); ok && len(jsonSchema) > 0 {
					fmt.Printf("     JSON Schema:\n")
					displayJSONSchemaFields(jsonSchema, "       ")
				}

				fmt.Println()
			}
		}
	}

	// Display metadata if present
	if metadata, ok := schema.Schema["metadata"].(map[string]interface{}); ok && len(metadata) > 0 {
		color.New(color.FgBlue, color.Bold).Printf("📊 Metadata Fields:\n")
		fmt.Println()

		for key, value := range metadata {
			fmt.Printf("  • ")
			color.New(color.FgYellow).Printf("%s", key)

			if valueMap, ok := value.(map[string]interface{}); ok {
				if fieldType, ok := valueMap["type"].(string); ok {
					fmt.Printf(": %s", fieldType)
				}

				// Display JSON schema if present
				if jsonSchema, ok := valueMap["json_schema"].(map[string]interface{}); ok && len(jsonSchema) > 0 {
					fmt.Printf("\n    JSON Schema:\n")
					displayJSONSchemaFields(jsonSchema, "      ")
				}
			} else {
				fmt.Printf(": %v", value)
			}

			fmt.Println()
		}
	}
}

func displayJSONSchemaFields(jsonSchema map[string]interface{}, indent string) {
	for field, fieldType := range jsonSchema {
		fmt.Printf("%s%s: %v\n", indent, field, fieldType)
	}
}
