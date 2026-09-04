//go:build cgo

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pdf

import (
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// extractOCRText extracts text from an image using Tesseract OCR.
func extractOCRText(imagePath string) string {
	client := gosseract.NewClient()
	defer client.Close()

	if err := client.SetImage(imagePath); err != nil {
		return ""
	}

	text, err := client.Text()
	if err != nil {
		return ""
	}

	text = strings.TrimSpace(text)
	if len(text) < 3 {
		return ""
	}

	return text
}
