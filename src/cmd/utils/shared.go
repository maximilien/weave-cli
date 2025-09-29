package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Shared utility functions for all command packages

// VirtualDocument represents a document aggregated from multiple chunks
type VirtualDocument struct {
	OriginalFilename string
	TotalChunks      int
	Chunks           []weaviate.Document
	Metadata         map[string]interface{}
}

// LoadConfigWithOverrides loads configuration with environment overrides
func LoadConfigWithOverrides() (*config.Config, error) {
	return config.LoadConfigWithOptions(config.LoadConfigOptions{
		ConfigFile:     "", // Will use default
		EnvFile:        "", // Will use default
		VectorDBType:   "", // Will use default
		WeaviateAPIKey: "", // Will use default
		WeaviateURL:    "", // Will use default
	})
}

// PrintError prints an error message to stderr
func PrintError(message string) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", message)
}

// PrintSuccess prints a success message to stdout
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintWarning prints a warning message to stdout
func PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintInfo prints an info message to stdout
func PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintHeader prints a header message to stdout
func PrintHeader(message string) {
	fmt.Printf("\n📋 %s\n", message)
}

// ConfirmAction prompts user for confirmation
func ConfirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}

// CreateWeaviateClient creates a Weaviate client from config
func CreateWeaviateClient(cfg *config.VectorDBConfig) (*weaviate.Client, error) {
	// Convert VectorDBConfig to weaviate.Config
	weaviateConfig := &weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	}
	return weaviate.NewClient(weaviateConfig)
}

// CreateMockClient creates a mock client from config
func CreateMockClient(cfg *config.VectorDBConfig) *mock.Client {
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

	return mock.NewClient(mockConfig)
}

// TruncateStringByLines truncates a string to a maximum number of lines
func TruncateStringByLines(text string, maxLines int) string {
	lines := make([]string, 0)
	currentLine := ""

	for _, char := range text {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
			if len(lines) >= maxLines {
				break
			}
		} else {
			currentLine += string(char)
		}
	}

	if currentLine != "" && len(lines) < maxLines {
		lines = append(lines, currentLine)
	}

	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}

	if len(lines) == maxLines && len(text) > len(result) {
		result += "\n... (truncated)"
	}

	return result
}

// Placeholder functions - these need to be implemented
// For now, they just print a message that the functionality is not yet implemented

func ListWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// For virtual mode, we need all documents to properly aggregate them
	// For regular mode, we can use the provided limit
	queryLimit := limit
	if virtual {
		queryLimit = 10000 // High limit to get all documents for proper aggregation
	}

	documents, err := client.ListDocuments(ctx, collectionName, queryLimit)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		PrintWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	if virtual {
		DisplayVirtualDocuments(documents, collectionName, showLong, shortLines, summary)
	} else {
		DisplayRegularDocuments(documents, collectionName, showLong, shortLines)
	}
}

func ListMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	PrintInfo("Mock document listing not yet implemented in new structure")
}

func CreateWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
	PrintInfo("Weaviate document creation not yet implemented in new structure")
}

func CreateMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
	PrintInfo("Mock document creation not yet implemented in new structure")
}

func ParseFieldDefinitions(fieldsStr string) ([]weaviate.FieldDefinition, error) {
	PrintInfo("Field parsing not yet implemented in new structure")
	return nil, fmt.Errorf("field parsing not yet implemented")
}

func CreateWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition, schemaType string) error {
	PrintInfo("Weaviate collection creation not yet implemented in new structure")
	return fmt.Errorf("collection creation not yet implemented")
}

func CreateMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) error {
	PrintInfo("Mock collection creation not yet implemented in new structure")
	return fmt.Errorf("collection creation not yet implemented")
}

