package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
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
	Use:     "list [database-name]",
	Aliases: []string{"ls", "l"},
	Short:   "List all collections",
	Long: `List all collections in the configured vector database.

This command shows:
- Collection names
- Document counts (if available)
- Collection metadata (if available)

If no database name is provided, it uses the default database.
Use 'weave config list' to see all available databases.`,
	Run: runCollectionList,
}

// collectionDeleteCmd represents the collection delete command
var collectionDeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME [COLLECTION_NAME...]",
	Aliases: []string{"del", "d"},
	Short:   "Delete one or more collections",
	Long: `Delete one or more collections from the configured vector database.

⚠️  WARNING: This is a destructive operation that will permanently
delete all data in the specified collection(s). Use with caution!

Examples:
  weave cols delete MyCollection
  weave cols d Collection1 Collection2 Collection3
  weave cols del MyCollection --force`,
	Args: cobra.MinimumNArgs(1),
	Run:  runCollectionDelete,
}

// collectionDeleteAllCmd represents the collection delete-all command
var collectionDeleteAllCmd = &cobra.Command{
	Use:     "delete-all",
	Aliases: []string{"del-all", "da"},
	Short:   "Delete all collections",
	Long: `Delete all collections from the configured vector database.

⚠️  WARNING: This is a destructive operation that will permanently
delete all data in all collections. Use with caution!`,
	Run: runCollectionDeleteAll,
}

// collectionCreateCmd represents the collection create command
var collectionCreateCmd = &cobra.Command{
	Use:     "create COLLECTION_NAME",
	Aliases: []string{"c"},
	Short:   "Create a new collection",
	Long: `Create a new collection in the configured vector database.

This command creates a collection with default fields and embedding model.
You can customize the collection by specifying custom fields and embedding model.

Examples:
  weave cols create MyCollection
  weave cols create MyCollection --embedding text-embedding-3-small
  weave cols create MyCollection --field title:text,content:text,metadata:text`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionCreate,
}

// collectionCountCmd represents the collection count command
var collectionCountCmd = &cobra.Command{
	Use:     "count [database-name]",
	Aliases: []string{"C"},
	Short:   "Count collections",
	Long: `Count the number of collections in the configured vector database.

This command returns the total number of collections available.`,
	Run: runCollectionCount,
}

// collectionShowCmd represents the collection show command
var collectionShowCmd = &cobra.Command{
	Use:     "show COLLECTION_NAME",
	Aliases: []string{"s"},
	Short:   "Show collection details",
	Long: `Show detailed information about a specific collection.

This command displays:
- Collection metadata and properties
- Document count
- Creation date (if available)
- Last document date (if available)
- Collection statistics`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionShow,
}

