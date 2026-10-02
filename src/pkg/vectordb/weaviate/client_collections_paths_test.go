package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateCollectionRESTSchemaPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/schema" {
			_, _ = io.WriteString(w, `{"classes":[]}`)
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/v1/schema" {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"class":"Docs"`) && !strings.Contains(string(body), `"class":"Images"`) && !strings.Contains(string(body), `"class":"NoVector"`) {
				t.Errorf("unexpected schema body: %s", body)
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, APIKey: "key", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := client.CreateCollection(ctx, "Docs", "text-embedding-3-small", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.CreateCollectionWithSchema(ctx, "Images", "text-embedding-3-small", []FieldDefinition{{Name: "caption", Type: "text"}}, "image"); err != nil {
		t.Fatal(err)
	}
	if err := client.CreateCollectionWithSchema(ctx, "NoVector", "none", []FieldDefinition{{Name: "title", Type: "text"}, {Name: "count", Type: "int"}}, "text"); err != nil {
		t.Fatal(err)
	}
	duplicateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"classes":[{"class":"Docs"}]}`)
	}))
	t.Cleanup(duplicateServer.Close)
	duplicateClient, err := NewClient(&Config{URL: duplicateServer.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := duplicateClient.CreateCollection(ctx, "Docs", "none", nil); err == nil {
		t.Fatal("expected duplicate collection error")
	}
	if count, err := client.DeleteDocumentsByMetadata(ctx, "Docs", []string{"invalid-filter"}); err == nil || count != 0 {
		t.Fatalf("expected invalid metadata filter, got count=%d err=%v", count, err)
	}
}
