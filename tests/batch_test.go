// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/cmd/document"
)

// TestBatchProcessedFileStatus tests the ProcessedFileStatus structure
func TestBatchProcessedFileStatus(t *testing.T) {
	t.Run("MarshalUnmarshal", func(t *testing.T) {
		status := document.ProcessedFileStatus{
			FilePath:         "/path/to/file.txt",
			ProcessedAt:      time.Now(),
			Success:          true,
			TextChunks:       5,
			Images:           0,
			ProcessingTimeMs: 1234,
			FileSize:         5000,
			RetryCount:       0,
		}

		// Marshal to JSON
		data, err := json.Marshal(&status)
		if err != nil {
			t.Fatalf("Failed to marshal status: %v", err)
		}

		// Unmarshal back
		var decoded document.ProcessedFileStatus
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal status: %v", err)
		}

		// Verify fields
		if decoded.FilePath != status.FilePath {
			t.Errorf("FilePath mismatch: got %s, want %s", decoded.FilePath, status.FilePath)
		}
		if decoded.Success != status.Success {
			t.Errorf("Success mismatch: got %v, want %v", decoded.Success, status.Success)
		}
		if decoded.TextChunks != status.TextChunks {
			t.Errorf("TextChunks mismatch: got %d, want %d", decoded.TextChunks, status.TextChunks)
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		statusFile := filepath.Join(tmpDir, "test.txt.processed")

		// Create status
		original := document.ProcessedFileStatus{
			FilePath:         filepath.Join(tmpDir, "test.txt"),
			ProcessedAt:      time.Now(),
			Success:          true,
			TextChunks:       3,
			Images:           2,
			ProcessingTimeMs: 500,
			FileSize:         1024,
			RetryCount:       1,
		}

		// Save
		data, err := json.MarshalIndent(&original, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}
		err = os.WriteFile(statusFile, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Load
		loadedData, err := os.ReadFile(statusFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var loaded document.ProcessedFileStatus
		err = json.Unmarshal(loadedData, &loaded)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Verify
		if loaded.FilePath != original.FilePath {
			t.Errorf("FilePath mismatch after load")
		}
		if loaded.TextChunks != original.TextChunks {
			t.Errorf("TextChunks mismatch after load")
		}
		if loaded.Images != original.Images {
			t.Errorf("Images mismatch after load")
		}
	})
}

// TestBatchProgress tests the BatchProgress structure
func TestBatchProgress(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		progress := document.BatchProgress{
			TotalFiles:     10,
			ProcessedFiles: 0,
			SuccessFiles:   0,
			FailedFiles:    0,
			SkippedFiles:   0,
			TotalChunks:    0,
			TotalImages:    0,
			StartTime:      time.Now(),
		}

		if progress.TotalFiles != 10 {
			t.Errorf("TotalFiles mismatch: got %d, want 10", progress.TotalFiles)
		}
		if progress.ProcessedFiles != 0 {
			t.Errorf("ProcessedFiles should be 0, got %d", progress.ProcessedFiles)
		}
	})

	t.Run("UpdateProgress", func(t *testing.T) {
		progress := document.BatchProgress{
			TotalFiles:     10,
			ProcessedFiles: 0,
			SuccessFiles:   0,
			StartTime:      time.Now(),
		}

		// Simulate processing 5 files successfully
		for i := 0; i < 5; i++ {
			progress.ProcessedFiles++
			progress.SuccessFiles++
			progress.TotalChunks += 3
		}

		if progress.ProcessedFiles != 5 {
			t.Errorf("ProcessedFiles mismatch: got %d, want 5", progress.ProcessedFiles)
		}
		if progress.SuccessFiles != 5 {
			t.Errorf("SuccessFiles mismatch: got %d, want 5", progress.SuccessFiles)
		}
		if progress.TotalChunks != 15 {
			t.Errorf("TotalChunks mismatch: got %d, want 15", progress.TotalChunks)
		}

		// Estimate remaining time
		elapsed := time.Since(progress.StartTime)
		avgTimePerFile := elapsed / time.Duration(progress.ProcessedFiles)
		remainingFiles := progress.TotalFiles - progress.ProcessedFiles
		estimatedRemaining := avgTimePerFile * time.Duration(remainingFiles)

		if estimatedRemaining <= 0 {
			t.Errorf("Estimated remaining time should be positive")
		}
	})
}

