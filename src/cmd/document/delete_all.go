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
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get selected databases
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	// Smart database selection for single-database operations
	dbConfig := utils.HandleSingleDatabaseSelection(ctx, selection, cfg, collectionName, fmt.Sprintf("weave docs delete-all %s", collectionName))

	// Confirmation prompt
	if !force {
		utils.PrintWarning(fmt.Sprintf("⚠️  Are you sure you want to delete ALL documents from collection '%s'? This action cannot be undone.", collectionName))
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
		utils.DeleteAllWeaviateDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeSupabase, config.VectorDBTypeMongoDB:
		utils.DeleteAllDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
		utils.DeleteAllDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
		utils.DeleteAllDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		utils.DeleteAllMockDocuments(ctx, dbConfig, collectionName)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
