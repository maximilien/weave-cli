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

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get database config
	var dbConfig *config.VectorDBConfig
	if len(args) > 0 {
		// Use specified database
		dbConfig, err = cfg.GetDatabase(args[0])
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to get database '%s': %v", args[0], err))
			os.Exit(1)
		}
	} else {
		// Use default database
		dbConfig, err = cfg.GetDefaultDatabase()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.ListWeaviateCollections(ctx, dbConfig, limit, virtual)
	case config.VectorDBTypeMock:
		utils.ListMockCollections(ctx, dbConfig, limit, virtual)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
