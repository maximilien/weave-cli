// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildContext_CrossVDB_DeduplicationKeepsHighestScore tests that when the same
// document appears from multiple VDBs with different scores, the highest-scored
// version is kept after deduplication.
//
// This is a regression test for the bug where cross-VDB queries only showed results
// from the first VDB because deduplication happened before sorting.
func TestBuildContext_CrossVDB_DeduplicationKeepsHighestScore(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   5,
			MinRelevanceScore:  0.3,
			DeduplicateSources: true,  // Enable deduplication
			SortByRelevance:    true,  // Enable sorting
		},
	}

	builder := NewContextBuilder(config)

	// Simulate the same document from 3 different VDBs with different scores
	// This mimics querying WeaveDocs:mongodb-cloud, WeaveDocs:weaviate-cloud, WeaveDocs:milvus-cloud
	results := []*vectordb.QueryResult{
		// MongoDB results (first in aggregation order, but NOT highest scores)
		{
			Document: vectordb.Document{
				ID:      "mongodb-doc1",
				Content: "Vector databases are specialized databases optimized for vector search",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.456, // Lower than weaviate version
		},
		{
			Document: vectordb.Document{
				ID:      "mongodb-doc2",
				Content: "Semantic search uses vector embeddings for similarity",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.441,
		},
		{
			Document: vectordb.Document{
				ID:      "mongodb-doc3",
				Content: "RAG combines retrieval with generation",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.430,
		},

		// Weaviate results (SAME content as some mongodb docs, but HIGHER scores!)
		{
			Document: vectordb.Document{
				ID:      "weaviate-doc1",
				Content: "Vector databases are specialized databases optimized for vector search", // Same as mongodb-doc1
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.823, // HIGHEST score! Should be kept after deduplication
		},
		{
			Document: vectordb.Document{
				ID:      "weaviate-doc2",
				Content: "Weaviate is a cloud-native vector database",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.678,
		},

		// Milvus results
		{
			Document: vectordb.Document{
				ID:      "milvus-doc1",
				Content: "Milvus provides high-performance vector similarity search",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "milvus-cloud",
				},
			},
			Score: 0.712,
		},
		{
			Document: vectordb.Document{
				ID:      "milvus-doc2",
				Content: "Semantic search uses vector embeddings for similarity", // Same as mongodb-doc2
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "milvus-cloud",
				},
			},
			Score: 0.889, // HIGHEST score for this content! Should be kept
		},
	}

	// Build context
	context, err := builder.BuildContext("vector database", results)
	require.NoError(t, err)
	require.NotNil(t, context)

	// Should have deduplicated results sorted by score
	// Expected order after sort + dedupe (keeping highest score for each unique content):
	// 1. milvus-doc2 (0.889) - "Semantic search..."
	// 2. weaviate-doc1 (0.823) - "Vector databases..."
	// 3. milvus-doc1 (0.712) - "Milvus provides..."
	// 4. weaviate-doc2 (0.678) - "Weaviate is..."
	// 5. mongodb-doc3 (0.430) - "RAG combines..." (unique, only from mongodb)

	// With max_context_chunks: 5, all 5 should be included
	assert.Len(t, context.Sources, 5, "Should have 5 sources after deduplication and limiting")

	// Debug: Print what we actually got
	t.Logf("Got %d sources:", len(context.Sources))
	for i, source := range context.Sources {
		vdb := "unknown"
		if v, ok := source.Metadata["_vdb"].(string); ok {
			vdb = v
		}
		t.Logf("  [%d] Score: %.3f, VDB: %s, Content: %.50s...", i+1, source.Score, vdb, source.Content)
	}

	// Verify the highest-scored version of each document was kept
	// First source should be milvus-doc2 (highest overall score)
	assert.Equal(t, 0.889, context.Sources[0].Score, "First source should have highest score (milvus-doc2)")
	assert.Contains(t, context.Sources[0].Content, "Semantic search")
	assert.Equal(t, "milvus-cloud", context.Sources[0].Metadata["_vdb"],
		"First source should be from milvus (highest score for this content)")

	// Second source should be weaviate-doc1
	assert.Equal(t, 0.823, context.Sources[1].Score, "Second source should be weaviate-doc1")
	assert.Contains(t, context.Sources[1].Content, "Vector databases")
	assert.Equal(t, "weaviate-cloud", context.Sources[1].Metadata["_vdb"],
		"Should keep weaviate version (higher score than mongodb version of same content)")

	// Verify we have sources from multiple VDBs, not just the first one
	vdbs := make(map[string]bool)
	for _, source := range context.Sources {
		if vdb, ok := source.Metadata["_vdb"].(string); ok {
			vdbs[vdb] = true
		}
	}

	assert.GreaterOrEqual(t, len(vdbs), 2, "Should have results from at least 2 different VDBs")
	assert.True(t, vdbs["milvus-cloud"], "Should have results from milvus-cloud")
	assert.True(t, vdbs["weaviate-cloud"], "Should have results from weaviate-cloud")
}

// TestBuildContext_CrossVDB_SortingOrder tests that results from multiple VDBs
// are properly sorted by score regardless of aggregation order
func TestBuildContext_CrossVDB_SortingOrder(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   3,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false, // Disable deduplication to test pure sorting
			SortByRelevance:    true,
		},
	}

	builder := NewContextBuilder(config)

	// Results in aggregation order: mongodb, weaviate, milvus
	// But scores are: milvus > weaviate > mongodb
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID:      "mongodb-1",
				Content: "MongoDB doc",
				Metadata: map[string]interface{}{
					"_vdb": "mongodb-cloud",
				},
			},
			Score: 0.456, // Lowest
		},
		{
			Document: vectordb.Document{
				ID:      "weaviate-1",
				Content: "Weaviate doc",
				Metadata: map[string]interface{}{
					"_vdb": "weaviate-cloud",
				},
			},
			Score: 0.678, // Middle
		},
		{
			Document: vectordb.Document{
				ID:      "milvus-1",
				Content: "Milvus doc",
				Metadata: map[string]interface{}{
					"_vdb": "milvus-cloud",
				},
			},
			Score: 0.823, // Highest
		},
	}

	context, err := builder.BuildContext("test", results)
	require.NoError(t, err)

	// Should be sorted by score descending
	assert.Len(t, context.Sources, 3)
	assert.Equal(t, 0.823, context.Sources[0].Score, "First should be highest score (milvus)")
	assert.Equal(t, 0.678, context.Sources[1].Score, "Second should be middle score (weaviate)")
	assert.Equal(t, 0.456, context.Sources[2].Score, "Third should be lowest score (mongodb)")

	// Verify VDBs are in score order, not aggregation order
	assert.Equal(t, "milvus-cloud", context.Sources[0].Metadata["_vdb"])
	assert.Equal(t, "weaviate-cloud", context.Sources[1].Metadata["_vdb"])
	assert.Equal(t, "mongodb-cloud", context.Sources[2].Metadata["_vdb"])
}

