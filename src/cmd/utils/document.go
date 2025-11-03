// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/image"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"github.com/spf13/viper"
)

// ProcessingReport tracks processing results for CSV reporting
type ProcessingReport struct {
	FilePath   string
	Collection string
	Timestamp  time.Time
	TextChunks []ChunkReport
	Images     []ImageReport
	ReportPath string
	ReportMode string
}

// ChunkReport tracks a single text chunk processing result
type ChunkReport struct {
	ChunkNumber int
	Success     bool
	Error       string
	SizeBytes   int
}

// ImageReport tracks a single image processing result
type ImageReport struct {
	ImageNumber int
	Filename    string
	Success     bool
	Error       string
	SizeBytes   int
	OCRWarnings []string
}

// GenerateDefaultReportPath generates a default CSV report path from input file
func GenerateDefaultReportPath(filePath string) string {
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	return filepath.Join(dir, nameWithoutExt+".csv")
}

// WriteProcessingReport writes or appends a processing report to CSV
func WriteProcessingReport(report *ProcessingReport) error {
	if report.ReportPath == "" {
		return nil // No report requested
	}

	// Determine if we need to create or append
	fileExists := false
	if _, err := os.Stat(report.ReportPath); err == nil {
		fileExists = true
	}

	// Create means overwrite, append means add to existing
	var file *os.File
	var err error
	if report.ReportMode == "create" || !fileExists {
		file, err = os.Create(report.ReportPath)
		if err != nil {
			return fmt.Errorf("failed to create report file: %v", err)
		}
	} else {
		file, err = os.OpenFile(report.ReportPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open report file: %v", err)
		}
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if creating new file
	if report.ReportMode == "create" || !fileExists {
		header := []string{"Timestamp", "File", "Collection", "Type", "Item", "Status", "Error", "Size_KB", "Warnings"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %v", err)
		}
	}

	timestamp := report.Timestamp.Format("2006-01-02 15:04:05")

	// Write text chunk records
	for _, chunk := range report.TextChunks {
		status := "success"
		if !chunk.Success {
			status = "failed"
		}
		record := []string{
			timestamp,
			filepath.Base(report.FilePath),
			report.Collection,
			"text_chunk",
			fmt.Sprintf("chunk_%d", chunk.ChunkNumber),
			status,
			chunk.Error,
			fmt.Sprintf("%.1f", float64(chunk.SizeBytes)/1024),
			"",
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write chunk record: %v", err)
		}
	}

	// Write image records
	for _, img := range report.Images {
		status := "success"
		if !img.Success {
			status = "failed"
		}
		warnings := strings.Join(img.OCRWarnings, "; ")
		record := []string{
			timestamp,
			filepath.Base(report.FilePath),
			report.Collection,
			"image",
			img.Filename,
			status,
			img.Error,
			fmt.Sprintf("%.1f", float64(img.SizeBytes)/1024),
			warnings,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write image record: %v", err)
		}
	}

	return nil
}

// ListWeaviateDocuments lists Weaviate documents
func ListWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, limit int, showLong bool, shortLines int, virtual bool, summary bool) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return
	}

	// Add timeout to context for API operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get documents from collection
	documents, err := client.ListDocuments(ctx, collectionName, limit)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			PrintError(fmt.Sprintf("Failed to connect to Weaviate API at %s: connection timeout after 30 seconds", cfg.URL))
			fmt.Println()
			PrintInfo("Please check that:")
			PrintInfo("  1. The Weaviate server is running and accessible")
			PrintInfo("  2. The URL in your config is correct")
			PrintInfo("  3. There are no network/firewall issues")
		} else {
			PrintError(fmt.Sprintf("Failed to list documents: %v", err))
		}
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
func CreateWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int, batchSize int, reportPath string, reportMode string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create client: %v", err))
		return err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		PrintError(fmt.Sprintf("File not found: %s", filePath))
		return err
	}

	// Initialize report if requested
	var report *ProcessingReport
	if reportPath != "" {
		report = &ProcessingReport{
			FilePath:   filePath,
			Collection: collectionName,
			Timestamp:  time.Now(),
			TextChunks: []ChunkReport{},
			Images:     []ImageReport{},
			ReportPath: reportPath,
			ReportMode: reportMode,
		}
	}

	// Determine file type and process accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	var processErr error
	switch ext {
	case ".pdf":
		processErr = processPDFFile(ctx, client, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, report)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		processErr = processImageFile(ctx, client, collectionName, filePath)
	default:
		processErr = processTextFile(ctx, client, collectionName, filePath, chunkSize)
	}

	// Write report if requested
	if report != nil {
		if err := WriteProcessingReport(report); err != nil {
			PrintWarning(fmt.Sprintf("Failed to write report: %v", err))
		} else {
			PrintSuccess(fmt.Sprintf("Report written to: %s", reportPath))
		}
	}

	return processErr
}

