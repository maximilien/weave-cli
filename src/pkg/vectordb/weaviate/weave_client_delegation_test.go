package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeaveClientDelegationCancellationPaths(t *testing.T) {
	client, err := NewWeaveClient(&Config{URL: "http://127.0.0.1:1", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.ListCollections(ctx); err == nil {
		t.Fatal("expected canceled list error")
	}
	if _, err := client.ListDocuments(ctx, "Docs", 1); err == nil {
		t.Fatal("expected canceled documents error")
	}
	if _, err := client.CountDocuments(ctx, "Docs"); err == nil {
		t.Fatal("expected canceled count error")
	}
	if _, err := client.GetDocument(ctx, "Docs", "id"); err == nil {
		t.Fatal("expected canceled get error")
	}
	if err := client.Health(ctx); err == nil {
		t.Fatal("expected canceled health error")
	}
	if _, err := client.GetCollectionSchema(ctx, "Docs"); err == nil {
		t.Fatal("expected canceled schema error")
	}
}

func TestWeaveClientMetadataQueryFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	docs, err := client.queryDocumentsByMetadata(context.Background(), "Docs", map[string]string{"filename": "a.pdf", "original_filename": "source.pdf", "url": "example", "kind": "pdf"})
	if err != nil || len(docs) != 1 || docs[0].ID != "doc-1" {
		t.Fatalf("metadata query = (%#v, %v)", docs, err)
	}
}

func TestWeaveClientMetadataQueryErrorResponses(t *testing.T) {
	responses := []string{"status", "malformed", "graphql", "shape"}
	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := "shape"
		if call < len(responses) {
			kind = responses[call]
		}
		call++
		switch kind {
		case "status":
			http.Error(w, "backend failed", http.StatusBadGateway)
		case "malformed":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{`)
		case "graphql":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"errors":[{"message":"bad filter"}]}`)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{}},"bad"]}}}`)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		_, _ = client.queryDocumentsByMetadata(context.Background(), "Docs", map[string]string{"kind": "x"})
	}
	if docs, err := client.queryDocumentsByMetadata(context.Background(), "Docs", map[string]string{"kind": "x"}); err != nil || len(docs) != 0 {
		t.Fatalf("shape response = (%#v, %v)", docs, err)
	}
	if _, err := client.GetDocumentsByMetadata(context.Background(), "Docs", []string{"invalid"}); err == nil {
		t.Fatal("expected invalid metadata filter")
	}
}

func TestWeaveClientDeleteDocumentsByMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}},{"_additional":{"id":"doc-2"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if count, err := client.DeleteDocumentsByMetadata(context.Background(), "Docs", []string{"kind=pdf", "filename=a.pdf"}); err != nil || count != 2 {
		t.Fatalf("deleted count = (%d, %v)", count, err)
	}
	if count, err := client.DeleteDocumentsByMetadata(context.Background(), "Docs", nil); err != nil || count != 2 {
		t.Fatalf("empty-filter count = (%d, %v)", count, err)
	}
	if _, err := client.DeleteDocumentsByMetadata(context.Background(), "Docs", []string{"invalid"}); err == nil {
		t.Fatal("expected invalid filter error")
	}
}

func TestWeaveClientGetDocumentsByMetadataHydrates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/objects/Docs/doc-1" {
			_, _ = io.WriteString(w, `{"id":"doc-1","properties":{"content":"hello","metadata":"{\"kind\":\"pdf\"}"}}`)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/objects/") {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}},{"_additional":{"id":"missing"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs, err := client.GetDocumentsByMetadata(context.Background(), "Docs", []string{"kind=pdf"})
	if err != nil || len(docs) != 1 || docs[0].Content != "hello" {
		t.Fatalf("hydrated documents = (%#v, %v)", docs, err)
	}
}

func TestWeaveClientMetadataOperationQueryErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend failed", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDocumentsByMetadata(context.Background(), "Docs", []string{"kind=pdf"}); err == nil {
		t.Fatal("expected get metadata query error")
	}
	if _, err := client.DeleteDocumentsByMetadata(context.Background(), "Docs", []string{"kind=pdf"}); err == nil {
		t.Fatal("expected delete metadata query error")
	}
}
