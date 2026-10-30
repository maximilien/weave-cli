// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestAnalyzeJSONFieldsMergesDocumentSchemas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/meta":
			_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
		case "/v1/objects/Docs/first":
			_, _ = io.WriteString(w, `{"id":"first","properties":{"details":"{\"name\":\"Ada\",\"score\":2}","plain":"not JSON"}}`)
		case "/v1/objects/Docs/second":
			_, _ = io.WriteString(w, `{"id":"second","properties":{"details":"{\"score\":3,\"enabled\":true}"}}`)
		case "/v1/objects/Docs/missing":
			http.NotFound(w, r)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := mustWeaviateClient(t, server.URL)
	docs := []weaviate.Document{{ID: "first"}, {ID: "missing"}, {ID: "second"}}
	got := analyzeJSONFields(docs, context.Background(), client, "Docs")
	want := map[string]map[string]interface{}{"details": {"name": "string", "score": "integer", "enabled": "boolean"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("schema = %#v; want %#v", got, want)
	}
}
