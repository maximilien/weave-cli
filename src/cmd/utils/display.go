// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

// VirtualDocument represents a virtual document structure
type VirtualDocument struct {
	OriginalFilename string
	TotalChunks      int
	Chunks           []weaviate.Document
	Metadata         map[string]interface{}
}

// ImageGroup represents a group of images or documents from the same source
type ImageGroup struct {
	SourcePDF   string
	DisplayName string
	TotalCount  int
	TotalChunks int
	Images      []VirtualDocument
	MinPage     int
	MaxPage     int
	IsImage     bool // true for images, false for regular documents
}

// groupImagesBySource groups images by their PDF source for compact display
func groupImagesBySource(virtualDocs []VirtualDocument) []ImageGroup {
	// Check if these are image documents by looking for image naming pattern
	if len(virtualDocs) == 0 {
		return nil
	}

	// Group by source PDF (extract from filename like "ragme-io_image_1.png" or "ragme-io.pdf_page_3_image_2.png")
	// Also handle standalone image files
	groups := make(map[string]*ImageGroup)

	for _, vdoc := range virtualDocs {
		filename := vdoc.OriginalFilename

		// Check if this matches PDF page pattern: source.pdf_page_X_image_Y.ext
		if strings.Contains(filename, "_page_") && strings.Contains(filename, "_image_") {
			// Extract source PDF name (everything before "_page_")
			parts := strings.Split(filename, "_page_")
			if len(parts) >= 2 {
				sourceName := parts[0]

				if _, exists := groups[sourceName]; !exists {
					groups[sourceName] = &ImageGroup{
						SourcePDF: sourceName,
						Images:    []VirtualDocument{},
						MinPage:   -1,
						MaxPage:   -1,
					}
				}
				groups[sourceName].Images = append(groups[sourceName].Images, vdoc)
			}
		} else if strings.Contains(filename, "_image_") {
			// Check if this matches image pattern: source_image_N.ext
			// Extract source name (everything before "_image_")
			parts := strings.Split(filename, "_image_")
			if len(parts) >= 2 {
				sourceName := parts[0]

				if _, exists := groups[sourceName]; !exists {
					groups[sourceName] = &ImageGroup{
						SourcePDF: sourceName,
						Images:    []VirtualDocument{},
						MinPage:   -1,
						MaxPage:   -1,
					}
				}
				groups[sourceName].Images = append(groups[sourceName].Images, vdoc)
			}
		} else {
			// Standalone image file - treat as its own group
			groups[filename] = &ImageGroup{
				SourcePDF: filename,
				Images:    []VirtualDocument{vdoc},
				MinPage:   -1,
				MaxPage:   -1,
			}
		}
	}

	// Convert to slice and create compact display names
	var result []ImageGroup
	for _, group := range groups {
		// Set total count to actual number of images in this group
		group.TotalCount = len(group.Images)

		// Calculate total chunks
		totalChunks := 0
		for _, img := range group.Images {
			totalChunks += img.TotalChunks
		}
		group.TotalChunks = totalChunks

		// Determine if this group contains images by checking metadata
		if len(group.Images) > 0 {
			// Check if the first document has type: image in metadata
			if docType, ok := group.Images[0].Metadata["type"].(string); ok && docType == "image" {
				group.IsImage = true
			}
		}

		// Sort images to find range
		sort.Slice(group.Images, func(i, j int) bool {
			return group.Images[i].OriginalFilename < group.Images[j].OriginalFilename
		})

		// Check if this is a single standalone image
		if len(group.Images) == 1 && !strings.Contains(group.Images[0].OriginalFilename, "_page_") && !strings.Contains(group.Images[0].OriginalFilename, "_image_") {
			// Single standalone image - use filename as-is
			group.DisplayName = group.SourcePDF
		} else if len(group.Images) > 0 && strings.Contains(group.Images[0].OriginalFilename, "_page_") {
			// Extract page numbers for PDF pattern
			minPage, maxPage := extractPageNumbers(group.Images)
			group.MinPage = minPage
			group.MaxPage = maxPage

			// Create display name for PDF pages
			if minPage > 0 && maxPage > 0 {
				if minPage == maxPage {
					group.DisplayName = fmt.Sprintf("%s (page %d)", group.SourcePDF, minPage)
				} else {
					group.DisplayName = fmt.Sprintf("%s (pages %d-%d)", group.SourcePDF, minPage, maxPage)
				}
			} else {
				group.DisplayName = fmt.Sprintf("%s (images)", group.SourcePDF)
			}
		} else {
			// Extract min and max image numbers for non-PDF pattern
			minNum, maxNum := extractImageNumbers(group.Images)

			// Create display name with range
			if minNum > 0 && maxNum > 0 {
				group.DisplayName = fmt.Sprintf("%s_image_%d...%d.png", group.SourcePDF, minNum, maxNum)
			} else {
				group.DisplayName = fmt.Sprintf("%s (images)", group.SourcePDF)
			}
		}

		result = append(result, *group)
	}

	// Sort by source name
	sort.Slice(result, func(i, j int) bool {
		return result[i].SourcePDF < result[j].SourcePDF
	})

	return result
}

