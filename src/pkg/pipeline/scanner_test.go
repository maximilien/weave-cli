// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileScannerGlobPatterns(t *testing.T) {
	// Create temp directory structure for testing
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"test.pdf",
		"result.pdf",
		"subdir/nested.pdf",
		"subdir/result-nested.pdf",
		"deep/subdir/test.pdf",
		"test.txt",
		"result.txt",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name      string
		pattern   string
		exclude   []string
		recursive bool
		expected  []string // basenames of expected files
	}{
		{
			name:      "Simple pattern *.pdf",
			pattern:   "*.pdf",
			recursive: true,
			expected:  []string{"test.pdf", "result.pdf", "nested.pdf", "result-nested.pdf", "test.pdf"},
		},
		{
			name:      "Doublestar pattern **/*.pdf",
			pattern:   "**/*.pdf",
			recursive: true,
			expected:  []string{"test.pdf", "result.pdf", "nested.pdf", "result-nested.pdf", "test.pdf"},
		},
		{
			name:      "Exclude pattern *result*.pdf",
			pattern:   "*.pdf",
			exclude:   []string{"*result*.pdf"},
			recursive: true,
			expected:  []string{"test.pdf", "nested.pdf", "test.pdf"},
		},
		{
			name:      "Exclude doublestar pattern **/*result*.pdf",
			pattern:   "**/*.pdf",
			exclude:   []string{"**/*result*.pdf"},
			recursive: true,
			expected:  []string{"test.pdf", "nested.pdf", "test.pdf"},
		},
		{
			name:      "Non-recursive *.pdf",
			pattern:   "*.pdf",
			recursive: false,
			expected:  []string{"test.pdf", "result.pdf"},
		},
		{
			name:      "Pattern with directory subdir/*.pdf",
			pattern:   "subdir/*.pdf",
			recursive: true,
			expected:  []string{"nested.pdf", "result-nested.pdf"},
		},
		{
			name:      "Multiple extensions **/*.{pdf,txt}",
			pattern:   "**/*.txt",
			recursive: true,
			expected:  []string{"test.txt", "result.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewFileScanner(tmpDir, tt.pattern, tt.exclude, tt.recursive)
			files, err := scanner.Scan(context.Background())
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// Extract basenames for comparison
			var basenames []string
			for _, f := range files {
				basenames = append(basenames, filepath.Base(f.Path))
			}

			// Check count matches
			if len(basenames) != len(tt.expected) {
				t.Errorf("Expected %d files, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(basenames), tt.expected, basenames)
				return
			}

			// Check all expected files are present
			expectedMap := make(map[string]bool)
			for _, e := range tt.expected {
				expectedMap[e] = true
			}

			for _, basename := range basenames {
				if !expectedMap[basename] {
					t.Errorf("Unexpected file in results: %s", basename)
				}
			}
		})
	}
}

func TestFileScannerExcludePatterns(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"document.pdf",
		"result-1.pdf",
		"result-2.pdf",
		"draft/document.pdf",
		"draft/result.pdf",
		"output/final.pdf",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name     string
		exclude  []string
		expected []string // basenames
	}{
		{
			name:     "Exclude *result*.pdf files",
			exclude:  []string{"*result*.pdf"},
			expected: []string{"document.pdf", "document.pdf", "final.pdf"},
		},
		{
			name:     "Exclude draft/** files",
			exclude:  []string{"draft/**"},
			expected: []string{"document.pdf", "result-1.pdf", "result-2.pdf", "final.pdf"},
		},
		{
			name:     "Exclude draft/* files (alternative)",
			exclude:  []string{"draft/*"},
			expected: []string{"document.pdf", "result-1.pdf", "result-2.pdf", "final.pdf"},
		},
		{
			name:     "Multiple exclude patterns",
			exclude:  []string{"*result*.pdf", "draft/**"},
			expected: []string{"document.pdf", "final.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewFileScanner(tmpDir, "*.pdf", tt.exclude, true)
			files, err := scanner.Scan(context.Background())
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			var basenames []string
			for _, f := range files {
				basenames = append(basenames, filepath.Base(f.Path))
			}

			if len(basenames) != len(tt.expected) {
				t.Errorf("Expected %d files, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(basenames), tt.expected, basenames)
			}
		})
	}
}
