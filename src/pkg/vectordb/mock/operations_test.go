// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mock

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAdapterCollectionAndDocumentLifecycle(t *testing.T) {
	adapter := createTestAdapter(t)
	ctx := context.Background()
	schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "docs")

	if err := adapter.CreateCollection(ctx, "docs", schema); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	exists, err := adapter.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	if exists, err := adapter.CollectionExists(ctx, "missing"); err != nil || exists {
		t.Fatalf("CollectionExists(missing) = %v, %v", exists, err)
	}

	docs := []*vectordb.Document{
		{ID: "one", Content: "machine learning guide", Metadata: map[string]interface{}{"kind": "guide", "year": 2026}},
		{ID: "two", Text: "vector database reference", Metadata: map[string]interface{}{"kind": "reference"}},
		{ID: "three", Content: "machine learning examples", Metadata: map[string]interface{}{"kind": "guide"}},
	}
	if err := adapter.CreateDocuments(ctx, "docs", docs); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	count, err := adapter.GetCollectionCount(ctx, "docs")
	if err != nil || count != 3 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 1 || collections[0].Name != "docs" || collections[0].Count != 3 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}

	got, err := adapter.GetDocument(ctx, "docs", "one")
	if err != nil || got.ID != "one" || got.Content != "machine learning guide" || got.Text != got.Content {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	update := &vectordb.Document{ID: "one", Text: "updated via text", Metadata: map[string]interface{}{"kind": "updated"}}
	if err := adapter.UpdateDocument(ctx, "docs", update); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	got, err = adapter.GetDocument(ctx, "docs", "one")
	if err != nil || got.Content != "updated via text" || got.Metadata["kind"] != "updated" {
		t.Fatalf("updated GetDocument() = %#v, %v", got, err)
	}

	listed, err := adapter.ListDocuments(ctx, "docs", 1, 1)
	if err != nil || len(listed) != 1 || listed[0].ID != "two" {
		t.Fatalf("ListDocuments() = %#v, %v", listed, err)
	}
	listed, err = adapter.ListDocuments(ctx, "docs", 5, 99)
	if err != nil || len(listed) != 0 {
		t.Fatalf("ListDocuments(offset past end) = %#v, %v", listed, err)
	}

	if err := adapter.DeleteDocument(ctx, "docs", "one"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{"two"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	count, err = adapter.GetCollectionCount(ctx, "docs")
	if err != nil || count != 0 {
		t.Fatalf("final GetCollectionCount() = %d, %v", count, err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	if _, err := adapter.GetCollectionCount(ctx, "docs"); err == nil {
		t.Fatal("GetCollectionCount(deleted) expected error")
	}
}

func TestAdapterSearchOperations(t *testing.T) {
	adapter := createTestAdapter(t)
	ctx := context.Background()
	if err := adapter.CreateCollection(ctx, "docs", adapter.GetDefaultSchema(vectordb.SchemaTypeText, "docs")); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	docs := []*vectordb.Document{
		{ID: "one", Content: "machine learning guide", Metadata: map[string]interface{}{"kind": "guide"}},
		{ID: "two", Content: "database reference", Metadata: map[string]interface{}{"kind": "reference"}},
		{ID: "three", Content: "machine learning examples", Metadata: map[string]interface{}{"kind": "guide"}},
	}
	if err := adapter.CreateDocuments(ctx, "docs", docs); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}

	options := &vectordb.QueryOptions{TopK: 2, SearchMetadata: true, NoTruncate: true}
	semantic, err := adapter.SearchSemantic(ctx, "docs", "machine learning", options)
	if err != nil || len(semantic) != 2 || semantic[0].Document.ID == "" {
		t.Fatalf("SearchSemantic() = %#v, %v", semantic, err)
	}
	bm25, err := adapter.SearchBM25(ctx, "docs", "machine", options)
	if err != nil || len(bm25) != 2 {
		t.Fatalf("SearchBM25() = %#v, %v", bm25, err)
	}
	hybrid, err := adapter.SearchHybrid(ctx, "docs", "database", nil)
	if err != nil || len(hybrid) == 0 {
		t.Fatalf("SearchHybrid() = %#v, %v", hybrid, err)
	}
	metadata, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, &vectordb.QueryOptions{TopK: 1})
	if err != nil || len(metadata) != 2 || metadata[0].Score != 1 {
		t.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
	}
	if _, err := adapter.SearchSemantic(ctx, "missing", "query", nil); err == nil {
		t.Fatal("SearchSemantic(missing) expected error")
	}
}

func TestAdapterSchemaOperationsAndHelpers(t *testing.T) {
	adapter := createTestAdapter(t)
	ctx := context.Background()

	schema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || schema.Class != "docs" || len(schema.Properties) != 2 {
		t.Fatalf("GetSchema() = %#v, %v", schema, err)
	}
	if err := adapter.UpdateSchema(ctx, "docs", schema); err != nil {
		t.Fatalf("UpdateSchema() error = %v", err)
	}
	imageSchema := adapter.GetDefaultSchema(vectordb.SchemaTypeImage, "images")
	if imageSchema.Vectorizer != "none" || len(imageSchema.Properties) != 3 {
		t.Fatalf("GetDefaultSchema(image) = %#v", imageSchema)
	}
	if err := adapter.ValidateSchema(nil); err == nil {
		t.Fatal("ValidateSchema(nil) expected error")
	}
	if err := adapter.ValidateSchema(&vectordb.CollectionSchema{}); err == nil {
		t.Fatal("ValidateSchema(empty class) expected error")
	}
	if err := adapter.ValidateSchema(schema); err != nil {
		t.Fatalf("ValidateSchema(valid) error = %v", err)
	}

	if got := convertToPointers(nil); got == nil || len(got) != 0 {
		t.Fatalf("convertToPointers(nil) = %#v", got)
	}
	for _, test := range []struct {
		value interface{}
		want  string
	}{
		{"text", "text"},
		{42, "42"},
		{2.5, "2.5"},
		{true, "true"},
		{[]string{"unsupported"}, ""},
	} {
		if got := toString(test.value); got != test.want {
			t.Fatalf("toString(%#v) = %q, want %q", test.value, got, test.want)
		}
	}
}
