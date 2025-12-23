// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"gopkg.in/yaml.v3"
)

func TestScanFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"test1.md":   "# Markdown file\n\nSome content.",
		"test2.txt":  "Plain text content.",
		"test3.pdf":  "fake pdf content",
		"test4.json": `{"key": "value"}`,
		"test5.yml":  "key: value",
		"ignore.go":  "package main", // Should be ignored
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Test scanning without pattern
	files, err := scanFiles(tmpDir, "")
	if err != nil {
		t.Fatalf("scanFiles failed: %v", err)
	}

	// Should find 5 files (md, txt, pdf, json, yml) - not .go
	expectedCount := 5
	if len(files) != expectedCount {
		t.Errorf("expected %d files, got %d", expectedCount, len(files))
	}

	// Verify correct extensions were found
	foundExtensions := make(map[string]bool)
	for _, file := range files {
		ext := filepath.Ext(file)
		foundExtensions[ext] = true
	}

	requiredExts := []string{".md", ".txt", ".pdf", ".json", ".yml"}
	for _, ext := range requiredExts {
		if !foundExtensions[ext] {
			t.Errorf("expected to find file with extension %s", ext)
		}
	}

	// .go files should not be included
	if foundExtensions[".go"] {
		t.Error("should not include .go files")
	}
}

func TestScanFiles_WithPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"test1.md", "test2.md", "test3.txt"}
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Test scanning with pattern for only .md files
	files, err := scanFiles(tmpDir, "*.md")
	if err != nil {
		t.Fatalf("scanFiles with pattern failed: %v", err)
	}

	// Should find only 2 .md files
	if len(files) != 2 {
		t.Errorf("expected 2 .md files, got %d", len(files))
	}

	for _, file := range files {
		if filepath.Ext(file) != ".md" {
			t.Errorf("expected only .md files, got %s", file)
		}
	}
}

