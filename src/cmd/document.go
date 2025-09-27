package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

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

You can show documents in three ways:
1. By document ID: weave doc show COLLECTION_NAME DOCUMENT_ID
2. By metadata filter: weave doc show COLLECTION_NAME --metadata key=value
3. By filename/name: weave doc show COLLECTION_NAME --name filename.pdf
   (or use --filename as an alias)

This command displays:
- Full document content
- Complete metadata
- Document ID and collection information

Use --schema to show the document schema including metadata structure.
Use --expand-metadata to show expanded metadata information.`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runDocumentShow,
}

// documentDeleteCmd represents the document delete command
var documentDeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME [DOCUMENT_ID] [DOCUMENT_ID...]",
	Aliases: []string{"del", "d"},
	Short:   "Delete documents from a collection",
	Long: `Delete documents from a collection.

You can delete documents in six ways:
1. By single document ID: weave doc delete COLLECTION_NAME DOCUMENT_ID
2. By multiple document IDs: weave doc delete COLLECTION_NAME DOC_ID1 DOC_ID2 DOC_ID3
3. By metadata filter: weave doc delete COLLECTION_NAME --metadata key=value
4. By filename/name: weave doc delete COLLECTION_NAME --name filename.pdf
   (or use --filename as an alias)
5. By original filename (virtual): weave doc delete COLLECTION_NAME ORIGINAL_FILENAME --virtual
6. By pattern: weave doc delete COLLECTION_NAME --pattern "tmp*.png"

Pattern types (auto-detected):
- Shell glob: tmp*.png, tmp?.png, tmp[0-9].png
- Regex: tmp.*\.png, ^tmp.*\.png$, .*\.(png|jpg)$

When using --virtual flag, all chunks and images associated with the original filename
will be deleted in one operation.

Examples:
  weave docs delete MyCollection doc123
  weave docs d MyCollection doc123 doc456 doc789
  weave docs delete MyCollection --metadata filename=test.pdf
  weave docs delete MyCollection --name test_image.png
  weave docs delete MyCollection --filename test_image.png
  weave docs delete MyCollection test.pdf --virtual
  weave docs delete MyCollection --pattern "tmp*.png"
  weave docs delete MyCollection --pattern "tmp.*\.png"

⚠️  WARNING: This is a destructive operation that will permanently
delete the specified documents. Use with caution!`,
	Args: cobra.MinimumNArgs(1),
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

// documentCreateCmd represents the document create command
var documentCreateCmd = &cobra.Command{
	Use:     "create COLLECTION_NAME FILE_PATH",
	Aliases: []string{"c"},
	Short:   "Create a document from a file",
	Long: `Create a document in a collection from a file.

Supported file types:
- Text files (.txt, .md, .json, etc.) - Content goes to 'text' field
- Image files (.jpg, .jpeg, .png, .gif, etc.) - Base64 data goes to 'image_data' field
- PDF files (.pdf) - Text extracted and chunked, goes to 'text' field

The command will automatically:
- Detect file type and process accordingly
- Generate appropriate metadata
- Chunk text content (default 1000 chars, configurable with --chunk-size)
- Create documents following RagMeDocs/RagMeImages schema

Examples:
  weave docs create MyCollection document.txt
  weave docs create MyCollection image.jpg
  weave docs create MyCollection document.pdf --chunk-size 500`,
	Args: cobra.ExactArgs(2),
	Run:  runDocumentCreate,
}

// documentCountCmd represents the document count command
var documentCountCmd = &cobra.Command{
	Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
	Aliases: []string{"C"},
	Short:   "Count documents in one or more collections",
	Long: `Count the number of documents in one or more collections.

This command returns the total number of documents in the specified collection(s).
You can specify multiple collections to get counts for each one.

