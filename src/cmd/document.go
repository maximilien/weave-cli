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

// documentCmd represents the document command
var documentCmd = &cobra.Command{
	Use:     "document",
	Aliases: []string{"doc", "docs"},
	Short:   "Document management",
	Long: `Manage documents in vector database collections.

This command provides subcommands to list, show, and delete documents.`,
}

// documentListCmd represents the document list command
var documentListCmd = &cobra.Command{
	Use:     "list COLLECTION_NAME",
	Aliases: []string{"ls", "l"},
	Short:   "List documents in a collection",
	Long: `List documents in a specific collection.

This command shows:
- Document IDs
- Content previews (truncated)
- Metadata information
- Document counts`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentList,
}

// documentShowCmd represents the document show command
var documentShowCmd = &cobra.Command{
	Use:     "show COLLECTION_NAME [DOCUMENT_ID]",
	Aliases: []string{"s"},
	Short:   "Show documents from a collection",
	Long: `Show detailed information about documents from a collection.

You can show documents in two ways:
1. By document ID: weave doc show COLLECTION_NAME DOCUMENT_ID
2. By metadata filter: weave doc show COLLECTION_NAME --metadata key=value

This command displays:
- Full document content
- Complete metadata
- Document ID and collection information`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runDocumentShow,
}

// documentDeleteCmd represents the document delete command
var documentDeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME [DOCUMENT_ID]",
	Aliases: []string{"del", "d"},
	Short:   "Delete documents from a collection",
	Long: `Delete documents from a collection.

You can delete documents in three ways:
1. By document ID: weave doc delete COLLECTION_NAME DOCUMENT_ID
2. By metadata filter: weave doc delete COLLECTION_NAME --metadata key=value
3. By original filename (virtual): weave doc delete COLLECTION_NAME ORIGINAL_FILENAME --virtual

When using --virtual flag, all chunks and images associated with the original filename
will be deleted in one operation.

⚠️  WARNING: This is a destructive operation that will permanently
delete the specified documents. Use with caution!`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runDocumentDelete,
}

// documentDeleteAllCmd represents the document delete-all command
var documentDeleteAllCmd = &cobra.Command{
	Use:     "delete-all COLLECTION_NAME",
	Aliases: []string{"del-all", "da"},
	Short:   "Delete all documents in a collection",
	Long: `Delete all documents from a specific collection.

⚠️  WARNING: This is a destructive operation that will permanently
delete ALL documents in the specified collection. Use with caution!`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentDeleteAll,
}

// documentCountCmd represents the document count command
var documentCountCmd = &cobra.Command{
	Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
	Aliases: []string{"c"},
	Short:   "Count documents in one or more collections",
	Long: `Count the number of documents in one or more collections.

This command returns the total number of documents in the specified collection(s).
You can specify multiple collections to get counts for each one.

Examples:
  weave docs c MyCollection
  weave docs c RagMeDocs RagMeImages
  weave docs c Collection1 Collection2 Collection3`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDocumentCount,
}

func init() {
	rootCmd.AddCommand(documentCmd)
	documentCmd.AddCommand(documentListCmd)
	documentCmd.AddCommand(documentShowCmd)
	documentCmd.AddCommand(documentCountCmd)
	documentCmd.AddCommand(documentDeleteCmd)
	documentCmd.AddCommand(documentDeleteAllCmd)

	// Add flags
	documentListCmd.Flags().IntP("limit", "l", 50, "Maximum number of documents to show")
	documentListCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	documentListCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	documentListCmd.Flags().BoolP("virtual", "w", false, "Show documents in virtual structure (aggregate chunks by original document)")
	documentListCmd.Flags().BoolP("summary", "S", false, "Show a clean summary of documents (works with --virtual)")

	documentShowCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	documentShowCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	documentShowCmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter (format: key=value)")

	documentDeleteCmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter (format: key=value)")
	documentDeleteCmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")
}