func TestSaveSchemaToYAML(t *testing.T) {
	schema := agents.SchemaConfig{
		CollectionName:   "test-collection",
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Fields: []agents.FieldConfig{
			{
				Name:        "content",
				Type:        "text",
				Description: "Main content field",
				Indexed:     true,
				Filterable:  false,
				Required:    true,
			},
			{
				Name:        "timestamp",
				Type:        "datetime",
				Description: "Creation timestamp",
				Indexed:     true,
				Filterable:  true,
				Required:    false,
			},
		},
		Indexes: []agents.IndexConfig{
			{
				Name:   "content_vector_index",
				Fields: []string{"content"},
				Type:   "vector",
			},
		},
		Metadata: map[string]interface{}{
			"version": "1.0",
			"created": "2025-12-23",
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test-schema.yaml")

	// Save schema to file
	err := saveSchemaToYAML(schema, outputFile)
	if err != nil {
		t.Fatalf("saveSchemaToYAML failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("schema file was not created")
	}

	// Read and parse the file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	// Verify it contains expected content
	contentStr := string(content)
	if !containsSubstring(contentStr, "collection_name: test-collection") {
		t.Error("schema file should contain collection name")
	}
	if !containsSubstring(contentStr, "vector_dimensions: 1536") {
		t.Error("schema file should contain vector dimensions")
	}
	if !containsSubstring(contentStr, "similarity_metric: cosine") {
		t.Error("schema file should contain similarity metric")
	}

	// Verify it's valid YAML
	var parsedSchema agents.SchemaConfig
	if err := yaml.Unmarshal(content, &parsedSchema); err != nil {
		t.Fatalf("saved schema is not valid YAML: %v", err)
	}

	// Verify parsed values match original
	if parsedSchema.CollectionName != schema.CollectionName {
		t.Errorf("collection name mismatch: expected %s, got %s",
			schema.CollectionName, parsedSchema.CollectionName)
	}
	if parsedSchema.VectorDimensions != schema.VectorDimensions {
		t.Errorf("vector dimensions mismatch: expected %d, got %d",
			schema.VectorDimensions, parsedSchema.VectorDimensions)
	}
}

func TestSchemaAnalysisOutput_JSONSerialization(t *testing.T) {
	// Test that SchemaAnalysisOutput with ChunkingRecommendation serializes correctly
	output := agents.SchemaAnalysisOutput{
		Schema: agents.SchemaConfig{
			CollectionName:   "test",
			VectorDimensions: 1536,
			SimilarityMetric: "cosine",
			Fields:           []agents.FieldConfig{},
			Indexes:          []agents.IndexConfig{},
		},
		Reasoning:     "Test reasoning",
		FieldAnalysis: []agents.FieldSuggestion{},
		Confidence:    0.85,
		Warnings:      []string{"Test warning"},
		ChunkingAdvice: &agents.ChunkingRecommendation{
			RecommendedSize: 1000,
			MinSize:         500,
			MaxSize:         1500,
			OverlapSize:     100,
			Reasoning:       "Based on content analysis",
			DocumentType:    "technical",
			Considerations:  []string{"Preserve code blocks", "Maintain context"},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal to JSON: %v", err)
	}

	// Unmarshal back
	var parsed agents.SchemaAnalysisOutput
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify chunking advice was preserved
	if parsed.ChunkingAdvice == nil {
		t.Fatal("chunking advice should not be nil after unmarshaling")
	}

	if parsed.ChunkingAdvice.RecommendedSize != 1000 {
		t.Errorf("expected recommended size 1000, got %d", parsed.ChunkingAdvice.RecommendedSize)
	}

	if parsed.ChunkingAdvice.DocumentType != "technical" {
		t.Errorf("expected document type 'technical', got %s", parsed.ChunkingAdvice.DocumentType)
	}

	if len(parsed.ChunkingAdvice.Considerations) != 2 {
		t.Errorf("expected 2 considerations, got %d", len(parsed.ChunkingAdvice.Considerations))
	}
}

func TestSchemaAnalysisOutput_YAMLSerialization(t *testing.T) {
	// Test YAML serialization with ChunkingRecommendation
	output := agents.SchemaAnalysisOutput{
		Schema: agents.SchemaConfig{
			CollectionName:   "test",
			VectorDimensions: 1536,
			SimilarityMetric: "cosine",
			Fields:           []agents.FieldConfig{},
			Indexes:          []agents.IndexConfig{},
		},
		ChunkingAdvice: &agents.ChunkingRecommendation{
			RecommendedSize: 800,
			MinSize:         400,
			MaxSize:         1200,
			OverlapSize:     80,
			Reasoning:       "Medium content density",
			DocumentType:    "mixed",
			Considerations:  []string{"Natural boundaries"},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal to YAML: %v", err)
	}

	// Unmarshal back
	var parsed agents.SchemaAnalysisOutput
	if err := yaml.Unmarshal(yamlData, &parsed); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Verify chunking advice
	if parsed.ChunkingAdvice == nil {
		t.Fatal("chunking advice should not be nil after unmarshaling YAML")
	}

	if parsed.ChunkingAdvice.RecommendedSize != 800 {
		t.Errorf("expected recommended size 800, got %d", parsed.ChunkingAdvice.RecommendedSize)
	}
}

func TestChunkingRecommendation_ValidationRanges(t *testing.T) {
	tests := []struct {
		name        string
		rec         agents.ChunkingRecommendation
		shouldValid bool
	}{
		{
			name: "valid technical",
			rec: agents.ChunkingRecommendation{
				RecommendedSize: 800,
				MinSize:         500,
				MaxSize:         1200,
				OverlapSize:     80,
				DocumentType:    "technical",
			},
			shouldValid: true,
		},
		{
			name: "valid narrative",
			rec: agents.ChunkingRecommendation{
				RecommendedSize: 1500,
				MinSize:         1000,
				MaxSize:         2000,
				OverlapSize:     150,
				DocumentType:    "narrative",
			},
			shouldValid: true,
		},
		{
			name: "recommended in range",
			rec: agents.ChunkingRecommendation{
				RecommendedSize: 1000,
				MinSize:         500,
				MaxSize:         1500,
				OverlapSize:     100,
			},
			shouldValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify recommended size is within min/max range
			if tt.rec.RecommendedSize < tt.rec.MinSize {
				t.Errorf("recommended size %d should be >= min size %d",
					tt.rec.RecommendedSize, tt.rec.MinSize)
			}
			if tt.rec.RecommendedSize > tt.rec.MaxSize {
				t.Errorf("recommended size %d should be <= max size %d",
					tt.rec.RecommendedSize, tt.rec.MaxSize)
			}

			// Verify overlap is reasonable (typically 10-20%)
			overlapPct := float64(tt.rec.OverlapSize) / float64(tt.rec.RecommendedSize) * 100
			if overlapPct < 5 || overlapPct > 30 {
				t.Logf("overlap percentage %.1f%% is outside typical range (10-20%%)", overlapPct)
			}
		})
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || containsSubstring(s[1:], substr))))
}
