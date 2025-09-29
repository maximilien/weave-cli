package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
)

// ListWeaviateDocuments lists Weaviate documents
func ListWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Get documents from collection
	documents, err := client.ListDocuments(ctx, collectionName, limit)
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

// ListMockDocuments lists mock documents
func ListMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	PrintInfo("Mock document listing not yet implemented in new structure")
}

// CreateWeaviateDocument creates a Weaviate document
func CreateWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
	PrintInfo("Weaviate document creation not yet implemented in new structure")
}

// CreateMockDocument creates a mock document
func CreateMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
	PrintInfo("Mock document creation not yet implemented in new structure")
}

// ShowWeaviateDocument shows Weaviate document details
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

// ShowMockDocument shows mock document details
func ShowMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, showLong bool, shortLines int, metadataFilters []string, name string, showSchema bool, expandMetadata bool) {
	PrintInfo("Mock document show not yet implemented in new structure")
}

// CountWeaviateDocuments counts Weaviate documents
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

// CountMockDocuments counts mock documents
func CountMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) (int, error) {
	PrintInfo("Mock document count not yet implemented in new structure")
	return 0, fmt.Errorf("document count not yet implemented")
}

// DeleteWeaviateDocuments deletes Weaviate documents
func DeleteWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, metadataFilters []string, virtual bool, pattern string, name string) {
	PrintInfo("Weaviate document deletion not yet implemented in new structure")
}

// DeleteMockDocuments deletes mock documents
func DeleteMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, metadataFilters []string, virtual bool, pattern string, name string) {
	PrintInfo("Mock document deletion not yet implemented in new structure")
}

// DeleteAllWeaviateDocuments deletes all Weaviate documents
func DeleteAllWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	PrintInfo("Weaviate delete all documents not yet implemented in new structure")
}

// DeleteAllMockDocuments deletes all mock documents
func DeleteAllMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	PrintInfo("Mock delete all documents not yet implemented in new structure")
}
