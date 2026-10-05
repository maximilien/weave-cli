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

func TestMetadataHydrationAndDeletionFlows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/objects/Docs/doc-1" {
			_, _ = io.WriteString(w, `{"id":"doc-1","properties":{"content":"hello","metadata":"{\"kind\":\"pdf\"}"}}`)
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/v1/objects/Docs/doc-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	docs, err := client.GetDocumentsByMetadata(ctx, "Docs", []string{"kind=pdf"})
	if err != nil || len(docs) != 1 || docs[0].Content != "hello" {
		t.Fatalf("hydrated docs=%#v err=%v", docs, err)
	}
	if count, err := client.DeleteDocumentsByMetadata(ctx, "Docs", []string{"kind=pdf"}); err != nil || count != 1 {
		t.Fatalf("deleted count=%d err=%v", count, err)
	}
}

func TestEmptyCollectionAndMissingDocumentPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/schema/Empty" {
			_, _ = io.WriteString(w, `{"properties":[{"name":"content","dataType":["text"]}]}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Empty":[]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := client.DeleteAllDocuments(ctx, "Empty"); err != nil {
		t.Fatalf("empty delete: %v", err)
	}
	if _, err := client.getDocumentSimple(ctx, "Empty", "missing"); err == nil {
		t.Fatal("expected missing document error")
	}
}

func TestDeleteAllDocumentsNonEmptyCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/v1/schema/Docs" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}},{"_additional":{"id":"doc-2"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAllDocuments(context.Background(), "Docs"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAllDocumentsFailureBranches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			http.Error(w, "failed", http.StatusBadGateway)
			return
		}
		if r.URL.Path == "/v1/schema/Docs" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"errors":[{"message":"class Docs not found"}]}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAllDocuments(context.Background(), "Docs"); err == nil {
		t.Fatal("expected list failure")
	}
}

func TestDeleteAllDocumentsPartialDeletionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			http.Error(w, "failed", http.StatusBadGateway)
			return
		}
		if r.URL.Path == "/v1/schema/Docs" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"}}]}}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAllDocuments(context.Background(), "Docs"); err == nil {
		t.Fatal("expected partial deletion failure")
	}
}

func TestListDocumentsSimpleGraphQLErrorMessages(t *testing.T) {
	responses := []string{
		`{"errors":[{"message":"class Docs not found"}]}`,
		`{"errors":[{"message":"Unknown class Docs"}]}`,
		`{"errors":[{"message":"Did you mean Other?"}]}`,
		`{"errors":[{"message":"generic failure"}]}`,
		`{"data":{"Get":{"Docs":[{"_additional":{}},"bad"]}}}`,
		`{"data":{"Get":{}}}`,
		`{"data":{}}`,
	}
	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := responses[len(responses)-1]
		if call < len(responses) {
			response = responses[call]
		}
		call++
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(responses); i++ {
		_, _ = client.listDocumentsSimple(context.Background(), "Docs", 2)
	}
}
