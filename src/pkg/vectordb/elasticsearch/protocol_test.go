// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package elasticsearch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type elasticsearchRecorder struct {
	mu     sync.Mutex
	paths  []string
	bodies []string
}

func (r *elasticsearchRecorder) record(req *http.Request) string {
	body, _ := io.ReadAll(req.Body)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, req.Method+" "+req.URL.Path)
	r.bodies = append(r.bodies, string(body))
	return string(body)
}

func writeBulkResponse(w http.ResponseWriter, body string) {
	items := make([]map[string]interface{}, 0)
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for i := 0; i < len(lines); i++ {
		var action map[string]map[string]interface{}
		if json.Unmarshal([]byte(lines[i]), &action) != nil {
			continue
		}
		for operation, detail := range action {
			id, _ := detail["_id"].(string)
			items = append(items, map[string]interface{}{operation: map[string]interface{}{
				"_index": "docs", "_id": id, "_version": 1, "result": "created", "status": 201,
				"_shards": map[string]int{"total": 1, "successful": 1, "failed": 0},
			}})
			if operation != "delete" {
				i++
			}
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"took": 1, "errors": false, "items": items})
}

func newElasticsearchProtocolAdapter(t *testing.T) (*Adapter, *elasticsearchRecorder) {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", "")
	recorder := &elasticsearchRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body := recorder.record(req)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		path := req.URL.Path
		switch {
		case path == "/" && req.Method == http.MethodHead:
			w.WriteHeader(http.StatusOK)
		case path == "/_bulk" || strings.HasSuffix(path, "/_bulk"):
			writeBulkResponse(w, body)
		case strings.HasSuffix(path, "/_delete_by_query"):
			_, _ = io.WriteString(w, `{"took":1,"timed_out":false,"total":2,"deleted":2,"batches":1,"version_conflicts":0,"noops":0,"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0,"failures":[]}`)
		case strings.HasSuffix(path, "/_search"):
			_, _ = io.WriteString(w, `{"took":1,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":2,"relation":"eq"},"max_score":0.9,"hits":[{"_index":"docs","_id":"one","_score":0.9,"_source":{"document_id":"one","text":"hello","content":"world","image":"image","image_data":"data","url":"https://example.test","metadata":{"kind":"guide"}}},{"_index":"docs","_id":"bad","_score":null,"_source":"invalid"}]}}`)
		case strings.HasSuffix(path, "/_mapping"):
			_, _ = io.WriteString(w, `{"docs":{"mappings":{"properties":{"vector_field":{"type":"dense_vector","dims":3,"index":true,"similarity":"cosine"}}}}}`)
		case path == "/*" || path == "/%2A":
			_, _ = io.WriteString(w, `{"docs":{"aliases":{},"mappings":{},"settings":{}},".system":{"aliases":{},"mappings":{},"settings":{}}}`)
		case strings.HasSuffix(path, "/_count"):
			_, _ = io.WriteString(w, `{"count":2,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0}}`)
		case strings.Contains(path, "/_doc/") && req.Method == http.MethodGet:
			if strings.HasSuffix(path, "/missing") {
				_, _ = io.WriteString(w, `{"_index":"docs","_id":"missing","found":false}`)
			} else if strings.HasSuffix(path, "/malformed") {
				_, _ = io.WriteString(w, `{"_index":"docs","_id":"malformed","found":true,"_source":"bad"}`)
			} else {
				_, _ = io.WriteString(w, `{"_index":"docs","_id":"one","_version":1,"_seq_no":0,"_primary_term":1,"found":true,"_source":{"document_id":"one","text":"hello","content":"world","image":"image","image_data":"data","url":"https://example.test","metadata":{"kind":"guide"}}}`)
			}
		case strings.Contains(path, "/_doc/") && req.Method == http.MethodDelete:
			_, _ = io.WriteString(w, `{"_index":"docs","_id":"one","_version":2,"result":"deleted","_shards":{"total":1,"successful":1,"failed":0},"_seq_no":1,"_primary_term":1}`)
		case strings.Contains(path, "/_doc/") && (req.Method == http.MethodPut || req.Method == http.MethodPost):
			_, _ = io.WriteString(w, `{"_index":"docs","_id":"one","_version":1,"result":"created","_shards":{"total":1,"successful":1,"failed":0},"_seq_no":0,"_primary_term":1}`)
		case strings.Contains(path, "/_update/"):
			_, _ = io.WriteString(w, `{"_index":"docs","_id":"one","_version":2,"result":"updated","_shards":{"total":1,"successful":1,"failed":0},"_seq_no":1,"_primary_term":1}`)
		case req.Method == http.MethodHead:
			if strings.Contains(path, "missing") {
				http.NotFound(w, req)
				return
			}
			w.WriteHeader(http.StatusOK)
		case req.Method == http.MethodPut:
			acknowledged := !strings.Contains(path, "unacknowledged")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"acknowledged": acknowledged, "shards_acknowledged": acknowledged, "index": strings.TrimPrefix(path, "/")})
		case req.Method == http.MethodDelete:
			acknowledged := !strings.Contains(path, "unacknowledged")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"acknowledged": acknowledged})
		default:
			http.Error(w, `{"error":{"type":"unexpected_request","reason":"unexpected request"},"status":400}`, http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)
	adapter, err := NewAdapter(&vectordb.Config{
		Type: vectordb.VectorDBTypeElasticsearchLocal, URL: server.URL,
		APIKey: "test-key", Timeout: 2, VectorDimensions: 3, SimilarityMetric: "cosine",
	})
	if err != nil {
		t.Fatalf("NewAdapter(): %v", err)
	}
	return adapter, recorder
}

