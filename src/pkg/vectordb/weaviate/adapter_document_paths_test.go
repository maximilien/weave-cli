package weaviate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestAdapterDocumentOperationErrorPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	weaveClient, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	adapter := &Adapter{client: client, weaveClient: weaveClient}
	ctx := context.Background()
	doc := &vectordb.Document{ID: "doc-1", Text: "text", Content: "content", Metadata: map[string]interface{}{"kind": "pdf"}}
	if err := adapter.CreateDocument(ctx, "Docs", doc); err == nil {
		t.Fatal("expected create document error")
	}
	if err := adapter.UpdateDocument(ctx, "Docs", doc); err == nil {
		t.Fatal("expected update document error")
	}
	if _, err := adapter.GetDocument(ctx, "Docs", "doc-1"); err == nil {
		t.Fatal("expected get document error")
	}
	if err := adapter.DeleteDocument(ctx, "Docs", "doc-1"); err == nil {
		t.Fatal("expected delete document error")
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "Docs", map[string]interface{}{"kind": "pdf"}); err == nil {
		t.Fatal("expected metadata delete error")
	}
	if _, err := adapter.ListDocuments(ctx, "Docs", 2, 0); err == nil {
		t.Fatal("expected list document error")
	}
	if err := adapter.CreateDocuments(ctx, "Docs", nil); err != nil {
		t.Fatal(err)
	}
	if err := adapter.DeleteDocuments(ctx, "Docs", nil); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterCollectionOperationErrorPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	weaveClient, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	adapter := &Adapter{client: client, weaveClient: weaveClient}
	ctx := context.Background()
	schema := &vectordb.CollectionSchema{Class: "Docs", Properties: []vectordb.SchemaProperty{{Name: "text", DataType: []string{"text"}}}}
	if err := adapter.CreateCollection(ctx, "Docs", schema); err == nil {
		t.Fatal("expected create collection error")
	}
	if err := adapter.DeleteCollection(ctx, "Docs"); err == nil {
		t.Fatal("expected delete collection error")
	}
	if _, err := adapter.ListCollections(ctx); err == nil {
		t.Fatal("expected list collections error")
	}
	if _, err := adapter.CollectionExists(ctx, "Docs"); err == nil {
		t.Fatal("expected collection exists error")
	}
	if _, err := adapter.GetCollectionCount(ctx, "Docs"); err == nil {
		t.Fatal("expected collection count error")
	}
	if _, err := adapter.GetSchema(ctx, "Docs"); err == nil {
		t.Fatal("expected schema error")
	}
}
