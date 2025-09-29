package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// collectionDeleteCmd represents the collection delete command
var collectionDeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME",
	Aliases: []string{"del", "d"},
	Short:   "Delete a collection",
	Long: `Delete a collection from the vector database.

This command will permanently delete the collection and all its documents.
Use with caution as this operation cannot be undone.

Example:
  weave collection delete MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionDelete,
}

func init() {
	collectionCmd.AddCommand(collectionDeleteCmd)

	collectionDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDelete(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation prompt
	if !force {
		if !confirmAction(fmt.Sprintf("Are you sure you want to delete collection '%s'? This action cannot be undone.", collectionName)) {
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
		err = deleteWeaviateCollection(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		err = deleteMockCollection(ctx, dbConfig, collectionName)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		printError(fmt.Sprintf("Failed to delete collection: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
}