// TestBuildContext_CrossVDB_SameIDDifferentVDBs tests the bug where the same document
// with the same ID exists in multiple VDBs with different scores. Deduplication should
// keep the highest-scored version, not just the first one encountered.
//
// This is the REAL bug the user reported!
func TestBuildContext_CrossVDB_SameIDDifferentVDBs(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   5,
			MinRelevanceScore:  0.0,
			DeduplicateSources: true,  // Enable deduplication
			SortByRelevance:    true,  // Enable sorting
		},
	}

	builder := NewContextBuilder(config)

	// Simulate the SAME documents (same IDs) synced across 3 VDBs with different scores
	// This is what happens when you query WeaveDocs:mongodb-cloud WeaveDocs:weaviate-cloud WeaveDocs:milvus-cloud
	results := []*vectordb.QueryResult{
		// MongoDB results (first in aggregation order)
		{
			Document: vectordb.Document{
				ID:      "doc-001", // SAME ID across all VDBs
				Content: "Vector databases are specialized for similarity search",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.456, // Lower score
		},
		{
			Document: vectordb.Document{
				ID:      "doc-002",
				Content: "Semantic search uses embeddings",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.441,
		},
		{
			Document: vectordb.Document{
				ID:      "doc-003",
				Content: "RAG combines retrieval and generation",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.430,
		},

		// Weaviate results (SAME IDs, HIGHER scores!)
		{
			Document: vectordb.Document{
				ID:      "doc-001", // SAME ID as mongodb version
				Content: "Vector databases are specialized for similarity search",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.823, // MUCH higher score!
		},
		{
			Document: vectordb.Document{
				ID:      "doc-002", // SAME ID
				Content: "Semantic search uses embeddings",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.789,
		},

		// Milvus results (SAME IDs, also HIGHER scores!)
		{
			Document: vectordb.Document{
				ID:      "doc-001", // SAME ID
				Content: "Vector databases are specialized for similarity search",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "milvus-cloud",
				},
			},
			Score: 0.712,
		},
		{
			Document: vectordb.Document{
				ID:      "doc-003", // SAME ID
				Content: "RAG combines retrieval and generation",
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "milvus-cloud",
				},
			},
			Score: 0.567,
		},
	}

	// Build context
	context, err := builder.BuildContext("vector database", results)
	require.NoError(t, err)
	require.NotNil(t, context)

	// Debug: Print what we got
	t.Logf("Got %d sources:", len(context.Sources))
	for i, source := range context.Sources {
		vdb := "unknown"
		if v, ok := source.Metadata["_vdb"].(string); ok {
			vdb = v
		}
		t.Logf("  [%d] ID: %s, Score: %.3f, VDB: %s", i+1, source.DocID, source.Score, vdb)
	}

	// Should have 3 unique documents (doc-001, doc-002, doc-003)
	assert.Len(t, context.Sources, 3, "Should have 3 unique documents after deduplication")

	// CRITICAL: For each document, should keep the HIGHEST-scored version, not just the first!
	// doc-001: weaviate has 0.823 (highest), mongodb has 0.456, milvus has 0.712
	assert.Equal(t, 0.823, context.Sources[0].Score, "doc-001 should use weaviate score (highest)")
	assert.Equal(t, "weaviate-cloud", context.Sources[0].Metadata["_vdb"],
		"doc-001 should be from weaviate (highest score), NOT mongodb (first in aggregation)")

	// doc-002: weaviate has 0.789 (higher than mongodb's 0.441)
	assert.Equal(t, 0.789, context.Sources[1].Score, "doc-002 should use weaviate score")
	assert.Equal(t, "weaviate-cloud", context.Sources[1].Metadata["_vdb"],
		"doc-002 should be from weaviate (highest score)")

	// doc-003: milvus has 0.567 (higher than mongodb's 0.430)
	assert.Equal(t, 0.567, context.Sources[2].Score, "doc-003 should use milvus score")
	assert.Equal(t, "milvus-cloud", context.Sources[2].Metadata["_vdb"],
		"doc-003 should be from milvus (highest score)")

	// Verify we're NOT just keeping mongodb results (the first VDB in aggregation order)
	vdbs := make(map[string]bool)
	for _, source := range context.Sources {
		if vdb, ok := source.Metadata["_vdb"].(string); ok {
			vdbs[vdb] = true
		}
	}

	t.Logf("VDBs represented: %v", vdbs)
	assert.True(t, vdbs["weaviate-cloud"] || vdbs["milvus-cloud"],
		"Should have results from weaviate or milvus, not ONLY from mongodb")
	assert.False(t, len(vdbs) == 1 && vdbs["mongodb-cloud"],
		"Should NOT have only mongodb results (the bug!)")
}

