// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"strings"
	"testing"
)

func TestChunkingAgentMetricsAndPrompt(t *testing.T) {
	agent := &ChunkingAgent{config: defaultConfig()}
	samples := []DocumentSample{
		{Type: "markdown", Size: 900, Preview: "# Intro\n\nFirst paragraph.\n\nSecond paragraph."},
		{Type: "json", Size: 1100, Preview: "fallback", Fields: map[string]interface{}{"content": "section\nline two\nline three"}},
	}
	metrics := agent.analyzeChunkingMetrics(samples)
	if metrics.AvgContentLength != 1000 || metrics.AvgParagraphs != 3 || metrics.AvgSections != 0 || metrics.AvgParagraphLen == 0 {
		t.Fatalf("analyzeChunkingMetrics() = %#v", metrics)
	}
	if metrics.ContentDensity != "medium" || len(metrics.FileTypes) != 2 {
		t.Fatalf("density/types = %#v", metrics)
	}
	prompt := agent.buildChunkingPrompt(metrics, &ChunkingAnalysisInput{
		CollectionName: "docs", VDBType: "weaviate", Requirements: "preserve headings",
	})
	for _, expected := range []string{"Collection: docs", "Vector Database: weaviate", "preserve headings", "Content Density: medium"} {
		if !strings.Contains(prompt, expected) {
			t.Errorf("buildChunkingPrompt() omitted %q", expected)
		}
	}

	for _, test := range []struct {
		size int64
		want string
	}{
		{500, "sparse"},
		{3000, "medium"},
		{6000, "dense"},
	} {
		got := agent.analyzeChunkingMetrics([]DocumentSample{{Type: "text", Size: test.size, Preview: "one\n\ntwo"}})
		if got.ContentDensity != test.want {
			t.Errorf("density for %d = %q, want %q", test.size, got.ContentDensity, test.want)
		}
	}
	if got := agent.analyzeChunkingMetrics(nil); got.ContentDensity != "sparse" || got.AvgParagraphLen != 0 {
		t.Errorf("empty metrics = %#v", got)
	}
}

func TestChunkingAgentValidationPaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	agent := NewChunkingAgent(nil)
	if agent.Name() != "chunking-agent" || agent.config == nil {
		t.Fatalf("NewChunkingAgent() = %#v", agent)
	}
	if result, err := agent.Execute(context.Background(), "invalid"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}
	if result, err := agent.Execute(context.Background(), &ChunkingAnalysisInput{SampleFiles: []string{"missing.txt"}, MaxSamples: 1}); err == nil || result != nil {
		t.Fatalf("Execute(no samples) = %#v, %v", result, err)
	}
}
