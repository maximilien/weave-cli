// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaAgentDocumentSampleBoundaries(t *testing.T) {
	agent := &SchemaAgent{}
	if agent.Name() != "schema-agent" {
		t.Fatalf("Name() = %q", agent.Name())
	}
	if _, err := agent.Execute(context.Background(), "invalid"); err == nil {
		t.Fatal("expected invalid input type error")
	}

	root := t.TempDir()
	jsonPath := filepath.Join(root, "sample.json")
	jsonData := `{"title":"fixture","count":2}`
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0o600); err != nil {
		t.Fatal(err)
	}
	jsonSample := agent.extractDocumentSample(jsonPath)
	if jsonSample == nil || jsonSample.Type != "json" || jsonSample.Fields["title"] != "fixture" {
		t.Fatalf("unexpected JSON sample: %#v", jsonSample)
	}

	textPath := filepath.Join(root, "sample.txt")
	longText := strings.Repeat("x", 1100)
	if err := os.WriteFile(textPath, []byte(longText), 0o600); err != nil {
		t.Fatal(err)
	}
	textSample := agent.extractDocumentSample(textPath)
	if textSample == nil || textSample.Type != "txt" || len(textSample.Preview) != 1000 {
		t.Fatalf("unexpected text sample: %#v", textSample)
	}

	markdownPath := filepath.Join(root, "sample.md")
	if err := os.WriteFile(markdownPath, []byte("# Heading"), 0o600); err != nil {
		t.Fatal(err)
	}
	samples := agent.extractSamples([]string{
		filepath.Join(root, "missing.txt"), jsonPath, markdownPath, textPath,
	}, 2)
	if len(samples) != 2 || samples[1].Type != "md" {
		t.Fatalf("unexpected samples: %#v", samples)
	}

	badJSON := filepath.Join(root, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{badJSON, filepath.Join(root, "unknown.bin"), filepath.Join(root, "missing.pdf")} {
		if sample := agent.extractDocumentSample(path); sample != nil {
			t.Fatalf("expected no sample for %s: %#v", path, sample)
		}
	}
}

func TestAnalyzeStructure_ChunkingMetrics(t *testing.T) {
	agent := &SchemaAgent{}

	tests := []struct {
		name     string
		samples  []DocumentSample
		expected struct {
			paragraphCount int
			sectionCount   int
			contentDensity string
		}
	}{
		{
			name: "simple markdown with paragraphs",
			samples: []DocumentSample{
				{
					Path: "test.md",
					Type: "md",
					Size: 500,
					Fields: map[string]interface{}{
						"content": "# Title\n\nParagraph 1\n\nParagraph 2\n\nParagraph 3",
					},
					Preview: "# Title\n\nParagraph 1\n\nParagraph 2\n\nParagraph 3",
				},
			},
			expected: struct {
				paragraphCount int
				sectionCount   int
				contentDensity string
			}{
				paragraphCount: 4, // Title + 3 paragraphs
				sectionCount:   1, // 1 header (#)
				contentDensity: "sparse",
			},
		},
		{
			name: "dense content",
			samples: []DocumentSample{
				{
					Path: "test.md",
					Type: "md",
					Size: 6000,
					Fields: map[string]interface{}{
						"content": "Long content" + string(make([]byte, 5900)),
					},
					Preview: "Long content",
				},
			},
			expected: struct {
				paragraphCount int
				sectionCount   int
				contentDensity string
			}{
				paragraphCount: 1,
				sectionCount:   0,
				contentDensity: "dense",
			},
		},
		{
			name: "medium content",
			samples: []DocumentSample{
				{
					Path: "test.txt",
					Type: "txt",
					Size: 2000,
					Fields: map[string]interface{}{
						"content": "Medium length content",
					},
					Preview: "Medium length content",
				},
			},
			expected: struct {
				paragraphCount int
				sectionCount   int
				contentDensity string
			}{
				paragraphCount: 1,
				sectionCount:   0,
				contentDensity: "medium",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structure := agent.analyzeStructure(tt.samples)

			// Check paragraph count
			if len(structure.ParagraphCounts) != len(tt.samples) {
				t.Errorf("expected %d paragraph counts, got %d", len(tt.samples), len(structure.ParagraphCounts))
			}
			if len(structure.ParagraphCounts) > 0 && structure.ParagraphCounts[0] != tt.expected.paragraphCount {
				t.Errorf("expected %d paragraphs, got %d", tt.expected.paragraphCount, structure.ParagraphCounts[0])
			}

			// Check section count
			if len(structure.SectionCounts) != len(tt.samples) {
				t.Errorf("expected %d section counts, got %d", len(tt.samples), len(structure.SectionCounts))
			}
			if len(structure.SectionCounts) > 0 && structure.SectionCounts[0] != tt.expected.sectionCount {
				t.Errorf("expected %d sections, got %d", tt.expected.sectionCount, structure.SectionCounts[0])
			}

			// Check content density
			if structure.ContentDensity != tt.expected.contentDensity {
				t.Errorf("expected content density %s, got %s", tt.expected.contentDensity, structure.ContentDensity)
			}
		})
	}
}

