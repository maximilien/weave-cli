// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// extractPDFImages extracts images from a PDF file
func extractPDFImages(filePath string, skipSmallImages bool, minImageSize int, noTips bool) ([]PDFImageData, error) {
	// Create a temporary directory for image extraction
	tempDir, err := os.MkdirTemp("", "pdf_images_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract images using pdfcpu with custom configuration
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	err = api.ExtractImagesFile(filePath, tempDir, []string{}, conf)
	if err != nil {
		// If extraction fails due to unsupported JPEG features, try alternative approach
		if strings.Contains(err.Error(), "unsupported JPEG feature") ||
			strings.Contains(err.Error(), "unknown color model") {
			fmt.Printf("⚠️  Warning: Some images have unsupported JPEG features (CMYK without APP14 metadata)\n")
			fmt.Printf("📝 Attempting alternative image extraction method...\n")
			return extractImagesWithFallback(filePath, tempDir, skipSmallImages, minImageSize, noTips)
		}
		return nil, fmt.Errorf("pdfcpu image extraction failed: %w", err)
	}

	// Find extracted image files
	var imageFiles []string
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".bmp" {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk temp directory: %w", err)
	}

	var imageData []PDFImageData
	for i, imagePath := range imageFiles {
		// Process each extracted image
		data, err := processExtractedImage(imagePath, filePath, i)
		if err != nil {
			continue // Skip problematic images
		}

		// Apply size filter if specified
		if skipSmallImages && len(data.ImageData) < minImageSize {
			continue
		}

		imageData = append(imageData, *data)
	}

	return imageData, nil
}

// extractImagesWithFallback attempts to extract images using pdfimages (poppler-utils)
func extractImagesWithFallback(filePath, tempDir string, skipSmallImages bool, minImageSize int, noTips bool) ([]PDFImageData, error) {
	fmt.Printf("🔄 Using fallback extraction with pdfimages for CMYK/incompatible PDFs...\n")

	// Check if pdfimages is available using exec.LookPath
	pdfimagesPath, err := exec.LookPath("pdfimages")
	if err != nil {
		// pdfimages not found, show installation tips
		fmt.Printf("⚠️  pdfimages not found - cannot extract images from CMYK PDFs\n")
		if !noTips {
			fmt.Printf("\n💡 Install poppler-utils to extract images from CMYK PDFs:\n")
			fmt.Printf("   macOS:  brew install poppler\n")
			fmt.Printf("   Ubuntu: sudo apt-get install poppler-utils\n\n")
		}
		return []PDFImageData{}, nil
	}

	fmt.Printf("✅ Found pdfimages at: %s\n", pdfimagesPath)

	// Use pdfimages to extract images
	// -j flag: extract as JPEG when possible (preserves quality)
	// -png flag: extract as PNG when JPEG not appropriate
	imagePrefix := filepath.Join(tempDir, "image")

	cmd := exec.Command("pdfimages", "-j", "-png", filePath, imagePrefix)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  pdfimages extraction failed: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", string(output))
		}
		return []PDFImageData{}, nil
	}

	fmt.Printf("✅ pdfimages extraction completed\n")

	// Find extracted image files
	var imageFiles []string
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".ppm" || ext == ".pbm" {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk temp directory: %w", err)
	}

	if len(imageFiles) == 0 {
		fmt.Printf("⚠️  No images found after fallback extraction\n")
		return []PDFImageData{}, nil
	}

	fmt.Printf("📸 Found %d images using fallback method\n", len(imageFiles))

	var imageData []PDFImageData
	for i, imagePath := range imageFiles {
		// Process each extracted image
		data, err := processExtractedImage(imagePath, filePath, i)
		if err != nil {
			fmt.Printf("⚠️  Failed to process image %d: %v\n", i+1, err)
			continue // Skip problematic images
		}

		// Apply size filter if specified
		if skipSmallImages && len(data.ImageData) < minImageSize {
			continue
		}

		imageData = append(imageData, *data)
	}

	fmt.Printf("✅ Processed %d images (after filtering)\n", len(imageData))

	return imageData, nil
}