func init() {
	rootCmd.AddCommand(collectionCmd)
	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionCountCmd)
	collectionCmd.AddCommand(collectionShowCmd)
	collectionCmd.AddCommand(collectionCreateCmd)
	collectionCmd.AddCommand(collectionDeleteCmd)
	collectionCmd.AddCommand(collectionDeleteAllCmd)

	// Add flags
	collectionListCmd.Flags().IntP("limit", "l", 100, "Maximum number of collections to show")
	collectionListCmd.Flags().BoolP("virtual", "w", false, "Show collections with virtual structure summary (chunks, images, stacks)")
	collectionShowCmd.Flags().IntP("short", "s", 10, "Show only first N lines of sample document metadata (default: 10)")
	collectionCreateCmd.Flags().StringP("embedding", "e", "text-embedding-3-small", "Embedding model to use for the collection")
	collectionCreateCmd.Flags().StringP("field", "f", "", "Custom fields for the collection (format: name1:type,name2:type)")
	collectionDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	limit, _ := cmd.Flags().GetInt("limit")
	virtual, _ := cmd.Flags().GetBool("virtual")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// If a specific database name is provided, use that database
	var dbConfig *config.VectorDBConfig
	var dbName string
	if len(args) > 0 {
		dbName = args[0]
		dbConfig, err = cfg.GetDatabase(dbName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", dbName, err))
			os.Exit(1)
		}
	} else {
		// Use default database
		dbName = "default"
		dbConfig, err = cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}
	}

	printHeader("Vector Database Collections")
	fmt.Println()

	color.New(color.FgCyan, color.Bold).Printf("Listing collections in %s database (%s)...\n", dbName, dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		listWeaviateCollections(ctx, dbConfig, limit, virtual)
	case config.VectorDBTypeLocal:
		listWeaviateCollections(ctx, dbConfig, limit, virtual)
	case config.VectorDBTypeMock:
		listMockCollections(ctx, dbConfig, limit, virtual)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runCollectionDelete(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	force, _ := cmd.Flags().GetBool("force")
	collectionNames := args

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete Collection(s)")
	fmt.Println()

	if len(collectionNames) == 1 {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete collection '%s' and all its data!", collectionNames[0]))
	} else {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete %d collections and all their data!", len(collectionNames)))
		fmt.Println()
		printInfo("Collections to delete:")
		for i, name := range collectionNames {
			fmt.Printf("  %d. %s\n", i+1, name)
		}
	}
	fmt.Println()

	// Confirm deletion unless --force is used
	if !force {
		var confirmMessage string
		if len(collectionNames) == 1 {
			confirmMessage = fmt.Sprintf("Are you sure you want to delete collection '%s'?", collectionNames[0])
		} else {
			confirmMessage = fmt.Sprintf("Are you sure you want to delete %d collections?", len(collectionNames))
		}

		if !confirmAction(confirmMessage) {
			printInfo("Operation cancelled by user")
			return
		}
	}

	// Get default database (for now, we'll use default for delete operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	if len(collectionNames) == 1 {
		color.New(color.FgCyan, color.Bold).Printf("Deleting collection '%s' in %s database...\n", collectionNames[0], dbConfig.Type)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("Deleting %d collections in %s database...\n", len(collectionNames), dbConfig.Type)
	}
	fmt.Println()

	ctx := context.Background()
	successCount := 0
	errorCount := 0

	for i, collectionName := range collectionNames {
		fmt.Printf("Deleting collection %d/%d: %s\n", i+1, len(collectionNames), collectionName)

		switch dbConfig.Type {
		case config.VectorDBTypeCloud:
			if err := deleteWeaviateCollection(ctx, dbConfig, collectionName); err != nil {
				printError(fmt.Sprintf("Failed to delete collection '%s': %v", collectionName, err))
				errorCount++
			} else {
				printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
				successCount++
			}
		case config.VectorDBTypeLocal:
			if err := deleteWeaviateCollection(ctx, dbConfig, collectionName); err != nil {
				printError(fmt.Sprintf("Failed to delete collection '%s': %v", collectionName, err))
				errorCount++
			} else {
				printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
				successCount++
			}
		case config.VectorDBTypeMock:
			if err := deleteMockCollection(ctx, dbConfig, collectionName); err != nil {
				printError(fmt.Sprintf("Failed to delete collection '%s': %v", collectionName, err))
				errorCount++
			} else {
				printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
				successCount++
			}
		default:
			printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			errorCount++
		}
		fmt.Println()
	}

	// Summary
	if len(collectionNames) > 1 {
		if errorCount == 0 {
			printSuccess(fmt.Sprintf("All %d collections deleted successfully!", successCount))
		} else if successCount == 0 {
			printError(fmt.Sprintf("Failed to delete all %d collections", errorCount))
		} else {
			printWarning(fmt.Sprintf("Deleted %d collections successfully, %d failed", successCount, errorCount))
		}
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

	// First confirmation
	if !confirmAction("Are you sure you want to delete all collections?") {
		printInfo("Operation cancelled by user")
		return
	}

	// Second confirmation with red warning
	fmt.Println()
	color.New(color.FgRed, color.Bold).Println("🚨 FINAL WARNING: This operation CANNOT be undone!")
	color.New(color.FgRed).Println("All collections and their data will be permanently deleted.")
	fmt.Println()
	
	// Require exact "yes" confirmation
	fmt.Print("Type 'yes' to confirm deletion: ")
	var response string
	fmt.Scanln(&response)
	
	if response != "yes" {
		printInfo("Operation cancelled - confirmation not received")
		return
	}

	// Get default database (for now, we'll use default for delete operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Deleting all collections in %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		deleteAllWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeLocal:
		deleteAllWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeMock:
		deleteAllMockCollections(ctx, dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runCollectionCreate(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	embeddingModel, _ := cmd.Flags().GetString("embedding")
	customFields, _ := cmd.Flags().GetString("field")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Create Collection")
	fmt.Println()

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Creating collection '%s' in %s database...\n", collectionName, dbConfig.Type)
	fmt.Println()

	// Parse custom fields if provided
	var fields []weaviate.FieldDefinition
	if customFields != "" {
		fields, err = parseFieldDefinitions(customFields)
		if err != nil {
			printError(fmt.Sprintf("Invalid field definition: %v", err))
			os.Exit(1)
		}
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		createWeaviateCollection(ctx, dbConfig, collectionName, embeddingModel, fields)
	case config.VectorDBTypeLocal:
		createWeaviateCollection(ctx, dbConfig, collectionName, embeddingModel, fields)
	case config.VectorDBTypeMock:
		createMockCollection(ctx, dbConfig, collectionName, embeddingModel, fields)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func listWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool) {
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

	// Apply limit if specified
	if limit > 0 && len(collections) > limit {
		collections = collections[:limit]
	}

	printSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
	if limit > 0 && len(collections) == limit {
		fmt.Printf("(showing first %d collections)\n", limit)
	}
	fmt.Println()

	for i, collection := range collections {
		color.New(color.FgGreen).Printf("%d. %s\n", i+1, collection)

		if virtual {
			// Show virtual structure summary
			showCollectionVirtualSummary(ctx, client, collection)
		} else {
			// Show regular document count using efficient method
			count, err := client.CountDocuments(ctx, collection)
			if err == nil {
				fmt.Printf("   Documents: %d\n", count)
			} else {
				fmt.Printf("   Documents: Unable to count\n")
			}
		}
		fmt.Println()
	}
}

func listMockCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

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

	// Apply limit if specified
	if limit > 0 && len(collections) > limit {
		collections = collections[:limit]
	}

	printSuccess(fmt.Sprintf("Found %d mock collections:", len(collections)))
	if limit > 0 && len(collections) == limit {
		fmt.Printf("(showing first %d collections)\n", limit)
	}
	fmt.Println()

	for i, collection := range collections {
		color.New(color.FgGreen).Printf("%d. %s\n", i+1, collection)

		if virtual {
			// Show virtual structure summary for mock collections
			showMockCollectionVirtualSummary(ctx, client, collection)
		} else {
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

func deleteWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) error {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Delete the collection
	if err := client.DeleteCollection(ctx, collectionName); err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
	}

	return nil
}

func deleteMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) error {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

	// Delete the collection
	if err := client.DeleteCollection(ctx, collectionName); err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
	}

	return nil
}

func deleteAllWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) {
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

func deleteAllMockCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

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

func runCollectionCount(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	var databaseName string
	if len(args) > 0 {
		databaseName = args[0]
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get database configuration
	var dbConfig *config.VectorDBConfig
	if databaseName != "" {
		dbConfig, err = cfg.GetDatabase(databaseName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", databaseName, err))
			os.Exit(1)
		}
	} else {
		dbConfig, err = cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}
	}

	printHeader("Collection Count")
	fmt.Println()

	color.New(color.FgCyan, color.Bold).Printf("Counting collections in %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	var count int
	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		count, err = countWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeLocal:
		count, err = countWeaviateCollections(ctx, dbConfig)
	case config.VectorDBTypeMock:
		count, err = countMockCollections(ctx, dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		verbose, _ := cmd.Flags().GetBool("verbose")
		printError(fmt.Sprintf("Failed to count collections: %v", err))
		if verbose {
			printWarning(fmt.Sprintf("Details: %v", err))
		}
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Found %d collections", count))
}

func countWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collections), nil
}

func countMockCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collections), nil
}

