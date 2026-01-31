// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/stretchr/testify/assert"
)

func TestBuildCombinedImageContent(t *testing.T) {
	tests := []struct {
		name     string
		imgData  pdf.PDFImageData
		expected string
	}{
		{
			name: "All fields present",
			imgData: pdf.PDFImageData{
				SectionHeading:  "Chapter 1: Introduction",
				SurroundingText: "This is the surrounding text from the page.",
				OCRText:         "Text extracted from image via OCR",
			},
			expected: "Chapter 1: Introduction\n\nThis is the surrounding text from the page.\n\nText extracted from image via OCR",
		},
		{
			name: "Only OCR text",
			imgData: pdf.PDFImageData{
				OCRText: "OCR text only",
			},
			expected: "OCR text only",
		},
		{
			name: "Only surrounding text",
			imgData: pdf.PDFImageData{
				SurroundingText: "Surrounding text only",
			},
			expected: "Surrounding text only",
		},
		{
			name: "Only section heading",
			imgData: pdf.PDFImageData{
				SectionHeading: "Section Heading",
			},
			expected: "Section Heading",
		},
		{
			name: "Section heading and OCR text",
			imgData: pdf.PDFImageData{
				SectionHeading: "Product Details",
				OCRText:        "Leica M3 Camera",
			},
			expected: "Product Details\n\nLeica M3 Camera",
		},
		{
			name: "Surrounding text and OCR text",
			imgData: pdf.PDFImageData{
				SurroundingText: "This auction features vintage cameras.",
				OCRText:         "Lot 42: Leica M3",
			},
			expected: "This auction features vintage cameras.\n\nLot 42: Leica M3",
		},
		{
			name:     "No content at all",
			imgData:  pdf.PDFImageData{},
			expected: "",
		},
		{
			name: "Empty strings should be ignored",
			imgData: pdf.PDFImageData{
				SectionHeading:  "",
				SurroundingText: "Valid text",
				OCRText:         "",
			},
			expected: "Valid text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCombinedImageContent(tt.imgData)
			assert.Equal(t, tt.expected, result, "Combined content should match expected value")
		})
	}
}

func TestBuildCombinedImageContent_RealWorldScenario(t *testing.T) {
	// Simulate a real auction catalog scenario
	imgData := pdf.PDFImageData{
		SectionHeading:  "Lot 42 - Vintage Cameras",
		SurroundingText: "Exceptional Leica M3 camera from 1954. Double stroke model with original leather case. Condition: Excellent. Estimated value: $2,000-$3,000.",
		OCRText:         "LEICA M3\nSerial #750001\nDouble Stroke\nOriginal Case Included",
	}

	result := buildCombinedImageContent(imgData)

	// Verify all components are present
	assert.Contains(t, result, "Lot 42 - Vintage Cameras")
	assert.Contains(t, result, "Exceptional Leica M3")
	assert.Contains(t, result, "LEICA M3")
	assert.Contains(t, result, "Serial #750001")

	// Verify structure (parts separated by double newline)
	assert.Contains(t, result, "\n\n")

	// Verify order (section heading first, then surrounding text, then OCR)
	firstPart := "Lot 42 - Vintage Cameras"
	secondPart := "Exceptional Leica M3"
	thirdPart := "LEICA M3"

	firstIndex := findStringIndex(result, firstPart)
	secondIndex := findStringIndex(result, secondPart)
	thirdIndex := findStringIndex(result, thirdPart)

	assert.True(t, firstIndex < secondIndex, "Section heading should come before surrounding text")
	assert.True(t, secondIndex < thirdIndex, "Surrounding text should come before OCR text")
}

// Helper function to find substring index
func findStringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