// TestBatchDirectoryScanning tests directory scanning functionality
func TestBatchDirectoryScanning(t *testing.T) {
	t.Run("ScanEmptyDirectory", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-scan-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Should return empty list
		files := []string{}
		supportedExts := map[string]bool{
			".txt": true, ".md": true, ".pdf": true,
		}

		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if supportedExts[ext] {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		if len(files) != 0 {
			t.Errorf("Expected 0 files, got %d", len(files))
		}
	})

	t.Run("ScanWithSupportedFiles", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-scan-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create test files
		testFiles := []string{
			"test1.txt",
			"test2.md",
			"test3.pdf",
			"test4.jpg",
			"test5.png",
			"ignore.xyz", // Should be ignored
		}

		for _, file := range testFiles {
			path := filepath.Join(tmpDir, file)
			err := os.WriteFile(path, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", file, err)
			}
		}

		// Scan directory
		files := []string{}
		supportedExts := map[string]bool{
			".txt": true, ".md": true, ".pdf": true,
			".jpg": true, ".png": true,
		}

		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if supportedExts[ext] {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Should find 5 files (excluding .xyz)
		if len(files) != 5 {
			t.Errorf("Expected 5 files, got %d", len(files))
		}
	})

	t.Run("ScanNestedDirectories", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-scan-nested-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create nested structure
		subDir1 := filepath.Join(tmpDir, "sub1")
		subDir2 := filepath.Join(tmpDir, "sub2")
		os.MkdirAll(subDir1, 0755)
		os.MkdirAll(subDir2, 0755)

		// Create files in different directories
		files := []string{
			filepath.Join(tmpDir, "root.txt"),
			filepath.Join(subDir1, "sub1.txt"),
			filepath.Join(subDir2, "sub2.txt"),
		}

		for _, file := range files {
			err := os.WriteFile(file, []byte("test"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", file, err)
			}
		}

		// Scan
		foundFiles := []string{}
		supportedExts := map[string]bool{".txt": true}

		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if supportedExts[ext] {
				foundFiles = append(foundFiles, path)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Should find all 3 files
		if len(foundFiles) != 3 {
			t.Errorf("Expected 3 files, got %d", len(foundFiles))
		}
	})
}

// TestBatchReportGeneration tests CSV report generation
func TestBatchReportGeneration(t *testing.T) {
	t.Run("WriteCSVReport", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-report-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		reportPath := filepath.Join(tmpDir, "report.csv")

		// Create sample statuses
		statuses := []document.ProcessedFileStatus{
			{
				FilePath:         "/path/to/file1.txt",
				ProcessedAt:      time.Now(),
				Success:          true,
				TextChunks:       5,
				Images:           0,
				ProcessingTimeMs: 1000,
				FileSize:         5000,
				RetryCount:       0,
			},
			{
				FilePath:         "/path/to/file2.pdf",
				ProcessedAt:      time.Now(),
				Success:          true,
				TextChunks:       10,
				Images:           3,
				ProcessingTimeMs: 2000,
				FileSize:         15000,
				RetryCount:       1,
			},
			{
				FilePath:         "/path/to/file3.txt",
				ProcessedAt:      time.Now(),
				Success:          false,
				Error:            "Failed to process",
				ProcessingTimeMs: 500,
				FileSize:         1000,
				RetryCount:       2,
			},
		}

		// Write report
		file, err := os.Create(reportPath)
		if err != nil {
			t.Fatalf("Failed to create report file: %v", err)
		}
		defer file.Close()

		// Write header
		header := "Timestamp,File,Status,TextChunks,Images,ProcessingTimeMs,FileSize,RetryCount,ChunksFailed,ImagesFailed,Error\n"
		file.WriteString(header)

		// Write each status
		for _, status := range statuses {
			statusStr := "success"
			if !status.Success {
				statusStr = "failed"
			}

			line := status.ProcessedAt.Format("2006-01-02 15:04:05") + "," +
				filepath.Base(status.FilePath) + "," +
				statusStr + "," +
				"," + "," + "," + "," + "," + "," + "," +
				status.Error + "\n"
			file.WriteString(line)
		}

		// Verify file exists
		if _, err := os.Stat(reportPath); os.IsNotExist(err) {
			t.Errorf("Report file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(reportPath)
		if err != nil {
			t.Fatalf("Failed to read report: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Timestamp") {
			t.Errorf("Report should contain header")
		}
		if !strings.Contains(contentStr, "file1.txt") {
			t.Errorf("Report should contain file1.txt")
		}
	})
}

// TestBatchFileTypeDetection tests file type detection
func TestBatchFileTypeDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"document.txt", "text"},
		{"document.md", "text"},
		{"document.pdf", "pdf"},
		{"image.jpg", "image"},
		{"image.jpeg", "image"},
		{"image.png", "image"},
		{"image.gif", "image"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			ext := filepath.Ext(tt.filename)
			ext = strings.ToLower(ext)

			var fileType string
			switch ext {
			case ".pdf":
				fileType = "pdf"
			case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
				fileType = "image"
			default:
				fileType = "text"
			}

			if fileType != tt.expected {
				t.Errorf("File type for %s: got %s, want %s", tt.filename, fileType, tt.expected)
			}
		})
	}
}