func runCollectionShow(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	shortLines, _ := cmd.Flags().GetInt("short")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader(fmt.Sprintf("Collection Details: %s", collectionName))
	fmt.Println()

	// Get default database (for now, we'll use default for collection operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Retrieving collection information from %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	verbose, _ := cmd.Flags().GetBool("verbose")
	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		showWeaviateCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose)
	case config.VectorDBTypeLocal:
		showWeaviateCollection(ctx, dbConfig, collectionName, shortLines, noTruncate, verbose)
	case config.VectorDBTypeMock:
		showMockCollection(ctx, dbConfig, collectionName, shortLines, noTruncate)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func showWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Check if collection exists by trying to list documents
	_, err = client.ListDocuments(ctx, collectionName, 1) // Just get 1 to check existence
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
			printError(fmt.Sprintf("Collection '%s' not found", collectionName))
		} else {
			printError(fmt.Sprintf("Collection '%s' not found or error accessing it: %v", collectionName, err))
		}
		if verbose {
			printWarning(fmt.Sprintf("Details: %v", err))
		}
		return
	}

	// Get document count using efficient method
	documentCount, err := client.CountDocuments(ctx, collectionName)
	if err != nil {
		printError(fmt.Sprintf("Failed to get document count: %v", err))
		return
	}

	// Display collection information with styling
	printStyledEmoji("📊")
	fmt.Printf(" ")
	printStyledKeyProminent("Collection")
	fmt.Printf(": ")
	printStyledValueDimmed(collectionName)
	fmt.Println()

	printStyledKeyProminent("Database Type")
	fmt.Printf(": ")
	printStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	printStyledKeyProminent("Document Count")
	fmt.Printf(": ")
	printStyledValueDimmed(fmt.Sprintf("%d", documentCount))
	fmt.Println()
	fmt.Println()

	// Show collection properties if available
	printStyledEmoji("🔧")
	fmt.Printf(" ")
	printStyledKeyProminent("Collection Properties")
	fmt.Println()
	printStyledKeyProminent("  Vector Database")
	fmt.Printf(": ")
	printStyledValueDimmed(string(cfg.Type))
	fmt.Println()
	printStyledKeyProminent("  URL")
	fmt.Printf(": ")
	printStyledValueDimmed(cfg.URL)
	fmt.Println()
	if cfg.APIKey != "" {
		printStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		printStyledValueDimmed("[CONFIGURED]")
		fmt.Println()
	} else {
		printStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		printStyledValueDimmed("[NOT CONFIGURED]")
		fmt.Println()
	}
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis (just one document)
		// Use ListDocuments but get the first document ID, then use GetDocument for accurate data
		sampleDocuments, err := client.ListDocuments(ctx, collectionName, 1)
		if err != nil {
			printWarning(fmt.Sprintf("Could not retrieve sample document: %v", err))
		} else if len(sampleDocuments) > 0 {
			// Get the actual document with full data using GetDocument
			sampleDoc, err := client.GetDocument(ctx, collectionName, sampleDocuments[0].ID)
			if err != nil {
				printWarning(fmt.Sprintf("Could not retrieve full sample document: %v", err))
				// Fall back to the basic document from ListDocuments
				sampleDoc = &sampleDocuments[0]
			}
			printStyledEmoji("📋")
			fmt.Printf(" ")
			printStyledKeyProminent("Sample Document Metadata")
			fmt.Println()
			if len(sampleDoc.Metadata) > 0 {
				metadataCount := 0
				for key, value := range sampleDoc.Metadata {
					if metadataCount >= shortLines {
						remainingFields := len(sampleDoc.Metadata) - shortLines
						fmt.Printf("... (truncated, %d more metadata fields)\n", remainingFields)
						break
					}

					var displayValue string
					if noTruncate {
						displayValue = fmt.Sprintf("%v", value)
					} else {
						displayValue = truncateMetadataValue(value, 100) // Limit each value to 100 chars
					}

					// Style the key-value pair directly
					fmt.Printf("  - ")
					printStyledKeyProminent(key)
					fmt.Printf(": ")
					if key == "id" {
						printStyledID(displayValue)
					} else {
						printStyledValueDimmed(displayValue)
					}
					fmt.Println()
					metadataCount++
				}
			} else {
				printStyledKey("  No metadata available")
				fmt.Println()
			}
			fmt.Println()

			// Show sample content
			if len(sampleDoc.Content) > 0 {
				printStyledEmoji("📄")
				fmt.Printf(" ")
				printStyledKeyProminent("Sample Document Content")
				fmt.Println()

				// Check if this is image content (base64 data)
				if isImageContent(sampleDoc.Content) {
					fmt.Printf("  📷 Image Document (Base64 encoded)\n")
					fmt.Printf("  📏 Content Size: %d characters\n", len(sampleDoc.Content))
					fmt.Printf("  🔍 Preview: %s...\n", sampleDoc.Content[:min(50, len(sampleDoc.Content))])
					if len(sampleDoc.Content) > 50 {
						fmt.Printf("  ℹ️  Full base64 data available (use --no-truncate to see all)\n")
					}
				} else {
					// Regular text content
					if noTruncate {
						fmt.Printf("%s\n", sampleDoc.Content)
					} else {
						contentLines := strings.Split(sampleDoc.Content, "\n")
						maxLines := shortLines
						if len(contentLines) > maxLines {
							for i := 0; i < maxLines; i++ {
								fmt.Printf("%s\n", contentLines[i])
							}
							fmt.Printf("  ... (%d more lines)\n", len(contentLines)-maxLines)
						} else {
							fmt.Printf("%s\n", sampleDoc.Content)
						}
					}
				}
				fmt.Println()
			} else {
				// No content, but show if it's an image document based on metadata
				if isImageDocument(sampleDoc.Metadata) {
					printStyledEmoji("📄")
					fmt.Printf(" ")
					printStyledKeyProminent("Sample Document Content")
					fmt.Println()
					fmt.Printf("  📷 Image Document (no text content)\n")
					fmt.Printf("  ℹ️  Image data stored in metadata fields\n")
					fmt.Println()
				}
			}
		}
	}

	printSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
}

func showMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

	// Check if collection exists
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	collectionExists := false
	for _, col := range collections {
		if col == collectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		printError(fmt.Sprintf("Collection '%s' not found", collectionName))
		return
	}

	// Get document count
	documents, err := client.ListDocuments(ctx, collectionName, 10000) // High limit to get count
	if err != nil {
		printError(fmt.Sprintf("Failed to get document count: %v", err))
		return
	}

	documentCount := len(documents)

	// Display collection information with styling
	printStyledEmoji("📊")
	fmt.Printf(" ")
	printStyledKeyProminent("Collection")
	fmt.Printf(": ")
	printStyledValueDimmed(collectionName)
	fmt.Println()

	printStyledKeyProminent("Database Type")
	fmt.Printf(": ")
	printStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	printStyledKeyProminent("Document Count")
	fmt.Printf(": ")
	printStyledValueDimmed(fmt.Sprintf("%d", documentCount))
	fmt.Println()
	fmt.Println()

	// Show collection properties
	printStyledEmoji("🔧")
	fmt.Printf(" ")
	printStyledKeyProminent("Collection Properties")
	fmt.Println()
	printStyledKeyProminent("  Vector Database")
	fmt.Printf(": ")
	printStyledValueDimmed(string(cfg.Type))
	fmt.Println()
	printStyledKeyProminent("  Simulate Embeddings")
	fmt.Printf(": ")
	printStyledValueDimmed(fmt.Sprintf("%t", cfg.SimulateEmbeddings))
	fmt.Println()
	printStyledKeyProminent("  Embedding Dimension")
	fmt.Printf(": ")
	printStyledValueDimmed(fmt.Sprintf("%d", cfg.EmbeddingDimension))
	fmt.Println()
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis
		sampleDoc := documents[0]
		printStyledEmoji("📋")
		fmt.Printf(" ")
		printStyledKeyProminent("Sample Document Metadata")
		fmt.Println()
		if len(sampleDoc.Metadata) > 0 {
			metadataCount := 0
			for key, value := range sampleDoc.Metadata {
				if metadataCount >= shortLines {
					remainingFields := len(sampleDoc.Metadata) - shortLines
					fmt.Printf("... (truncated, %d more metadata fields)\n", remainingFields)
					break
				}

				var displayValue string
				if noTruncate {
					displayValue = fmt.Sprintf("%v", value)
				} else {
					displayValue = truncateMetadataValue(value, 100) // Limit each value to 100 chars
				}

				// Style the key-value pair directly
				fmt.Printf("  - ")
				printStyledKeyProminent(key)
				fmt.Printf(": ")
				if key == "id" {
					printStyledID(displayValue)
				} else {
					printStyledValueDimmed(displayValue)
				}
				fmt.Println()
				metadataCount++
			}
		} else {
			printStyledKey("  No metadata available")
			fmt.Println()
		}
		fmt.Println()
	}

	printSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
}

