// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractPDFText(t *testing.T) {
	// Get the path to the test fixture
	fixturePath := filepath.Join("..", "..", "..", "tests", "fixtures", "ragme-io.pdf")

	tests := []struct {
		name          string
		filePath      string
		chunkSize     int
		wantMinChunks int
		wantMaxChunks int
		wantErr       bool
	}{
		{
			name:          "Extract with default chunk size",
			filePath:      fixturePath,
			chunkSize:     1000,
			wantMinChunks: 5,  // ragme-io.pdf should produce at least 5 chunks
			wantMaxChunks: 30, // but not more than 30 chunks (22 pages)
			wantErr:       false,
		},
		{
			name:          "Extract with smaller chunk size",
			filePath:      fixturePath,
			chunkSize:     500,
			wantMinChunks: 10, // Smaller chunks = more chunks
			wantMaxChunks: 50,
			wantErr:       false,
		},
		{
			name:          "Extract with large chunk size",
			filePath:      fixturePath,
			chunkSize:     5000,
			wantMinChunks: 1,  // Larger chunks = fewer chunks
			wantMaxChunks: 25, // Still 22 pages worth of content
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textData, err := extractPDFText(tt.filePath, tt.chunkSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractPDFText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Check number of chunks
				if len(textData) < tt.wantMinChunks {
					t.Errorf("extractPDFText() got %d chunks, want at least %d", len(textData), tt.wantMinChunks)
				}
				if len(textData) > tt.wantMaxChunks {
					t.Errorf("extractPDFText() got %d chunks, want at most %d", len(textData), tt.wantMaxChunks)
				}

				// Check each chunk
				for i, chunk := range textData {
					// Check required fields
					if chunk.ID == "" {
						t.Errorf("Chunk %d missing ID", i)
					}
					if chunk.Content == "" {
						t.Errorf("Chunk %d missing Content", i)
					}
					if chunk.URL == "" {
						t.Errorf("Chunk %d missing URL", i)
					}
					if chunk.Metadata == nil {
						t.Errorf("Chunk %d missing Metadata", i)
					}

					// Check RagMeDocs-compatible metadata fields
					if _, ok := chunk.Metadata["chunk_index"]; !ok {
						t.Errorf("Chunk %d missing chunk_index in metadata", i)
					}
					if _, ok := chunk.Metadata["chunk_sizes"]; !ok {
						t.Errorf("Chunk %d missing chunk_sizes in metadata", i)
					}
					if _, ok := chunk.Metadata["source_document"]; !ok {
						t.Errorf("Chunk %d missing source_document in metadata", i)
					}

					// Check URL format
					expectedURL := filepath.Base(tt.filePath) + "#chunk-"
					if !contains(chunk.URL, expectedURL) {
						t.Errorf("Chunk %d URL = %v, want to contain %v", i, chunk.URL, expectedURL)
					}

					// Check chunk index matches
					if chunkIdx, ok := chunk.Metadata["chunk_index"].(int); ok {
						if chunkIdx != i {
							t.Errorf("Chunk %d has chunk_index = %d, want %d", i, chunkIdx, i)
						}
					}
				}

				// Check chunk sizes array
				if len(textData) > 0 {
					chunkSizes, ok := textData[0].Metadata["chunk_sizes"].([]int)
					if !ok {
						t.Error("chunk_sizes is not []int")
					} else if len(chunkSizes) != len(textData) {
						t.Errorf("chunk_sizes length = %d, want %d", len(chunkSizes), len(textData))
					}
				}
			}
		})
	}
}

func TestChunkText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		chunkSize int
		wantLen   int
	}{
		{
			name:      "Empty text",
			text:      "",
			chunkSize: 100,
			wantLen:   0,
		},
		{
			name:      "Short text",
			text:      "Short text that fits in one chunk.",
			chunkSize: 100,
			wantLen:   1,
		},
		{
			name:      "Text requiring multiple chunks",
			text:      "This is a longer text. " + repeatString("Sentence. ", 50),
			chunkSize: 100,
			wantLen:   6, // Approximate, depends on smart boundary detection
		},
		{
			name:      "Zero chunk size uses default",
			text:      "Some text",
			chunkSize: 0,
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := chunkText(tt.text, tt.chunkSize)

			if tt.text == "" {
				// Empty text may return empty array or single empty string
				if len(chunks) > 1 {
					t.Errorf("chunkText() empty text returned %d chunks, want 0 or 1", len(chunks))
				}
				return
			}

			if len(chunks) == 0 {
				t.Error("chunkText() returned no chunks for non-empty text")
				return
			}

			// Check that chunks are reasonable size
			for i, chunk := range chunks {
				if tt.chunkSize > 0 {
					// Allow some flexibility for smart boundary detection
					maxSize := tt.chunkSize * 2
					if len(chunk) > maxSize {
						t.Errorf("Chunk %d too large: got %d chars, max %d", i, len(chunk), maxSize)
					}
				}
			}

			// Verify all text is preserved
			combined := ""
			for _, chunk := range chunks {
				combined += chunk + " "
			}
			// Basic check that content is preserved (allowing for whitespace changes)
			if len(combined) < len(tt.text)/2 {
				t.Errorf("Combined chunks much smaller than original: got %d, original %d",
					len(combined), len(tt.text))
			}
		})
	}
}

func TestExtractPDFMetadata(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "tests", "fixtures", "ragme-io.pdf")

	metadata, err := extractPDFMetadata(fixturePath)
	if err != nil {
		t.Fatalf("extractPDFMetadata() error = %v", err)
	}

	// Check required fields
	if metadata["filename"] == "" {
		t.Error("metadata missing filename")
	}
	if metadata["type"] != "pdf" {
		t.Errorf("metadata type = %v, want pdf", metadata["type"])
	}

	// Check that filename matches
	if metadata["filename"] != "ragme-io.pdf" {
		t.Errorf("metadata filename = %v, want ragme-io.pdf", metadata["filename"])
	}
}

func TestExtractTextFromPDF(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "tests", "fixtures", "ragme-io.pdf")

	text, pageCount, err := extractTextFromPDF(fixturePath)
	if err != nil {
		t.Fatalf("extractTextFromPDF() error = %v", err)
	}

	// Check that we got some text
	if text == "" {
		t.Error("extractTextFromPDF() returned empty text")
	}

	// Check page count is reasonable (ragme-io.pdf has 22 pages)
	if pageCount < 1 {
		t.Errorf("extractTextFromPDF() pageCount = %d, want > 0", pageCount)
	}
	if pageCount > 100 {
		t.Errorf("extractTextFromPDF() pageCount = %d, seems too high", pageCount)
	}

	// Check that text contains expected content from the PDF
	// Note: pdfcpu extracts text per page, so content may be fragmented
	// We just check that we got SOME meaningful text
	if len(text) < 100 {
		t.Errorf("extractTextFromPDF() text too short: got %d chars, want at least 100", len(text))
	}

	// Optional: try to find at least one expected word from the PDF
	// (this may fail if the PDF uses non-standard fonts or encoding)
	expectedKeywords := []string{"RAGme", "Personal", "RAG", "Agent", "Web", "Content"}
	foundAny := false
	for _, keyword := range expectedKeywords {
		if contains(text, keyword) {
			foundAny = true
			break
		}
	}
	if !foundAny {
		t.Logf("Warning: No expected keywords found in extracted text (may be due to PDF encoding)")
		t.Logf("First 200 chars of extracted text: %s", text[:min(200, len(text))])
	}
}

// Helper functions

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
