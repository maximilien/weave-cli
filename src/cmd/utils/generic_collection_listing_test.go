// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// These presentation helpers share the VectorDBClient boundary. A local
// Weaviate fixture supplies collection data without requiring six services;
// adapter protocol behavior is tested in the respective adapter packages.
func TestGenericCollectionListingPresentation(t *testing.T) {
	for _, helper := range []struct {
		name string
		list func(context.Context, *config.VectorDBConfig, int, bool, bool)
	}{
		{"Supabase", ListSupabaseCollections},
		{"MongoDB", ListMongoDBCollections},
		{"Chroma", ListChromaCollections},
		{"Qdrant", ListQdrantCollections},
		{"Neo4j", ListNeo4jCollections},
		{"Milvus", ListMilvusCollections},
	} {
		t.Run(helper.name, func(t *testing.T) {
			for _, scenario := range []struct {
				name        string
				empty, fail bool
				limit       int
				json        bool
			}{
				{name: "text"},
				{name: "limited text", limit: 1},
				{name: "json", json: true},
				{name: "limited json", limit: 1, json: true},
				{name: "empty text", empty: true},
				{name: "empty json", empty: true, json: true},
				{name: "list failure", fail: true},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						switch r.URL.Path {
						case "/v1/meta":
							_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
						case "/v1/schema":
							if scenario.fail {
								http.Error(w, "fixture unavailable", http.StatusInternalServerError)
								return
							}
							if scenario.empty {
								_, _ = io.WriteString(w, `{"classes":[]}`)
								return
							}
							_, _ = io.WriteString(w, `{"classes":[{"class":"ZImages","description":"image archive","vectorizer":"none"},{"class":"Alpha","description":"text archive","vectorizer":"none"}]}`)
						case "/v1/graphql":
							_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Alpha":[{"meta":{"count":2}}],"ZImages":[{"meta":{"count":0}}]}}}`)
						default:
							t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
							http.NotFound(w, r)
						}
					}))
					defer server.Close()
					cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}
					output, diagnostics := captureCollectionListing(t, func() { helper.list(context.Background(), cfg, scenario.limit, false, scenario.json) })
					if scenario.fail {
						if !strings.Contains(diagnostics, "Failed to list collections") {
							t.Fatalf("missing failure diagnostic: %q", output)
						}
						if strings.Contains(output, "Found ") {
							t.Fatalf("failure reported success: %q", output)
						}
						return
					}
					wantCount := 2
					if scenario.empty {
						wantCount = 0
					} else if scenario.limit > 0 {
						wantCount = scenario.limit
					}
					if scenario.json {
						var got struct {
							Collections []struct {
								Name  string
								Count int64
							} `json:"collections"`
							Total int `json:"total"`
						}
						if err := json.Unmarshal([]byte(output), &got); err != nil {
							t.Fatalf("invalid JSON: %v: %q", err, output)
						}
						if got.Total != wantCount || len(got.Collections) != wantCount || got.Collections == nil {
							t.Fatalf("unexpected collection count: %+v", got)
						}
						if wantCount > 0 && (got.Collections[0].Name != "Alpha" || got.Collections[0].Count != 2) {
							t.Fatalf("wrong first collection: %+v", got)
						}
						if wantCount == 2 && (got.Collections[1].Name != "ZImages" || got.Collections[1].Count != 0) {
							t.Fatalf("wrong empty collection: %+v", got)
						}
						return
					}
					if scenario.empty {
						if !strings.Contains(diagnostics, "No collections found") {
							t.Fatalf("missing empty diagnostic: %q", output)
						}
						return
					}
					for _, want := range []string{"Alpha", "2 items"} {
						if !strings.Contains(output, want) {
							t.Errorf("missing %q: %q", want, output)
						}
					}
					if scenario.limit > 0 {
						if strings.Contains(output, "ZImages") || !strings.Contains(diagnostics, "showing first 1 collections") {
							t.Fatalf("limit ignored: %q", output)
						}
					} else if strings.Index(output, "ZImages") < strings.Index(output, "Alpha") || !strings.Contains(output, "empty") {
						t.Fatalf("wrong sorting or empty count: %q", output)
					}
				})
			}
			t.Run("invalid configuration", func(t *testing.T) {
				output, diagnostics := captureCollectionListing(t, func() { helper.list(context.Background(), &config.VectorDBConfig{Type: "invalid"}, 0, false, false) })
				if !strings.Contains(diagnostics, "Failed to create "+helper.name+" client") {
					t.Fatalf("missing client diagnostic: %q (stdout %q)", diagnostics, output)
				}
			})
		})
	}
}

// Keep diagnostics separate so JSON assertions also protect stdout purity.
func captureCollectionListing(t *testing.T, run func()) (string, string) {
	t.Helper()
	file, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	previous := os.Stderr
	os.Stderr = file
	defer func() { os.Stderr = previous }()
	output := captureUtilsOutput(t, run)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	diagnostics, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	return output, string(diagnostics)
}
