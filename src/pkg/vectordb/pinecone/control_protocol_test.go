// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pinecone

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	pcsdk "github.com/pinecone-io/go-pinecone/pinecone"
)

type pineconeRoundTripFunc func(*http.Request) (*http.Response, error)

func (f pineconeRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func pineconeResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func newPineconeControlAdapter(t *testing.T, transport http.RoundTripper) *Adapter {
	t.Helper()
	client, err := pcsdk.NewClient(pcsdk.NewClientParams{
		ApiKey:     "fixture-key",
		Host:       "https://control.fixture",
		RestClient: &http.Client{Transport: transport},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return &Adapter{
		client: client,
		config: &vectordb.Config{
			Type:             vectordb.VectorDBTypePinecone,
			Timeout:          1,
			VectorDimensions: 384,
			SimilarityMetric: "euclidean",
		},
		apiKey: "fixture-key",
	}
}

func TestPineconeControlPlaneOperations(t *testing.T) {
	var requests []*http.Request
	transport := pineconeRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req)
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/indexes":
			return pineconeResponse(http.StatusOK, `{"indexes":[{"name":"docs","dimension":384,"host":"","metric":"euclidean"},{"name":"notes","dimension":768,"host":"","metric":"cosine"}]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/indexes":
			return pineconeResponse(http.StatusCreated, `{"name":"docs","dimension":384,"host":"","metric":"euclidean"}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/indexes/docs":
			return pineconeResponse(http.StatusOK, `{"name":"docs","dimension":384,"host":"","metric":"euclidean"}`), nil
		case req.Method == http.MethodDelete && req.URL.Path == "/indexes/docs":
			return pineconeResponse(http.StatusAccepted, ``), nil
		default:
			return pineconeResponse(http.StatusNotFound, `{"message":"index not found"}`), nil
		}
	})
	adapter := newPineconeControlAdapter(t, transport)
	ctx := context.Background()

	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	schema := &vectordb.CollectionSchema{Class: "docs", Properties: []vectordb.SchemaProperty{{Name: "vector", DataType: []string{"number[]"}}}}
	if err := adapter.CreateCollection(ctx, "docs", schema); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	exists, err := adapter.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 2 || collections[1].Name != "notes" {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	info, err := adapter.GetCollectionInfo(ctx, "docs")
	if err != nil || info.Name != "docs" || info.Count != 0 {
		t.Fatalf("GetCollectionInfo() = %#v, %v", info, err)
	}
	gotSchema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || gotSchema.Class != "docs" || gotSchema.Vectorizer != "sentence-transformers/all-MiniLM-L6-v2" {
		t.Fatalf("GetSchema() = %#v, %v", gotSchema, err)
	}
	if len(requests) < 8 || requests[0].Header.Get("Api-Key") != "fixture-key" {
		t.Fatalf("requests = %d, API key = %q", len(requests), requests[0].Header.Get("Api-Key"))
	}
}

func TestPineconeControlPlaneNotFoundAndFallbacks(t *testing.T) {
	transport := pineconeRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/indexes/missing" {
			return pineconeResponse(http.StatusNotFound, `{"message":"index not found"}`), nil
		}
		return pineconeResponse(http.StatusInternalServerError, `{"message":"fixture failure"}`), nil
	})
	adapter := newPineconeControlAdapter(t, transport)
	ctx := context.Background()

	exists, err := adapter.CollectionExists(ctx, "missing")
	if err == nil || exists {
		t.Fatalf("CollectionExists(missing) = %v, %v", exists, err)
	}
	if dims, err := adapter.getIndexDimensions(ctx, "missing"); err != nil || dims != 384 {
		t.Fatalf("getIndexDimensions() = %d, %v", dims, err)
	}
	adapter.config.VectorDimensions = 0
	if dims, err := adapter.getIndexDimensions(ctx, "missing"); err != nil || dims != 1536 {
		t.Fatalf("getIndexDimensions(default) = %d, %v", dims, err)
	}
	if _, err := adapter.GetCollectionCount(ctx, "missing"); err == nil {
		t.Fatal("GetCollectionCount() expected error")
	}
	if _, err := adapter.GetCollectionInfo(ctx, "missing"); err == nil {
		t.Fatal("GetCollectionInfo() expected error")
	}
	if _, err := adapter.GetSchema(ctx, "missing"); err != nil {
		t.Fatalf("GetSchema() should use dimension fallback: %v", err)
	}
}