// TestBatchProcessedFileFilter tests filtering of already processed files
func TestBatchProcessedFileFilter(t *testing.T) {
	t.Run("FilterProcessedFiles", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "batch-filter-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create test files
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		file3 := filepath.Join(tmpDir, "file3.txt")

		os.WriteFile(file1, []byte("content1"), 0644)
		os.WriteFile(file2, []byte("content2"), 0644)
		os.WriteFile(file3, []byte("content3"), 0644)

		// Create .processed file for file1 (success)
		status1 := document.ProcessedFileStatus{
			FilePath:    file1,
			ProcessedAt: time.Now(),
			Success:     true,
		}
		data1, _ := json.MarshalIndent(&status1, "", "  ")
		os.WriteFile(file1+".processed", data1, 0644)

		// Create .processed file for file2 (failed)
		status2 := document.ProcessedFileStatus{
			FilePath:    file2,
			ProcessedAt: time.Now(),
			Success:     false,
			Error:       "Test error",
		}
		data2, _ := json.MarshalIndent(&status2, "", "  ")
		os.WriteFile(file2+".processed", data2, 0644)

		// No .processed file for file3

		// Filter files
		allFiles := []string{file1, file2, file3}
		var unprocessed []string

		for _, file := range allFiles {
			processedFile := file + ".processed"
			if _, err := os.Stat(processedFile); os.IsNotExist(err) {
				unprocessed = append(unprocessed, file)
			} else {
				// Check if processed file is valid
				data, err := os.ReadFile(processedFile)
				if err != nil {
					unprocessed = append(unprocessed, file)
					continue
				}
				var status document.ProcessedFileStatus
				err = json.Unmarshal(data, &status)
				if err != nil || !status.Success {
					unprocessed = append(unprocessed, file)
				}
			}
		}

		// Should have file2 (failed) and file3 (not processed)
		if len(unprocessed) != 2 {
			t.Errorf("Expected 2 unprocessed files, got %d", len(unprocessed))
		}

		// Verify file2 and file3 are in unprocessed
		hasFile2 := false
		hasFile3 := false
		for _, f := range unprocessed {
			if f == file2 {
				hasFile2 = true
			}
			if f == file3 {
				hasFile3 = true
			}
		}

		if !hasFile2 {
			t.Errorf("file2 should be in unprocessed (failed)")
		}
		if !hasFile3 {
			t.Errorf("file3 should be in unprocessed (not processed)")
		}
	})
}
