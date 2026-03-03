// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/version"
)

const (
	// BackupVersion is the current backup format version
	BackupVersion = "1.0.0"

	// BackupExtension is the file extension for backup files
	BackupExtension = ".weavebak"

	// CompressedExtension is the additional extension for compressed backups
	CompressedExtension = ".gz"
)

// WriteBackup writes a backup to a file with optional compression
func WriteBackup(backup *BackupFormat, outputPath string, compress bool) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Add proper extension if not present
	if !strings.HasSuffix(outputPath, BackupExtension) && !strings.HasSuffix(outputPath, BackupExtension+CompressedExtension) {
		outputPath = outputPath + BackupExtension
		if compress {
			outputPath = outputPath + CompressedExtension
		}
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file

	// Add gzip compression if requested
	if compress {
		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()
		writer = gzipWriter
	}

	// Encode backup as JSON
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Pretty print for readability
	if err := encoder.Encode(backup); err != nil {
		return fmt.Errorf("failed to encode backup: %w", err)
	}

	return nil
}

// ReadBackup reads a backup from a file (auto-detects compression)
func ReadBackup(inputPath string) (*BackupFormat, error) {
	// Open backup file
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// Auto-detect gzip compression
	if strings.HasSuffix(inputPath, CompressedExtension) {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Decode backup from JSON
	var backup BackupFormat
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&backup); err != nil {
		return nil, fmt.Errorf("failed to decode backup: %w", err)
	}

	return &backup, nil
}

// ValidateBackup validates a backup file
func ValidateBackup(inputPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check if file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, "backup file does not exist")
		return result, nil
	}

	// Try to read the backup
	backup, err := ReadBackup(inputPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read backup: %v", err))
		return result, nil
	}

	// Validate version
	if backup.Version == "" {
		result.Warnings = append(result.Warnings, "backup version not specified")
	} else if backup.Version != BackupVersion {
		result.Warnings = append(result.Warnings, fmt.Sprintf("backup version mismatch (expected %s, got %s)", BackupVersion, backup.Version))
	}
	result.Version = backup.Version

	// Validate metadata
	if backup.Metadata.Collection == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "collection name is missing")
	}
	result.Collection = backup.Metadata.Collection

	if backup.Metadata.VectorDimensions <= 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "vector dimensions must be positive")
	}

	// Validate documents
	result.TotalDocuments = len(backup.Documents)
	if result.TotalDocuments == 0 {
		result.Warnings = append(result.Warnings, "backup contains no documents")
	}

	// Validate each document has required fields
	for i, doc := range backup.Documents {
		if doc.ID == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("document %d missing ID", i))
		}
		if len(doc.Embedding) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("document %s missing embedding", doc.ID))
		}
		if len(doc.Embedding) != backup.Metadata.VectorDimensions {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("document %s embedding dimension mismatch (expected %d, got %d)",
				doc.ID, backup.Metadata.VectorDimensions, len(doc.Embedding)))
		}
	}

	// Get file size
	fileInfo, err := os.Stat(inputPath)
	if err == nil {
		result.BackupSizeBytes = fileInfo.Size()
	}

	return result, nil
}

// NewBackupFormat creates a new backup format with current version and metadata
func NewBackupFormat(collection string, vdbType string, embeddingModel string, vectorDims int) *BackupFormat {
	versionInfo := version.Get()
	return &BackupFormat{
		Version: BackupVersion,
		Metadata: BackupMetadata{
			Collection:       collection,
			VDBType:          vdbType,
			EmbeddingModel:   embeddingModel,
			VectorDimensions: vectorDims,
			WeaveVersion:     versionInfo.Version,
			CreatedAt:        time.Now(),
		},
		Documents: []BackupDocument{},
	}
}
