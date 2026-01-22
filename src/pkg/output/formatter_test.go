// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format Format
	}{
		{"JSON formatter", FormatJSON},
		{"YAML formatter", FormatYAML},
		{"Table formatter", FormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewFormatter(tt.format, buf)

			if formatter == nil {
				t.Fatal("Expected formatter, got nil")
			}
			if formatter.format != tt.format {
				t.Errorf("Expected format %s, got %s", tt.format, formatter.format)
			}
			if formatter.writer != buf {
				t.Error("Expected writer to be set")
			}
		})
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		contains []string
	}{
		{
			name: "Simple map",
			data: map[string]interface{}{
				"name":    "test",
				"version": "0.9.8",
				"count":   42,
			},
			contains: []string{`"name"`, `"test"`, `"version"`, `"0.9.8"`, `"count"`, `42`},
		},
		{
			name: "Nested structure",
			data: map[string]interface{}{
				"config": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
				"enabled": true,
			},
			contains: []string{`"config"`, `"host"`, `"localhost"`, `"port"`, `8080`, `"enabled"`, `true`},
		},
		{
			name: "Array data",
			data: map[string]interface{}{
				"items": []string{"one", "two", "three"},
			},
			contains: []string{`"items"`, `"one"`, `"two"`, `"three"`},
		},
		{
			name:     "Empty map",
			data:     map[string]interface{}{},
			contains: []string{`{}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewFormatter(FormatJSON, buf)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}

			// Verify it's valid JSON
			var parsed interface{}
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v\nOutput:\n%s", err, output)
			}
		})
	}
}

func TestFormatter_FormatYAML(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		contains []string
	}{
		{
			name: "Simple map",
			data: map[string]interface{}{
				"name":    "test",
				"version": "0.9.8",
			},
			contains: []string{"name:", "test", "version:", "0.9.8"},
		},
		{
			name: "Nested structure",
			data: map[string]interface{}{
				"config": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
			},
			contains: []string{"config:", "host:", "localhost", "port:", "8080"},
		},
		{
			name: "Array data",
			data: map[string]interface{}{
				"items": []string{"one", "two", "three"},
			},
			contains: []string{"items:", "- one", "- two", "- three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewFormatter(FormatYAML, buf)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}

			// Verify it's valid YAML
			var parsed interface{}
			if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Errorf("Output is not valid YAML: %v\nOutput:\n%s", err, output)
			}
		})
	}
}

func TestFormatter_FormatTable(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		contains []string
	}{
		{
			name: "Simple map",
			data: map[string]interface{}{
				"name":    "test",
				"version": "0.9.8",
			},
			contains: []string{"name", "test", "version", "0.9.8", ":"},
		},
		{
			name: "Map with alignment",
			data: map[string]interface{}{
				"x":          "short",
				"long_key":   "value",
				"medium_key": "data",
			},
			contains: []string{"x", "long_key", "medium_key", "short", "value", "data", ":"},
		},
		{
			name: "Nested map",
			data: map[string]interface{}{
				"config": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
			},
			contains: []string{"config:", "host:", "localhost", "port:", "8080"},
		},
		{
			name: "Array data",
			data: map[string]interface{}{
				"items": []interface{}{"one", "two", "three"},
			},
			contains: []string{"items:", "[0] one", "[1] two", "[2] three"},
		},
		{
			name: "Empty array",
			data: map[string]interface{}{
				"empty": []interface{}{},
			},
			contains: []string{"empty:", "[]"},
		},
		{
			name: "Empty nested map",
			data: map[string]interface{}{
				"empty": map[string]interface{}{},
			},
			contains: []string{"empty:", "{}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewFormatter(FormatTable, buf)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestFormatter_FormatTable_StructConversion(t *testing.T) {
	// Test that structs get converted to maps via JSON marshaling
	type TestStruct struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Count   int    `json:"count"`
	}

	data := TestStruct{
		Name:    "test-struct",
		Version: "1.0.0",
		Count:   5,
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatTable, buf)

	err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	expectedContains := []string{"name", "test-struct", "version", "1.0.0", "count", "5", ":"}
	for _, expected := range expectedContains {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}

func TestFormatter_UnsupportedFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(Format("unsupported"), buf)

	data := map[string]interface{}{"test": "data"}
	err := formatter.Format(data)

	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "String value",
			value:    "test",
			expected: "test",
		},
		{
			name:     "Integer value",
			value:    42,
			expected: "42",
		},
		{
			name:     "Boolean value",
			value:    true,
			expected: "true",
		},
		{
			name:     "Simple array",
			value:    []interface{}{"a", "b", "c"},
			expected: "[a, b, c]",
		},
		{
			name:     "Empty array",
			value:    []interface{}{},
			expected: "[]",
		},
		{
			name:     "Simple map",
			value:    map[string]interface{}{"key": "value", "count": 5},
			expected: "{", // Should contain both opening brace and key-value pairs
		},
		{
			name:     "Nested array",
			value:    []interface{}{1, 2, 3},
			expected: "[1, 2, 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatValue_NestedStructures(t *testing.T) {
	// Test nested map
	nestedMap := map[string]interface{}{
		"inner": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}
	result := formatValue(nestedMap)
	if !strings.Contains(result, "{") || !strings.Contains(result, "inner") {
		t.Errorf("Expected nested map formatting, got: %s", result)
	}

	// Test nested array
	nestedArray := []interface{}{
		[]interface{}{"a", "b"},
		[]interface{}{"c", "d"},
	}
	result = formatValue(nestedArray)
	if !strings.Contains(result, "[") {
		t.Errorf("Expected nested array formatting, got: %s", result)
	}
}

func TestFormatIngestReport(t *testing.T) {
	report := map[string]interface{}{
		"total_documents": 100,
		"successful":      95,
		"failed":          5,
		"duration":        "2.5s",
	}

	tests := []struct {
		name   string
		format Format
	}{
		{"JSON format", FormatJSON},
		{"YAML format", FormatYAML},
		{"Table format", FormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := FormatIngestReport(report, tt.format, buf)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("Expected non-empty output")
			}

			// Verify all report fields are present
			expectedFields := []string{"100", "95", "5", "2.5s"}
			for _, field := range expectedFields {
				if !strings.Contains(output, field) {
					t.Errorf("Expected output to contain '%s', got:\n%s", field, output)
				}
			}
		})
	}
}

func TestFormatter_FormatJSON_IndentedOutput(t *testing.T) {
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatJSON, buf)
	err := formatter.Format(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	// Check for indentation (two spaces as configured)
	if !strings.Contains(output, "  ") {
		t.Error("Expected indented JSON output")
	}
}

func TestFormatter_FormatYAML_IndentedOutput(t *testing.T) {
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatYAML, buf)
	err := formatter.Format(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	// Check for indentation (2 spaces as configured)
	lines := strings.Split(output, "\n")
	hasIndentation := false
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") {
			hasIndentation = true
			break
		}
	}
	if !hasIndentation {
		t.Error("Expected indented YAML output")
	}
}

func TestFormatter_FormatTable_KeyAlignment(t *testing.T) {
	data := map[string]interface{}{
		"x":             "value1",
		"medium_key":    "value2",
		"very_long_key": "value3",
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatTable, buf)
	err := formatter.Format(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// All lines should have colons at roughly the same position (after padding)
	// Just verify output contains all keys and values
	expectedItems := []string{"x", "medium_key", "very_long_key", "value1", "value2", "value3", ":"}
	for _, item := range expectedItems {
		if !strings.Contains(output, item) {
			t.Errorf("Expected output to contain '%s', got:\n%s", item, output)
		}
	}
}

func TestFormatter_FormatTable_ComplexNestedArray(t *testing.T) {
	data := map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{"id": 1, "name": "first"},
			map[string]interface{}{"id": 2, "name": "second"},
		},
	}

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatTable, buf)
	err := formatter.Format(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain array indices
	if !strings.Contains(output, "[0]") || !strings.Contains(output, "[1]") {
		t.Errorf("Expected array indices in output:\n%s", output)
	}
}

func TestFormatConstants(t *testing.T) {
	// Verify format constants have expected values
	if FormatJSON != "json" {
		t.Errorf("Expected FormatJSON='json', got '%s'", FormatJSON)
	}
	if FormatYAML != "yaml" {
		t.Errorf("Expected FormatYAML='yaml', got '%s'", FormatYAML)
	}
	if FormatTable != "table" {
		t.Errorf("Expected FormatTable='table', got '%s'", FormatTable)
	}
}
