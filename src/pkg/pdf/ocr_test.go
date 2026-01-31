// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractOCRText(t *testing.T) {
	// Create a temporary test image
	tempFile, err := os.CreateTemp("", "test_image_*.png")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Create a simple test image with white background
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for x := 0; x < 200; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.White)
		}
	}

	// Save the image
	err = png.Encode(tempFile, img)
	assert.NoError(t, err)
	tempFile.Close()

	// Test OCR extraction
	text := extractOCRText(tempFile.Name())

	// For a blank image, OCR might return empty string or whitespace
	// This is expected behavior
	t.Logf("OCR result for blank image: '%s'", text)
	assert.True(t, len(text) == 0 || strings.TrimSpace(text) == "",
		"Blank image should produce empty or whitespace text")
}

func TestExtractOCRText_NonExistentFile(t *testing.T) {
	// Test with non-existent file
	text := extractOCRText("/nonexistent/file.png")

	// Should return empty string on error (non-fatal)
	assert.Equal(t, "", text, "Non-existent file should return empty string")
}

func TestExtractOCRText_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires actual PDF images with text
	// It will be skipped if no test images are available
	testImagePath := os.Getenv("TEST_IMAGE_PATH")
	if testImagePath == "" {
		t.Skip("TEST_IMAGE_PATH not set, skipping OCR integration test")
	}

	// Check if file exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skip("Test image not found, skipping OCR integration test")
	}

	// Extract OCR text
	text := extractOCRText(testImagePath)

	t.Logf("OCR result length: %d characters", len(text))
	t.Logf("OCR result preview: %s", text[:minInt(len(text), 100)])

	// For real images with text, we expect some content
	assert.NotEmpty(t, text, "Real image should produce some OCR text")
	assert.Greater(t, len(text), 3, "OCR text should be longer than 3 characters")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark OCR performance
func BenchmarkExtractOCRText(b *testing.B) {
	// Create a test image
	tempFile, err := os.CreateTemp("", "bench_image_*.png")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for x := 0; x < 200; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.White)
		}
	}
	png.Encode(tempFile, img)
	tempFile.Close()

	// Benchmark OCR extraction
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractOCRText(tempFile.Name())
	}
}
