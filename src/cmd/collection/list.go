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

// ListCmd represents the collection list command
var ListCmd = &cobra.Command{
	Use:     "list [database-name]",
	Aliases: []string{"ls", "l"},
	Short:   "List all collections",
	Long: `List all collections in the configured vector database.

This command shows:
- Collection names
- Document counts (if available)
- Collection metadata (if available)

If no database name is provided, it uses the default database.
Use 'weave config list' to see all available databases.`,
	Run: runCollectionList,
}

func init() {
	CollectionCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntP("limit", "l", 100, "Maximum number of collections to show")
	ListCmd.Flags().BoolP("virtual", "", false, "Show collections in virtual structure")
}

func runCollectionList(cmd *cobra.Command, args []string) {
	limit, _ := cmd.Flags().GetInt("limit")
	virtual, _ := cmd.Flags().GetBool("virtual")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load configuration with interactive help
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	ctx := context.Background()

	// Check if a specific database name is provided (legacy behavior)
	if len(args) > 0 {
		// Use specified database (legacy behavior)
		dbConfig, err := cfg.GetDatabase(args[0])
		if err != nil {
			utils.HandleConfigError(err, true)
			os.Exit(1)
		}
		listCollectionsForDatabase(ctx, dbConfig, limit, virtual, jsonOutput)
		return
	}

	// Use new vector database selector based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector databases: %v", err))
		os.Exit(1)
	}

	// List collections for each selected database
	for i, dbConfig := range selection.Configs {
		// If multiple databases, show which database we're listing
		if len(selection.Configs) > 1 {
			dbType := utils.GetVectorDBTypeFromConfig(&dbConfig)
			if !jsonOutput {
				utils.PrintInfo(fmt.Sprintf("\n=== %s ===", dbType))
			}
		}

		listCollectionsForDatabase(ctx, &dbConfig, limit, virtual, jsonOutput)

		// Add spacing between databases (except for the last one)
		if i < len(selection.Configs)-1 && !jsonOutput {
			fmt.Println()
		}
	}
}

// listCollectionsForDatabase lists collections for a specific database configuration
func listCollectionsForDatabase(ctx context.Context, dbConfig *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool) {
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.ListWeaviateCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMock:
		utils.ListMockCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeSupabase:
		utils.ListSupabaseCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMongoDB:
		utils.ListMongoDBCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
		utils.ListMilvusCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
		utils.ListChromaCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeQdrantLocal, config.VectorDBTypeQdrantCloud:
		utils.ListQdrantCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
	}
}
