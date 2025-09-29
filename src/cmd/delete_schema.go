package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// collectionDeleteSchemaCmd represents the collection delete-schema command
var collectionDeleteSchemaCmd = &cobra.Command{
	Use:     "delete-schema COLLECTION_NAME",
	Aliases: []string{"ds"},
	Short:   "Delete collection schema",
	Long: `Delete the schema of a collection.

This command removes the schema definition from the collection.
The collection and its documents will remain, but the schema will be deleted.

Example:
  weave collection delete-schema MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionDeleteSchema,
}

func init() {
	collectionCmd.AddCommand(collectionDeleteSchemaCmd)

	collectionDeleteSchemaCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDeleteSchema(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation prompt
	if !force {
		if !confirmAction(fmt.Sprintf("Are you sure you want to delete the schema for collection '%s'?", collectionName)) {
			fmt.Println("Operation cancelled")
			return
		}
	}

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		err = deleteWeaviateCollectionSchema(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		// Mock collections don't have separate schemas
		printWarning("Mock collections don't have separate schemas to delete")
		return
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		printError(fmt.Sprintf("Failed to delete collection schema: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted schema for collection: %s", collectionName))
}
