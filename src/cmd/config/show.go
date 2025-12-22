// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	configpkg "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// showCmd represents the config show command
var showCmd = &cobra.Command{
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
	Run: runShow,
}

func runShow(cobraCmd *cobra.Command, args []string) {
	outputFormat, _ := cobraCmd.Flags().GetString("output")

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// If JSON or YAML output requested, output the whole config and exit
	if outputFormat == "json" || outputFormat == "yaml" {
		var data []byte
		var marshalErr error

		if outputFormat == "json" {
			data, marshalErr = json.MarshalIndent(cfg, "", "  ")
		} else {
			data, marshalErr = yaml.Marshal(cfg)
		}

		if marshalErr != nil {
			printError(fmt.Sprintf("Failed to marshal configuration: %v", marshalErr))
			os.Exit(1)
		}

		fmt.Println(string(data))
		return
	}

	// Text output (original behavior)
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

	// Find config paths to show location
	configPaths, err := configpkg.FindConfigPaths()
	if err == nil {
		// Show location information
		if configPaths.Location == "global" {
			globalDir, _ := configpkg.GetGlobalConfigDir()
			color.New(color.FgCyan).Printf("📍 Location: Global (%s)\n", globalDir)
		} else {
			color.New(color.FgCyan).Printf("📍 Location: Local (current directory)\n")
		}
		fmt.Println()
	}

	fmt.Printf("Config file: %s\n", configpkg.GetConfigFile())
	fmt.Printf("Env file: %s\n", configpkg.GetEnvFile())

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
