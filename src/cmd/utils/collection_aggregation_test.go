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

func TestCollectionQueryAggregationSurvivesFailures(t *testing.T) {
	for _, cross := range []bool{false, true} {
		name := "single database"
		if cross {
			name = "cross database"
		}
		t.Run(name, func(t *testing.T) {
			cfg, _ := documentOperationsFixture(t, "text2vec-openai", false, false, false, false)
			options := weaviate.QueryOptions{TopK: 2, JSONOutput: true}
			out, diagnostics := captureCollectionListing(t, func() {
				if cross {
					QueryMultipleCollectionsCrossVDB(context.Background(), []CollectionSpec{{Name: "Unconfigured"}, {Name: "Missing"}, {Name: "Invalid"}, {Name: "Docs", VDBKey: "archive"}}, map[string]*config.VectorDBConfig{"Missing": cfg, "Docs": cfg, "Invalid": {Type: "invalid"}}, "lookup", options)
				} else {
					QueryMultipleCollections(context.Background(), cfg, []string{"Missing", "Docs"}, "lookup", options)
				}
			})
			if !strings.Contains(diagnostics, "Failed to query collection 'Missing'") {
				t.Fatalf("missing query diagnostic: %s", diagnostics)
			}
			if cross && (!strings.Contains(diagnostics, "No VDB configuration") || !strings.Contains(diagnostics, "Failed to create client")) {
				t.Fatalf("missing configuration diagnostics: %s", diagnostics)
			}
			start := strings.Index(out, "{\n")
			if start < 0 {
				t.Fatalf("missing JSON result: %s", out)
			}
			var got struct {
				Count   int                    `json:"count"`
				Query   string                 `json:"query"`
				Results []weaviate.QueryResult `json:"results"`
			}
			if err := json.Unmarshal([]byte(out[start:]), &got); err != nil {
				t.Fatalf("decode: %v: %s", err, out)
			}
			if got.Count != 2 || len(got.Results) != 2 || got.Query != "lookup" {
				t.Fatalf("results = %#v", got)
			}
			for i, result := range got.Results {
				want := []string{"first", "second"}[i]
				if result.ID != want || result.Content == "" || result.Metadata["_collection"] != "Docs" {
					t.Fatalf("result = %#v", result)
				}
				if cross && (result.Metadata["_vdb"] != "archive" || result.Metadata["_vdb_type"] != string(cfg.Type)) {
					t.Fatalf("missing source: %#v", result.Metadata)
				}
			}
		})
	}
}

func TestSingleCollectionQueryJSON(t *testing.T) {
	cfg, _ := documentOperationsFixture(t, "text2vec-openai", false, false, false, false)
	out, diagnostics := captureCollectionListing(t, func() {
		QueryWeaviateCollection(context.Background(), cfg, "Docs", "lookup", weaviate.QueryOptions{TopK: 2, JSONOutput: true})
	})
	var got struct {
		Count      int
		Collection string
		Results    []weaviate.QueryResult
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode: %v: %s %s", err, out, diagnostics)
	}
	if got.Count != 2 || got.Collection != "Docs" || len(got.Results) != 2 {
		t.Fatalf("results = %#v", got)
	}
}

func TestCollectionDeletionHelpersRespectPatternAndFailure(t *testing.T) {
	// These helpers operate through VectorDBClient; the HTTP fixture provides
	// that boundary without requiring a live Supabase instance.
	for _, tc := range []struct {
		name, pattern string
	}{
		{name: "explicit"}, {name: "matching pattern", pattern: "Doc*"}, {name: "unmatched pattern", pattern: "Other*"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, mutations := documentOperationsFixture(t, "none", false, false, false, false)
			var err error
			if tc.pattern != "" {
				err = DeleteSupabaseCollectionsByPattern(context.Background(), cfg, tc.pattern)
			} else {
				err = DeleteSupabaseCollections(context.Background(), cfg, []string{"Docs"})
			}
			if err != nil {
				t.Fatal(err)
			}
			want := 1
			if tc.pattern == "Other*" {
				want = 0
			}
			if len(mutations) != want {
				t.Fatalf("delete count = %d; want %d", len(mutations), want)
			}
			if want == 1 {
				if path := <-mutations; path != "/v1/schema/Docs" {
					t.Fatalf("wrong collection deleted: %s", path)
				}
			}
		})
	}
	for _, pattern := range []string{"", "*"} {
		cfg := &config.VectorDBConfig{Type: "invalid"}
		var err error
		if pattern == "" {
			err = DeleteSupabaseCollections(context.Background(), cfg, []string{"Docs"})
		} else {
			err = DeleteSupabaseCollectionsByPattern(context.Background(), cfg, pattern)
		}
		if err == nil || !strings.Contains(err.Error(), "failed to create Supabase client") {
			t.Fatalf("configuration error = %v", err)
		}
	}
}

func TestCollectionQueryInvalidClients(t *testing.T) {
	for _, dbType := range []config.VectorDBType{config.VectorDBTypeLocal, "invalid"} {
		for _, cross := range []bool{false, true} {
			cfg := &config.VectorDBConfig{Type: dbType, URL: "://bad"}
			out, diagnostics := captureCollectionListing(t, func() {
				if cross {
					QueryMultipleCollectionsCrossVDB(context.Background(), []CollectionSpec{{Name: "Docs"}}, map[string]*config.VectorDBConfig{"Docs": cfg}, "lookup", weaviate.QueryOptions{})
				} else {
					QueryMultipleCollections(context.Background(), cfg, []string{"Docs"}, "lookup", weaviate.QueryOptions{})
				}
			})
			want := "Failed to create"
			if dbType == config.VectorDBTypeLocal {
				want = "Failed to query"
			}
			if !strings.Contains(diagnostics, want) || !strings.Contains(diagnostics, "No results found from any collection") || strings.Contains(out, "Aggregated") {
				t.Fatalf("type=%s cross=%v output=%s diagnostics=%s", dbType, cross, out, diagnostics)
			}
		}
	}
}

func TestCollectionDeletionReportsBackendFailure(t *testing.T) {
	for _, pattern := range []bool{false, true} {
		cfg, mutations := documentOperationsFixture(t, "none", false, false, false, true)
		var err error
		if pattern {
			err = DeleteSupabaseCollectionsByPattern(context.Background(), cfg, "Doc*")
		} else {
			err = DeleteSupabaseCollections(context.Background(), cfg, []string{"Docs", "NeverDelete"})
		}
		if err == nil || !strings.Contains(err.Error(), "failed to delete collection Docs") {
			t.Fatalf("error = %v", err)
		}
		if len(mutations) != 1 {
			t.Fatalf("deletions after failure = %d", len(mutations))
		}
		if path := <-mutations; path != "/v1/schema/Docs" {
			t.Fatalf("wrong target: %s", path)
		}
	}
}
