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

func TestEmbeddingWriteProtocol(t *testing.T) {
	for _, mode := range []string{"single", "batch", "update", "single failure", "batch failure", "update failure"} {
		t.Run(mode, func(t *testing.T) {
			a, executor, _ := newNeo4jProtocolFixture()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct{ Input string }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if body.Input != "guide" {
					t.Errorf("input=%s", body.Input)
				}
				data := []interface{}{map[string]interface{}{"index": 0, "embedding": []float64{.25, .5, .75}}}
				if mode == "single failure" || mode == "batch failure" || mode == "update failure" {
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
			doc := &vectordb.Document{ID: "550e8400-e29b-41d4-a716-446655440000", Text: "guide", Metadata: map[string]interface{}{"kind": "technical"}}
			switch mode {
			case "single", "single failure":
				err = a.CreateDocument(context.Background(), "docs", doc)
			case "batch", "batch failure":
				err = a.CreateDocuments(context.Background(), "docs", []*vectordb.Document{doc})
			default:
				err = a.UpdateDocument(context.Background(), "docs", doc)
			}
			if mode == "single failure" || mode == "batch failure" || mode == "update failure" {
				if err == nil || len(executor.queries) != 0 {
					t.Fatalf("error=%v upserts=%d", err, len(executor.queries))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, params := range executor.params {
				encoded, _ := json.Marshal(params)
				if strings.Contains(string(encoded), "0.25") && strings.Contains(string(encoded), "guide") {
					found = true
				}
			}
			if !found {
				t.Fatalf("embedding payload missing: %v", executor.params)
			}
		})
	}
}
