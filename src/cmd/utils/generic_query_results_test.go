// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestGenericQueryResultsAndFailures(t *testing.T) {
	for _, tc := range []struct {
		name       string
		bm25, fail bool
	}{
		{name: "semantic"}, {name: "keyword", bm25: true}, {name: "semantic failure", fail: true}, {name: "keyword failure", bm25: true, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "none", false, tc.fail, false, false)
			out, diagnostics := captureCollectionListing(t, func() {
				QueryCollection(context.Background(), cfg, "Docs", "fixture query", weaviate.QueryOptions{TopK: 2, UseBM25: tc.bm25, JSONOutput: true})
			})
			if tc.fail {
				if !strings.Contains(diagnostics, "Failed to query collection 'Docs'") {
					t.Fatalf("diagnostics = %s", diagnostics)
				}
				return
			}
			var got struct {
				Collection string                 `json:"collection"`
				Query      string                 `json:"query"`
				Count      int                    `json:"count"`
				Results    []weaviate.QueryResult `json:"results"`
			}
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("decode: %v: %s %s", err, out, diagnostics)
			}
			if got.Collection != "Docs" || got.Query != "fixture query" || got.Count != 2 || len(got.Results) != 2 || got.Results[0].ID != "first" || got.Results[1].ID != "second" {
				t.Fatalf("results = %#v", got)
			}
		})
	}
}

func TestGenericAgentQueryStopsOnEmptyMockResults(t *testing.T) {
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeMock, Enabled: true, Collections: []config.Collection{{Name: "Photos", Type: "image"}, {Name: "Docs", Type: "text"}}}
	for _, cross := range []bool{false, true} {
		for _, bm25 := range []bool{false, true} {
			options := weaviate.QueryOptions{TopK: 5, TopKImages: 2, UseBM25: bm25, Verbose: true}
			_, diagnostics := captureCollectionListing(t, func() {
				if cross {
					QueryMultipleCollectionsWithAgentCrossVDB(context.Background(), []CollectionSpec{{Name: "Photos"}, {Name: "Docs"}}, map[string]*config.VectorDBConfig{"Photos": cfg, "Docs": cfg}, "query", options, []string{"unused"}, "text", false)
				} else {
					QueryMultipleCollectionsWithAgent(context.Background(), cfg, []string{"Photos", "Docs"}, "query", options, []string{"unused"}, "text", false)
				}
			})
			if !strings.Contains(diagnostics, "No results found from any collection") {
				t.Fatalf("cross=%v bm25=%v diagnostics=%s", cross, bm25, diagnostics)
			}
		}
	}
}

func TestGenericAgentQueryReportsInvalidDatabase(t *testing.T) {
	cfg := &config.VectorDBConfig{Type: "invalid"}
	for _, cross := range []bool{false, true} {
		_, diagnostics := captureCollectionListing(t, func() {
			if cross {
				QueryMultipleCollectionsWithAgentCrossVDB(context.Background(), []CollectionSpec{{Name: "Docs"}, {Name: "Missing"}}, map[string]*config.VectorDBConfig{"Docs": cfg}, "query", weaviate.QueryOptions{}, []string{"unused"}, "text", false)
			} else {
				QueryMultipleCollectionsWithAgent(context.Background(), cfg, []string{"Docs"}, "query", weaviate.QueryOptions{}, []string{"unused"}, "text", false)
			}
		})
		if !strings.Contains(diagnostics, "Failed to create client") || !strings.Contains(diagnostics, "No results found from any collection") {
			t.Fatalf("cross=%v diagnostics=%s", cross, diagnostics)
		}
		if cross && !strings.Contains(diagnostics, "No VDB configuration found") {
			t.Fatalf("missing config diagnostic: %s", diagnostics)
		}
	}
}
