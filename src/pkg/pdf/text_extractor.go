// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// extractPDFText extracts text content from a PDF file
func extractPDFText(filePath string, chunkSize int) ([]PDFTextData, error) {
	// Extract raw text from PDF using pdfcpu
	text, pageCount, err := extractTextFromPDF(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract PDF metadata
	pdfMetadata, err := extractPDFMetadata(filePath)
	if err != nil {
		// Non-fatal, continue with empty metadata
		pdfMetadata = make(map[string]string)
	}

	// If text extraction failed or returned empty, generate fallback content
	if text == "" || len(strings.TrimSpace(text)) < 10 {
		text = generateFallbackPDFContent(filepath.Base(filePath), fileInfo.Size(), pageCount)
	}

	// Chunk the text - default to 1000 chars if not specified
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	chunks := chunkText(text, chunkSize)

	// Calculate chunk sizes for metadata (compatible with RagMeDocs)
	chunkSizes := make([]int, len(chunks))
	for i, chunk := range chunks {
		chunkSizes[i] = len(chunk)
	}

	// Create PDFTextData objects with RagMeDocs-compatible metadata
	var textData []PDFTextData
	fileName := filepath.Base(filePath)

	for i, chunk := range chunks {
		// Build metadata compatible with RagMeDocs format
		metadata := map[string]interface{}{
			"author":            pdfMetadata["Author"], // PDF author if available
			"chunk_index":       i,                     // Current chunk index
			"chunk_sizes":       chunkSizes,            // Array of all chunk sizes
			"source_document":   fileName,              // Original filename
			"processing_info":   fmt.Sprintf("Processed on %s", time.Now().Format(time.RFC3339)),
			"type":              "pdf",
			"filename":          fileName,
			"date_added":        time.Now().Format(time.RFC3339),
			"storage_path":      filePath,
			"file_size":         fileInfo.Size(),
			"page_count":        pageCount,
			"title":             pdfMetadata["Title"],
			"subject":           pdfMetadata["Subject"],
			"keywords":          pdfMetadata["Keywords"],
			"creator":           pdfMetadata["Creator"],
			"producer":          pdfMetadata["Producer"],
			"creation_date":     pdfMetadata["CreationDate"],
			"is_chunked":        len(chunks) > 1, // Boolean flag indicating if document is chunked
			"total_chunks":      len(chunks),     // Total number of chunks
			"original_filename": fileName,        // Original filename (matches RagMeDocs)
		}

		// Remove empty metadata fields
		for k, v := range metadata {
			if v == "" || v == nil {
				delete(metadata, k)
			}
		}

		textData = append(textData, PDFTextData{
			ID:         uuid.New().String(),
			Content:    chunk,
			URL:        fmt.Sprintf("file://%s#chunk-%d", filePath, i),
			Metadata:   metadata,
			PageNumber: i + 1,
			SourcePDF:  filePath,
		})
	}

	return textData, nil
}

// extractTextFromPDF extracts actual text content from a PDF file using pdfcpu
func extractTextFromPDF(filePath string) (string, int, error) {
	// Try pdftotext first (from poppler-utils) as it handles presentation PDFs better
	text, pageCount, err := extractTextWithPdftotext(filePath)
	if err == nil && text != "" && !looksLikeRenderingCommands(text) {
		return text, pageCount, nil
	}

	// Fall back to pdfcpu extraction
	return extractTextWithPdfcpu(filePath)
}

// extractTextWithPdftotext uses the pdftotext command-line tool
func extractTextWithPdftotext(filePath string) (string, int, error) {
	// Check if pdftotext is available in PATH
	_, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", 0, fmt.Errorf("pdftotext not found in PATH")
	}

	// Create temporary output file
	tmpFile, err := os.CreateTemp("", "pdftext-*.txt")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Run pdftotext
	cmd := exec.Command("pdftotext", "-layout", filePath, tmpPath)
	err = cmd.Run()
	if err != nil {
		return "", 0, fmt.Errorf("pdftotext failed: %w", err)
	}

	// Read the extracted text
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read extracted text: %w", err)
	}

	text := strings.TrimSpace(string(content))

	// Count pages (approximate by form feed characters)
	pageCount := strings.Count(text, "\f") + 1
	if pageCount == 1 && text != "" {
		// Estimate based on text length
		pageCount = max(1, len(text)/2000)
	}

	return text, pageCount, nil
}

