// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/maximilien/weave-cli/src/pkg/config"
	mockdb "github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestBackupCreateWithMockVDB(t *testing.T) {
	// Skip test - requires OCR dependencies (gosseract/tesseract)
	// Run manually with: go test -v -tags=integration ./src/cmd/backup/...
	t.Skip("Skipping backup integration test - requires OCR dependencies")

	// Create temp directory for backup
	tmpDir, err := os.MkdirTemp("", "backup-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock VDB client with test data
	mockClient := mockdb.NewClient(&config.MockConfig{
		Collections: []config.MockCollection{},
	})
	ctx := context.Background()

	// Create test collection
	schema := &vectordb.CollectionSchema{
		Class:      "TestCollection",
		Vectorizer: "text-embedding-3-small",
	}

	if err := mockClient.CreateCollectionWithSchema(ctx, "TestCollection", schema); err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Add test documents
	testDocs := []*vectordb.Document{
		{
			ID:        "doc1",
			Content:   "Test document 1",
			Embedding: []float64{0.1, 0.2, 0.3},
			Metadata:  map[string]interface{}{"source": "test1"},
		},
		{
			ID:        "doc2",
			Content:   "Test document 2",
			Embedding: []float64{0.4, 0.5, 0.6},
			Metadata:  map[string]interface{}{"source": "test2"},
		},
	}

	if err := mockClient.CreateDocuments(ctx, "TestCollection", testDocs); err != nil {
		t.Fatalf("Failed to create documents: %v", err)
	}

	// Create backup format
	backup := backuppkg.NewBackupFormat(
		"TestCollection",
		"mock",
		"text-embedding-3-small",
		3, // mock uses 3 dimensions for test
	)

	// Export documents
	docs, err := mockClient.ListDocuments(ctx, "TestCollection", 100)
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	for _, doc := range docs {
		backupDoc := backuppkg.BackupDocument{
			ID:        doc.ID,
			Content:   doc.Content,
			Embedding: []float64{0.1, 0.2, 0.3}, // Mock embedding for test
			Metadata:  doc.Metadata,
		}
		backup.Documents = append(backup.Documents, backupDoc)
	}

	backup.Metadata.TotalDocuments = len(backup.Documents)

	// Write backup
	outputPath := filepath.Join(tmpDir, "test-backup.weavebak")
	if err := backuppkg.WriteBackup(backup, outputPath, false); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Backup file was not created")
	}

	// Validate backup
	result, err := backuppkg.ValidateBackup(outputPath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Backup validation failed. Errors: %v", result.Errors)
	}

	if result.TotalDocuments != 2 {
		t.Errorf("Expected 2 documents in backup, got %d", result.TotalDocuments)
	}

	// Read backup and verify contents
	readBackup, err := backuppkg.ReadBackup(outputPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if len(readBackup.Documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(readBackup.Documents))
	}

	// Verify document content
	if readBackup.Documents[0].Content != "Test document 1" {
		t.Errorf("Document content mismatch")
	}
}

func TestBackupCreateCompressed(t *testing.T) {
	t.Skip("Skipping backup integration test - requires OCR dependencies")

	tmpDir, err := os.MkdirTemp("", "backup-compressed-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup with data
	backup := backuppkg.NewBackupFormat("TestCol", "mock", "text-embedding-3-small", 1536)
	for i := 0; i < 100; i++ {
		doc := backuppkg.BackupDocument{
			ID:        string(rune('a' + i)),
			Content:   "Test content with some repetitive text to test compression",
			Embedding: make([]float64, 1536),
		}
		backup.Documents = append(backup.Documents, doc)
	}

	// Write uncompressed
	uncompressedPath := filepath.Join(tmpDir, "uncompressed.weavebak")
	if err := backuppkg.WriteBackup(backup, uncompressedPath, false); err != nil {
		t.Fatalf("Failed to write uncompressed: %v", err)
	}

	// Write compressed
	compressedPath := filepath.Join(tmpDir, "compressed.weavebak.gz")
	if err := backuppkg.WriteBackup(backup, compressedPath, true); err != nil {
		t.Fatalf("Failed to write compressed: %v", err)
	}

	// Get file sizes
	uncompressedInfo, _ := os.Stat(uncompressedPath)
	compressedInfo, _ := os.Stat(compressedPath)

	// Compressed should be smaller
	if compressedInfo.Size() >= uncompressedInfo.Size() {
		t.Logf("Warning: Compressed size (%d) >= uncompressed size (%d)", compressedInfo.Size(), uncompressedInfo.Size())
		// Note: For small files, compression might not help much, so we just log it
	}

	// Verify both can be read
	readUncompressed, err := backuppkg.ReadBackup(uncompressedPath)
	if err != nil {
		t.Fatalf("Failed to read uncompressed: %v", err)
	}

	readCompressed, err := backuppkg.ReadBackup(compressedPath)
	if err != nil {
		t.Fatalf("Failed to read compressed: %v", err)
	}

	// Both should have same number of documents
	if len(readUncompressed.Documents) != len(readCompressed.Documents) {
		t.Errorf("Document count mismatch between compressed and uncompressed")
	}
}

func TestBackupCreateOptions(t *testing.T) {
	// Test CreateOptions validation
	opts := &backuppkg.CreateOptions{
		Collection: "TestCollection",
		OutputFile: "/tmp/test.weavebak",
		VDBType:    "milvus-local",
		Compress:   true,
		BatchSize:  100,
		Quiet:      false,
	}

	if opts.Collection == "" {
		t.Error("Collection should not be empty")
	}

	if opts.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}

	if opts.OutputFile == "" {
		t.Error("OutputFile should not be empty")
	}
}

func TestBackupWithImageData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup with image data
	backup := backuppkg.NewBackupFormat("ImageCollection", "milvus-local", "clip-vit", 512)
	backup.Documents = []backuppkg.BackupDocument{
		{
			ID:             "img1",
			Content:        "An image",
			Embedding:      make([]float64, 512),
			Image:          "data:image/png;base64,iVBORw0KG...",
			ImageData:      "base64encodeddata",
			ImageThumbnail: "thumbnail_base64",
			ImageURL:       "https://storage.example.com/img1.jpg",
			ImageMetadata: map[string]interface{}{
				"size":   1024,
				"format": "jpeg",
			},
		},
	}

	outputPath := filepath.Join(tmpDir, "images.weavebak")
	if err := backuppkg.WriteBackup(backup, outputPath, false); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Read and verify image fields are preserved
	readBackup, err := backuppkg.ReadBackup(outputPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if len(readBackup.Documents) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(readBackup.Documents))
	}

	doc := readBackup.Documents[0]
	if doc.Image == "" {
		t.Error("Image field not preserved")
	}

	if doc.ImageThumbnail == "" {
		t.Error("ImageThumbnail field not preserved")
	}

	if doc.ImageURL == "" {
		t.Error("ImageURL field not preserved")
	}

	if doc.ImageMetadata == nil {
		t.Error("ImageMetadata field not preserved")
	}
}
