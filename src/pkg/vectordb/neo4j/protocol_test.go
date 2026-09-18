// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package neo4j

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	neo4jsdk "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type neo4jFixtureDriver struct {
	neo4jsdk.DriverWithContext
	healthErr error
	closed    bool
}

func (d *neo4jFixtureDriver) VerifyConnectivity(context.Context) error { return d.healthErr }
func (d *neo4jFixtureDriver) Close(context.Context) error {
	d.closed = true
	return nil
}

type neo4jFixtureExecutor struct {
	queries []string
	params  []map[string]interface{}
	fail    error
}

func neo4jRecord(values map[string]interface{}) *neo4jsdk.Record {
	keys := make([]string, 0, len(values))
	vals := make([]interface{}, 0, len(values))
	for key, value := range values {
		keys = append(keys, key)
		vals = append(vals, value)
	}
	return &neo4jsdk.Record{Keys: keys, Values: vals}
}

func (e *neo4jFixtureExecutor) ExecuteQuery(_ context.Context, query string, params map[string]interface{}, _ string) (*neo4jsdk.EagerResult, error) {
	e.queries = append(e.queries, query)
	e.params = append(e.params, params)
	if e.fail != nil {
		return nil, e.fail
	}

	var records []*neo4jsdk.Record
	switch {
	case strings.Contains(query, "SHOW INDEXES") && strings.Contains(query, "labelsOrTypes"):
		records = []*neo4jsdk.Record{neo4jRecord(map[string]interface{}{
			"name": "docs_vector_idx", "type": "VECTOR", "entityType": "NODE",
			"labelsOrTypes": []interface{}{"docs"}, "properties": []interface{}{"embedding"},
			"options": map[string]interface{}{"indexConfig": map[string]interface{}{"vector.dimensions": int64(3)}},
		})}
	case strings.Contains(query, "SHOW INDEXES") && strings.Contains(query, "type = 'VECTOR'"):
		records = []*neo4jsdk.Record{
			neo4jRecord(map[string]interface{}{"name": "docs_vector_idx"}),
			neo4jRecord(map[string]interface{}{"name": "ignored"}),
		}
	case strings.Contains(query, "SHOW INDEXES"):
		records = []*neo4jsdk.Record{neo4jRecord(map[string]interface{}{"name": "docs_vector_idx"})}
	case strings.Contains(query, "count(n) as count"):
		records = []*neo4jsdk.Record{neo4jRecord(map[string]interface{}{"count": int64(2)})}
	case strings.Contains(query, "RETURN d.id as id") || strings.Contains(query, "RETURN n.id as id") || strings.Contains(query, "RETURN node.id as id"):
		records = []*neo4jsdk.Record{neo4jRecord(map[string]interface{}{
			"id": "doc-1", "content": "hello", "embedding": []interface{}{float64(1), float64(2), float64(3)},
			"props": map[string]interface{}{"id": "doc-1", "content": "hello", "embedding": []float64{1}, "kind": "guide"},
			"score": float64(0.75),
		})}
	case strings.Contains(query, "RETURN d.id as id"):
		records = []*neo4jsdk.Record{neo4jRecord(map[string]interface{}{"id": "doc-1"})}
	}
	return &neo4jsdk.EagerResult{Records: records}, nil
}

func newNeo4jProtocolFixture() (*Adapter, *neo4jFixtureExecutor, *neo4jFixtureDriver) {
	executor := &neo4jFixtureExecutor{}
	driver := &neo4jFixtureDriver{}
	client := &Client{
		driver:   driver,
		executor: executor,
		config: &Config{
			URI: "bolt://fixture:7687", Database: "neo4j", Timeout: time.Second,
			VectorDimensions: 3, SimilarityMetric: "cosine",
		},
	}
	return &Adapter{client: client}, executor, driver
}

