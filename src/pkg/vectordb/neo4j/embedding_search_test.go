// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package neo4j

import (
	"context"
	"encoding/json"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSemanticEmbeddingProtocol(t *testing.T) {
	for _, mode := range []string{"default", "limited", "mismatch", "embedding failure", "missing key"} {
		t.Run(mode, func(t *testing.T) {
			a, e, _ := newNeo4jProtocolFixture()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct{ Input string }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Input != "find guides" {
					t.Errorf("query=%s", body.Input)
				}
				v := []float64{.25, .5, .75}
				if mode == "mismatch" {
					v = v[:1]
				}
				data := []interface{}{map[string]interface{}{"embedding": v, "index": 0}}
				if mode == "embedding failure" {
					data = nil
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			}))
			defer server.Close()
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1")
			if mode != "missing key" {
				var err error
				a.llmClient, err = llm.NewOpenAIClientWithHTTP("fixture", server.Client())
				if err != nil {
					t.Fatal(err)
				}
			}
			opts := &vectordb.QueryOptions{}
			if mode == "limited" {
				opts.TopK = 2
			}
			results, err := a.SearchSemantic(context.Background(), "docs", "find guides", opts)
			if mode == "mismatch" || mode == "embedding failure" || mode == "missing key" {
				if err == nil {
					t.Fatal("expected embedding failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || results[0].Document.ID != "doc-1" || results[0].Score != .75 {
				t.Fatalf("results=%+v", results)
			}
			found := false
			for i, q := range e.queries {
				if strings.Contains(q, "queryNodes") {
					found = true
					encoded, _ := json.Marshal(e.params[i])
					if !strings.Contains(string(encoded), "0.25") {
						t.Fatalf("vector missing: %s", encoded)
					}
				}
			}
			if !found {
				t.Fatal("vector query missing")
			}
		})
	}
}
