// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestNewContextBuilder(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.3,
		},
	}

	builder := NewContextBuilder(config)
	if builder == nil {
		t.Fatal("NewContextBuilder() returned nil")
	}
	if builder.config != config {
		t.Error("ContextBuilder config not set correctly")
	}
}

func TestBuildContext_EmptyResults(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.3,
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	if context.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", context.Query)
	}
	if len(context.Sources) != 0 {
		t.Errorf("Expected 0 sources, got %d", len(context.Sources))
	}
	if context.TotalResults != 0 {
		t.Errorf("Expected TotalResults 0, got %d", context.TotalResults)
	}
}

func TestBuildContext_WithResults(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0, // Accept all scores
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0, // No truncation
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID:      "doc1",
				Content: "First document content",
				Metadata: map[string]interface{}{
					"source": "test1",
				},
			},
			Score: 0.9,
		},
		{
			Document: vectordb.Document{
				ID:      "doc2",
				Content: "Second document content",
				Metadata: map[string]interface{}{
					"source": "test2",
				},
			},
			Score: 0.7,
		},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	if context.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", context.Query)
	}
	if len(context.Sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(context.Sources))
	}
	if context.TotalResults != 2 {
		t.Errorf("Expected TotalResults 2, got %d", context.TotalResults)
	}

	// Check first source
	if context.Sources[0].Index != 1 {
		t.Errorf("Expected Index 1, got %d", context.Sources[0].Index)
	}
	if context.Sources[0].Content != "First document content" {
		t.Errorf("Expected Content 'First document content', got '%s'", context.Sources[0].Content)
	}
	if context.Sources[0].Score != 0.9 {
		t.Errorf("Expected Score 0.9, got %f", context.Sources[0].Score)
	}
	if context.Sources[0].DocID != "doc1" {
		t.Errorf("Expected DocID 'doc1', got '%s'", context.Sources[0].DocID)
	}

	// Check second source
	if context.Sources[1].Index != 2 {
		t.Errorf("Expected Index 2, got %d", context.Sources[1].Index)
	}
	if context.Sources[1].Content != "Second document content" {
		t.Errorf("Expected Content 'Second document content', got '%s'", context.Sources[1].Content)
	}
}

func TestBuildContext_FilterByRelevance(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.5, // Filter out scores below 0.5
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{ID: "doc1", Content: "High relevance"},
			Score:    0.9,
		},
		{
			Document: vectordb.Document{ID: "doc2", Content: "Low relevance"},
			Score:    0.3,
		},
		{
			Document: vectordb.Document{ID: "doc3", Content: "Medium relevance"},
			Score:    0.6,
		},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	// Should only include doc1 and doc3 (scores >= 0.5)
	if len(context.Sources) != 2 {
		t.Errorf("Expected 2 sources after filtering, got %d", len(context.Sources))
	}
	if context.TotalResults != 3 {
		t.Errorf("Expected TotalResults 3, got %d", context.TotalResults)
	}
}

func TestBuildContext_MaxContextChunks(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  2, // Limit to 2 sources
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "doc1", Content: "First"}, Score: 0.9},
		{Document: vectordb.Document{ID: "doc2", Content: "Second"}, Score: 0.8},
		{Document: vectordb.Document{ID: "doc3", Content: "Third"}, Score: 0.7},
		{Document: vectordb.Document{ID: "doc4", Content: "Fourth"}, Score: 0.6},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	// Should only include first 2 results
	if len(context.Sources) != 2 {
		t.Errorf("Expected 2 sources (max chunks), got %d", len(context.Sources))
	}
	if context.TotalResults != 4 {
		t.Errorf("Expected TotalResults 4, got %d", context.TotalResults)
	}
}

func TestBuildContext_SortByRelevance(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
			SortByRelevance:   true, // Enable sorting
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "doc1", Content: "Low"}, Score: 0.5},
		{Document: vectordb.Document{ID: "doc2", Content: "High"}, Score: 0.9},
		{Document: vectordb.Document{ID: "doc3", Content: "Medium"}, Score: 0.7},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	// Should be sorted by score descending
	if len(context.Sources) != 3 {
		t.Fatalf("Expected 3 sources, got %d", len(context.Sources))
	}

	expectedOrder := []float64{0.9, 0.7, 0.5}
	for i, expected := range expectedOrder {
		if context.Sources[i].Score != expected {
			t.Errorf("Expected source %d to have score %f, got %f", i, expected, context.Sources[i].Score)
		}
	}
}

func TestBuildContext_Deduplicate(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   5,
			MinRelevanceScore:  0.0,
			DeduplicateSources: true, // Enable deduplication
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "doc1", Content: "Unique content 1"}, Score: 0.9},
		{Document: vectordb.Document{ID: "doc1", Content: "Unique content 1"}, Score: 0.8}, // Duplicate ID
		{Document: vectordb.Document{ID: "doc2", Content: "Unique content 2"}, Score: 0.7},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	// Should only have 2 unique sources
	if len(context.Sources) != 2 {
		t.Errorf("Expected 2 sources after deduplication, got %d", len(context.Sources))
	}
}

func TestBuildContext_TruncateContent(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 20, // Truncate to 20 chars
		},
	}

	builder := NewContextBuilder(config)
	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID:      "doc1",
				Content: "This is a very long content that should be truncated",
			},
			Score: 0.9,
		},
	}

	context, err := builder.BuildContext("test query", results)
	if err != nil {
		t.Fatalf("BuildContext() failed: %v", err)
	}

	if len(context.Sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(context.Sources))
	}

	expected := "This is a very long ..."
	if context.Sources[0].Content != expected {
		t.Errorf("Expected truncated content '%s', got '%s'", expected, context.Sources[0].Content)
	}
}

