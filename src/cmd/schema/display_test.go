// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package schema

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/spf13/viper"
)

func captureSchemaOutput(t *testing.T, run func()) string {
	t.Helper()
	old := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = old })
	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestDisplaySchemaAnalysisCompleteRecommendation(t *testing.T) {
	output := &agents.SchemaAnalysisOutput{
		Schema: agents.SchemaConfig{
			CollectionName:   "Products",
			VectorDimensions: 1536,
			SimilarityMetric: "cosine",
			Fields: []agents.FieldConfig{
				{Name: "content", Type: "text", Description: "Searchable body", Indexed: true, Filterable: true, Required: true},
				{Name: "notes", Type: "text"},
			},
			Indexes: []agents.IndexConfig{{Name: "content-vector", Type: "vector", Fields: []string{"content"}}},
		},
		Reasoning:  "The samples contain product descriptions.",
		Confidence: 0.875,
		Warnings:   []string{"Review sparse fields"},
		FieldAnalysis: []agents.FieldSuggestion{{
			Name: "content", Frequency: 0.75, Cardinality: 12, Reasoning: "Present in most samples",
		}},
		Alternatives: []agents.SchemaConfig{{CollectionName: "ProductsCompact"}},
		ChunkingAdvice: &agents.ChunkingRecommendation{
			RecommendedSize: 800,
			MinSize:         400,
			MaxSize:         1200,
			OverlapSize:     80,
			DocumentType:    "catalog",
			Reasoning:       "Keep each product together.",
			Considerations:  []string{"Preserve specifications", "Keep titles with descriptions"},
		},
	}

	display := captureSchemaOutput(t, func() { displaySchemaAnalysis(output) })
	for _, want := range []string{
		"Schema Analysis for 'Products'",
		"Confidence: 87.5%",
		"The samples contain product descriptions.",
		"Vector Dimensions: 1536",
		"content              text [indexed] [filterable] [required]",
		"Searchable body",
		"content-vector (vector): [content]",
		"Review sparse fields",
		"content (frequency: 75.0%, cardinality: 12)",
		"Alternative Schemas Available: 1",
		"Recommended Size: 800 characters (~200 tokens)",
		"Overlap Size: 80 characters (~10%)",
		"Preserve specifications",
	} {
		if !strings.Contains(display, want) {
			t.Errorf("display missing %q:\n%s", want, display)
		}
	}
}

func TestDisplaySchemaAnalysisMinimalRecommendation(t *testing.T) {
	display := captureSchemaOutput(t, func() {
		displaySchemaAnalysis(&agents.SchemaAnalysisOutput{Schema: agents.SchemaConfig{CollectionName: "Empty"}})
	})
	if !strings.Contains(display, "Schema Analysis for 'Empty'") {
		t.Fatalf("unexpected minimal output:\n%s", display)
	}
	for _, absent := range []string{"Fields (", "Indexes (", "Warnings:", "Chunking Recommendations:"} {
		if strings.Contains(display, absent) {
			t.Errorf("minimal output unexpectedly contains %q:\n%s", absent, display)
		}
	}
}

func TestRunSchemaSuggestInputValidation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	viper.Set("OPENAI_API_KEY", "")
	t.Cleanup(viper.Reset)
	if err := runSchemaSuggest(suggestCmd, []string{filepath.Join(t.TempDir(), "missing")}); err == nil || !strings.Contains(err.Error(), "source path does not exist") {
		t.Fatalf("expected source error, got %v", err)
	}
	if err := runSchemaSuggest(suggestCmd, []string{t.TempDir()}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY required") {
		t.Fatalf("expected API key error, got %v", err)
	}
}

func TestSchemaFileAndGlobFailures(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.md"), []byte("sample"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := scanFiles(dir, "["); err == nil {
		t.Fatal("expected malformed glob error")
	}
	missingParent := filepath.Join(t.TempDir(), "missing", "schema.yaml")
	if err := saveSchemaToYAML(agents.SchemaConfig{CollectionName: "Products"}, missingParent); err == nil || !strings.Contains(err.Error(), "failed to write file") {
		t.Fatalf("expected write error, got %v", err)
	}
}
