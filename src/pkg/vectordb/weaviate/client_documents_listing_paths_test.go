package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListDocumentsGraphQLPropertiesPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/schema/Docs" {
			_, _ = io.WriteString(w, `{"properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}`)
			return
		}
		if r.URL.Path == "/v1/graphql" {
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "Get") {
				_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1"},"content":"hello","metadata":"{\"author\":\"Ada\"}"}]}}}`)
				return
			}
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs, err := client.ListDocuments(context.Background(), "Docs", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].ID != "doc-1" {
		t.Fatalf("unexpected documents: %#v", docs)
	}
}