func TestNeo4jCollectionProtocol(t *testing.T) {
	adapter, _, _ := newNeo4jProtocolFixture()
	client := adapter.client
	ctx := context.Background()

	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if err := adapter.CreateCollection(ctx, "docs", nil); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 1 || collections[0].Name != "docs" || collections[0].Count != 2 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	exists, err := adapter.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	count, err := adapter.GetCollectionCount(ctx, "docs")
	if err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	info, err := client.GetCollectionInfo(ctx, "docs")
	if err != nil || info["type"] != "VECTOR" {
		t.Fatalf("GetCollectionInfo() = %#v, %v", info, err)
	}
	dimensions, err := client.getIndexDimensions(ctx, "docs")
	if err != nil || dimensions != 3 {
		t.Fatalf("getIndexDimensions() = %d, %v", dimensions, err)
	}
	schema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || schema.Vectorizer != "text-embedding-3-small" {
		t.Fatalf("GetSchema() = %#v, %v", schema, err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	if err := client.DeleteCollection(ctx, "docs", false); err != nil {
		t.Fatalf("DeleteCollection(false) error = %v", err)
	}
}

func TestNeo4jDocumentProtocol(t *testing.T) {
	adapter, _, _ := newNeo4jProtocolFixture()
	client := adapter.client
	ctx := context.Background()
	doc := &vectordb.Document{ID: "doc-1", Content: "hello", Text: "fallback", Metadata: map[string]interface{}{"kind": "guide"}}

	if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc, {ID: "doc-2", Text: "second"}}); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	if err := client.BatchCreateDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("empty BatchCreateDocuments() error = %v", err)
	}
	got, err := adapter.GetDocument(ctx, "docs", "doc-1")
	if err != nil || got.ID != "doc-1" || got.Content != "hello" || got.Metadata["kind"] != "guide" {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := adapter.UpdateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	if err := adapter.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{"doc-1", "doc-2"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("empty DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide", "rank": 2}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	if err := client.DeleteAllDocuments(ctx, "docs"); err != nil {
		t.Fatalf("DeleteAllDocuments() error = %v", err)
	}
	docs, err := adapter.ListDocuments(ctx, "docs", 0, 2)
	if err != nil || len(docs) != 1 || docs[0].Metadata["kind"] != "guide" {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
	clientDocs, err := client.ListDocuments(ctx, "docs", 0)
	if err != nil || len(clientDocs) != 1 || len(clientDocs[0].Vector) != 3 {
		t.Fatalf("client.ListDocuments() = %#v, %v", clientDocs, err)
	}
}

func TestNeo4jQueryProtocol(t *testing.T) {
	adapter, _, _ := newNeo4jProtocolFixture()
	client := adapter.client
	ctx := context.Background()

	results, err := client.VectorSearch(ctx, "docs", []float32{1, 2, 3}, 0)
	if err != nil || len(results) != 1 || results[0].Score != 0.75 || results[0].Document.ID != "doc-1" {
		t.Fatalf("VectorSearch() = %#v, %v", results, err)
	}
	results, err = client.VectorSearchWithFilter(ctx, "docs", []float32{1, 2, 3}, 0, map[string]interface{}{"kind": "guide"})
	if err != nil || len(results) != 1 || results[0].Document.Metadata["kind"] != "guide" {
		t.Fatalf("VectorSearchWithFilter() = %#v, %v", results, err)
	}
	results, err = client.VectorSearchWithFilter(ctx, "docs", nil, 2, nil)
	if err != nil || len(results) != 1 {
		t.Fatalf("unfiltered VectorSearchWithFilter() = %#v, %v", results, err)
	}
	metadata, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide", "rank": 2}, &vectordb.QueryOptions{TopK: 3})
	if err != nil || len(metadata) != 1 || metadata[0].Document.ID != "doc-1" {
		t.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
	}
	if _, err := adapter.SearchBM25(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchBM25() expected unsupported error")
	}
	if _, err := adapter.SearchHybrid(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchHybrid() expected unsupported error")
	}
}

func TestNeo4jProtocolErrorsAndHelpers(t *testing.T) {
	ctx := context.Background()
	adapter, executor, driver := newNeo4jProtocolFixture()
	client := adapter.client

	executor.fail = errors.New("fixture query failure")
	for name, call := range map[string]func() error{
		"create collection": func() error { return client.CreateCollection(ctx, "docs", 3, "cosine") },
		"list collections":  func() error { _, err := client.ListCollections(ctx); return err },
		"exists":            func() error { _, err := client.CollectionExists(ctx, "docs"); return err },
		"delete":            func() error { return client.DeleteCollection(ctx, "docs", true) },
		"count":             func() error { _, err := client.GetCollectionCount(ctx, "docs"); return err },
		"create document":   func() error { return client.CreateDocument(ctx, "docs", &Document{ID: "id"}) },
		"get document":      func() error { _, err := client.GetDocument(ctx, "docs", "id"); return err },
		"update document":   func() error { return client.UpdateDocument(ctx, "docs", &Document{ID: "id"}) },
		"delete document":   func() error { return client.DeleteDocument(ctx, "docs", "id") },
		"list documents":    func() error { _, err := client.ListDocuments(ctx, "docs", 1); return err },
		"vector search":     func() error { _, err := client.VectorSearch(ctx, "docs", nil, 1); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	executor.fail = nil

	for _, message := range []string{"connection refused", "deadline timeout", "authentication failed", "plain failure"} {
		driver.healthErr = errors.New(message)
		err := client.Health(ctx)
		if err == nil || !strings.Contains(err.Error(), strings.Fields(message)[0]) {
			t.Fatalf("Health(%q) error = %v", message, err)
		}
	}
	driver.healthErr = nil
	if err := adapter.Close(ctx); err != nil || !driver.closed {
		t.Fatalf("Close() = %v, closed = %v", err, driver.closed)
	}
	if err := (&Adapter{}).Close(ctx); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}

	for dims, want := range map[int]string{
		768: "sentence-transformers/all-mpnet-base-v2", 384: "sentence-transformers/all-MiniLM-L6-v2",
		1536: "text-embedding-3-small", 3072: "text-embedding-3-large", 1024: "nomic-embed-text", 42: "text-embedding-3-small",
	} {
		if got := inferEmbeddingModelFromDimensions(dims); got != want {
			t.Errorf("inferEmbeddingModelFromDimensions(%d) = %q, want %q", dims, got, want)
		}
	}
}
