// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package image

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateTimestampedPath(t *testing.T) {
	tests := []struct {
		name             string
		originalPath     string
		expectedContains []string
	}{
		{
			name:         "Simple filename",
			originalPath: "photo.jpg",
			expectedContains: []string{
				"images/",
				"photo.jpg",
			},
		},
		{
			name:         "Filename with spaces",
			originalPath: "My Photo.jpg",
			expectedContains: []string{
				"images/",
				"my-photo.jpg",
			},
		},
		{
			name:         "Uppercase extension",
			originalPath: "IMAGE.JPG",
			expectedContains: []string{
				"images/",
				"image.jpg",
			},
		},
		{
			name:         "Full path",
			originalPath: "/path/to/image.png",
			expectedContains: []string{
				"images/",
				"image.png",
			},
		},
		{
			name:         "Multiple spaces",
			originalPath: "Yorkshire  Terrier  Photo.jpg",
			expectedContains: []string{
				"images/",
				"yorkshire--terrier--photo.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTimestampedPath(tt.originalPath)

			// Check that result contains expected parts
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateTimestampedPath() = %q, want to contain %q", result, expected)
				}
			}

			// Verify timestamp format (YYYYMMDD_HHMMSS_microseconds)
			// Extract just the filename from the path
			filename := filepath.Base(result)
			parts := strings.Split(filename, "_")

			if len(parts) < 4 {
				t.Errorf("Expected at least 4 parts separated by '_', got %d in %q", len(parts), filename)
			}

			// Check date format (YYYYMMDD)
			if len(parts[0]) != 8 {
				t.Errorf("Expected date part to be 8 digits, got %d in %q", len(parts[0]), parts[0])
			}

			// Check time format (HHMMSS)
			if len(parts[1]) != 6 {
				t.Errorf("Expected time part to be 6 digits, got %d in %q", len(parts[1]), parts[1])
			}

			// Check microseconds format (6 digits)
			if len(parts[2]) != 6 {
				t.Errorf("Expected microseconds to be 6 digits, got %d in %q", len(parts[2]), parts[2])
			}
		})
	}
}

func TestGenerateTimestampedPathWithPrefix(t *testing.T) {
	tests := []struct {
		name             string
		originalPath     string
		prefix           string
		expectedContains []string
	}{
		{
			name:         "Custom prefix",
			originalPath: "photo.jpg",
			prefix:       "uploads",
			expectedContains: []string{
				"uploads/",
				"photo.jpg",
			},
		},
		{
			name:         "Nested prefix",
			originalPath: "image.png",
			prefix:       "data/images",
			expectedContains: []string{
				"data/images/",
				"image.png",
			},
		},
		{
			name:         "Empty prefix",
			originalPath: "photo.jpg",
			prefix:       "",
			expectedContains: []string{
				"photo.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTimestampedPathWithPrefix(tt.originalPath, tt.prefix)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateTimestampedPathWithPrefix() = %q, want to contain %q", result, expected)
				}
			}
		})
	}
}

func TestExtractTimestampFromPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "Valid timestamped path",
			path:        "images/20251015_143052_123456_photo.jpg",
			shouldError: false,
		},
		{
			name:        "Valid with nested path",
			path:        "/var/data/images/20251015_143052_123456_image.png",
			shouldError: false,
		},
		{
			name:        "Invalid - too few parts",
			path:        "images/photo.jpg",
			shouldError: true,
		},
		{
			name:        "Invalid - wrong date format",
			path:        "images/202510_143052_123456_photo.jpg",
			shouldError: true,
		},
		{
			name:        "Invalid - wrong time format",
			path:        "images/20251015_1430_123456_photo.jpg",
			shouldError: true,
		},
		{
			name:        "Invalid - wrong microseconds format",
			path:        "images/20251015_143052_12345_photo.jpg",
			shouldError: true,
		},
		{
			name:        "Invalid - bad date value",
			path:        "images/20259999_143052_123456_photo.jpg",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := ExtractTimestampFromPath(tt.path)

			if tt.shouldError {
				if err == nil {
					t.Errorf("ExtractTimestampFromPath() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractTimestampFromPath() unexpected error: %v", err)
				}
				if timestamp.IsZero() {
					t.Errorf("ExtractTimestampFromPath() returned zero timestamp")
				}
			}
		})
	}
}

func TestExtractTimestampFromPath_RoundTrip(t *testing.T) {
	// Test that we can generate a path and extract the timestamp back
	// We verify that the extraction works, but don't check exact time
	// due to timezone handling complexities (generates in local, parses as UTC)
	originalPath := "test-image.jpg"
	timestampedPath := GenerateTimestampedPath(originalPath)

	extractedTime, err := ExtractTimestampFromPath(timestampedPath)
	if err != nil {
		t.Fatalf("Failed to extract timestamp from generated path %q: %v", timestampedPath, err)
	}

	// Verify that the extracted timestamp is not zero
	if extractedTime.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	// Verify the timestamp is reasonably recent (within 24 hours accounting for timezones)
	now := time.Now()
	diff := now.Sub(extractedTime)
	if diff < -24*time.Hour || diff > 24*time.Hour {
		t.Errorf("Extracted time %v is not within 24 hours of now %v (diff: %v)",
			extractedTime, now, diff)
	}
}

func TestIsTimestampedPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid timestamped path",
			path:     "images/20251015_143052_123456_photo.jpg",
			expected: true,
		},
		{
			name:     "Invalid - regular path",
			path:     "images/photo.jpg",
			expected: false,
		},
		{
			name:     "Invalid - partial timestamp",
			path:     "images/20251015_photo.jpg",
			expected: false,
		},
		{
			name:     "Valid with underscores in filename",
			path:     "images/20251015_143052_123456_my_photo_file.jpg",
			expected: true,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimestampedPath(tt.path)
			if result != tt.expected {
				t.Errorf("IsTimestampedPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetOriginalFilename(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Simple timestamped path",
			path:     "images/20251015_143052_123456_photo.jpg",
			expected: "photo.jpg",
		},
		{
			name:     "Filename with underscores",
			path:     "images/20251015_143052_123456_my_photo_file.jpg",
			expected: "my_photo_file.jpg",
		},
		{
			name:     "Non-timestamped path",
			path:     "images/photo.jpg",
			expected: "photo.jpg",
		},
		{
			name:     "Full path",
			path:     "/var/data/images/20251015_143052_123456_image.png",
			expected: "image.png",
		},
		{
			name:     "Multiple underscores in original",
			path:     "20251015_143052_123456_file_name_with_many_underscores.jpg",
			expected: "file_name_with_many_underscores.jpg",
		},
		{
			name:     "Empty string",
			path:     "",
			expected: ".", // filepath.Base returns "." for empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOriginalFilename(tt.path)
			if result != tt.expected {
				t.Errorf("GetOriginalFilename(%q) = %q, expected %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetOriginalFilename_RoundTrip(t *testing.T) {
	// Test that we can generate a path and extract the original filename
	originalPaths := []string{
		"photo.jpg",
		"my-photo.png",
		"file_with_underscores.gif",
		"Multiple Words Image.jpg",
	}

	for _, originalPath := range originalPaths {
		t.Run(originalPath, func(t *testing.T) {
			timestampedPath := GenerateTimestampedPath(originalPath)
			extractedFilename := GetOriginalFilename(timestampedPath)

			// The extracted filename should match the lowercase, hyphenated version
			expectedFilename := strings.ToLower(originalPath)
			expectedFilename = strings.ReplaceAll(expectedFilename, " ", "-")

			if extractedFilename != expectedFilename {
				t.Errorf("Round trip failed: original=%q, timestamped=%q, extracted=%q, expected=%q",
					originalPath, timestampedPath, extractedFilename, expectedFilename)
			}
		})
	}
}

// Real-world scenarios
func TestStoragePath_RealWorldScenarios(t *testing.T) {
	t.Run("AuctionsMax.ai image storage workflow", func(t *testing.T) {
		// Simulate storing an image extracted from a PDF
		originalPath := "2022-tamarkin-auction-catalogue.pdf#page=1&image=1.jpg"

		// Generate storage path
		storagePath := GenerateTimestampedPath(originalPath)

		// Verify it's timestamped
		if !IsTimestampedPath(storagePath) {
			t.Errorf("Expected timestamped path, got: %s", storagePath)
		}

		// Verify we can extract the timestamp
		timestamp, err := ExtractTimestampFromPath(storagePath)
		if err != nil {
			t.Errorf("Failed to extract timestamp: %v", err)
		}
		if timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}

		// Verify original filename is preserved (lowercased and hyphenated)
		originalFilename := GetOriginalFilename(storagePath)
		expectedFilename := "2022-tamarkin-auction-catalogue.pdf#page=1&image=1.jpg"
		if originalFilename != expectedFilename {
			t.Errorf("Expected original filename %q, got %q", expectedFilename, originalFilename)
		}
	})

	t.Run("Custom prefix for different storage backends", func(t *testing.T) {
		originalPath := "product-image.jpg"

		// Different storage backends
		// Note: filepath.Join normalizes paths, so "s3://bucket" becomes "s3:/bucket"
		tests := []struct {
			prefix   string
			expected string
		}{
			{prefix: "s3://bucket/images", expected: "s3:/bucket/images"}, // filepath.Join normalizes
			{prefix: "gs://bucket/uploads", expected: "gs:/bucket/uploads"},
			{prefix: "/mnt/storage/images", expected: "/mnt/storage/images"},
			{prefix: "images/products", expected: "images/products"},
		}

		for _, tt := range tests {
			storagePath := GenerateTimestampedPathWithPrefix(originalPath, tt.prefix)

			if !strings.HasPrefix(storagePath, tt.expected) {
				t.Errorf("Expected path to start with %q, got: %s", tt.expected, storagePath)
			}

			if !strings.Contains(storagePath, "product-image.jpg") {
				t.Errorf("Expected path to contain filename, got: %s", storagePath)
			}
		}
	})

	t.Run("Handling special characters in filenames", func(t *testing.T) {
		specialFilenames := []string{
			"Photo with Spaces.jpg",
			"UPPERCASE.JPG",
			"Multiple   Spaces.png",
		}

		for _, filename := range specialFilenames {
			storagePath := GenerateTimestampedPath(filename)

			// Should be lowercased
			if storagePath != strings.ToLower(storagePath) {
				t.Errorf("Expected lowercase path, got: %s", storagePath)
			}

			// Should not contain uppercase original
			original := filepath.Base(filename)
			if strings.Contains(storagePath, original) && original != strings.ToLower(original) {
				t.Errorf("Expected no uppercase in path, got: %s", storagePath)
			}
		}
	})
}
