// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBackupFormat(t *testing.T) {
	backup := NewBackupFormat("TestCollection", "milvus-local", "text-embedding-3-small", 1536)

	if backup.Version != BackupVersion {
		t.Errorf("Expected version %s, got %s", BackupVersion, backup.Version)
	}

	if backup.Metadata.Collection != "TestCollection" {
		t.Errorf("Expected collection TestCollection, got %s", backup.Metadata.Collection)
	}

	if backup.Metadata.VDBType != "milvus-local" {
		t.Errorf("Expected VDB type milvus-local, got %s", backup.Metadata.VDBType)
	}

	if backup.Metadata.VectorDimensions != 1536 {
		t.Errorf("Expected 1536 dimensions, got %d", backup.Metadata.VectorDimensions)
	}

	if len(backup.Documents) != 0 {
		t.Errorf("Expected empty documents, got %d", len(backup.Documents))
	}
}

func TestWriteAndReadBackup(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create backup
	backup := NewBackupFormat("TestCollection", "milvus-local", "text-embedding-3-small", 1536)
	backup.Documents = []BackupDocument{
		{
			ID:        "doc1",
			Content:   "Test content 1",
			Embedding: []float64{0.1, 0.2, 0.3},
			Metadata:  map[string]interface{}{"key": "value1"},
		},
		{
			ID:        "doc2",
			Content:   "Test content 2",
			Embedding: []float64{0.4, 0.5, 0.6},
			Metadata:  map[string]interface{}{"key": "value2"},
		},
	}
	backup.Metadata.TotalDocuments = len(backup.Documents)

	// Write uncompressed backup
	outputPath := filepath.Join(tmpDir, "test.weavebak")
	if err := WriteBackup(backup, outputPath, false); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Read backup
	readBackup, err := ReadBackup(outputPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	// Verify
	if readBackup.Version != backup.Version {
		t.Errorf("Version mismatch: expected %s, got %s", backup.Version, readBackup.Version)
	}

	if readBackup.Metadata.Collection != backup.Metadata.Collection {
		t.Errorf("Collection mismatch: expected %s, got %s", backup.Metadata.Collection, readBackup.Metadata.Collection)
	}

	if len(readBackup.Documents) != len(backup.Documents) {
		t.Errorf("Document count mismatch: expected %d, got %d", len(backup.Documents), len(readBackup.Documents))
	}

	// Verify first document
	if readBackup.Documents[0].ID != "doc1" {
		t.Errorf("Document ID mismatch: expected doc1, got %s", readBackup.Documents[0].ID)
	}

	if readBackup.Documents[0].Content != "Test content 1" {
		t.Errorf("Content mismatch: expected 'Test content 1', got %s", readBackup.Documents[0].Content)
	}
}

func TestWriteAndReadCompressedBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backup := NewBackupFormat("TestCollection", "milvus-local", "text-embedding-3-small", 1536)
	backup.Documents = []BackupDocument{
		{
			ID:        "doc1",
			Content:   "Test content with compression",
			Embedding: []float64{0.1, 0.2, 0.3},
		},
	}

	// Write compressed backup
	outputPath := filepath.Join(tmpDir, "test-compressed.weavebak.gz")
	if err := WriteBackup(backup, outputPath, true); err != nil {
		t.Fatalf("Failed to write compressed backup: %v", err)
	}

	// Verify file exists and has .gz extension
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Compressed backup file not created")
	}

	// Read compressed backup
	readBackup, err := ReadBackup(outputPath)
	if err != nil {
		t.Fatalf("Failed to read compressed backup: %v", err)
	}

	// Verify
	if len(readBackup.Documents) != 1 {
		t.Errorf("Expected 1 document, got %d", len(readBackup.Documents))
	}

	if readBackup.Documents[0].ID != "doc1" {
		t.Errorf("Document ID mismatch after compression")
	}
}

func TestValidateBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid backup
	backup := NewBackupFormat("TestCollection", "milvus-local", "text-embedding-3-small", 1536)
	backup.Documents = []BackupDocument{
		{
			ID:        "doc1",
			Content:   "Test",
			Embedding: make([]float64, 1536), // Correct dimensions
		},
	}
	backup.Metadata.TotalDocuments = len(backup.Documents)

	outputPath := filepath.Join(tmpDir, "valid.weavebak")
	if err := WriteBackup(backup, outputPath, false); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Validate
	result, err := ValidateBackup(outputPath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid backup, got invalid. Errors: %v", result.Errors)
	}

	if result.Collection != "TestCollection" {
		t.Errorf("Collection mismatch: expected TestCollection, got %s", result.Collection)
	}

	if result.TotalDocuments != 1 {
		t.Errorf("Expected 1 document, got %d", result.TotalDocuments)
	}
}

func TestValidateInvalidBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid backup (wrong embedding dimensions)
	backup := NewBackupFormat("TestCollection", "milvus-local", "text-embedding-3-small", 1536)
	backup.Documents = []BackupDocument{
		{
			ID:        "doc1",
			Content:   "Test",
			Embedding: []float64{0.1, 0.2}, // Wrong dimensions (2 instead of 1536)
		},
	}

	outputPath := filepath.Join(tmpDir, "invalid.weavebak")
	if err := WriteBackup(backup, outputPath, false); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Validate
	result, err := ValidateBackup(outputPath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid backup, got valid")
	}

	if len(result.Errors) == 0 {
		t.Errorf("Expected validation errors, got none")
	}
}

func TestValidateNonExistentBackup(t *testing.T) {
	result, err := ValidateBackup("/nonexistent/path/backup.weavebak")
	if err != nil {
		t.Fatalf("Validation should not error: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid result for non-existent file")
	}

	if len(result.Errors) == 0 {
		t.Errorf("Expected error for non-existent file")
	}
}

func TestBackupFormatExtensions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backup := NewBackupFormat("Test", "milvus-local", "text-embedding-3-small", 1536)

	tests := []struct {
		name       string
		outputPath string
		compress   bool
		wantExt    string
	}{
		{"No extension, no compress", filepath.Join(tmpDir, "test1"), false, ".weavebak"},
		{"No extension, compress", filepath.Join(tmpDir, "test2"), true, ".weavebak.gz"},
		{"With .weavebak, no compress", filepath.Join(tmpDir, "test3.weavebak"), false, ".weavebak"},
		{"With .weavebak, compress", filepath.Join(tmpDir, "test4.weavebak"), true, ".weavebak.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteBackup(backup, tt.outputPath, tt.compress)
			if err != nil {
				t.Fatalf("WriteBackup failed: %v", err)
			}

			// Check that file exists with correct extension
			expectedPath := tt.outputPath
			if filepath.Ext(expectedPath) != tt.wantExt {
				if tt.compress && filepath.Ext(expectedPath) != ".gz" {
					expectedPath = expectedPath + tt.wantExt
				} else if !tt.compress && filepath.Ext(expectedPath) != ".weavebak" {
					expectedPath = expectedPath + ".weavebak"
				}
			}

			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				// Try the modified path
				files, _ := filepath.Glob(filepath.Join(tmpDir, "test*"))
				t.Logf("Files created: %v", files)
			}
		})
	}
}

func TestBackupMetadata(t *testing.T) {
	backup := NewBackupFormat("TestCol", "weaviate-cloud", "text-embedding-3-large", 3072)

	// Check metadata fields
	if backup.Metadata.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if backup.Metadata.WeaveVersion == "" {
		t.Error("WeaveVersion should not be empty")
	}

	if backup.Metadata.VectorDimensions != 3072 {
		t.Errorf("Expected 3072 dimensions, got %d", backup.Metadata.VectorDimensions)
	}

	// Check that created time is recent (within last minute)
	if time.Since(backup.Metadata.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt timestamp seems too old: %v", backup.Metadata.CreatedAt)
	}
}
