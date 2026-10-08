// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package qdrant

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	qdrant "github.com/qdrant/go-client/qdrant"
)

type qdrantAdapterRoundTripFunc func(*http.Request) (*http.Response, error)

func (f qdrantAdapterRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAdapterFixtureOperationPaths(t *testing.T) {
	state := &qdrantFixtureState{
		getResult:    []*qdrant.RetrievedPoint{fixturePoint()},
		scrollResult: []*qdrant.RetrievedPoint{fixturePoint()},
		searchResult: []*qdrant.ScoredPoint{fixtureScoredPoint()},
	}
	client := newQdrantFixtureClient(state)
	client.config.VectorDimensions = 3
	client.config.SimilarityMetric = "Cosine"
	adapter := &Adapter{Client: client}
	ctx := context.Background()

	if err := adapter.CreateCollection(ctx, "docs", nil); err != nil {
		t.Fatal(err)
	}
	if collections, err := adapter.ListCollections(ctx); err != nil || len(collections) != 2 {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	doc := &vectordb.Document{ID: "doc-1", Text: "hello", Metadata: map[string]interface{}{"kind": "guide"}}
	if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatal(err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc}); err != nil {
		t.Fatal(err)
	}
	if got, err := adapter.GetDocument(ctx, "docs", "doc-1"); err != nil || got.ID != "doc-1" {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := adapter.UpdateDocument(ctx, "docs", doc); err != nil {
		t.Fatal(err)
	}
	if docs, err := adapter.ListDocuments(ctx, "docs", 10, 4); err != nil || len(docs) != 1 {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.SearchSemantic(ctx, "docs", "hello", &vectordb.QueryOptions{TopK: 2}); err == nil {
		t.Fatal("SearchSemantic unexpectedly succeeded without an embedding client")
	}
	if _, err := adapter.SearchHybrid(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchHybrid unexpectedly succeeded without an embedding client")
	}
	if _, err := adapter.SearchBM25(ctx, "docs", "hello", nil); err == nil {
		t.Fatal("SearchBM25 unexpectedly succeeded")
	}
	if _, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, &vectordb.QueryOptions{TopK: 2}); err != nil {
		t.Fatal(err)
	}
	if schema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, "docs"); schema.Class != "docs" {
		t.Fatalf("GetDefaultSchema() = %#v", schema)
	}
	if err := adapter.ValidateSchema(nil); err != nil {
		t.Fatal(err)
	}
	if err := adapter.UpdateSchema(ctx, "docs", nil); err == nil {
		t.Fatal("UpdateSchema unexpectedly succeeded")
	}
	embeddingClient, err := llm.NewOpenAIClientWithHTTP("fixture", &http.Client{Transport: qdrantAdapterRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[{"object":"embedding","embedding":[0.1,0.2,0.3],"index":0}],"model":"fixture","usage":{"prompt_tokens":1,"total_tokens":1}}`))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	adapter.llmClient = embeddingClient
	if results, err := adapter.SearchSemantic(ctx, "docs", "hello", &vectordb.QueryOptions{TopK: 2}); err != nil || len(results) != 1 {
		t.Fatalf("SearchSemantic() = %#v, %v", results, err)
	}
	if err := adapter.CreateDocument(ctx, "docs", &vectordb.Document{ID: "embedded", Text: "hello"}); err != nil {
		t.Fatal(err)
	}
}
