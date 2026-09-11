// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestExtractImageAndPageNumbers(t *testing.T) {
	if min, max := extractImageNumbers(nil); min != 0 || max != 0 {
		t.Fatalf("extractImageNumbers(nil) = %d, %d", min, max)
	}
	images := []VirtualDocument{
		{OriginalFilename: "manual_image_12.png"},
		{OriginalFilename: "manual_image_bad.png"},
		{OriginalFilename: "manual_image_2.png"},
		{OriginalFilename: "manual.png"},
	}
	if min, max := extractImageNumbers(images); min != 2 || max != 12 {
		t.Fatalf("extractImageNumbers() = %d, %d", min, max)
	}

	if min, max := extractPageNumbers(nil); min != 0 || max != 0 {
		t.Fatalf("extractPageNumbers(nil) = %d, %d", min, max)
	}
	pages := []VirtualDocument{
		{OriginalFilename: "guide.pdf_page_9_image_1.png"},
		{OriginalFilename: "guide.pdf_page_bad_image_2.png"},
		{OriginalFilename: "guide.pdf_page_3_image_1.png"},
		{OriginalFilename: "guide.pdf_page_4.png"},
	}
	if min, max := extractPageNumbers(pages); min != 3 || max != 9 {
		t.Fatalf("extractPageNumbers() = %d, %d", min, max)
	}
}

func TestGroupImagesBySource(t *testing.T) {
	if groups := groupImagesBySource(nil); groups != nil {
		t.Fatalf("groupImagesBySource(nil) = %#v", groups)
	}
	docs := []VirtualDocument{
		{OriginalFilename: "guide.pdf_page_3_image_2.png", TotalChunks: 2, Metadata: map[string]interface{}{"type": "image"}},
		{OriginalFilename: "guide.pdf_page_1_image_1.png", TotalChunks: 1, Metadata: map[string]interface{}{"type": "image"}},
		{OriginalFilename: "photos_image_7.png", TotalChunks: 1},
		{OriginalFilename: "photos_image_4.png", TotalChunks: 2},
		{OriginalFilename: "standalone.png", TotalChunks: 1},
		{OriginalFilename: "broken_page_x_image_bad.png", TotalChunks: 1},
	}
	groups := groupImagesBySource(docs)
	if len(groups) != 4 {
		t.Fatalf("groupImagesBySource() returned %d groups: %#v", len(groups), groups)
	}
	bySource := make(map[string]ImageGroup, len(groups))
	for _, group := range groups {
		bySource[group.SourcePDF] = group
	}
	guide := bySource["guide.pdf"]
	if guide.DisplayName != "guide.pdf (pages 1-3)" || guide.TotalCount != 2 || guide.TotalChunks != 3 || !guide.IsImage {
		t.Fatalf("guide group = %#v", guide)
	}
	photos := bySource["photos"]
	if photos.DisplayName != "photos_image_4...7.png" || photos.TotalCount != 2 || photos.TotalChunks != 3 {
		t.Fatalf("photos group = %#v", photos)
	}
	if got := bySource["standalone.png"].DisplayName; got != "standalone.png" {
		t.Fatalf("standalone display name = %q", got)
	}
	if got := bySource["broken"].DisplayName; got != "broken (images)" {
		t.Fatalf("broken display name = %q", got)
	}

	onePage := groupImagesBySource([]VirtualDocument{{OriginalFilename: "one.pdf_page_2_image_1.png"}})
	if len(onePage) != 1 || onePage[0].DisplayName != "one.pdf (page 2)" {
		t.Fatalf("one-page group = %#v", onePage)
	}
}

func TestAggregateDocumentsByOriginal(t *testing.T) {
	docs := []weaviate.Document{
		{ID: "00000001-a", Content: "one", Metadata: map[string]interface{}{"source_document": "a.txt"}},
		{ID: "00000002-a", Content: "two", Metadata: map[string]interface{}{"source_document": "a.txt"}},
		{ID: "00000003-b", Metadata: map[string]interface{}{"source_file": "/tmp/b.txt"}},
		{ID: "00000004-c", Metadata: map[string]interface{}{"original_filename": "c.txt"}},
		{ID: "00000005-d", Metadata: map[string]interface{}{"filename": "d.txt"}},
		{ID: "00000006-e", Metadata: map[string]interface{}{"metadata": `{"filename":"e.txt","tag":"flat"}`}},
		{ID: "00000007-f", Metadata: map[string]interface{}{"metadata": `{"metadata":"{\"original_filename\":\"f.txt\",\"tag\":\"nested\"}"}`}},
		{ID: "00000008-g", Content: "fallback", Metadata: map[string]interface{}{"metadata": "invalid"}},
	}
	virtual := AggregateDocumentsByOriginal(docs)
	if len(virtual) != 7 {
		t.Fatalf("AggregateDocumentsByOriginal() returned %d documents: %#v", len(virtual), virtual)
	}
	if virtual[0].OriginalFilename != "a.txt" || virtual[0].TotalChunks != 2 {
		t.Fatalf("first virtual document = %#v", virtual[0])
	}
	if virtual[len(virtual)-1].OriginalFilename != "standalone-00000008-8" {
		t.Fatalf("fallback virtual document = %#v", virtual[len(virtual)-1])
	}
}

