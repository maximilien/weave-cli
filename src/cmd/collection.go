package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/spf13/cobra"
)

// collectionCmd represents the collection command
var collectionCmd = &cobra.Command{
	Use:     "collection",
	Aliases: []string{"col", "cols"},
	Short:   "Collection management",
	Long: `Manage vector database collections.

This command provides subcommands to list, view, and delete collections.`,
}

// collectionListCmd represents the collection list command
var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	Long: `List all collections in the configured vector database.

This command shows:
- Collection names
- Document counts (if available)
- Collection metadata (if available)`,
	Run: runCollectionList,
}

// collectionDeleteCmd represents the collection delete command
var collectionDeleteCmd = &cobra.Command{
	Use:   "delete COLLECTION_NAME",
	Short: "Delete a specific collection",
	Long: `Delete a specific collection from the configured vector database.

⚠️  WARNING: This is a destructive operation that will permanently
delete all data in the specified collection. Use with caution!`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionDelete,
}

// collectionDeleteAllCmd represents the collection delete-all command
var collectionDeleteAllCmd = &cobra.Command{
	Use:   "delete-all",
	Short: "Delete all collections",
	Long: `Delete all collections from the configured vector database.

⚠️  WARNING: This is a destructive operation that will permanently
delete all data in all collections. Use with caution!`,
	Run: runCollectionDeleteAll,
}

func init() {
	rootCmd.AddCommand(collectionCmd)
	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionDeleteCmd)
	collectionCmd.AddCommand(collectionDeleteAllCmd)
}

func runCollectionList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Vector Database Collections")
	fmt.Println()

	dbType := cfg.Database.VectorDB.Type
	color.New(color.FgCyan, color.Bold).Printf("Listing collections in %s database...\n", dbType)
	fmt.Println()

	ctx := context.Background()

	switch dbType {
	case config.VectorDBTypeCloud:
		listWeaviateCollections(ctx, &cfg.Database.VectorDB.WeaviateCloud)
	case config.VectorDBTypeLocal:
		listWeaviateCollections(ctx, &cfg.Database.VectorDB.WeaviateLocal)
	case config.VectorDBTypeMock:
		listMockCollections(ctx, &cfg.Database.VectorDB.Mock)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbType))
		os.Exit(1)
	}
}

func runCollectionDelete(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete Collection")
	fmt.Println()

	printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete collection '%s' and all its data!", collectionName))
	fmt.Println()

	// Confirm deletion
	if !confirmAction(fmt.Sprintf("Are you sure you want to delete collection '%s'?", collectionName)) {
		printInfo("Operation cancelled by user")
		return
	}

	dbType := cfg.Database.VectorDB.Type
	color.New(color.FgCyan, color.Bold).Printf("Deleting collection '%s' in %s database...\n", collectionName, dbType)
	fmt.Println()

	ctx := context.Background()

	switch dbType {
	case config.VectorDBTypeCloud:
		deleteWeaviateCollection(ctx, &cfg.Database.VectorDB.WeaviateCloud, collectionName)
	case config.VectorDBTypeLocal:
		deleteWeaviateCollection(ctx, &cfg.Database.VectorDB.WeaviateLocal, collectionName)
	case config.VectorDBTypeMock:
		deleteMockCollection(ctx, &cfg.Database.VectorDB.Mock, collectionName)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbType))
		os.Exit(1)
	}
}

func runCollectionDeleteAll(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete All Collections")
	fmt.Println()

	printWarning("⚠️  WARNING: This will permanently delete ALL collections and their data!")
	fmt.Println()

	// Confirm deletion
	if !confirmAction("Are you sure you want to delete all collections?") {
		printInfo("Operation cancelled by user")
		return
	}

	dbType := cfg.Database.VectorDB.Type
	color.New(color.FgCyan, color.Bold).Printf("Deleting all collections in %s database...\n", dbType)
	fmt.Println()

	ctx := context.Background()

	switch dbType {
	case config.VectorDBTypeCloud:
		deleteAllWeaviateCollections(ctx, &cfg.Database.VectorDB.WeaviateCloud)
	case config.VectorDBTypeLocal:
		deleteAllWeaviateCollections(ctx, &cfg.Database.VectorDB.WeaviateLocal)
	case config.VectorDBTypeMock:
		deleteAllMockCollections(ctx, &cfg.Database.VectorDB.Mock)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbType))
		os.Exit(1)
	}
}