// CreateMockDocument creates a mock document
func CreateMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int, batchSize int, reportPath string, reportMode string) {
	client := CreateMockClient(cfg)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		PrintError(fmt.Sprintf("File not found: %s", filePath))
		return
	}

	// Determine file type and process accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		processPDFFileMock(ctx, client, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		processImageFileMock(ctx, client, collectionName, filePath)
	default:
		processTextFileMock(ctx, client, collectionName, filePath, chunkSize)
	}
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
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create Weaviate client: %v", err))
		return
	}

	var deletedCount int

	// Handle different deletion methods
	if len(documentIDs) > 0 {
		// Delete by document IDs
		var err error
		deletedCount, err = client.DeleteDocumentsBulk(ctx, collectionName, documentIDs)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete documents by IDs: %v", err))
			return
		}
		if deletedCount == 0 {
			PrintError(fmt.Sprintf("Failed to delete any documents. All %d deletion(s) failed. Use --name flag to delete by filename.", len(documentIDs)))
			return
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) by ID", deletedCount))
	} else if len(metadataFilters) > 0 {
		// Delete by metadata filters
		var err error
		deletedCount, err = client.DeleteDocumentsByMetadata(ctx, collectionName, metadataFilters)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete documents by metadata: %v", err))
			return
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) by metadata", deletedCount))
	} else if name != "" {
		// Delete by filename/name - search for documents with matching filename
		filters := []string{fmt.Sprintf("filename=%s", name)}
		var err error
		deletedCount, err = client.DeleteDocumentsByMetadata(ctx, collectionName, filters)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete document by name '%s': %v", name, err))
			return
		}
		if deletedCount == 0 {
			PrintInfo(fmt.Sprintf("No documents found with filename '%s'", name))
		} else {
			PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) with filename '%s'", deletedCount, name))
		}
	} else if pattern != "" {
		// Delete by pattern - get all documents and filter by pattern
		documents, err := client.ListDocuments(ctx, collectionName, 1000) // Get up to 1000 documents
		if err != nil {
			PrintError(fmt.Sprintf("Failed to list documents for pattern matching: %v", err))
			return
		}

		var matchingIDs []string
		for _, doc := range documents {
			// Check if filename matches pattern
			if filename, ok := doc.Metadata["filename"].(string); ok && matchPattern(filename, pattern) {
				matchingIDs = append(matchingIDs, doc.ID)
			}
		}

		if len(matchingIDs) == 0 {
			PrintInfo(fmt.Sprintf("No documents found matching pattern '%s'", pattern))
			return
		}

		// Delete matching documents
		deletedCount, err = client.DeleteDocumentsBulk(ctx, collectionName, matchingIDs)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete documents by pattern '%s': %v", pattern, err))
			return
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) matching pattern '%s'", deletedCount, pattern))
	} else {
		PrintError("No deletion criteria specified. Please provide document IDs, metadata filters, filename, or pattern.")
		return
	}
}

// DeleteMockDocuments deletes mock documents
func DeleteMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string, documentIDs []string, metadataFilters []string, virtual bool, pattern string, name string) {
	client := CreateMockClient(cfg)

	var deletedCount int

	// Handle different deletion methods
	if len(documentIDs) > 0 {
		// Delete by document IDs
		for _, docID := range documentIDs {
			err := client.DeleteDocument(ctx, collectionName, docID)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete document %s: %v", docID, err))
				continue
			}
			deletedCount++
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) by ID", deletedCount))
	} else if len(metadataFilters) > 0 {
		// Delete by metadata filters
		var err error
		deletedCount, err = client.DeleteDocumentsByMetadata(ctx, collectionName, metadataFilters)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete documents by metadata: %v", err))
			return
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) by metadata", deletedCount))
	} else if name != "" {
		// Delete by filename/name - search for documents with matching filename
		filters := []string{fmt.Sprintf("filename=%s", name)}
		var err error
		deletedCount, err = client.DeleteDocumentsByMetadata(ctx, collectionName, filters)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to delete document by name '%s': %v", name, err))
			return
		}
		if deletedCount == 0 {
			PrintInfo(fmt.Sprintf("No documents found with filename '%s'", name))
		} else {
			PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) with filename '%s'", deletedCount, name))
		}
	} else if pattern != "" {
		// Delete by pattern - get all documents and filter by pattern
		documents, err := client.ListDocuments(ctx, collectionName, 1000) // Get up to 1000 documents
		if err != nil {
			PrintError(fmt.Sprintf("Failed to list documents for pattern matching: %v", err))
			return
		}

		var matchingIDs []string
		for _, doc := range documents {
			// Check if filename matches pattern
			if filename, ok := doc.Metadata["filename"].(string); ok && matchPattern(filename, pattern) {
				matchingIDs = append(matchingIDs, doc.ID)
			}
		}

		if len(matchingIDs) == 0 {
			PrintInfo(fmt.Sprintf("No documents found matching pattern '%s'", pattern))
			return
		}

		// Delete matching documents
		for _, docID := range matchingIDs {
			err := client.DeleteDocument(ctx, collectionName, docID)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete document %s: %v", docID, err))
				continue
			}
			deletedCount++
		}
		PrintSuccess(fmt.Sprintf("✅ Successfully deleted %d document(s) matching pattern '%s'", deletedCount, pattern))
	} else {
		PrintError("No deletion criteria specified. Please provide document IDs, metadata filters, filename, or pattern.")
		return
	}
}

// DeleteAllWeaviateDocuments deletes all Weaviate documents in a collection
func DeleteAllWeaviateDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create Weaviate client: %v", err))
		return
	}

	// Delete all documents in the collection
	err = client.DeleteAllDocuments(ctx, collectionName)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to delete all documents from collection '%s': %v", collectionName, err))
		return
	}

	PrintSuccess(fmt.Sprintf("✅ Successfully deleted all documents from collection: %s", collectionName))
}

// DeleteAllMockDocuments deletes all mock documents in a collection
func DeleteAllMockDocuments(ctx context.Context, cfg *config.VectorDBConfig, collectionName string) {
	client := CreateMockClient(cfg)

	// Delete all documents in the collection
	err := client.DeleteAllDocuments(ctx, collectionName)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to delete all documents from collection '%s': %v", collectionName, err))
		return
	}

	PrintSuccess(fmt.Sprintf("✅ Successfully deleted all documents from collection: %s", collectionName))
}

