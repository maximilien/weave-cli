// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"strings"
	"testing"
)

func TestTruncateStringByLines(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLines int
		expected string
	}{
		{
			name:     "Empty string",
			text:     "",
			maxLines: 5,
			expected: "",
		},
		{
			name:     "Single line within limit",
			text:     "Hello world",
			maxLines: 5,
			expected: "Hello world",
		},
		{
			name:     "Multiple lines within limit",
			text:     "Line 1\nLine 2\nLine 3",
			maxLines: 5,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Multiple lines exceeding limit",
			text:     "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6",
			maxLines: 3,
			expected: "Line 1\nLine 2\nLine 3\n... (truncated)",
		},
		{
			name:     "Exactly at limit",
			text:     "Line 1\nLine 2\nLine 3",
			maxLines: 3,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Zero maxLines returns original",
			text:     "Line 1\nLine 2\nLine 3",
			maxLines: 0,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Negative maxLines returns original",
			text:     "Line 1\nLine 2\nLine 3",
			maxLines: -1,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Single line truncation",
			text:     "Line 1\nLine 2",
			maxLines: 1,
			expected: "Line 1\n... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateStringByLines(tt.text, tt.maxLines)
			if result != tt.expected {
				t.Errorf("TruncateStringByLines() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestTruncateMetadataValue(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		maxLength int
		expected  string
	}{
		{
			name:      "Short string within limit",
			value:     "hello",
			maxLength: 10,
			expected:  "hello",
		},
		{
			name:      "String exceeding limit",
			value:     "This is a very long string",
			maxLength: 10,
			expected:  "This is a ...",
		},
		{
			name:      "Exactly at limit",
			value:     "1234567890",
			maxLength: 10,
			expected:  "1234567890",
		},
		{
			name:      "Integer value",
			value:     42,
			maxLength: 10,
			expected:  "42",
		},
		{
			name:      "Float value",
			value:     3.14159,
			maxLength: 5,
			expected:  "3.141...",
		},
		{
			name:      "Boolean value",
			value:     true,
			maxLength: 10,
			expected:  "true",
		},
		{
			name:      "Slice value truncated",
			value:     []int{1, 2, 3, 4, 5},
			maxLength: 10,
			expected:  "[1 2 3 4 5...",
		},
		{
			name:      "Map value truncated",
			value:     map[string]int{"a": 1, "b": 2},
			maxLength: 15,
			expected:  "map[a:1 b:2]",
		},
		{
			name:      "Empty string",
			value:     "",
			maxLength: 10,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateMetadataValue(tt.value, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateMetadataValue() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsImageContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "Short string (< 100 chars)",
			content:  "Hello world",
			expected: false,
		},
		{
			name:     "Long base64 string (> 80% base64 chars)",
			content:  strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=", 10),
			expected: true,
		},
		{
			name:     "Long non-base64 string",
			content:  strings.Repeat("🎉 Special chars: @#$%^&*(){}[]|\\<>? Non-ASCII: 你好世界! ", 10),
			expected: false,
		},
		{
			name:     "Mixed content mostly base64",
			content:  "data:image/jpeg;base64," + strings.Repeat("iVBORw0KGgoAAAANSUhEUgAA", 20),
			expected: true,
		},
		{
			name:     "Exactly 100 chars base64",
			content:  strings.Repeat("A", 100),
			expected: true,
		},
		{
			name:     "Long text with < 80% base64 chars",
			content:  strings.Repeat("@#$ Special!!! 你好 ", 20),
			expected: false,
		},
		{
			name:     "Typical base64 image prefix",
			content:  "data:image/png;base64," + strings.Repeat("iVBORw0KGgoAAAANSUhEUgAAAAUA", 10),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsImageContent(tt.content)
			if result != tt.expected {
				t.Errorf("IsImageContent() = %v, expected %v (content length: %d)", result, tt.expected, len(tt.content))
			}
		})
	}
}

func TestIsImageDocument(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		expected bool
	}{
		{
			name:     "Nil metadata",
			metadata: nil,
			expected: false,
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
			expected: false,
		},
		{
			name: "Has 'image' field",
			metadata: map[string]interface{}{
				"image": "some_image_data",
			},
			expected: true,
		},
		{
			name: "Has 'image_data' field",
			metadata: map[string]interface{}{
				"image_data": "base64data",
			},
			expected: true,
		},
		{
			name: "Has 'base64_data' field",
			metadata: map[string]interface{}{
				"base64_data": "data",
			},
			expected: true,
		},
		{
			name: "Has 'content_type' field",
			metadata: map[string]interface{}{
				"content_type": "image/jpeg",
			},
			expected: true,
		},
		{
			name: "Nested metadata with content_type=image",
			metadata: map[string]interface{}{
				"metadata": `{"content_type":"image"}`,
			},
			expected: true,
		},
		{
			name: "Nested metadata with content_type=text",
			metadata: map[string]interface{}{
				"metadata": `{"content_type":"text"}`,
			},
			expected: false,
		},
		{
			name: "Nested metadata with invalid JSON",
			metadata: map[string]interface{}{
				"metadata": "not valid json",
			},
			expected: false,
		},
		{
			name: "Nested metadata not a string",
			metadata: map[string]interface{}{
				"metadata": 123,
			},
			expected: false,
		},
		{
			name: "Regular text document",
			metadata: map[string]interface{}{
				"title":  "Document Title",
				"author": "John Doe",
			},
			expected: false,
		},
		{
			name: "Multiple image indicators",
			metadata: map[string]interface{}{
				"image":        "data",
				"image_data":   "more_data",
				"content_type": "image/png",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsImageDocument(tt.metadata)
			if result != tt.expected {
				t.Errorf("IsImageDocument() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetValueType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "String value",
			value:    "hello",
			expected: "string",
		},
		{
			name:     "Int value",
			value:    42,
			expected: "integer",
		},
		{
			name:     "Int8 value",
			value:    int8(8),
			expected: "integer",
		},
		{
			name:     "Int16 value",
			value:    int16(16),
			expected: "integer",
		},
		{
			name:     "Int32 value",
			value:    int32(32),
			expected: "integer",
		},
		{
			name:     "Int64 value",
			value:    int64(64),
			expected: "integer",
		},
		{
			name:     "Float32 value",
			value:    float32(3.14),
			expected: "float",
		},
		{
			name:     "Float64 value",
			value:    3.14159,
			expected: "float",
		},
		{
			name:     "Boolean true",
			value:    true,
			expected: "boolean",
		},
		{
			name:     "Boolean false",
			value:    false,
			expected: "boolean",
		},
		{
			name:     "Array/slice",
			value:    []interface{}{1, 2, 3},
			expected: "array",
		},
		{
			name:     "Map/object",
			value:    map[string]interface{}{"key": "value"},
			expected: "object",
		},
		{
			name:     "Nil value",
			value:    nil,
			expected: "unknown",
		},
		{
			name:     "Struct value",
			value:    struct{ Name string }{Name: "test"},
			expected: "unknown",
		},
		{
			name:     "Empty slice",
			value:    []interface{}{},
			expected: "array",
		},
		{
			name:     "Empty map",
			value:    map[string]interface{}{},
			expected: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValueType(tt.value)
			if result != tt.expected {
				t.Errorf("GetValueType(%v) = %q, expected %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a < b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "a > b",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "a == b",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "Negative numbers",
			a:        -5,
			b:        -10,
			expected: -10,
		},
		{
			name:     "Negative and positive",
			a:        -5,
			b:        10,
			expected: -5,
		},
		{
			name:     "Zero values",
			a:        0,
			b:        0,
			expected: 0,
		},
		{
			name:     "Zero and positive",
			a:        0,
			b:        5,
			expected: 0,
		},
		{
			name:     "Zero and negative",
			a:        0,
			b:        -5,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Integration tests for realistic scenarios
func TestHelpers_RealWorldScenarios(t *testing.T) {
	t.Run("Image detection for PDF extracted images", func(t *testing.T) {
		// Simulate base64 image from PDF extraction (AuctionsMax.ai use case)
		base64Image := "data:image/jpeg;base64," + strings.Repeat("iVBORw0KGgoAAAANSUhEUgAA", 100)

		if !IsImageContent(base64Image) {
			t.Error("Expected to detect base64 image content from PDF extraction")
		}
	})

	t.Run("Metadata detection for image documents", func(t *testing.T) {
		// Typical metadata from image extraction
		metadata := map[string]interface{}{
			"content_type": "image",
			"source":       "2022-tamarkin-auction-catalogue.pdf",
			"page":         1,
			"image_index":  1,
		}

		if !IsImageDocument(metadata) {
			t.Error("Expected to detect image document from metadata")
		}
	})

	t.Run("Truncate long error messages for display", func(t *testing.T) {
		longError := strings.Repeat("Error detail line\n", 20)
		truncated := TruncateStringByLines(longError, 5)

		lines := strings.Split(truncated, "\n")
		if len(lines) > 6 { // 5 lines + truncation indicator
			t.Errorf("Expected at most 6 lines, got %d", len(lines))
		}
		if !strings.Contains(truncated, "... (truncated)") {
			t.Error("Expected truncation indicator")
		}
	})

	t.Run("Type detection for schema validation", func(t *testing.T) {
		values := []interface{}{
			"text",
			42,
			3.14,
			true,
			[]interface{}{1, 2, 3},
			map[string]interface{}{"key": "value"},
		}

		expectedTypes := []string{"string", "integer", "float", "boolean", "array", "object"}

		for i, value := range values {
			valueType := GetValueType(value)
			if valueType != expectedTypes[i] {
				t.Errorf("Expected type %q for value %v, got %q", expectedTypes[i], value, valueType)
			}
		}
	})
}
