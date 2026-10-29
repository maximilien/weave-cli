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
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestAgentQueriesUseImageLimitAndStopOnEmpty(t *testing.T) {
	for _, cross := range []bool{false, true} {
		name := "single database"
		if cross {
			name = "cross database"
		}
		t.Run(name, func(t *testing.T) {
			queries := make(chan string, 8)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/v1/meta":
					_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
				case "/v1/schema":
					_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","vectorizer":"none","properties":[{"name":"content","dataType":["text"]}]},{"class":"Photos","vectorizer":"none","properties":[{"name":"url","dataType":["text"]}]}]}`)
				case "/v1/graphql":
					var request struct{ Query string }
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
					}
					queries <- request.Query
					class := "Docs"
					if strings.Contains(request.Query, "Photos") {
						class = "Photos"
					}
					_, _ = io.WriteString(w, `{"data":{"Get":{"`+class+`":[]}}}`)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL, Timeout: 1}
			options := weaviate.QueryOptions{TopK: 5, TopKImages: 2, Verbose: true}
			_, diagnostics := captureCollectionListing(t, func() {
				if cross {
					QueryMultipleCollectionsWithAgentCrossVDB(context.Background(), []CollectionSpec{{Name: "MissingConfig"}, {Name: "Docs", VDBKey: "fixture"}, {Name: "Photos", VDBKey: "fixture"}}, map[string]*config.VectorDBConfig{"Docs": cfg, "Photos": cfg}, "query", options, []string{"unused"}, "text", false)
				} else {
					QueryMultipleCollectionsWithAgent(context.Background(), cfg, []string{"Docs", "Photos"}, "query", options, []string{"unused"}, "text", false)
				}
			})
			if !strings.Contains(diagnostics, "No results found from any collection") {
				t.Fatalf("missing empty result: %s", diagnostics)
			}
			if cross && !strings.Contains(diagnostics, "No VDB configuration found") {
				t.Fatalf("missing configuration diagnostic: %s", diagnostics)
			}
			if len(queries) != 2 {
				t.Fatalf("queries=%d; want 2", len(queries))
			}
			for len(queries) > 0 {
				query := <-queries
				want := "limit: 5"
				if strings.Contains(query, "Photos") {
					want = "limit: 2"
				}
				if !strings.Contains(query, want) {
					t.Errorf("wrong query limit: %s", query)
				}
			}
		})
	}
}
