// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type comparisonClient struct {
	vectordb.VectorDBClient
	results map[string][]*vectordb.QueryResult
	errors  map[string]error
	calls   []comparisonCall
}

type comparisonCall struct {
	collection string
	query      string
	topK       int
}

func (c *comparisonClient) SearchSemantic(_ context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	c.calls = append(c.calls, comparisonCall{collection: collectionName, query: query, topK: options.TopK})
	key := collectionName + ":" + query
	if err := c.errors[key]; err != nil {
		return nil, err
	}
	return c.results[key], nil
}

func TestGenerateComparisonCollectsResultsAndContinuesAfterFailure(t *testing.T) {
	client := &comparisonClient{
		results: map[string][]*vectordb.QueryResult{
			"alpha:first": {
				{Document: vectordb.Document{ID: "a1"}, Score: 0.9},
				{Document: vectordb.Document{ID: "a2"}, Score: 0.5},
			},
			"beta:first": {},
			"alpha:second": {
				{Document: vectordb.Document{ID: "a3"}, Score: 0.8},
			},
		},
		errors: map[string]error{"beta:second": errors.New("query unavailable")},
	}
	report := generateComparison(context.Background(), client, []string{"alpha", "beta"}, []string{"first", "second"}, 7)

	if len(client.calls) != 4 {
		t.Fatalf("expected four searches, got %#v", client.calls)
	}
	for _, call := range client.calls {
		if call.topK != 7 {
			t.Errorf("expected top-k 7, got %#v", call)
		}
	}
	if len(report.Queries) != 2 || len(report.Queries[0].Collections) != 2 || len(report.Queries[1].Collections) != 1 {
		t.Fatalf("unexpected report structure: %#v", report)
	}
	alpha := report.Queries[0].Collections[0]
	if alpha.Collection != "alpha" || len(alpha.TopResults) != 2 || alpha.TopResults[0].Rank != 1 || alpha.TopResults[1].Rank != 2 {
		t.Fatalf("unexpected alpha results: %#v", alpha)
	}
	if alpha.AvgScore < 0.699 || alpha.AvgScore > 0.701 {
		t.Fatalf("unexpected average score: %f", alpha.AvgScore)
	}
	if got := report.Queries[0].Collections[1].AvgScore; got != 0 {
		t.Fatalf("empty results should have zero average, got %f", got)
	}
}

func TestComparisonReportMarkdown(t *testing.T) {
	report := &ComparisonReport{
		Collections: []string{"alpha", "beta", "unused"},
		Generated:   time.Date(2026, 9, 14, 10, 30, 0, 0, time.UTC),
		Queries: []QueryComparison{{
			Query: "find cameras",
			Collections: []CollectionResults{
				{Collection: "alpha", AvgScore: 0.5, Latency: 20 * time.Millisecond, TopResults: []Result{{DocID: "a", Score: 0.5, Rank: 1}}},
				{Collection: "beta", AvgScore: 0.9, Latency: 10 * time.Millisecond, TopResults: []Result{{DocID: "b", Score: 0.9, Rank: 1}}},
			},
		}, {
			Query: "find lenses",
			Collections: []CollectionResults{
				{Collection: "alpha", AvgScore: 0.7, Latency: 40 * time.Millisecond},
				{Collection: "beta", AvgScore: 0.3, Latency: 30 * time.Millisecond},
			},
		}},
	}

	markdown := report.ToMarkdown()
	for _, want := range []string{
		"# Embedding Model Comparison Report",
		"2026-09-14 10:30:00",
		"1. `alpha`",
		"| alpha | 0.600 | 30ms |",
		"| beta | 0.600 | 20ms |",
		"## Query: \"find cameras\"",
		"### 1. beta",
		"| 1 | `b` | 0.900 |",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("markdown missing %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, "| unused |") {
		t.Fatalf("unused collection should not have summary metrics:\n%s", markdown)
	}
	if strings.Index(markdown, "### 1. beta") > strings.Index(markdown, "### 2. alpha") {
		t.Fatalf("query results should be score-sorted:\n%s", markdown)
	}
}

func TestComparisonReportJSON(t *testing.T) {
	report := &ComparisonReport{
		Collections: []string{"alpha", "beta"},
		Generated:   time.Date(2026, 9, 14, 10, 30, 0, 0, time.UTC),
		Queries: []QueryComparison{{
			Query:       "find cameras",
			Collections: []CollectionResults{{Collection: "alpha", AvgScore: 0.875, Latency: 1250 * time.Millisecond}},
		}},
	}
	output := report.ToJSON()
	var parsed struct {
		Generated   string   `json:"generated"`
		Collections []string `json:"collections"`
		Queries     []struct {
			Query   string `json:"query"`
			Results []struct {
				Collection string  `json:"collection"`
				AvgScore   float64 `json:"avg_score"`
				LatencyMS  int64   `json:"latency_ms"`
			} `json:"results"`
		} `json:"queries"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid report JSON: %v\n%s", err, output)
	}
	if parsed.Generated != "2026-09-14T10:30:00Z" || len(parsed.Collections) != 2 || parsed.Queries[0].Query != "find cameras" {
		t.Fatalf("unexpected report JSON: %#v", parsed)
	}
	result := parsed.Queries[0].Results[0]
	if result.Collection != "alpha" || result.AvgScore != 0.875 || result.LatencyMS != 1250 {
		t.Fatalf("unexpected JSON result: %#v", result)
	}

	empty := (&ComparisonReport{Generated: report.Generated}).ToJSON()
	if err := json.Unmarshal([]byte(empty), &map[string]interface{}{}); err != nil {
		t.Fatalf("invalid empty report JSON: %v\n%s", err, empty)
	}
}