// processTextFile processes a text file and creates documents
func processTextFile(ctx context.Context, client *weaviate.Client, collectionName, filePath string, chunkSize int) error {
	// Convert to absolute path for consistent storage_path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		// Fallback to original path if absolute path cannot be determined
		absPath = filePath
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to read file: %v", err))
		return err
	}

	// Chunk the text content first to get total count
	chunks := chunkText(string(content), chunkSize)

	// Determine if this is a WeaveDocs collection (new schema) or RagMeDocs (legacy)
	isWeaveDocs := isWeaveDocsCollection(collectionName)

	// Get the collection metadata mode - use the same client that will be used for document creation
	metadataMode := getCollectionMetadataMode(ctx, client, collectionName)

	// Create documents for each chunk
	successCount := 0
	for i, chunk := range chunks {
		docID := uuid.New().String()

		var document weaviate.Document
		if isWeaveDocs {
			if metadataMode == "flat" {
				// Use flat metadata structure - individual fields ONLY, ignore auto-generated metadata field
				now := time.Now().Format(time.RFC3339)
				document = weaviate.Document{
					ID:      docID,
					Content: chunk,
					URL:     fmt.Sprintf("file://%s#chunk-%d", absPath, i),
					Metadata: map[string]interface{}{
						"id":                docID,
						"date_added":        now,
						"creation_date":     now,
						"modified_date":     now,
						"creator":           "",
						"producer":          "",
						"title":             filepath.Base(filePath),
						"ai_summary":        "",
						"filename":          filepath.Base(filePath),
						"is_chunked":        len(chunks) > 1,
						"total_chunks":      len(chunks),
						"chunk_index":       i,
						"chunk_sizes":       getChunkSizes(chunks),
						"original_filename": filepath.Base(filePath),
						"storage_path":      absPath,
						"type":              "text",
						"content":           chunk, // Use content field as per Weaviate's expectation
					},
				}
			} else {
				// Use JSON metadata structure - single metadata field ONLY
				// Content is stored in Document.Content, NOT in metadata
				now := time.Now().Format(time.RFC3339)
				metadataJSON := map[string]interface{}{
					"id":                docID,
					"date_added":        now,
					"creation_date":     now,
					"modified_date":     now,
					"creator":           "",
					"producer":          "",
					"title":             filepath.Base(filePath),
					"ai_summary":        "",
					"filename":          filepath.Base(filePath),
					"is_chunked":        len(chunks) > 1,
					"total_chunks":      len(chunks),
					"chunk_index":       i,
					"chunk_sizes":       getChunkSizes(chunks),
					"original_filename": filepath.Base(filePath),
					"storage_path":      absPath,
					"type":              "text",
				}
				// Use the metadataJSON map directly - don't wrap it
				document = weaviate.Document{
					ID:       docID,
					Content:  chunk,
					URL:      fmt.Sprintf("file://%s#chunk-%d", absPath, i),
					Metadata: metadataJSON,
				}
			}
		} else {
			// Use RagMeDocs schema (nested metadata structure for backward compatibility)
			metadataMap := map[string]interface{}{
				"type":       "text",
				"filename":   filepath.Base(filePath),
				"date_added": time.Now().Format(time.RFC3339),
			}

			// Convert metadata to JSON string (like existing RagMeDocs)
			metadataJSON, marshalErr := json.Marshal(metadataMap)
			if marshalErr != nil {
				PrintError(fmt.Sprintf("Failed to marshal metadata: %v", marshalErr))
				continue
			}

			chunkMetadata := map[string]interface{}{
				"metadata": string(metadataJSON),
			}

			document = weaviate.Document{
				ID:       docID,
				Content:  chunk,
				URL:      fmt.Sprintf("file://%s#chunk-%d", filePath, i),
				Metadata: chunkMetadata,
			}
		}

		err := client.CreateDocument(ctx, collectionName, document)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to create document chunk %d: %v", i, err))
			continue
		}
		successCount++
	}

	PrintSuccess(fmt.Sprintf("Successfully created document: %s (%d chunks)", filepath.Base(filePath), successCount))
	return nil
}

