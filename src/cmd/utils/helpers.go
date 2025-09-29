// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TruncateStringByLines truncates text to a maximum number of lines
func TruncateStringByLines(text string, maxLines int) string {
	if maxLines <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}

	// Return first maxLines lines with truncation indicator
	truncated := strings.Join(lines[:maxLines], "\n")
	return truncated + "\n... (truncated)"
}

// TruncateMetadataValue truncates a metadata value to a maximum length
func TruncateMetadataValue(value interface{}, maxLength int) string {
	valueStr := fmt.Sprintf("%v", value)
	if len(valueStr) > maxLength {
		return valueStr[:maxLength] + "..."
	}
	return valueStr
}

// IsImageContent checks if content looks like base64 image data
func IsImageContent(content string) bool {
	// Check if content looks like base64 image data
	if len(content) < 100 {
		return false
	}
	// Simple heuristic: base64 image data is usually very long and contains base64 characters
	base64Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	base64Count := 0
	for _, char := range content[:min(200, len(content))] {
		if strings.ContainsRune(base64Chars, char) {
			base64Count++
		}
	}
	return float64(base64Count)/float64(min(200, len(content))) > 0.8
}

// IsImageDocument checks if metadata indicates this is an image document
func IsImageDocument(metadata map[string]interface{}) bool {
	// Check if metadata indicates this is an image document
	if metadata == nil {
		return false
	}

	// Check for image-related metadata fields
	imageFields := []string{"image", "image_data", "base64_data", "content_type"}
	for _, field := range imageFields {
		if _, exists := metadata[field]; exists {
			return true
		}
	}

	// Check nested metadata
	if metadataField, ok := metadata["metadata"]; ok {
		if metadataStr, ok := metadataField.(string); ok {
			var metadataObj map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err == nil {
				if contentType, ok := metadataObj["content_type"].(string); ok {
					return contentType == "image"
				}
			}
		}
	}

	return false
}

// GetValueType returns a string representation of the Go type
func GetValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case float32, float64:
		return "float"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
