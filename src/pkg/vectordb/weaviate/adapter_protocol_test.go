// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package weaviate

import (
	"context"
	"errors"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAdapterQueryProtocols(t *testing.T) {
	client, _ := newQueryProtocolClient(t)
	adapter := &Adapter{client: client}
	ctx := context.Background()
	opts := &vectordb.QueryOptions{TopK: 2, SearchMetadata: true}

	semantic, err := adapter.SearchSemantic(ctx, "Docs", "semantic", opts)
	if err != nil || len(semantic) != 1 || semantic[0].Document.Content != "semantic result" || semantic[0].Document.ID != "result-1" {
		t.Fatalf("SearchSemantic() = (%#v, %v)", semantic, err)
	}
	bm25, err := adapter.SearchBM25(ctx, "Docs", "keyword", opts)
	if err != nil || len(bm25) != 1 || bm25[0].Document.Content != "keyword result" {
		t.Fatalf("SearchBM25() = (%#v, %v)", bm25, err)
	}
	filtered, err := adapter.SearchByMetadata(ctx, "Docs", map[string]interface{}{"source": "fake"}, opts)
	if err != nil || len(filtered) != 1 {
		t.Fatalf("SearchByMetadata() = (%#v, %v)", filtered, err)
	}

	hybrid, err := adapter.SearchHybrid(ctx, "Docs", "semantic", opts)
	if err != nil || len(hybrid) != 1 {
		t.Fatalf("SearchHybrid(semantic) = (%#v, %v)", hybrid, err)
	}
	fallback, err := adapter.SearchHybrid(ctx, "FallbackDocs", "fallback", &vectordb.QueryOptions{TopK: 4})
	if err != nil || len(fallback) != 1 {
		t.Fatalf("SearchHybrid(fallback) = (%#v, %v)", fallback, err)
	}

	for name, run := range map[string]func() error{
		"semantic": func() error { _, err := adapter.SearchSemantic(ctx, "Missing", "query", opts); return err },
		"bm25":     func() error { _, err := adapter.SearchBM25(ctx, "Missing", "query", opts); return err },
		"metadata": func() error { _, err := adapter.SearchByMetadata(ctx, "Missing", nil, opts); return err },
	} {
		t.Run(name+" error", func(t *testing.T) {
			if err := run(); err == nil {
				t.Fatal("expected wrapped query error")
			}
		})
	}

	if got := adapter.convertWeaviateQueryResults(nil); got != nil {
		t.Fatalf("convert nil results = %#v", got)
	}
}

func TestAdapterCollectionProtocols(t *testing.T) {
	client, _ := newQueryProtocolClient(t)
	adapter := &Adapter{client: client}
	ctx := context.Background()

	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) == 0 {
		t.Fatalf("ListCollections() = (%#v, %v)", collections, err)
	}
	exists, err := adapter.CollectionExists(ctx, "Docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists(Docs) = (%t, %v)", exists, err)
	}
	exists, err = adapter.CollectionExists(ctx, "Missing")
	if err != nil || exists {
		t.Fatalf("CollectionExists(Missing) = (%t, %v)", exists, err)
	}
	schema, err := adapter.GetSchema(ctx, "Docs")
	if err != nil || schema.Class != "Docs" || len(schema.Properties) == 0 {
		t.Fatalf("GetSchema() = (%#v, %v)", schema, err)
	}
	if _, err := adapter.GetSchema(ctx, "Missing"); err == nil {
		t.Fatal("GetSchema(Missing) expected error")
	}
	if err := adapter.DeleteCollection(ctx, "Docs"); err == nil {
		t.Fatal("DeleteCollection() expected fake-server error")
	}
}

func TestAdapterSchemaProtocols(t *testing.T) {
	adapter := &Adapter{}

	text := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "Docs")
	if text.Class != "Docs" || len(text.Properties) != 3 || text.Vectorizer != "text-embedding-3-small" {
		t.Fatalf("text schema = %#v", text)
	}
	image := adapter.GetDefaultSchema(vectordb.SchemaTypeImage, "Images")
	if len(image.Properties) != 5 || image.Properties[3].Name != "image" || image.Properties[4].Name != "image_data" {
		t.Fatalf("image schema = %#v", image)
	}

	valid := &vectordb.CollectionSchema{
		Class: "Docs",
		Properties: []vectordb.SchemaProperty{{
			Name: "metadata", DataType: []string{"object"},
			NestedProperties: []vectordb.SchemaProperty{{
				Name: "author", DataType: []string{"text"},
				NestedProperties: []vectordb.SchemaProperty{{Name: "name", DataType: []string{"text"}}},
			}},
		}},
	}
	if err := adapter.ValidateSchema(valid); err != nil {
		t.Fatalf("ValidateSchema(valid) error = %v", err)
	}

	cases := map[string]*vectordb.CollectionSchema{
		"nil":                  nil,
		"class":                {Properties: []vectordb.SchemaProperty{{Name: "text", DataType: []string{"text"}}}},
		"properties":           {Class: "Docs"},
		"property name":        {Class: "Docs", Properties: []vectordb.SchemaProperty{{DataType: []string{"text"}}}},
		"property type":        {Class: "Docs", Properties: []vectordb.SchemaProperty{{Name: "text"}}},
		"nested name":          {Class: "Docs", Properties: []vectordb.SchemaProperty{{Name: "metadata", DataType: []string{"object"}, NestedProperties: []vectordb.SchemaProperty{{DataType: []string{"text"}}}}}},
		"nested type":          {Class: "Docs", Properties: []vectordb.SchemaProperty{{Name: "metadata", DataType: []string{"object"}, NestedProperties: []vectordb.SchemaProperty{{Name: "author"}}}}},
		"recursive nested bad": {Class: "Docs", Properties: []vectordb.SchemaProperty{{Name: "metadata", DataType: []string{"object"}, NestedProperties: []vectordb.SchemaProperty{{Name: "author", DataType: []string{"object"}, NestedProperties: []vectordb.SchemaProperty{{DataType: []string{"text"}}}}}}}},
	}
	for name, schema := range cases {
		t.Run(name, func(t *testing.T) {
			if err := adapter.ValidateSchema(schema); err == nil {
				t.Fatal("expected invalid schema error")
			}
		})
	}

	if err := adapter.UpdateSchema(context.Background(), "Docs", valid); err == nil {
		t.Fatal("UpdateSchema() expected unsupported error")
	}
}

func TestAdapterErrorCategoriesAndStringConversion(t *testing.T) {
	adapter := &Adapter{}
	for _, message := range []string{"connection refused", "no such host", "timeout", "unauthorized", "forbidden", "not found", "unexpected"} {
		if err := adapter.wrapError(errors.New(message), "operation"); err == nil {
			t.Fatalf("wrapError(%q) returned nil", message)
		}
	}
	if err := adapter.wrapError(nil, "operation"); err != nil {
		t.Fatalf("wrapError(nil) = %v", err)
	}

	cases := map[interface{}]string{"text": "text", 7: "7", 2.5: "2.5", true: "true", struct{}{}: ""}
	for value, want := range cases {
		if got := toString(value); got != want {
			t.Errorf("toString(%v) = %q, want %q", value, got, want)
		}
	}
}