// processImageFile processes an image file and creates a document
func processImageFile(ctx context.Context, client *weaviate.Client, collectionName, filePath string) error {
	// Read image file
	imageBytes, err := os.ReadFile(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to read image file: %v", err))
		return err
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get file info: %v", err))
		return err
	}

	// Generate base64 data
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", strings.TrimPrefix(filepath.Ext(filePath), "."), base64Data)

	// Record processing start time (Phase 5)
	processingStartTime := time.Now()

	// Extract EXIF metadata (Phase 1)
	exifData, err := image.ExtractEXIFFromBytes(imageBytes)
	if err != nil {
		PrintWarning(fmt.Sprintf("Failed to extract EXIF data: %v", err))
		exifData = &image.EXIFData{} // Use empty EXIF data
	}

	// Extract OCR text (Phase 3)
	ocrExtractor := image.GetOCRExtractor()
	ocrData, err := ocrExtractor.ExtractFromBytes(imageBytes)
	if err != nil {
		PrintWarning(fmt.Sprintf("Failed to extract OCR text: %v", err))
		ocrData = &image.OCRData{} // Use empty OCR data
	}

	// Generate timestamped storage path (Phase 4)
	timestampedPath := image.GenerateTimestampedPath(filePath)

	// Calculate processing duration (Phase 5)
	processingDuration := time.Since(processingStartTime)

	// Determine if this is a WeaveImages collection (new schema) or RagMeImages (legacy)
	isWeaveImages := isWeaveImagesCollection(collectionName)

	// Create document
	docID := uuid.New().String()
	var document weaviate.Document

	if isWeaveImages {
		// Use WeaveImages schema (flat metadata structure)
		now := time.Now().Format(time.RFC3339)

		// Build metadata with EXIF, OCR, and processing data
		metadata := map[string]interface{}{
			"id":                     docID,
			"date_added":             now,
			"creation_date":          now,
			"modified_date":          now,
			"creator":                "",
			"producer":               "",
			"title":                  filepath.Base(filePath),
			"ai_summary":             "",
			"filename":               filepath.Base(filePath),
			"is_chunked":             false,
			"total_chunks":           1,
			"chunk_index":            0,
			"chunk_sizes":            []int{len(imageBytes)},
			"original_filename":      filepath.Base(filePath),
			"storage_path":           filePath,        // Absolute path (original)
			"storage_path_relative":  timestampedPath, // Phase 4: Timestamped relative path
			"type":                   "image",
			"file_size":              fileInfo.Size(),
			"image_format":           strings.ToLower(filepath.Ext(filePath)),
			"image_size":             len(imageBytes),
			"content":                "",                                // Images don't have text content
			"processing_timestamp":   now,                               // Phase 5: When processed
			"processing_duration_ms": processingDuration.Milliseconds(), // Phase 5: How long it took
		}

		// Add EXIF data if available
		if !exifData.IsEmpty() {
			if exifData.Make != "" {
				metadata["exif_make"] = exifData.Make
			}
			if exifData.Model != "" {
				metadata["exif_model"] = exifData.Model
			}
			if exifData.DateTime != "" {
				metadata["exif_datetime"] = exifData.DateTime
			}
			if exifData.Orientation != 0 {
				metadata["exif_orientation"] = exifData.Orientation
			}
			if exifData.Width != 0 {
				metadata["exif_width"] = exifData.Width
			}
			if exifData.Height != 0 {
				metadata["exif_height"] = exifData.Height
			}
			if exifData.FNumber != 0 {
				metadata["exif_f_number"] = exifData.FNumber
			}
			if exifData.ExposureTime != "" {
				metadata["exif_exposure_time"] = exifData.ExposureTime
			}
			if exifData.ISO != 0 {
				metadata["exif_iso"] = exifData.ISO
			}
			if exifData.FocalLength != "" {
				metadata["exif_focal_length"] = exifData.FocalLength
			}
			if exifData.GPSLatitude != 0 || exifData.GPSLongitude != 0 {
				metadata["exif_gps_latitude"] = exifData.GPSLatitude
				metadata["exif_gps_longitude"] = exifData.GPSLongitude
			}
			if exifData.GPSAltitude != 0 {
				metadata["exif_gps_altitude"] = exifData.GPSAltitude
			}
		}

		// Add OCR data if text was extracted (Phase 3)
		if !ocrData.IsEmpty() {
			metadata["ocr_text"] = ocrData.Text
			metadata["ocr_confidence"] = ocrData.Confidence
			metadata["ocr_language"] = ocrData.Language
			metadata["ocr_word_count"] = ocrData.WordCount()
			metadata["ocr_text_summary"] = ocrData.GetTextSummary(200) // First 200 chars
		}

		document = weaviate.Document{
			ID:        docID,
			Image:     dataURL,
			ImageData: base64Data,
			URL:       fmt.Sprintf("file://%s", filePath),
			Metadata:  metadata,
		}
	} else {
		// Use RagMeImages schema (legacy structure for backward compatibility)
		metadata := map[string]interface{}{
			"type":         "image",
			"filename":     filepath.Base(filePath),
			"date_added":   time.Now().Format(time.RFC3339),
			"storage_path": filePath,
			"file_size":    fileInfo.Size(),
			"image_format": strings.ToLower(filepath.Ext(filePath)),
			"image_size":   len(imageBytes),
		}

		document = weaviate.Document{
			ID:        docID,
			Image:     dataURL,
			ImageData: base64Data,
			URL:       fmt.Sprintf("file://%s", filePath),
			Metadata:  metadata,
		}
	}

	err = client.CreateDocument(ctx, collectionName, document)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create image document: %v", err))
		return err
	}

	PrintSuccess(fmt.Sprintf("Successfully created image document: %s", filepath.Base(filePath)))
	return nil
}

