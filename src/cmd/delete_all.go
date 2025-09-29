package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// collectionDeleteAllCmd represents the collection delete-all command
var collectionDeleteAllCmd = &cobra.Command{
	Use:     "delete-all",
	Aliases: []string{"da"},
	Short:   "Delete all collections",
	Long: `Delete all collections from the vector database.

This command will permanently delete ALL collections and their documents.
Use with extreme caution as this operation cannot be undone.

Example:
  weave collection delete-all`,
	Args: cobra.NoArgs,
	Run:  runCollectionDeleteAll,
}

func init() {
	collectionCmd.AddCommand(collectionDeleteAllCmd)

	collectionDeleteAllCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDeleteAll(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation prompt
	if !force {
		if !confirmAction("Are you sure you want to delete ALL collections? This action cannot be undone.") {
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
		deleteAllWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeMock:
		deleteAllMockCollections(ctx, dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	printSuccess("Successfully deleted all collections")
}
