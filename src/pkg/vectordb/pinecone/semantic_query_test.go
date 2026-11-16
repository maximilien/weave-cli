// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pinecone

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	pcsdk "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type queryRecordingIndex struct {
	*fakeVectorIndex
	request *pcsdk.QueryByVectorValuesRequest
}

func (i *queryRecordingIndex) QueryByVectorValues(ctx context.Context, request *pcsdk.QueryByVectorValuesRequest) (*pcsdk.QueryVectorsResponse, error) {
	i.request = request
	return i.fakeVectorIndex.QueryByVectorValues(ctx, request)
}

func embeddingDataFixture(t *testing.T, mode string) (*Adapter, *queryRecordingIndex, *[]string) {
	t.Helper()
	adapter := newPineconeControlAdapter(t, pineconeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return pineconeResponse(200, `{"name":"docs","dimension":3,"host":"data.fixture","metric":"cosine"}`), nil
	}))
	metadata, _ := structpb.NewStruct(map[string]interface{}{"content": "matched text", "owner": "Ada"})
	index := &queryRecordingIndex{fakeVectorIndex: &fakeVectorIndex{query: &pcsdk.QueryVectorsResponse{Matches: []*pcsdk.ScoredVector{{Vector: &pcsdk.Vector{Id: "matched", Metadata: metadata}, Score: 0.75}, {Vector: &pcsdk.Vector{Id: "bare"}, Score: 0.5}}}}}
	executor := &fakeVectorExecutor{index: index}
	adapter.executor = executor
	if mode == "connection failure" {
		executor.indexErr = errors.New("fixture connection failed")
	}
	if mode == "describe failure" {
		executor.describeErr = errors.New("fixture describe failed")
	}
	if mode == "query failure" || mode == "write failure" {
		index.err = errors.New("fixture operation failed")
	}
	inputs := []string{}
	client, err := llm.NewOpenAIClientWithHTTP("fixture-key", &http.Client{Transport: pineconeRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		var request struct {
			Input string `json:"input"`
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.Model != "text-embedding-3-small" {
			t.Errorf("embedding model=%q", request.Model)
		}
		inputs = append(inputs, request.Input)
		if mode == "embedding failure" {
			response := pineconeResponse(400, `{"error":{"message":"fixture embedding failed","type":"invalid_request_error"}}`)
			response.Header.Set("Content-Type", "application/json")
			return response, nil
		}
		vector := []float64{0.25, 0.5, 0.75}
		if mode == "dimension mismatch" {
			vector = vector[:2]
		}
		data, _ := json.Marshal(map[string]interface{}{"object": "list", "data": []interface{}{map[string]interface{}{"index": 0, "embedding": vector}}})
		response := pineconeResponse(200, string(data))
		response.Header.Set("Content-Type", "application/json")
		return response, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	adapter.llmClient = client
	return adapter, index, &inputs
}

func TestSemanticQueryEmbeddingAndResults(t *testing.T) {
	for _, mode := range []string{"default", "custom limit", "hybrid", "dimension mismatch", "embedding failure", "connection failure", "describe failure", "query failure"} {
		t.Run(mode, func(t *testing.T) {
			adapter, index, inputs := embeddingDataFixture(t, mode)
			var options *vectordb.QueryOptions
			if mode == "custom limit" {
				options = &vectordb.QueryOptions{TopK: 2}
			}
			var results []*vectordb.QueryResult
			var err error
			if mode == "hybrid" {
				results, err = adapter.SearchHybrid(context.Background(), "docs", "find text", options)
			} else {
				results, err = adapter.SearchSemantic(context.Background(), "docs", "find text", options)
			}
			wantError := map[string]string{"dimension mismatch": "embedding dimension mismatch", "embedding failure": "failed to generate query embedding", "connection failure": "failed to connect to index", "describe failure": "failed to describe index", "query failure": "failed to query vectors"}[mode]
			if wantError != "" {
				if err == nil || !strings.Contains(err.Error(), wantError) {
					t.Fatalf("query error=%v want %q", err, wantError)
				}
			} else {
				if err != nil || len(results) != 2 {
					t.Fatalf("results=%+v error=%v", results, err)
				}
				if results[0].Document.ID != "matched" || results[0].Document.Content != "matched text" || results[0].Document.Metadata["owner"] != "Ada" || results[0].Score != 0.75 || results[1].Document.ID != "bare" {
					t.Fatalf("converted results=%+v", results)
				}
				wantK := uint32(10)
				if options != nil {
					wantK = 2
				}
				if index.request.TopK != wantK || !index.request.IncludeMetadata || !index.request.IncludeValues || !reflect.DeepEqual(index.request.Vector, []float32{0.25, 0.5, 0.75}) {
					t.Fatalf("query request=%+v", index.request)
				}
			}
			if !reflect.DeepEqual(*inputs, []string{"find text"}) {
				t.Fatalf("embedding inputs=%v", *inputs)
			}
			wantClosed := 0
			if index.request != nil {
				wantClosed = 1
			}
			if index.closed != wantClosed {
				t.Fatalf("index closes=%d want=%d", index.closed, wantClosed)
			}
		})
	}
}
