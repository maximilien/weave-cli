// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func mockOperationConfig() *config.VectorDBConfig {
	return &config.VectorDBConfig{
		Name:               "fixture",
		Type:               config.VectorDBTypeMock,
		URL:                "http://fixture.invalid",
		APIKey:             "fixture-key",
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 3,
		Collections: []config.Collection{
			{Name: "AlphaDocs", Type: "text"},
			{Name: "BetaImages", Type: "image"},
		},
	}
}

func TestGenericCollectionOperationBoundaries(t *testing.T) {
	ctx := context.Background()
	cfg := mockOperationConfig()

	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil || !IsVectorDBTypeSupported("mock") || IsVectorDBTypeSupported("not-a-database") {
		t.Fatalf("vector database factory/support mismatch: %#v", GetSupportedVectorDBTypes())
	}
	if _, err := CreateVectorDBClient(&config.VectorDBConfig{Type: config.VectorDBType("invalid")}); err == nil {
		t.Fatal("unsupported database type was accepted")
	}

	if err := CreateGenericCollection(ctx, cfg, "CreatedDocs", "text-embedding-3-small"); err != nil {
		t.Fatalf("CreateGenericCollection() error = %v", err)
	}
	if err := CreateGenericCollectionWithSchemaType(ctx, cfg, "CreatedImages", "", vectordb.SchemaTypeImage); err != nil {
		t.Fatalf("CreateGenericCollectionWithSchemaType() error = %v", err)
	}
	if err := CreateGenericCollectionWithSchemaType(ctx, &config.VectorDBConfig{Type: "invalid"}, "bad", "", vectordb.SchemaTypeText); err == nil {
		t.Fatal("generic collection creation accepted invalid database")
	}

	count, err := CountGenericCollections(ctx, cfg)
	if err != nil || count != 2 {
		t.Fatalf("CountGenericCollections() = %d, %v", count, err)
	}
	if _, err := CountGenericCollections(ctx, &config.VectorDBConfig{Type: "invalid"}); err == nil {
		t.Fatal("generic collection count accepted invalid database")
	}

	if err := DeleteGenericCollections(ctx, cfg, []string{"AlphaDocs"}); err != nil {
		t.Fatalf("DeleteGenericCollections() error = %v", err)
	}
	if err := DeleteGenericCollections(ctx, cfg, []string{"Missing"}); err == nil {
		t.Fatal("generic deletion accepted missing collection")
	}
	if err := DeleteGenericCollections(ctx, &config.VectorDBConfig{Type: "invalid"}, nil); err == nil {
		t.Fatal("generic deletion accepted invalid database")
	}

	if err := DeleteGenericCollectionsByPattern(ctx, cfg, "Alpha*"); err != nil {
		t.Fatalf("pattern deletion error = %v", err)
	}
	if err := DeleteGenericCollectionsByPattern(ctx, cfg, "Missing*"); err != nil {
		t.Fatalf("no-match pattern deletion error = %v", err)
	}
	if err := DeleteGenericCollectionsByPattern(ctx, &config.VectorDBConfig{Type: "invalid"}, "*"); err == nil {
		t.Fatal("pattern deletion accepted invalid database")
	}

	ShowGenericCollection(ctx, cfg, "AlphaDocs", 3, false, false, false, false, false, false, false, "", "", false)
	ShowGenericCollection(ctx, cfg, "Missing", 3, false, false, false, false, false, false, false, "", "", false)
	ShowGenericCollection(ctx, &config.VectorDBConfig{Type: "invalid"}, "Missing", 3, false, false, false, false, false, false, false, "", "", false)

	QueryMockCollection(ctx, cfg, "AlphaDocs", "query", weaviate.QueryOptions{TopK: 2})
	QueryMockCollection(ctx, cfg, "Missing", "query", weaviate.QueryOptions{})
	ListDocuments(ctx, cfg, "AlphaDocs", 10, 0, false, 3, false, false, false, false)
}

func TestCollectionPatternAndPlaceholderBoundaries(t *testing.T) {
	for _, test := range []struct {
		value   string
		pattern string
		want    bool
	}{
		{value: "AlphaDocs", pattern: "Alpha*", want: true},
		{value: "AlphaDocs", pattern: "^Alpha.*$", want: true},
		{value: "BetaImages", pattern: "^Alpha.*$", want: false},
		{value: "AlphaDocs", pattern: "[", want: false},
		{value: "AlphaDocs", pattern: "*Docs", want: true},
	} {
		if got := matchPattern(test.value, test.pattern); got != test.want {
			t.Errorf("matchPattern(%q, %q) = %t", test.value, test.pattern, got)
		}
	}
	if !isRegexPattern("^docs$") || isRegexPattern("docs*") {
		t.Fatal("regex pattern classification mismatch")
	}
	if !isImageCollection("VacationPhotos") || isImageCollection("Documents") {
		t.Fatal("image collection classification mismatch")
	}
	if isConnectionRefusedError(nil) || !isConnectionRefusedError(errors.New("dial: connection refused")) || isConnectionRefusedError(errors.New("timeout")) {
		t.Fatal("connection-refused classification mismatch")
	}

	if _, err := ParseFieldDefinitions("title:string"); err == nil {
		t.Fatal("field parser placeholder returned no error")
	}
	cfg := mockOperationConfig()
	if err := CreateMockCollection(context.Background(), cfg, "Docs", "", nil); err == nil {
		t.Fatal("mock create placeholder returned no error")
	}
	if _, err := CountMockCollections(context.Background(), cfg); err == nil {
		t.Fatal("mock count placeholder returned no error")
	}
	if err := DeleteMockCollections(context.Background(), cfg, []string{"Docs"}); err == nil {
		t.Fatal("mock delete placeholder returned no error")
	}
	if err := DeleteMockCollectionsByPattern(context.Background(), cfg, "*"); err == nil {
		t.Fatal("mock pattern-delete placeholder returned no error")
	}
	ShowMockCollection(context.Background(), cfg, "Docs", 3, false, false, false, false, false, false, false, "", "", false)
}

