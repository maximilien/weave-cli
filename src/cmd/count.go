package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// collectionCountCmd represents the collection count command
var collectionCountCmd = &cobra.Command{
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
	collectionCmd.AddCommand(collectionCountCmd)
}

func runCollectionCount(cmd *cobra.Command, args []string) {
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

	var count int
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		count, err = countWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeMock:
		count, err = countMockCollections(ctx, dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		printError(fmt.Sprintf("Failed to count collections: %v", err))
		os.Exit(1)
	}

	printHeader("Collection Count")
	fmt.Printf("Total collections: %d\n", count)
}
