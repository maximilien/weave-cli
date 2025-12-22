// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ListCmd represents the document list command
var ListCmd = &cobra.Command{
	Use:     "list COLLECTION_NAME",
	Aliases: []string{"ls", "l"},
	Short:   "List documents in a collection",
	Long: `List documents in a specific collection.

This command shows:
- Document IDs
- Content previews (truncated)
- Metadata information
- Document counts`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentList,
}

func init() {
	DocumentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntP("limit", "l", 50, "Maximum number of documents to show")
	ListCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	ListCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	ListCmd.Flags().BoolP("virtual", "w", false, "Show documents in virtual structure (aggregate chunks by original document)")
	ListCmd.Flags().BoolP("summary", "S", false, "Show a clean summary of documents (works with --virtual)")
	ListCmd.Flags().StringP("output", "o", "text", "Output format: text, json, yaml")
}

func runDocumentList(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")
	virtual, _ := cmd.Flags().GetBool("virtual")
	summary, _ := cmd.Flags().GetBool("summary")
	outputFormat, _ := cmd.Flags().GetString("output")

	// Support legacy --json flag for backward compatibility
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		outputFormat = "json"
	}
	jsonOutput := (outputFormat == "json")

	// Load configuration with interactive help
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get selected databases based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	// For single-database operations, handle multiple databases intelligently
	var dbConfig *config.VectorDBConfig
	if len(selection.Configs) > 1 {
		// Try to use the default database type from config/env
		defaultDB := utils.SelectDefaultDatabase(selection.Configs, cfg)
		if defaultDB != nil {
			dbConfig = defaultDB
			utils.PrintInfo(fmt.Sprintf("Using default database: %s (%s)", dbConfig.Name, dbConfig.Type))
			fmt.Println()
		} else {
			// If multiple Weaviate databases, try each until one works
			if utils.AreAllWeaviateDatabases(selection.Configs) {
				utils.PrintInfo(fmt.Sprintf("Trying %d Weaviate databases...", len(selection.Configs)))
				fmt.Println()
				dbConfig = utils.TryWeaviateDatabasesForCollection(context.Background(), selection.Configs, collectionName)
				if dbConfig == nil {
					utils.PrintError(fmt.Sprintf("Collection '%s' not found in any Weaviate database", collectionName))
					os.Exit(1)
				}
			} else {
				// Multiple different database types - need user to specify
				utils.PrintError(fmt.Sprintf("Document list requires a single database, but found %d databases:", len(selection.Configs)))
				for _, cfg := range selection.Configs {
					utils.PrintInfo(fmt.Sprintf("  • %s (%s)", cfg.Name, cfg.Type))
				}
				fmt.Println()
				utils.PrintInfo("Please use --vector-db-type (or --vdb) to specify which database:")
				for _, cfg := range selection.Configs {
					utils.PrintInfo(fmt.Sprintf("  weave docs ls %s --vector-db-type %s", collectionName, cfg.Type))
				}
				os.Exit(1)
			}
		}
	} else {
		dbConfig = &selection.Configs[0]
	}

	ctx := context.Background()

	// Use generic ListDocuments that works with all database types via vectordb abstraction
	utils.ListDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary, jsonOutput)
}