// Collection functions
func ListWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		PrintWarning("No collections found in the database")
		return
	}

	// Sort collections alphabetically
	sort.Strings(collections)

	// Apply limit if specified
	if limit > 0 && len(collections) > limit {
		collections = collections[:limit]
	}

	PrintSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
	if limit > 0 && len(collections) == limit {
		fmt.Printf("(showing first %d collections)\n", limit)
	}
	fmt.Println()

	for i, collection := range collections {
		fmt.Printf("%d. %s\n", i+1, collection)

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

func ListMockCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool) {
	PrintInfo("Mock collection listing not yet implemented in new structure")
}

func ShowWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Check if collection exists by listing all collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	// Check if the specific collection exists
	collectionExists := false
	for _, existingCollection := range collections {
		if existingCollection == collectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		PrintError(fmt.Sprintf("Collection '%s' not found", collectionName))
		return
	}

	// Get document count using efficient method
	documentCount, err := client.CountDocuments(ctx, collectionName)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get document count: %v", err))
		return
	}

	// Display collection information with styling
	PrintStyledEmoji("📊")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection")
	fmt.Printf(": ")
	PrintStyledValueDimmed(collectionName)
	fmt.Println()

	PrintStyledKeyProminent("Database Type")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()

	PrintStyledKeyProminent("Document Count")
	fmt.Printf(": ")
	PrintStyledValueDimmed(fmt.Sprintf("%d", documentCount))
	fmt.Println()
	fmt.Println()

	// Show collection properties if available
	PrintStyledEmoji("🔧")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Collection Properties")
	fmt.Println()
	PrintStyledKeyProminent("  Vector Database")
	fmt.Printf(": ")
	PrintStyledValueDimmed(string(cfg.Type))
	fmt.Println()
	PrintStyledKeyProminent("  URL")
	fmt.Printf(": ")
	PrintStyledValueDimmed(cfg.URL)
	fmt.Println()
	if cfg.APIKey != "" {
		PrintStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		PrintStyledValueDimmed("[CONFIGURED]")
		fmt.Println()
	} else {
		PrintStyledKeyProminent("  API Key")
		fmt.Printf(": ")
		PrintStyledValueDimmed("[NOT CONFIGURED]")
		fmt.Println()
	}
	fmt.Println()

	if documentCount > 0 {
		// Get sample document for metadata analysis (just one document)
		sampleDocuments, err := client.ListDocuments(ctx, collectionName, 1)
		if err != nil {
			PrintWarning(fmt.Sprintf("Could not retrieve sample document: %v", err))
		} else if len(sampleDocuments) > 0 {
			// Get the actual document with full data using GetDocument
			sampleDoc, err := client.GetDocument(ctx, collectionName, sampleDocuments[0].ID)
			if err != nil {
				PrintWarning(fmt.Sprintf("Could not retrieve full sample document: %v", err))
				// Fall back to the basic document from ListDocuments
				sampleDoc = &sampleDocuments[0]
			}
			PrintStyledEmoji("📋")
			fmt.Printf(" ")
			PrintStyledKeyProminent("Sample Document Metadata")
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
						displayValue = TruncateMetadataValue(value, 100) // Limit each value to 100 chars
					}

					// Style the key-value pair directly
					fmt.Printf("  - ")
					PrintStyledKeyProminent(key)
					fmt.Printf(": ")
					if key == "id" {
						PrintStyledID(displayValue)
					} else {
						PrintStyledValueDimmed(displayValue)
					}
					fmt.Println()
					metadataCount++
				}
			} else {
				PrintStyledKey("  No metadata available")
				fmt.Println()
			}
			fmt.Println()

			// Show sample content
			if len(sampleDoc.Content) > 0 {
				PrintStyledEmoji("📄")
				fmt.Printf(" ")
				PrintStyledKeyProminent("Sample Document Content")
				fmt.Println()

				// Check if this is image content (base64 data)
				if IsImageContent(sampleDoc.Content) {
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
				if IsImageDocument(sampleDoc.Metadata) {
					PrintStyledEmoji("📄")
					fmt.Printf(" ")
					PrintStyledKeyProminent("Sample Document Content")
					fmt.Println()
					fmt.Printf("  📷 Image Document (no text content)\n")
					fmt.Printf("  ℹ️  Image data stored in metadata fields\n")
					fmt.Println()
				}
			}
		}
	}

	// Show schema if requested
	if showSchema {
		// ShowCollectionSchema(ctx, client, collectionName) // TODO: Implement
		PrintInfo("Collection schema display not yet implemented")
	}

	// Show expanded metadata if requested
	if showMetadata {
		// ShowCollectionMetadata(ctx, client, collectionName) // TODO: Implement
		PrintInfo("Collection metadata display not yet implemented")
	}

	PrintSuccess(fmt.Sprintf("Collection '%s' summary retrieved successfully", collectionName))
}

func ShowMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool) {
	PrintInfo("Mock collection show not yet implemented in new structure")
}

func CountWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %w", err)
	}

	return len(collections), nil
}

func CountMockCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	PrintInfo("Mock collection count not yet implemented in new structure")
	return 0, fmt.Errorf("collection count not yet implemented")
}

func DeleteWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	PrintInfo("Weaviate collection deletion not yet implemented in new structure")
	return fmt.Errorf("collection deletion not yet implemented")
}

func DeleteMockCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	PrintInfo("Mock collection deletion not yet implemented in new structure")
	return fmt.Errorf("collection deletion not yet implemented")
}

func DeleteWeaviateCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	PrintInfo("Weaviate collection deletion by pattern not yet implemented in new structure")
	return fmt.Errorf("collection deletion by pattern not yet implemented")
}

func DeleteMockCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	PrintInfo("Mock collection deletion by pattern not yet implemented in new structure")
	return fmt.Errorf("collection deletion by pattern not yet implemented")
}

func DeleteAllWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	PrintInfo("Weaviate delete all collections not yet implemented in new structure")
}

func DeleteAllMockCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	PrintInfo("Mock delete all collections not yet implemented in new structure")
}

func DeleteWeaviateCollectionSchema(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) error {
	PrintInfo("Weaviate collection schema deletion not yet implemented in new structure")
	return fmt.Errorf("collection schema deletion not yet implemented")
}

// Document functions
func ShowWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, showLong bool, shortLines int, metadataFilters []string, name string, showSchema bool, expandMetadata bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// If no specific document IDs provided, show all documents
	if len(documentIDs) == 0 {
		documents, err := client.ListDocuments(ctx, collectionName, 50) // Limit to 50 for show
		if err != nil {
			PrintError(fmt.Sprintf("Failed to list documents: %v", err))
			return
		}

		if len(documents) == 0 {
			PrintWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
			return
		}

		PrintSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
		fmt.Println()

		for i, doc := range documents {
			if i > 0 {
				fmt.Println(strings.Repeat("=", 80))
				fmt.Println()
			}

			// Display document details (matching original format)
			color.New(color.FgGreen).Printf("Document ID: %s\n", doc.ID)
			fmt.Printf("Collection: %s\n", collectionName)
			fmt.Println()

			fmt.Printf("Content:\n")
			if showLong {
				fmt.Printf("%s\n", doc.Content)
			} else {
				// Use shortLines to limit content by lines instead of characters
				preview := TruncateStringByLines(doc.Content, shortLines)
				fmt.Printf("%s\n", preview)
			}
			fmt.Println()

			if len(doc.Metadata) > 0 {
				fmt.Printf("Metadata:\n")
				for key, value := range doc.Metadata {
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := SmartTruncate(valueStr, key, shortLines)
					fmt.Printf("  %s: %s\n", key, truncatedValue)
				}
			}

			// Show schema if requested (only for first document to avoid repetition)
			if showSchema && i == 0 {
				ShowDocumentSchema(doc, collectionName)
			}

			// Show expanded metadata if requested (only for first document to avoid repetition)
			if expandMetadata && i == 0 {
				ShowDocumentMetadata(doc, collectionName)
			}
		}
	} else {
		// Show specific documents by ID
		for _, docID := range documentIDs {
			doc, err := client.GetDocument(ctx, collectionName, docID)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to get document '%s': %v", docID, err))
				continue
			}

			// Display document details (matching original format)
			color.New(color.FgGreen).Printf("Document ID: %s\n", doc.ID)
			fmt.Printf("Collection: %s\n", collectionName)
			fmt.Println()

			fmt.Printf("Content:\n")
			if showLong {
				fmt.Printf("%s\n", doc.Content)
			} else {
				// Use shortLines to limit content by lines instead of characters
				preview := TruncateStringByLines(doc.Content, shortLines)
				fmt.Printf("%s\n", preview)
			}
			fmt.Println()

			if len(doc.Metadata) > 0 {
				fmt.Printf("Metadata:\n")
				for key, value := range doc.Metadata {
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := SmartTruncate(valueStr, key, shortLines)
					fmt.Printf("  %s: %s\n", key, truncatedValue)
				}
			}

			// Show schema if requested
			if showSchema {
				ShowDocumentSchema(*doc, collectionName)
			}

			// Show expanded metadata if requested
			if expandMetadata {
				ShowDocumentMetadata(*doc, collectionName)
			}
		}
	}
}

func ShowMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, showLong bool, shortLines int, metadataFilters []string, name string, showSchema bool, expandMetadata bool) {
	PrintInfo("Mock document show not yet implemented in new structure")
}

func CountWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	count, err := client.CountDocuments(ctx, collectionName)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func CountMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
	PrintInfo("Mock document count not yet implemented in new structure")
	return 0, fmt.Errorf("document count not yet implemented")
}

func DeleteWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, metadataFilters []string, virtual bool, pattern string, name string) {
	PrintInfo("Weaviate document deletion not yet implemented in new structure")
}

func DeleteMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, metadataFilters []string, virtual bool, pattern string, name string) {
	PrintInfo("Mock document deletion not yet implemented in new structure")
}

func DeleteAllWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	PrintInfo("Weaviate delete all documents not yet implemented in new structure")
}

func DeleteAllMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	PrintInfo("Mock delete all documents not yet implemented in new structure")
}

// Display functions for virtual and regular documents
func DisplayRegularDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int) {
	PrintSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	fmt.Println()

	for i, doc := range documents {
		fmt.Printf("%d. ", i+1)
		PrintStyledEmoji("📄")
		fmt.Printf(" ")
		PrintStyledKeyProminent("ID")
		fmt.Printf(": ")
		PrintStyledID(doc.ID)
		fmt.Println()

		// Only show content if it's not just the redundant "Document ID: [ID]"
		if doc.Content != fmt.Sprintf("Document ID: %s", doc.ID) {
			fmt.Printf("   ")
			PrintStyledKeyProminent("Content")
			fmt.Printf(": ")
			if showLong {
				PrintStyledValue(doc.Content)
			} else {
				// Use smartTruncate to handle base64 content appropriately
				preview := SmartTruncate(doc.Content, "content", shortLines)
				PrintStyledValue(preview)
			}
			fmt.Println()
		}

		if len(doc.Metadata) > 0 {
			fmt.Printf("   ")
			PrintStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range doc.Metadata {
				if key != "id" { // Skip ID since it's already shown
					// Use smartTruncate to handle base64 content appropriately
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := SmartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					PrintStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}
}

func DisplayVirtualDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int, summary bool) {
	virtualDocs := AggregateDocumentsByOriginal(documents)

	PrintSuccess(fmt.Sprintf("Found %d virtual documents in collection '%s' (aggregated from %d total documents):", len(virtualDocs), collectionName, len(documents)))
	fmt.Println()

	for i, vdoc := range virtualDocs {
		fmt.Printf("%d. ", i+1)
		PrintStyledEmoji("📄")
		fmt.Printf(" Document: ")
		PrintStyledFilename(vdoc.OriginalFilename)
		fmt.Println()

		// Determine if this is an image collection
		isImageCollection := IsImageVirtualDocument(vdoc)

		fmt.Printf("   ")
		if vdoc.TotalChunks > 0 {
			PrintStyledKeyNumberProminentWithEmoji("Chunks", len(vdoc.Chunks), "📝")
			fmt.Printf("/")
			PrintStyledNumber(vdoc.TotalChunks)
			fmt.Println()
		} else if isImageCollection {
			PrintStyledKeyNumberProminentWithEmoji("Images", len(vdoc.Chunks), "🖼️")
			fmt.Println()
		} else {
			PrintStyledKeyValueProminentWithEmoji("Type", "Single document (no chunks)", "📄")
			fmt.Println()
		}

		// Show metadata from the first chunk or document
		if len(vdoc.Metadata) > 0 {
			fmt.Printf("   ")
			PrintStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range vdoc.Metadata {
				if key != "id" && key != "chunk_index" && key != "total_chunks" && key != "is_chunked" {
					// Truncate value based on shortLines directive
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := SmartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					PrintStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}

		// Show details if there are items
		if len(vdoc.Chunks) > 0 {
			fmt.Printf("   ")
			if isImageCollection {
				PrintStyledKeyValueProminentWithEmoji("Stack Details", "", "🗂️")
			} else {
				PrintStyledKeyValueProminentWithEmoji("Chunk Details", "", "📝")
			}
			fmt.Println()

			for j, chunk := range vdoc.Chunks {
				fmt.Printf("     %d. ", j+1)
				PrintStyledKeyProminent("ID")
				fmt.Printf(": ")
				PrintStyledID(chunk.ID)
				if chunkIndex, ok := chunk.Metadata["chunk_index"]; ok {
					fmt.Printf(" (")
					PrintStyledKeyProminent("chunk")
					fmt.Printf(" ")
					PrintStyledNumber(int(chunkIndex.(float64)))
					fmt.Printf(")")
				}
				fmt.Println()

				if chunk.Content != fmt.Sprintf("Document ID: %s", chunk.ID) {
					fmt.Printf("        ")
					PrintStyledKeyProminent("Content")
					fmt.Printf(": ")
					if showLong {
						PrintStyledValueDimmed(chunk.Content)
					} else {
						preview := SmartTruncate(chunk.Content, "content", shortLines)
						PrintStyledValueDimmed(preview)
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
		PrintStyledKeyValueProminentWithEmoji("Summary", "", "📋")
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
			} else if IsImageVirtualDocument(vdoc) {
				totalImages += len(vdoc.Chunks)
				imageStackCount++
			} else {
				documentCount++
			}
		}

		// Show individual document details
		for i, vdoc := range virtualDocs {
			fmt.Printf("   %d. ", i+1)
			PrintStyledFilename(vdoc.OriginalFilename)
			if vdoc.TotalChunks > 0 {
				fmt.Printf(" - %d chunks", len(vdoc.Chunks))
			} else if IsImageVirtualDocument(vdoc) {
				fmt.Printf(" - %d images", len(vdoc.Chunks))
			} else {
				fmt.Printf(" - Single document")
			}
			fmt.Println()
		}

		// Show totals
		if len(virtualDocs) > 1 {
			fmt.Println()
			PrintStyledKeyValueProminentWithEmoji("Totals", "", "📊")
			fmt.Printf("   ")
			PrintStyledKeyNumberProminentWithEmoji("Documents", documentCount, "📄")
			fmt.Printf(" (")
			PrintStyledKeyNumberProminentWithEmoji("Total Chunks", totalChunks, "📝")
			fmt.Printf(")")
			fmt.Println()

			if imageStackCount > 0 {
				fmt.Printf("   ")
				PrintStyledKeyNumberProminentWithEmoji("Image Stacks", imageStackCount, "🗂️")
				fmt.Printf(" (")
				PrintStyledKeyNumberProminentWithEmoji("Total Images", totalImages, "🖼️")
				fmt.Printf(")")
				fmt.Println()
			}
		}
	}
}

// Helper functions for virtual document processing
func AggregateDocumentsByOriginal(documents []weaviate.Document) []VirtualDocument {
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

		// For standalone documents, use URL or filename as key
		groupKey := GetStandaloneDocumentKey(doc)

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

	// Convert map to slice and calculate total chunks for documents that don't have it set
	var virtualDocs []VirtualDocument
	for _, vdoc := range docMap {
		// If TotalChunks is 0 but we have chunks, determine if this is a text or image document
		if vdoc.TotalChunks == 0 && len(vdoc.Chunks) > 0 {
			// Check if this is an image document
			isImageDoc := false
			for _, chunk := range vdoc.Chunks {
				if metadata, ok := chunk.Metadata["metadata"]; ok {
					var metadataObj map[string]interface{}
					if metadataStr, ok := metadata.(string); ok {
						if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
							if contentType, ok := metadataObj["content_type"].(string); ok {
								if contentType == "image" {
									isImageDoc = true
									break
								}
							}
						}
					} else if metadataMap, ok := metadata.(map[string]interface{}); ok {
						metadataObj = metadataMap
						// Check for nested metadata
						if nestedMetadata, ok := metadataObj["metadata"]; ok {
							if nestedStr, ok := nestedMetadata.(string); ok {
								if err := json.Unmarshal([]byte(nestedStr), &metadataObj); err == nil {
									if contentType, ok := metadataObj["content_type"].(string); ok {
										if contentType == "image" {
											isImageDoc = true
											break
										}
									}
								}
							}
						}
					}
				}
			}

			if !isImageDoc {
				// This is a text document, set total chunks to the number of chunks we have
				vdoc.TotalChunks = len(vdoc.Chunks)
			}
		}

		virtualDocs = append(virtualDocs, *vdoc)
	}

	// Sort by original filename for consistent output
	sort.Slice(virtualDocs, func(i, j int) bool {
		return virtualDocs[i].OriginalFilename < virtualDocs[j].OriginalFilename
	})

	return virtualDocs
}

func IsImageVirtualDocument(vdoc VirtualDocument) bool {
	// Check if any chunk has image content type
	for _, chunk := range vdoc.Chunks {
		if metadata, ok := chunk.Metadata["metadata"]; ok {
			var metadataObj map[string]interface{}
			if metadataStr, ok := metadata.(string); ok {
				if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
					if contentType, ok := metadataObj["content_type"].(string); ok {
						if contentType == "image" {
							return true
						}
					}
				}
			} else if metadataMap, ok := metadata.(map[string]interface{}); ok {
				metadataObj = metadataMap
				if contentType, ok := metadataObj["content_type"].(string); ok {
					if contentType == "image" {
						return true
					}
				}
			}
		}
	}
	return false
}

func GetStandaloneDocumentKey(doc weaviate.Document) string {
	// Try to get filename from URL first
	if url, ok := doc.Metadata["url"].(string); ok {
		parts := strings.Split(url, "#")
		if len(parts) > 0 {
			filename := strings.TrimPrefix(parts[0], "file://")
			return filepath.Base(filename)
		}
	}

	// Try to get filename from metadata
	if metadata, ok := doc.Metadata["metadata"]; ok {
		var metadataObj map[string]interface{}
		if metadataStr, ok := metadata.(string); ok {
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if filename, ok := metadataObj["filename"].(string); ok {
					return filename
				}
			}
		} else if metadataMap, ok := metadata.(map[string]interface{}); ok {
			if filename, ok := metadataMap["filename"].(string); ok {
				return filename
			}
		}
	}

	// Fallback to document ID
	return doc.ID
}

func SmartTruncate(value, key string, shortLines int) string {
	// For content fields, truncate by lines
	if key == "content" || key == "text" || key == "ai_summary" {
		return TruncateStringByLines(value, shortLines)
	}

	// For other fields, truncate by characters
	if len(value) > 100 {
		return value[:100] + "..."
	}
	return value
}

// Styled output functions for virtual structure display

// PrintStyledKey prints a styled key (dimmed)
func PrintStyledKey(key string) {
	color.New(color.FgWhite, color.Faint).Printf("%s", key)
}

// PrintStyledValue prints a styled value (normal color)
func PrintStyledValue(value string) {
	color.New(color.FgWhite).Printf("%s", value)
}

// PrintStyledValueDimmed prints a styled value (dimmed color)
func PrintStyledValueDimmed(value string) {
	color.New(color.FgWhite, color.Faint).Printf("%s", value)
}

// PrintStyledID prints a styled ID (highlighted)
func PrintStyledID(id string) {
	color.New(color.FgYellow, color.Bold).Printf("%s", id)
}

// PrintStyledFilename prints a styled filename (highlighted)
func PrintStyledFilename(filename string) {
	color.New(color.FgCyan, color.Bold).Printf("%s", filename)
}

// PrintStyledNumber prints a styled number (highlighted)
func PrintStyledNumber(num int) {
	color.New(color.FgGreen, color.Bold).Printf("%d", num)
}

// PrintStyledEmoji prints an emoji
func PrintStyledEmoji(emoji string) {
	fmt.Printf("%s", emoji)
}

// PrintStyledKeyValueDimmed prints a key-value pair with dimmed value styling
func PrintStyledKeyValueDimmed(key, value string) {
	PrintStyledKey(key)
	fmt.Printf(": ")
	PrintStyledValueDimmed(value)
}

// PrintStyledKeyProminent prints a prominent key (normal color, not dimmed)
func PrintStyledKeyProminent(key string) {
	color.New(color.FgWhite).Printf("%s", key)
}

// PrintStyledKeyValueProminentWithEmoji prints a key-value pair with emoji and prominent key styling
func PrintStyledKeyValueProminentWithEmoji(key, value, emoji string) {
	PrintStyledEmoji(emoji)
	fmt.Printf(" ")
	PrintStyledKeyProminent(key)
	fmt.Printf(": ")
	PrintStyledValue(value)
}

// PrintStyledKeyNumberProminentWithEmoji prints a key-number pair with emoji and prominent key styling
func PrintStyledKeyNumberProminentWithEmoji(key string, num int, emoji string) {
	PrintStyledEmoji(emoji)
	fmt.Printf(" ")
	PrintStyledKeyProminent(key)
	fmt.Printf(": ")
	PrintStyledNumber(num)
}

// Helper functions for collection and document analysis
func TruncateMetadataValue(value interface{}, maxLength int) string {
	valueStr := fmt.Sprintf("%v", value)
	if len(valueStr) > maxLength {
		return valueStr[:maxLength] + "..."
	}
	return valueStr
}

func IsImageContent(content string) bool {
	// Check if content looks like base64 image data
	if len(content) < 100 {
		return false
	}
	// Simple heuristic: base64 image data is usually very long and contains base64 characters
	base64Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	base64Count := 0
	for _, char := range content[:min(200, len(content))] {
		if strings.ContainsRune(base64Chars, char) {
			base64Count++
		}
	}
	return float64(base64Count)/float64(min(200, len(content))) > 0.8
}

func IsImageDocument(metadata map[string]interface{}) bool {
	// Check if metadata indicates this is an image document
	if metadata == nil {
		return false
	}
	
	// Check for image-related metadata fields
	imageFields := []string{"image", "image_data", "base64_data", "content_type"}
	for _, field := range imageFields {
		if _, exists := metadata[field]; exists {
			return true
		}
	}
	
	// Check nested metadata
	if metadataField, ok := metadata["metadata"]; ok {
		if metadataStr, ok := metadataField.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if contentType, ok := metadataObj["content_type"].(string); ok {
					return contentType == "image"
				}
			}
		}
	}
	
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Document schema and metadata functions
func ShowDocumentSchema(doc weaviate.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("📋 Document Schema: %s\n", collectionName)
	fmt.Println()

	// Document structure
	PrintStyledEmoji("🏗️")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Document Structure")
	fmt.Println()

	fmt.Printf("  • ")
	PrintStyledKeyProminent("id")
	fmt.Printf(" (")
	PrintStyledValueDimmed("string")
	fmt.Printf(") - Unique document identifier")
	fmt.Println()

	fmt.Printf("  • ")
	PrintStyledKeyProminent("content")
	fmt.Printf(" (")
	PrintStyledValueDimmed("text")
	fmt.Printf(") - Document text content")
	fmt.Println()

	// Metadata fields
	if len(doc.Metadata) > 0 {
		fmt.Printf("  • ")
		PrintStyledKeyProminent("metadata")
		fmt.Printf(" (")
		PrintStyledValueDimmed("object")
		fmt.Printf(") - Document metadata")
		fmt.Println()

		// Show metadata field types
		PrintStyledEmoji("📊")
		fmt.Printf(" ")
		PrintStyledKeyProminent("Metadata Fields")
		fmt.Println()

		for key, value := range doc.Metadata {
			fmt.Printf("    • ")
			PrintStyledKeyProminent(key)
			fmt.Printf(" (")
			PrintStyledValueDimmed(GetValueType(value))
			fmt.Printf(")")
			fmt.Println()
		}
	} else {
		fmt.Printf("  • ")
		PrintStyledKeyProminent("metadata")
		fmt.Printf(" (")
		PrintStyledValueDimmed("object")
		fmt.Printf(") - No metadata fields")
		fmt.Println()
	}

	fmt.Println()
}

func ShowDocumentMetadata(doc weaviate.Document, collectionName string) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("📊 Document Metadata: %s\n", collectionName)
	fmt.Println()

	if len(doc.Metadata) == 0 {
		PrintStyledValueDimmed("No metadata available for this document")
		fmt.Println()
		return
	}

	// Display metadata analysis
	PrintStyledEmoji("📈")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Metadata Analysis")
	fmt.Println()

	fmt.Printf("  • Total Metadata Fields: ")
	PrintStyledValueDimmed(fmt.Sprintf("%d", len(doc.Metadata)))
	fmt.Println()

	fmt.Printf("  • Document ID: ")
	PrintStyledValueDimmed(doc.ID)
	fmt.Println()

	fmt.Println()

	// Show detailed metadata fields
	PrintStyledEmoji("🔍")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Metadata Fields")
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
		PrintStyledKeyProminent(key)
		fmt.Printf(" (")
		PrintStyledValueDimmed(GetValueType(value))
		fmt.Printf(")")
		fmt.Println()

		// Show the actual value
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 200 {
			valueStr = valueStr[:197] + "..."
		}

		fmt.Printf("    Value: ")
		PrintStyledValueDimmed(valueStr)
		fmt.Println()
		fmt.Println()
	}

	// Show metadata statistics
	PrintStyledEmoji("📊")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Metadata Statistics")
	fmt.Println()

	// Calculate field type distribution
	typeCounts := make(map[string]int)
	for _, value := range doc.Metadata {
		typeCounts[GetValueType(value)]++
	}

	fmt.Printf("  • Field Type Distribution: ")
	var typeDist []string
	for fieldType, count := range typeCounts {
		typeDist = append(typeDist, fmt.Sprintf("%s (%d)", fieldType, count))
	}
	PrintStyledValueDimmed(strings.Join(typeDist, ", "))
	fmt.Println()

	fmt.Printf("  • Average Value Length: ")
	totalLength := 0
	for _, value := range doc.Metadata {
		valueStr := fmt.Sprintf("%v", value)
		totalLength += len(valueStr)
	}
	avgLength := float64(totalLength) / float64(len(doc.Metadata))
	PrintStyledValueDimmed(fmt.Sprintf("%.1f characters", avgLength))
	fmt.Println()
}

func GetValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case float32, float64:
		return "float"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func showCollectionVirtualSummary(ctx context.Context, client *weaviate.Client, collectionName string) {
	// This is a simplified implementation
	// In the real implementation, this would show virtual structure summary
	fmt.Printf("   Virtual structure: Not implemented yet\n")
}
