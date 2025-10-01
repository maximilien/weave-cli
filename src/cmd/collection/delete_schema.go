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

// DeleteSchemaCmd represents the collection delete-schema command
var DeleteSchemaCmd = &cobra.Command{
	Use:     "delete-schema COLLECTION_NAME",
	Aliases: []string{"ds"},
	Short:   "Delete collection schema",
	Long: `Delete the schema of a collection.

This command removes the schema definition from the collection.
The collection and its documents will remain, but the schema will be deleted.

⚠️  WARNING: This is a destructive operation that will permanently
delete the schema definition. Use with caution!

This command requires double confirmation:
1. First confirmation: Standard y/N prompt
2. Second confirmation: Red warning requiring exact "yes" input

Use --force to skip confirmations in scripts.

You can specify collections in multiple ways:
1. By collection name: weave cols delete-schema MyCollection
2. By pattern: weave cols delete-schema --pattern "Test*"

Pattern types (auto-detected):
- Shell glob: Test*, WeaveDocs*, *Docs
- Regex: Test.*, ^WeaveDocs.*$, .*Docs$

Examples:
  weave collection delete-schema MyCollection
  weave collection delete-schema --pattern "Test*"`,
	Args: func(cmd *cobra.Command, args []string) error {
		pattern, _ := cmd.Flags().GetString("pattern")
		if pattern == "" && len(args) == 0 {
			return fmt.Errorf("requires collection name or --pattern flag")
		}
		return nil
	},
	Run: runCollectionDeleteSchema,
}

func init() {
	CollectionCmd.AddCommand(DeleteSchemaCmd)

	DeleteSchemaCmd.Flags().StringP("pattern", "p", "", "Delete schemas of collections matching pattern (auto-detects shell glob vs regex)")
	DeleteSchemaCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDeleteSchema(cmd *cobra.Command, args []string) {
	pattern, _ := cmd.Flags().GetString("pattern")
	force, _ := cmd.Flags().GetBool("force")

	var collectionName string
	if len(args) > 0 {
		collectionName = args[0]
	}

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
		var message string
		if pattern != "" {
			message = fmt.Sprintf("⚠️  Are you sure you want to delete schemas for collections matching pattern '%s'?", pattern)
		} else {
			message = fmt.Sprintf("⚠️  Are you sure you want to delete the schema for collection '%s'?", collectionName)
		}
		utils.PrintWarning(message)
		if !utils.ConfirmAction("") {
			fmt.Println("Operation cancelled")
			return
		}

		// Second confirmation with prominent red warning
		utils.PrintError("🚨 This will permanently delete the schema(s). Type 'yes' to confirm:")
		if !utils.ConfirmAction("") {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		if pattern != "" {
			err = utils.DeleteWeaviateCollectionSchemasByPattern(ctx, dbConfig, pattern)
		} else {
			err = utils.DeleteWeaviateCollectionSchema(ctx, dbConfig, collectionName)
		}
	case config.VectorDBTypeMock:
		// Mock collections don't have separate schemas
		utils.PrintWarning("Mock collections don't have separate schemas to delete")
		return
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to delete collection schema: %v", err))
		os.Exit(1)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully deleted schema for collection: %s", collectionName))
}
