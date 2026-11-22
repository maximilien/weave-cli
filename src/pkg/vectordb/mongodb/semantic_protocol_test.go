// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mongodb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoSemanticSearchProtocol(t *testing.T) {
	for _, mode := range []string{"results", "empty", "aggregate failure", "decode failure", "missing key", "embedding failure", "dimension mismatch"} {
		runMongoMock(t, mode, func(mt *mtest.T, adapter *Adapter) {
			ctx := context.Background()
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" || r.Header.Get("Authorization") != "Bearer fixture-key" {
					mt.Errorf("unexpected embedding request: %s %s", r.Method, r.URL.Path)
				}
				var body struct {
					Input string
					Model string
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					mt.Error(err)
				}
				if body.Input != "find technical guides" || body.Model != "text-embedding-3-small" {
					mt.Errorf("embedding request=%+v", body)
				}
				w.Header().Set("Content-Type", "application/json")
				data := []interface{}{map[string]interface{}{"index": 0, "embedding": []float64{.25, .75}}}
				if mode == "embedding failure" {
					data = nil
				}
				if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": data}); err != nil {
					mt.Error(err)
				}
			}))
			defer server.Close()
			mt.Setenv("OPENAI_BASE_URL", server.URL+"/v1")
			if mode != "missing key" {
				var err error
				adapter.llmClient, err = llm.NewOpenAIClientWithHTTP("fixture-key", server.Client())
				if err != nil {
					mt.Fatal(err)
				}
			}
			dims := 2
			if mode == "dimension mismatch" {
				dims = 3
			}
			metadata := bson.D{{Key: "vector_dimensions", Value: dims}}
			mt.AddMockResponses(
				mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, metadata),
				mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, metadata),
			)
			doc := append(mongoFixtureDocument(), bson.E{Key: "score", Value: .875})
			switch mode {
			case "results":
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, doc))
			case "empty":
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch))
			case "aggregate failure":
				mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 42, Message: "fixture search unavailable"}))
			case "decode failure":
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, bson.D{{Key: "score", Value: "invalid-score"}}))
			}
			results, err := adapter.SearchSemantic(ctx, "docs", "find technical guides", &vectordb.QueryOptions{TopK: 3})
			wantErr := map[string]string{"aggregate failure": "vector search failed", "decode failure": "failed to decode result", "missing key": "requires OpenAI API key", "embedding failure": "failed to generate query embedding", "dimension mismatch": "embedding dimension mismatch"}[mode]
			if wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), wantErr) {
					mt.Fatalf("search error=%v want %q", err, wantErr)
				}
			} else if err != nil {
				mt.Fatal(err)
			}
			wantCalls := 1
			if mode == "missing key" {
				wantCalls = 0
			}
			if calls != wantCalls {
				mt.Fatalf("embedding calls=%d want=%d", calls, wantCalls)
			}
			if mode == "results" {
				if len(results) != 1 || results[0].Document.ID != "doc-1" || results[0].Document.Content != "full content" || results[0].Document.Metadata["kind"] != "guide" || results[0].Score != .875 {
					mt.Fatalf("results=%+v", results)
				}
			}
			if mode == "empty" && len(results) != 0 {
				mt.Fatalf("empty results=%+v", results)
			}
			if mode == "results" || mode == "empty" || mode == "aggregate failure" || mode == "decode failure" {
				var pipeline bson.A
				for _, event := range mt.GetAllStartedEvents() {
					if event.CommandName == "aggregate" {
						var command struct{ Pipeline bson.A }
						if err := bson.Unmarshal(event.Command, &command); err != nil {
							mt.Fatal(err)
						}
						pipeline = command.Pipeline
					}
				}
				if len(pipeline) != 2 {
					mt.Fatalf("pipeline=%v", pipeline)
				}
				stage := pipeline[0].(bson.D)[0]
				search := stage.Value.(bson.D).Map()
				if stage.Key != "$vectorSearch" || search["index"] != "vector_index" || search["path"] != "embedding" || search["limit"] != int32(3) || search["numCandidates"] != int32(30) || !reflect.DeepEqual(search["queryVector"], bson.A{.25, .75}) {
					mt.Fatalf("vector search=%v", stage)
				}
			}
		})
	}
}

func TestMongoCollectionListingProtocol(t *testing.T) {
	for _, mode := range []string{"collections", "empty", "list failure", "count failure", "schema fallback"} {
		runMongoMock(t, mode, func(mt *mtest.T, adapter *Adapter) {
			if mode == "list failure" {
				mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 42, Message: "fixture list unavailable"}))
			} else if mode == "empty" {
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.$cmd.listCollections", mtest.FirstBatch))
			} else {
				mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.$cmd.listCollections", mtest.FirstBatch, bson.D{{Key: "name", Value: "docs"}}))
				if mode == "count failure" {
					mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 42, Message: "fixture count unavailable"}))
				} else {
					mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(7)}}))
				}
				if mode == "schema fallback" {
					mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch))
				} else {
					mt.AddMockResponses(mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, bson.D{{Key: "vector_dimensions", Value: 3072}}))
				}
			}
			collections, err := adapter.ListCollections(context.Background())
			if mode == "list failure" {
				if err == nil || !strings.Contains(err.Error(), "failed to list collections") {
					mt.Fatalf("listing error=%v", err)
				}
				return
			}
			if err != nil {
				mt.Fatal(err)
			}
			if mode == "empty" {
				if len(collections) != 0 {
					mt.Fatalf("collections=%v", collections)
				}
				return
			}
			count := int64(7)
			if mode == "count failure" {
				count = 0
			}
			model := "text-embedding-3-large"
			if mode == "schema fallback" {
				model = "text-embedding-3-small"
			}
			if len(collections) != 1 || collections[0].Name != "docs" || collections[0].Count != count || collections[0].Vectorizer != model {
				mt.Fatalf("collections=%+v", collections)
			}
		})
	}
}
