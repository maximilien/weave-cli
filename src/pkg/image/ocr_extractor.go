package image

import (
	"fmt"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// OCRData represents extracted OCR text and confidence
type OCRData struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
	HasText    bool    `json:"has_text"`
}

// ExtractOCR extracts text from an image using Tesseract OCR
func ExtractOCR(filePath string) (*OCRData, error) {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetImage(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to set image: %w", err)
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Clean up text
	text = strings.TrimSpace(text)
	hasText := len(text) > 0

	// Confidence is not easily available in gosseract v2, default to 0
	return &OCRData{
		Text:       text,
		Confidence: 0.0,
		Language:   "eng", // Default language
		HasText:    hasText,
	}, nil
}

// ExtractOCRFromBytes extracts text from image bytes using Tesseract OCR
func ExtractOCRFromBytes(imageData []byte) (*OCRData, error) {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetImageFromBytes(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to set image from bytes: %w", err)
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Clean up text
	text = strings.TrimSpace(text)
	hasText := len(text) > 0

	// Confidence is not easily available in gosseract v2, default to 0
	return &OCRData{
		Text:       text,
		Confidence: 0.0,
		Language:   "eng", // Default language
		HasText:    hasText,
	}, nil
}

// ExtractOCRWithLanguage extracts text with specific language
func ExtractOCRWithLanguage(filePath string, language string) (*OCRData, error) {
	client := gosseract.NewClient()
	defer client.Close()

	err := client.SetLanguage(language)
	if err != nil {
		return nil, fmt.Errorf("failed to set language: %w", err)
	}

	err = client.SetImage(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to set image: %w", err)
	}

	// Extract text
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Clean up text
	text = strings.TrimSpace(text)
	hasText := len(text) > 0

	// Confidence is not easily available in gosseract v2, default to 0
	return &OCRData{
		Text:       text,
		Confidence: 0.0,
		Language:   language,
		HasText:    hasText,
	}, nil
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
