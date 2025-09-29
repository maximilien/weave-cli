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

// DeleteAllCmd represents the document delete-all command
var DeleteAllCmd = &cobra.Command{
	Use:     "delete-all COLLECTION_NAME",
	Aliases: []string{"del-all", "da"},
	Short:   "Delete all documents in a collection",
	Long: `Delete all documents from a specific collection.

⚠️  WARNING: This is a destructive operation that will permanently
delete ALL documents in the specified collection. Use with caution!

This command requires double confirmation:
1. First confirmation: Standard y/N prompt
2. Second confirmation: Red warning requiring exact "yes" input

Use --force to skip confirmations in scripts.

Example:
  weave document delete-all MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentDeleteAll,
}

func init() {
	DocumentCmd.AddCommand(DeleteAllCmd)

	DeleteAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runDocumentDeleteAll(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	force, _ := cmd.Flags().GetBool("force")

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

	// Confirmation prompt
	if !force {
		if !utils.ConfirmAction(fmt.Sprintf("Are you sure you want to delete ALL documents from collection '%s'? This action cannot be undone.", collectionName)) {
			fmt.Println("Operation cancelled")
			return
		}

		// Second confirmation
		if !utils.ConfirmAction("This will permanently delete ALL documents. Type 'yes' to confirm: ") {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.DeleteAllWeaviateDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		utils.DeleteAllMockDocuments(ctx, dbConfig, collectionName)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully deleted all documents from collection: %s", collectionName))
}
