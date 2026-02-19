// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"testing"
)

func TestParsePageNumberFromFilename(t *testing.T) {
	tests := []struct {
		name       string
		imagePath  string
		imageIndex int
		wantPage   int
	}{
		{
			name:       "pdfcpu standard format single digit page",
			imagePath:  "/tmp/pdf_images_123/catalogue_001_Im0.jpg",
			imageIndex: 0,
			wantPage:   1,
		},
		{
			name:       "pdfcpu standard format double digit page",
			imagePath:  "/tmp/pdf_images_123/catalogue_012_Im3.png",
			imageIndex: 5,
			wantPage:   12,
		},
		{
			name:       "pdfcpu standard format triple digit page",
			imagePath:  "/tmp/pdf_images_123/tamarkin-2023_044_Im1.jpg",
			imageIndex: 10,
			wantPage:   44,
		},
		{
			name:       "pdfcpu format with hyphen in pdf name",
			imagePath:  "/tmp/catalogue-2022_003_Im0.jpg",
			imageIndex: 2,
			wantPage:   3,
		},
		{
			name:       "fallback: no underscore pattern",
			imagePath:  "/tmp/image.jpg",
			imageIndex: 4,
			wantPage:   5, // fallback: imageIndex + 1
		},
		{
			name:       "fallback: non-numeric page segment",
			imagePath:  "/tmp/catalogue_abc_Im0.jpg",
			imageIndex: 2,
			wantPage:   3, // fallback: imageIndex + 1
		},
		{
			name:       "fallback: only two parts",
			imagePath:  "/tmp/catalogue_001.jpg",
			imageIndex: 0,
			wantPage:   1, // fallback: imageIndex + 1 = 1, happens to match
		},
		{
			name:       "page zero rejected, falls back",
			imagePath:  "/tmp/catalogue_000_Im0.jpg",
			imageIndex: 3,
			wantPage:   4, // fallback: 0 is not > 0, so imageIndex + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePageNumberFromFilename(tt.imagePath, tt.imageIndex)
			if got != tt.wantPage {
				t.Errorf("parsePageNumberFromFilename(%q, %d) = %d, want %d",
					tt.imagePath, tt.imageIndex, got, tt.wantPage)
			}
		})
	}
}