func TestPineconeControlPlaneErrors(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("fixture transport failure")
	adapter := newPineconeControlAdapter(t, pineconeRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, boom
	}))
	schema := &vectordb.CollectionSchema{Class: "docs", Properties: []vectordb.SchemaProperty{{Name: "vector"}}}

	for name, call := range map[string]func() error{
		"health": func() error { return adapter.Health(ctx) },
		"create": func() error { return adapter.CreateCollection(ctx, "docs", schema) },
		"delete": func() error { return adapter.DeleteCollection(ctx, "docs") },
		"exists": func() error { _, err := adapter.CollectionExists(ctx, "docs"); return err },
		"list":   func() error { _, err := adapter.ListCollections(ctx); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	for _, message := range []string{"401 Unauthorized", "deadline timeout", "plain failure"} {
		adapter := newPineconeControlAdapter(t, pineconeRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New(message)
		}))
		if err := adapter.Health(ctx); err == nil || !strings.Contains(err.Error(), message) {
			t.Fatalf("Health(%q) error = %v", message, err)
		}
	}
}

func TestPineconeNilClientAndSchemaHelpers(t *testing.T) {
	adapter := &Adapter{config: &vectordb.Config{}}
	ctx := context.Background()
	schema := &vectordb.CollectionSchema{Class: "docs", Properties: []vectordb.SchemaProperty{{Name: "vector"}}}

	for name, call := range map[string]func() error{
		"health":     func() error { return adapter.Health(ctx) },
		"create":     func() error { return adapter.CreateCollection(ctx, "docs", schema) },
		"delete":     func() error { return adapter.DeleteCollection(ctx, "docs") },
		"exists":     func() error { _, err := adapter.CollectionExists(ctx, "docs"); return err },
		"list":       func() error { _, err := adapter.ListCollections(ctx); return err },
		"count":      func() error { _, err := adapter.GetCollectionCount(ctx, "docs"); return err },
		"info":       func() error { _, err := adapter.GetCollectionInfo(ctx, "docs"); return err },
		"schema":     func() error { _, err := adapter.GetSchema(ctx, "docs"); return err },
		"create doc": func() error { return adapter.CreateDocument(ctx, "docs", &vectordb.Document{}) },
		"get doc":    func() error { _, err := adapter.GetDocument(ctx, "docs", "id"); return err },
		"delete doc": func() error { return adapter.DeleteDocument(ctx, "docs", "id") },
		"list docs":  func() error { _, err := adapter.ListDocuments(ctx, "docs", 10, 0); return err },
		"semantic":   func() error { _, err := adapter.SearchSemantic(ctx, "docs", "query", nil); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected nil-client error")
			}
		})
	}

	for dims, want := range map[int]string{
		768:  "sentence-transformers/all-mpnet-base-v2",
		384:  "sentence-transformers/all-MiniLM-L6-v2",
		1536: "text-embedding-3-small",
		3072: "text-embedding-3-large",
		1024: "nomic-embed-text",
		999:  "text-embedding-3-small",
	} {
		if got := inferEmbeddingModelFromDimensions(dims); got != want {
			t.Errorf("inferEmbeddingModelFromDimensions(%d) = %q, want %q", dims, got, want)
		}
	}
	if err := adapter.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := adapter.SearchBM25(ctx, "docs", "query", nil); err == nil {
		t.Fatal("SearchBM25() expected unsupported error")
	}
}
