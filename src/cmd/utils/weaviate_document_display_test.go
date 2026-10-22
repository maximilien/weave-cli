// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestWeaviateDocumentListAndShowWithFakeServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/meta":
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","vectorizer":"none","properties":[{"name":"content","dataType":["text"]}]}]}`)
		case "/v1/graphql":
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "Aggregate") {
				_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":2}}]}}}`)
			} else {
				_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-one"},"content":"first line\nsecond line","metadata":"{\"author\":\"Ada\"}"},{"_additional":{"id":"doc-two"},"content":"another document"}]}}}`)
			}
		case "/v1/objects/Docs/doc-one":
			_, _ = io.WriteString(w, `{"id":"doc-one","class":"Docs","properties":{"content":"first line\nsecond line","metadata":"{\"author\":\"Ada\"}"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}
	ctx := context.Background()

	list := captureUtilsOutput(t, func() {
		ListWeaviateDocuments(ctx, cfg, "Docs", 10, false, 1, false, false)
	})
	if !strings.Contains(list, "1.") || !strings.Contains(list, "2.") {
		t.Fatalf("list output = %q", list)
	}
	all := captureUtilsOutput(t, func() {
		ShowWeaviateDocument(ctx, cfg, "Docs", nil, true, 2, nil, "", true, true, false)
	})
	if strings.Count(all, "Collection: Docs") != 2 || !strings.Contains(all, "second line") {
		t.Fatalf("show all output = %q", all)
	}
	one := captureUtilsOutput(t, func() {
		ShowWeaviateDocument(ctx, cfg, "Docs", []string{"doc-one", "missing"}, false, 1, nil, "", true, true, false)
	})
	if !strings.Contains(one, "first line") || !strings.Contains(one, "Collection: Docs") {
		t.Fatalf("show selected output = %q", one)
	}
}
