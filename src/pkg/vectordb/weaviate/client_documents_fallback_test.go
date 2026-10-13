package weaviate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListDocumentsFallbackAggregationAndSimpleQueries(t *testing.T) {
	cases := []struct {
		name      string
		handler   func(http.ResponseWriter, *http.Request)
		wantEmpty bool
		wantError bool
	}{
		{
			name: "aggregation confirms empty",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/schema/Docs" {
					_, _ = io.WriteString(w, `{"properties":[{"name":"chunk_index","dataType":["int"]}]}`)
					return
				}
				_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":0}}]}}}`)
			},
			wantEmpty: true,
		},
		{
			name: "aggregation failure then simple empty",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/schema/Docs" {
					_, _ = io.WriteString(w, `{"properties":[{"name":"chunk_index","dataType":["int"]}]}`)
					return
				}
				body, _ := io.ReadAll(r.Body)
				if strings.Contains(string(body), "Aggregate") {
					_, _ = io.WriteString(w, `{"errors":[{"message":"aggregation unavailable"}]}`)
				} else {
					_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[]}}}`)
				}
			},
			wantEmpty: true,
		},
		{
			name: "all fallbacks fail",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/schema/Docs" {
					_, _ = io.WriteString(w, `{"properties":[{"name":"chunk_index","dataType":["int"]}]}`)
					return
				}
				_, _ = io.WriteString(w, `{"errors":[{"message":"not found"}]}`)
			},
			wantError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()
			client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
			if err != nil {
				t.Fatal(err)
			}
			docs, err := client.ListDocuments(context.Background(), "Docs", 5)
			if tc.wantError {
				if err == nil || docs != nil {
					t.Fatalf("expected fallback error, docs=%#v err=%v", docs, err)
				}
				return
			}
			if err != nil || !tc.wantEmpty || len(docs) != 0 {
				t.Fatalf("expected empty fallback result, docs=%#v err=%v", docs, err)
			}
		})
	}
}

func TestListDocumentsRichPropertiesAndEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/schema/Images" {
			_, _ = io.WriteString(w, `{"properties":[{"name":"text","dataType":["text"]},{"name":"metadata","dataType":["text"]},{"name":"image","dataType":["blob"]}]}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"Get":{"Images":[{"_additional":{"id":"img-1","vector":[1,2,3]},"text":"caption","metadata":"{\"camera\":\"A\"}"}]}}}`)
	}))
	defer server.Close()
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs, err := client.ListDocuments(context.Background(), "Images", 2)
	if err != nil || len(docs) != 1 {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
	if docs[0].ID != "img-1" || docs[0].Content != "caption" || len(docs[0].Embedding) != 3 || docs[0].Metadata["camera"] != "A" {
		t.Fatalf("unexpected rich document: %#v", docs[0])
	}
}