// CollectionVirtualSummary represents the virtual structure summary of a collection
type CollectionVirtualSummary struct {
	TotalDocuments   int
	VirtualDocuments int
	TotalChunks      int
	TotalImages      int
	ImageStacks      int
	ChunkedDocuments int
	StandaloneImages int
}

// showCollectionVirtualSummary shows virtual structure summary for a collection
func showCollectionVirtualSummary(ctx context.Context, client interface{}, collectionName string) {
	// Get all documents from the collection
	var documents []interface{}

	// Handle different client types
	switch c := client.(type) {
	case *weaviate.WeaveClient:
		docs, listErr := c.ListDocuments(ctx, collectionName, 1000)
		if listErr != nil {
			fmt.Printf("   Virtual Summary: Unable to analyze\n")
			return
		}
		// Convert to interface{} slice
		for _, doc := range docs {
			documents = append(documents, doc)
		}
	default:
		fmt.Printf("   ")
		printStyledKeyValueWithEmoji("Virtual Summary", "Unable to analyze", "📊")
		fmt.Println()
		return
	}

	if len(documents) == 0 {
		fmt.Printf("   ")
		printStyledKeyValueWithEmoji("Virtual Summary", "No documents", "📊")
		fmt.Println()
		return
	}

	// Analyze the collection
	summary := analyzeCollectionVirtualStructure(documents)

	// Display the summary
	fmt.Printf("   ")
	printStyledKeyValueWithEmoji("Virtual Summary", "", "📊")
	fmt.Println()
	fmt.Printf("     ")
	printStyledKeyNumberProminent("Total Documents", summary.TotalDocuments)
	fmt.Println()
	fmt.Printf("     ")
	printStyledKeyNumberProminent("Virtual Documents", summary.VirtualDocuments)
	fmt.Println()

	if summary.ChunkedDocuments > 0 {
		fmt.Printf("     ")
		printStyledKeyProminent("Chunked Documents")
		fmt.Printf(": ")
		printStyledNumber(summary.ChunkedDocuments)
		fmt.Printf(" (")
		printStyledNumber(summary.TotalChunks)
		fmt.Printf(" chunks)\n")
	}

	if summary.TotalImages > 0 {
		fmt.Printf("     ")
		printStyledKeyNumberProminentWithEmoji("Images", summary.TotalImages, "🖼️")
		fmt.Println()
		if summary.ImageStacks > 0 {
			fmt.Printf("     ")
			printStyledKeyNumberProminentWithEmoji("Image Stacks", summary.ImageStacks, "🗂️")
			fmt.Println()
		}
		if summary.StandaloneImages > 0 {
			fmt.Printf("     ")
			printStyledKeyNumberProminentWithEmoji("Standalone Images", summary.StandaloneImages, "🖼️")
			fmt.Println()
		}
	}
}