// processPDFFile processes a PDF file and creates documents
func processPDFFile(ctx context.Context, client *weaviate.Client, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int, batchSize int, report *ProcessingReport) error {
	fmt.Printf("📄 Processing PDF: %s\n", filepath.Base(filePath))

	// Extract PDF content using the existing PDF processor
	fmt.Println("🔍 Extracting content from PDF...")
	noTips := viper.GetBool("no-tips")
	textData, imageData, err := pdf.ExtractPDFContent(filePath, chunkSize, skipSmallImages, minImageSize, noTips)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to extract PDF content: %v", err))
		return err
	}

	fmt.Printf("✅ Found %d text chunks and %d images\n", len(textData), len(imageData))

	// Create text documents with progress
	textSuccessCount := 0
	if len(textData) > 0 {
		fmt.Printf("\n📝 Creating text documents (%d chunks):\n", len(textData))
		for i, textDoc := range textData {
			document := weaviate.Document{
				ID:       textDoc.ID,
				Content:  textDoc.Content,
				URL:      textDoc.URL,
				Metadata: textDoc.Metadata,
			}

			err := client.CreateDocument(ctx, collectionName, document)

			// Track in report if requested
			if report != nil {
				chunkReport := ChunkReport{
					ChunkNumber: i + 1,
					Success:     err == nil,
					Error:       "",
					SizeBytes:   len(textDoc.Content),
				}
				if err != nil {
					chunkReport.Error = err.Error()
				}
				report.TextChunks = append(report.TextChunks, chunkReport)
			}

			if err != nil {
				PrintError(fmt.Sprintf("Failed to create PDF text document %s: %v", textDoc.ID, err))
				continue
			}
			textSuccessCount++

			// Progress indicator
			if (i+1)%5 == 0 || i == len(textData)-1 {
				fmt.Printf("  [%d/%d] chunks created\n", textSuccessCount, len(textData))
			}
		}
	}

	// Create image documents if image collection is specified
	imageSuccessCount := 0
	imageFailureCount := 0
	if imageCollection != "" && len(imageData) > 0 {
		fmt.Printf("\n🖼️  Processing extracted images (%d total):\n", len(imageData))

		// Process images in batches to avoid memory issues with large PDFs
		for i := 0; i < len(imageData); i++ {
			imageDoc := imageData[i]

			// Apply full image processing pipeline to extracted images
			// Pass i+1 as sequential index for clean numbering (image_1, image_2, etc.)
			processedDoc, ocrWarnings, err := processExtractedPDFImage(ctx, imageDoc, imageCollection, filePath, i+1)

			// Get filename for reporting
			filename := "unknown"
			if fn, ok := processedDoc.Metadata["filename"].(string); ok && fn != "" {
				filename = fn
			}

			if err != nil {
				PrintError(fmt.Sprintf("Failed to process PDF image %d: %v", i+1, err))

				// Track in report if requested
				if report != nil {
					imgReport := ImageReport{
						ImageNumber: i + 1,
						Filename:    filename,
						Success:     false,
						Error:       err.Error(),
						SizeBytes:   len(imageDoc.ImageData),
						OCRWarnings: ocrWarnings,
					}
					report.Images = append(report.Images, imgReport)
				}
				continue
			}

			err = client.CreateDocument(ctx, imageCollection, processedDoc)

			// Track in report if requested
			if report != nil {
				imgReport := ImageReport{
					ImageNumber: i + 1,
					Filename:    filename,
					Success:     err == nil,
					Error:       "",
					SizeBytes:   len(imageDoc.ImageData),
					OCRWarnings: ocrWarnings,
				}
				if err != nil {
					imgReport.Error = err.Error()
				}
				report.Images = append(report.Images, imgReport)
			}

			if err != nil {
				imageFailureCount++
				// Check if it's a memory error and provide helpful message
				if strings.Contains(err.Error(), "not enough memory") || strings.Contains(err.Error(), "memory") {
					PrintError(fmt.Sprintf("Failed to create PDF image document %s: %v", imageDoc.ID, err))
					if imageFailureCount == 1 && !viper.GetBool("no-tips") {
						PrintWarning("💡 Tip: Try reducing batch size with --batch-size flag (e.g., --batch-size 5)")
						PrintWarning("💡 Or increase min-image-size to skip large images (e.g., --min-image-size 10240)")
					}
				} else {
					PrintError(fmt.Sprintf("Failed to create PDF image document %s: %v", imageDoc.ID, err))
				}
				continue
			}
			imageSuccessCount++

			// Progress indicator
			fmt.Printf("  [%d/%d] Image %d: %s (%.1f KB)\n",
				imageSuccessCount, len(imageData), i+1,
				filename,
				float64(len(imageDoc.ImageData))/1024)

			// Force garbage collection every batch to free memory
			// Also add small delay to let Weaviate server process and free memory
			if (i+1)%batchSize == 0 {
				runtime.GC()
				time.Sleep(500 * time.Millisecond) // Half-second pause between batches
			}
		}

		// Final garbage collection
		runtime.GC()
	}

	fmt.Println()
	PrintSuccess(fmt.Sprintf("✅ PDF processed: %s", filepath.Base(filePath)))
	fmt.Printf("   • Text chunks: %d/%d created\n", textSuccessCount, len(textData))
	if imageCollection != "" {
		fmt.Printf("   • Images: %d/%d created", imageSuccessCount, len(imageData))
		if imageFailureCount > 0 {
			fmt.Printf(" (%d failed)\n", imageFailureCount)
		} else {
			fmt.Println()
		}
	}

	return nil
}