// extractImageNumbers extracts the min and max image numbers from filenames
func extractImageNumbers(images []VirtualDocument) (int, int) {
	if len(images) == 0 {
		return 0, 0
	}

	min := -1
	max := -1

	for _, img := range images {
		// Extract number from filename like "source_image_123.png"
		filename := img.OriginalFilename
		if strings.Contains(filename, "_image_") {
			parts := strings.Split(filename, "_image_")
			if len(parts) >= 2 {
				// Get number part (before .ext)
				numPart := strings.Split(parts[1], ".")[0]
				var num int
				_, err := fmt.Sscanf(numPart, "%d", &num)
				if err != nil {
					// If parsing fails, skip this file
					continue
				}

				if min == -1 || num < min {
					min = num
				}
				if max == -1 || num > max {
					max = num
				}
			}
		}
	}

	return min, max
}

// extractPageNumbers extracts the min and max page numbers from PDF image filenames
func extractPageNumbers(images []VirtualDocument) (int, int) {
	if len(images) == 0 {
		return 0, 0
	}

	min := -1
	max := -1

	for _, img := range images {
		// Extract page number from filename like "ragme-io.pdf_page_3_image_2.png"
		filename := img.OriginalFilename
		if strings.Contains(filename, "_page_") {
			parts := strings.Split(filename, "_page_")
			if len(parts) >= 2 {
				// Get page number (between "_page_" and "_image_")
				pageImagePart := parts[1]
				if strings.Contains(pageImagePart, "_image_") {
					pageParts := strings.Split(pageImagePart, "_image_")
					if len(pageParts) >= 1 {
						var pageNum int
						_, err := fmt.Sscanf(pageParts[0], "%d", &pageNum)
						if err != nil {
							// If parsing fails, skip this file
							continue
						}

						if min == -1 || pageNum < min {
							min = pageNum
						}
						if max == -1 || pageNum > max {
							max = pageNum
						}
					}
				}
			}
		}
	}

	return min, max
}

// DisplayRegularDocuments displays regular documents with styling
func DisplayRegularDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int, noTruncate bool, jsonOutput bool, vectorizer string) {
	// If JSON output is requested, marshal and print JSON
	if jsonOutput {
		output := map[string]interface{}{
			"collection":      collectionName,
			"total_documents": len(documents),
			"documents":       documents,
		}
		if vectorizer != "" {
			output["vectorizer"] = vectorizer
		}
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal documents to JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
		return
	}

	PrintSuccess(fmt.Sprintf("Found %d documents in collection '%s':", len(documents), collectionName))
	if vectorizer != "" {
		PrintInfo(fmt.Sprintf("  Embedding model: %s", GetStyledValueDimmed(vectorizer)))
	}
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
			if showLong || noTruncate {
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
					valueStr := fmt.Sprintf("%v", value)
					var displayValue string
					if noTruncate {
						displayValue = valueStr
					} else {
						// Use smartTruncate to handle base64 content appropriately
						displayValue = SmartTruncate(valueStr, key, shortLines)
					}
					fmt.Printf("     ")
					PrintStyledKeyValueDimmed(key, displayValue)
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}
}

