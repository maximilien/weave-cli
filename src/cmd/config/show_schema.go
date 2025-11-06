// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

// showSchemaCmd represents the config show-schema command
var showSchemaCmd = &cobra.Command{
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
	Run:  runShowSchema,
}

func runShowSchema(cobraCmd *cobra.Command, args []string) {
	schemaName := args[0]

	// Get flags
	yamlOutput, _ := cobraCmd.Flags().GetBool("yaml")
	jsonOutput, _ := cobraCmd.Flags().GetBool("json")

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
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

	// Handle output formats
	if yamlOutput {
		outputSchemaAsYAML(schema)
		return
	}

	if jsonOutput {
		outputSchemaAsJSON(schema)
		return
	}

	// Default: formatted output
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

func outputSchemaAsYAML(schema *config.SchemaDefinition) {
	// Use the same export function from utils package
	data, err := yaml.Marshal(schema)
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal schema to YAML: %v", err))
		os.Exit(1)
	}

	// Fix json_schema indentation for YAML linting
	result := fixSchemaJSONIndentation(string(data))

	// Add YAML document separator and print
	fmt.Println("---")
	fmt.Print(result)
}

func outputSchemaAsJSON(schema *config.SchemaDefinition) {
	// Use json.MarshalIndent for pretty printing
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal schema to JSON: %v", err))
		os.Exit(1)
	}

	fmt.Println(string(data))
}

// fixSchemaJSONIndentation fixes json_schema indentation in YAML output
// Similar to the function in utils/export.go but for schema output
func fixSchemaJSONIndentation(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	result := make([]string, 0, len(lines))
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check if this is a json_schema: line
		if trimmed == "json_schema:" {
			// Get the indentation of the json_schema: line
			jsonSchemaIndent := len(line) - len(strings.TrimLeft(line, " "))
			// YAML linting requires +4 spaces for nested maps in lists (not +2)
			expectedChildIndent := jsonSchemaIndent + 4
			result = append(result, line)
			i++

			// Collect all children of json_schema
			firstChildIndent := -1
			for i < len(lines) {
				childLine := lines[i]
				childTrimmed := strings.TrimSpace(childLine)

				if childTrimmed == "" {
					result = append(result, childLine)
					i++
					continue
				}

				childIndent := len(childLine) - len(strings.TrimLeft(childLine, " "))

				// If child has same or less indentation than json_schema:, we're done
				if childIndent <= jsonSchemaIndent {
					break
				}

				// Track first child's indentation
				if firstChildIndent == -1 {
					firstChildIndent = childIndent
				}

				// Calculate the indentation delta
				indentDelta := childIndent - firstChildIndent
				correctedIndent := expectedChildIndent + indentDelta
				correctedLine := strings.Repeat(" ", correctedIndent) + childTrimmed
				result = append(result, correctedLine)
				i++
			}
		} else {
			result = append(result, line)
			i++
		}
	}

	return strings.Join(result, "\n")
}
