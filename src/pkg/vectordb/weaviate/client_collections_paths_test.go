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
			if !strings.Contains(string(body), `"class":"Docs"`) && !strings.Contains(string(body), `"class":"Images"`) {
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
}