// DisplayVirtualDocuments displays virtual documents with aggregation and styling
func DisplayVirtualDocuments(documents []weaviate.Document, collectionName string, showLong bool, shortLines int, noTruncate bool, summary bool, jsonOutput bool, vectorizer string) {
	// Aggregate documents by original filename/source
	virtualDocs := AggregateDocumentsByOriginal(documents)

	if len(virtualDocs) == 0 {
		if jsonOutput {
			output := map[string]interface{}{
				"collection":        collectionName,
				"total_documents":   0,
				"virtual_documents": []VirtualDocument{},
			}
			if vectorizer != "" {
				output["vectorizer"] = vectorizer
			}
			jsonBytes, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			PrintWarning(fmt.Sprintf("No virtual documents found in collection '%s'", collectionName))
		}
		return
	}

	// If JSON output is requested, marshal and print JSON
	if jsonOutput {
		output := map[string]interface{}{
			"collection":              collectionName,
			"total_documents":         len(documents),
			"total_virtual_documents": len(virtualDocs),
			"virtual_documents":       virtualDocs,
		}
		if vectorizer != "" {
			output["vectorizer"] = vectorizer
		}
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal virtual documents to JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
		return
	}

	PrintSuccess(fmt.Sprintf("Found %d virtual documents in collection '%s' (aggregated from %d total documents):", len(virtualDocs), collectionName, len(documents)))
	if vectorizer != "" {
		PrintInfo(fmt.Sprintf("  Embedding model: %s", GetStyledValueDimmed(vectorizer)))
	}
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
					var displayValue string
					if noTruncate {
						displayValue = valueStr
					} else {
						displayValue = SmartTruncate(valueStr, key, shortLines)
					}
					fmt.Printf("     ")
					PrintStyledKeyValueDimmed(key, displayValue)
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
			if showLong || noTruncate {
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

		// Group images by PDF source for compact display
		imageGroups := groupImagesBySource(virtualDocs)

		if len(imageGroups) > 0 {
			// Display grouped images compactly
			displayIdx := 1
			for _, group := range imageGroups {
				fmt.Printf("   %d. ", displayIdx)
				PrintStyledFilename(group.DisplayName)
				fmt.Printf(" ")
				PrintStyledNumber(group.TotalCount)

				// Use "image" or "document" based on the group type
				if group.IsImage {
					if group.TotalCount == 1 {
						fmt.Printf(" image - ")
					} else {
						fmt.Printf(" images - ")
					}
				} else {
					if group.TotalCount == 1 {
						fmt.Printf(" document - ")
					} else {
						fmt.Printf(" documents - ")
					}
				}

				PrintStyledNumber(group.TotalChunks)
				if group.TotalChunks == 1 {
					fmt.Printf(" chunk")
				} else {
					fmt.Printf(" chunks")
				}
				fmt.Println()
				displayIdx++
			}
		} else {
			// Fallback to original display for non-image documents
			for i, vdoc := range virtualDocs {
				fmt.Printf("   %d. ", i+1)
				PrintStyledFilename(vdoc.OriginalFilename)
				fmt.Printf(" - ")
				PrintStyledNumber(vdoc.TotalChunks)
				fmt.Printf(" chunks")
				fmt.Println()
			}
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

		// Check for WeaveDocs schema first (flat metadata structure)
		if filename, ok := doc.Metadata["original_filename"].(string); ok {
			originalFilename = filename
			metadata = doc.Metadata
		} else if filename, ok := doc.Metadata["filename"].(string); ok {
			// Fallback to filename if original_filename is not available
			originalFilename = filename
			metadata = doc.Metadata
		} else {
			// Check if this is a RagMeDocs document with nested metadata (legacy structure)
			if metadataField, ok := doc.Metadata["metadata"]; ok {
				if metadataStr, ok := metadataField.(string); ok {
					var metadataObj map[string]interface{}
					if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
						// Check if there's another nested metadata field that needs parsing
						if nestedMetadataStr, ok := metadataObj["metadata"].(string); ok {
							var nestedMetadataObj map[string]interface{}
							if err := json.Unmarshal([]byte(nestedMetadataStr), &nestedMetadataObj); err == nil {
								// Look for filename in the nested metadata (legacy RagMeDocs structure)
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
							// Look for filename in the parsed metadata (legacy RagMeDocs structure)
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

// DisplayQueryResults displays semantic search results with styling
func DisplayQueryResults(results []weaviate.QueryResult, collectionName, queryText string, noTruncate bool, jsonOutput bool) {
	// Handle JSON output
	if jsonOutput {
		output := map[string]interface{}{
			"collection": collectionName,
			"query":      queryText,
			"results":    results,
			"count":      len(results),
		}
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return
		}
		fmt.Println(string(jsonBytes))
		return
	}

	PrintSuccess(fmt.Sprintf("Semantic search results for '%s' in collection '%s':", queryText, collectionName))
	fmt.Println()

	if len(results) == 0 {
		PrintWarning("No results found for your query.")
		return
	}

	for i, result := range results {
		fmt.Printf("%d. ", i+1)
		PrintStyledEmoji("🔍")
		fmt.Printf(" ")
		PrintStyledKeyProminent("Score")
		fmt.Printf(": ")
		PrintStyledValue(fmt.Sprintf("%.3f", result.Score))
		fmt.Println()

		fmt.Printf("   ")
		PrintStyledKeyProminent("ID")
		fmt.Printf(": ")
		PrintStyledID(result.ID)
		fmt.Println()

		if result.Content != "" {
			fmt.Printf("   ")
			PrintStyledKeyProminent("Content")
			fmt.Printf(": ")
			// Truncate content for better readability unless --no-truncate is specified
			var content string
			if noTruncate {
				content = result.Content
			} else {
				content = SmartTruncate(result.Content, "content", 3)
			}
			PrintStyledValue(content)
			fmt.Println()
		}

		// Show relevant metadata fields
		if len(result.Metadata) > 0 {
			fmt.Printf("   ")
			PrintStyledKeyProminent("Metadata")
			fmt.Printf(": ")

			// Show key metadata fields
			keyFields := []string{"filename", "original_filename", "title", "type", "creation_date"}
			var metadataParts []string

			for _, field := range keyFields {
				if value, exists := result.Metadata[field]; exists && value != "" {
					metadataParts = append(metadataParts, fmt.Sprintf("%s: %v", field, value))
				}
			}

			if len(metadataParts) > 0 {
				PrintStyledValueDimmed(strings.Join(metadataParts, ", "))
			} else {
				PrintStyledValueDimmed(fmt.Sprintf("%d fields", len(result.Metadata)))
			}
			fmt.Println()
		}

		fmt.Println()
	}

	// Show summary
	PrintStyledEmoji("📊")
	fmt.Printf(" ")
	PrintStyledKeyProminent("Summary")
	fmt.Printf(": Found %d results", len(results))
	fmt.Println()

	// Check if all results have low scores (< 0.3 after normalization)
	allLowScores := true
	for _, result := range results {
		if result.Score >= 0.3 {
			allLowScores = false
			break
		}
	}

	// Show score interpretation guidance
	fmt.Println()
	if allLowScores {
		PrintStyledEmoji("⚠️")
		fmt.Printf("  ")
		PrintWarning("All results have low scores (< 0.3) - no good matches found for your query")
		fmt.Println()
		PrintStyledEmoji("💡")
		fmt.Printf("  ")
		PrintStyledValueDimmed("Try rephrasing your query or use different keywords")
		fmt.Println()
	} else {
		PrintStyledEmoji("ℹ️")
		fmt.Printf("  ")
		PrintStyledValueDimmed("Score interpretation: < 0.3 = no match, 0.3-0.5 = weak, 0.5-0.7 = good, > 0.7 = strong")
		fmt.Println()
	}
}
