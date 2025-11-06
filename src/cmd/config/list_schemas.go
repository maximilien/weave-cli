// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/spf13/cobra"
)

// listSchemasCmd represents the config list-schemas command
var listSchemasCmd = &cobra.Command{
	Use:     "list-schemas",
	Aliases: []string{"ls-schemas"},
	Short:   "List all configured schemas",
	Long: `List all configured schemas defined in config.yaml.

This command displays:
- All configured schema names
- Schema class names
- Schema vectorizer types`,
	Run: runListSchemas,
}

func runListSchemas(cobraCmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
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
