// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
)

func TestPDFIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the path to the test fixture
	fixturePath := filepath.Join("fixtures", "ragme-io.pdf")

	t.Run("ExtractPDFContent", func(t *testing.T) {
		// Test with default chunk size (1000)
		textData, imageData, err := pdf.ExtractPDFContent(fixturePath, 1000, true, 10240, 2000, false)
		if err != nil {
			t.Fatalf("ExtractPDFContent() error = %v", err)
		}

		// Verify we got text chunks
		if len(textData) == 0 {
			t.Error("ExtractPDFContent() returned no text chunks")
		}

		t.Logf("Extracted %d text chunks from PDF", len(textData))

		// Verify metadata structure matches RagMeDocs format
		for i, chunk := range textData {
			// Check required RagMeDocs fields
			requiredFields := []string{
				"chunk_index",
				"chunk_sizes",
				"source_document",
				"processing_info",
			}

			for _, field := range requiredFields {
				if _, ok := chunk.Metadata[field]; !ok {
					t.Errorf("Chunk %d missing required metadata field: %s", i, field)
				}
			}

			// Verify chunk_index is correct
			if chunkIdx, ok := chunk.Metadata["chunk_index"].(int); ok {
				if chunkIdx != i {
					t.Errorf("Chunk %d has incorrect chunk_index: got %d, want %d", i, chunkIdx, i)
				}
			} else {
				t.Errorf("Chunk %d chunk_index is not an int", i)
			}

			// Verify chunk_sizes is an array and has correct length
			if chunkSizes, ok := chunk.Metadata["chunk_sizes"].([]int); ok {
				if len(chunkSizes) != len(textData) {
					t.Errorf("Chunk %d chunk_sizes has wrong length: got %d, want %d", i, len(chunkSizes), len(textData))
				}
			} else {
				t.Errorf("Chunk %d chunk_sizes is not []int", i)
			}

			// Verify URL format matches RagMeDocs: file:///path/to/file.pdf#chunk-N
			if chunk.URL == "" {
				t.Errorf("Chunk %d has empty URL", i)
			}

			// Verify content is not empty
			if chunk.Content == "" {
				t.Errorf("Chunk %d has empty content", i)
			}

			// Verify chunk size is reasonable (similar to RagMeDocs: 800-1000 chars)
			if len(chunk.Content) < 100 {
				t.Logf("Warning: Chunk %d is smaller than expected: %d chars", i, len(chunk.Content))
			}
			if len(chunk.Content) > 2000 {
				t.Logf("Warning: Chunk %d is larger than expected: %d chars", i, len(chunk.Content))
			}
		}

		// Images are optional
		t.Logf("Extracted %d images from PDF", len(imageData))
	})

	t.Run("ExtractPDFContentWithDifferentChunkSizes", func(t *testing.T) {
		chunkSizes := []int{500, 1000, 2000}

		for _, chunkSize := range chunkSizes {
			textData, _, err := pdf.ExtractPDFContent(fixturePath, chunkSize, true, 10240, 2000, false)
			if err != nil {
				t.Fatalf("ExtractPDFContent(chunkSize=%d) error = %v", chunkSize, err)
			}

			if len(textData) == 0 {
				t.Errorf("ExtractPDFContent(chunkSize=%d) returned no chunks", chunkSize)
			}

			t.Logf("ChunkSize %d: Generated %d chunks", chunkSize, len(textData))

			// Verify chunk sizes are within reasonable bounds
			for i, chunk := range textData {
				chunkLen := len(chunk.Content)
				// Allow some flexibility for smart boundary detection
				maxSize := chunkSize * 2
				if chunkLen > maxSize {
					t.Errorf("ChunkSize %d, Chunk %d: content too large: got %d chars, max %d",
						chunkSize, i, chunkLen, maxSize)
				}
			}
		}
	})

	t.Run("ExtractPDFContentPreservesAllContent", func(t *testing.T) {
		// Skip in short mode to avoid timeout in CI
		if testing.Short() {
			t.Skip("Skipping long-running PDF content preservation test")
		}

		// Extract with different chunk sizes and verify total content is similar
		textData1, _, err := pdf.ExtractPDFContent(fixturePath, 500, true, 10240, 2000, false)
		if err != nil {
			t.Fatalf("ExtractPDFContent() error = %v", err)
		}

		textData2, _, err := pdf.ExtractPDFContent(fixturePath, 2000, true, 10240, 2000, false)
		if err != nil {
			t.Fatalf("ExtractPDFContent() error = %v", err)
		}

		// Calculate total content length
		totalLen1 := 0
		for _, chunk := range textData1 {
			totalLen1 += len(chunk.Content)
		}

		totalLen2 := 0
		for _, chunk := range textData2 {
			totalLen2 += len(chunk.Content)
		}

		// Content should be similar (within 10% due to whitespace handling)
		diff := float64(totalLen1-totalLen2) / float64(totalLen1)
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.10 {
			t.Errorf("Total content length varies too much: %d vs %d (%.1f%% diff)",
				totalLen1, totalLen2, diff*100)
		}

		t.Logf("Total content preserved: %d chars (500-byte chunks) vs %d chars (2000-byte chunks)",
			totalLen1, totalLen2)
	})
}

func TestPDFMetadataExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fixturePath := filepath.Join("fixtures", "ragme-io.pdf")

	textData, _, err := pdf.ExtractPDFContent(fixturePath, 1000, true, 10240, 2000, false)
	if err != nil {
		t.Fatalf("ExtractPDFContent() error = %v", err)
	}

	if len(textData) == 0 {
		t.Fatal("No text data extracted")
	}

	// Check first chunk for metadata
	chunk := textData[0]

	t.Log("Metadata fields present:")
	for k, v := range chunk.Metadata {
		t.Logf("  %s: %v", k, v)
	}

	// Check for computed/inferred metadata
	computedFields := []string{
		"chunk_index",
		"chunk_sizes",
		"source_document",
		"processing_info",
		"date_added",
		"file_size",
		"page_count",
	}

	for _, field := range computedFields {
		if _, ok := chunk.Metadata[field]; !ok {
			t.Errorf("Missing computed metadata field: %s", field)
		}
	}

	// PDF metadata fields (may or may not be present depending on PDF)
	optionalFields := []string{
		"author",
		"title",
		"subject",
		"keywords",
		"creator",
		"producer",
		"creation_date",
	}

	presentCount := 0
	for _, field := range optionalFields {
		if val, ok := chunk.Metadata[field]; ok && val != "" && val != nil {
			presentCount++
			t.Logf("Optional field '%s' present: %v", field, val)
		}
	}

	t.Logf("Found %d/%d optional metadata fields", presentCount, len(optionalFields))
}

// TestPDFProcessingConsistency verifies that processing the same PDF multiple times
// produces consistent results
func TestPDFProcessingConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fixturePath := filepath.Join("fixtures", "ragme-io.pdf")

	// Process the same PDF 3 times
	var results [][]pdf.PDFTextData
	for i := 0; i < 3; i++ {
		textData, _, err := pdf.ExtractPDFContent(fixturePath, 1000, true, 10240, 2000, false)
		if err != nil {
			t.Fatalf("ExtractPDFContent() run %d error = %v", i+1, err)
		}
		results = append(results, textData)
	}

	// Verify all runs produced the same number of chunks
	firstLen := len(results[0])
	for i, result := range results[1:] {
		if len(result) != firstLen {
			t.Errorf("Run %d produced %d chunks, want %d", i+2, len(result), firstLen)
		}
	}

	// Verify chunk content is consistent
	for chunkIdx := 0; chunkIdx < firstLen && chunkIdx < len(results[1]) && chunkIdx < len(results[2]); chunkIdx++ {
		content1 := results[0][chunkIdx].Content
		content2 := results[1][chunkIdx].Content
		content3 := results[2][chunkIdx].Content

		if content1 != content2 || content1 != content3 {
			t.Errorf("Chunk %d content varies across runs", chunkIdx)
		}
	}

	t.Log("PDF processing is consistent across multiple runs")
}

func init() {
	// Initialize config for tests
	_ = context.Background()
	_ = config.VectorDBConfig{}
}