// analyzeCollectionVirtualStructure analyzes documents to determine virtual structure
func analyzeCollectionVirtualStructure(documents []interface{}) CollectionVirtualSummary {
	summary := CollectionVirtualSummary{
		TotalDocuments: len(documents),
	}

	docMap := make(map[string]bool)   // Track unique virtual documents
	imageMap := make(map[string]bool) // Track unique image sources

	for _, docInterface := range documents {
		// Convert to the appropriate document type
		if doc, ok := docInterface.(weaviate.Document); ok {
			// Check if this is a chunked document
			if metadata, ok := doc.Metadata["metadata"]; ok {
				if metadataStr, ok := metadata.(string); ok {
					var metadataObj map[string]interface{}
					if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
						if originalFilename, ok := metadataObj["original_filename"].(string); ok {
							if isChunked, ok := metadataObj["is_chunked"].(bool); ok && isChunked {
								// This is a chunk
								summary.TotalChunks++
								docMap[originalFilename] = true
								continue
							}
						}
					}
				}
			}

			// Check if this is an image
			if isImageDocument(doc.Metadata) {
				summary.TotalImages++
				groupKey := getImageGroupKey(doc)
				imageMap[groupKey] = true

				// Check if it's a standalone image or from a document
				if strings.Contains(groupKey, ".pdf") {
					summary.ImageStacks++
				} else {
					summary.StandaloneImages++
				}
			} else {
				// Regular document
				docMap[doc.ID] = true
			}
		}
	}

	summary.VirtualDocuments = len(docMap) + len(imageMap)
	summary.ChunkedDocuments = len(docMap)

	return summary
}

