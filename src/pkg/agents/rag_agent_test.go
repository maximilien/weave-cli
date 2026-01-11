// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// MockLLMClient is a mock implementation of llm.Client for testing
type MockLLMClient struct {
	response string
	err      error
	metrics  *llm.Metrics
}

func NewMockLLMClient(response string, err error) *MockLLMClient {
	return &MockLLMClient{
		response: response,
		err:      err,
		metrics:  &llm.Metrics{},
	}
}

func (m *MockLLMClient) Complete(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	m.metrics.Invocations++
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockLLMClient) CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...llm.Option) (interface{}, error) {
	m.metrics.Invocations++
	if m.err != nil {
		return nil, m.err
	}
	return schema, nil
}

func (m *MockLLMClient) GetMetrics() *llm.Metrics {
	return m.metrics
}

func TestNewRAGAgent(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-rag-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		SystemPrompt: "You are a helpful assistant.",
		Response: CustomAgentResponseConfig{
			IncludeReferences: true,
			MaxContextChunks:  5,
			MinRelevanceScore: 0.3,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test response", nil)

	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	if agent.Name() != "test-rag-agent" {
		t.Errorf("Expected name 'test-rag-agent', got '%s'", agent.Name())
	}
	if agent.GetType() != "rag" {
		t.Errorf("Expected type 'rag', got '%s'", agent.GetType())
	}
	if !agent.IsRAGType() {
		t.Error("Expected IsRAGType() to return true")
	}
}

func TestNewRAGAgent_NilConfig(t *testing.T) {
	mockClient := NewMockLLMClient("Test", nil)
	_, err := NewRAGAgent(nil, mockClient)
	if err == nil {
		t.Error("Expected error for nil config, got nil")
	}
}

func TestNewRAGAgent_NilClient(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
	}
	config.applyDefaults()

	_, err := NewRAGAgent(config, nil)
	if err == nil {
		t.Error("Expected error for nil client, got nil")
	}
}