// parsePageNumberFromFilename extracts the PDF page number from pdfcpu's output filename.
// pdfcpu names extracted images as: {pdfname}_{pageNr}_{imgName}.{ext}
// e.g. "catalogue_001_Im0.jpg" → page 1, "catalogue_012_Im3.png" → page 12.
// Falls back to imageIndex+1 if the filename doesn't match the expected pattern.
func parsePageNumberFromFilename(imagePath string, imageIndex int) int {
	base := filepath.Base(imagePath)
	// Remove extension
	noExt := strings.TrimSuffix(base, filepath.Ext(base))
	// Split on underscore — format is {name}_{page}_{imgname}
	// We want the second-to-last underscore segment (page number)
	parts := strings.Split(noExt, "_")
	// Need at least 3 parts: name, page, imgname
	if len(parts) >= 3 {
		// Page number is the second-to-last segment
		pageStr := parts[len(parts)-2]
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			return pageNum
		}
	}
	// Fallback: use sequential index (old behavior)
	return imageIndex + 1
}

// processExtractedImage processes a single extracted image
func processExtractedImage(imagePath, sourcePDF string, imageIndex int) (*PDFImageData, error) {
	// Read image file
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	// Generate base64 data
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Data)

	// Extract EXIF data
	exifData := extractEXIFData(imageBytes)

	// Generate OCR text (placeholder - would need OCR library)
	ocrText := extractOCRText(imagePath)

	// Generate descriptive filename
	pdfBasename := filepath.Base(sourcePDF)
	pdfName := strings.TrimSuffix(pdfBasename, filepath.Ext(pdfBasename))
	imageExt := strings.ToLower(filepath.Ext(imagePath))
	if imageExt == "" {
		imageExt = ".jpg" // Default extension
	}
	generatedFilename := fmt.Sprintf("%s_image_%d%s", pdfName, imageIndex+1, imageExt)

	// Generate metadata
	metadata := map[string]interface{}{
		"type":         "image",
		"source_pdf":   sourcePDF,
		"image_index":  imageIndex,
		"image_format": strings.ToLower(filepath.Ext(imagePath)),
		"image_size":   len(imageBytes),
		"date_added":   time.Now().Format(time.RFC3339),
		"filename":     generatedFilename,
	}

	// Add EXIF data to metadata
	for key, value := range exifData {
		metadata[key] = value
	}

	// Parse actual page number from pdfcpu's filename (e.g. "catalogue_003_Im0.jpg" → page 3).
	// This is the critical fix for Issue #41: images were previously associated with the wrong
	// text chunks because sequential imageIndex was used instead of the actual PDF page number.
	pageNumber := parsePageNumberFromFilename(imagePath, imageIndex)
	metadata["page_number"] = pageNumber

	return &PDFImageData{
		ID:         uuid.New().String(),
		ImageData:  base64Data,
		Image:      dataURL,
		URL:        fmt.Sprintf("file://%s#page-%d-image-%d", sourcePDF, pageNumber, imageIndex),
		Metadata:   metadata,
		OCRText:    ocrText,
		EXIFData:   exifData,
		Caption:    fmt.Sprintf("Image %d (page %d) from %s", imageIndex+1, pageNumber, filepath.Base(sourcePDF)),
		PageNumber: pageNumber,
		ImageIndex: imageIndex,
		SourcePDF:  sourcePDF,
	}, nil
}

// extractEXIFData extracts EXIF data from image bytes
func extractEXIFData(imageData []byte) map[string]interface{} {
	// This is a placeholder implementation
	// In a real implementation, you would use an EXIF library
	return map[string]interface{}{
		"width":     0,
		"height":    0,
		"format":    "unknown",
		"timestamp": time.Now().Format(time.RFC3339),
	}
}
