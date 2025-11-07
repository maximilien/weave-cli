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

// DeleteAllCmd represents the collection delete-all command
var DeleteAllCmd = &cobra.Command{
	Use:     "delete-all",
	Aliases: []string{"da"},
	Short:   "Clear all collections (delete all documents)",
	Long: `Clear all collections by deleting all documents from them.

⚠️  WARNING: This is a destructive operation that will permanently
delete ALL documents in ALL collections. The collection schemas will
remain but will be empty. Use with extreme caution!

This command requires double confirmation:
1. First confirmation: Standard y/N prompt
2. Second confirmation: Red warning requiring exact "yes" input

Use --force to skip confirmations in scripts.

Example:
  weave collection delete-all`,
	Args: cobra.NoArgs,
	Run:  runCollectionDeleteAll,
}

func init() {
	CollectionCmd.AddCommand(DeleteAllCmd)

	DeleteAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDeleteAll(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")

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

	// Confirmation prompt
	if !force {
		utils.PrintWarning("⚠️  Are you sure you want to delete ALL documents from ALL collections? This action cannot be undone.")
		if !utils.ConfirmAction("") {
			fmt.Println("Operation cancelled")
			return
		}

		// Second confirmation with prominent red warning
		utils.PrintError("🚨 This will permanently delete ALL documents. Type 'yes' to confirm:")
		if !utils.ConfirmAction("") {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.DeleteAllWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeMock:
		utils.DeleteAllMockCollections(ctx, dbConfig)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	utils.PrintSuccess("Successfully deleted all documents from all collections")
}
