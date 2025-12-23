// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stats

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAnalyzeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		docs     []*vectordb.Document
		topN     int
		expected int // expected number of metadata fields
	}{
		{
			name:     "empty documents",
			docs:     []*vectordb.Document{},
			topN:     5,
			expected: 0,
		},
		{
			name: "single document with metadata",
			docs: []*vectordb.Document{
				{
					ID: "doc1",
					Metadata: map[string]interface{}{
						"author": "John Doe",
						"year":   2024,
					},
				},
			},
			topN:     5,
			expected: 2,
		},
		{
			name: "multiple documents with same fields",
			docs: []*vectordb.Document{
				{
					ID: "doc1",
					Metadata: map[string]interface{}{
						"author": "John Doe",
						"year":   2024,
					},
				},
				{
					ID: "doc2",
					Metadata: map[string]interface{}{
						"author": "Jane Smith",
						"year":   2024,
					},
				},
			},
			topN:     5,
			expected: 2,
		},
		{
			name: "documents with different fields",
			docs: []*vectordb.Document{
				{
					ID: "doc1",
					Metadata: map[string]interface{}{
						"author": "John Doe",
					},
				},
				{
					ID: "doc2",
					Metadata: map[string]interface{}{
						"year": 2024,
					},
				},
			},
			topN:     5,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeMetadata(tt.docs, tt.topN)
			if len(result) != tt.expected {
				t.Errorf("analyzeMetadata() returned %d fields, expected %d", len(result), tt.expected)
			}
		})
	}
}

func TestMetadataFieldStats(t *testing.T) {
	docs := []*vectordb.Document{
		{
			ID: "doc1",
			Metadata: map[string]interface{}{
				"category": "tech",
				"author":   "John",
			},
		},
		{
			ID: "doc2",
			Metadata: map[string]interface{}{
				"category": "tech",
				"author":   "Jane",
			},
		},
		{
			ID: "doc3",
			Metadata: map[string]interface{}{
				"category": "science",
				"author":   "John",
			},
		},
	}

	result := analyzeMetadata(docs, 5)

	// Check category field
	if categoryStats, exists := result["category"]; exists {
		if categoryStats.UniqueValues != 2 {
			t.Errorf("category field should have 2 unique values, got %d", categoryStats.UniqueValues)
		}
		if len(categoryStats.TopValues) == 0 {
			t.Error("category field should have top values")
		}
		// "tech" should be the top value (appears 2 times)
		if categoryStats.TopValues[0].Value != "tech" {
			t.Errorf("top category value should be 'tech', got '%s'", categoryStats.TopValues[0].Value)
		}
		if categoryStats.TopValues[0].Count != 2 {
			t.Errorf("top category count should be 2, got %d", categoryStats.TopValues[0].Count)
		}
	} else {
		t.Error("category field not found in results")
	}

	// Check author field
	if authorStats, exists := result["author"]; exists {
		if authorStats.UniqueValues != 2 {
			t.Errorf("author field should have 2 unique values, got %d", authorStats.UniqueValues)
		}
		// "John" should be the top value (appears 2 times)
		if authorStats.TopValues[0].Value != "John" {
			t.Errorf("top author value should be 'John', got '%s'", authorStats.TopValues[0].Value)
		}
	} else {
		t.Error("author field not found in results")
	}
}

func TestTopNValues(t *testing.T) {
	docs := []*vectordb.Document{
		{ID: "doc1", Metadata: map[string]interface{}{"tag": "a"}},
		{ID: "doc2", Metadata: map[string]interface{}{"tag": "b"}},
		{ID: "doc3", Metadata: map[string]interface{}{"tag": "c"}},
		{ID: "doc4", Metadata: map[string]interface{}{"tag": "d"}},
		{ID: "doc5", Metadata: map[string]interface{}{"tag": "e"}},
		{ID: "doc6", Metadata: map[string]interface{}{"tag": "f"}},
	}

	tests := []struct {
		name     string
		topN     int
		expected int
	}{
		{"top 3", 3, 3},
		{"top 5", 5, 5},
		{"top 10", 10, 6}, // only 6 unique values exist
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeMetadata(docs, tt.topN)
			if tagStats, exists := result["tag"]; exists {
				if len(tagStats.TopValues) != tt.expected {
					t.Errorf("expected %d top values, got %d", tt.expected, len(tagStats.TopValues))
				}
			} else {
				t.Error("tag field not found")
			}
		})
	}
}

func TestMetadataTypes(t *testing.T) {
	docs := []*vectordb.Document{
		{
			ID: "doc1",
			Metadata: map[string]interface{}{
				"string_field": "text",
				"int_field":    42,
				"float_field":  3.14,
				"bool_field":   true,
			},
		},
	}

	result := analyzeMetadata(docs, 5)

	// Check that different types are correctly identified
	if stringStats, exists := result["string_field"]; exists {
		if stringStats.Type != "string" {
			t.Errorf("string_field type should be 'string', got '%s'", stringStats.Type)
		}
	}

	if intStats, exists := result["int_field"]; exists {
		if intStats.Type != "int" {
			t.Errorf("int_field type should be 'int', got '%s'", intStats.Type)
		}
	}

	if boolStats, exists := result["bool_field"]; exists {
		if boolStats.Type != "bool" {
			t.Errorf("bool_field type should be 'bool', got '%s'", boolStats.Type)
		}
	}
}
