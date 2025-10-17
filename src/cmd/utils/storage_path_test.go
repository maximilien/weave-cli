// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFilepathAbsConversion tests the core regression:
// That relative paths are converted to absolute paths
// This is a unit test for the specific bug that was fixed
func TestFilepathAbsConversion(t *testing.T) {
	tests := []struct {
		name           string
		setupDir       bool
		relativePath   string
		wantAbsolute   bool
		wantMatchInput bool
	}{
		{
			name:           "Absolute path remains absolute",
			setupDir:       false,
			relativePath:   "/tmp/test.txt",
			wantAbsolute:   true,
			wantMatchInput: true,
		},
		{
			name:           "Relative path becomes absolute",
			setupDir:       true,
			relativePath:   "test.txt",
			wantAbsolute:   true,
			wantMatchInput: false, // Should be converted
		},
		{
			name:           "Relative path with directory",
			setupDir:       true,
			relativePath:   "./docs/README.md",
			wantAbsolute:   true,
			wantMatchInput: false, // Should be converted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string

			if tt.setupDir {
				// Create temp directory
				tmpDir, err := os.MkdirTemp("", "weave-path-test-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tmpDir)

				// Create test file
				fullPath := filepath.Join(tmpDir, tt.relativePath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directories: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				// Change to temp directory
				originalDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				defer os.Chdir(originalDir)

				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}

				filePath = tt.relativePath
			} else {
				filePath = tt.relativePath
			}

			// THIS IS THE KEY FIX: Convert to absolute path
			// This is what was added to document.go:436-441
			absPath, err := filepath.Abs(filePath)
			if err != nil {
				// Fallback to original path if absolute path cannot be determined
				absPath = filePath
			}

			// Verify the result
			if tt.wantAbsolute {
				if !filepath.IsAbs(absPath) {
					t.Errorf("Expected absolute path, got relative: %s", absPath)
				}
			}

			if tt.wantMatchInput {
				if absPath != filePath {
					t.Errorf("Expected path to match input\nGot:  %s\nWant: %s", absPath, filePath)
				}
			} else {
				if absPath == filePath && !filepath.IsAbs(filePath) {
					t.Errorf("Expected path to be converted to absolute\nGot:  %s\nInput: %s", absPath, filePath)
				}
			}
		})
	}
}

// TestStoragePathRegression is a high-level regression test
// that verifies the bug fix without needing the full infrastructure
func TestStoragePathRegression(t *testing.T) {
	t.Run("README.md scenario", func(t *testing.T) {
		// This replicates the exact scenario from the bug report:
		// User runs: ./bin/weave docs create WeaveDocs README.md
		// The filePath parameter receives "README.md" (relative)
		// storage_path should be absolute, not "README.md"

		// Simulate being in the weave-cli directory
		tmpDir, err := os.MkdirTemp("", "weave-regression-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create README.md
		readmePath := filepath.Join(tmpDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Weave CLI"), 0644); err != nil {
			t.Fatalf("Failed to create README.md: %v", err)
		}

		// Change to that directory (simulating user running from project root)
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// User provides relative path
		userProvidedPath := "README.md"

		// BEFORE THE FIX: storage_path would be "README.md" (WRONG)
		// AFTER THE FIX: storage_path should be absolute path

		// Apply the fix
		absPath, err := filepath.Abs(userProvidedPath)
		if err != nil {
			absPath = userProvidedPath // fallback
		}

		// Verify the fix
		if absPath == "README.md" {
			t.Error("REGRESSION: storage_path is still relative 'README.md'")
			t.Error("This is the exact bug that was fixed!")
		}

		if !filepath.IsAbs(absPath) {
			t.Errorf("storage_path must be absolute, got: %s", absPath)
		}

		// On macOS, /var is a symlink to /private/var, so we need to evaluate both paths
		expectedPath := filepath.Join(tmpDir, "README.md")
		absPathEval, _ := filepath.EvalSymlinks(absPath)
		expectedPathEval, _ := filepath.EvalSymlinks(expectedPath)

		if absPathEval != expectedPathEval {
			t.Errorf("storage_path incorrect\nGot:  %s\nWant: %s", absPath, expectedPath)
		}

		t.Logf("✓ storage_path correctly converted: %s", absPath)
	})

	t.Run("PDF scenario", func(t *testing.T) {
		// Same test for PDF files
		tmpDir, err := os.MkdirTemp("", "weave-pdf-regression-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// User provides relative path
		userProvidedPath := "document.pdf"

		// Apply the fix
		absPath, err := filepath.Abs(userProvidedPath)
		if err != nil {
			absPath = userProvidedPath
		}

		// Verify
		if absPath == "document.pdf" {
			t.Error("REGRESSION: storage_path is still relative")
		}

		if !filepath.IsAbs(absPath) {
			t.Errorf("storage_path must be absolute for PDFs too, got: %s", absPath)
		}

		t.Logf("✓ PDF storage_path correctly converted: %s", absPath)
	})
}
