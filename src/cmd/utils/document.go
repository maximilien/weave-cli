// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
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
func CreateWeaviateDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) error {
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

	// Determine file type and process accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	var processErr error
	switch ext {
	case ".pdf":
		processErr = processPDFFile(ctx, client, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		processErr = processImageFile(ctx, client, collectionName, filePath)
	default:
		processErr = processTextFile(ctx, client, collectionName, filePath, chunkSize)
	}

	return processErr
}

// CreateMockDocument creates a mock document
func CreateMockDocument(ctx context.Context, cfg *config.VectorDBConfig, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) {
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
	} else {
		PrintError("No deletion criteria specified. Please provide document IDs, metadata filters, or filename.")
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
	} else {
		PrintError("No deletion criteria specified. Please provide document IDs, metadata filters, or filename.")
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
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to read file: %v", err))
		return err
	}

	// Chunk the text content first to get total count
	chunks := chunkText(string(content), chunkSize)

	// Create documents for each chunk
	successCount := 0
	for i, chunk := range chunks {
		docID := uuid.New().String()

		// Create metadata as a JSON string (exactly like existing RagMeDocs)
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

		document := weaviate.Document{
			ID:       docID,
			Content:  chunk,
			URL:      fmt.Sprintf("file://%s#chunk-%d", filePath, i),
			Metadata: chunkMetadata,
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

	// Generate metadata
	metadata := map[string]interface{}{
		"type":         "image",
		"filename":     filepath.Base(filePath),
		"date_added":   time.Now().Format(time.RFC3339),
		"storage_path": filePath,
		"file_size":    fileInfo.Size(),
		"image_format": strings.ToLower(filepath.Ext(filePath)),
		"image_size":   len(imageBytes),
	}

	// Create document
	docID := uuid.New().String()
	document := weaviate.Document{
		ID:        docID,
		Image:     dataURL,
		ImageData: base64Data,
		URL:       fmt.Sprintf("file://%s", filePath),
		Metadata:  metadata,
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
func processPDFFile(ctx context.Context, client *weaviate.Client, collectionName, filePath string, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int) error {
	// Extract PDF content using the existing PDF processor
	textData, imageData, err := pdf.ExtractPDFContent(filePath, chunkSize, skipSmallImages, minImageSize)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to extract PDF content: %v", err))
		return err
	}

	// Create text documents
	textSuccessCount := 0
	for _, textDoc := range textData {
		document := weaviate.Document{
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
			document := weaviate.Document{
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

	PrintSuccess(fmt.Sprintf("Successfully created PDF document: %s (%d text chunks", filepath.Base(filePath), textSuccessCount))
	if imageSuccessCount > 0 {
		fmt.Printf(", %d images)", imageSuccessCount)
	}
	fmt.Println()
	return nil
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

	// Generate metadata
	metadata := map[string]interface{}{
		"type":              "text",
		"filename":          filepath.Base(filePath),
		"original_filename": filepath.Base(filePath),
		"date_added":        time.Now().Format(time.RFC3339),
		"storage_path":      filePath,
		"file_size":         fileInfo.Size(),
		"is_chunked":        true,
	}

	// Chunk the text content
	chunks := chunkText(string(content), chunkSize)

	// Create documents for each chunk
	successCount := 0
	for i, chunk := range chunks {
		docID := fmt.Sprintf("%s-chunk-%d", filepath.Base(filePath), i)

		document := mock.Document{
			ID:       docID,
			Content:  chunk,
			URL:      fmt.Sprintf("file://%s#chunk-%d", filePath, i),
			Metadata: metadata,
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

	// Generate metadata
	metadata := map[string]interface{}{
		"type":         "image",
		"filename":     filepath.Base(filePath),
		"date_added":   time.Now().Format(time.RFC3339),
		"storage_path": filePath,
		"file_size":    fileInfo.Size(),
		"image_format": strings.ToLower(filepath.Ext(filePath)),
		"image_size":   len(imageBytes),
	}

	// Create document
	docID := fmt.Sprintf("%s-image", filepath.Base(filePath))
	document := mock.Document{
		ID:        docID,
		Image:     dataURL,
		ImageData: base64Data,
		URL:       fmt.Sprintf("file://%s", filePath),
		Metadata:  metadata,
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
	textData, imageData, err := pdf.ExtractPDFContent(filePath, chunkSize, skipSmallImages, minImageSize)
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

	PrintSuccess(fmt.Sprintf("Successfully created PDF document: %s (%d text chunks", filepath.Base(filePath), textSuccessCount))
	if imageSuccessCount > 0 {
		fmt.Printf(", %d images)", imageSuccessCount)
	}
	fmt.Println()
}