func TestProcessingReportLifecycle(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "report.csv")
	report := &ProcessingReport{
		FilePath:   filepath.Join(root, "guide.pdf"),
		Collection: "Docs",
		Timestamp:  time.Date(2026, time.September, 22, 12, 0, 0, 0, time.UTC),
		ReportPath: path,
		ReportMode: "create",
		TextChunks: []ChunkReport{
			{ChunkNumber: 1, Success: true, SizeBytes: 2048},
			{ChunkNumber: 2, Success: false, Error: "chunk failed", SizeBytes: 1024},
		},
		Images: []ImageReport{
			{ImageNumber: 1, Filename: "image.png", Success: true, SizeBytes: 512, OCRWarnings: []string{"low contrast", "small text"}},
			{ImageNumber: 2, Filename: "bad.png", Success: false, Error: "ocr failed"},
		},
	}
	if got := GenerateDefaultReportPath(report.FilePath); got != filepath.Join(root, "guide.csv") {
		t.Fatalf("GenerateDefaultReportPath() = %q", got)
	}
	if err := WriteProcessingReport(report); err != nil {
		t.Fatal(err)
	}
	report.ReportMode = "append"
	report.TextChunks = report.TextChunks[:1]
	report.Images = nil
	if err := WriteProcessingReport(report); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"Timestamp,File,Collection", "chunk_1,success", "chunk_2,failed", "low contrast; small text"} {
		if !strings.Contains(text, want) {
			t.Errorf("report missing %q: %s", want, text)
		}
	}
	if err := WriteProcessingReport(&ProcessingReport{}); err != nil {
		t.Fatalf("empty report error = %v", err)
	}
	bad := &ProcessingReport{ReportPath: filepath.Join(root, "missing", "report.csv")}
	if err := WriteProcessingReport(bad); err == nil {
		t.Fatal("report in missing directory returned no error")
	}
}

func TestGlobPrintAndStyleBoundaries(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"one.txt", "two.txt", "image.png"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "dir.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !IsGlobPattern("*.txt") || IsGlobPattern("one.txt") {
		t.Fatal("glob pattern classification mismatch")
	}
	files, err := ExpandGlobPattern(filepath.Join(root, "*.txt"))
	if err != nil || len(files) != 2 {
		t.Fatalf("ExpandGlobPattern() = %#v, %v", files, err)
	}
	if _, err := ExpandGlobPattern("["); err == nil {
		t.Fatal("invalid glob returned no error")
	}

	withUtilsStdin(t, "yes\n", func() {
		if !ConfirmAction("continue") {
			t.Fatal("yes confirmation rejected")
		}
	})
	withUtilsStdin(t, "yes\n", func() {
		if !ConfirmActionStrict("continue") {
			t.Fatal("strict yes confirmation rejected")
		}
	})
	withUtilsStdin(t, "n\n", func() {
		if ConfirmAction("continue") {
			t.Fatal("negative confirmation accepted")
		}
	})

	PrintHeader("header")
	PrintStyledKey("key")
	PrintStyledValue("value")
	PrintStyledValueDimmed("dim")
	PrintStyledID("id")
	PrintStyledFilename("file")
	PrintStyledNumber(3)
	PrintStyledEmoji("x")
	PrintStyledKeyValueDimmed("key", "value")
	PrintStyledKeyProminent("key")
	PrintStyledKeyValueProminentWithEmoji("key", "value", "x")
	PrintStyledKeyValueProminentWithEmoji("key", "", "x")
	PrintStyledKeyNumberProminentWithEmoji("key", 3, "x")
	for _, value := range []string{
		GetStyledKeyProminent("key"), GetStyledKeyDimmed("key"), GetStyledNumber(3),
		GetStyledValueDimmed("value"), GetStyledEmoji("x"),
	} {
		if value == "" {
			t.Fatal("style helper returned empty output")
		}
	}
}

func withUtilsStdin(t *testing.T, input string, run func()) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	previous := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = previous
		_ = reader.Close()
	}()
	run()
}
