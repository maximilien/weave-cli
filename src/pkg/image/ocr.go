package image

import (
	"fmt"
	"strings"
)

// OCRData represents extracted OCR text and confidence
type OCRData struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
	HasText    bool    `json:"has_text"`
}

// OCRExtractor interface for OCR functionality
type OCRExtractor interface {
	ExtractFromBytes(imageData []byte) (*OCRData, error)
	ExtractFromFile(filePath string) (*OCRData, error)
	ExtractWithLanguage(filePath string, language string) (*OCRData, error)
}

// NoOpOCRExtractor provides a no-op implementation when Tesseract is not available
type NoOpOCRExtractor struct{}

// ExtractFromBytes returns an error indicating OCR is not available
func (n *NoOpOCRExtractor) ExtractFromBytes(imageData []byte) (*OCRData, error) {
	return nil, fmt.Errorf("OCR functionality not available: Tesseract not installed")
}

// ExtractFromFile returns an error indicating OCR is not available
func (n *NoOpOCRExtractor) ExtractFromFile(filePath string) (*OCRData, error) {
	return nil, fmt.Errorf("OCR functionality not available: Tesseract not installed")
}

// ExtractWithLanguage returns an error indicating OCR is not available
func (n *NoOpOCRExtractor) ExtractWithLanguage(filePath string, language string) (*OCRData, error) {
	return nil, fmt.Errorf("OCR functionality not available: Tesseract not installed")
}

// GetOCRExtractor returns an OCR extractor instance
// This will return a NoOpOCRExtractor if Tesseract is not available
func GetOCRExtractor() OCRExtractor {
	// Try to create a Tesseract extractor
	// If it fails, return the no-op extractor
	extractor, err := NewTesseractExtractor()
	if err != nil {
		return &NoOpOCRExtractor{}
	}
	return extractor
}

// NewTesseractExtractor creates a new Tesseract OCR extractor
// Returns an error if Tesseract is not available
func NewTesseractExtractor() (OCRExtractor, error) {
	// This will be implemented in a separate file with build tags
	// For now, return an error to indicate Tesseract is not available
	return nil, fmt.Errorf("Tesseract OCR not available in this build")
}

// GetTextSummary returns a short summary of the OCR text
func (o *OCRData) GetTextSummary(maxLength int) string {
	if !o.HasText {
		return ""
	}

	// Replace newlines with spaces
	text := strings.ReplaceAll(o.Text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Collapse multiple spaces
	text = strings.Join(strings.Fields(text), " ")

	if len(text) <= maxLength {
		return text
	}

	return text[:maxLength] + "..."
}

// WordCount returns the number of words in the extracted text
func (o *OCRData) WordCount() int {
	if !o.HasText {
		return 0
	}
	return len(strings.Fields(o.Text))
}

// IsEmpty returns true if no text was extracted
func (o *OCRData) IsEmpty() bool {
	return !o.HasText || o.Text == ""
}