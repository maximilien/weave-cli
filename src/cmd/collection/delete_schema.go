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

Example:
  weave collection delete-schema MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionDeleteSchema,
}

func init() {
	CollectionCmd.AddCommand(DeleteSchemaCmd)
	
	DeleteSchemaCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDeleteSchema(cmd *cobra.Command, args []string) {
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
		if !utils.ConfirmAction(fmt.Sprintf("Are you sure you want to delete the schema for collection '%s'?", collectionName)) {
			fmt.Println("Operation cancelled")
			return
		}
		
		// Second confirmation
		if !utils.ConfirmAction("This will permanently delete the schema. Type 'yes' to confirm: ") {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		err = utils.DeleteWeaviateCollectionSchema(ctx, dbConfig, collectionName)
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