func TestRAGAgent_Execute(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-rag-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		SystemPrompt: "You are a helpful assistant.",
		Response: CustomAgentResponseConfig{
			IncludeReferences: true,
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
		Output: CustomAgentOutputConfig{
			Format:          "json",
			ShowSources:     true,
			IncludeMetadata: true,
			ShowConfidence:  true,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("The answer is 42, based on the sources [1] and [2].", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	input := &RAGInput{
		Query: "What is the answer?",
		Results: []*vectordb.QueryResult{
			{
				Document: vectordb.Document{
					ID:      "doc1",
					Content: "The answer to life, the universe, and everything is 42.",
					Metadata: map[string]interface{}{
						"source": "book",
					},
				},
				Score: 0.95,
			},
			{
				Document: vectordb.Document{
					ID:      "doc2",
					Content: "Deep Thought calculated 42 as the answer.",
					Metadata: map[string]interface{}{
						"source": "movie",
					},
				},
				Score: 0.85,
			},
		},
	}

	result, err := agent.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	output, ok := result.(*RAGOutput)
	if !ok {
		t.Fatalf("Expected *RAGOutput, got %T", result)
	}

	if output.Answer == "" {
		t.Error("Expected non-empty answer")
	}
	if !strings.Contains(output.Answer, "42") {
		t.Errorf("Expected answer to contain '42', got '%s'", output.Answer)
	}
	if len(output.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(output.Sources))
	}
	if output.Metadata == nil {
		t.Error("Expected metadata to be present")
	}
	if output.Confidence == 0 {
		t.Error("Expected non-zero confidence")
	}
}

func TestRAGAgent_Execute_InvalidInput(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	// Pass wrong input type
	_, err = agent.Execute(context.Background(), "invalid input")
	if err == nil {
		t.Error("Expected error for invalid input type, got nil")
	}
}

func TestRAGAgent_FormatOutput_JSON(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Output: CustomAgentOutputConfig{
			Format: "json",
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	output := &RAGOutput{
		Answer: "Test answer",
		Sources: []SourceCitation{
			{Index: 1, Score: 0.9, DocID: "doc1"},
		},
	}

	formatted, err := agent.FormatOutput(output)
	if err != nil {
		t.Fatalf("FormatOutput() failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(formatted), &parsed); err != nil {
		t.Errorf("FormatOutput() did not produce valid JSON: %v", err)
	}

	if !strings.Contains(formatted, "Test answer") {
		t.Error("JSON output should contain the answer")
	}
}

func TestRAGAgent_FormatOutput_Markdown(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Output: CustomAgentOutputConfig{
			Format:      "markdown",
			ShowSources: true,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	output := &RAGOutput{
		Answer: "Test answer",
		Sources: []SourceCitation{
			{Index: 1, Score: 0.9, DocID: "doc1", Content: "Source content"},
		},
	}

	formatted, err := agent.FormatOutput(output)
	if err != nil {
		t.Fatalf("FormatOutput() failed: %v", err)
	}

	if !strings.Contains(formatted, "## Answer") {
		t.Error("Markdown output should contain '## Answer' header")
	}
	if !strings.Contains(formatted, "## Sources") {
		t.Error("Markdown output should contain '## Sources' header")
	}
	if !strings.Contains(formatted, "Test answer") {
		t.Error("Markdown output should contain the answer")
	}
}

func TestRAGAgent_FormatOutput_Text(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Output: CustomAgentOutputConfig{
			Format:      "text",
			ShowSources: true,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	output := &RAGOutput{
		Answer: "Test answer",
		Sources: []SourceCitation{
			{Index: 1, Score: 0.9, DocID: "doc1"},
		},
	}

	formatted, err := agent.FormatOutput(output)
	if err != nil {
		t.Fatalf("FormatOutput() failed: %v", err)
	}

	if !strings.Contains(formatted, "Test answer") {
		t.Error("Text output should contain the answer")
	}
	if !strings.Contains(formatted, "Sources:") {
		t.Error("Text output should contain 'Sources:' label")
	}
}

func TestRAGAgent_CalculateConfidence(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
		Response: CustomAgentResponseConfig{
			MaxContextChunks: 5,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	// Test with high-score sources
	contextHigh := &QueryContext{
		TotalResults: 10,
		Sources: []SourceContext{
			{Score: 0.9},
			{Score: 0.85},
			{Score: 0.8},
		},
	}
	confidenceHigh := agent.calculateConfidence(contextHigh)
	if confidenceHigh <= 0 || confidenceHigh > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidenceHigh)
	}

	// Test with low-score sources
	contextLow := &QueryContext{
		TotalResults: 10,
		Sources: []SourceContext{
			{Score: 0.3},
			{Score: 0.2},
		},
	}
	confidenceLow := agent.calculateConfidence(contextLow)
	if confidenceLow <= 0 || confidenceLow > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidenceLow)
	}

	// High confidence should be greater than low confidence
	if confidenceHigh <= confidenceLow {
		t.Errorf("High-score confidence (%f) should be greater than low-score confidence (%f)", confidenceHigh, confidenceLow)
	}

	// Test with no sources
	contextEmpty := &QueryContext{
		TotalResults: 0,
		Sources:      []SourceContext{},
	}
	confidenceEmpty := agent.calculateConfidence(contextEmpty)
	if confidenceEmpty != 0 {
		t.Errorf("Confidence for empty sources should be 0, got %f", confidenceEmpty)
	}
}

func TestRAGAgent_BuildMetadata(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.7,
		},
		SystemPrompt: "Test",
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Test", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	context := &QueryContext{
		TotalResults: 10,
		Sources: []SourceContext{
			{Score: 0.9},
			{Score: 0.7},
			{Score: 0.5},
		},
	}

	metadata := agent.buildMetadata(context)

	if metadata.TotalSources != 10 {
		t.Errorf("Expected TotalSources 10, got %d", metadata.TotalSources)
	}
	if metadata.SourcesUsed != 3 {
		t.Errorf("Expected SourcesUsed 3, got %d", metadata.SourcesUsed)
	}
	if metadata.Model != "gpt-4o" {
		t.Errorf("Expected Model 'gpt-4o', got '%s'", metadata.Model)
	}
	if metadata.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", metadata.Temperature)
	}
	if metadata.MinRelevance != 0.5 {
		t.Errorf("Expected MinRelevance 0.5, got %f", metadata.MinRelevance)
	}
	if metadata.MaxRelevance != 0.9 {
		t.Errorf("Expected MaxRelevance 0.9, got %f", metadata.MaxRelevance)
	}

	expectedAvg := (0.9 + 0.7 + 0.5) / 3.0
	if !floatEquals(metadata.AvgRelevance, expectedAvg, 0.0001) {
		t.Errorf("Expected AvgRelevance %f, got %f", expectedAvg, metadata.AvgRelevance)
	}
}

func TestRAGAgent_CustomPromptTemplate(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt:       "You are a helpful assistant.",
		UserPromptTemplate: "Query: {{query}}\nSources: {{sources}}\nTotal: {{source_count}}",
		Response: CustomAgentResponseConfig{
			MaxContextChunks:  5,
			MinRelevanceScore: 0.0,
		},
	}
	config.applyDefaults()

	mockClient := NewMockLLMClient("Custom response", nil)
	agent, err := NewRAGAgent(config, mockClient)
	if err != nil {
		t.Fatalf("NewRAGAgent() failed: %v", err)
	}

	input := &RAGInput{
		Query: "Test query",
		Results: []*vectordb.QueryResult{
			{
				Document: vectordb.Document{ID: "doc1", Content: "Test content"},
				Score:    0.9,
			},
		},
	}

	result, err := agent.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	output, ok := result.(*RAGOutput)
	if !ok {
		t.Fatalf("Expected *RAGOutput, got %T", result)
	}

	if output.Answer != "Custom response" {
		t.Errorf("Expected answer 'Custom response', got '%s'", output.Answer)
	}
}

// floatEquals checks if two floats are equal within a tolerance
func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
