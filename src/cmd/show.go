package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// collectionShowCmd represents the collection show command
var collectionShowCmd = &cobra.Command{
	Use:     "show COLLECTION_NAME",
	Aliases: []string{"s"},
	Short:   "Show collection details",
	Long: `Show detailed information about a collection.

This command displays:
- Collection schema
- Metadata information
- Document count
- Field definitions

Example:
  weave collection show MyCollection`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionShow,
}

func init() {
	collectionCmd.AddCommand(collectionShowCmd)

	collectionShowCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	collectionShowCmd.Flags().BoolP("no-truncate", "n", false, "Don't truncate long content")
	collectionShowCmd.Flags().BoolP("verbose", "v", false, "Show verbose information")
	collectionShowCmd.Flags().BoolP("schema", "", false, "Show collection schema")
	collectionShowCmd.Flags().BoolP("metadata", "", false, "Show collection metadata")
}

func runCollectionShow(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	shortLines, _ := cmd.Flags().GetInt("short")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	verbose, _ := cmd.Flags().GetBool("verbose")
	showSchema, _ := cmd.Flags().GetBool("schema")
	showMetadata, _ := cmd.Flags().GetBool("metadata")

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
		showWeaviateCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose, showSchema, showMetadata)
	case config.VectorDBTypeMock:
		showMockCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose, showSchema, showMetadata)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
