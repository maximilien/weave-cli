// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package milvus

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbeddingSearchProtocol(t *testing.T) {
	for _, mode := range []string{"semantic", "hybrid", "keyword failure", "vector failure", "embedding failure"} {
		t.Run(mode, func(t *testing.T) {
			fields := fixtureResultSet("doc-1")
			f := &milvusFixture{described: fixtureCollection("docs", "text-embedding-3-small", 3), queryResults: []client.ResultSet{fields}, searchResults: []client.SearchResult{{ResultCount: 1, Fields: fields, Scores: []float32{.75}}}, fail: map[string]error{}}
			if mode == "keyword failure" {
				f.fail["Query"] = errors.New("fixture keyword failure")
			}
			if mode == "vector failure" {
				f.fail["Search"] = errors.New("fixture vector failure")
			}
			a := newMilvusAdapter(f)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct{ Input string }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Input != "guide" {
					t.Errorf("input=%s", body.Input)
				}
				data := []interface{}{map[string]interface{}{"embedding": []float64{.25, .5, .75}, "index": 0}}
				if mode == "embedding failure" {
					data = nil
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			}))
			defer server.Close()
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1")
			var err error
			a.llmClient, err = llm.NewOpenAIClientWithHTTP("fixture", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			var results []*vectordb.QueryResult
			if mode == "semantic" || mode == "embedding failure" {
				results, err = a.SearchSemantic(context.Background(), "docs", "guide", &vectordb.QueryOptions{TopK: 2})
			} else {
				results, err = a.SearchHybrid(context.Background(), "docs", "guide", &vectordb.QueryOptions{TopK: 2})
			}
			if mode == "embedding failure" {
				if err == nil {
					t.Fatal("expected embedding failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || results[0].Document.ID != "doc-1" {
				t.Fatalf("results=%+v", results)
			}
		})
	}
}