// extractTextWithPdfcpu extracts text using pdfcpu library
func extractTextWithPdfcpu(filePath string) (string, int, error) {
	// Create a temporary directory for text extraction
	tmpDir, err := os.MkdirTemp("", "pdfextract-*")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract text using pdfcpu
	err = api.ExtractContentFile(filePath, tmpDir, nil, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract content: %w", err)
	}

	// Read the extracted text files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read temp directory: %w", err)
	}

	var textContent strings.Builder
	pageCount := 0

	// Combine text from all pages
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		pageCount++
		textPath := filepath.Join(tmpDir, entry.Name())
		content, err := os.ReadFile(textPath)
		if err != nil {
			continue // Skip files we can't read
		}

		if textContent.Len() > 0 {
			textContent.WriteString("\n\n")
		}
		textContent.Write(content)
	}

	return strings.TrimSpace(textContent.String()), pageCount, nil
}

// looksLikeRenderingCommands checks if the text looks like PDF rendering commands
func looksLikeRenderingCommands(text string) bool {
	if len(text) < 50 {
		return false
	}

	// Check for common PDF operators
	operators := []string{" cm\n", " RG ", " rg\n", "\nq\n", "\nQ\n", " gs\n", " re\n", "\nf\n", "\nW* n\n"}
	matchCount := 0

	for _, op := range operators {
		if strings.Contains(text[:min(500, len(text))], op) {
			matchCount++
		}
	}

	// If we see 3 or more PDF operators in the first 500 chars, it's rendering commands
	return matchCount >= 3
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// extractPDFMetadata extracts metadata from a PDF file using pdfcpu
func extractPDFMetadata(filePath string) (map[string]string, error) {
	metadata := make(map[string]string)

	// Read PDF configuration
	conf := model.NewDefaultConfiguration()

	// Open the PDF file
	file, err := os.Open(filePath)
	if err != nil {
		metadata["filename"] = filepath.Base(filePath)
		metadata["type"] = "pdf"
		return metadata, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Get PDF properties (metadata)
	props, err := api.Properties(file, conf)
	if err != nil {
		// Non-fatal, return basic metadata
		metadata["filename"] = filepath.Base(filePath)
		metadata["type"] = "pdf"
		return metadata, nil
	}

	// Copy properties to our metadata map
	for k, v := range props {
		// Only include non-empty values
		if v != "" {
			metadata[k] = v
		}
	}

	// Ensure we have the basic filename
	metadata["filename"] = filepath.Base(filePath)
	metadata["type"] = "pdf"

	return metadata, nil
}

// generateFallbackPDFContent generates fallback content for PDFs that can't be fully processed
func generateFallbackPDFContent(fileName string, fileSize int64, pageCount int) string {
	if pageCount > 0 {
		return fmt.Sprintf("PDF Document: %s\nFile Size: %d bytes\nPages: %d\n\nThis document could not be fully extracted. It may contain scanned images or use an unsupported encoding.", fileName, fileSize, pageCount)
	}

	// Fallback when page count is unknown
	if fileSize < 1000 {
		return fmt.Sprintf("Small PDF file: %s (%d bytes)", fileName, fileSize)
	} else if fileSize < 100000 {
		return fmt.Sprintf("PDF Document: %s (%d bytes)\n\nThis appears to be a document with moderate content that could not be fully extracted.", fileName, fileSize)
	} else {
		return fmt.Sprintf("Large PDF Document: %s (%d bytes)\n\nThis appears to be a substantial document that could not be fully extracted. It may contain scanned images or use an unsupported encoding.", fileName, fileSize)
	}
}

// chunkText splits text into chunks of specified size with smart boundary detection
func chunkText(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size similar to RagMeDocs
	}

	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string

	// Split by paragraphs first (double newline)
	paragraphs := strings.Split(text, "\n\n")

	currentChunk := ""

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph exceeds chunk size and we have content
		if len(currentChunk)+len(para)+2 > chunkSize && currentChunk != "" {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = para
		} else {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += para
		}

		// If current chunk with this paragraph is too large, split it
		if len(currentChunk) > chunkSize {
			// Try to split by sentences
			sentences := strings.Split(currentChunk, ". ")
			tempChunk := ""

			for i, sent := range sentences {
				sent = strings.TrimSpace(sent)
				if sent == "" {
					continue
				}

				// Add period back if not last sentence
				if i < len(sentences)-1 && !strings.HasSuffix(sent, ".") {
					sent += "."
				}

				if len(tempChunk)+len(sent)+1 > chunkSize && tempChunk != "" {
					chunks = append(chunks, strings.TrimSpace(tempChunk))
					tempChunk = sent
				} else {
					if tempChunk != "" {
						tempChunk += " "
					}
					tempChunk += sent
				}
			}

			currentChunk = tempChunk
		}
	}

	// Add any remaining content
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	// If no chunks were created, force split by character count
	if len(chunks) == 0 {
		for i := 0; i < len(text); i += chunkSize {
			end := i + chunkSize
			if end > len(text) {
				end = len(text)
			}
			chunks = append(chunks, text[i:end])
		}
	}

	return chunks
}

