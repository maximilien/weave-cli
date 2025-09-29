package pdf

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// extractPDFImages extracts images from a PDF file
func extractPDFImages(filePath string, skipSmallImages bool, minImageSize int) ([]PDFImageData, error) {
	// Create a temporary directory for image extraction
	tempDir, err := os.MkdirTemp("", "pdf_images_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract images using pdfcpu
	err = api.ExtractImagesFile(filePath, tempDir, nil)
	if err != nil {
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

	// Generate metadata
	metadata := map[string]interface{}{
		"type":         "image",
		"source_pdf":  sourcePDF,
		"image_index":  imageIndex,
		"image_format": strings.ToLower(filepath.Ext(imagePath)),
		"image_size":   len(imageBytes),
		"date_added":   time.Now().Format(time.RFC3339),
	}

	// Add EXIF data to metadata
	for key, value := range exifData {
		metadata[key] = value
	}

	return &PDFImageData{
		ID:         fmt.Sprintf("%s-image-%d", filepath.Base(sourcePDF), imageIndex),
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