func TestVirtualDocumentHelpers(t *testing.T) {
	vdoc := VirtualDocument{Chunks: []weaviate.Document{{Metadata: map[string]interface{}{"content_type": "image/png"}}}}
	if !IsImageVirtualDocument(vdoc) {
		t.Fatal("IsImageVirtualDocument() = false")
	}
	if IsImageVirtualDocument(VirtualDocument{}) {
		t.Fatal("IsImageVirtualDocument(empty) = true")
	}
	doc := weaviate.Document{ID: "12345678-abcd", Content: "hello"}
	if got := GetStandaloneDocumentKey(doc); got != "standalone-12345678-5" {
		t.Fatalf("GetStandaloneDocumentKey() = %q", got)
	}
}

func TestSmartTruncate(t *testing.T) {
	if got := SmartTruncate("one\ntwo\nthree", "content", 2); got != "one\ntwo\n... (truncated)" {
		t.Fatalf("content truncation = %q", got)
	}
	longMetadata := strings.Repeat("m", 201)
	if got := SmartTruncate(longMetadata, "metadata", 2); len(got) != 200 || !strings.HasSuffix(got, "...") {
		t.Fatalf("metadata truncation length = %d", len(got))
	}
	if got := SmartTruncate("short", "metadata", 2); got != "short" {
		t.Fatalf("short metadata = %q", got)
	}
	longValue := strings.Repeat("v", 101)
	if got := SmartTruncate(longValue, "title", 2); len(got) != 100 || !strings.HasSuffix(got, "...") {
		t.Fatalf("value truncation length = %d", len(got))
	}
	if got := SmartTruncate("short", "title", 2); got != "short" {
		t.Fatalf("short value = %q", got)
	}
}

func TestDocumentDisplayModes(t *testing.T) {
	documents := []weaviate.Document{{
		ID:       "12345678-abcd",
		Content:  "line one\nline two\nline three",
		Metadata: map[string]interface{}{"filename": "guide.txt", "id": "duplicate", "note": strings.Repeat("x", 110)},
	}}
	DisplayRegularDocuments(documents, "Docs", false, 1, false, true, "text2vec", 3, 1, 1)
	DisplayRegularDocuments(documents, "Docs", false, 1, false, false, "text2vec", 3, 1, 1)
	DisplayRegularDocuments(documents, "Docs", true, 1, true, false, "", 1, 10, 0)
	DisplayRegularDocuments([]weaviate.Document{{ID: "12345678-abcd", Content: "Document ID: 12345678-abcd"}}, "Docs", false, 1, false, false, "", -1, 10, 0)
	DisplayRegularDocuments([]weaviate.Document{{ID: "12345678-abcd", Metadata: map[string]interface{}{"bad": make(chan int)}}}, "Docs", false, 1, false, true, "", -1, 10, 0)

	DisplayVirtualDocuments(nil, "Docs", false, 1, false, false, false, "", -1, 10, 0)
	DisplayVirtualDocuments(nil, "Docs", false, 1, false, false, true, "text2vec", -1, 10, 0)
	DisplayVirtualDocuments(documents, "Docs", false, 1, false, true, true, "text2vec", 3, 1, 0)
	DisplayVirtualDocuments(documents, "Docs", false, 1, false, true, false, "text2vec", 3, 1, 0)
	DisplayVirtualDocuments(documents, "Docs", true, 1, true, false, false, "", 1, 10, 0)
	DisplayVirtualDocuments([]weaviate.Document{{ID: "12345678-abcd", Metadata: map[string]interface{}{"filename": "bad.txt", "bad": make(chan int)}}}, "Docs", false, 1, false, false, true, "", -1, 10, 0)
}

func TestDocumentSchemaAndMetadataDisplays(t *testing.T) {
	ShowDocumentSchema(weaviate.Document{}, "Docs")
	doc := weaviate.Document{ID: "12345678-abcd", Metadata: map[string]interface{}{
		"title":  "Guide",
		"pages":  3,
		"score":  0.9,
		"active": true,
		"tags":   []interface{}{"go", "cli"},
		"extra":  map[string]interface{}{"version": 1},
		"long":   strings.Repeat("x", 210),
	}}
	ShowDocumentSchema(doc, "Docs")
	ShowDocumentMetadata(weaviate.Document{ID: "12345678-abcd"}, "Docs")
	ShowDocumentMetadata(doc, "Docs")
}

func TestDisplayQueryResultModes(t *testing.T) {
	results := []weaviate.QueryResult{
		{ID: "one", Content: "one\ntwo\nthree\nfour", Score: 0.2, Metadata: map[string]interface{}{"filename": "guide.txt"}},
		{ID: "two", Content: "content", Score: 0.1, Metadata: map[string]interface{}{"custom": "value"}},
	}
	DisplayQueryResults(results, "Docs", "guide", false, true, false)
	DisplayQueryResults(nil, "Docs", "missing", false, false, false)
	DisplayQueryResults(results, "Docs", "guide", false, false, false)
	results[1].Score = 0.8
	DisplayQueryResults(results, "Docs", "guide", true, false, true)
	DisplayQueryResults([]weaviate.QueryResult{{ID: "bad", Metadata: map[string]interface{}{"bad": make(chan int)}}}, "Docs", "bad", false, true, false)
}
