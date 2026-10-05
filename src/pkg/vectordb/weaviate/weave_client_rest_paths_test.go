package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeaveClientCreateAndDeleteRESTPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/objects" {
			w.WriteHeader(http.StatusCreated)
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/v1/objects/doc-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, APIKey: "key", OpenAIAPIKey: "openai", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	doc := Document{ID: "doc-1", Text: "text", Content: "content", Metadata: map[string]interface{}{"kind": "pdf"}}
	if err := client.CreateDocument(context.Background(), "Docs", doc); err != nil {
		t.Fatal(err)
	}
	if err := client.deleteObjectViaREST(context.Background(), "doc-1"); err != nil {
		t.Fatal(err)
	}
}

func TestWeaveClientRESTErrorPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateDocument(context.Background(), "Docs", Document{Metadata: map[string]interface{}{}}); err == nil {
		t.Fatal("expected create error")
	}
	if err := client.deleteObjectViaREST(context.Background(), "doc-1"); err == nil {
		t.Fatal("expected delete error")
	}
}

func TestWeaveClientDeleteCollectionGraphQLPaths(t *testing.T) {
	responses := []string{
		`{"data":{"delete":{"Docs":{"successful":2}}}}`,
		`{"data":{"delete":{"Docs":{"successful":0}}}}`,
		`{"errors":[{"message":"delete failed"}]}`,
		`{`,
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
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(responses); i++ {
		_ = client.deleteCollectionViaGraphQL(context.Background(), "Docs")
	}
}

func TestWeaveClientDeleteCollectionRESTPaths(t *testing.T) {
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
	objects, err := client.getAllObjectsInCollection(context.Background(), "Docs")
	if err != nil || len(objects) != 2 {
		t.Fatalf("objects = (%#v, %v)", objects, err)
	}
	if err := client.deleteCollectionViaREST(context.Background(), "Docs"); err != nil {
		t.Fatal(err)
	}
}

func TestWeaveClientDeleteCollectionRESTEmptyAndQueryErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	t.Cleanup(server.Close)
	client, err := NewWeaveClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if objects, err := client.getAllObjectsInCollection(context.Background(), "Docs"); err != nil || len(objects) != 0 {
		t.Fatalf("empty objects = (%#v, %v)", objects, err)
	}
	if err := client.deleteCollectionViaREST(context.Background(), "Docs"); err != nil {
		t.Fatal(err)
	}
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadGateway)
	}))
	t.Cleanup(badServer.Close)
	badClient, err := NewWeaveClient(&Config{URL: badServer.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := badClient.getAllObjectsInCollection(context.Background(), "Docs"); err == nil {
		t.Fatal("expected object query error")
	}
	if err := badClient.deleteCollectionViaREST(context.Background(), "Docs"); err == nil {
		t.Fatal("expected collection deletion error")
	}
}
