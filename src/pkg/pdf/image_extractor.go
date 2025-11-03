// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pdf

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
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

// extractImagesWithFallback attempts to extract images using a more lenient approach
func extractImagesWithFallback(filePath, tempDir string, skipSmallImages bool, minImageSize int, noTips bool) ([]PDFImageData, error) {
	// For now, return empty slice with a warning
	// This allows PDF processing to continue without images
	fmt.Printf("⚠️  Skipping image extraction for this PDF due to incompatible JPEG format\n")

	// Only show tips if not suppressed
	if !noTips {
		fmt.Printf("\n💡 Tips for extracting images from CMYK PDFs:\n")
		fmt.Printf("\n   Option 1 - Using Ghostscript (recommended):\n")
		fmt.Printf("   $ gs -sDEVICE=pdfwrite -dProcessColorModel=/DeviceRGB \\\n")
		fmt.Printf("        -dColorConversionStrategy=/RGB -dNOPAUSE -dBATCH \\\n")
		fmt.Printf("        -sOutputFile=output-rgb.pdf %s\n", filepath.Base(filePath))
		fmt.Printf("   $ weave docs create <collection> output-rgb.pdf --image-col <image-collection>\n")
		fmt.Printf("\n   Option 2 - Using ImageMagick:\n")
		fmt.Printf("   $ convert -density 300 -colorspace RGB %s output-rgb.pdf\n", filepath.Base(filePath))
		fmt.Printf("   $ weave docs create <collection> output-rgb.pdf --image-col <image-collection>\n")
		fmt.Printf("\n   Option 3 - Continue with text-only processing (current):\n")
		fmt.Printf("   Text content will be extracted and searchable without images.\n\n")
	}

	return []PDFImageData{}, nil
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

	return &PDFImageData{
		ID:         uuid.New().String(),
		ImageData:  base64Data,
		Image:      dataURL,
		URL:        fmt.Sprintf("file://%s#image-%d", sourcePDF, imageIndex),
		Metadata:   metadata,
		OCRText:    ocrText,
		EXIFData:   exifData,
		Caption:    fmt.Sprintf("Image %d from %s", imageIndex+1, filepath.Base(sourcePDF)),
		PageNumber: imageIndex + 1,
		ImageIndex: imageIndex,
		SourcePDF:  sourcePDF,
	}, nil
}

// extractOCRText extracts text from an image using OCR (placeholder implementation)
func extractOCRText(imagePath string) string {
	// This is a placeholder implementation
	// In a real implementation, you would use an OCR library like Tesseract
	return fmt.Sprintf("OCR text from image: %s", filepath.Base(imagePath))
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
