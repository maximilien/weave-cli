package utils

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Shared utility functions for all command packages

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

	// List documents
	documents, err := client.ListDocuments(ctx, collectionName, limit)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to list documents: %v", err))
		return
	}

	if len(documents) == 0 {
		PrintWarning(fmt.Sprintf("No documents found in collection '%s'", collectionName))
		return
	}

	PrintSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	if limit > 0 && len(documents) == limit {
		fmt.Printf("(showing first %d documents)\n", limit)
	}
	fmt.Println()

	for i, doc := range documents {
		fmt.Printf("%d. ID: %s\n", i+1, doc.ID)
		
		if showLong {
			fmt.Printf("   Content: %s\n", doc.Content)
		} else {
			preview := TruncateStringByLines(doc.Content, shortLines)
			fmt.Printf("   Content: %s\n", preview)
		}
		
		if len(doc.Metadata) > 0 {
			fmt.Printf("   Metadata: %v\n", doc.Metadata)
		}
		fmt.Println()
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
	PrintInfo("Weaviate collection show not yet implemented in new structure")
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
			fmt.Printf("Document %d:\n", i+1)
			fmt.Printf("  ID: %s\n", doc.ID)
			
			if showLong {
				fmt.Printf("  Content: %s\n", doc.Content)
			} else {
				preview := TruncateStringByLines(doc.Content, shortLines)
				fmt.Printf("  Content: %s\n", preview)
			}
			
			if len(doc.Metadata) > 0 {
				fmt.Printf("  Metadata: %v\n", doc.Metadata)
			}
			fmt.Println()
		}
	} else {
		// Show specific documents by ID
		for _, docID := range documentIDs {
			doc, err := client.GetDocument(ctx, collectionName, docID)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to get document '%s': %v", docID, err))
				continue
			}

			fmt.Printf("Document ID: %s\n", doc.ID)
			
			if showLong {
				fmt.Printf("Content: %s\n", doc.Content)
			} else {
				preview := TruncateStringByLines(doc.Content, shortLines)
				fmt.Printf("Content: %s\n", preview)
			}
			
			if len(doc.Metadata) > 0 {
				fmt.Printf("Metadata: %v\n", doc.Metadata)
			}
			fmt.Println()
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

// Helper functions
func showCollectionVirtualSummary(ctx context.Context, client *weaviate.Client, collectionName string) {
	// This is a simplified implementation
	// In the real implementation, this would show virtual structure summary
	fmt.Printf("   Virtual structure: Not implemented yet\n")
}
