// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"fmt"
)

// PDFImageData represents an extracted image from a PDF
type PDFImageData struct {
	ID              string
	ImageData       string // Base64 encoded image data
	Image           string // Data URL format
	URL             string
	Metadata        map[string]interface{}
	OCRText         string
	EXIFData        map[string]interface{}
	Caption         string
	PageNumber      int
	ImageIndex      int
	SourcePDF       string
	SurroundingText string // Text context from the page where the image appears
	SectionHeading  string // Section/chapter heading if available
}

// PDFTextData represents extracted text from a PDF
type PDFTextData struct {
	ID         string
	Content    string
	URL        string
	Metadata   map[string]interface{}
	PageNumber int
	SourcePDF  string
}

// ExtractPDFContent extracts both text and images from a PDF file
func ExtractPDFContent(filePath string, chunkSize int, skipSmallImages bool, minImageSize int, maxMetadataLength int, noTips bool) ([]PDFTextData, []PDFImageData, error) {
	// Extract text content
	textData, err := extractPDFText(filePath, chunkSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Extract images
	imageData, err := extractPDFImages(filePath, skipSmallImages, minImageSize, noTips)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract images: %w", err)
	}

	// Extract page-by-page text for image context
	pageTexts, err := extractTextByPage(filePath)
	if err != nil {
		// Non-fatal - images can still be processed without context
		pageTexts = make(map[int]string)
	}

	// Enrich images with surrounding text context
	for i := range imageData {
		enrichImageWithContext(&imageData[i], pageTexts, maxMetadataLength)
	}

	return textData, imageData, nil
}
