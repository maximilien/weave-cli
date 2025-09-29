package utils

import (
	"context"
	"fmt"
	"os"

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
	PrintInfo("Weaviate document listing not yet implemented in new structure")
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
	PrintInfo("Weaviate collection listing not yet implemented in new structure")
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
	PrintInfo("Weaviate collection count not yet implemented in new structure")
	return 0, fmt.Errorf("collection count not yet implemented")
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
	PrintInfo("Weaviate document show not yet implemented in new structure")
}

func ShowMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, showLong bool, shortLines int, metadataFilters []string, name string, showSchema bool, expandMetadata bool) {
	PrintInfo("Mock document show not yet implemented in new structure")
}

func CountWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
	PrintInfo("Weaviate document count not yet implemented in new structure")
	return 0, fmt.Errorf("document count not yet implemented")
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