// showMockCollectionVirtualSummary shows virtual structure summary for a mock collection
func showMockCollectionVirtualSummary(ctx context.Context, client *mock.Client, collectionName string) {
	// Get all documents from the collection
	documents, err := client.ListDocuments(ctx, collectionName, 1000)
	if err != nil {
		fmt.Printf("   Virtual Summary: Unable to analyze\n")
		return
	}

	if len(documents) == 0 {
		fmt.Printf("   Virtual Summary: No documents\n")
		return
	}

	// Convert to interface{} slice
	var docs []interface{}
	for _, doc := range documents {
		docs = append(docs, doc)
	}

	// Analyze the collection
	summary := analyzeMockCollectionVirtualStructure(docs)

	// Display the summary
	fmt.Printf("   Virtual Summary:\n")
	fmt.Printf("     Total Documents: %d\n", summary.TotalDocuments)
	fmt.Printf("     Virtual Documents: %d\n", summary.VirtualDocuments)

	if summary.ChunkedDocuments > 0 {
		fmt.Printf("     Chunked Documents: %d (%d chunks)\n", summary.ChunkedDocuments, summary.TotalChunks)
	}

	if summary.TotalImages > 0 {
		fmt.Printf("     Images: %d\n", summary.TotalImages)
		if summary.ImageStacks > 0 {
			fmt.Printf("     Image Stacks: %d\n", summary.ImageStacks)
		}
		if summary.StandaloneImages > 0 {
			fmt.Printf("     Standalone Images: %d\n", summary.StandaloneImages)
		}
	}
}

// analyzeMockCollectionVirtualStructure analyzes mock documents to determine virtual structure
func analyzeMockCollectionVirtualStructure(documents []interface{}) CollectionVirtualSummary {
	summary := CollectionVirtualSummary{
		TotalDocuments: len(documents),
	}

	docMap := make(map[string]bool)   // Track unique virtual documents
	imageMap := make(map[string]bool) // Track unique image sources

	for _, docInterface := range documents {
		// Convert to the appropriate document type
		if doc, ok := docInterface.(mock.Document); ok {
			// Check if this is a chunked document
			if metadata, ok := doc.Metadata["metadata"]; ok {
				if metadataStr, ok := metadata.(string); ok {
					var metadataObj map[string]interface{}
					if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
						if originalFilename, ok := metadataObj["original_filename"].(string); ok {
							if isChunked, ok := metadataObj["is_chunked"].(bool); ok && isChunked {
								// This is a chunk
								summary.TotalChunks++
								docMap[originalFilename] = true
								continue
							}
						}
					}
				}
			}

			// Check if this is an image
			if isMockImageDocument(doc) {
				summary.TotalImages++
				groupKey := getMockImageGroupKey(doc)
				imageMap[groupKey] = true

				// Check if it's a standalone image or from a document
				if strings.Contains(groupKey, ".pdf") {
					summary.ImageStacks++
				} else {
					summary.StandaloneImages++
				}
			} else {
				// Regular document
				docMap[doc.ID] = true
			}
		}
	}

	summary.VirtualDocuments = len(docMap) + len(imageMap)
	summary.ChunkedDocuments = len(docMap)

	return summary
}

// isMockImageDocument checks if a mock document represents an image
func isMockImageDocument(doc mock.Document) bool {
	// Check for image field
	if _, hasImage := doc.Metadata["image"]; hasImage {
		return true
	}

	// Check metadata for image-related fields
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if _, hasBase64Data := metadataObj["base64_data"]; hasBase64Data {
					return true
				}
				if _, hasClassification := metadataObj["classification"]; hasClassification {
					return true
				}
			}
		}
	}

	return false
}

// truncateMetadataValue truncates a metadata value to prevent massive dumps
func truncateMetadataValue(value interface{}, maxLength int) string {
	valueStr := fmt.Sprintf("%v", value)

	// If it's already short enough, return as is
	if len(valueStr) <= maxLength {
		return valueStr
	}

	// Truncate and add ellipsis
	truncated := valueStr[:maxLength]
	remainingChars := len(valueStr) - maxLength
	return fmt.Sprintf("%s... (truncated, %d more characters)", truncated, remainingChars)
}

// isImageContent checks if content appears to be base64 encoded image data
func isImageContent(content string) bool {
	// Check if content looks like base64 data (starts with common image prefixes)
	content = strings.TrimSpace(content)
	if len(content) < 20 {
		return false
	}

	// Common base64 image prefixes
	imagePrefixes := []string{
		"data:image/",
		"/9j/",        // JPEG
		"iVBORw0KGgo", // PNG
		"R0lGOD",      // GIF
		"UklGR",       // WebP
	}

	for _, prefix := range imagePrefixes {
		if strings.HasPrefix(content, prefix) {
			return true
		}
	}

	// Check if it's a long base64 string (likely image data)
	if len(content) > 1000 && isBase64String(content) {
		return true
	}

	return false
}