func runDocumentList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")
	virtual, _ := cmd.Flags().GetBool("virtual")
	summary, _ := cmd.Flags().GetBool("summary")

	// Adjust limit for virtual listings to prevent timeouts
	// Check if user explicitly set a limit (different from default 50)
	userSetLimit := cmd.Flags().Changed("limit")
	if virtual && !userSetLimit {
		// Use a more conservative default for virtual listings
		// This prevents connection timeouts with large image collections
		limit = 20
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader(fmt.Sprintf("Documents in Collection: %s", collectionName))
	fmt.Println()

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Listing documents in %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		listWeaviateDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
	case config.VectorDBTypeLocal:
		listWeaviateDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
	case config.VectorDBTypeMock:
		listMockDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentShow(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")

	var documentID string
	if len(args) > 1 {
		documentID = args[1]
	}

	// Validate arguments
	if len(metadataFilters) == 0 && documentID == "" {
		printError("Either DOCUMENT_ID or --metadata filter must be provided")
		os.Exit(1)
	}

	if len(metadataFilters) > 0 && documentID != "" {
		printError("Cannot specify both DOCUMENT_ID and --metadata filter")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	if len(metadataFilters) > 0 {
		printHeader("Documents matching metadata filters")
		fmt.Printf("Metadata filters: %v\n", metadataFilters)
	} else {
		printHeader(fmt.Sprintf("Document Details: %s", documentID))
	}
	fmt.Println()

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Retrieving documents from %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		if len(metadataFilters) > 0 {
			showWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines)
		} else {
			showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
		}
	case config.VectorDBTypeLocal:
		if len(metadataFilters) > 0 {
			showWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines)
		} else {
			showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
		}
	case config.VectorDBTypeMock:
		if len(metadataFilters) > 0 {
			showMockDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines)
		} else {
			showMockDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
		}
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentDelete(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	virtual, _ := cmd.Flags().GetBool("virtual")

	var documentID string
	if len(args) > 1 {
		documentID = args[1]
	}

	// Validate arguments
	if len(metadataFilters) == 0 && documentID == "" {
		printError("Either DOCUMENT_ID or --metadata filter must be provided")
		os.Exit(1)
	}

	if virtual && len(metadataFilters) > 0 {
		printError("Cannot use --virtual flag with --metadata filter")
		os.Exit(1)
	}

	if len(metadataFilters) > 0 && documentID != "" && !virtual {
		printError("Cannot specify both DOCUMENT_ID and --metadata filter")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete Document")
	fmt.Println()

	if len(metadataFilters) > 0 {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete documents matching metadata filters from collection '%s'!", collectionName))
		fmt.Printf("Metadata filters: %v\n", metadataFilters)
		fmt.Println()

		// Confirm deletion
		if !confirmAction("Are you sure you want to delete documents matching these metadata filters?") {
			printInfo("Operation cancelled by user")
			return
		}
	} else if virtual {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete ALL chunks and images associated with original filename '%s' from collection '%s'!", documentID, collectionName))
		fmt.Println()

		// Confirm deletion
		if !confirmAction(fmt.Sprintf("Are you sure you want to delete all chunks and images for original filename '%s'?", documentID)) {
			printInfo("Operation cancelled by user")
			return
		}
	} else {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete document '%s' from collection '%s'!", documentID, collectionName))
		fmt.Println()

		// Confirm deletion
		if !confirmAction(fmt.Sprintf("Are you sure you want to delete document '%s'?", documentID)) {
			printInfo("Operation cancelled by user")
			return
		}
	}

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Deleting document from %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		if len(metadataFilters) > 0 {
			deleteWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteWeaviateDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentID)
		} else {
			deleteWeaviateDocument(ctx, dbConfig, collectionName, documentID)
		}
	case config.VectorDBTypeLocal:
		if len(metadataFilters) > 0 {
			deleteWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteWeaviateDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentID)
		} else {
			deleteWeaviateDocument(ctx, dbConfig, collectionName, documentID)
		}
	case config.VectorDBTypeMock:
		if len(metadataFilters) > 0 {
			deleteMockDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteMockDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentID)
		} else {
			deleteMockDocument(ctx, dbConfig, collectionName, documentID)
		}
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentDeleteAll(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete All Documents")
	fmt.Println()

	printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete ALL documents from collection '%s'!", collectionName))
	fmt.Println()

	// Confirm deletion
	if !confirmAction(fmt.Sprintf("Are you sure you want to delete ALL documents from collection '%s'?", collectionName)) {
		printInfo("Operation cancelled by user")
		return
	}

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Deleting all documents from %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		deleteAllWeaviateDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeLocal:
		deleteAllWeaviateDocuments(ctx, dbConfig, collectionName)
	case config.VectorDBTypeMock:
		deleteAllMockDocuments(ctx, dbConfig, collectionName)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func listWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// For virtual mode, we need all documents to properly aggregate them
	// For regular mode, we can use the provided limit
	queryLimit := limit
	if virtual {
		queryLimit = 10000 // High limit to get all documents for proper aggregation
	}

	// List documents
	documents, err := client.ListDocuments(ctx, collectionName, queryLimit)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	if virtual {
		// For virtual mode, also get documents from the corresponding image collection
		// to properly aggregate images with their source documents
		var allDocuments []weaviate.Document
		allDocuments = append(allDocuments, documents...)

		// Check if there's a corresponding image collection
		imageCollectionName := getImageCollectionName(collectionName)
		if imageCollectionName != "" {
			imageDocuments, err := client.ListDocuments(ctx, imageCollectionName, queryLimit)
			if err == nil && len(imageDocuments) > 0 {
				allDocuments = append(allDocuments, imageDocuments...)
			}
		}

		displayVirtualDocuments(allDocuments, collectionName, showLong, shortLines, summary)
	} else {
		displayRegularDocuments(documents, collectionName, showLong, shortLines)
	}
}

func listMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
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

	// For virtual mode, we need all documents to properly aggregate them
	// For regular mode, we can use the provided limit
	queryLimit := limit
	if virtual {
		queryLimit = 10000 // High limit to get all documents for proper aggregation
	}

	// List documents
	documents, err := client.ListDocuments(ctx, collectionName, queryLimit)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	if virtual {
		displayVirtualMockDocuments(documents, collectionName, showLong, shortLines, summary)
	} else {
		displayRegularMockDocuments(documents, collectionName, showLong, shortLines)
	}
}

func showWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string, showLong bool, shortLines int) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Get document
	document, err := client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		printError(fmt.Sprintf("Failed to get document: %v", err))
		os.Exit(1)
	}

	// Display document details
	color.New(color.FgGreen).Printf("Document ID: %s\n", document.ID)
	fmt.Printf("Collection: %s\n", collectionName)
	fmt.Println()

	fmt.Printf("Content:\n")
	if showLong {
		fmt.Printf("%s\n", document.Content)
	} else {
		// Use shortLines to limit content by lines instead of characters
		preview := truncateStringByLines(document.Content, shortLines)
		fmt.Printf("%s\n", preview)
	}
	fmt.Println()

	if len(document.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for key, value := range document.Metadata {
			// Truncate value based on shortLines directive
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := smartTruncate(valueStr, key, shortLines)
			fmt.Printf("  %s: %s\n", key, truncatedValue)
		}
	}
}

func showMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string, showLong bool, shortLines int) {
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

	// Get document
	document, err := client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		printError(fmt.Sprintf("Failed to get document: %v", err))
		os.Exit(1)
	}

	// Display document details
	color.New(color.FgGreen).Printf("Document ID: %s\n", document.ID)
	fmt.Printf("Collection: %s\n", collectionName)
	fmt.Println()

	fmt.Printf("Content:\n")
	if showLong {
		fmt.Printf("%s\n", document.Content)
	} else {
		// Use shortLines to limit content by lines instead of characters
		preview := truncateStringByLines(document.Content, shortLines)
		fmt.Printf("%s\n", preview)
	}
	fmt.Println()

	if len(document.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for key, value := range document.Metadata {
			// Truncate value based on shortLines directive
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := smartTruncate(valueStr, key, shortLines)
			fmt.Printf("  %s: %s\n", key, truncatedValue)
		}
	}
}

func showWeaviateDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string, showLong bool, shortLines int) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Get documents matching the metadata filters
	documents, err := client.GetDocumentsByMetadata(ctx, collectionName, metadataFilters)
	if err != nil {
		printError(fmt.Sprintf("Failed to get documents by metadata: %v", err))
		os.Exit(1)
	}

	if len(documents) == 0 {
		printWarning("No documents found matching the metadata filters")
		return
	}

	// Show each document
	for i, document := range documents {
		if i > 0 {
			fmt.Println(strings.Repeat("=", 80))
			fmt.Println()
		}

		// Display document details
		color.New(color.FgGreen).Printf("Document ID: %s\n", document.ID)
		fmt.Printf("Collection: %s\n", collectionName)
		fmt.Println()

		fmt.Printf("Content:\n")
		if showLong {
			fmt.Printf("%s\n", document.Content)
		} else {
			// Use shortLines to limit content by lines instead of characters
			preview := truncateStringByLines(document.Content, shortLines)
			fmt.Printf("%s\n", preview)
		}
		fmt.Println()

		if len(document.Metadata) > 0 {
			fmt.Printf("Metadata:\n")
			for key, value := range document.Metadata {
				// Truncate value based on shortLines directive
				valueStr := fmt.Sprintf("%v", value)
				truncatedValue := smartTruncate(valueStr, key, shortLines)
				fmt.Printf("  %s: %s\n", key, truncatedValue)
			}
		}
	}

	printSuccess(fmt.Sprintf("Found and displayed %d documents matching metadata filters", len(documents)))
}

func showMockDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string, showLong bool, shortLines int) {
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

	// Get documents matching the metadata filters
	documents, err := client.GetDocumentsByMetadata(ctx, collectionName, metadataFilters)
	if err != nil {
		printError(fmt.Sprintf("Failed to get documents by metadata: %v", err))
		os.Exit(1)
	}

	if len(documents) == 0 {
		printWarning("No documents found matching the metadata filters")
		return
	}

	// Show each document
	for i, document := range documents {
		if i > 0 {
			fmt.Println(strings.Repeat("=", 80))
			fmt.Println()
		}

		// Display document details
		color.New(color.FgGreen).Printf("Document ID: %s\n", document.ID)
		fmt.Printf("Collection: %s\n", collectionName)
		fmt.Println()

		fmt.Printf("Content:\n")
		if showLong {
			fmt.Printf("%s\n", document.Content)
		} else {
			// Use shortLines to limit content by lines instead of characters
			preview := truncateStringByLines(document.Content, shortLines)
			fmt.Printf("%s\n", preview)
		}
		fmt.Println()

		if len(document.Metadata) > 0 {
			fmt.Printf("Metadata:\n")
			for key, value := range document.Metadata {
				// Truncate value based on shortLines directive
				valueStr := fmt.Sprintf("%v", value)
				truncatedValue := smartTruncate(valueStr, key, shortLines)
				fmt.Printf("  %s: %s\n", key, truncatedValue)
			}
		}
	}

	printSuccess(fmt.Sprintf("Found and displayed %d documents matching metadata filters", len(documents)))
}

func deleteWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Delete document
	if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
		printError(fmt.Sprintf("Failed to delete document: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted document '%s' from collection '%s'", documentID, collectionName))
}

func deleteWeaviateDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Delete documents by metadata
	deletedCount, err := client.DeleteDocumentsByMetadata(ctx, collectionName, metadataFilters)
	if err != nil {
		printError(fmt.Sprintf("Failed to delete documents by metadata: %v", err))
		os.Exit(1)
	}

	if deletedCount == 0 {
		printWarning("No documents found matching the metadata filters")
	} else {
		printSuccess(fmt.Sprintf("Successfully deleted %d documents from collection '%s' matching metadata filters", deletedCount, collectionName))
	}
}

func deleteMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string) {
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

	// Delete document
	if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
		printError(fmt.Sprintf("Failed to delete document: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted document '%s' from collection '%s'", documentID, collectionName))
}

func deleteMockDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string) {
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

	// Delete documents by metadata
	deletedCount, err := client.DeleteDocumentsByMetadata(ctx, collectionName, metadataFilters)
	if err != nil {
		printError(fmt.Sprintf("Failed to delete documents by metadata: %v", err))
		os.Exit(1)
	}

	if deletedCount == 0 {
		printWarning("No documents found matching the metadata filters")
	} else {
		printSuccess(fmt.Sprintf("Successfully deleted %d documents from collection '%s' matching metadata filters", deletedCount, collectionName))
	}
}

func deleteAllWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// List documents first to get count
	documents, err := client.ListDocuments(ctx, collectionName, 1000) // Get up to 1000 documents
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printInfo("No documents to delete")
		return
	}

	printInfo(fmt.Sprintf("Found %d documents to delete", len(documents)))

	// Delete each document
	deletedCount := 0
	for _, doc := range documents {
		printInfo(fmt.Sprintf("Deleting document: %s", doc.ID))
		if err := client.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			printError(fmt.Sprintf("Failed to delete document %s: %v", doc.ID, err))
		} else {
			deletedCount++
			printSuccess(fmt.Sprintf("Successfully deleted document: %s", doc.ID))
		}
	}

	printSuccess(fmt.Sprintf("Successfully deleted %d out of %d documents from collection '%s'", deletedCount, len(documents), collectionName))
}

func deleteAllMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
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

	// List documents first to get count
	documents, err := client.ListDocuments(ctx, collectionName, 1000) // Get up to 1000 documents
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printInfo("No documents to delete")
		return
	}

	printInfo(fmt.Sprintf("Found %d documents to delete", len(documents)))

	// Delete each document
	deletedCount := 0
	for _, doc := range documents {
		printInfo(fmt.Sprintf("Deleting document: %s", doc.ID))
		if err := client.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			printError(fmt.Sprintf("Failed to delete document %s: %v", doc.ID, err))
		} else {
			deletedCount++
			printSuccess(fmt.Sprintf("Successfully deleted document: %s", doc.ID))
		}
	}

	printSuccess(fmt.Sprintf("Successfully deleted %d out of %d documents from collection '%s'", deletedCount, len(documents), collectionName))
}

