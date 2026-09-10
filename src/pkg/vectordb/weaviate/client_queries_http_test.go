// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package weaviate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type queryRecorder struct {
	mu      sync.Mutex
	queries []string
}

func (r *queryRecorder) add(query string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queries = append(r.queries, query)
}

func (r *queryRecorder) contains(fragment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, query := range r.queries {
		if strings.Contains(query, fragment) {
			return true
		}
	}
	return false
}

func newQueryProtocolServer(t *testing.T) (*httptest.Server, *queryRecorder) {
	t.Helper()
	recorder := &queryRecorder{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/meta":
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
			return
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[
				{"class":"Docs","properties":[{"name":"content"},{"name":"text"},{"name":"metadata"}]},
				{"class":"ProductImages","properties":[{"name":"url"},{"name":"image"},{"name":"image_data"},{"name":"metadata"}]},
				{"class":"ErrorImages","properties":[{"name":"url"},{"name":"image_data"}]},
				{"class":"FallbackDocs","properties":[{"name":"content"},{"name":"text"},{"name":"metadata"}]},
				{"class":"SimpleDocs","properties":[{"name":"content"},{"name":"text"},{"name":"metadata"},{"name":"url"}]},
				{"class":"NoFields","properties":[{"name":"metadata"}]},
				{"class":"NearFailure","properties":[{"name":"content"}]},
				{"class":"HybridFailure","properties":[{"name":"content"}]},
				{"class":"SimpleFailure","properties":[{"name":"content"}]}
			]}`)
			return
		case "/v1/graphql":
			var payload struct {
				Query string `json:"query"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode GraphQL request: %v", err)
			}
			recorder.add(payload.Query)
			query := payload.Query

			switch {
			case strings.Contains(query, "NearFailure"),
				strings.Contains(query, "HybridFailure"),
				strings.Contains(query, "SimpleFailure"):
				http.Error(w, "query failed", http.StatusBadGateway)
			case strings.Contains(query, "ErrorImages"):
				_, _ = io.WriteString(w, `{"errors":[{"message":"near image unavailable"}]}`)
			case strings.Contains(query, "FallbackDocs") && strings.Contains(query, "nearText"):
				_, _ = io.WriteString(w, `{"errors":[{"message":"nearText unavailable"}]}`)
			case strings.Contains(query, "FallbackDocs") && strings.Contains(query, "hybrid"):
				writeQueryData(w, "FallbackDocs", "content", "hybrid result", "score", 0.8)
			case strings.Contains(query, "SimpleDocs") && (strings.Contains(query, "nearText") || strings.Contains(query, "hybrid")):
				_, _ = io.WriteString(w, `{"errors":[{"message":"operator unavailable"}]}`)
			case strings.Contains(query, "SimpleDocs"):
				writeQueryData(w, "SimpleDocs", "content", "simple result", "certainty", 0.2)
			case strings.Contains(query, "ProductImages") && strings.Contains(query, "nearImage"):
				_, _ = io.WriteString(w, `{"data":{"Get":{"ProductImages":[{
					"url":"https://example.test/image.png","image_data":"encoded",
					"metadata":{"filename":"image.png"},
					"_additional":{"id":"image-1","certainty":0.9}
				}]}}}`)
			case strings.Contains(query, "ProductImages"):
				writeQueryData(w, "ProductImages", "url", "https://example.test/image.png", "distance", 0.2)
			case strings.Contains(query, "Docs") && strings.Contains(query, "bm25"):
				writeQueryData(w, "Docs", "content", "keyword result", "score", 0.7)
			case strings.Contains(query, "Docs"):
				writeQueryData(w, "Docs", "content", "semantic result", "distance", 0.4)
			default:
				_, _ = io.WriteString(w, `{"data":{"Get":{}}}`)
			}
			return
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server, recorder
}