func TestAdapterProtocolLifecycle(t *testing.T) {
	adapter, recorder := newElasticsearchProtocolAdapter(t)
	ctx := context.Background()
	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health(): %v", err)
	}
	if err := adapter.CreateCollection(ctx, "docs", &vectordb.CollectionSchema{}); err != nil {
		t.Fatalf("CreateCollection(): %v", err)
	}
	if err := adapter.CreateCollection(ctx, "unacknowledged", nil); err == nil {
		t.Fatal("CreateCollection() accepted an unacknowledged response")
	}
	if exists, err := adapter.CollectionExists(ctx, "docs"); err != nil || !exists {
		t.Fatalf("CollectionExists(docs) = %t, %v", exists, err)
	}
	if exists, err := adapter.CollectionExists(ctx, "missing"); err != nil || exists {
		t.Fatalf("CollectionExists(missing) = %t, %v", exists, err)
	}
	schema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || schema.Class != "docs" || len(schema.Properties) != 4 {
		t.Fatalf("GetSchema() = %#v, %v", schema, err)
	}
	if _, err := adapter.GetSchema(ctx, "missing"); err == nil {
		t.Fatal("GetSchema() accepted a missing collection")
	}
	if dimensions, err := adapter.getIndexDimensions(ctx, "docs"); err != nil || dimensions != 3 {
		t.Fatalf("getIndexDimensions() = %d, %v", dimensions, err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 1 || collections[0].Name != "docs" || collections[0].Count != 2 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	if count, err := adapter.GetCollectionCount(ctx, "docs"); err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}

	doc := &vectordb.Document{ID: "one", Text: "hello", Content: "world", Image: "image", ImageData: "data", URL: "https://example.test", Metadata: map[string]interface{}{"kind": "guide"}}
	if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument(): %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("CreateDocuments(empty): %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc, {ID: "two", Text: "two"}}); err != nil {
		t.Fatalf("CreateDocuments(): %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{{ID: "bad", Metadata: map[string]interface{}{"bad": make(chan int)}}}); err == nil {
		t.Fatal("CreateDocuments() accepted non-JSON metadata")
	}
	got, err := adapter.GetDocument(ctx, "docs", "one")
	if err != nil || got.ID != "one" || got.Image != "image" || got.ImageData != "data" || got.Metadata["kind"] != "guide" {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if _, err := adapter.GetDocument(ctx, "docs", "missing"); err == nil {
		t.Fatal("GetDocument() accepted a missing document")
	}
	if _, err := adapter.GetDocument(ctx, "docs", "malformed"); err == nil {
		t.Fatal("GetDocument() accepted malformed source")
	}
	if err := adapter.UpdateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("UpdateDocument(): %v", err)
	}
	if err := adapter.DeleteDocument(ctx, "docs", "one"); err != nil {
		t.Fatalf("DeleteDocument(): %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("DeleteDocuments(empty): %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{"one", "two"}); err != nil {
		t.Fatalf("DeleteDocuments(): %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata(): %v", err)
	}
	documents, err := adapter.ListDocuments(ctx, "docs", 0, 0)
	if err != nil || len(documents) != 1 || documents[0].ID != "one" {
		t.Fatalf("ListDocuments() = %#v, %v", documents, err)
	}
	if err := adapter.UpdateSchema(ctx, "docs", nil); err == nil {
		t.Fatal("UpdateSchema() accepted a nil schema")
	}
	if err := adapter.UpdateSchema(ctx, "docs", &vectordb.CollectionSchema{Class: "docs"}); err == nil || !strings.Contains(err.Error(), "reindexing") {
		t.Fatalf("UpdateSchema(valid) error = %v", err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection(): %v", err)
	}
	if err := adapter.DeleteCollection(ctx, "unacknowledged"); err == nil {
		t.Fatal("DeleteCollection() accepted an unacknowledged response")
	}
	if err := adapter.Close(ctx); err != nil {
		t.Fatalf("Close(): %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if len(recorder.paths) < 20 {
		t.Fatalf("recorded only %d protocol requests: %#v", len(recorder.paths), recorder.paths)
	}
}

func TestAdapterProtocolSearch(t *testing.T) {
	adapter, recorder := newElasticsearchProtocolAdapter(t)
	ctx := context.Background()
	for _, test := range []struct {
		name string
		run  func() ([]*vectordb.QueryResult, error)
	}{
		{"bm25 defaults", func() ([]*vectordb.QueryResult, error) { return adapter.SearchBM25(ctx, "docs", "hello", nil) }},
		{"bm25 options", func() ([]*vectordb.QueryResult, error) {
			return adapter.SearchBM25(ctx, "docs", "hello", &vectordb.QueryOptions{TopK: 2})
		}},
		{"metadata defaults", func() ([]*vectordb.QueryResult, error) {
			return adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, nil)
		}},
		{"metadata options", func() ([]*vectordb.QueryResult, error) {
			return adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, &vectordb.QueryOptions{TopK: 1})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			results, err := test.run()
			if err != nil || len(results) != 1 || results[0].Document.ID != "one" || results[0].Score < 0.89 {
				t.Fatalf("search results = %#v, %v", results, err)
			}
		})
	}
	if _, err := adapter.SearchSemantic(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchSemantic() succeeded without an LLM client")
	}
	if _, err := adapter.SearchHybrid(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchHybrid() succeeded without an LLM client")
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if !strings.Contains(strings.Join(recorder.bodies, "\n"), "multi_match") {
		t.Fatalf("search bodies = %#v", recorder.bodies)
	}
}

func TestSimilarityMetricFallback(t *testing.T) {
	for _, metric := range []string{"cosine", "dot_product", "l2_norm", "unknown"} {
		if got := getSimilarityMetric(metric); got == nil {
			t.Fatalf("getSimilarityMetric(%q) returned nil", metric)
		}
	}
}
