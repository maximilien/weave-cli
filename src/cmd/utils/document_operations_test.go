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

// Keep the protocol fixture local and inspect mutations separately from output.
func documentOperationsFixture(t *testing.T, vectorizer string, empty, failList, failCount, failDelete bool) (*config.VectorDBConfig, <-chan string) {
	t.Helper()
	mutations := make(chan string, 16)
	schema := map[string]interface{}{"class": "Docs", "vectorizer": vectorizer, "properties": []map[string]interface{}{{"name": "content", "dataType": []string{"text"}}, {"name": "metadata", "dataType": []string{"text"}}}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/meta":
			_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
		case r.Method == http.MethodDelete:
			mutations <- r.URL.Path
			if failDelete && strings.HasSuffix(r.URL.Path, "/second") {
				http.Error(w, "delete rejected", 500)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/v1/schema":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"classes": []interface{}{schema}})
		case r.URL.Path == "/v1/schema/Docs":
			_ = json.NewEncoder(w).Encode(schema)
		case r.URL.Path == "/v1/graphql":
			var body struct{ Query string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if strings.Contains(body.Query, "Aggregate") {
				if failCount {
					http.Error(w, "count rejected", 500)
					return
				}
				_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":2}}]}}}`)
				return
			}
			if failList {
				http.Error(w, "query rejected", 500)
				return
			}
			if empty {
				_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[]}}}`)
				return
			}
			_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"first"},"content":"First content","metadata":"{\"filename\":\"first.txt\"}"},{"_additional":{"id":"second"},"content":"Second content","metadata":"{\"filename\":\"second.txt\"}"}]}}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return &config.VectorDBConfig{Name: "fixture", Type: config.VectorDBTypeLocal, URL: server.URL, Timeout: 1}, mutations
}

func TestDocumentListingPaginationAndFailures(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		offset                              int
		virtual, empty, failList, failCount bool
	}{
		{name: "first page"}, {name: "second page", offset: 1}, {name: "virtual", virtual: true}, {name: "count unavailable", failCount: true}, {name: "empty", empty: true}, {name: "offset past end", offset: 2}, {name: "list failure", failList: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "none", tc.empty, tc.failList, tc.failCount, false)
			out, diagnostics := captureCollectionListing(t, func() {
				ListDocuments(context.Background(), cfg, "Docs", 1, tc.offset, false, 3, false, tc.virtual, false, true)
			})
			if tc.failList {
				if !strings.Contains(diagnostics, "Failed to list documents") {
					t.Fatalf("diagnostics = %q", diagnostics)
				}
				return
			}
			if tc.empty || tc.offset == 2 {
				if !strings.Contains(out+diagnostics, "No documents found") {
					t.Fatalf("output = %q %q", out, diagnostics)
				}
				return
			}
			var got map[string]interface{}
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("invalid JSON: %v: %s", err, out)
			}
			if got["collection"] != "Docs" {
				t.Fatalf("collection = %#v", got)
			}
			if tc.virtual {
				if !strings.Contains(out, "first.txt") {
					t.Fatalf("missing virtual document: %s", out)
				}
				return
			}
			docs, ok := got["documents"].([]interface{})
			if !ok || len(docs) != 1 {
				t.Fatalf("documents = %#v", got)
			}
			want := "first"
			if tc.offset == 1 {
				want = "second"
			}
			if docs[0].(map[string]interface{})["id"] != want {
				t.Fatalf("wrong page: %#v", docs)
			}
			count, exists := got["total_documents"]
			if tc.failCount {
				if exists {
					t.Fatalf("invented total: %#v", count)
				}
			} else if count != float64(2) {
				t.Fatalf("total = %#v", count)
			}
		})
	}
}

func TestDocumentCountAndClearFailures(t *testing.T) {
	for _, tc := range []struct {
		name                                   string
		empty, failList, failCount, failDelete bool
	}{
		{name: "success"}, {name: "empty", empty: true}, {name: "list failure", failList: true}, {name: "count failure", failCount: true}, {name: "partial delete", failDelete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, mutations := documentOperationsFixture(t, "none", tc.empty, tc.failList, tc.failCount, tc.failDelete)
			n, err := CountWeaviateDocuments(context.Background(), cfg, "Docs")
			if tc.failCount {
				if err == nil || !strings.Contains(err.Error(), "failed to count documents") {
					t.Fatalf("count error = %v", err)
				}
			} else if err != nil || n != 2 {
				t.Fatalf("count = %d, %v", n, err)
			}
			out, diagnostics := captureCollectionListing(t, func() { DeleteAllDocuments(context.Background(), cfg, "Docs") })
			if tc.failList || tc.failDelete {
				if !strings.Contains(diagnostics, "Failed to delete all documents") || strings.Contains(out+diagnostics, "Successfully deleted") {
					t.Fatalf("false success: %s %s", out, diagnostics)
				}
			} else if !strings.Contains(out+diagnostics, "Successfully deleted all documents") {
				t.Fatalf("output = %s %s", out, diagnostics)
			}
			want := 2
			if tc.empty || tc.failList {
				want = 0
			}
			if len(mutations) != want {
				t.Fatalf("deletions = %d, want %d", len(mutations), want)
			}
			for i := 0; i < want; i++ {
				path := <-mutations
				if !strings.Contains(path, "/Docs/") || !(strings.HasSuffix(path, "/first") || strings.HasSuffix(path, "/second")) {
					t.Fatalf("wrong deletion target: %s", path)
				}
			}
		})
	}
}

func TestEmbeddingCompatibilityDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, vectorizer, model, collection, want string
		wantErr                                   bool
	}{
		{"disabled", "none", "custom", "Docs", "will be ignored", false},
		{"other vectorizer", "text2vec-cohere", "custom", "Docs", "only works with", false},
		{"recommended", "text2vec-openai", "text-embedding-3-small", "Docs", "Using embedding model", false},
		{"custom model", "text2vec-openai", "future-model", "Docs", "will be used as specified", false},
		{"missing collection", "none", "custom", "Missing", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, tc.vectorizer, false, false, false, false)
			client, err := CreateWeaviateClient(cfg)
			if err != nil {
				t.Fatal(err)
			}
			var validationErr error
			out, diagnostics := captureCollectionListing(t, func() {
				validationErr = validateEmbeddingForCollection(context.Background(), client, tc.collection, tc.model)
			})
			if tc.wantErr {
				if validationErr == nil || !strings.Contains(validationErr.Error(), "failed to get collection schema") {
					t.Fatalf("error = %v", validationErr)
				}
				return
			}
			if validationErr != nil || !strings.Contains(out+diagnostics, tc.want) {
				t.Fatalf("result = %v, %s %s", validationErr, out, diagnostics)
			}
			if tc.model == "text-embedding-3-small" && strings.Contains(out+diagnostics, "Recommended") {
				t.Fatal("recommended model warned as unknown")
			}
		})
	}
}

func TestFindCollectionAcrossWeaviateDatabases(t *testing.T) {
	cfg, _ := documentOperationsFixture(t, "none", false, false, false, false)
	configs := []config.VectorDBConfig{{Name: "invalid", URL: "://bad", Type: config.VectorDBTypeLocal}, *cfg}
	for _, name := range []string{"Docs", "Missing"} {
		got := TryWeaviateDatabasesForCollection(context.Background(), configs, name)
		if name == "Missing" {
			if got != nil {
				t.Fatalf("selected missing collection: %#v", got)
			}
		} else if got == nil || got.Name != "fixture" {
			t.Fatalf("selected = %#v", got)
		}
	}
}