// processExtractedPDFImage processes an extracted PDF image with full metadata pipeline
func processExtractedPDFImage(ctx context.Context, imageDoc pdf.PDFImageData, collectionName, sourcePDF string, sequentialIndex int) (weaviate.Document, []string, error) {
	// Decode base64 image data
	imageBytes, err := base64.StdEncoding.DecodeString(imageDoc.ImageData)
	if err != nil {
		return weaviate.Document{}, nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	// Record processing start time
	processingStartTime := time.Now()

	// Track OCR warnings
	var ocrWarnings []string

	// Extract EXIF metadata
	exifData, err := image.ExtractEXIFFromBytes(imageBytes)
	if err != nil {
		// Non-critical, continue with empty EXIF
		exifData = &image.EXIFData{}
	}

	// Extract OCR text
	ocrExtractor := image.GetOCRExtractor()
	ocrData, err := ocrExtractor.ExtractFromBytes(imageBytes)
	if err != nil {
		// Non-critical, continue with empty OCR
		ocrData = &image.OCRData{}
		ocrWarnings = append(ocrWarnings, fmt.Sprintf("OCR extraction failed: %v", err))
	}

	// Check for common OCR warnings in the error message (these come from tesseract stderr)
	if err != nil && strings.Contains(err.Error(), "Image too small") {
		ocrWarnings = append(ocrWarnings, "Image too small to scale")
	}
	if err != nil && strings.Contains(err.Error(), "Line cannot be recognized") {
		ocrWarnings = append(ocrWarnings, "Line cannot be recognized")
	}

	// Generate timestamped storage path
	timestampedPath := image.GenerateTimestampedPathWithPrefix(sourcePDF, "images")

	// Calculate processing duration
	processingDuration := time.Since(processingStartTime)

	// Determine if this is WeaveImages or RagMeImages
	isWeaveImages := isWeaveImagesCollection(collectionName)

	// Create RagMe-compatible URL format for grouping
	// Format: pdf://filename.pdf/image_N (using sequential index for clean numbering)
	pdfFilename := filepath.Base(sourcePDF)
	pdfName := strings.TrimSuffix(pdfFilename, filepath.Ext(pdfFilename))

	// Generate sequential filename (e.g., ragme-io_image_1.png)
	imageExt := ".png" // Default to PNG for PDF-extracted images
	if format, ok := imageDoc.Metadata["image_format"].(string); ok && format != "" {
		imageExt = format
		if !strings.HasPrefix(imageExt, ".") {
			imageExt = "." + imageExt
		}
	}
	sequentialFilename := fmt.Sprintf("%s_image_%d%s", pdfName, sequentialIndex, imageExt)

	// Use RagMe-compatible URL format with sequential numbering
	ragmeURL := fmt.Sprintf("pdf://%s/image_%d", pdfFilename, sequentialIndex)

	now := time.Now().Format(time.RFC3339)

	if isWeaveImages {
		// Build comprehensive metadata with all enhancements

		metadata := map[string]interface{}{
			"id":                     imageDoc.ID,
			"date_added":             now,
			"creation_date":          now,
			"modified_date":          now,
			"creator":                "PDF Extraction",
			"producer":               "Weave CLI",
			"title":                  fmt.Sprintf("Image %d from %s", sequentialIndex, pdfFilename),
			"ai_summary":             "",
			"filename":               sequentialFilename,
			"is_chunked":             false,
			"total_chunks":           1,
			"chunk_index":            0,
			"chunk_sizes":            []int{len(imageBytes)},
			"original_filename":      sequentialFilename,
			"storage_path":           imageDoc.Metadata["source_pdf"],
			"storage_path_relative":  timestampedPath,
			"type":                   "image",
			"source_pdf":             sourcePDF,
			"pdf_filename":           pdfFilename, // RAGme-io compatibility
			"pdf_page":               imageDoc.Metadata["pdf_page"],
			"pdf_image_index":        sequentialIndex, // Use sequential index
			"file_size":              len(imageBytes),
			"image_format":           imageDoc.Metadata["image_format"],
			"image_size":             len(imageBytes),
			"content":                "",
			"processing_timestamp":   now,
			"processing_duration_ms": processingDuration.Milliseconds(),
		}

		// Add EXIF data
		if !exifData.IsEmpty() {
			if exifData.Make != "" {
				metadata["exif_make"] = exifData.Make
			}
			if exifData.Model != "" {
				metadata["exif_model"] = exifData.Model
			}
			if exifData.DateTime != "" {
				metadata["exif_datetime"] = exifData.DateTime
			}
			if exifData.Orientation != 0 {
				metadata["exif_orientation"] = exifData.Orientation
			}
			if exifData.Width != 0 {
				metadata["exif_width"] = exifData.Width
			}
			if exifData.Height != 0 {
				metadata["exif_height"] = exifData.Height
			}
		}

		// Add OCR data
		if !ocrData.IsEmpty() {
			metadata["ocr_text"] = ocrData.Text
			metadata["ocr_confidence"] = ocrData.Confidence
			metadata["ocr_language"] = ocrData.Language
			metadata["ocr_word_count"] = ocrData.WordCount()
			metadata["ocr_text_summary"] = ocrData.GetTextSummary(200)
		}

		return weaviate.Document{
			ID:        imageDoc.ID,
			Image:     imageDoc.Image,
			ImageData: imageDoc.ImageData,
			URL:       ragmeURL, // Use RagMe-compatible URL format
			Metadata:  metadata,
		}, ocrWarnings, nil
	}

	// RagMeImages format (legacy)
	return weaviate.Document{
		ID:        imageDoc.ID,
		Image:     imageDoc.Image,
		ImageData: imageDoc.ImageData,
		URL:       imageDoc.URL,
		Metadata:  imageDoc.Metadata,
	}, ocrWarnings, nil
}

// chunkText splits text into chunks of specified size
func chunkText(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	currentChunk := ""
	currentSize := 0

	for _, line := range lines {
		lineSize := len(line) + 1 // +1 for newline

		if currentSize+lineSize > chunkSize && currentChunk != "" {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = line + "\n"
			currentSize = lineSize
		} else {
			currentChunk += line + "\n"
			currentSize += lineSize
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	// If no chunks were created, return the original text
	if len(chunks) == 0 {
		chunks = []string{text}
	}

	return chunks
}

// isWeaveDocsCollection determines if a collection uses the WeaveDocs schema
func isWeaveDocsCollection(collectionName string) bool {
	// WeaveDocs is the default for --text collections
	// Check if collection name suggests it's using the new schema
	weaveDocsNames := []string{"WeaveDocs", "weavedocs", "weave-docs", "weave_docs"}
	for _, name := range weaveDocsNames {
		if strings.EqualFold(collectionName, name) {
			return true
		}
	}

	// Default to WeaveDocs schema for new collections
	// Legacy RagMeDocs collections should be explicitly named
	ragMeDocsNames := []string{"RagMeDocs", "ragmedocs", "ragme-docs", "ragme_docs"}
	for _, name := range ragMeDocsNames {
		if strings.EqualFold(collectionName, name) {
			return false
		}
	}

	// Default to new WeaveDocs schema
	return true
}

// getCollectionMetadataMode determines if a collection uses flat or JSON metadata
func getCollectionMetadataMode(ctx context.Context, client *weaviate.Client, collectionName string) string {
	// Get the collection schema
	schema, err := client.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		// If we can't get the schema, default to flat metadata
		return "flat"
	}

	// Check if the collection has individual metadata fields
	hasFlatMetadataFields := false
	for _, prop := range schema.Properties {
		if prop.Name == "is_chunked" || prop.Name == "total_chunks" || prop.Name == "chunk_index" {
			hasFlatMetadataFields = true
			break
		}
	}

	if hasFlatMetadataFields {
		return "flat"
	}

	// Check if it has a metadata field (JSON metadata)
	for _, prop := range schema.Properties {
		if prop.Name == "metadata" {
			return "json"
		}
	}

	// Default to flat metadata
	return "flat"
}

// isWeaveImagesCollection determines if a collection uses the WeaveImages schema
func isWeaveImagesCollection(collectionName string) bool {
	// WeaveImages is the default for --image collections
	// Check if collection name suggests it's using the new schema
	weaveImagesNames := []string{"WeaveImages", "weaveimages", "weave-images", "weave_images"}
	for _, name := range weaveImagesNames {
		if strings.EqualFold(collectionName, name) {
			return true
		}
	}

	// Default to WeaveImages schema for new collections
	// Legacy RagMeImages collections should be explicitly named
	ragMeImagesNames := []string{"RagMeImages", "ragmeimages", "ragme-images", "ragme_images"}
	for _, name := range ragMeImagesNames {
		if strings.EqualFold(collectionName, name) {
			return false
		}
	}

	// Default to new WeaveImages schema
	return true
}

// getChunkSizes returns an array of chunk sizes
func getChunkSizes(chunks []string) []int {
	sizes := make([]int, len(chunks))
	for i, chunk := range chunks {
		sizes[i] = len(chunk)
	}
	return sizes
}

// Mock helper functions
func processTextFileMock(ctx context.Context, client *mock.Client, collectionName, filePath string, chunkSize int) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to read file: %v", err))
		return
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get file info: %v", err))
		return
	}

	// Chunk the text content
	chunks := chunkText(string(content), chunkSize)

	// Determine if this is a WeaveDocs collection (new schema) or RagMeDocs (legacy)
	isWeaveDocs := isWeaveDocsCollection(collectionName)

	// Create documents for each chunk
	successCount := 0
	for i, chunk := range chunks {
		docID := fmt.Sprintf("%s-chunk-%d", filepath.Base(filePath), i)

		var document mock.Document
		if isWeaveDocs {
			// Use WeaveDocs schema (flat metadata structure)
			document = mock.Document{
				ID:      docID,
				Content: chunk,
				URL:     fmt.Sprintf("file://%s#chunk-%d", filePath, i),
				Metadata: map[string]interface{}{
					"id":                docID,
					"added_date":        time.Now().Format(time.RFC3339),
					"creation_date":     time.Now().Format(time.RFC3339),
					"modified_date":     time.Now().Format(time.RFC3339),
					"creator":           "",
					"producer":          "",
					"title":             filepath.Base(filePath),
					"ai_summary":        "",
					"filename":          filepath.Base(filePath),
					"is_chunked":        len(chunks) > 1,
					"total_chunks":      len(chunks),
					"chunk_index":       i,
					"chunk_sizes":       getChunkSizes(chunks),
					"original_filename": filepath.Base(filePath),
					"storage_path":      filePath,
					"type":              "text",
				},
			}
		} else {
			// Use RagMeDocs schema (legacy structure for backward compatibility)
			metadata := map[string]interface{}{
				"type":              "text",
				"filename":          filepath.Base(filePath),
				"original_filename": filepath.Base(filePath),
				"date_added":        time.Now().Format(time.RFC3339),
				"storage_path":      filePath,
				"file_size":         fileInfo.Size(),
				"is_chunked":        true,
			}

			document = mock.Document{
				ID:       docID,
				Content:  chunk,
				URL:      fmt.Sprintf("file://%s#chunk-%d", filePath, i),
				Metadata: metadata,
			}
		}

		err := client.CreateDocument(ctx, collectionName, document)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to create document chunk %d: %v", i, err))
			continue
		}
		successCount++
	}

	PrintSuccess(fmt.Sprintf("Successfully created document: %s (%d chunks)", filepath.Base(filePath), successCount))
}

