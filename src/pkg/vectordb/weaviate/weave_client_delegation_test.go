package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
