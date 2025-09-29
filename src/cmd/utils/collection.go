package utils

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// ParseFieldDefinitions parses field definitions from a string
func ParseFieldDefinitions(fieldsStr string) ([]weaviate.FieldDefinition, error) {
	PrintInfo("Field parsing not yet implemented in new structure")
	return nil, fmt.Errorf("field parsing not yet implemented")
}

// CreateWeaviateCollection creates a Weaviate collection
func CreateWeaviateCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition, schemaType string) error {
	PrintInfo("Weaviate collection creation not yet implemented in new structure")
	return fmt.Errorf("collection creation not yet implemented")
}

// CreateMockCollection creates a mock collection
func CreateMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) error {
	PrintInfo("Mock collection creation not yet implemented in new structure")
	return fmt.Errorf("collection creation not yet implemented")
}

// ListWeaviateCollections lists Weaviate collections
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

// ListMockCollections lists mock collections
func ListMockCollections(ctx context.Context, cfg *config.VectorDBConfig, limit int, virtual bool) {
	PrintInfo("Mock collection listing not yet implemented in new structure")
}

// ShowWeaviateCollection shows Weaviate collection details
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

// ShowMockCollection shows mock collection details
func ShowMockCollection(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, shortLines int, noTruncate bool, verbose bool, showSchema bool, showMetadata bool) {
	PrintInfo("Mock collection show not yet implemented in new structure")
}

// CountWeaviateCollections counts Weaviate collections
func CountWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %v", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list collections: %v", err)
	}

	return len(collections), nil
}

// CountMockCollections counts mock collections
func CountMockCollections(ctx context.Context, cfg *config.VectorDBConfig) (int, error) {
	PrintInfo("Mock collection counting not yet implemented in new structure")
	return 0, fmt.Errorf("mock collection counting not yet implemented")
}

// DeleteWeaviateCollections deletes Weaviate collections
func DeleteWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	PrintInfo("Weaviate collection deletion not yet implemented in new structure")
	return fmt.Errorf("collection deletion not yet implemented")
}

// DeleteMockCollections deletes mock collections
func DeleteMockCollections(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string) error {
	PrintInfo("Mock collection deletion not yet implemented in new structure")
	return fmt.Errorf("collection deletion not yet implemented")
}

// DeleteWeaviateCollectionsByPattern deletes Weaviate collections by pattern
func DeleteWeaviateCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	PrintInfo("Weaviate collection pattern deletion not yet implemented in new structure")
	return fmt.Errorf("collection pattern deletion not yet implemented")
}

// DeleteMockCollectionsByPattern deletes mock collections by pattern
func DeleteMockCollectionsByPattern(ctx context.Context, cfg *config.VectorDBConfig, pattern string) error {
	PrintInfo("Mock collection pattern deletion not yet implemented in new structure")
	return fmt.Errorf("collection pattern deletion not yet implemented")
}

// DeleteAllWeaviateCollections deletes all Weaviate collections
func DeleteAllWeaviateCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	PrintInfo("Weaviate collection deletion all not yet implemented in new structure")
}

// DeleteAllMockCollections deletes all mock collections
func DeleteAllMockCollections(ctx context.Context, cfg *config.VectorDBConfig) {
	PrintInfo("Mock collection deletion all not yet implemented in new structure")
}

// DeleteWeaviateCollectionSchema deletes Weaviate collection schema
func DeleteWeaviateCollectionSchema(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) error {
	PrintInfo("Weaviate collection schema deletion not yet implemented in new structure")
	return fmt.Errorf("collection schema deletion not yet implemented")
}

// showCollectionVirtualSummary shows virtual structure summary for a collection
func showCollectionVirtualSummary(ctx context.Context, client *weaviate.Client, collectionName string) {
	// This is a simplified implementation
	// In the real implementation, this would show virtual structure summary
	fmt.Printf("   Virtual structure: Not implemented yet\n")
}