func runDocumentCount(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionNames := args

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	// Handle single collection (backward compatibility)
	if len(collectionNames) == 1 {
		collectionName := collectionNames[0]
		printHeader(fmt.Sprintf("Document Count: %s", collectionName))
		fmt.Println()

		color.New(color.FgCyan, color.Bold).Printf("Counting documents in %s database...\n", dbConfig.Type)
		fmt.Println()

		var count int
		switch dbConfig.Type {
		case config.VectorDBTypeCloud:
			count, err = countWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeLocal:
			count, err = countWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeMock:
			count, err = countMockDocuments(ctx, dbConfig, collectionName)
		default:
			printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			verbose, _ := cmd.Flags().GetBool("verbose")
			// Provide more specific error messages
			if strings.Contains(err.Error(), "does not exist") {
				printError(fmt.Sprintf("Collection '%s' does not exist", collectionName))
				if verbose {
					printWarning(fmt.Sprintf("Details: %v", err))
				}
			} else if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
				printError(fmt.Sprintf("Collection '%s' not found, check database configuration", collectionName))
				if verbose {
					printWarning(fmt.Sprintf("Details: %v", err))
				}
			} else {
				printError(fmt.Sprintf("Failed to count documents: %v", err))
			}
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Found %d documents in collection '%s'", count, collectionName))
		return
	}

	// Handle multiple collections
	printHeader(fmt.Sprintf("Document Count: %d Collections", len(collectionNames)))
	fmt.Println()

	color.New(color.FgCyan, color.Bold).Printf("Counting documents in %s database...\n", dbConfig.Type)
	fmt.Println()

	totalCount := 0
	successCount := 0
	errorCount := 0

	for i, collectionName := range collectionNames {
		color.New(color.FgYellow).Printf("%d. %s: ", i+1, collectionName)

		var count int
		switch dbConfig.Type {
		case config.VectorDBTypeCloud:
			count, err = countWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeLocal:
			count, err = countWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeMock:
			count, err = countMockDocuments(ctx, dbConfig, collectionName)
		default:
			printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			verbose, _ := cmd.Flags().GetBool("verbose")
			// Provide more specific error messages
			if strings.Contains(err.Error(), "does not exist") {
				color.New(color.FgRed).Printf("ERROR - Collection '%s' does not exist\n", collectionName)
				if verbose {
					color.New(color.FgYellow).Printf("  Details: %v\n", err)
				}
			} else if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
				color.New(color.FgRed).Printf("ERROR - Collection '%s' not found, check database configuration\n", collectionName)
				if verbose {
					color.New(color.FgYellow).Printf("  Details: %v\n", err)
				}
			} else {
				color.New(color.FgRed).Printf("ERROR - %v\n", err)
			}
			errorCount++
		} else {
			color.New(color.FgGreen).Printf("%d documents\n", count)
			totalCount += count
			successCount++
		}
	}

	fmt.Println()
	if successCount > 0 {
		printSuccess(fmt.Sprintf("Total documents across %d collections: %d", successCount, totalCount))
	}
	if errorCount > 0 {
		printWarning(fmt.Sprintf("Failed to count %d collection(s)", errorCount))
	}
}

func countWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	// Use the efficient CountDocuments method that doesn't fetch document content
	// This is much faster for collections with large data like base64 images
	count, err := client.CountDocuments(ctx, collectionName)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func countMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
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

	// Get documents with a high limit to count them all
	documents, err := client.ListDocuments(ctx, collectionName, 10000) // High limit to get all documents
	if err != nil {
		return 0, fmt.Errorf("failed to list documents: %w", err)
	}

	return len(documents), nil
}

// VirtualDocument represents a document with its chunks aggregated
type VirtualDocument struct {
	OriginalFilename string
	TotalChunks      int
	Chunks           []weaviate.Document
	Metadata         map[string]interface{}
}

