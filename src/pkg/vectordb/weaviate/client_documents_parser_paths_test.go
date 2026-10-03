package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDocumentParserPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"},"content":"hello","metadata":"{\"author\":\"Ada\"}"}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	docs, err := client.listDocumentsWithSimpleMetadata(ctx, "Docs", 5, []string{"content", "metadata"}, map[string]bool{})
	if err != nil || len(docs) != 1 || docs[0].ID != "doc-1" {
		t.Fatalf("simple metadata docs=%#v err=%v", docs, err)
	}
	if docs[0].Metadata["author"] != "Ada" {
		t.Fatalf("metadata=%#v", docs[0].Metadata)
	}
	simple, err := client.listDocumentsSimple(ctx, "Docs", 5)
	if err != nil || len(simple) != 1 || simple[0].ID != "doc-1" {
		t.Fatalf("simple docs=%#v err=%v", simple, err)
	}
	doc, err := client.getDocumentSimple(ctx, "Docs", "doc-1")
	if err != nil || doc.ID != "doc-1" || doc.Content != "Document ID: doc-1" {
		t.Fatalf("simple document=%#v err=%v", doc, err)
	}
}

func TestMetadataQueryFilterBranches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}},{"_additional":{"id":"doc-2"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs, err := client.queryDocumentsByMetadata(context.Background(), "Docs", map[string]string{
		"filename": "a.pdf", "original_filename": "source.pdf", "url": "example", "kind": "pdf",
	})
	if err != nil || len(docs) != 2 || docs[0].ID != "doc-1" {
		t.Fatalf("metadata query=%#v err=%v", docs, err)
	}
	if _, err := client.GetDocumentsByMetadata(context.Background(), "Docs", []string{"invalid"}); err == nil {
		t.Fatal("expected invalid metadata filter")
	}
}
