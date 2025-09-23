package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
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
	Use:   "list COLLECTION_NAME",
	Short: "List documents in a collection",
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
	Use:   "show COLLECTION_NAME DOCUMENT_ID",
	Short: "Show a specific document",
	Long: `Show detailed information about a specific document.

This command displays:
- Full document content
- Complete metadata
- Document ID and collection information`,
	Args: cobra.ExactArgs(2),
	Run:  runDocumentShow,
}

// documentDeleteCmd represents the document delete command
var documentDeleteCmd = &cobra.Command{
	Use:   "delete COLLECTION_NAME DOCUMENT_ID",
	Short: "Delete a specific document",
	Long: `Delete a specific document from a collection.

⚠️  WARNING: This is a destructive operation that will permanently
delete the specified document. Use with caution!`,
	Args: cobra.ExactArgs(2),
	Run:  runDocumentDelete,
}

// documentDeleteAllCmd represents the document delete-all command
var documentDeleteAllCmd = &cobra.Command{
	Use:   "delete-all COLLECTION_NAME",
	Short: "Delete all documents in a collection",
	Long: `Delete all documents from a specific collection.

⚠️  WARNING: This is a destructive operation that will permanently
delete ALL documents in the specified collection. Use with caution!`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentDeleteAll,
}

func init() {
	rootCmd.AddCommand(documentCmd)
	documentCmd.AddCommand(documentListCmd)
	documentCmd.AddCommand(documentShowCmd)
	documentCmd.AddCommand(documentDeleteCmd)
	documentCmd.AddCommand(documentDeleteAllCmd)

	// Add flags
	documentListCmd.Flags().IntP("limit", "l", 10, "Maximum number of documents to show")
	documentListCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	documentListCmd.Flags().IntP("short", "s", 10, "Show only first N lines of content (default: 10)")

	documentShowCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	documentShowCmd.Flags().IntP("short", "s", 10, "Show only first N lines of content (default: 10)")
}

func runDocumentList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")

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
		listWeaviateDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines)
	case config.VectorDBTypeLocal:
		listWeaviateDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines)
	case config.VectorDBTypeMock:
		listMockDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentShow(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	documentID := args[1]
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader(fmt.Sprintf("Document Details: %s", documentID))
	fmt.Println()

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan, color.Bold).Printf("Retrieving document from %s database...\n", dbConfig.Type)
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
	case config.VectorDBTypeLocal:
		showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
	case config.VectorDBTypeMock:
		showMockDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentDelete(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")
	collectionName := args[0]
	documentID := args[1]

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete Document")
	fmt.Println()

	printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete document '%s' from collection '%s'!", documentID, collectionName))
	fmt.Println()

	// Confirm deletion
	if !confirmAction(fmt.Sprintf("Are you sure you want to delete document '%s'?", documentID)) {
		printInfo("Operation cancelled by user")
		return
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
		deleteWeaviateDocument(ctx, dbConfig, collectionName, documentID)
	case config.VectorDBTypeLocal:
		deleteWeaviateDocument(ctx, dbConfig, collectionName, documentID)
	case config.VectorDBTypeMock:
		deleteMockDocument(ctx, dbConfig, collectionName, documentID)
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

func listWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// List documents
	documents, err := client.ListDocuments(ctx, collectionName, limit)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	printSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	fmt.Println()

	for i, doc := range documents {
		color.New(color.FgGreen).Printf("%d. ID: %s\n", i+1, doc.ID)

		// Only show content if it's not just the redundant "Document ID: [ID]"
		if doc.Content != fmt.Sprintf("Document ID: %s", doc.ID) {
			if showLong {
				fmt.Printf("   Content: %s\n", doc.Content)
			} else {
				// Use shortLines to limit content by lines instead of characters
				preview := truncateStringByLines(doc.Content, shortLines)
				fmt.Printf("   Content: %s\n", preview)
			}
		}

		if len(doc.Metadata) > 0 {
			// Check if there's any metadata beyond the ID
			hasNonIdMetadata := false
			for key := range doc.Metadata {
				if key != "id" {
					hasNonIdMetadata = true
					break
				}
			}

			if hasNonIdMetadata {
				fmt.Printf("   Metadata:\n")
				for key, value := range doc.Metadata {
					if key != "id" { // Skip ID since it's already shown
						// Show raw value as string, even if it's JSON, truncated by lines
						valueStr := fmt.Sprintf("%v", value)
						truncatedValue := truncateStringByLines(valueStr, shortLines)
						fmt.Printf("     %s: %s\n", key, truncatedValue)
					}
				}
			}
		}
		fmt.Println()
	}
}

func listMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection{
			Name:        col.Name,
			Type:        col.Type,
			Description: col.Description,
		}
	}

	client := mock.NewClient(mockConfig)

	// List documents
	documents, err := client.ListDocuments(ctx, collectionName, limit)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		printWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	printSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	fmt.Println()

	for i, doc := range documents {
		color.New(color.FgGreen).Printf("%d. ID: %s\n", i+1, doc.ID)

		// Only show content if it's not just the redundant "Document ID: [ID]"
		if doc.Content != fmt.Sprintf("Document ID: %s", doc.ID) {
			if showLong {
				fmt.Printf("   Content: %s\n", doc.Content)
			} else {
				// Use shortLines to limit content by lines instead of characters
				preview := truncateStringByLines(doc.Content, shortLines)
				fmt.Printf("   Content: %s\n", preview)
			}
		}

		if len(doc.Metadata) > 0 {
			// Check if there's any metadata beyond the ID
			hasNonIdMetadata := false
			for key := range doc.Metadata {
				if key != "id" {
					hasNonIdMetadata = true
					break
				}
			}

			if hasNonIdMetadata {
				fmt.Printf("   Metadata:\n")
				for key, value := range doc.Metadata {
					if key != "id" { // Skip ID since it's already shown
						// Show raw value as string, even if it's JSON, truncated by lines
						valueStr := fmt.Sprintf("%v", value)
						truncatedValue := truncateStringByLines(valueStr, shortLines)
						fmt.Printf("     %s: %s\n", key, truncatedValue)
					}
				}
			}
		}
		fmt.Println()
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
			// Show raw value as string, even if it's JSON, truncated by lines
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := truncateStringByLines(valueStr, shortLines)
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
		mockConfig.Collections[i] = config.MockCollection{
			Name:        col.Name,
			Type:        col.Type,
			Description: col.Description,
		}
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
			// Show raw value as string, even if it's JSON, truncated by lines
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := truncateStringByLines(valueStr, shortLines)
			fmt.Printf("  %s: %s\n", key, truncatedValue)
		}
	}
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

func deleteMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection{
			Name:        col.Name,
			Type:        col.Type,
			Description: col.Description,
		}
	}

	client := mock.NewClient(mockConfig)

	// Delete document
	if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
		printError(fmt.Sprintf("Failed to delete document: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully deleted document '%s' from collection '%s'", documentID, collectionName))
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
		mockConfig.Collections[i] = config.MockCollection{
			Name:        col.Name,
			Type:        col.Type,
			Description: col.Description,
		}
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

// truncateStringByLines truncates a string to the specified number of lines
func truncateStringByLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}

	// Join first maxLines lines and add ellipsis with remaining line count
	truncated := strings.Join(lines[:maxLines], "\n")
	remainingLines := len(lines) - maxLines
	return truncated + fmt.Sprintf("\n... (truncated, %d more lines)", remainingLines)
}
