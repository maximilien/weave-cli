// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ShowCmd represents the collection show command
var ShowCmd = &cobra.Command{
	Use:     "show COLLECTION_NAME",
	Aliases: []string{"s"},
	Short:   "Show collection details",
	Long: `Show detailed information about a collection.

This command displays:
- Collection schema
- Metadata information
- Document count
- Field definitions

Example:
  weave collection show MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionShow,
}

func init() {
	CollectionCmd.AddCommand(ShowCmd)

	ShowCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	ShowCmd.Flags().BoolP("no-truncate", "n", false, "Don't truncate long content")
	ShowCmd.Flags().BoolP("verbose", "", false, "Show verbose information")
	ShowCmd.Flags().BoolP("schema", "", false, "Show collection schema")
	ShowCmd.Flags().BoolP("metadata", "", false, "Show collection metadata")
	ShowCmd.Flags().BoolP("expand-metadata", "", false, "Show expanded metadata information")
	ShowCmd.Flags().BoolP("yaml", "", false, "Output schema/metadata as YAML")
	ShowCmd.Flags().BoolP("json", "", false, "Output schema/metadata as JSON")
	ShowCmd.Flags().StringP("yaml-file", "", "", "Write schema/metadata to YAML file")
	ShowCmd.Flags().StringP("json-file", "", "", "Write schema/metadata to JSON file")
	ShowCmd.Flags().BoolP("compact", "", false, "Remove occurrences, samples, and empty nested properties from output")
}

func runCollectionShow(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	shortLines, _ := cmd.Flags().GetInt("short")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	verbose, _ := cmd.Flags().GetBool("verbose")
	showSchema, _ := cmd.Flags().GetBool("schema")
	showMetadata, _ := cmd.Flags().GetBool("metadata")
	expandMetadata, _ := cmd.Flags().GetBool("expand-metadata")
	outputYAML, _ := cmd.Flags().GetBool("yaml")
	outputJSON, _ := cmd.Flags().GetBool("json")
	yamlFile, _ := cmd.Flags().GetString("yaml-file")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	compact, _ := cmd.Flags().GetBool("compact")

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.ShowWeaviateCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose, showSchema, showMetadata, expandMetadata, outputYAML, outputJSON, yamlFile, jsonFile, compact)
	case config.VectorDBTypeMock:
		utils.ShowMockCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose, showSchema, showMetadata, expandMetadata, outputYAML, outputJSON, yamlFile, jsonFile, compact)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
