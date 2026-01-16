// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSortByRelevance_RandomTieBreaking tests that equal scores are randomly shuffled
// to avoid bias toward the first VDB in aggregation order
func TestSortByRelevance_RandomTieBreaking(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    true,
		},
	}

	builder := NewContextBuilder(config)

	// Create results with SAME scores from different VDBs
	// If sorting is deterministic, these would always be in aggregation order
	// With random tie-breaking, order should vary
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID:      "mongodb-1",
				Content: "Document from MongoDB",
				Metadata: map[string]interface{}{
					"_vdb": "mongodb-cloud",
				},
			},
			Score: 0.75, // Same score
		},
		{
			Document: vectordb.Document{
				ID:      "weaviate-1",
				Content: "Document from Weaviate",
				Metadata: map[string]interface{}{
					"_vdb": "weaviate-cloud",
				},
			},
			Score: 0.75, // Same score
		},
		{
			Document: vectordb.Document{
				ID:      "milvus-1",
				Content: "Document from Milvus",
				Metadata: map[string]interface{}{
					"_vdb": "milvus-cloud",
				},
			},
			Score: 0.75, // Same score
		},
	}

	// Run sorting multiple times and track which VDB appears first
	firstPositions := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		// Create fresh copy for each iteration
		testResults := make([]*vectordb.QueryResult, len(results))
		copy(testResults, results)

		// Sort with random tie-breaking
		builder.sortByRelevance(testResults)

		// Track which VDB ended up first
		firstVDB := testResults[0].Document.Metadata["_vdb"].(string)
		firstPositions[firstVDB]++
	}

	t.Logf("First position distribution over %d iterations:", iterations)
	for vdb, count := range firstPositions {
		percentage := float64(count) / float64(iterations) * 100
		t.Logf("  %s: %d times (%.1f%%)", vdb, count, percentage)
	}

	// All three VDBs should appear first at least once (very high probability)
	// If sorting was deterministic, only one VDB would ever be first
	assert.GreaterOrEqual(t, len(firstPositions), 2,
		"At least 2 different VDBs should appear first (proves randomization)")

	// Each VDB should appear first roughly 1/3 of the time (with tolerance)
	// We use a loose tolerance since it's random
	for vdb, count := range firstPositions {
		percentage := float64(count) / float64(iterations) * 100
		assert.GreaterOrEqual(t, percentage, 10.0,
			"VDB %s should appear first at least 10%% of the time (got %.1f%%)", vdb, percentage)
	}
}

// TestSortByRelevance_RandomTieBreakerWithDifferentScores verifies that
// random tie-breaking only affects equal scores, not different scores
func TestSortByRelevance_RandomTieBreakerWithDifferentScores(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    true,
		},
	}

	builder := NewContextBuilder(config)

	// Mix of different scores and equal scores
	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "low-1"}, Score: 0.60},
		{Document: vectordb.Document{ID: "high-1"}, Score: 0.90},
		{Document: vectordb.Document{ID: "mid-tie-1"}, Score: 0.75},
		{Document: vectordb.Document{ID: "mid-tie-2"}, Score: 0.75}, // Tie
		{Document: vectordb.Document{ID: "mid-tie-3"}, Score: 0.75}, // Tie
		{Document: vectordb.Document{ID: "low-2"}, Score: 0.60},     // Tie with low-1
	}

	// Run multiple times to verify behavior
	for i := 0; i < 10; i++ {
		testResults := make([]*vectordb.QueryResult, len(results))
		copy(testResults, results)

		builder.sortByRelevance(testResults)

		// Highest score should ALWAYS be first (not random)
		assert.Equal(t, 0.90, testResults[0].Score, "Highest score should always be first")
		assert.Equal(t, "high-1", testResults[0].Document.ID)

		// Next 3 should all be 0.75 (order may vary randomly)
		assert.Equal(t, 0.75, testResults[1].Score, "Second should be 0.75 (tied for 2nd)")
		assert.Equal(t, 0.75, testResults[2].Score, "Third should be 0.75 (tied for 2nd)")
		assert.Equal(t, 0.75, testResults[3].Score, "Fourth should be 0.75 (tied for 2nd)")

		// Last 2 should be 0.60 (order may vary randomly)
		assert.Equal(t, 0.60, testResults[4].Score, "Fifth should be 0.60 (tied for last)")
		assert.Equal(t, 0.60, testResults[5].Score, "Sixth should be 0.60 (tied for last)")
	}
}

