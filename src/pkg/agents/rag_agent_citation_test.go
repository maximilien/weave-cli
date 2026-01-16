// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRAGAgent_FormatOutput_CrossVDBCitations tests that citations include
// document ID, score, VDB, and collection name for cross-VDB queries
func TestRAGAgent_FormatOutput_CrossVDBCitations(t *testing.T) {
	config := &CustomAgentConfig{
		Name:         "test-rag-agent",
		Type:         "rag",
		SystemPrompt: "Test system prompt",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.0,
			DeduplicateSources: true,
			SortByRelevance:    true,
		},
		Output: CustomAgentOutputConfig{
			Format:          "text",
			IncludeMetadata: false,
			ShowConfidence:  false,
			ShowSources:     true,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 60,
		},
	}

	agent, err := NewRAGAgent(config, &MockLLMClient{})
	require.NoError(t, err)

	// Create RAG output with cross-VDB citations
	output := &RAGOutput{
		Answer: "Vector databases provide efficient similarity search capabilities.",
		Sources: []SourceCitation{
			{
				Index: 1,
				DocID: "doc-001",
				Score: 0.823,
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
			{
				Index: 2,
				DocID: "doc-002",
				Score: 0.789,
				Metadata: map[string]interface{}{
					"_collection": "WeaveDocs",
					"_vdb":        "milvus-cloud",
				},
			},
			{
				Index: 3,
				DocID: "doc-003",
				Score: 0.712,
				Metadata: map[string]interface{}{
					"_collection": "AuctionDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
		},
	}

	// Format output
	formatted, err := agent.FormatOutput(output)
	require.NoError(t, err)

	t.Logf("Formatted output:\n%s", formatted)

	// Verify the new citation format includes all components
	// Expected format: [1] ID: doc-001, Score: 82.3%, VDB: weaviate-cloud, Collection: WeaveDocs

	// Check first citation
	assert.Contains(t, formatted, "[1] ID: doc-001, Score: 82.3%, VDB: weaviate-cloud, Collection: WeaveDocs",
		"Citation should include ID, score, VDB, and collection")

	// Check second citation
	assert.Contains(t, formatted, "[2] ID: doc-002, Score: 78.9%, VDB: milvus-cloud, Collection: WeaveDocs",
		"Second citation should show different VDB")

	// Check third citation with different collection
	assert.Contains(t, formatted, "[3] ID: doc-003, Score: 71.2%, VDB: mongodb-cloud, Collection: AuctionDocs",
		"Third citation should show different collection and VDB")

	// Verify all citations show different VDBs (proving cross-VDB query)
	assert.Contains(t, formatted, "weaviate-cloud", "Should show weaviate VDB")
	assert.Contains(t, formatted, "milvus-cloud", "Should show milvus VDB")
	assert.Contains(t, formatted, "mongodb-cloud", "Should show mongodb VDB")
}

// TestRAGAgent_FormatOutput_MarkdownCrossVDBCitations tests markdown format citations
func TestRAGAgent_FormatOutput_MarkdownCrossVDBCitations(t *testing.T) {
	config := &CustomAgentConfig{
		Name:         "test-rag-agent",
		Type:         "rag",
		SystemPrompt: "Test system prompt",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.0,
			DeduplicateSources: true,
			SortByRelevance:    true,
		},
		Output: CustomAgentOutputConfig{
			Format:          "markdown",
			IncludeMetadata: false,
			ShowConfidence:  false,
			ShowSources:     true,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 60,
		},
	}

	agent, err := NewRAGAgent(config, &MockLLMClient{})
	require.NoError(t, err)

	output := &RAGOutput{
		Answer: "Vector databases provide efficient similarity search capabilities.",
		Sources: []SourceCitation{
			{
				Index:   1,
				DocID:   "abc-123",
				Score:   0.856,
				Content: "Weaviate is a cloud-native vector database...",
				Metadata: map[string]interface{}{
					"_collection": "TechDocs",
					"_vdb":        "weaviate-cloud",
				},
			},
		},
	}

	formatted, err := agent.FormatOutput(output)
	require.NoError(t, err)

	t.Logf("Markdown formatted output:\n%s", formatted)

	// Verify markdown format with backticks for ID
	assert.Contains(t, formatted, "**[1]** ID: `abc-123`, Score: 85.6%, VDB: weaviate-cloud, Collection: TechDocs",
		"Markdown citation should include ID with backticks, score, VDB, and collection")

	// Verify content is shown on separate line
	assert.Contains(t, formatted, "- Weaviate is a cloud-native vector database",
		"Content should be shown as a bullet point")
}

// TestRAGAgent_FormatOutput_CitationWithoutDocID tests citation when DocID is missing
func TestRAGAgent_FormatOutput_CitationWithoutDocID(t *testing.T) {
	config := &CustomAgentConfig{
		Name:         "test-rag-agent",
		Type:         "rag",
		SystemPrompt: "Test system prompt",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
		Output: CustomAgentOutputConfig{
			Format:      "text",
			ShowSources: true,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 60,
		},
	}

	agent, err := NewRAGAgent(config, &MockLLMClient{})
	require.NoError(t, err)

	output := &RAGOutput{
		Answer: "Test answer",
		Sources: []SourceCitation{
			{
				Index: 1,
				DocID: "", // No document ID
				Score: 0.75,
				Metadata: map[string]interface{}{
					"_collection": "TestDocs",
					"_vdb":        "test-vdb",
				},
			},
		},
	}

	formatted, err := agent.FormatOutput(output)
	require.NoError(t, err)

	// Should show citation without ID field
	assert.Contains(t, formatted, "[1] Score: 75.0%, VDB: test-vdb, Collection: TestDocs",
		"Citation without DocID should still show score, VDB, and collection")

	// Should NOT contain "ID:" when no DocID
	assert.False(t, strings.Contains(formatted, "ID:"), "Should not show ID field when DocID is empty")
}