// isImageDocument checks if metadata indicates this is an image document
func isImageDocument(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}

	// Check for image-related metadata fields
	imageFields := []string{"content_type", "file_type", "type", "mime_type"}
	for _, field := range imageFields {
		if value, exists := metadata[field]; exists {
			valueStr := strings.ToLower(fmt.Sprintf("%v", value))
			if strings.Contains(valueStr, "image") {
				return true
			}
		}
	}

	// Check for image-related field names
	for key := range metadata {
		keyLower := strings.ToLower(key)
		if strings.Contains(keyLower, "image") || strings.Contains(keyLower, "base64") {
			return true
		}
	}

	return false
}

// isBase64String checks if a string appears to be base64 encoded
func isBase64String(s string) bool {
	// Base64 characters: A-Z, a-z, 0-9, +, /, =
	base64Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="

	if len(s) == 0 {
		return false
	}

	// Check if most characters are base64 characters
	base64Count := 0
	for _, char := range s {
		if strings.ContainsRune(base64Chars, char) {
			base64Count++
		}
	}

	// If more than 90% of characters are base64 characters, likely base64
	return float64(base64Count)/float64(len(s)) > 0.9
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// parseFieldDefinitions parses field definitions from command line input
func parseFieldDefinitions(fieldStr string) ([]weaviate.FieldDefinition, error) {
	var fields []weaviate.FieldDefinition

	// Split by comma to get individual field definitions
	fieldParts := strings.Split(fieldStr, ",")

	for _, part := range fieldParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon to get name:type
		nameTypeParts := strings.Split(part, ":")
		if len(nameTypeParts) != 2 {
			return nil, fmt.Errorf("invalid field format '%s', expected 'name:type'", part)
		}

		name := strings.TrimSpace(nameTypeParts[0])
		fieldType := strings.TrimSpace(nameTypeParts[1])

		if name == "" || fieldType == "" {
			return nil, fmt.Errorf("field name and type cannot be empty in '%s'", part)
		}

		// Validate field type
		if !isValidFieldType(fieldType) {
			return nil, fmt.Errorf("invalid field type '%s', supported types: text, int, float, bool, date, object", fieldType)
		}

		fields = append(fields, weaviate.FieldDefinition{
			Name: name,
			Type: fieldType,
		})
	}

	return fields, nil
}

// isValidFieldType checks if a field type is valid
func isValidFieldType(fieldType string) bool {
	validTypes := []string{"text", "int", "float", "bool", "date", "object"}
	for _, validType := range validTypes {
		if fieldType == validType {
			return true
		}
	}
	return false
}

// createWeaviateCollection creates a collection in Weaviate
func createWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Create the collection using Weaviate's REST API
	err = client.CreateCollection(ctx, collectionName, embeddingModel, customFields)
	if err != nil {
		printError(fmt.Sprintf("Failed to create collection '%s': %v", collectionName, err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully created collection: %s", collectionName))

	// Show collection details
	if len(customFields) > 0 {
		fmt.Println()
		printInfo("Custom fields:")
		for _, field := range customFields {
			fmt.Printf("  - %s: %s\n", field.Name, field.Type)
		}
	}

	fmt.Println()
	printInfo(fmt.Sprintf("Embedding model: %s", embeddingModel))
}

// createMockCollection creates a collection in Mock database
func createMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 1536, // Default OpenAI embedding dimension
		Collections:        []config.MockCollection{},
	}

	client := mock.NewClient(mockConfig)

	// Create the collection
	err := client.CreateCollection(ctx, collectionName, embeddingModel, customFields)
	if err != nil {
		printError(fmt.Sprintf("Failed to create collection '%s': %v", collectionName, err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully created collection: %s", collectionName))

	// Show collection details
	if len(customFields) > 0 {
		fmt.Println()
		printInfo("Custom fields:")
		for _, field := range customFields {
			fmt.Printf("  - %s: %s\n", field.Name, field.Type)
		}
	}

	fmt.Println()
	printInfo(fmt.Sprintf("Embedding model: %s", embeddingModel))
}
