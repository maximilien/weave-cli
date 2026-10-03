package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateDocumentMergePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && (r.URL.Path == "/v1/objects/Docs/doc-1" || r.URL.Path == "/v1/objects/NoSchema/doc-1"):
			_, _ = io.WriteString(w, `{"id":"doc-1","properties":{"text":"old","content":"old","metadata":"{\"author\":\"Ada\"}"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/schema/Docs":
			_, _ = io.WriteString(w, `{"properties":[{"name":"text","dataType":["text"]},{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/objects/TextOnly/doc-1":
			_, _ = io.WriteString(w, `{"id":"doc-1","properties":{"text":"old"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/schema/TextOnly":
			_, _ = io.WriteString(w, `{"properties":[{"name":"text","dataType":["text"]}]}`)
		case (r.Method == http.MethodPatch || r.Method == http.MethodPut) && (r.URL.Path == "/v1/objects/Docs/doc-1" || r.URL.Path == "/v1/objects/NoSchema/doc-1" || r.URL.Path == "/v1/objects/TextOnly/doc-1" || r.URL.Path == "/v1/objects/doc-1"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "Docs", "doc-1", "new content", map[string]interface{}{"status": "published"}); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "Docs", "doc-1", "", map[string]interface{}{"_update_text": "text update", "_update_content": "content update"}); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "Docs", "doc-1", "", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "Docs", "doc-1", "", map[string]interface{}{"_update_text": 123, "_update_content": ""}); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "NoSchema", "doc-1", "new", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "TextOnly", "doc-1", "new", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateDocument(context.Background(), "Docs", "missing", "new", nil); err == nil {
		t.Fatal("expected missing document update error")
	}
	if err := client.UpdateDocument(context.Background(), "Missing", "doc-1", "new", nil); err == nil {
		t.Fatal("expected missing collection update error")
	}
}