// TestSortByRelevance_CrossVDBFairness tests that when multiple VDBs return
// results with the same score, no single VDB is favored in the final ranking
func TestSortByRelevance_CrossVDBFairness(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   3, // Only take top 3
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    true,
		},
	}

	builder := NewContextBuilder(config)

	// 3 VDBs all return results with score 0.80
	// Without random tie-breaking, aggregation order determines selection
	// With random tie-breaking, each VDB has fair chance
	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "mongodb-1", Metadata: map[string]interface{}{"_vdb": "mongodb"}}, Score: 0.80},
		{Document: vectordb.Document{ID: "weaviate-1", Metadata: map[string]interface{}{"_vdb": "weaviate"}}, Score: 0.80},
		{Document: vectordb.Document{ID: "milvus-1", Metadata: map[string]interface{}{"_vdb": "milvus"}}, Score: 0.80},
	}

	// Track which VDBs appear in top 3 over multiple runs
	vdbCounts := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		testResults := make([]*vectordb.QueryResult, len(results))
		copy(testResults, results)

		// Build context (which sorts and limits to top 3)
		context, err := builder.BuildContext("test", testResults)
		require.NoError(t, err)

		// Count VDB representation in results
		for _, source := range context.Sources {
			vdb := source.Metadata["_vdb"].(string)
			vdbCounts[vdb]++
		}
	}

	t.Logf("VDB representation in top 3 over %d iterations:", iterations)
	for vdb, count := range vdbCounts {
		percentage := float64(count) / float64(iterations*3) * 100
		t.Logf("  %s: %d times (%.1f%%)", vdb, count, percentage)
	}

	// Each VDB should appear roughly 1/3 of the time (fair distribution)
	// With loose tolerance since it's random
	for _, vdb := range []string{"mongodb", "weaviate", "milvus"} {
		count := vdbCounts[vdb]
		percentage := float64(count) / float64(iterations*3) * 100
		assert.GreaterOrEqual(t, percentage, 20.0,
			"VDB %s should appear at least 20%% of the time for fair distribution (got %.1f%%)",
			vdb, percentage)
	}
}

// TestSortByRelevance_EpsilonBasedShuffling tests that scores within epsilon
// are treated as equal and randomly shuffled for fairness
func TestSortByRelevance_EpsilonBasedShuffling(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    true,
		},
	}

	builder := NewContextBuilder(config)

	// Create results with scores that are VERY close (within epsilon 0.001)
	// These should be shuffled randomly despite not being exactly equal
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID: "mongodb-1",
				Metadata: map[string]interface{}{
					"_vdb": "mongodb-cloud",
				},
			},
			Score: 0.82300, // Base score
		},
		{
			Document: vectordb.Document{
				ID: "weaviate-1",
				Metadata: map[string]interface{}{
					"_vdb": "weaviate-cloud",
				},
			},
			Score: 0.82305, // Within epsilon (diff: 0.00005)
		},
		{
			Document: vectordb.Document{
				ID: "milvus-1",
				Metadata: map[string]interface{}{
					"_vdb": "milvus-cloud",
				},
			},
			Score: 0.82298, // Within epsilon (diff: 0.00002)
		},
		{
			Document: vectordb.Document{
				ID: "qdrant-1",
				Metadata: map[string]interface{}{
					"_vdb": "qdrant-cloud",
				},
			},
			Score: 0.81500, // OUTSIDE epsilon (diff: 0.008) - should always be last
		},
	}

	// Track which VDB appears first among the "tied" results
	firstPositions := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		testResults := make([]*vectordb.QueryResult, len(results))
		copy(testResults, results)

		builder.sortByRelevance(testResults)

		// First 3 should all be within epsilon (order varies randomly)
		assert.InDelta(t, 0.823, testResults[0].Score, 0.001,
			"First 3 should all be around 0.823 (within epsilon)")
		assert.InDelta(t, 0.823, testResults[1].Score, 0.001,
			"First 3 should all be around 0.823 (within epsilon)")
		assert.InDelta(t, 0.823, testResults[2].Score, 0.001,
			"First 3 should all be around 0.823 (within epsilon)")

		// Last one should ALWAYS be the 0.815 (outside epsilon)
		assert.Equal(t, 0.81500, testResults[3].Score,
			"Score outside epsilon should always be last")
		assert.Equal(t, "qdrant-1", testResults[3].Document.ID)

		// Track which VDB is first among the tied scores
		firstVDB := testResults[0].Document.Metadata["_vdb"].(string)
		firstPositions[firstVDB]++
	}

	t.Logf("First position distribution for epsilon-tied scores over %d iterations:", iterations)
	for vdb, count := range firstPositions {
		percentage := float64(count) / float64(iterations) * 100
		t.Logf("  %s: %d times (%.1f%%)", vdb, count, percentage)
	}

	// All VDBs with scores within epsilon should appear first at least once
	assert.GreaterOrEqual(t, len(firstPositions), 2,
		"At least 2 different VDBs should appear first among epsilon-tied scores")

	// VDB outside epsilon should NEVER appear first
	assert.Equal(t, 0, firstPositions["qdrant-cloud"],
		"VDB with score outside epsilon should never be first")
}
