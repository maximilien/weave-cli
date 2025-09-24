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
	Use:     "delete COLLECTION_NAME",
	Aliases: []string{"del", "d"},
	Short:   "Delete a specific collection",
	Long: `Delete a specific collection from the configured vector database.

⚠️  WARNING: This is a destructive operation that will permanently
delete all data in the specified collection. Use with caution!`,
	Args: cobra.ExactArgs(1),
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

// collectionCountCmd represents the collection count command
var collectionCountCmd = &cobra.Command{
	Use:     "count [database-name]",
	Aliases: []string{"c"},
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
	collectionCmd.AddCommand(collectionDeleteCmd)
	collectionCmd.AddCommand(collectionDeleteAllCmd)

	// Add flags
	collectionListCmd.Flags().IntP("limit", "l", 100, "Maximum number of collections to show")
	collectionShowCmd.Flags().IntP("short", "s", 10, "Show only first N lines of sample document metadata (default: 10)")
}

func runCollectionList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	limit, _ := cmd.Flags().GetInt("limit")

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
		listWeaviateCollections(ctx, dbConfig, limit)
	case config.VectorDBTypeLocal:
		listWeaviateCollections(ctx, dbConfig, limit)
	case config.VectorDBTypeMock:
		listMockCollections(ctx, dbConfig, limit)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
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

	// Get default database (for now, we'll use default for delete operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Deleting collection '%s' in %s database...\n", collectionName, dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		deleteWeaviateCollection(ctx, dbConfig, collectionName)
	case config.VectorDBTypeLocal:
		deleteWeaviateCollection(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		deleteMockCollection(ctx, dbConfig, collectionName)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
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

func listWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int) {
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

func listMockCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int) {
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

func deleteWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
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

func deleteMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
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
		printError(fmt.Sprintf("Failed to delete collection %s: %v", collectionName, err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted collection: %s", collectionName))
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
		printError(fmt.Sprintf("Failed to count collections: %v", err))
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

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		showWeaviateCollection(ctx, dbConfig, collectionName, shortLines)
	case config.VectorDBTypeLocal:
		showWeaviateCollection(ctx, dbConfig, collectionName, shortLines)
	case config.VectorDBTypeMock:
		showMockCollection(ctx, dbConfig, collectionName, shortLines)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func showWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Check if collection exists by trying to list documents
	_, err = client.ListDocuments(ctx, collectionName, 1) // Just get 1 to check existence
	if err != nil {
		printError(fmt.Sprintf("Collection '%s' not found or error accessing it: %v", collectionName, err))
		return
	}

	// Get document count
	allDocuments, err := client.ListDocuments(ctx, collectionName, 10000) // High limit to get count
	if err != nil {
		printError(fmt.Sprintf("Failed to get document count: %v", err))
		return
	}

	documentCount := len(allDocuments)

	// Display collection information
	color.New(color.FgGreen).Printf("Collection Name: %s\n", collectionName)
	fmt.Printf("Database Type: %s\n", cfg.Type)
	fmt.Printf("Document Count: %d\n", documentCount)
	fmt.Println()

	// Show collection properties if available
	fmt.Printf("Collection Properties:\n")
	fmt.Printf("  - Vector Database: %s\n", cfg.Type)
	fmt.Printf("  - URL: %s\n", cfg.URL)
	if cfg.APIKey != "" {
		fmt.Printf("  - API Key: [CONFIGURED]\n")
	} else {
		fmt.Printf("  - API Key: [NOT CONFIGURED]\n")
	}
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis
		sampleDoc := allDocuments[0]
		fmt.Printf("Sample Document Metadata:\n")
		if len(sampleDoc.Metadata) > 0 {
			metadataLines := make([]string, 0, len(sampleDoc.Metadata))
			for key, value := range sampleDoc.Metadata {
				truncatedValue := truncateMetadataValue(value, 100) // Limit each value to 100 chars
				metadataLines = append(metadataLines, fmt.Sprintf("  - %s: %s", key, truncatedValue))
			}

			// Apply truncation if needed
			if len(metadataLines) > shortLines {
				for i := 0; i < shortLines; i++ {
					fmt.Println(metadataLines[i])
				}
				remainingLines := len(metadataLines) - shortLines
				fmt.Printf("... (truncated, %d more metadata fields)\n", remainingLines)
			} else {
				for _, line := range metadataLines {
					fmt.Println(line)
				}
			}
		} else {
			fmt.Printf("  - No metadata available\n")
		}
		fmt.Println()
	}

	printSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
}

func showMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int) {
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

	// Display collection information
	color.New(color.FgGreen).Printf("Collection Name: %s\n", collectionName)
	fmt.Printf("Database Type: %s\n", cfg.Type)
	fmt.Printf("Document Count: %d\n", documentCount)
	fmt.Println()

	// Show collection properties
	fmt.Printf("Collection Properties:\n")
	fmt.Printf("  - Vector Database: %s\n", cfg.Type)
	fmt.Printf("  - Simulate Embeddings: %t\n", cfg.SimulateEmbeddings)
	fmt.Printf("  - Embedding Dimension: %d\n", cfg.EmbeddingDimension)
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis
		sampleDoc := documents[0]
		fmt.Printf("Sample Document Metadata:\n")
		if len(sampleDoc.Metadata) > 0 {
			metadataLines := make([]string, 0, len(sampleDoc.Metadata))
			for key, value := range sampleDoc.Metadata {
				truncatedValue := truncateMetadataValue(value, 100) // Limit each value to 100 chars
				metadataLines = append(metadataLines, fmt.Sprintf("  - %s: %s", key, truncatedValue))
			}

			// Apply truncation if needed
			if len(metadataLines) > shortLines {
				for i := 0; i < shortLines; i++ {
					fmt.Println(metadataLines[i])
				}
				remainingLines := len(metadataLines) - shortLines
				fmt.Printf("... (truncated, %d more metadata fields)\n", remainingLines)
			} else {
				for _, line := range metadataLines {
					fmt.Println(line)
				}
			}
		} else {
			fmt.Printf("  - No metadata available\n")
		}
		fmt.Println()
	}

	printSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
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