// extractTextByPage extracts text from PDF page by page for image context correlation
func extractTextByPage(filePath string) (map[int]string, error) {
	pageTexts := make(map[int]string)

	// Try pdftotext with page separation first
	text, _, err := extractTextWithPdftotext(filePath)
	if err == nil && text != "" {
		// Split by form feed character (page separator in pdftotext output)
		pages := strings.Split(text, "\f")
		for i, pageText := range pages {
			pageNum := i + 1
			pageTexts[pageNum] = strings.TrimSpace(pageText)
		}
		return pageTexts, nil
	}

	// Fall back to pdfcpu extraction (per-page)
	tmpDir, err := os.MkdirTemp("", "pdfpages-*")
	if err != nil {
		return pageTexts, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract text using pdfcpu
	err = api.ExtractContentFile(filePath, tmpDir, nil, nil)
	if err != nil {
		return pageTexts, fmt.Errorf("failed to extract content: %w", err)
	}

	// Read per-page text files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return pageTexts, fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		// Parse page number from filename (usually "page_N.txt")
		var pageNum int
		if _, err := fmt.Sscanf(entry.Name(), "page_%d.txt", &pageNum); err != nil {
			continue
		}

		textPath := filepath.Join(tmpDir, entry.Name())
		content, err := os.ReadFile(textPath)
		if err != nil {
			continue
		}

		pageTexts[pageNum] = strings.TrimSpace(string(content))
	}

	return pageTexts, nil
}

// enrichImageWithContext adds caption and surrounding text to image metadata
func enrichImageWithContext(image *PDFImageData, pageTexts map[int]string, maxMetadataLength int) {
	if image == nil {
		return
	}

	// Use default 2000 if maxMetadataLength is 0 (unlimited not recommended for metadata)
	if maxMetadataLength == 0 {
		maxMetadataLength = 2000
	}

	// Get text from the page where this image appears
	pageText, hasPageText := pageTexts[image.PageNumber]
	if !hasPageText || pageText == "" {
		// No page text available, use OCR text if available
		if image.OCRText != "" {
			image.Caption = extractCaptionFromOCR(image.OCRText)
			image.SurroundingText = truncateText(image.OCRText, maxMetadataLength)
		}
		return
	}

	// Extract caption (look for "Figure", "Fig.", "Image", "Diagram", etc.)
	image.Caption = extractCaptionFromText(pageText, image.ImageIndex)

	// Extract surrounding text using configurable max length
	image.SurroundingText = truncateText(pageText, maxMetadataLength)

	// Extract section heading (first line or heading-like text) - limit to 200 chars
	image.SectionHeading = truncateText(extractSectionHeading(pageText), 200)
}

// extractCaptionFromText tries to find a caption related to the image
func extractCaptionFromText(pageText string, imageIndex int) string {
	// Look for common caption patterns
	captionPatterns := []string{
		fmt.Sprintf("Figure %d:", imageIndex+1),
		fmt.Sprintf("Fig. %d:", imageIndex+1),
		fmt.Sprintf("Image %d:", imageIndex+1),
		"Figure:",
		"Fig.:",
		"Diagram:",
		"Chart:",
		"Graph:",
	}

	lines := strings.Split(pageText, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		for _, pattern := range captionPatterns {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				// Found a caption, get this line and maybe the next one
				caption := line
				if i+1 < len(lines) && len(lines[i+1]) < 200 {
					caption += " " + strings.TrimSpace(lines[i+1])
				}
				return truncateText(caption, 300)
			}
		}
	}

	// No caption found, return empty
	return ""
}

// extractCaptionFromOCR extracts caption from OCR text
func extractCaptionFromOCR(ocrText string) string {
	// For OCR text, just take the first meaningful line
	lines := strings.Split(ocrText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 { // Skip very short lines
			return truncateText(line, 200)
		}
	}
	return ""
}

// extractSectionHeading tries to extract the section/chapter heading from page text
func extractSectionHeading(pageText string) string {
	lines := strings.Split(pageText, "\n")

	// Look for heading-like patterns:
	// - Short lines (< 80 chars)
	// - At the beginning of the page
	// - ALL CAPS or Title Case
	// - May have numbers (e.g., "1. Introduction")

	for i, line := range lines {
		if i > 5 {
			break // Only check first few lines
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || len(line) > 100 {
			continue
		}

		// Check if it looks like a heading
		isAllCaps := strings.ToUpper(line) == line && len(line) > 5
		hasNumbers := strings.Contains(line, "1.") || strings.Contains(line, "2.") || strings.Contains(line, "3.")

		if isAllCaps || hasNumbers {
			return truncateText(line, 150)
		}
	}

	// No heading found, return first substantial line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 20 && len(line) < 100 {
			return truncateText(line, 150)
		}
	}

	return ""
}

// truncateText truncates text to maxLen characters with ellipsis if needed
func truncateText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
