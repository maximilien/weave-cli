// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stats

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/viper"
)

func captureStatsOutput(t *testing.T, run func()) string {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldColorOutput := color.Output
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	os.Stderr = writer
	color.Output = writer
	t.Cleanup(func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		color.Output = oldColorOutput
	})
	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	color.Output = oldColorOutput
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}

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

func TestCollectStatsWithMockDatabase(t *testing.T) {
	dbConfig := &config.VectorDBConfig{
		Name:               "mock",
		Type:               config.VectorDBTypeMock,
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
	}
	stats, err := collectStats(context.Background(), dbConfig, "Documents", 100, 5)
	if err != nil {
		t.Fatal(err)
	}
	if stats.CollectionName != "Documents" || stats.VectorDB != "mock" || stats.DocumentCount != 0 || len(stats.Metadata) != 0 {
		t.Fatalf("unexpected mock statistics: %#v", stats)
	}

	badConfig := &config.VectorDBConfig{Type: config.VectorDBType("unsupported")}
	if _, err := collectStats(context.Background(), badConfig, "Documents", 100, 5); err == nil || !strings.Contains(err.Error(), "failed to create client") {
		t.Fatalf("expected client creation error, got %v", err)
	}
}

func TestDisplayStatsWithMetadata(t *testing.T) {
	stats := &CollectionStats{
		CollectionName: "Documents",
		DocumentCount:  3,
		VectorDB:       "mock",
		Metadata: map[string]MetadataFieldStat{
			"year": {
				Type:         "int",
				UniqueValues: 2,
				TopValues:    []ValueCount{{Value: "2026", Count: 2}, {Value: "2025", Count: 1}},
			},
			"author": {
				Type:         "string",
				UniqueValues: 1,
				TopValues:    []ValueCount{{Value: "Max", Count: 3}},
			},
		},
	}
	output := captureStatsOutput(t, func() { displayStats(stats, 5) })
	for _, want := range []string{
		"Collection Statistics: Documents",
		"Collection: Documents",
		"Vector DB: mock",
		"Documents: 3",
		"Metadata Fields (2 total)",
		"author",
		"Unique values: 1",
		"Max (3 occurrences)",
		"2026 (2 occurrences)",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("display output missing %q:\n%s", want, output)
		}
	}
	if strings.Index(output, "author") > strings.Index(output, "year") {
		t.Fatalf("metadata fields should be alphabetically sorted:\n%s", output)
	}
}

func TestDisplayStatsWithoutMetadata(t *testing.T) {
	output := captureStatsOutput(t, func() {
		displayStats(&CollectionStats{CollectionName: "Empty", VectorDB: "mock"}, 5)
	})
	if !strings.Contains(output, "No metadata fields found") {
		t.Fatalf("unexpected empty metadata output:\n%s", output)
	}
}

func TestRunStatsOutputFormats(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
	t.Setenv("VECTOR_DB_TYPE", "mock")
	t.Setenv("WEAVIATE_URL", "https://fixture.invalid")
	t.Setenv("WEAVIATE_API_KEY", "fixture")
	t.Setenv("OPENAI_API_KEY", "fixture")
	configPath := filepath.Join(root, "config.yaml")
	configData := `databases:
  default: fixture
  vector_databases:
    - name: fixture
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 3
`
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	viper.Set("config", configPath)
	viper.Set("env", "")
	viper.Set("quiet", true)
	viper.Set("no-color", true)
	t.Cleanup(func() {
		_ = Cmd.Flags().Set("output", "text")
		viper.Reset()
	})

	tests := []struct {
		format string
		want   string
	}{
		{format: "text", want: "Collection Statistics: Docs"},
		{format: "json", want: `"collection_name": "Docs"`},
		{format: "yaml", want: "collection_name: Docs"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if err := Cmd.Flags().Set("output", tt.format); err != nil {
				t.Fatal(err)
			}
			output := captureStatsOutput(t, func() { runStats(Cmd, []string{"Docs"}) })
			if !strings.Contains(output, tt.want) {
				t.Fatalf("%s output missing %q:\n%s", tt.format, tt.want, output)
			}
		})
	}
}
