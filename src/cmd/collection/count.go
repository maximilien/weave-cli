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

// CountCmd represents the collection count command
var CountCmd = &cobra.Command{
	Use:     "count",
	Aliases: []string{"c"},
	Short:   "Count collections",
	Long: `Count the number of collections in the vector database.

This command returns the total number of collections.

Example:
  weave collection count`,
	Args: cobra.NoArgs,
	Run:  runCollectionCount,
}

func init() {
	CollectionCmd.AddCommand(CountCmd)
}

func runCollectionCount(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		utils.HandleConfigError(err, true)
		os.Exit(1)
	}

	ctx := context.Background()

	var count int
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		count, err = utils.CountWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeSupabase:
		utils.PrintError("Collection count not yet implemented for Supabase")
		os.Exit(1)
	case config.VectorDBTypeMongoDB:
		utils.PrintError("Collection count not yet implemented for MongoDB")
		os.Exit(1)
	case config.VectorDBTypeMock:
		count, err = utils.CountMockCollections(ctx, dbConfig)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to count collections: %v", err))
		os.Exit(1)
	}

	utils.PrintHeader("Collection Count")
	fmt.Printf("Total collections: %d\n", count)
}
