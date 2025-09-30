// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// VirtualDocument represents a virtual document structure
type VirtualDocument struct {
	OriginalFilename string
	TotalChunks      int
	Chunks           []weaviate.Document
	Metadata         map[string]interface{}
}

// DisplayRegularDocuments displays regular documents with styling
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

// DisplayVirtualDocuments displays virtual documents with aggregation and styling
func DisplayVirtualDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int, summary bool) {
	// Aggregate documents by original filename/source
	virtualDocs := AggregateDocumentsByOriginal(documents)

	if len(virtualDocs) == 0 {
		PrintWarning(fmt.Sprintf("No virtual documents found in collection '%s'", collectionName))
		return
	}

	PrintSuccess(fmt.Sprintf("Found %d virtual documents in collection '%s' (aggregated from %d total documents):", len(virtualDocs), collectionName, len(documents)))
	fmt.Println()

	for i, vdoc := range virtualDocs {
		fmt.Printf("%d. ", i+1)
		PrintStyledEmoji("📄")
		fmt.Printf(" ")
		PrintStyledKeyProminent("Document")
		fmt.Printf(": ")
		PrintStyledFilename(vdoc.OriginalFilename)
		fmt.Println()

		fmt.Printf("   ")
		PrintStyledKeyNumberProminentWithEmoji("Chunks", vdoc.TotalChunks, "📝")
		fmt.Println()

		// Show metadata if available
		if len(vdoc.Metadata) > 0 {
			fmt.Printf("   ")
			PrintStyledKeyValueProminentWithEmoji("Metadata", "", "📋")
			fmt.Println()
			for key, value := range vdoc.Metadata {
				if key != "id" { // Skip ID since it's redundant
					valueStr := fmt.Sprintf("%v", value)
					truncatedValue := SmartTruncate(valueStr, key, shortLines)
					fmt.Printf("     ")
					PrintStyledKeyValueDimmed(key, truncatedValue)
					fmt.Println()
				}
			}
		}

		// Show chunk details
		fmt.Printf("   ")
		PrintStyledKeyValueProminentWithEmoji("Chunk Details", "", "📝")
		fmt.Println()

		for j, chunk := range vdoc.Chunks {
			fmt.Printf("     %d. ", j+1)
			PrintStyledKeyProminent("ID")
			fmt.Printf(": ")
			PrintStyledID(chunk.ID)
			fmt.Println()

			fmt.Printf("        ")
			PrintStyledKeyProminent("Content")
			fmt.Printf(": ")
			if showLong {
				PrintStyledValue(chunk.Content)
			} else {
				preview := SmartTruncate(chunk.Content, "content", shortLines)
				PrintStyledValue(preview)
			}
			fmt.Println()
		}

		fmt.Println()
	}

	// Show summary if requested
	if summary {
		PrintStyledKeyValueProminentWithEmoji("Summary", "", "📋")
		fmt.Println()
		for i, vdoc := range virtualDocs {
			fmt.Printf("   %d. ", i+1)
			PrintStyledFilename(vdoc.OriginalFilename)
			fmt.Printf(" - ")
			PrintStyledNumber(vdoc.TotalChunks)
			fmt.Printf(" chunks")
			fmt.Println()
		}
		fmt.Println()
	}
}

// AggregateDocumentsByOriginal aggregates documents by their original filename/source
func AggregateDocumentsByOriginal(documents []weaviate.Document) []VirtualDocument {
	docMap := make(map[string]*VirtualDocument)

	for _, doc := range documents {
		// Try to get original filename from metadata
		var originalFilename string
		var metadata map[string]interface{}

		// First, check if original_filename is directly in the metadata
		if filename, ok := doc.Metadata["original_filename"].(string); ok {
			originalFilename = filename
			metadata = doc.Metadata
		} else if filename, ok := doc.Metadata["filename"].(string); ok {
			// Fallback to filename if original_filename is not available
			originalFilename = filename
			metadata = doc.Metadata
		} else {
			// Check if this is a chunked document with nested metadata
			if metadataField, ok := doc.Metadata["metadata"]; ok {
				if metadataStr, ok := metadataField.(string); ok {
					var metadataObj map[string]interface{}
					if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
						// Check if there's another nested metadata field that needs parsing
						if nestedMetadataStr, ok := metadataObj["metadata"].(string); ok {
							var nestedMetadataObj map[string]interface{}
							if err := json.Unmarshal([]byte(nestedMetadataStr), &nestedMetadataObj); err == nil {
								// Look for filename in the nested metadata (this is the current structure)
								if filename, ok := nestedMetadataObj["filename"].(string); ok {
									originalFilename = filename
									metadata = nestedMetadataObj
								} else if filename, ok := nestedMetadataObj["original_filename"].(string); ok {
									// Fallback to original_filename if available
									originalFilename = filename
									metadata = nestedMetadataObj
								}
							}
						} else {
							// Look for filename in the parsed metadata (this is the current structure)
							if filename, ok := metadataObj["filename"].(string); ok {
								originalFilename = filename
								metadata = metadataObj
							} else if filename, ok := metadataObj["original_filename"].(string); ok {
								// Fallback to original_filename if available
								originalFilename = filename
								metadata = metadataObj
							}
						}
					}
				}
			}
		}

		// If no original filename found, use document ID as key
		if originalFilename == "" {
			originalFilename = GetStandaloneDocumentKey(doc)
			metadata = doc.Metadata
		}

		// Add to virtual document
		if vdoc, exists := docMap[originalFilename]; exists {
			vdoc.Chunks = append(vdoc.Chunks, doc)
			vdoc.TotalChunks++
		} else {
			docMap[originalFilename] = &VirtualDocument{
				OriginalFilename: originalFilename,
				TotalChunks:      1,
				Chunks:           []weaviate.Document{doc},
				Metadata:         metadata,
			}
		}
	}

	// Convert map to slice and sort by filename
	var virtualDocs []VirtualDocument
	for _, vdoc := range docMap {
		virtualDocs = append(virtualDocs, *vdoc)
	}

	sort.Slice(virtualDocs, func(i, j int) bool {
		return virtualDocs[i].OriginalFilename < virtualDocs[j].OriginalFilename
	})

	return virtualDocs
}

// IsImageVirtualDocument checks if a virtual document represents an image collection
func IsImageVirtualDocument(vdoc VirtualDocument) bool {
	// Check if any chunk has image-related metadata
	for _, chunk := range vdoc.Chunks {
		if IsImageDocument(chunk.Metadata) {
			return true
		}
	}
	return false
}

// GetStandaloneDocumentKey extracts a unique key for standalone documents
func GetStandaloneDocumentKey(doc weaviate.Document) string {
	// For standalone documents, use a combination of ID and content hash
	contentHash := fmt.Sprintf("%d", len(doc.Content))
	return fmt.Sprintf("standalone-%s-%s", doc.ID[:8], contentHash)
}

// SmartTruncate intelligently truncates content based on key type
func SmartTruncate(value, key string, shortLines int) string {
	// For certain keys, use different truncation strategies
	switch key {
	case "content", "text":
		// For content, truncate by lines
		return TruncateStringByLines(value, shortLines)
	case "metadata":
		// For metadata JSON, truncate by characters but preserve structure
		if len(value) > 200 {
			return value[:197] + "..."
		}
		return value
	default:
		// For other fields, truncate by characters
		if len(value) > 100 {
			return value[:97] + "..."
		}
		return value
	}
}

// ShowDocumentSchema shows the schema of a Weaviate document
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

// ShowDocumentMetadata shows expanded metadata for a document
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
