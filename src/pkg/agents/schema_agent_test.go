// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"testing"
)

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
