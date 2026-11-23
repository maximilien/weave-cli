// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package opensearch

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

func TestEmbeddingSearchProtocol(t *testing.T) {
	for _, mode := range []string{"semantic default", "semantic limited", "hybrid default", "hybrid limited", "mismatch", "embedding failure"} {
		t.Run(mode, func(t *testing.T) {
			a, rec := newOpenSearchProtocolAdapter(t, "green")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/embeddings" || r.Header.Get("Authorization") != "Bearer fixture-key" {
					t.Errorf("unexpected embedding request %s", r.URL.Path)
				}
				var body struct{ Input string }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Input != "find guides" {
					t.Errorf("query=%q", body.Input)
				}
				vector := []float64{.25, .5, .75}
				if mode == "mismatch" {
					vector = vector[:1]
				}
				data := []interface{}{map[string]interface{}{"embedding": vector, "index": 0}}
				if mode == "embedding failure" {
					data = nil
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			}))
			defer server.Close()
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1")
			var err error
			a.llmClient, err = llm.NewOpenAIClientWithHTTP("fixture-key", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			var opts *vectordb.QueryOptions
			wantK := float64(10)
			if strings.Contains(mode, "limited") {
				opts = &vectordb.QueryOptions{TopK: 2}
				wantK = 2
			}
			var results []*vectordb.QueryResult
			if strings.HasPrefix(mode, "hybrid") {
				results, err = a.SearchHybrid(context.Background(), "docs", "find guides", opts)
			} else {
				results, err = a.SearchSemantic(context.Background(), "docs", "find guides", opts)
			}
			if mode == "mismatch" || mode == "embedding failure" {
				if err == nil {
					t.Fatal("expected embedding error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(results) == 0 || results[0].Document.Content != "world" {
				t.Fatalf("results=%+v", results)
			}
			rec.mu.Lock()
			defer rec.mu.Unlock()
			var body map[string]interface{}
			for i, path := range rec.paths {
				if strings.Contains(path, "_search") {
					if err := json.Unmarshal([]byte(rec.bodies[i]), &body); err != nil {
						t.Fatal(err)
					}
				}
			}
			if body["size"] != wantK {
				t.Fatalf("search body=%v", body)
			}
			encoded, _ := json.Marshal(body)
			if !strings.Contains(string(encoded), "0.25") || !strings.Contains(string(encoded), "0.75") {
				t.Fatalf("vector missing: %s", encoded)
			}
			if strings.HasPrefix(mode, "hybrid") && !strings.Contains(string(encoded), "find guides") {
				t.Fatalf("text query missing: %s", encoded)
			}
		})
	}
}
