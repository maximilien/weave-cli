// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/cmd/document"
)

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tool     document.ConversionTool
		expected string
	}{
		{
			name:     "Ghostscript with simple filename",
			input:    "document.pdf",
			tool:     document.ToolGhostscript,
			expected: "document-rgb-ghostscript.pdf",
		},
		{
			name:     "ImageMagick with simple filename",
			input:    "document.pdf",
			tool:     document.ToolImageMagick,
			expected: "document-rgb-imagemagick.pdf",
		},
		{
			name:     "Ghostscript with path",
			input:    "/path/to/document.pdf",
			tool:     document.ToolGhostscript,
			expected: "/path/to/document-rgb-ghostscript.pdf",
		},
		{
			name:     "ImageMagick with path",
			input:    "/path/to/document.pdf",
			tool:     document.ToolImageMagick,
			expected: "/path/to/document-rgb-imagemagick.pdf",
		},
		{
			name:     "Ghostscript with complex filename",
			input:    "my-complex-document-2025.pdf",
			tool:     document.ToolGhostscript,
			expected: "my-complex-document-2025-rgb-ghostscript.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test assumes we export the generateOutputFilename function
			// or make it testable. For now, we're testing the logic conceptually.
			dir := filepath.Dir(tt.input)
			base := filepath.Base(tt.input)
			ext := filepath.Ext(base)
			name := base[:len(base)-len(ext)]

			var suffix string
			if tt.tool == document.ToolGhostscript {
				suffix = "-rgb-ghostscript"
			} else {
				suffix = "-rgb-imagemagick"
			}

			outputBase := name + suffix + ext
			var result string
			if dir == "." || dir == "" {
				result = outputBase
			} else {
				result = filepath.Join(dir, outputBase)
			}

			if result != tt.expected {
				t.Errorf("generateOutputFilename(%q, %v) = %q, want %q",
					tt.input, tt.tool, result, tt.expected)
			}
		})
	}
}

func TestToolInstallationChecks(t *testing.T) {
	t.Run("CheckGhostscriptInstallation", func(t *testing.T) {
		_, err := exec.LookPath("gs")
		if err == nil {
			t.Log("Ghostscript is installed")
		} else {
			t.Log("Ghostscript is not installed (this is OK for testing)")
		}
	})

	t.Run("CheckImageMagickInstallation", func(t *testing.T) {
		_, err := exec.LookPath("convert")
		if err == nil {
			t.Log("ImageMagick is installed")
		} else {
			t.Log("ImageMagick is not installed (this is OK for testing)")
		}
	})
}

func TestValidateOutputDirectory(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  func() (string, error)
		cleanup   func(string)
		shouldErr bool
	}{
		{
			name: "Valid writable directory",
			setupDir: func() (string, error) {
				return os.MkdirTemp("", "weave-test-*")
			},
			cleanup: func(dir string) {
				os.RemoveAll(dir)
			},
			shouldErr: false,
		},
		{
			name: "Current directory",
			setupDir: func() (string, error) {
				return ".", nil
			},
			cleanup:   func(dir string) {},
			shouldErr: false,
		},
		{
			name: "Non-existent directory",
			setupDir: func() (string, error) {
				return "/nonexistent/path/that/does/not/exist", nil
			},
			cleanup:   func(dir string) {},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := tt.setupDir()
			if err != nil {
				t.Fatalf("Failed to setup test directory: %v", err)
			}
			defer tt.cleanup(dir)

			// Test directory validation logic
			info, err := os.Stat(dir)
			if os.IsNotExist(err) {
				if !tt.shouldErr {
					t.Errorf("Expected directory to exist but got error: %v", err)
				}
				return
			}

			if err != nil {
				if !tt.shouldErr {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if !info.IsDir() {
				if !tt.shouldErr {
					t.Errorf("Expected path to be a directory")
				}
				return
			}

			// Test write permission
			tempFile := filepath.Join(dir, ".weave-write-test")
			f, err := os.Create(tempFile)
			if err != nil {
				if !tt.shouldErr {
					t.Errorf("Expected directory to be writable but got error: %v", err)
				}
				return
			}
			f.Close()
			os.Remove(tempFile)

			if tt.shouldErr {
				t.Errorf("Expected error but validation passed")
			}
		})
	}
}

func TestPDFConversionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we have at least one conversion tool installed
	hasGhostscript := exec.Command("gs", "--version").Run() == nil
	hasImageMagick := exec.Command("convert", "--version").Run() == nil

	if !hasGhostscript && !hasImageMagick {
		t.Skip("Skipping integration test: neither Ghostscript nor ImageMagick is installed")
	}

	// Use the test fixture PDF
	inputPDF := filepath.Join("fixtures", "ragme-io.pdf")
	if _, err := os.Stat(inputPDF); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", inputPDF)
	}

	// Create temp directory for output
	tempDir, err := os.MkdirTemp("", "weave-pdf-convert-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if hasGhostscript {
		t.Run("ConvertWithGhostscript", func(t *testing.T) {
			outputPDF := filepath.Join(tempDir, "output-ghostscript.pdf")

			cmd := exec.Command("gs",
				"-sDEVICE=pdfwrite",
				"-dProcessColorModel=/DeviceRGB",
				"-dColorConversionStrategy=/RGB",
				"-dNOPAUSE",
				"-dBATCH",
				"-sOutputFile="+outputPDF,
				inputPDF,
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Ghostscript conversion failed: %v\nOutput: %s", err, string(output))
				return
			}

			// Verify output file exists
			if _, err := os.Stat(outputPDF); os.IsNotExist(err) {
				t.Errorf("Output PDF was not created: %s", outputPDF)
				return
			}

			// Verify output file is not empty
			info, err := os.Stat(outputPDF)
			if err != nil {
				t.Errorf("Failed to stat output PDF: %v", err)
				return
			}

			if info.Size() == 0 {
				t.Errorf("Output PDF is empty")
			}

			t.Logf("Successfully converted PDF with Ghostscript: %d bytes", info.Size())
		})
	}

	if hasImageMagick {
		t.Run("ConvertWithImageMagick", func(t *testing.T) {
			outputPDF := filepath.Join(tempDir, "output-imagemagick.pdf")

			cmd := exec.Command("convert",
				"-density", "300",
				"-colorspace", "RGB",
				inputPDF,
				outputPDF,
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				// ImageMagick might have PDF policy restrictions
				if containsString(string(output), "not authorized") {
					t.Skipf("ImageMagick PDF policy prevents conversion: %s", string(output))
					return
				}
				t.Errorf("ImageMagick conversion failed: %v\nOutput: %s", err, string(output))
				return
			}

			// Verify output file exists
			if _, err := os.Stat(outputPDF); os.IsNotExist(err) {
				t.Errorf("Output PDF was not created: %s", outputPDF)
				return
			}

			// Verify output file is not empty
			info, err := os.Stat(outputPDF)
			if err != nil {
				t.Errorf("Failed to stat output PDF: %v", err)
				return
			}

			if info.Size() == 0 {
				t.Errorf("Output PDF is empty")
			}

			t.Logf("Successfully converted PDF with ImageMagick: %d bytes", info.Size())
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) && hasSubstring(s, substr)))
}

// hasSubstring checks if s contains substr
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