// displayRegularDocuments shows documents in the traditional format
func displayRegularDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int) {
	printSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	fmt.Println()

	for i, doc := range documents {
		fmt.Printf("%d. ", i+1)
		printStyledEmoji("📄")
		fmt.Printf(" ")
		printStyledKeyProminent("ID")
		fmt.Printf(": ")
		printStyledID(doc.ID)
		fmt.Println()

		// Only show content if it's not just the redundant "Document ID: [ID]"
		if doc.Content != fmt.Sprintf("Document ID: %s", doc.ID) {
			fmt.Printf("   ")
			printStyledKeyProminent("Content")
			fmt.Printf(": ")
			if showLong {
				printStyledValue(doc.Content)
			} else {
				// Use smartTruncate to handle base64 content appropriately
				preview := smartTruncate(doc.Content, "content", shortLines)
				printStyledValue(preview)
			}
			fmt.Println()
		}

		if len(doc.Metadata) > 0 {
			fmt.Printf("   ")
			printStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range doc.Metadata {
				if key != "id" { // Skip ID since it's already shown
					// Use smartTruncate to handle base64 content appropriately
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := smartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					printStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}
}

// displayVirtualDocuments shows documents aggregated by their original document
func displayVirtualDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int, summary bool) {
	virtualDocs := aggregateDocumentsByOriginal(documents)

	printSuccess(fmt.Sprintf("Found %d virtual documents in collection '%s' (aggregated from %d total documents):", len(virtualDocs), collectionName, len(documents)))
	fmt.Println()

	for i, vdoc := range virtualDocs {
		fmt.Printf("%d. ", i+1)
		printStyledEmoji("📄")
		fmt.Printf(" Document: ")
		printStyledFilename(vdoc.OriginalFilename)
		fmt.Println()

		// Determine if this is an image collection
		isImageCollection := isImageVirtualDocument(vdoc)

		fmt.Printf("   ")
		if vdoc.TotalChunks > 0 {
			printStyledKeyNumberProminentWithEmoji("Chunks", len(vdoc.Chunks), "📝")
			fmt.Printf("/")
			printStyledNumber(vdoc.TotalChunks)
			fmt.Println()
		} else if isImageCollection {
			printStyledKeyNumberProminentWithEmoji("Images", len(vdoc.Chunks), "🖼️")
			fmt.Println()
		} else {
			printStyledKeyValueProminentWithEmoji("Type", "Single document (no chunks)", "📄")
			fmt.Println()
		}

		// Show metadata from the first chunk or document
		if len(vdoc.Metadata) > 0 {
			fmt.Printf("   ")
			printStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range vdoc.Metadata {
				if key != "id" && key != "chunk_index" && key != "total_chunks" && key != "is_chunked" {
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := smartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					printStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}

		// Show details if there are items
		if len(vdoc.Chunks) > 0 {
			fmt.Printf("   ")
			if isImageCollection {
				printStyledKeyValueProminentWithEmoji("Stack Details", "", "🗂️")
			} else {
				printStyledKeyValueProminentWithEmoji("Chunk Details", "", "📝")
			}
			fmt.Println()

			for j, chunk := range vdoc.Chunks {
				fmt.Printf("     %d. ", j+1)
				printStyledKeyProminent("ID")
				fmt.Printf(": ")
				printStyledID(chunk.ID)
				if chunkIndex, ok := chunk.Metadata["chunk_index"]; ok {
					fmt.Printf(" (")
					printStyledKeyProminent("chunk")
					fmt.Printf(" ")
					printStyledNumber(int(chunkIndex.(float64)))
					fmt.Printf(")")
				}
				fmt.Println()

				if chunk.Content != fmt.Sprintf("Document ID: %s", chunk.ID) {
					fmt.Printf("        ")
					printStyledKeyProminent("Content")
					fmt.Printf(": ")
					if showLong {
						printStyledValueDimmed(chunk.Content)
					} else {
						preview := smartTruncate(chunk.Content, "content", shortLines)
						printStyledValueDimmed(preview)
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	// Show summary if requested
	if summary {
		fmt.Println()
		printStyledKeyValueProminentWithEmoji("Summary", "", "📋")
		fmt.Println()
		for i, vdoc := range virtualDocs {
			fmt.Printf("   %d. ", i+1)
			printStyledFilename(vdoc.OriginalFilename)
			if vdoc.TotalChunks > 0 {
				fmt.Printf(" - %d chunks", len(vdoc.Chunks))
			} else if isImageVirtualDocument(vdoc) {
				fmt.Printf(" - %d images", len(vdoc.Chunks))
			} else {
				fmt.Printf(" - Single document")
			}
			fmt.Println()
		}
	}
}

// aggregateDocumentsByOriginal groups documents by their original filename
func aggregateDocumentsByOriginal(documents []weaviate.Document) []VirtualDocument {
	docMap := make(map[string]*VirtualDocument)

	for _, doc := range documents {
		// Check if this is a chunked document
		if metadata, ok := doc.Metadata["metadata"]; ok {
			if metadataStr, ok := metadata.(string); ok {
				// Parse the JSON metadata to extract original filename
				var metadataObj map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
					if originalFilename, ok := metadataObj["original_filename"].(string); ok {
						if isChunked, ok := metadataObj["is_chunked"].(bool); ok && isChunked {
							// This is a chunk
							if vdoc, exists := docMap[originalFilename]; exists {
								vdoc.Chunks = append(vdoc.Chunks, doc)
							} else {
								totalChunks := 0
								if tc, ok := metadataObj["total_chunks"].(float64); ok {
									totalChunks = int(tc)
								}
								docMap[originalFilename] = &VirtualDocument{
									OriginalFilename: originalFilename,
									TotalChunks:      totalChunks,
									Chunks:           []weaviate.Document{doc},
									Metadata:         metadataObj,
								}
							}
							continue
						}
					}
				}
			}
		}

		// Check if this is an image extracted from a PDF
		groupKey := getImageGroupKey(doc)

		if vdoc, exists := docMap[groupKey]; exists {
			// Add to existing group
			vdoc.Chunks = append(vdoc.Chunks, doc)
		} else {
			// Create new group
			docMap[groupKey] = &VirtualDocument{
				OriginalFilename: groupKey,
				TotalChunks:      0, // Images are not chunks
				Chunks:           []weaviate.Document{doc},
				Metadata:         doc.Metadata,
			}
		}
	}

	// Convert map to slice
	var virtualDocs []VirtualDocument
	for _, vdoc := range docMap {
		virtualDocs = append(virtualDocs, *vdoc)
	}

	// Sort by original filename for consistent output
	sort.Slice(virtualDocs, func(i, j int) bool {
		return virtualDocs[i].OriginalFilename < virtualDocs[j].OriginalFilename
	})

	return virtualDocs
}

// isImageVirtualDocument checks if a virtual document represents an image collection
func isImageVirtualDocument(vdoc VirtualDocument) bool {
	// Check if any of the chunks have image-related metadata
	for _, chunk := range vdoc.Chunks {
		if _, hasImage := chunk.Metadata["image"]; hasImage {
			return true
		}
		if metadata, ok := chunk.Metadata["metadata"]; ok {
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
	}
	return false
}

// getImageCollectionName returns the corresponding image collection name for a document collection
func getImageCollectionName(collectionName string) string {
	// Common patterns for image collections
	imageSuffixes := []string{"Images", "Image", "images", "image"}

	for _, suffix := range imageSuffixes {
		if strings.HasSuffix(collectionName, suffix) {
			return collectionName // Already an image collection
		}
	}

	// Try to find corresponding image collection
	// For example: RagMeDocs -> RagMeImages
	if strings.HasSuffix(collectionName, "Docs") {
		return strings.TrimSuffix(collectionName, "Docs") + "Images"
	}
	if strings.HasSuffix(collectionName, "docs") {
		return strings.TrimSuffix(collectionName, "docs") + "images"
	}

	// Default pattern: add "Images" suffix
	return collectionName + "Images"
}

// getImageGroupKey determines the grouping key for an image document
func getImageGroupKey(doc weaviate.Document) string {
	// Check if this is an image extracted from a PDF
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				// Check for PDF filename
				if pdfFilename, ok := metadataObj["pdf_filename"].(string); ok {
					return pdfFilename
				}
			}
		}
	}

	// Check URL for PDF source
	if url, ok := doc.Metadata["url"].(string); ok {
		if strings.HasPrefix(url, "pdf://") {
			// Extract PDF name from URL like "pdf://ragme-io.pdf/page_4/image_1"
			parts := strings.Split(url, "/")
			if len(parts) >= 2 {
				return parts[1] // Return "ragme-io.pdf"
			}
		}
	}

	// Check filename for PDF source
	if filename, ok := doc.Metadata["filename"].(string); ok {
		if strings.Contains(filename, ".pdf_") {
			// Extract PDF name from filename like "ragme-io.pdf_page_4_image_1.png"
			parts := strings.Split(filename, ".pdf_")
			if len(parts) >= 1 {
				return parts[0] + ".pdf"
			}
		}
	}

	// For standalone images, use the URL or filename as the key
	if url, ok := doc.Metadata["url"].(string); ok {
		return url
	}

	if filename, ok := doc.Metadata["filename"].(string); ok {
		return filename
	}

	// Fallback
	return "Unknown Image"
}

// displayRegularMockDocuments shows mock documents in the traditional format
func displayRegularMockDocuments(documents []mock.Document, collectionName string, showLong bool, shortLines int) {
	printSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	fmt.Println()

	for i, doc := range documents {
		fmt.Printf("%d. ", i+1)
		printStyledEmoji("📄")
		fmt.Printf(" ")
		printStyledKeyProminent("ID")
		fmt.Printf(": ")
		printStyledID(doc.ID)
		fmt.Println()

		// Only show content if it's not just the redundant "Document ID: [ID]"
		if doc.Content != fmt.Sprintf("Document ID: %s", doc.ID) {
			fmt.Printf("   ")
			printStyledKeyProminent("Content")
			fmt.Printf(": ")
			if showLong {
				printStyledValue(doc.Content)
			} else {
				// Use shortLines to limit content by lines instead of characters
				preview := smartTruncate(doc.Content, "content", shortLines)
				printStyledValue(preview)
			}
			fmt.Println()
		}

		if len(doc.Metadata) > 0 {
			fmt.Printf("   ")
			printStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range doc.Metadata {
				if key != "id" { // Skip ID since it's already shown
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := smartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					printStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}
}

// displayVirtualMockDocuments shows mock documents aggregated by their original document
func displayVirtualMockDocuments(documents []mock.Document, collectionName string, showLong bool, shortLines int, summary bool) {
	virtualDocs := aggregateMockDocumentsByOriginal(documents)

	printSuccess(fmt.Sprintf("Found %d virtual documents in collection '%s' (aggregated from %d total documents):", len(virtualDocs), collectionName, len(documents)))
	fmt.Println()

	for i, vdoc := range virtualDocs {
		fmt.Printf("%d. ", i+1)
		printStyledEmoji("📄")
		fmt.Printf(" Document: ")
		printStyledFilename(vdoc.OriginalFilename)
		fmt.Println()

		// Determine if this is an image collection
		isImageCollection := isMockImageVirtualDocument(vdoc)

		fmt.Printf("   ")
		if vdoc.TotalChunks > 0 {
			printStyledKeyNumberProminentWithEmoji("Chunks", len(vdoc.Chunks), "📝")
			fmt.Printf("/")
			printStyledNumber(vdoc.TotalChunks)
			fmt.Println()
		} else if isImageCollection {
			printStyledKeyNumberProminentWithEmoji("Images", len(vdoc.Chunks), "🖼️")
			fmt.Println()
		} else {
			printStyledKeyValueProminentWithEmoji("Type", "Single document (no chunks)", "📄")
			fmt.Println()
		}

		// Show metadata from the first chunk or document
		if len(vdoc.Metadata) > 0 {
			fmt.Printf("   ")
			printStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range vdoc.Metadata {
				if key != "id" && key != "chunk_index" && key != "total_chunks" && key != "is_chunked" {
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := smartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					printStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}

		// Show details if there are items
		if len(vdoc.Chunks) > 0 {
			fmt.Printf("   ")
			if isImageCollection {
				printStyledKeyValueProminentWithEmoji("Stack Details", "", "🗂️")
			} else {
				printStyledKeyValueProminentWithEmoji("Chunk Details", "", "📝")
			}
			fmt.Println()
			for j, chunk := range vdoc.Chunks {
				fmt.Printf("     %d. ", j+1)
				printStyledKeyProminent("ID")
				fmt.Printf(": ")
				printStyledID(chunk.ID)
				if chunkIndex, ok := chunk.Metadata["chunk_index"]; ok {
					fmt.Printf(" (")
					printStyledKeyProminent("chunk")
					fmt.Printf(" ")
					printStyledNumber(int(chunkIndex.(float64)))
					fmt.Printf(")")
				}
				fmt.Println()

				if chunk.Content != fmt.Sprintf("Document ID: %s", chunk.ID) {
					fmt.Printf("        ")
					printStyledKeyProminent("Content")
					fmt.Printf(": ")
					if showLong {
						printStyledValueDimmed(chunk.Content)
					} else {
						preview := smartTruncate(chunk.Content, "content", shortLines)
						printStyledValueDimmed(preview)
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	// Show summary if requested
	if summary {
		fmt.Println()
		printStyledKeyValueProminentWithEmoji("Summary", "", "📋")
		fmt.Println()
		for i, vdoc := range virtualDocs {
			fmt.Printf("   %d. ", i+1)
			printStyledFilename(vdoc.OriginalFilename)
			if vdoc.TotalChunks > 0 {
				fmt.Printf(" - %d chunks", len(vdoc.Chunks))
			} else if isMockImageVirtualDocument(vdoc) {
				fmt.Printf(" - %d images", len(vdoc.Chunks))
			} else {
				fmt.Printf(" - Single document")
			}
			fmt.Println()
		}
	}
}

// MockVirtualDocument represents a mock document with its chunks aggregated
type MockVirtualDocument struct {
	OriginalFilename string
	TotalChunks      int
	Chunks           []mock.Document
	Metadata         map[string]interface{}
}

// aggregateMockDocumentsByOriginal groups mock documents by their original filename
func aggregateMockDocumentsByOriginal(documents []mock.Document) []MockVirtualDocument {
	docMap := make(map[string]*MockVirtualDocument)

	for _, doc := range documents {
		// Check if this is a chunked document
		if metadata, ok := doc.Metadata["metadata"]; ok {
			if metadataStr, ok := metadata.(string); ok {
				// Parse the JSON metadata to extract original filename
				var metadataObj map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
					if originalFilename, ok := metadataObj["original_filename"].(string); ok {
						if isChunked, ok := metadataObj["is_chunked"].(bool); ok && isChunked {
							// This is a chunk
							if vdoc, exists := docMap[originalFilename]; exists {
								vdoc.Chunks = append(vdoc.Chunks, doc)
							} else {
								totalChunks := 0
								if tc, ok := metadataObj["total_chunks"].(float64); ok {
									totalChunks = int(tc)
								}
								docMap[originalFilename] = &MockVirtualDocument{
									OriginalFilename: originalFilename,
									TotalChunks:      totalChunks,
									Chunks:           []mock.Document{doc},
									Metadata:         metadataObj,
								}
							}
							continue
						}
					}
				}
			}
		}

		// Check if this is an image extracted from a PDF
		groupKey := getMockImageGroupKey(doc)

		if vdoc, exists := docMap[groupKey]; exists {
			// Add to existing group
			vdoc.Chunks = append(vdoc.Chunks, doc)
		} else {
			// Create new group
			docMap[groupKey] = &MockVirtualDocument{
				OriginalFilename: groupKey,
				TotalChunks:      0, // Images are not chunks
				Chunks:           []mock.Document{doc},
				Metadata:         doc.Metadata,
			}
		}
	}

	// Convert map to slice
	var virtualDocs []MockVirtualDocument
	for _, vdoc := range docMap {
		virtualDocs = append(virtualDocs, *vdoc)
	}

	// Sort by original filename for consistent output
	sort.Slice(virtualDocs, func(i, j int) bool {
		return virtualDocs[i].OriginalFilename < virtualDocs[j].OriginalFilename
	})

	return virtualDocs
}

// isMockImageVirtualDocument checks if a mock virtual document represents an image collection
func isMockImageVirtualDocument(vdoc MockVirtualDocument) bool {
	// Check if any of the chunks have image-related metadata
	for _, chunk := range vdoc.Chunks {
		if _, hasImage := chunk.Metadata["image"]; hasImage {
			return true
		}
		if metadata, ok := chunk.Metadata["metadata"]; ok {
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
	}
	return false
}

// getMockImageGroupKey determines the grouping key for a mock image document
func getMockImageGroupKey(doc mock.Document) string {
	// Check if this is an image extracted from a PDF
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				// Check for PDF filename
				if pdfFilename, ok := metadataObj["pdf_filename"].(string); ok {
					return pdfFilename
				}
			}
		}
	}

	// Check URL for PDF source
	if url, ok := doc.Metadata["url"].(string); ok {
		if strings.HasPrefix(url, "pdf://") {
			// Extract PDF name from URL like "pdf://ragme-io.pdf/page_4/image_1"
			parts := strings.Split(url, "/")
			if len(parts) >= 2 {
				return parts[1] // Return "ragme-io.pdf"
			}
		}
	}

	// Check filename for PDF source
	if filename, ok := doc.Metadata["filename"].(string); ok {
		if strings.Contains(filename, ".pdf_") {
			// Extract PDF name from filename like "ragme-io.pdf_page_4_image_1.png"
			parts := strings.Split(filename, ".pdf_")
			if len(parts) >= 1 {
				return parts[0] + ".pdf"
			}
		}
	}

	// For standalone images, use the URL or filename as the key
	if url, ok := doc.Metadata["url"].(string); ok {
		return url
	}

	if filename, ok := doc.Metadata["filename"].(string); ok {
		return filename
	}

	// Fallback
	return "Unknown Image"
}

// truncateStringByLines truncates a string to the specified number of lines
func truncateStringByLines(s string, maxLines int) string {
	// If --no-truncate flag is set, return content as-is
	if noTruncate {
		return s
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}

	// Join first maxLines lines and add ellipsis with remaining line count
	truncated := strings.Join(lines[:maxLines], "\n")
	remainingLines := len(lines) - maxLines
	return truncated + fmt.Sprintf("\n... (truncated, %d more lines)", remainingLines)
}

// truncateBase64Content truncates base64 content to a reasonable length
func truncateBase64Content(s string, maxChars int) string {
	// If --no-truncate flag is set, return content as-is
	if noTruncate {
		return s
	}

	if len(s) <= maxChars {
		return s
	}

	// For base64 content, show first part and indicate truncation
	truncated := s[:maxChars]
	return truncated + fmt.Sprintf("... (truncated, %d more characters)", len(s)-maxChars)
}

// isBase64Field checks if a field name suggests it contains base64 data
func isBase64Field(fieldName string) bool {
	base64Fields := []string{
		"image", "base64_data", "data", "content", "payload",
		"attachment", "file_data", "binary_data", "encoded_data",
	}

	fieldNameLower := strings.ToLower(fieldName)
	for _, field := range base64Fields {
		if strings.Contains(fieldNameLower, field) {
			return true
		}
	}
	return false
}

// isBase64Content checks if content looks like base64 data
func isBase64Content(content string) bool {
	// Base64 content is typically very long and contains only base64 characters
	if len(content) < 100 {
		return false
	}

	// Check if it starts with data: URL format
	if strings.HasPrefix(content, "data:") {
		return true
	}

	// Check if it's mostly base64 characters (A-Z, a-z, 0-9, +, /, =)
	base64Chars := 0
	for _, char := range content {
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '+' || char == '/' || char == '=' {
			base64Chars++
		}
	}

	// If more than 90% of characters are base64, consider it base64 content
	return float64(base64Chars)/float64(len(content)) > 0.9
}

// smartTruncate intelligently truncates content based on its type
func smartTruncate(content string, fieldName string, maxLines int) string {
	// If --no-truncate flag is set, return content as-is
	if noTruncate {
		return content
	}

	// Special handling for metadata field - preserve JSON structure
	if fieldName == "metadata" {
		return truncateJSONMetadata(content, maxLines)
	}

	// Check if this looks like base64 content
	if isBase64Field(fieldName) || isBase64Content(content) {
		// For base64 content, limit to a reasonable number of characters
		return truncateBase64Content(content, 200)
	}

	// For regular content, use line-based truncation
	return truncateStringByLines(content, maxLines)
}

// truncateJSONMetadata truncates JSON metadata while preserving structure
func truncateJSONMetadata(jsonStr string, maxLines int) string {
	// If --no-truncate flag is set, return content as-is
	if noTruncate {
		return jsonStr
	}

	// Try to parse the JSON to preserve structure
	var metadataObj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &metadataObj); err != nil {
		// If it's not valid JSON, fall back to regular truncation
		return truncateStringByLines(jsonStr, maxLines)
	}

	// Create a truncated version of the metadata
	truncatedObj := make(map[string]interface{})
	for key, value := range metadataObj {
		valueStr := fmt.Sprintf("%v", value)

		// Check if this value looks like base64 content
		if isBase64Content(valueStr) {
			truncatedObj[key] = truncateBase64Content(valueStr, 200)
		} else if len(valueStr) > 500 {
			// For very long non-base64 values, truncate them
			truncatedObj[key] = truncateStringByLines(valueStr, 3)
		} else {
			truncatedObj[key] = value
		}
	}

	// Convert back to JSON string
	jsonBytes, err := json.MarshalIndent(truncatedObj, "", "  ")
	if err != nil {
		// If marshaling fails, fall back to regular truncation
		return truncateStringByLines(jsonStr, maxLines)
	}

	// Apply line-based truncation to the formatted JSON
	return truncateStringByLines(string(jsonBytes), maxLines)
}

// isURL checks if a string is a valid URL
func isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") || strings.HasPrefix(str, "file://") || strings.HasPrefix(str, "pdf://")
}

// deleteWeaviateDocumentsByOriginalFilename deletes all documents (chunks and images) associated with an original filename
func deleteWeaviateDocumentsByOriginalFilename(ctx context.Context, cfg *config.VectorDBConfig, collectionName, originalFilename string) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Determine if this is a URL or filename and search accordingly
	var documents []weaviate.Document
	if isURL(originalFilename) {
		// For URLs, search in the url field
		documents, err = client.GetDocumentsByMetadata(ctx, collectionName, []string{fmt.Sprintf("url=%s", originalFilename)})
		if err != nil {
			printError(fmt.Sprintf("Failed to query documents by URL: %v", err))
			return
		}
	} else {
		// For filenames, first try original_filename field
		documents, err = client.GetDocumentsByMetadata(ctx, collectionName, []string{fmt.Sprintf("original_filename=%s", originalFilename)})
		if err != nil {
			printError(fmt.Sprintf("Failed to query documents by original filename: %v", err))
			return
		}

		// If no documents found and this looks like a PDF filename, try searching for PDF images
		if len(documents) == 0 && strings.HasSuffix(originalFilename, ".pdf") {
			// Search for PDF images with URLs starting with pdf://filename/
			documents, err = client.GetDocumentsByMetadata(ctx, collectionName, []string{fmt.Sprintf("url=pdf://%s/", originalFilename)})
			if err != nil {
				printError(fmt.Sprintf("Failed to query PDF images by URL pattern: %v", err))
				return
			}
		}

		// If still no documents found, try searching for standalone images by URL
		// Only do this if it's not a PDF filename (to avoid conflicts with PDF images)
		if len(documents) == 0 && !strings.HasSuffix(originalFilename, ".pdf") {
			// Search for standalone images with URL matching the filename exactly
			documents, err = client.GetDocumentsByMetadata(ctx, collectionName, []string{fmt.Sprintf("url=%s", originalFilename)})
			if err != nil {
				printError(fmt.Sprintf("Failed to query standalone images by URL: %v", err))
				return
			}
		}
	}

	if len(documents) == 0 {
		printWarning(fmt.Sprintf("No documents found with original filename '%s' in collection '%s'", originalFilename, collectionName))
		return
	}

	// Delete each document individually
	deletedCount := 0
	for _, doc := range documents {
		if err := client.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to delete document %s: %v\n", doc.ID, err)
			continue
		}
		deletedCount++
	}

	if deletedCount == 0 {
		printWarning("No documents were successfully deleted")
	} else {
		printSuccess(fmt.Sprintf("Successfully deleted %d documents (chunks/images) associated with original filename '%s' from collection '%s'", deletedCount, originalFilename, collectionName))
	}
}

// deleteMockDocumentsByOriginalFilename deletes all mock documents (chunks and images) associated with an original filename
func deleteMockDocumentsByOriginalFilename(ctx context.Context, cfg *config.VectorDBConfig, collectionName, originalFilename string) {
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

	// First, find all documents associated with this original filename
	documents, err := client.ListDocuments(ctx, collectionName, 1000) // Get up to 1000 documents
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	// Filter documents by original filename
	var matchingDocuments []mock.Document
	for _, doc := range documents {
		if metadata, ok := doc.Metadata["metadata"]; ok {
			if metadataStr, ok := metadata.(string); ok {
				// Parse the JSON metadata to extract original filename
				var metadataObj map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
					if docOriginalFilename, ok := metadataObj["original_filename"].(string); ok {
						if docOriginalFilename == originalFilename {
							matchingDocuments = append(matchingDocuments, doc)
						}
					}
				}
			}
		}
	}

	if len(matchingDocuments) == 0 {
		printWarning(fmt.Sprintf("No documents found with original filename '%s' in collection '%s'", originalFilename, collectionName))
		return
	}

	// Delete each document individually
	deletedCount := 0
	for _, doc := range matchingDocuments {
		if err := client.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to delete document %s: %v\n", doc.ID, err)
			continue
		}
		deletedCount++
	}

	if deletedCount == 0 {
		printWarning("No documents were successfully deleted")
	} else {
		printSuccess(fmt.Sprintf("Successfully deleted %d documents (chunks/images) associated with original filename '%s' from collection '%s'", deletedCount, originalFilename, collectionName))
	}
}
