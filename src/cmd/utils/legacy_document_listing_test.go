// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestLegacyWeaviateDocumentListing(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		virtual, empty, failList, failCount bool
		want                                string
	}{
		{name: "regular", want: "Found 2 documents"},
		{name: "virtual", virtual: true, want: "Found 2 virtual documents"},
		{name: "count unavailable", failCount: true, want: "Found 2 documents"},
		{name: "empty", empty: true, want: "No documents found"},
		{name: "list failure", failList: true, want: "Failed to list documents"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "none", tc.empty, tc.failList, tc.failCount, false)
			out, diagnostics := captureCollectionListing(t, func() { ListWeaviateDocuments(context.Background(), cfg, "Docs", 2, false, 3, tc.virtual, false) })
			if !strings.Contains(out+diagnostics, tc.want) {
				t.Fatalf("output = %q %q; want %q", out, diagnostics, tc.want)
			}
			if tc.failList && strings.Contains(out, "First content") {
				t.Fatalf("failed query displayed document: %s", out)
			}
		})
	}
}

func TestClearAllWeaviateCollectionsContinuesAfterFailures(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		empty, failList, failA, failB bool
		want                          string
		wantDeletes                   int
	}{
		{name: "both succeed", want: "Successfully cleared 2 collections", wantDeletes: 2},
		{name: "partial failure", failA: true, want: "Successfully cleared 1 collections", wantDeletes: 2},
		{name: "all fail", failA: true, failB: true, want: "Failed to clear any collections", wantDeletes: 2},
		{name: "empty", empty: true, want: "No collections found to delete"},
		{name: "list failure", failList: true, want: "Failed to list collections"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deletes := make(chan string, 3)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/v1/meta":
					_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
				case r.URL.Path == "/v1/schema":
					if tc.failList {
						http.Error(w, "list rejected", 500)
						return
					}
					if tc.empty {
						_, _ = io.WriteString(w, `{"classes":[]}`)
						return
					}
					_, _ = io.WriteString(w, `{"classes":[{"class":"Alpha","properties":[{"name":"content","dataType":["text"]}]},{"class":"Beta","properties":[{"name":"content","dataType":["text"]}]}]}`)
				case r.URL.Path == "/v1/schema/Alpha" || r.URL.Path == "/v1/schema/Beta":
					_, _ = io.WriteString(w, `{"properties":[{"name":"content","dataType":["text"]}]}`)
				case r.URL.Path == "/v1/graphql":
					var request struct{ Query string }
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
					}
					class := "Alpha"
					if strings.Contains(request.Query, "Beta") {
						class = "Beta"
					}
					_, _ = io.WriteString(w, `{"data":{"Get":{"`+class+`":[{"_additional":{"id":"`+strings.ToLower(class)+`-id"},"content":"fixture"}]}}}`)
				case r.Method == http.MethodDelete:
					deletes <- r.URL.Path
					if (tc.failA && strings.Contains(r.URL.Path, "alpha-id")) || (tc.failB && strings.Contains(r.URL.Path, "beta-id")) {
						http.Error(w, "delete rejected", 500)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL, Timeout: 1}
			out, diagnostics := captureCollectionListing(t, func() { DeleteAllWeaviateCollections(context.Background(), cfg) })
			if !strings.Contains(out+diagnostics, tc.want) {
				t.Fatalf("output = %q %q; want %q", out, diagnostics, tc.want)
			}
			if len(deletes) != tc.wantDeletes {
				t.Fatalf("delete count = %d; want %d", len(deletes), tc.wantDeletes)
			}
			seen := map[string]bool{}
			for len(deletes) > 0 {
				seen[<-deletes] = true
			}
			if tc.wantDeletes == 2 && (!seen["/v1/objects/Alpha/alpha-id"] || !seen["/v1/objects/Beta/beta-id"]) {
				t.Fatalf("wrong deletion targets: %#v", seen)
			}
			if tc.failA && !strings.Contains(diagnostics, "Failed to clear collection 'Alpha'") {
				t.Fatalf("missing failure: %s", diagnostics)
			}
		})
	}
}

func TestShowWeaviateCollectionReportsMissingAndCountFailure(t *testing.T) {
	for _, tc := range []struct {
		name, collection, want string
		failCount              bool
	}{
		{name: "missing", collection: "Missing", want: "Collection 'Missing' not found"},
		{name: "count failure", collection: "Docs", failCount: true, want: "Failed to get document count"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "none", false, false, tc.failCount, false)
			out, diagnostics := captureCollectionListing(t, func() {
				ShowWeaviateCollection(context.Background(), cfg, tc.collection, 3, false, false, false, false, false, false, false, "", "", false)
			})
			if !strings.Contains(diagnostics, tc.want) {
				t.Fatalf("diagnostics = %q; want %q", diagnostics, tc.want)
			}
			if strings.Contains(out, "Document Count") {
				t.Fatalf("show reported success after error: %q", out)
			}
		})
	}
}