func TestBuildContext_ExtractContent(t *testing.T) {
	tests := []struct {
		name     string
		doc      vectordb.Document
		expected string
	}{
		{
			name:     "Content field",
			doc:      vectordb.Document{Content: "Content text", Text: "Text field", URL: "http://example.com"},
			expected: "Content text",
		},
		{
			name:     "Text field (no Content)",
			doc:      vectordb.Document{Text: "Text field", URL: "http://example.com"},
			expected: "Text field",
		},
		{
			name:     "URL field (no Content or Text)",
			doc:      vectordb.Document{URL: "http://example.com"},
			expected: "[Document URL: http://example.com]",
		},
		{
			name:     "Empty document",
			doc:      vectordb.Document{},
			expected: "[Empty document]",
		},
	}

	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := []*vectordb.QueryResult{
				{Document: tt.doc, Score: 0.9},
			}

			context, err := builder.BuildContext("test", results)
			if err != nil {
				t.Fatalf("BuildContext() failed: %v", err)
			}

			if len(context.Sources) != 1 {
				t.Fatalf("Expected 1 source, got %d", len(context.Sources))
			}

			if context.Sources[0].Content != tt.expected {
				t.Errorf("Expected content '%s', got '%s'", tt.expected, context.Sources[0].Content)
			}
		})
	}
}

func TestFormatContextForPrompt(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	context := &QueryContext{
		Query: "test query",
		Sources: []SourceContext{
			{Index: 1, Content: "First source", Score: 0.9, DocID: "doc1"},
			{Index: 2, Content: "Second source", Score: 0.7, DocID: "doc2"},
		},
	}

	formatted := builder.FormatContextForPrompt(context)

	// Check that formatted string contains key elements
	if !contains(formatted, "Query: test query") {
		t.Error("Formatted context should contain query")
	}
	if !contains(formatted, "[1]") {
		t.Error("Formatted context should contain citation [1]")
	}
	if !contains(formatted, "First source") {
		t.Error("Formatted context should contain first source content")
	}
	if !contains(formatted, "0.900") {
		t.Error("Formatted context should contain score")
	}
}

func TestFormatContextForPromptWithCitations_Numeric(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
			CitationFormat:    "numeric",
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	context := &QueryContext{
		Query: "test query",
		Sources: []SourceContext{
			{Index: 1, Content: "First source", Score: 0.9, DocID: "doc1"},
		},
	}

	formatted := builder.FormatContextForPromptWithCitations(context)

	if !contains(formatted, "Use [1], [2]") {
		t.Error("Formatted context should contain numeric citation instructions")
	}
	if !contains(formatted, "First source") {
		t.Error("Formatted context should contain source content")
	}
}

func TestFormatContextForPromptWithCitations_StrictMode(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
			CitationFormat:    "numeric",
			StrictMode:        true,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	builder := NewContextBuilder(config)
	context := &QueryContext{
		Query: "test query",
		Sources: []SourceContext{
			{Index: 1, Content: "Test", Score: 0.9, DocID: "doc1"},
		},
	}

	formatted := builder.FormatContextForPromptWithCitations(context)

	if !contains(formatted, "STRICT MODE ENABLED") {
		t.Error("Formatted context should contain strict mode warning")
	}
	if !contains(formatted, "Only use information from the provided sources") {
		t.Error("Formatted context should contain strict mode instructions")
	}
}

func TestGetSourceByIndex(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
	}

	builder := NewContextBuilder(config)
	context := &QueryContext{
		Sources: []SourceContext{
			{Index: 1, Content: "First", Score: 0.9, DocID: "doc1"},
			{Index: 2, Content: "Second", Score: 0.7, DocID: "doc2"},
		},
	}

	// Valid index
	source, err := builder.GetSourceByIndex(context, 1)
	if err != nil {
		t.Errorf("GetSourceByIndex(1) failed: %v", err)
	}
	if source.Content != "First" {
		t.Errorf("Expected content 'First', got '%s'", source.Content)
	}

	// Invalid index (too low)
	_, err = builder.GetSourceByIndex(context, 0)
	if err == nil {
		t.Error("GetSourceByIndex(0) should return error")
	}

	// Invalid index (too high)
	_, err = builder.GetSourceByIndex(context, 3)
	if err == nil {
		t.Error("GetSourceByIndex(3) should return error")
	}
}

func TestGetSourceCount(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
	}

	builder := NewContextBuilder(config)
	context := &QueryContext{
		Sources: []SourceContext{
			{Index: 1, Content: "First", Score: 0.9, DocID: "doc1"},
			{Index: 2, Content: "Second", Score: 0.7, DocID: "doc2"},
		},
	}

	count := builder.GetSourceCount(context)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestBuildContextFromQueryResults_ConvenienceFunction(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			TruncateSources: 0,
		},
	}

	results := []*vectordb.QueryResult{
		{Document: vectordb.Document{ID: "doc1", Content: "Test"}, Score: 0.9},
	}

	context, err := BuildContextFromQueryResults("test query", results, config)
	if err != nil {
		t.Fatalf("BuildContextFromQueryResults() failed: %v", err)
	}

	if context.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", context.Query)
	}
	if len(context.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(context.Sources))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