func processImageFileMock(ctx context.Context, client *mock.Client, collectionName, filePath string) {
	// Read image file
	imageBytes, err := os.ReadFile(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to read image file: %v", err))
		return
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to get file info: %v", err))
		return
	}

	// Generate base64 data
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", strings.TrimPrefix(filepath.Ext(filePath), "."), base64Data)

	// Determine if this is a WeaveImages collection (new schema) or RagMeImages (legacy)
	isWeaveImages := isWeaveImagesCollection(collectionName)

	// Create document
	docID := fmt.Sprintf("%s-image", filepath.Base(filePath))
	var document mock.Document

	if isWeaveImages {
		// Use WeaveImages schema (flat metadata structure)
		document = mock.Document{
			ID:        docID,
			Image:     dataURL,
			ImageData: base64Data,
			URL:       fmt.Sprintf("file://%s", filePath),
			Metadata: map[string]interface{}{
				"id":                docID,
				"added_date":        time.Now().Format(time.RFC3339),
				"creation_date":     time.Now().Format(time.RFC3339),
				"modified_date":     time.Now().Format(time.RFC3339),
				"creator":           "",
				"producer":          "",
				"title":             filepath.Base(filePath),
				"ai_summary":        "",
				"filename":          filepath.Base(filePath),
				"is_chunked":        false,
				"total_chunks":      1,
				"chunk_index":       0,
				"chunk_sizes":       []int{len(imageBytes)},
				"original_filename": filepath.Base(filePath),
				"storage_path":      filePath,
				"type":              "image",
				"file_size":         fileInfo.Size(),
				"image_format":      strings.ToLower(filepath.Ext(filePath)),
				"image_size":        len(imageBytes),
			},
		}
	} else {
		// Use RagMeImages schema (legacy structure for backward compatibility)
		metadata := map[string]interface{}{
			"type":         "image",
			"filename":     filepath.Base(filePath),
			"date_added":   time.Now().Format(time.RFC3339),
			"storage_path": filePath,
			"file_size":    fileInfo.Size(),
			"image_format": strings.ToLower(filepath.Ext(filePath)),
			"image_size":   len(imageBytes),
		}

		document = mock.Document{
			ID:        docID,
			Image:     dataURL,
			ImageData: base64Data,
			URL:       fmt.Sprintf("file://%s", filePath),
			Metadata:  metadata,
		}
	}

	err = client.CreateDocument(ctx, collectionName, document)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create image document: %v", err))
		return
	}

	PrintSuccess(fmt.Sprintf("Successfully created image document: %s", filepath.Base(filePath)))
}