func writeQueryData(w io.Writer, collection, field, content, scoreField string, score float64) {
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"Get": map[string]interface{}{
				collection: []interface{}{map[string]interface{}{
					field:      content,
					"metadata": map[string]interface{}{"source": "fake"},
					"_additional": map[string]interface{}{
						"id":       "result-1",
						scoreField: score,
					},
				}},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(response)
}

func newQueryProtocolClient(t *testing.T) (*Client, *queryRecorder) {
	t.Helper()
	server, recorder := newQueryProtocolServer(t)
	client, err := NewClient(&Config{URL: server.URL, APIKey: "api-key", Timeout: 1})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, recorder
}

func TestQueryProtocolsWithFakeHTTPServer(t *testing.T) {
	client, queries := newQueryProtocolClient(t)
	ctx := context.Background()

	results, err := client.Query(ctx, "Docs", `quoted "phrase"`, QueryOptions{Verbose: true})
	if err != nil || len(results) != 1 || results[0].Content != "semantic result" {
		t.Fatalf("Query(nearText) = (%#v, %v)", results, err)
	}
	if !queries.contains(`concepts: ["quoted \"phrase\""]`) || !queries.contains(`targetVectors: ["default"]`) || !queries.contains("limit: 5") {
		t.Fatal("nearText query did not include escaping, default vector, and default limit")
	}

	results, err = client.Query(ctx, "ProductImages", "", QueryOptions{
		TopK: 2, ImageQuery: "base64-image", IncludeImages: true, Verbose: true,
	})
	if err != nil || len(results) != 1 || results[0].Metadata["image_base64"] != "encoded" {
		t.Fatalf("Query(nearImage) = (%#v, %v)", results, err)
	}
	if !queries.contains("nearImage") || !queries.contains(`targetVectors: ["image_vector"]`) || !queries.contains("image_data") {
		t.Fatal("nearImage query did not include image vector and image data")
	}

	results, err = client.Query(ctx, "ProductImages", "camera", QueryOptions{UseImageVector: true})
	if err != nil || len(results) != 1 || results[0].Content != "https://example.test/image.png" {
		t.Fatalf("Query(image nearText) = (%#v, %v)", results, err)
	}

	results, err = client.Query(ctx, "Docs", `keyword "query"`, QueryOptions{
		TopK: 3, UseBM25: true, SearchMetadata: true, Verbose: true,
	})
	if err != nil || len(results) != 1 || results[0].Content != "keyword result" {
		t.Fatalf("Query(BM25) = (%#v, %v)", results, err)
	}
	if !queries.contains("bm25") || !queries.contains(`properties: ["content", "text", "metadata"]`) {
		t.Fatal("BM25 query did not include all searchable fields")
	}

	if _, err := client.Query(ctx, "NoFields", "query", QueryOptions{UseBM25: true}); err == nil || !strings.Contains(err.Error(), "no searchable fields") {
		t.Fatalf("Query(BM25 no fields) error = %v", err)
	}
	if _, err := client.Query(ctx, "Missing", "query", QueryOptions{}); err == nil || !strings.Contains(err.Error(), "not found in schema") {
		t.Fatalf("Query(missing schema) error = %v", err)
	}
}

func TestQueryFallbackProtocolsWithFakeHTTPServer(t *testing.T) {
	client, queries := newQueryProtocolClient(t)
	ctx := context.Background()

	results, err := client.Query(ctx, "FallbackDocs", `fallback "query"`, QueryOptions{TopK: 4, SearchMetadata: true, Verbose: true})
	if err != nil || len(results) != 1 || results[0].Content != "hybrid result" {
		t.Fatalf("Query(hybrid fallback) = (%#v, %v)", results, err)
	}
	if !queries.contains("hybrid") || !queries.contains(`targetVectors: ["default"]`) {
		t.Fatal("hybrid fallback query was not issued")
	}

	results, err = client.Query(ctx, "SimpleDocs", "simple", QueryOptions{SearchMetadata: true})
	if err != nil || len(results) != 1 || results[0].Score != 1 {
		t.Fatalf("Query(simple fallback) = (%#v, %v)", results, err)
	}
	if !queries.contains("operator: Or") || !queries.contains(`path: ["metadata"]`) || !queries.contains(`path: ["url"]`) {
		t.Fatal("simple fallback query did not include expected operands")
	}

	if _, err := client.queryWithFallback(ctx, "NoFields", "query", QueryOptions{}, "content"); err == nil || !strings.Contains(err.Error(), "no searchable fields") {
		t.Fatalf("queryWithFallback(no fields) error = %v", err)
	}
	if _, err := client.queryWithSimpleFallback(ctx, "NoFields", "query", QueryOptions{}, "content"); err == nil || !strings.Contains(err.Error(), "no searchable fields") {
		t.Fatalf("queryWithSimpleFallback(no fields) error = %v", err)
	}
	if _, err := client.Query(ctx, "NearFailure", "query", QueryOptions{}); err == nil || !strings.Contains(err.Error(), "failed to execute semantic search") {
		t.Fatalf("Query(transport failure) error = %v", err)
	}
	if _, err := client.Query(ctx, "ErrorImages", "query", QueryOptions{ImageQuery: "encoded"}); err == nil || !strings.Contains(err.Error(), "nearImage query returned errors") {
		t.Fatalf("Query(nearImage errors) error = %v", err)
	}
	if _, err := client.queryWithFallback(ctx, "HybridFailure", "query", QueryOptions{}, "content"); err == nil || !strings.Contains(err.Error(), "failed to execute hybrid fallback") {
		t.Fatalf("queryWithFallback(transport failure) error = %v", err)
	}
	if _, err := client.queryWithSimpleFallback(ctx, "SimpleFailure", "query", QueryOptions{}, "content"); err == nil || !strings.Contains(err.Error(), "failed to execute simple fallback") {
		t.Fatalf("queryWithSimpleFallback(transport failure) error = %v", err)
	}
}

func TestQueryWithFiltersProtocol(t *testing.T) {
	client, queries := newQueryProtocolClient(t)
	ctx := context.Background()

	results, err := client.QueryWithFilters(ctx, "Docs", `filtered "query"`, QueryOptions{Verbose: true}, map[string]interface{}{
		"tenant": "one", "year": 2026,
	})
	if err != nil || len(results) != 1 || results[0].Content != "semantic result" {
		t.Fatalf("QueryWithFilters() = (%#v, %v)", results, err)
	}
	if !queries.contains("where:") || !queries.contains(`equal: "one"`) || !queries.contains(`equal: "2026"`) || !queries.contains("limit: 5") {
		t.Fatal("filtered query did not include filters and default limit")
	}

	results, err = client.QueryWithFilters(ctx, "ProductImages", "image", QueryOptions{TopK: 1}, nil)
	if err != nil || len(results) != 1 {
		t.Fatalf("QueryWithFilters(image) = (%#v, %v)", results, err)
	}
	if !queries.contains(`targetVectors: ["text_vector"]`) {
		t.Fatal("image filtered query did not select text_vector")
	}

	if _, err := client.QueryWithFilters(ctx, "Missing", "query", QueryOptions{}, nil); err == nil || !strings.Contains(err.Error(), "failed to get collection schema") {
		t.Fatalf("QueryWithFilters(missing) error = %v", err)
	}
}