func TestAnalyzeStructure_AverageParagraphLength(t *testing.T) {
	agent := &SchemaAgent{}

	samples := []DocumentSample{
		{
			Path: "test.md",
			Type: "md",
			Size: 100,
			Fields: map[string]interface{}{
				"content": "Short paragraph.\n\nAnother short one.",
			},
			Preview: "Short paragraph.\n\nAnother short one.",
		},
	}

	structure := agent.analyzeStructure(samples)

	if structure.AvgParagraphLen == 0 {
		t.Error("expected non-zero average paragraph length")
	}

	// Should have some reasonable average (both paragraphs are ~15-20 chars)
	if structure.AvgParagraphLen < 10 || structure.AvgParagraphLen > 30 {
		t.Errorf("expected average paragraph length between 10-30, got %d", structure.AvgParagraphLen)
	}
}

func TestAnalyzeStructure_MultipleFileTypes(t *testing.T) {
	agent := &SchemaAgent{}

	samples := []DocumentSample{
		{
			Path: "test1.md",
			Type: "md",
			Size: 100,
			Fields: map[string]interface{}{
				"content": "Markdown content",
			},
			Preview: "Markdown content",
		},
		{
			Path: "test2.pdf",
			Type: "pdf",
			Size: 200,
			Fields: map[string]interface{}{
				"content": "PDF content",
			},
			Preview: "PDF content",
		},
		{
			Path: "test3.txt",
			Type: "txt",
			Size: 150,
			Fields: map[string]interface{}{
				"content": "Text content",
			},
			Preview: "Text content",
		},
	}

	structure := agent.analyzeStructure(samples)

	// Should have 3 file types
	if len(structure.FileTypes) != 3 {
		t.Errorf("expected 3 file types, got %d", len(structure.FileTypes))
	}

	// Should have 3 content lengths
	if len(structure.ContentLengths) != 3 {
		t.Errorf("expected 3 content lengths, got %d", len(structure.ContentLengths))
	}

	// Average content length should be (100+200+150)/3 = 150
	expectedAvg := 150
	if avgLength(structure.ContentLengths) != expectedAvg {
		t.Errorf("expected average content length %d, got %d", expectedAvg, avgLength(structure.ContentLengths))
	}
}

func TestAnalyzeStructure_EmptySamples(t *testing.T) {
	agent := &SchemaAgent{}

	samples := []DocumentSample{}
	structure := agent.analyzeStructure(samples)

	// Should handle empty samples gracefully
	if len(structure.FileTypes) != 0 {
		t.Errorf("expected 0 file types for empty samples, got %d", len(structure.FileTypes))
	}

	if structure.AvgParagraphLen != 0 {
		t.Errorf("expected 0 average paragraph length for empty samples, got %d", structure.AvgParagraphLen)
	}

	if structure.ContentDensity != "sparse" {
		t.Errorf("expected sparse content density for empty samples, got %s", structure.ContentDensity)
	}
}

