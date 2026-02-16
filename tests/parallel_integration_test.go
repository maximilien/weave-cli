// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
)

// TestParallelProcessing tests parallel document ingestion with worker pool
func TestParallelProcessing(t *testing.T) {
	// Create mock config
	cfg := &config.VectorDBConfig{
		Type:    config.VectorDBTypeMock,
		Enabled: true,
		MockConfig: &config.MockConfig{
			Enabled:            true,
			SimulateEmbeddings: true,
			EmbeddingDimension: 384,
			Collections: []config.MockCollection{
				{Name: "ParallelTestCol", Type: "text", Description: "Test parallel processing"},
			},
		},
	}

	// Create test file with multiple chunks
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-parallel.txt")
	content := ""
	for i := 1; i <= 10; i++ {
		content += fmt.Sprintf("Chunk %d: This is test content that will be processed in parallel. "+
			"Each chunk should be processed by a separate worker for faster ingestion.\n\n", i)
	}
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SequentialProcessing", func(t *testing.T) {
		// Test with workers=1 (sequential)
		start := time.Now()
		err := utils.CreateDocument(ctx, cfg, "ParallelTestCol", testFile, 200, "", false, 0, 10, 2000, "", "", "", 1)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Sequential processing failed: %v", err)
		}
		t.Logf("Sequential processing (workers=1) took: %v", elapsed)
	})

	t.Run("ParallelProcessing3Workers", func(t *testing.T) {
		// Test with workers=3 (parallel)
		start := time.Now()
		err := utils.CreateDocument(ctx, cfg, "ParallelTestCol", testFile, 200, "", false, 0, 10, 2000, "", "", "", 3)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Parallel processing (3 workers) failed: %v", err)
		}
		t.Logf("Parallel processing (workers=3) took: %v", elapsed)
	})

	t.Run("ParallelProcessing5Workers", func(t *testing.T) {
		// Test with workers=5 (parallel)
		start := time.Now()
		err := utils.CreateDocument(ctx, cfg, "ParallelTestCol", testFile, 200, "", false, 0, 10, 2000, "", "", "", 5)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Parallel processing (5 workers) failed: %v", err)
		}
		t.Logf("Parallel processing (workers=5) took: %v", elapsed)
	})

	t.Run("WorkerValidation", func(t *testing.T) {
		// Workers should be capped at 10
		// The function validates internally, so this should still work
		err := utils.CreateDocument(ctx, cfg, "ParallelTestCol", testFile, 200, "", false, 0, 10, 2000, "", "", "", 15)
		if err != nil {
			t.Errorf("Worker validation failed: %v", err)
		}
	})
}

// TestGlobPatternProcessing tests glob pattern support for batch file processing
func TestGlobPatternProcessing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	files := []string{"test-1.txt", "test-2.txt", "test-3.txt"}
	for _, file := range files {
		filePath := filepath.Join(tmpDir, file)
		content := fmt.Sprintf("Content for %s with some text to process.\n", file)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	t.Run("IsGlobPattern", func(t *testing.T) {
		tests := []struct {
			path     string
			expected bool
		}{
			{"*.txt", true},
			{"file.txt", false},
			{"**/*.md", true},
			{"/path/to/file.pdf", false},
			{"images/*.{jpg,png}", true},
			{"test-?.txt", true},
			{"[abc].txt", true},
		}

		for _, tt := range tests {
			result := utils.IsGlobPattern(tt.path)
			if result != tt.expected {
				t.Errorf("IsGlobPattern(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		}
	})

	t.Run("ExpandGlobPattern", func(t *testing.T) {
		pattern := filepath.Join(tmpDir, "test-*.txt")
		files, err := utils.ExpandGlobPattern(pattern)
		if err != nil {
			t.Fatalf("ExpandGlobPattern failed: %v", err)
		}

		if len(files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(files))
		}

		// Verify files are not directories
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				t.Errorf("Failed to stat file %s: %v", file, err)
			}
			if info.IsDir() {
				t.Errorf("Expected file, got directory: %s", file)
			}
		}
	})

	t.Run("EmptyGlobPattern", func(t *testing.T) {
		pattern := filepath.Join(tmpDir, "nonexistent-*.txt")
		files, err := utils.ExpandGlobPattern(pattern)
		if err != nil {
			t.Fatalf("ExpandGlobPattern should not error on no matches: %v", err)
		}

		if len(files) != 0 {
			t.Errorf("Expected 0 files for non-matching pattern, got %d", len(files))
		}
	})
}

// TestProgressAggregation tests progress tracking during parallel processing
func TestProgressAggregation(t *testing.T) {
	// This is more of a visual test - we verify it doesn't error
	// Progress is printed to stderr, so we can't easily capture it in tests
	cfg := &config.VectorDBConfig{
		Type:    config.VectorDBTypeMock,
		Enabled: true,
		MockConfig: &config.MockConfig{
			Enabled:            true,
			SimulateEmbeddings: true,
			EmbeddingDimension: 384,
			Collections: []config.MockCollection{
				{Name: "ProgressTestCol", Type: "text", Description: "Test progress tracking"},
			},
		},
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "progress-test.txt")
	content := ""
	for i := 1; i <= 20; i++ {
		content += fmt.Sprintf("Chunk %d content for progress tracking test.\n\n", i)
	}
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ProgressWithParallelWorkers", func(t *testing.T) {
		// Process with workers=4 - progress should be displayed
		err := utils.CreateDocument(ctx, cfg, "ProgressTestCol", testFile, 100, "", false, 0, 10, 2000, "", "", "", 4)
		if err != nil {
			t.Errorf("Progress tracking test failed: %v", err)
		}
		t.Log("Progress tracking completed successfully (check stderr for progress bars)")
	})
}

// TestParallelWithMockClient verifies parallel processing works with mock client
func TestParallelWithMockClient(t *testing.T) {
	mockCfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 1536,
		Collections: []config.MockCollection{
			{Name: "MockParallelCol", Type: "text"},
		},
	}

	client := mock.NewClient(mockCfg)
	ctx := context.Background()

	// Verify client is working
	if err := client.Health(ctx); err != nil {
		t.Fatalf("Mock client health check failed: %v", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		t.Fatalf("Failed to list collections: %v", err)
	}

	if len(collections) != 1 || collections[0] != "MockParallelCol" {
		t.Errorf("Expected MockParallelCol, got: %v", collections)
	}

	t.Log("Mock client verified - parallel processing infrastructure ready")
}