func processPDFFileMock(ctx context.Context, client *mock.Client, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
	// Extract PDF content using the existing PDF processor
	noTips := viper.GetBool("no-tips")
	textData, imageData, err := pdf.ExtractPDFContent(filePath, chunkSize, skipSmallImages, minImageSize, noTips)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to extract PDF content: %v", err))
		return
	}

	// Create text documents
	textSuccessCount := 0
	for _, textDoc := range textData {
		document := mock.Document{
			ID:       textDoc.ID,
			Content:  textDoc.Content,
			URL:      textDoc.URL,
			Metadata: textDoc.Metadata,
		}

		err := client.CreateDocument(ctx, collectionName, document)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to create PDF text document %s: %v", textDoc.ID, err))
			continue
		}
		textSuccessCount++
	}

	// Create image documents if image collection is specified
	imageSuccessCount := 0
	if imageCollection != "" && len(imageData) > 0 {
		for _, imageDoc := range imageData {
			document := mock.Document{
				ID:        imageDoc.ID,
				Image:     imageDoc.Image,
				ImageData: imageDoc.ImageData,
				URL:       imageDoc.URL,
				Metadata:  imageDoc.Metadata,
			}

			err := client.CreateDocument(ctx, imageCollection, document)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to create PDF image document %s: %v", imageDoc.ID, err))
				continue
			}
			imageSuccessCount++
		}
	}

	PrintSuccess(fmt.Sprintf("Successfully created PDF document: %s (%d text chunks)", filepath.Base(filePath), textSuccessCount))
	if imageSuccessCount > 0 {
		fmt.Printf(", %d images)", imageSuccessCount)
	}
	fmt.Println()
}

// UpdateWeaviateDocument updates an existing document in Weaviate
func UpdateWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID, docName string, metadataFilter []string, content, filePath string, metadata []string) error {
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Parse metadata updates
	metadataUpdates := make(map[string]interface{})
	for _, m := range metadata {
		parts := strings.SplitN(m, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid metadata format: %s (expected key=value)", m)
		}
		metadataUpdates[parts[0]] = parts[1]
	}

	// Get new content if file path is specified
	var newContent string
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		newContent = string(data)
	} else if content != "" {
		newContent = content
	}

	// Find document by ID, name, or metadata filter
	var targetDocID string
	if documentID != "" {
		targetDocID = documentID
	} else if docName != "" {
		// Find document by name
		docs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}
		for _, doc := range docs {
			if filename, ok := doc.Metadata["filename"].(string); ok && filename == docName {
				targetDocID = doc.ID
				break
			}
			if origFilename, ok := doc.Metadata["original_filename"].(string); ok && origFilename == docName {
				targetDocID = doc.ID
				break
			}
		}
		if targetDocID == "" {
			return fmt.Errorf("document with name '%s' not found", docName)
		}
	} else if len(metadataFilter) > 0 {
		// Find document by metadata filter
		docs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}
		for _, doc := range docs {
			match := true
			for _, filter := range metadataFilter {
				parts := strings.SplitN(filter, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid metadata filter: %s", filter)
				}
				key, value := parts[0], parts[1]
				if docValue, ok := doc.Metadata[key]; !ok || fmt.Sprint(docValue) != value {
					match = false
					break
				}
			}
			if match {
				targetDocID = doc.ID
				break
			}
		}
		if targetDocID == "" {
			return fmt.Errorf("no document found matching metadata filters")
		}
	}

	// Update the document
	if err := client.UpdateDocument(ctx, collectionName, targetDocID, newContent, metadataUpdates); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	PrintSuccess(fmt.Sprintf("Successfully updated document: %s", targetDocID))
	if newContent != "" {
		fmt.Printf("  Content updated (%d characters)\n", len(newContent))
	}
	if len(metadataUpdates) > 0 {
		fmt.Printf("  Metadata updated (%d fields)\n", len(metadataUpdates))
		for key, value := range metadataUpdates {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	return nil
}

// UpdateMockDocument updates a document in the mock database
func UpdateMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, documentID, docName string, metadataFilter []string, content, filePath string, metadata []string) {
	client := CreateMockClient(cfg)

	// Parse metadata updates
	metadataUpdates := make(map[string]interface{})
	for _, m := range metadata {
		parts := strings.SplitN(m, "=", 2)
		if len(parts) != 2 {
			PrintError(fmt.Sprintf("Invalid metadata format: %s", m))
			return
		}
		metadataUpdates[parts[0]] = parts[1]
	}

	// Get new content if file path is specified
	var newContent string
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to read file: %v", err))
			return
		}
		newContent = string(data)
	} else if content != "" {
		newContent = content
	}

	// Find document by ID, name, or metadata filter
	var targetDocID string
	if documentID != "" {
		targetDocID = documentID
	} else if docName != "" {
		// Find document by name
		docs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to list documents: %v", err))
			return
		}
		for _, doc := range docs {
			if filename, ok := doc.Metadata["filename"].(string); ok && filename == docName {
				targetDocID = doc.ID
				break
			}
			if origFilename, ok := doc.Metadata["original_filename"].(string); ok && origFilename == docName {
				targetDocID = doc.ID
				break
			}
		}
		if targetDocID == "" {
			PrintError(fmt.Sprintf("Document with name '%s' not found", docName))
			return
		}
	} else if len(metadataFilter) > 0 {
		// Find document by metadata filter
		docs, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			PrintError(fmt.Sprintf("Failed to list documents: %v", err))
			return
		}
		for _, doc := range docs {
			match := true
			for _, filter := range metadataFilter {
				parts := strings.SplitN(filter, "=", 2)
				if len(parts) != 2 {
					PrintError(fmt.Sprintf("Invalid metadata filter: %s", filter))
					return
				}
				key, value := parts[0], parts[1]
				if docValue, ok := doc.Metadata[key]; !ok || fmt.Sprint(docValue) != value {
					match = false
					break
				}
			}
			if match {
				targetDocID = doc.ID
				break
			}
		}
		if targetDocID == "" {
			PrintError("No document found matching metadata filters")
			return
		}
	}

	// Update the document
	if err := client.UpdateDocument(ctx, collectionName, targetDocID, newContent, metadataUpdates); err != nil {
		PrintError(fmt.Sprintf("Failed to update document: %v", err))
		return
	}

	PrintSuccess(fmt.Sprintf("Successfully updated document: %s", targetDocID))
	if newContent != "" {
		fmt.Printf("  Content updated (%d characters)\n", len(newContent))
	}
	if len(metadataUpdates) > 0 {
		fmt.Printf("  Metadata updated (%d fields)\n", len(metadataUpdates))
		for key, value := range metadataUpdates {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}
}
