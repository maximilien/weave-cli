// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/pipeline"
)

// TestFileScannerBasic tests basic file scanning functionality
func TestPipelineFileScannerBasic(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"file1.txt":        "Test content 1",
		"file2.md":         "# Test Markdown",
		"file3.pdf":        "PDF content",
		"subdir/file4.txt": "Nested file",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	t.Run("ScanAllFiles", func(t *testing.T) {
		scanner := pipeline.NewFileScanner(tmpDir, "", []string{}, true)
		files, err := scanner.Scan(context.Background())

		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if len(files) != 4 {
			t.Errorf("Expected 4 files, got %d", len(files))
		}
	})

	t.Run("ScanWithGlobPattern", func(t *testing.T) {
		scanner := pipeline.NewFileScanner(tmpDir, "*.txt", []string{}, true)
		files, err := scanner.Scan(context.Background())

		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Should find file1.txt and subdir/file4.txt
		if len(files) != 2 {
			t.Errorf("Expected 2 .txt files, got %d", len(files))
		}

		for _, file := range files {
			if filepath.Ext(file.Path) != ".txt" {
				t.Errorf("Expected .txt file, got %s", file.Path)
			}
		}
	})

	t.Run("ScanWithExclusion", func(t *testing.T) {
		scanner := pipeline.NewFileScanner(tmpDir, "", []string{"*.pdf"}, true)
		files, err := scanner.Scan(context.Background())

		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Should find everything except file3.pdf
		if len(files) != 3 {
			t.Errorf("Expected 3 files (excluding .pdf), got %d", len(files))
		}

		for _, file := range files {
			if filepath.Ext(file.Path) == ".pdf" {
				t.Errorf("PDF file should have been excluded: %s", file.Path)
			}
		}
	})

	t.Run("ScanNonRecursive", func(t *testing.T) {
		scanner := pipeline.NewFileScanner(tmpDir, "", []string{}, false)
		files, err := scanner.Scan(context.Background())

		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Should only find files in root directory (not subdir)
		if len(files) != 3 {
			t.Errorf("Expected 3 files in root dir, got %d", len(files))
		}
	})
}

// TestFileScannerFileTypes tests file type detection
func TestPipelineFileScannerFileTypes(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		filename     string
		expectedType pipeline.FileType
	}{
		{"test.pdf", pipeline.FileTypePDF},
		{"test.txt", pipeline.FileTypeTXT},
		{"test.md", pipeline.FileTypeMD},
		{"test.json", pipeline.FileTypeJSON},
		{"test.yaml", pipeline.FileTypeYAML},
		{"test.yml", pipeline.FileTypeYAML},
		{"test.unknown", pipeline.FileTypeUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			fullPath := filepath.Join(tmpDir, tc.filename)
			os.WriteFile(fullPath, []byte("test content"), 0644)

			scanner := pipeline.NewFileScanner(tmpDir, tc.filename, []string{}, false)
			files, err := scanner.Scan(context.Background())

			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("Expected 1 file, got %d", len(files))
			}

			if files[0].Type != tc.expectedType {
				t.Errorf("Expected type %s, got %s", tc.expectedType, files[0].Type)
			}
		})
	}
}

// TestFileScannerHash tests file hash generation
func TestPipelineFileScannerHash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two identical files with same content
	content := "identical content"
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	os.WriteFile(file1, []byte(content), 0644)
	os.WriteFile(file2, []byte(content), 0644)

	scanner := pipeline.NewFileScanner(tmpDir, "", []string{}, false)
	files, err := scanner.Scan(context.Background())

	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// Same content should produce same hash
	if files[0].Hash != files[1].Hash {
		t.Errorf("Expected identical hashes for same content, got %s and %s",
			files[0].Hash, files[1].Hash)
	}

	// Hash should not be empty
	if files[0].Hash == "" {
		t.Error("Hash should not be empty")
	}
}

// TestProgressTracker tests progress tracking functionality
func TestPipelineProgressTracker(t *testing.T) {
	t.Run("EnabledProgress", func(t *testing.T) {
		tracker := pipeline.NewProgressTracker(true, false)

		// These should not panic
		tracker.StartScanning()
		tracker.FinishScanning(10)

		tracker.StartProcessing(10, 5, 2)
		tracker.UpdateProgress(1)
		tracker.UpdateProgress(1)
		tracker.FinishProcessing()

		tracker.StartBatching(10, 5)
		tracker.UpdateBatch(1, 2, 5)
		tracker.FinishBatching()
	})

	t.Run("QuietProgress", func(t *testing.T) {
		tracker := pipeline.NewProgressTracker(true, true)

		// These should be no-ops
		tracker.StartScanning()
		tracker.FinishScanning(10)
		tracker.ShowInfo("test info")
		tracker.ShowWarning("test warning")
		tracker.ShowError("test error")
	})

	t.Run("DisabledProgress", func(t *testing.T) {
		tracker := pipeline.NewProgressTracker(false, false)

		// These should be no-ops
		tracker.StartProcessing(10, 5, 2)
		tracker.UpdateProgress(1)
		tracker.FinishProcessing()
	})
}

// TestIngestReport tests the IngestReport structure
func TestPipelineIngestReport(t *testing.T) {
	report := &pipeline.IngestReport{
		Status:           "success",
		StartTime:        time.Now().Add(-5 * time.Second),
		EndTime:          time.Now(),
		Duration:         5.0,
		FilesScanned:     100,
		FilesProcessed:   95,
		FilesSkipped:     3,
		FilesFailed:      2,
		DocumentsCreated: 150,
		ThroughputFiles:  19.0,
		ThroughputDocs:   30.0,
		Collection:       "test-collection",
		VDBType:          "mock",
		BatchSize:        50,
		Workers:          4,
	}

	// Test basic fields
	if report.Status != "success" {
		t.Errorf("Expected status 'success', got %s", report.Status)
	}

	if report.FilesProcessed+report.FilesSkipped+report.FilesFailed != report.FilesScanned {
		t.Error("File counts don't add up correctly")
	}

	if report.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	if report.ThroughputFiles <= 0 {
		t.Error("Throughput should be positive")
	}
}