func listWeaviateCollections(ctx context.Context, cfg interface{}) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printWarning("No collections found in the database")
		return
	}

	// Sort collections alphabetically
	sort.Strings(collections)

	printSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
	fmt.Println()

	for i, collection := range collections {
		color.New(color.FgGreen).Printf("%d. %s\n", i+1, collection)

		// Try to get document count (use a reasonable limit for counting)
		documents, err := client.ListDocuments(ctx, collection, 1000) // Use 1000 as a reasonable limit for counting
		if err == nil {
			if len(documents) >= 1000 {
				fmt.Printf("   Documents: %d+ (showing first 1000)\n", len(documents))
			} else {
				fmt.Printf("   Documents: %d\n", len(documents))
			}
		} else {
			fmt.Printf("   Documents: Unable to count\n")
		}
		fmt.Println()
	}
}

func listMockCollections(ctx context.Context, cfg *config.MockConfig) {
	client := mock.NewClient(cfg)

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printWarning("No collections found in the mock database")
		return
	}

	// Sort collections alphabetically
	sort.Strings(collections)

	printSuccess(fmt.Sprintf("Found %d mock collections:", len(collections)))
	fmt.Println()

	for i, collection := range collections {
		color.New(color.FgGreen).Printf("%d. %s\n", i+1, collection)

		// Get document count
		documents, err := client.ListDocuments(ctx, collection, 1000)
		if err == nil {
			if len(documents) >= 1000 {
				fmt.Printf("   Documents: %d+ (showing first 1000)\n", len(documents))
			} else {
				fmt.Printf("   Documents: %d\n", len(documents))
			}
		} else {
			fmt.Printf("   Documents: Unable to count\n")
		}

		// Get collection stats
		stats, err := client.GetCollectionStats(ctx, collection)
		if err == nil {
			if embeddingDim, ok := stats["embedding_dimension"].(int); ok {
				fmt.Printf("   Embedding Dimension: %d\n", embeddingDim)
			}
		}
		fmt.Println()
	}
}

func deleteWeaviateCollection(ctx context.Context, cfg interface{}, collectionName string) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Delete the collection
	if err := client.DeleteCollection(ctx, collectionName); err != nil {
		printError(fmt.Sprintf("Failed to delete collection %s: %v", collectionName, err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
}

func deleteMockCollection(ctx context.Context, cfg *config.MockConfig, collectionName string) {
	client := mock.NewClient(cfg)

	// Delete the collection
	if err := client.DeleteCollection(ctx, collectionName); err != nil {
		printError(fmt.Sprintf("Failed to delete collection %s: %v", collectionName, err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
}

func deleteAllWeaviateCollections(ctx context.Context, cfg interface{}) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// List collections first
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printInfo("No collections to delete")
		return
	}

	// Delete each collection
	for _, collection := range collections {
		printInfo(fmt.Sprintf("Deleting collection: %s", collection))
		if err := client.DeleteCollection(ctx, collection); err != nil {
			printError(fmt.Sprintf("Failed to delete collection %s: %v", collection, err))
		} else {
			printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collection))
		}
	}

	printSuccess("All collections deleted successfully!")
}

func deleteAllMockCollections(ctx context.Context, cfg *config.MockConfig) {
	client := mock.NewClient(cfg)

	// List collections first
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printInfo("No collections to delete")
		return
	}

	// Delete each collection
	for _, collection := range collections {
		printInfo(fmt.Sprintf("Deleting collection: %s", collection))
		if err := client.DeleteCollection(ctx, collection); err != nil {
			printError(fmt.Sprintf("Failed to delete collection %s: %v", collection, err))
		} else {
			printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collection))
		}
	}

	printSuccess("All collections deleted successfully!")
}

// confirmAction prompts the user for confirmation
func confirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		// If there's an error reading input, default to "no"
		return false
	}
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