func TestAnalyzeStructure_LargeDocuments(t *testing.T) {
	agent := &SchemaAgent{}

	// Create a large document with multiple sections and paragraphs
	largeContent := "# Introduction\n\n"
	for i := 0; i < 100; i++ {
		largeContent += "This is paragraph number with some substantial content that makes it longer. "
		largeContent += "Adding more text to make each paragraph more realistic in size. "
		largeContent += "Typical paragraphs in documentation are around 100-200 words.\n\n"
	}
	largeContent += "## Section 1\n\nMore content here with details.\n\n## Section 2\n\nEven more detailed content."

	samples := []DocumentSample{
		{
			Path: "large.md",
			Type: "md",
			Size: int64(len(largeContent)),
			Fields: map[string]interface{}{
				"content": largeContent,
			},
			Preview: largeContent,
		},
	}

	structure := agent.analyzeStructure(samples)

	// Should have detected multiple sections (3 headers: #, ##, ##)
	if len(structure.SectionCounts) != 1 {
		t.Errorf("expected 1 section count entry, got %d", len(structure.SectionCounts))
	}
	if structure.SectionCounts[0] < 3 {
		t.Errorf("expected at least 3 sections, got %d", structure.SectionCounts[0])
	}

	// Should have detected many paragraphs
	if len(structure.ParagraphCounts) != 1 {
		t.Errorf("expected 1 paragraph count entry, got %d", len(structure.ParagraphCounts))
	}
	if structure.ParagraphCounts[0] < 50 {
		t.Errorf("expected at least 50 paragraphs, got %d", structure.ParagraphCounts[0])
	}

	// Document with 100+ paragraphs should be medium or dense (content length > 10000 bytes)
	// Just verify it's been classified (any classification is fine for this test)
	if structure.ContentDensity == "" {
		t.Error("content density should be classified")
	}
}

func TestAnalyzeStructure_TechnicalContent(t *testing.T) {
	agent := &SchemaAgent{}

	// Technical content with code blocks
	technicalContent := `# API Documentation

## Overview
This API provides access to the database.

## Code Example

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

## Parameters
- param1: string
- param2: int`

	samples := []DocumentSample{
		{
			Path: "api.md",
			Type: "md",
			Size: int64(len(technicalContent)),
			Fields: map[string]interface{}{
				"content": technicalContent,
			},
			Preview: technicalContent,
		},
	}

	structure := agent.analyzeStructure(samples)

	// Should detect sections
	if len(structure.SectionCounts) == 0 {
		t.Error("expected section counts to be populated")
	}

	// Should have reasonable paragraph count
	if len(structure.ParagraphCounts) == 0 {
		t.Error("expected paragraph counts to be populated")
	}
}

func TestBuildAnalysisPrompt_IncludesChunkingMetrics(t *testing.T) {
	agent := &SchemaAgent{}

	structure := DocumentStructure{
		Samples: []DocumentSample{
			{Path: "test.md", Type: "md", Size: 1000},
		},
		FileTypes:       []string{"md"},
		CommonFields:    map[string]int{"content": 1},
		FieldTypes:      map[string]string{"content": "text"},
		ContentLengths:  []int{1000},
		ParagraphCounts: []int{10},
		SectionCounts:   []int{3},
		AvgParagraphLen: 100,
		ContentDensity:  "medium",
	}

	input := &SchemaAnalysisInput{
		CollectionName: "test-collection",
		VDBType:        "weaviate",
		Requirements:   "Test requirements",
	}

	prompt := agent.buildAnalysisPrompt(structure, input)

	// Verify prompt includes chunking-related information
	if !containsString(prompt, "Average Paragraphs per Document: 10") {
		t.Error("prompt should include average paragraphs")
	}

	if !containsString(prompt, "Average Sections per Document: 3") {
		t.Error("prompt should include average sections")
	}

	if !containsString(prompt, "Average Paragraph Length: 100") {
		t.Error("prompt should include average paragraph length")
	}

	if !containsString(prompt, "Content Density: medium") {
		t.Error("prompt should include content density")
	}

	if !containsString(prompt, "chunking_advice") {
		t.Error("prompt should request chunking advice in JSON schema")
	}

	if !containsString(prompt, "recommended_size") {
		t.Error("prompt should request recommended chunk size")
	}

	if !containsString(prompt, "overlap_size") {
		t.Error("prompt should request overlap size")
	}
}

func TestContentDensityClassification(t *testing.T) {
	agent := &SchemaAgent{}

	tests := []struct {
		name            string
		size            int64
		expectedDensity string
	}{
		{"very small", 500, "sparse"},
		{"small", 999, "sparse"},
		{"medium low", 1000, "medium"},
		{"medium high", 4999, "medium"},
		{"large", 5000, "dense"},
		{"very large", 10000, "dense"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			samples := []DocumentSample{
				{
					Path:    "test.txt",
					Type:    "txt",
					Size:    tt.size,
					Fields:  map[string]interface{}{"content": "test"},
					Preview: "test",
				},
			}

			structure := agent.analyzeStructure(samples)

			if structure.ContentDensity != tt.expectedDensity {
				t.Errorf("for size %d, expected density %s, got %s",
					tt.size, tt.expectedDensity, structure.ContentDensity)
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr ||
				containsString(s[1:], substr)))
}