Examples:
  weave docs C MyCollection
  weave docs C RagMeDocs RagMeImages
  weave docs C Collection1 Collection2 Collection3`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDocumentCount,
}

func init() {
	rootCmd.AddCommand(documentCmd)
	documentCmd.AddCommand(documentListCmd)
	documentCmd.AddCommand(documentShowCmd)
	documentCmd.AddCommand(documentCountCmd)
	documentCmd.AddCommand(documentCreateCmd)
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
	documentShowCmd.Flags().StringP("name", "n", "", "Show document by filename/name")
	documentShowCmd.Flags().StringP("filename", "f", "", "Show document by filename/name (alias for --name)")
	documentShowCmd.Flags().Bool("schema", false, "Show document schema including metadata structure")
	documentShowCmd.Flags().Bool("expand-metadata", false, "Show expanded metadata information")

	documentCreateCmd.Flags().IntP("chunk-size", "s", 1000, "Chunk size for text content (default: 1000 characters)")

	documentDeleteCmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter (format: key=value)")
	documentDeleteCmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")
	documentDeleteCmd.Flags().StringP("pattern", "p", "", "Delete documents matching pattern (auto-detects shell glob vs regex)")
	documentDeleteCmd.Flags().StringP("name", "n", "", "Delete document by filename/name")
	documentDeleteCmd.Flags().StringP("filename", "F", "", "Delete document by filename/name (alias for --name)")
	documentDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runDocumentList(cmd *cobra.Command, args []string) {
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
	cfg, err := loadConfigWithOverrides()
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
	collectionName := args[0]
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	documentName, _ := cmd.Flags().GetString("name")
	if documentName == "" {
		documentName, _ = cmd.Flags().GetString("filename")
	}
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")
	showSchema, _ := cmd.Flags().GetBool("schema")
	showMetadata, _ := cmd.Flags().GetBool("expand-metadata")

	var documentID string
	if len(args) > 1 {
		documentID = args[1]
	}

	// Validate arguments
	if len(metadataFilters) == 0 && documentID == "" && documentName == "" {
		printError("Either DOCUMENT_ID, --metadata filter, or --name must be provided")
		os.Exit(1)
	}

	// Check for conflicting arguments
	providedArgs := 0
	if len(metadataFilters) > 0 {
		providedArgs++
	}
	if documentID != "" {
		providedArgs++
	}
	if documentName != "" {
		providedArgs++
	}

	if providedArgs > 1 {
		printError("Cannot specify multiple methods (DOCUMENT_ID, --metadata filter, or --name) at the same time")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	if len(metadataFilters) > 0 {
		printHeader("Documents matching metadata filters")
		fmt.Printf("Metadata filters: %v\n", metadataFilters)
	} else if documentName != "" {
		printHeader(fmt.Sprintf("Document Details: %s", documentName))
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
			showWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines, showSchema, showMetadata)
		} else if documentName != "" {
			showWeaviateDocumentByName(ctx, dbConfig, collectionName, documentName, showLong, shortLines, showSchema, showMetadata)
		} else {
			showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines, showSchema, showMetadata)
		}
	case config.VectorDBTypeLocal:
		if len(metadataFilters) > 0 {
			showWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines, showSchema, showMetadata)
		} else if documentName != "" {
			showWeaviateDocumentByName(ctx, dbConfig, collectionName, documentName, showLong, shortLines, showSchema, showMetadata)
		} else {
			showWeaviateDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines, showSchema, showMetadata)
		}
	case config.VectorDBTypeMock:
		if len(metadataFilters) > 0 {
			showMockDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters, showLong, shortLines, showSchema, showMetadata)
		} else if documentName != "" {
			showMockDocumentByName(ctx, dbConfig, collectionName, documentName, showLong, shortLines, showSchema, showMetadata)
		} else {
			showMockDocument(ctx, dbConfig, collectionName, documentID, showLong, shortLines, showSchema, showMetadata)
		}
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentDelete(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	collectionName := args[0]
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	virtual, _ := cmd.Flags().GetBool("virtual")
	pattern, _ := cmd.Flags().GetString("pattern")
	documentName, _ := cmd.Flags().GetString("name")
	if documentName == "" {
		documentName, _ = cmd.Flags().GetString("filename")
	}

	// Get document IDs (all args after collection name)
	documentIDs := args[1:]

	// Validate arguments
	if len(metadataFilters) == 0 && len(documentIDs) == 0 && pattern == "" && documentName == "" {
		printError("Either DOCUMENT_ID(s), --metadata filter, --pattern, or --name must be provided")
		os.Exit(1)
	}

	if virtual && len(metadataFilters) > 0 {
		printError("Cannot use --virtual flag with --metadata filter")
		os.Exit(1)
	}

	if virtual && pattern != "" {
		printError("Cannot use --virtual flag with --pattern")
		os.Exit(1)
	}

	if virtual && documentName != "" {
		printError("Cannot use --virtual flag with --name")
		os.Exit(1)
	}

	if len(metadataFilters) > 0 && len(documentIDs) > 0 && !virtual && pattern == "" && documentName == "" {
		printError("Cannot specify both DOCUMENT_ID(s) and --metadata filter")
		os.Exit(1)
	}

	if pattern != "" && len(documentIDs) > 0 {
		printError("Cannot specify both DOCUMENT_ID(s) and --pattern")
		os.Exit(1)
	}

	if pattern != "" && len(metadataFilters) > 0 {
		printError("Cannot specify both --metadata filter and --pattern")
		os.Exit(1)
	}

	if documentName != "" && len(documentIDs) > 0 {
		printError("Cannot specify both DOCUMENT_ID(s) and --name")
		os.Exit(1)
	}

	if documentName != "" && len(metadataFilters) > 0 {
		printError("Cannot specify both --metadata filter and --name")
		os.Exit(1)
	}

	if documentName != "" && pattern != "" {
		printError("Cannot specify both --name and --pattern")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete Document(s)")
	fmt.Println()

	if len(metadataFilters) > 0 {
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete documents matching metadata filters from collection '%s'!", collectionName))
		fmt.Printf("Metadata filters: %v\n", metadataFilters)
		fmt.Println()

		// Confirm deletion unless --force is used
		if !force && !confirmAction("Are you sure you want to delete documents matching these metadata filters?") {
			printInfo("Operation cancelled by user")
			return
		}
	} else if virtual {
		if len(documentIDs) != 1 {
			printError("Virtual deletion requires exactly one original filename")
			os.Exit(1)
		}
		originalFilename := documentIDs[0]
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete ALL chunks and images associated with original filename '%s' from collection '%s'!", originalFilename, collectionName))
		fmt.Println()

		// Confirm deletion unless --force is used
		if !force && !confirmAction(fmt.Sprintf("Are you sure you want to delete all chunks and images for original filename '%s'?", originalFilename)) {
			printInfo("Operation cancelled by user")
			return
		}
	} else if pattern != "" {
		// Pattern-based deletion
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete documents matching pattern '%s' from collection '%s'!", pattern, collectionName))
		fmt.Println()

		// First, let's find matching documents to show the user
		// Get default database (for now, we'll use default for document operations)
		dbConfig, err := cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}

		matchingDocs, err := findDocumentsByPattern(cfg, dbConfig, collectionName, pattern, false)
		if err != nil {
			printError(fmt.Sprintf("Failed to find documents matching pattern: %v", err))
			os.Exit(1)
		}

		if len(matchingDocs) == 0 {
			printInfo(fmt.Sprintf("No documents found matching pattern '%s'", pattern))
			return
		}

		printInfo(fmt.Sprintf("Found %d documents matching pattern '%s':", len(matchingDocs), pattern))
		for i, doc := range matchingDocs {
			filename := getDocumentDisplayName(doc)
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, filename, doc.ID)
		}
		fmt.Println()

		// Confirm deletion unless --force is used
		if !force && !confirmAction(fmt.Sprintf("Are you sure you want to delete %d documents matching pattern '%s'?", len(matchingDocs), pattern)) {
			printInfo("Operation cancelled by user")
			return
		}
	} else if documentName != "" {
		// Name-based deletion
		printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete document '%s' from collection '%s'!", documentName, collectionName))
		fmt.Println()

		// Confirm deletion unless --force is used
		if !force && !confirmAction(fmt.Sprintf("Are you sure you want to delete document '%s'?", documentName)) {
			printInfo("Operation cancelled by user")
			return
		}
	} else {
		// Multiple document IDs
		if len(documentIDs) == 1 {
			printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete document '%s' from collection '%s'!", documentIDs[0], collectionName))
		} else {
			printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete %d documents from collection '%s'!", len(documentIDs), collectionName))
			fmt.Println()
			printInfo("Documents to delete:")
			for i, docID := range documentIDs {
				fmt.Printf("  %d. %s\n", i+1, docID)
			}
		}
		fmt.Println()

		// Confirm deletion unless --force is used
		if !force {
			var confirmMessage string
			if len(documentIDs) == 1 {
				confirmMessage = fmt.Sprintf("Are you sure you want to delete document '%s'?", documentIDs[0])
			} else {
				confirmMessage = fmt.Sprintf("Are you sure you want to delete %d documents?", len(documentIDs))
			}

			if !confirmAction(confirmMessage) {
				printInfo("Operation cancelled by user")
				return
			}
		}
	}

	// Get default database (for now, we'll use default for document operations)
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	if len(metadataFilters) > 0 {
		color.New(color.FgCyan, color.Bold).Printf("Deleting documents by metadata from %s database...\n", dbConfig.Type)
	} else if virtual {
		color.New(color.FgCyan, color.Bold).Printf("Deleting documents by original filename from %s database...\n", dbConfig.Type)
	} else if pattern != "" {
		color.New(color.FgCyan, color.Bold).Printf("Deleting documents by pattern from %s database...\n", dbConfig.Type)
	} else if documentName != "" {
		color.New(color.FgCyan, color.Bold).Printf("Deleting document by name from %s database...\n", dbConfig.Type)
	} else if len(documentIDs) == 1 {
		color.New(color.FgCyan, color.Bold).Printf("Deleting document from %s database...\n", dbConfig.Type)
	} else {
		color.New(color.FgCyan, color.Bold).Printf("Deleting %d documents from %s database...\n", len(documentIDs), dbConfig.Type)
	}
	fmt.Println()

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		if len(metadataFilters) > 0 {
			deleteWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteWeaviateDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentIDs[0])
		} else if pattern != "" {
			deleteWeaviateDocumentsByPattern(ctx, dbConfig, collectionName, pattern)
		} else if documentName != "" {
			deleteWeaviateDocumentByName(ctx, dbConfig, collectionName, documentName)
		} else {
			deleteMultipleWeaviateDocuments(ctx, dbConfig, collectionName, documentIDs)
		}
	case config.VectorDBTypeLocal:
		if len(metadataFilters) > 0 {
			deleteWeaviateDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteWeaviateDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentIDs[0])
		} else if pattern != "" {
			deleteWeaviateDocumentsByPattern(ctx, dbConfig, collectionName, pattern)
		} else if documentName != "" {
			deleteWeaviateDocumentByName(ctx, dbConfig, collectionName, documentName)
		} else {
			deleteMultipleWeaviateDocuments(ctx, dbConfig, collectionName, documentIDs)
		}
	case config.VectorDBTypeMock:
		if len(metadataFilters) > 0 {
			deleteMockDocumentsByMetadata(ctx, dbConfig, collectionName, metadataFilters)
		} else if virtual {
			deleteMockDocumentsByOriginalFilename(ctx, dbConfig, collectionName, documentIDs[0])
		} else if pattern != "" {
			deleteMockDocumentsByPattern(ctx, dbConfig, collectionName, pattern)
		} else if documentName != "" {
			deleteMockDocumentByName(ctx, dbConfig, collectionName, documentName)
		} else {
			deleteMultipleMockDocuments(ctx, dbConfig, collectionName, documentIDs)
		}
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func runDocumentDeleteAll(cmd *cobra.Command, args []string) {
	collectionName := args[0]

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Delete All Documents")
	fmt.Println()

	printWarning(fmt.Sprintf("⚠️  WARNING: This will permanently delete ALL documents from collection '%s'!", collectionName))
	fmt.Println()

	// First confirmation
	if !confirmAction(fmt.Sprintf("Are you sure you want to delete ALL documents from collection '%s'?", collectionName)) {
		printInfo("Operation cancelled by user")
		return
	}

	// Second confirmation with red warning
	fmt.Println()
	color.New(color.FgRed, color.Bold).Println("🚨 FINAL WARNING: This operation CANNOT be undone!")
	color.New(color.FgRed).Printf("All documents in collection '%s' will be permanently deleted.\n", collectionName)
	fmt.Println()

	// Require exact "yes" confirmation
	fmt.Print("Type 'yes' to confirm deletion: ")
	var response string
	_, _ = fmt.Scanln(&response)

	if response != "yes" {
		printInfo("Operation cancelled - confirmation not received")
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

func showWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
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

	// Show schema if requested
	if showSchema {
		showDocumentSchema(*document, collectionName)
	}

	// Show expanded metadata if requested
	if showMetadata {
		showDocumentMetadata(*document, collectionName)
	}
}

func showMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
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

	// Show schema if requested
	if showSchema {
		showMockDocumentSchema(*document, collectionName)
	}

	// Show expanded metadata if requested
	if showMetadata {
		showMockDocumentMetadata(*document, collectionName)
	}
}

func showWeaviateDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
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

	// Show schema if requested (only for first document to avoid repetition)
	if showSchema && len(documents) > 0 {
		showDocumentSchema(documents[0], collectionName)
	}

	// Show expanded metadata if requested (only for first document to avoid repetition)
	if showMetadata && len(documents) > 0 {
		showDocumentMetadata(documents[0], collectionName)
	}
}

func showMockDocumentsByMetadata(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, metadataFilters []string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
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

	// Show schema if requested (only for first document to avoid repetition)
	if showSchema && len(documents) > 0 {
		showMockDocumentSchema(documents[0], collectionName)
	}

	// Show expanded metadata if requested (only for first document to avoid repetition)
	if showMetadata && len(documents) > 0 {
		showMockDocumentMetadata(documents[0], collectionName)
	}
}

func deleteMultipleWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string) {
	client, err := createWeaviateClient(cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Use bulk delete for better performance when deleting multiple documents
	if len(documentIDs) > 1 {
		fmt.Printf("Deleting %d documents using bulk operation...\n", len(documentIDs))

		successCount, err := client.DeleteDocumentsBulk(ctx, collectionName, documentIDs)
		if err != nil {
			printError(fmt.Sprintf("Bulk delete failed: %v", err))
			printInfo("Falling back to individual deletions...")

			// Fallback to individual deletions
			successCount = 0
			errorCount := 0

			for i, documentID := range documentIDs {
				fmt.Printf("Deleting document %d/%d: %s\n", i+1, len(documentIDs), documentID)

				if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
					printError(fmt.Sprintf("Failed to delete document %s: %v", documentID, err))
					errorCount++
				} else {
					successCount++
				}
			}

			if errorCount == 0 {
				printSuccess(fmt.Sprintf("All %d documents deleted successfully!", successCount))
			} else if successCount == 0 {
				printError(fmt.Sprintf("Failed to delete all %d documents", errorCount))
			} else {
				printWarning(fmt.Sprintf("Deleted %d documents successfully, %d failed", successCount, errorCount))
			}
		} else {
			if successCount == len(documentIDs) {
				printSuccess(fmt.Sprintf("All %d documents deleted successfully using bulk operation!", successCount))
			} else {
				printWarning(fmt.Sprintf("Bulk delete completed: %d/%d documents deleted successfully", successCount, len(documentIDs)))
			}
		}
	} else {
		// Single document deletion
		documentID := documentIDs[0]
		fmt.Printf("Deleting document: %s\n", documentID)

		if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
			printError(fmt.Sprintf("Failed to delete document %s: %v", documentID, err))
		} else {
			printSuccess(fmt.Sprintf("Successfully deleted document: %s", documentID))
		}
	}
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

func deleteMultipleMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string) {
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

	successCount := 0
	errorCount := 0

	for i, documentID := range documentIDs {
		fmt.Printf("Deleting document %d/%d: %s\n", i+1, len(documentIDs), documentID)

		if err := client.DeleteDocument(ctx, collectionName, documentID); err != nil {
			printError(fmt.Sprintf("Failed to delete document %s: %v", documentID, err))
			errorCount++
		} else {
			printSuccess(fmt.Sprintf("Successfully deleted document: %s", documentID))
			successCount++
		}
		fmt.Println()
	}

	// Summary
	if len(documentIDs) > 1 {
		if errorCount == 0 {
			printSuccess(fmt.Sprintf("All %d documents deleted successfully!", successCount))
		} else if successCount == 0 {
			printError(fmt.Sprintf("Failed to delete all %d documents", errorCount))
		} else {
			printWarning(fmt.Sprintf("Deleted %d documents successfully, %d failed", successCount, errorCount))
		}
	}
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
	collectionNames := args

	// Load configuration
	cfg, err := loadConfigWithOverrides()
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

		// Calculate totals
		totalChunks := 0
		totalImages := 0
		documentCount := 0
		imageStackCount := 0

		for _, vdoc := range virtualDocs {
			if vdoc.TotalChunks > 0 {
				totalChunks += len(vdoc.Chunks)
				documentCount++
			} else if isImageVirtualDocument(vdoc) {
				totalImages += len(vdoc.Chunks)
				imageStackCount++
			} else {
				documentCount++
			}
		}

		// Show individual document details
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

		// Show totals
		if len(virtualDocs) > 1 {
			fmt.Println()
			printStyledKeyValueProminentWithEmoji("Totals", "", "📊")
			fmt.Printf("   ")
			printStyledKeyNumberProminentWithEmoji("Documents", documentCount, "📄")
			fmt.Printf(" (")
			printStyledKeyNumberProminentWithEmoji("Total Chunks", totalChunks, "📝")
			fmt.Printf(")")
			fmt.Println()

			if imageStackCount > 0 {
				fmt.Printf("   ")
				printStyledKeyNumberProminentWithEmoji("Image Stacks", imageStackCount, "🗂️")
				fmt.Printf(" (")
				printStyledKeyNumberProminentWithEmoji("Total Images", totalImages, "🖼️")
				fmt.Printf(")")
				fmt.Println()
			}
		}
	}
}

// aggregateDocumentsByOriginal groups documents by their original filename
func aggregateDocumentsByOriginal(documents []weaviate.Document) []VirtualDocument {
	docMap := make(map[string]*VirtualDocument)

	for _, doc := range documents {
		// Check if this is a chunked document
		if metadata, ok := doc.Metadata["metadata"]; ok {
			// Handle both string and map metadata formats
			var metadataObj map[string]interface{}

			if metadataStr, ok := metadata.(string); ok {
				// Parse the JSON metadata to extract original filename
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err != nil {
					continue
				}
			} else if metadataMap, ok := metadata.(map[string]interface{}); ok {
				// Metadata is already a map
				metadataObj = metadataMap

				// Check if there's nested metadata (like in TestCollection)
				if nestedMetadata, ok := metadataObj["metadata"]; ok {
					if nestedStr, ok := nestedMetadata.(string); ok {
						// Parse the nested JSON metadata
						if err := json.Unmarshal([]byte(nestedStr), &metadataObj); err != nil {
							continue
						}
					} else if nestedMap, ok := nestedMetadata.(map[string]interface{}); ok {
						// Use the nested metadata map
						metadataObj = nestedMap
					}
				}
			} else {
				continue
			}

			// Check for chunked document using current metadata structure
			if filename, ok := metadataObj["filename"].(string); ok {
				if isExtracted, ok := metadataObj["is_extracted_from_document"].(bool); ok && isExtracted {
					// This is a chunk from a document
					if vdoc, exists := docMap[filename]; exists {
						vdoc.Chunks = append(vdoc.Chunks, doc)
					} else {
						totalChunks := 0
						if tc, ok := metadataObj["total_chunks"].(float64); ok {
							totalChunks = int(tc)
						}
						docMap[filename] = &VirtualDocument{
							OriginalFilename: filename,
							TotalChunks:      totalChunks,
							Chunks:           []weaviate.Document{doc},
							Metadata:         metadataObj,
						}
					}
					continue
				}
			}

			// Check for chunked document using RagMeDocs metadata structure
			if _, ok := metadataObj["chunk_index"].(float64); ok {
				// This is a chunked document from RagMeDocs
				// Extract original filename from URL
				if url, ok := doc.Metadata["url"].(string); ok {
					// Extract filename from URL like "file:///path/to/file.pdf#chunk-0"
					parts := strings.Split(url, "#")
					if len(parts) > 0 {
						originalFilename := strings.TrimPrefix(parts[0], "file://")
						originalFilename = filepath.Base(originalFilename)

						if vdoc, exists := docMap[originalFilename]; exists {
							vdoc.Chunks = append(vdoc.Chunks, doc)
						} else {
							totalChunks := 0
							if chunkSizes, ok := metadataObj["chunk_sizes"].([]interface{}); ok {
								totalChunks = len(chunkSizes)
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

		// Check if this is an image document that should be grouped with its source
		if isImageDocumentFromMetadata(doc.Metadata) {
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
		} else {
			// For standalone documents, use URL or filename as key
			groupKey := getStandaloneDocumentKey(doc)

			if vdoc, exists := docMap[groupKey]; exists {
				// Add to existing group
				vdoc.Chunks = append(vdoc.Chunks, doc)
			} else {
				// Create new group
				docMap[groupKey] = &VirtualDocument{
					OriginalFilename: groupKey,
					TotalChunks:      0,
					Chunks:           []weaviate.Document{doc},
					Metadata:         doc.Metadata,
				}
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

// isImageDocumentFromMetadata checks if a document is an image based on metadata
func isImageDocumentFromMetadata(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}

	// Check metadata for image indicators
	if metadataField, ok := metadata["metadata"]; ok {
		if metadataStr, ok := metadataField.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				// Check for image content type
				if contentType, ok := metadataObj["content_type"].(string); ok {
					if contentType == "image" {
						return true
					}
				}
				// Check for source type indicating PDF extracted image
				if sourceType, ok := metadataObj["source_type"].(string); ok {
					if sourceType == "pdf_extracted_image" {
						return true
					}
				}
			}
		}
	}

	// Check URL for image indicators
	if url, ok := metadata["url"].(string); ok {
		if strings.HasSuffix(strings.ToLower(url), ".png") ||
			strings.HasSuffix(strings.ToLower(url), ".jpg") ||
			strings.HasSuffix(strings.ToLower(url), ".jpeg") ||
			strings.HasSuffix(strings.ToLower(url), ".gif") ||
			strings.HasSuffix(strings.ToLower(url), ".bmp") {
			return true
		}
	}

	// Check filename for image indicators
	if filename, ok := metadata["filename"].(string); ok {
		if strings.HasSuffix(strings.ToLower(filename), ".png") ||
			strings.HasSuffix(strings.ToLower(filename), ".jpg") ||
			strings.HasSuffix(strings.ToLower(filename), ".jpeg") ||
			strings.HasSuffix(strings.ToLower(filename), ".gif") ||
			strings.HasSuffix(strings.ToLower(filename), ".bmp") {
			return true
		}
	}

	return false
}

// getStandaloneDocumentKey determines the grouping key for standalone documents
func getStandaloneDocumentKey(doc weaviate.Document) string {
	// Use URL as primary key
	if url, ok := doc.Metadata["url"].(string); ok {
		return url
	}

	// Use filename as fallback
	if filename, ok := doc.Metadata["filename"].(string); ok {
		return filename
	}

	// Fallback
	return "Unknown Document"
}

// getImageGroupKey determines the grouping key for an image document
func getImageGroupKey(doc weaviate.Document) string {
	// Check if this is an image extracted from a PDF
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				// Check for PDF filename (RagMeDocs format)
				if pdfFilename, ok := metadataObj["pdf_filename"].(string); ok {
					return pdfFilename
				}
				// Check for source document (RagMeDocs format)
				if sourceDoc, ok := metadataObj["source_document"].(string); ok {
					return sourceDoc
				}
				// Check for source type to confirm it's from PDF
				if sourceType, ok := metadataObj["source_type"].(string); ok {
					if sourceType == "pdf_extracted_image" {
						// Try to extract PDF name from filename
						if filename, ok := metadataObj["filename"].(string); ok {
							// For tmp files like "tmp00cnn9fb_p23_i0.png", we need to find the source PDF
							// This should be handled by pdf_filename or source_document above
							return filename
						}
					}
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

		// Calculate totals
		totalChunks := 0
		totalImages := 0
		documentCount := 0
		imageStackCount := 0

		for _, vdoc := range virtualDocs {
			if vdoc.TotalChunks > 0 {
				totalChunks += len(vdoc.Chunks)
				documentCount++
			} else if isMockImageVirtualDocument(vdoc) {
				totalImages += len(vdoc.Chunks)
				imageStackCount++
			} else {
				documentCount++
			}
		}

		// Show individual document details
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

		// Show totals
		if len(virtualDocs) > 1 {
			fmt.Println()
			printStyledKeyValueProminentWithEmoji("Totals", "", "📊")
			fmt.Printf("   ")
			printStyledKeyNumberProminentWithEmoji("Documents", documentCount, "📄")
			fmt.Printf(" (")
			printStyledKeyNumberProminentWithEmoji("Total Chunks", totalChunks, "📝")
			fmt.Printf(")")
			fmt.Println()

			if imageStackCount > 0 {
				fmt.Printf("   ")
				printStyledKeyNumberProminentWithEmoji("Image Stacks", imageStackCount, "🗂️")
				fmt.Printf(" (")
				printStyledKeyNumberProminentWithEmoji("Total Images", totalImages, "🖼️")
				fmt.Printf(")")
				fmt.Println()
			}
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

func runDocumentCreate(cmd *cobra.Command, args []string) {
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")

	collectionName := args[0]
	filePath := args[1]

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		printError(fmt.Sprintf("File not found: %s", filePath))
		os.Exit(1)
	}

	// Get database configuration
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get database configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Create Document")
	fmt.Println()

	// Process the file based on its type
	documents, err := processFile(filePath, chunkSize)
	if err != nil {
		printError(fmt.Sprintf("Failed to process file: %v", err))
		os.Exit(1)
	}

	// Create documents in the database
	ctx := context.Background()
	successCount := 0
	errorCount := 0

	for i, doc := range documents {
		printInfo(fmt.Sprintf("Creating document %d/%d: %s", i+1, len(documents), doc.ID))

		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			if err := createWeaviateDocument(ctx, dbConfig, collectionName, doc); err != nil {
				errorMsg := formatDocumentCreationError(doc.ID, err)
				printError(errorMsg)
				errorCount++
			} else {
				printSuccess(fmt.Sprintf("Successfully created document: %s", doc.ID))
				successCount++
			}
		case config.VectorDBTypeMock:
			if err := createMockDocument(ctx, dbConfig, collectionName, doc); err != nil {
				printError(fmt.Sprintf("Failed to create document '%s': %v", doc.ID, err))
				errorCount++
			} else {
				printSuccess(fmt.Sprintf("Successfully created document: %s", doc.ID))
				successCount++
			}
		default:
			printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			errorCount++
		}
	}

	// Summary
	if len(documents) > 1 {
		if errorCount == 0 {
			printSuccess(fmt.Sprintf("All %d documents created successfully!", successCount))
		} else if successCount == 0 {
			printError(fmt.Sprintf("Failed to create all %d documents", errorCount))
		} else {
			printWarning(fmt.Sprintf("Created %d documents successfully, %d failed", successCount, errorCount))
		}
	}
}

// DocumentData represents the data structure for creating a document
type DocumentData struct {
	ID        string
	Content   string
	Image     string
	ImageData string
	URL       string
	Metadata  map[string]interface{}
}

// processFile processes a file and returns document data
func processFile(filePath string, chunkSize int) ([]DocumentData, error) {
	// Determine file type
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return processPDFFile(filePath, chunkSize)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return processImageFile(filePath)
	default:
		return processTextFile(filePath, chunkSize)
	}
}

// processTextFile processes a text file and chunks it
func processTextFile(filePath string, chunkSize int) ([]DocumentData, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)
	if len(text) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// Chunk the text
	chunks := chunkText(text, chunkSize)

	// Generate documents
	var documents []DocumentData
	for i, chunk := range chunks {
		docID := generateDocumentID()

		// Generate metadata
		metadata := generateTextMetadata(filePath, i, len(chunks), len(chunk))

		doc := DocumentData{
			ID:       docID,
			Content:  chunk,
			URL:      fmt.Sprintf("file://%s#chunk-%d", filePath, i),
			Metadata: metadata,
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

// processImageFile processes an image file
func processImageFile(filePath string) ([]DocumentData, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(content)

	// Determine MIME type
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "image/jpeg" // default
	}

	// Generate document
	docID := generateDocumentID()
	metadata := generateImageMetadata(filePath, len(content))

	doc := DocumentData{
		ID:        docID,
		Image:     fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data),
		ImageData: base64Data,
		URL:       fmt.Sprintf("file://%s", filePath),
		Metadata:  metadata,
	}

	return []DocumentData{doc}, nil
}

// processPDFFile processes a PDF file (placeholder - will implement with pdfcpu)
func processPDFFile(filePath string, chunkSize int) ([]DocumentData, error) {
	// For now, return an error indicating PDF support is not yet implemented
	return nil, fmt.Errorf("PDF processing not yet implemented - will add pdfcpu support")
}

// chunkText splits text into chunks of specified size
func chunkText(text string, chunkSize int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + chunkSize

		// Ensure we don't go beyond the text length
		if end > len(text) {
			end = len(text)
		}

		// Try to break at word boundary
		if end < len(text) {
			// Look for the last space within the chunk
			for i := end; i > start; i-- {
				if text[i] == ' ' || text[i] == '\n' {
					end = i
					break
				}
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		start = end
		if start < len(text) && text[start] == ' ' {
			start++ // Skip the space
		}
	}

	return chunks
}

// generateDocumentID generates a unique document ID
func generateDocumentID() string {
	// Simple UUID-like ID generation using timestamp and random components
	now := time.Now().UnixNano()
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		now&0xffffffff,
		(now>>32)&0xffff,
		(now>>48)&0xffff,
		(now>>16)&0xffff,
		(now>>0)&0xffffffff)
}

// generateTextMetadata generates metadata for text documents
func generateTextMetadata(filePath string, chunkIndex, totalChunks int, chunkSize int) map[string]interface{} {
	fileName := filepath.Base(filePath)
	fileSize := int64(0)

	if stat, err := os.Stat(filePath); err == nil {
		fileSize = stat.Size()
	}

	metadata := map[string]interface{}{
		"filename":                   fileName,
		"file_size":                  float64(fileSize),
		"content_type":               "text",
		"date_added":                 time.Now().Format(time.RFC3339),
		"chunk_index":                chunkIndex,
		"chunk_size":                 chunkSize,
		"total_chunks":               totalChunks,
		"source_document":            fileName,
		"processed_by":               "weave-cli document create",
		"processing_time":            time.Now().Unix(),
		"is_extracted_from_document": true,
	}

	return metadata
}

// generateImageMetadata generates metadata for image documents
func generateImageMetadata(filePath string, fileSize int) map[string]interface{} {
	fileName := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	metadata := map[string]interface{}{
		"filename":                   fileName,
		"file_size":                  float64(fileSize),
		"content_type":               "image",
		"date_added":                 time.Now().Format(time.RFC3339),
		"source_document":            fileName,
		"processed_by":               "weave-cli document create",
		"processing_time":            time.Now().Unix(),
		"is_extracted_from_document": true,
		"file_extension":             ext,
	}

	return metadata
}

// createWeaviateDocument creates a document in Weaviate
func createWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, doc DocumentData) error {
	// Convert VectorDBConfig to weaviate.Config
	weaviateConfig := &weaviate.Config{
		URL:          cfg.URL,
		APIKey:       cfg.APIKey,
		OpenAIAPIKey: cfg.OpenAIAPIKey,
	}

	// Create Weaviate client
	client, err := weaviate.NewWeaveClient(weaviateConfig)
	if err != nil {
		return fmt.Errorf("failed to create weaviate client: %w", err)
	}

	// Create document object
	document := weaviate.Document{
		ID:        doc.ID,
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
	}

	// Create the document
	return client.CreateDocument(ctx, collectionName, document)
}

// createMockDocument creates a document in mock database
func createMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, doc DocumentData) error {
	// Convert VectorDBConfig to mock.Config
	mockConfig := &config.MockConfig{
		Collections: []config.MockCollection{},
	}

	// Create mock client
	client := mock.NewClient(mockConfig)

	// Create document object
	document := mock.Document{
		ID:        doc.ID,
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
	}

	// Create the document
	return client.CreateDocument(ctx, collectionName, document)
}

// formatDocumentCreationError formats document creation errors with user-friendly messages
func formatDocumentCreationError(docID string, err error) string {
	errorStr := err.Error()

	// Check for OpenAI API key missing error
	if strings.Contains(errorStr, "no api key found") && (strings.Contains(errorStr, "OPENAI_APIKEY") || strings.Contains(errorStr, "OPENAI_API_KEY")) {
		return fmt.Sprintf("Failed to create document '%s': Missing OpenAI API key for vectorization.\n"+
			"   💡 Solution: Add OPENAI_API_KEY to your .env file or use mock database (VECTOR_DB_TYPE=mock)", docID)
	}

	// Check for other common vectorization errors
	if strings.Contains(errorStr, "vectorize target vector") {
		return fmt.Sprintf("Failed to create document '%s': Vectorization error.\n"+
			"   💡 Check your embedding model configuration and API keys", docID)
	}

	// Check for authentication errors
	if strings.Contains(errorStr, "unauthorized") || strings.Contains(errorStr, "401") {
		return fmt.Sprintf("Failed to create document '%s': Authentication failed.\n"+
			"   💡 Check your WEAVIATE_API_KEY in .env file", docID)
	}

	// Check for connection errors
	if strings.Contains(errorStr, "connection") || strings.Contains(errorStr, "timeout") {
		return fmt.Sprintf("Failed to create document '%s': Connection error.\n"+
			"   💡 Check your WEAVIATE_URL and network connection", docID)
	}

	// Default error message
	return fmt.Sprintf("Failed to create document '%s': %v", docID, err)
}

// findDocumentsByPattern finds documents matching a pattern (auto-detects glob vs regex)
func findDocumentsByPattern(cfg *config.Config, dbConfig *config.VectorDBConfig, collectionName, pattern string, forceRegex bool) ([]weaviate.Document, error) {
	ctx := context.Background()

	// Auto-detect pattern type if not forced to regex
	var regex *regexp.Regexp
	var err error

	if forceRegex || isRegexPattern(pattern) {
		// Compile as regex pattern
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
	} else {
		// Convert glob pattern to regex
		regexPattern := globToRegex(pattern)
		regex, err = regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %v", err)
		}
	}

	// Create client based on database type
	var client weaviate.WeaveClient
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		weaviateClient, err := createWeaviateClient(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Weaviate client: %v", err)
		}
		client = *weaviateClient
	case config.VectorDBTypeMock:
		return nil, fmt.Errorf("pattern deletion not yet supported for mock database")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	// Get all documents from the collection
	allDocs, err := client.ListDocuments(ctx, collectionName, 10000) // Large limit to get all docs
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %v", err)
	}

	// Filter documents that match the pattern
	var matchingDocs []weaviate.Document
	for _, doc := range allDocs {
		if matchesPattern(doc, regex) {
			matchingDocs = append(matchingDocs, doc)
		}
	}

	return matchingDocs, nil
}

// isRegexPattern detects if a pattern looks like regex rather than glob
func isRegexPattern(pattern string) bool {
	// Common regex indicators
	regexIndicators := []string{
		"^", "$", "\\", ".*", ".+", ".?", "[", "]", "(", ")", "{", "}", "|", "+", "?",
	}

	// If pattern contains regex-specific characters, treat as regex
	for _, indicator := range regexIndicators {
		if strings.Contains(pattern, indicator) {
			return true
		}
	}

	// If pattern contains only glob characters (*, ?, []) and no regex chars, treat as glob
	// This is a simple heuristic - could be improved
	return false
}

// globToRegex converts a shell glob pattern to regex
func globToRegex(glob string) string {
	// Escape special regex characters first
	result := regexp.QuoteMeta(glob)

	// Convert glob patterns to regex equivalents
	result = strings.ReplaceAll(result, "\\*", ".*") // * -> .*
	result = strings.ReplaceAll(result, "\\?", ".")  // ? -> .
	result = strings.ReplaceAll(result, "\\[", "[")  // [ -> [
	result = strings.ReplaceAll(result, "\\]", "]")  // ] -> ]

	// Add anchors for exact matching (optional - could be made configurable)
	// result = "^" + result + "$"

	return result
}

// matchesPattern checks if a document matches the regex pattern
func matchesPattern(doc weaviate.Document, regex *regexp.Regexp) bool {
	// Check filename field directly (for image documents)
	if filename, ok := doc.Metadata["filename"].(string); ok {
		if regex.MatchString(filename) {
			return true
		}
	}

	// Check filename in nested metadata (for text documents)
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if filename, ok := metadataObj["filename"].(string); ok {
					if regex.MatchString(filename) {
						return true
					}
				}
			}
		}
	}

	// Check URL field
	if url, ok := doc.Metadata["url"].(string); ok {
		if regex.MatchString(url) {
			return true
		}
	}

	return false
}

// getDocumentDisplayName returns a display name for a document
func getDocumentDisplayName(doc weaviate.Document) string {
	// Try to get filename from metadata
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if filename, ok := metadataObj["filename"].(string); ok {
					return filename
				}
			}
		}
	}

	// Try URL
	if url, ok := doc.Metadata["url"].(string); ok {
		return url
	}

	// Try filename field directly
	if filename, ok := doc.Metadata["filename"].(string); ok {
		return filename
	}

	// Fallback to ID
	return doc.ID
}

// deleteWeaviateDocumentsByPattern deletes documents matching a pattern from Weaviate
func deleteWeaviateDocumentsByPattern(ctx context.Context, dbConfig *config.VectorDBConfig, collectionName, pattern string) {
	// Find matching documents
	matchingDocs, err := findDocumentsByPattern(nil, dbConfig, collectionName, pattern, false)
	if err != nil {
		printError(fmt.Sprintf("Failed to find documents: %v", err))
		return
	}

	if len(matchingDocs) == 0 {
		printInfo("No documents found matching pattern")
		return
	}

	// Extract document IDs
	var docIDs []string
	for _, doc := range matchingDocs {
		docIDs = append(docIDs, doc.ID)
	}

	// Delete the documents
	deleteMultipleWeaviateDocuments(ctx, dbConfig, collectionName, docIDs)
}

// deleteMockDocumentsByPattern deletes documents matching a pattern from mock database
func deleteMockDocumentsByPattern(ctx context.Context, dbConfig *config.VectorDBConfig, collectionName, pattern string) {
	printError("Pattern deletion not yet supported for mock database")
}

// Helper functions for document operations by name

// DocumentWithCollection represents a document and the collection it was found in
type DocumentWithCollection struct {
	Document   *weaviate.Document
	Collection string
}

// findDocumentByName finds a document by its filename/name
func findDocumentByName(cfg *config.Config, dbConfig *config.VectorDBConfig, collectionName, documentName string) (*DocumentWithCollection, error) {
	ctx := context.Background()

	client, err := createWeaviateClient(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// First, try the specified collection
	doc, err := searchInCollection(client, ctx, collectionName, documentName)
	if err == nil {
		return &DocumentWithCollection{Document: doc, Collection: collectionName}, nil
	}

	// If not found, try related collections
	relatedCollections := getRelatedCollections(collectionName)
	for _, relatedCollection := range relatedCollections {
		doc, err := searchInCollection(client, ctx, relatedCollection, documentName)
		if err == nil {
			return &DocumentWithCollection{Document: doc, Collection: relatedCollection}, nil
		}
	}

	return nil, fmt.Errorf("document with name '%s' not found in collection '%s' or related collections", documentName, collectionName)
}

// searchInCollection searches for a document in a specific collection
func searchInCollection(client *weaviate.WeaveClient, ctx context.Context, collectionName, documentName string) (*weaviate.Document, error) {
	// Get all documents from the collection - use a higher limit to ensure we get all documents
	documents, err := client.ListDocuments(ctx, collectionName, 10000) // Get up to 10000 documents
	if err != nil {
		return nil, fmt.Errorf("failed to list documents from collection '%s': %v", collectionName, err)
	}

	// Search for document with matching name
	for _, doc := range documents {
		if documentNameMatches(doc, documentName) {
			return &doc, nil
		}
	}

	return nil, fmt.Errorf("document with name '%s' not found in collection '%s'", documentName, collectionName)
}

// getRelatedCollections returns related collection names based on common patterns
func getRelatedCollections(collectionName string) []string {
	var related []string

	// Common patterns for related collections
	patterns := map[string][]string{
		"Docs":   {"Images", "Images_test"},
		"Images": {"Docs", "Docs_test"},
		"_test":  {"_test"},
	}

	// Check for patterns and add related collections
	for pattern, suffixes := range patterns {
		if strings.Contains(collectionName, pattern) {
			baseName := strings.ReplaceAll(collectionName, pattern, "")
			for _, suffix := range suffixes {
				relatedCollection := baseName + suffix
				if relatedCollection != collectionName {
					related = append(related, relatedCollection)
				}
			}
		}
	}

	// Special cases for known collection pairs
	specialCases := map[string][]string{
		"RagMeDocs":       {"RagMeImages"},
		"RagMeImages":     {"RagMeDocs"},
		"NotesMaxDocs":    {"NotesMaxDocs_test"},
		"CalendarMaxDocs": {"CalendarMaxDocs_test"},
		"TODOsMaxDocs":    {"TODOsMaxDocs_test"},
		"WeaveDocs":       {"WeaveImages"},
		"WeaveImages":     {"WeaveDocs"},
	}

	if relatedCols, exists := specialCases[collectionName]; exists {
		related = append(related, relatedCols...)
	}

	return related
}

// documentNameMatches checks if a document matches the given name
func documentNameMatches(doc weaviate.Document, name string) bool {
	// Check URL field (common for documents)
	if url, ok := doc.Metadata["url"].(string); ok {
		if url == name || filepath.Base(url) == name {
			return true
		}
		// Check if URL contains the name (for PDF image URLs like "pdf://ragme-io.pdf/page_1/image_1")
		if strings.Contains(url, name) {
			return true
		}
	}

	// Check filename field directly
	if filename, ok := doc.Metadata["filename"].(string); ok {
		if filename == name || filepath.Base(filename) == name {
			return true
		}
	}

	// Check nested metadata for filename (common in text documents)
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if filename, ok := metadataObj["filename"].(string); ok {
					if filename == name || filepath.Base(filename) == name {
						return true
					}
				}
				if originalFilename, ok := metadataObj["original_filename"].(string); ok {
					if originalFilename == name || filepath.Base(originalFilename) == name {
						return true
					}
				}
			}
		}
	}

	return false
}

// showWeaviateDocumentByName shows a Weaviate document by its name
func showWeaviateDocumentByName(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentName string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
	docWithCollection, err := findDocumentByName(nil, cfg, collectionName, documentName)
	if err != nil {
		printError(fmt.Sprintf("Failed to find document: %v", err))
		return
	}

	doc := docWithCollection.Document
	actualCollection := docWithCollection.Collection

	// Display document details
	color.New(color.FgGreen).Printf("Document ID: %s\n", doc.ID)
	fmt.Printf("Collection: %s\n", actualCollection)
	if actualCollection != collectionName {
		fmt.Printf("Note: Document found in '%s' (searched in '%s')\n", actualCollection, collectionName)
	}
	fmt.Printf("Name: %s\n", documentName)
	fmt.Println()

	fmt.Printf("Content:\n")
	if showLong {
		fmt.Printf("%s\n", doc.Content)
	} else {
		// Use shortLines to limit content by lines instead of characters
		preview := truncateStringByLines(doc.Content, shortLines)
		fmt.Printf("%s\n", preview)
	}
	fmt.Println()

	if len(doc.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for key, value := range doc.Metadata {
			// Truncate value based on shortLines directive
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := smartTruncate(valueStr, key, shortLines)
			fmt.Printf("  %s: %s\n", key, truncatedValue)
		}
		fmt.Println()
	}

	// Show schema if requested
	if showSchema {
		showDocumentSchema(*doc, actualCollection)
	}

	// Show expanded metadata if requested
	if showMetadata {
		showDocumentMetadata(*doc, actualCollection)
	}
}

// showMockDocumentByName shows a mock document by its name
func showMockDocumentByName(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentName string, showLong bool, shortLines int, showSchema bool, showMetadata bool) {
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

	// Get all documents from the collection
	documents, err := client.ListDocuments(ctx, collectionName, 1000)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	// Search for document with matching name
	var foundDoc *mock.Document
	for _, doc := range documents {
		if mockDocumentNameMatches(doc, documentName) {
			foundDoc = &doc
			break
		}
	}

	if foundDoc == nil {
		printError(fmt.Sprintf("Document with name '%s' not found", documentName))
		return
	}

	// Display document details
	color.New(color.FgGreen).Printf("Document ID: %s\n", foundDoc.ID)
	fmt.Printf("Collection: %s\n", collectionName)
	fmt.Printf("Name: %s\n", documentName)
	fmt.Println()

	fmt.Printf("Content:\n")
	if showLong {
		fmt.Printf("%s\n", foundDoc.Content)
	} else {
		// Use shortLines to limit content by lines instead of characters
		preview := truncateStringByLines(foundDoc.Content, shortLines)
		fmt.Printf("%s\n", preview)
	}
	fmt.Println()

	if len(foundDoc.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for key, value := range foundDoc.Metadata {
			// Truncate value based on shortLines directive
			valueStr := fmt.Sprintf("%v", value)
			truncatedValue := smartTruncate(valueStr, key, shortLines)
			fmt.Printf("  %s: %s\n", key, truncatedValue)
		}
		fmt.Println()
	}

	// Show schema if requested
	if showSchema {
		showMockDocumentSchema(*foundDoc, collectionName)
	}

	// Show expanded metadata if requested
	if showMetadata {
		showMockDocumentMetadata(*foundDoc, collectionName)
	}
}

// mockDocumentNameMatches checks if a mock document matches the given name
func mockDocumentNameMatches(doc mock.Document, name string) bool {
	// Check URL field (common for documents)
	if url, ok := doc.Metadata["url"].(string); ok {
		if url == name || filepath.Base(url) == name {
			return true
		}
		// Check if URL contains the name (for PDF image URLs like "pdf://ragme-io.pdf/page_1/image_1")
		if strings.Contains(url, name) {
			return true
		}
	}

	// Check filename field directly
	if filename, ok := doc.Metadata["filename"].(string); ok {
		if filename == name || filepath.Base(filename) == name {
			return true
		}
	}

	// Check nested metadata for filename (common in text documents)
	if metadata, ok := doc.Metadata["metadata"]; ok {
		if metadataStr, ok := metadata.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if filename, ok := metadataObj["filename"].(string); ok {
					if filename == name || filepath.Base(filename) == name {
						return true
					}
				}
				if originalFilename, ok := metadataObj["original_filename"].(string); ok {
					if originalFilename == name || filepath.Base(originalFilename) == name {
						return true
					}
				}
			}
		}
	}

	return false
}

// deleteWeaviateDocumentByName deletes a Weaviate document by its name
func deleteWeaviateDocumentByName(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentName string) {
	// Find all documents that match the name (for virtual documents like image stacks)
	documentsWithCollections, err := findAllDocumentsByName(nil, cfg, collectionName, documentName)
	if err != nil {
		printError(fmt.Sprintf("Failed to find documents: %v", err))
		return
	}

	if len(documentsWithCollections) == 0 {
		printError(fmt.Sprintf("No documents found with name '%s'", documentName))
		return
	}

	// Group documents by collection for efficient deletion
	collectionGroups := make(map[string][]string)
	for _, docWithCollection := range documentsWithCollections {
		collectionGroups[docWithCollection.Collection] = append(collectionGroups[docWithCollection.Collection], docWithCollection.Document.ID)
	}

	// Show detailed info about what was found
	if len(documentsWithCollections) == 1 {
		doc := documentsWithCollections[0]
		if doc.Collection != collectionName {
			printInfo(fmt.Sprintf("Document found in collection '%s' (searched in '%s')", doc.Collection, collectionName))
		}
	} else {
		printInfo(fmt.Sprintf("Found %d documents with name '%s' across %d collections", len(documentsWithCollections), documentName, len(collectionGroups)))

		// Show detailed breakdown
		for collection, docIDs := range collectionGroups {
			fmt.Printf("  - %s: %d documents", collection, len(docIDs))

			// Determine document type for better user understanding
			if strings.Contains(collection, "Images") || strings.Contains(collection, "Image") {
				fmt.Printf(" (images in stack)")
			} else if len(docIDs) > 1 {
				fmt.Printf(" (chunks)")
			}
			fmt.Println()
		}

		// Show warning about complete deletion
		fmt.Println()
		color.New(color.FgYellow).Printf("⚠️  WARNING: This will delete ALL related documents including:\n")
		for collection, docIDs := range collectionGroups {
			if strings.Contains(collection, "Images") || strings.Contains(collection, "Image") {
				fmt.Printf("   • %d images from the image stack in %s\n", len(docIDs), collection)
			} else if len(docIDs) > 1 {
				fmt.Printf("   • %d text chunks in %s\n", len(docIDs), collection)
			} else {
				fmt.Printf("   • 1 document in %s\n", collection)
			}
		}
		fmt.Println()
	}

	// Delete documents from each collection with individual progress
	totalDeleted := 0
	for collection, docIDs := range collectionGroups {
		if len(docIDs) > 1 {
			printInfo(fmt.Sprintf("Deleting %d documents from collection '%s'...", len(docIDs), collection))
		} else {
			printInfo(fmt.Sprintf("Deleting document from collection '%s'...", collection))
		}

		// Show individual deletion progress for better user experience
		if len(docIDs) > 1 {
			fmt.Printf("  Deleting documents: ")
			for i, docID := range docIDs {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", docID[:8]+"...")
			}
			fmt.Println()
		}

		deleteMultipleWeaviateDocuments(ctx, cfg, collection, docIDs)
		totalDeleted += len(docIDs)
	}

	if totalDeleted > 1 {
		printSuccess(fmt.Sprintf("Successfully deleted %d documents with name '%s'", totalDeleted, documentName))
	} else {
		printSuccess(fmt.Sprintf("Successfully deleted document '%s'", documentName))
	}
}

// findAllDocumentsByName finds all documents that match the given name across related collections
func findAllDocumentsByName(cfg *config.Config, dbConfig *config.VectorDBConfig, collectionName, documentName string) ([]*DocumentWithCollection, error) {
	ctx := context.Background()

	client, err := createWeaviateClient(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	var allMatches []*DocumentWithCollection

	// Search in the specified collection first
	matches, err := findAllInCollection(client, ctx, collectionName, documentName)
	if err == nil {
		for _, doc := range matches {
			allMatches = append(allMatches, &DocumentWithCollection{Document: doc, Collection: collectionName})
		}
	}

	// Search in related collections
	relatedCollections := getRelatedCollections(collectionName)
	for _, relatedCollection := range relatedCollections {
		matches, err := findAllInCollection(client, ctx, relatedCollection, documentName)
		if err == nil {
			for _, doc := range matches {
				allMatches = append(allMatches, &DocumentWithCollection{Document: doc, Collection: relatedCollection})
			}
		}
	}

	if len(allMatches) == 0 {
		return nil, fmt.Errorf("no documents found with name '%s' in collection '%s' or related collections", documentName, collectionName)
	}

	return allMatches, nil
}

// findAllInCollection finds all documents in a specific collection that match the given name
func findAllInCollection(client *weaviate.WeaveClient, ctx context.Context, collectionName, documentName string) ([]*weaviate.Document, error) {
	// Get all documents from the collection
	documents, err := client.ListDocuments(ctx, collectionName, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents from collection '%s': %v", collectionName, err)
	}

	var matches []*weaviate.Document
	for _, doc := range documents {
		if documentNameMatches(doc, documentName) {
			matches = append(matches, &doc)
		}
	}

	return matches, nil
}

// deleteMockDocumentByName deletes a mock document by its name
func deleteMockDocumentByName(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentName string) {
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

	// Get all documents from the collection
	documents, err := client.ListDocuments(ctx, collectionName, 1000)
	if err != nil {
		printError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	// Find all documents with matching name
	var docIDs []string
	for _, doc := range documents {
		if mockDocumentNameMatches(doc, documentName) {
			docIDs = append(docIDs, doc.ID)
		}
	}

	if len(docIDs) == 0 {
		printError(fmt.Sprintf("Document with name '%s' not found", documentName))
		return
	}

	if len(docIDs) > 1 {
		printInfo(fmt.Sprintf("Found %d documents with name '%s'", len(docIDs), documentName))

		// Show detailed breakdown for mock documents
		if len(docIDs) > 1 {
			fmt.Printf("  - %s: %d documents", collectionName, len(docIDs))
			if strings.Contains(collectionName, "Images") || strings.Contains(collectionName, "Image") {
				fmt.Printf(" (images in stack)")
			} else {
				fmt.Printf(" (chunks)")
			}
			fmt.Println()

			// Show warning about complete deletion
			fmt.Println()
			color.New(color.FgYellow).Printf("⚠️  WARNING: This will delete ALL related documents including:\n")
			if strings.Contains(collectionName, "Images") || strings.Contains(collectionName, "Image") {
				fmt.Printf("   • %d images from the image stack in %s\n", len(docIDs), collectionName)
			} else {
				fmt.Printf("   • %d text chunks in %s\n", len(docIDs), collectionName)
			}
			fmt.Println()
		}
	}

	// Delete all matching documents
	deleteMultipleMockDocuments(ctx, cfg, collectionName, docIDs)
}

// showDocumentSchema shows the schema of a Weaviate document
func showDocumentSchema(doc weaviate.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("📋 Document Schema: %s\n", collectionName)
	fmt.Println()

	// Document structure
	printStyledEmoji("🏗️")
	fmt.Printf(" ")
	printStyledKeyProminent("Document Structure")
	fmt.Println()

	fmt.Printf("  • ")
	printStyledKeyProminent("id")
	fmt.Printf(" (")
	printStyledValueDimmed("string")
	fmt.Printf(") - Unique document identifier")
	fmt.Println()

	fmt.Printf("  • ")
	printStyledKeyProminent("content")
	fmt.Printf(" (")
	printStyledValueDimmed("text")
	fmt.Printf(") - Document text content")
	fmt.Println()

	// Metadata fields
	if len(doc.Metadata) > 0 {
		fmt.Printf("  • ")
		printStyledKeyProminent("metadata")
		fmt.Printf(" (")
		printStyledValueDimmed("object")
		fmt.Printf(") - Document metadata")
		fmt.Println()

		// Show metadata field types
		printStyledEmoji("📊")
		fmt.Printf(" ")
		printStyledKeyProminent("Metadata Fields")
		fmt.Println()

		for key, value := range doc.Metadata {
			fmt.Printf("    • ")
			printStyledKeyProminent(key)
			fmt.Printf(" (")
			printStyledValueDimmed(getValueType(value))
			fmt.Printf(")")
			fmt.Println()
		}
	} else {
		fmt.Printf("  • ")
		printStyledKeyProminent("metadata")
		fmt.Printf(" (")
		printStyledValueDimmed("object")
		fmt.Printf(") - No metadata fields")
		fmt.Println()
	}

	fmt.Println()
}

// showMockDocumentSchema shows the schema of a mock document
func showMockDocumentSchema(doc mock.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("📋 Document Schema: %s\n", collectionName)
	fmt.Println()

	// Document structure
	printStyledEmoji("🏗️")
	fmt.Printf(" ")
	printStyledKeyProminent("Document Structure")
	fmt.Println()

	fmt.Printf("  • ")
	printStyledKeyProminent("id")
	fmt.Printf(" (")
	printStyledValueDimmed("string")
	fmt.Printf(") - Unique document identifier")
	fmt.Println()

	fmt.Printf("  • ")
	printStyledKeyProminent("content")
	fmt.Printf(" (")
	printStyledValueDimmed("text")
	fmt.Printf(") - Document text content")
	fmt.Println()

	// Metadata fields
	if len(doc.Metadata) > 0 {
		fmt.Printf("  • ")
		printStyledKeyProminent("metadata")
		fmt.Printf(" (")
		printStyledValueDimmed("object")
		fmt.Printf(") - Document metadata")
		fmt.Println()

		// Show metadata field types
		printStyledEmoji("📊")
		fmt.Printf(" ")
		printStyledKeyProminent("Metadata Fields")
		fmt.Println()

		for key, value := range doc.Metadata {
			fmt.Printf("    • ")
			printStyledKeyProminent(key)
			fmt.Printf(" (")
			printStyledValueDimmed(getValueType(value))
			fmt.Printf(")")
			fmt.Println()
		}
	} else {
		fmt.Printf("  • ")
		printStyledKeyProminent("metadata")
		fmt.Printf(" (")
		printStyledValueDimmed("object")
		fmt.Printf(") - No metadata fields")
		fmt.Println()
	}

	fmt.Println()

	// Mock-specific information
	printStyledEmoji("🎭")
	fmt.Printf(" ")
	printStyledKeyProminent("Mock Document Info")
	fmt.Println()
	fmt.Printf("  • Type: ")
	printStyledValueDimmed("Mock Document")
	fmt.Println()
	fmt.Printf("  • Collection: ")
	printStyledValueDimmed(collectionName)
	fmt.Println()
	fmt.Println()
}

// showDocumentMetadata shows expanded metadata for a Weaviate document
func showDocumentMetadata(doc weaviate.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("📊 Document Metadata: %s\n", collectionName)
	fmt.Println()

	if len(doc.Metadata) == 0 {
		printStyledValueDimmed("No metadata available for this document")
		fmt.Println()
		return
	}

	// Display metadata analysis
	printStyledEmoji("📈")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Analysis")
	fmt.Println()

	fmt.Printf("  • Total Metadata Fields: ")
	printStyledValueDimmed(fmt.Sprintf("%d", len(doc.Metadata)))
	fmt.Println()

	fmt.Printf("  • Document ID: ")
	printStyledValueDimmed(doc.ID)
	fmt.Println()

	fmt.Println()

	// Show detailed metadata fields
	printStyledEmoji("🔍")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Fields")
	fmt.Println()

	// Sort fields alphabetically for consistent display
	var sortedKeys []string
	for key := range doc.Metadata {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := doc.Metadata[key]

		fmt.Printf("  • ")
		printStyledKeyProminent(key)
		fmt.Printf(" (")
		printStyledValueDimmed(getValueType(value))
		fmt.Printf(")")
		fmt.Println()

		// Show the actual value
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 200 {
			valueStr = valueStr[:197] + "..."
		}

		fmt.Printf("    Value: ")
		printStyledValueDimmed(valueStr)
		fmt.Println()
		fmt.Println()
	}

	// Show metadata statistics
	printStyledEmoji("📊")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Statistics")
	fmt.Println()

	// Calculate field type distribution
	typeCounts := make(map[string]int)
	for _, value := range doc.Metadata {
		typeCounts[getValueType(value)]++
	}

	fmt.Printf("  • Field Type Distribution: ")
	var typeDist []string
	for fieldType, count := range typeCounts {
		typeDist = append(typeDist, fmt.Sprintf("%s (%d)", fieldType, count))
	}
	printStyledValueDimmed(strings.Join(typeDist, ", "))
	fmt.Println()

	fmt.Printf("  • Average Value Length: ")
	totalLength := 0
	for _, value := range doc.Metadata {
		valueStr := fmt.Sprintf("%v", value)
		totalLength += len(valueStr)
	}
	avgLength := float64(totalLength) / float64(len(doc.Metadata))
	printStyledValueDimmed(fmt.Sprintf("%.1f characters", avgLength))
	fmt.Println()
}

// showMockDocumentMetadata shows expanded metadata for a mock document
func showMockDocumentMetadata(doc mock.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("📊 Document Metadata: %s\n", collectionName)
	fmt.Println()

	if len(doc.Metadata) == 0 {
		printStyledValueDimmed("No metadata available for this document")
		fmt.Println()
		return
	}

	// Display metadata analysis
	printStyledEmoji("📈")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Analysis")
	fmt.Println()

	fmt.Printf("  • Total Metadata Fields: ")
	printStyledValueDimmed(fmt.Sprintf("%d", len(doc.Metadata)))
	fmt.Println()

	fmt.Printf("  • Document ID: ")
	printStyledValueDimmed(doc.ID)
	fmt.Println()

	fmt.Println()

	// Show detailed metadata fields
	printStyledEmoji("🔍")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Fields")
	fmt.Println()

	// Sort fields alphabetically for consistent display
	var sortedKeys []string
	for key := range doc.Metadata {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := doc.Metadata[key]

		fmt.Printf("  • ")
		printStyledKeyProminent(key)
		fmt.Printf(" (")
		printStyledValueDimmed(getValueType(value))
		fmt.Printf(")")
		fmt.Println()

		// Show the actual value
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 200 {
			valueStr = valueStr[:197] + "..."
		}

		fmt.Printf("    Value: ")
		printStyledValueDimmed(valueStr)
		fmt.Println()
		fmt.Println()
	}

	// Show metadata statistics
	printStyledEmoji("📊")
	fmt.Printf(" ")
	printStyledKeyProminent("Metadata Statistics")
	fmt.Println()

	// Calculate field type distribution
	typeCounts := make(map[string]int)
	for _, value := range doc.Metadata {
		typeCounts[getValueType(value)]++
	}

	fmt.Printf("  • Field Type Distribution: ")
	var typeDist []string
	for fieldType, count := range typeCounts {
		typeDist = append(typeDist, fmt.Sprintf("%s (%d)", fieldType, count))
	}
	printStyledValueDimmed(strings.Join(typeDist, ", "))
	fmt.Println()

	fmt.Printf("  • Average Value Length: ")
	totalLength := 0
	for _, value := range doc.Metadata {
		valueStr := fmt.Sprintf("%v", value)
		totalLength += len(valueStr)
	}
	avgLength := float64(totalLength) / float64(len(doc.Metadata))
	printStyledValueDimmed(fmt.Sprintf("%.1f characters", avgLength))
	fmt.Println()

	// Mock-specific information
	printStyledEmoji("🎭")
	fmt.Printf(" ")
	printStyledKeyProminent("Mock Document Info")
	fmt.Println()
	fmt.Printf("  • Type: ")
	printStyledValueDimmed("Mock Document")
	fmt.Println()
	fmt.Printf("  • Collection: ")
	printStyledValueDimmed(collectionName)
	fmt.Println()
	fmt.Println()
}
