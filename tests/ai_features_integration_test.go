// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// TestSchemaAndChunkingSuggestions tests end-to-end AI features
func TestSchemaAndChunkingSuggestions(t *testing.T) {
	// Skip if no OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	// Create LLM client
	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	t.Run("SchemaSuggestion_SmallSample", func(t *testing.T) {
		agent := agents.NewSchemaAgent(llmClient)

		// Use docs directory for testing
		input := &agents.SchemaAnalysisInput{
			SampleFiles:    []string{"../docs/USER_GUIDE.md", "../README.md"},
			CollectionName: "TestDocs",
			VDBType:        "weaviate",
			MaxSamples:     2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := agent.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Schema suggestion failed: %v", err)
		}

		output, ok := result.(*agents.SchemaAnalysisOutput)
		if !ok {
			t.Fatal("Invalid output type from schema agent")
		}

		// Verify output
		if output.Schema.CollectionName != "TestDocs" {
			t.Errorf("Expected collection name 'TestDocs', got '%s'", output.Schema.CollectionName)
		}

		if output.Confidence < 0 || output.Confidence > 1 {
			t.Errorf("Confidence should be between 0 and 1, got %f", output.Confidence)
		}

		if len(output.Schema.Fields) == 0 {
			t.Error("Schema should have at least one field")
		}

		t.Logf("✅ Schema suggested with %d fields, confidence: %.1f%%",
			len(output.Schema.Fields), output.Confidence*100)
	})

	t.Run("ChunkingSuggestion_SmallSample", func(t *testing.T) {
		agent := agents.NewChunkingAgent(llmClient)

		input := &agents.ChunkingAnalysisInput{
			SampleFiles:    []string{"../docs/USER_GUIDE.md", "../README.md"},
			CollectionName: "TestDocs",
			VDBType:        "weaviate",
			MaxSamples:     2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := agent.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Chunking suggestion failed: %v", err)
		}

		output, ok := result.(*agents.ChunkingAnalysisOutput)
		if !ok {
			t.Fatal("Invalid output type from chunking agent")
		}

		// Verify output
		rec := output.Recommendation
		if rec.RecommendedSize <= 0 {
			t.Error("Recommended chunk size should be positive")
		}

		if rec.MinSize >= rec.MaxSize {
			t.Errorf("MinSize (%d) should be less than MaxSize (%d)", rec.MinSize, rec.MaxSize)
		}

		if rec.OverlapSize < 0 || rec.OverlapSize > rec.RecommendedSize {
			t.Errorf("Invalid overlap size: %d (recommended: %d)", rec.OverlapSize, rec.RecommendedSize)
		}

		if rec.DocumentType == "" {
			t.Error("Document type should not be empty")
		}

		if output.Confidence < 0 || output.Confidence > 1 {
			t.Errorf("Confidence should be between 0 and 1, got %f", output.Confidence)
		}

		t.Logf("✅ Chunking: %d chars (range: %d-%d), overlap: %d, type: %s, confidence: %.1f%%",
			rec.RecommendedSize, rec.MinSize, rec.MaxSize, rec.OverlapSize,
			rec.DocumentType, output.Confidence*100)
	})

	t.Run("AgentConfigLoading", func(t *testing.T) {
		// Test that agent config loads without error
		config, err := agents.LoadAgentConfig()
		if err != nil {
			t.Fatalf("Failed to load agent config: %v", err)
		}

		// Verify defaults
		if config.LLM.DefaultModel != "gpt-4o" {
			t.Errorf("Expected default model 'gpt-4o', got '%s'", config.LLM.DefaultModel)
		}

		if config.SchemaAgent.MaxSamples != 50 {
			t.Errorf("Expected max samples 50, got %d", config.SchemaAgent.MaxSamples)
		}

		if config.ChunkingAgent.DefaultChunkSize != 1000 {
			t.Errorf("Expected default chunk size 1000, got %d", config.ChunkingAgent.DefaultChunkSize)
		}

		t.Log("✅ Agent config loaded successfully with defaults")
	})
}

// TestSchemaAndChunkingWithRequirements tests custom requirements
func TestSchemaAndChunkingWithRequirements(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	t.Run("SchemaWithRequirements", func(t *testing.T) {
		agent := agents.NewSchemaAgent(llmClient)

		input := &agents.SchemaAnalysisInput{
			SampleFiles:    []string{"../docs/USER_GUIDE.md"},
			CollectionName: "TechDocs",
			VDBType:        "weaviate",
			Requirements:   "Technical documentation with code examples",
			MaxSamples:     1,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := agent.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Schema suggestion with requirements failed: %v", err)
		}

		output := result.(*agents.SchemaAnalysisOutput)
		if output.Reasoning == "" {
			t.Error("Reasoning should be provided")
		}

		t.Logf("✅ Schema with requirements: %s", output.Reasoning[:100]+"...")
	})

	t.Run("ChunkingWithRequirements", func(t *testing.T) {
		agent := agents.NewChunkingAgent(llmClient)

		input := &agents.ChunkingAnalysisInput{
			SampleFiles:    []string{"../docs/USER_GUIDE.md"},
			CollectionName: "TechDocs",
			VDBType:        "weaviate",
			Requirements:   "Code-heavy technical documentation",
			MaxSamples:     1,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := agent.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Chunking suggestion with requirements failed: %v", err)
		}

		output := result.(*agents.ChunkingAnalysisOutput)
		if output.Recommendation.Reasoning == "" {
			t.Error("Reasoning should be provided")
		}

		// Technical docs typically have smaller chunks
		if output.Recommendation.RecommendedSize > 1500 {
			t.Logf("Note: Technical docs usually recommend smaller chunks, got %d",
				output.Recommendation.RecommendedSize)
		}

		t.Logf("✅ Chunking with requirements: %s", output.Recommendation.Reasoning[:100]+"...")
	})
}

// TestAgentErrorHandling tests error scenarios
func TestAgentErrorHandling(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	t.Run("EmptyFileList", func(t *testing.T) {
		agent := agents.NewSchemaAgent(llmClient)

		input := &agents.SchemaAnalysisInput{
			SampleFiles:    []string{},
			CollectionName: "TestDocs",
			VDBType:        "weaviate",
			MaxSamples:     10,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := agent.Execute(ctx, input)
		if err == nil {
			t.Error("Should fail with empty file list")
		}

		t.Logf("✅ Correctly failed with empty files: %v", err)
	})

	t.Run("InvalidFilePath", func(t *testing.T) {
		agent := agents.NewChunkingAgent(llmClient)

		input := &agents.ChunkingAnalysisInput{
			SampleFiles:    []string{"/nonexistent/path/file.txt"},
			CollectionName: "TestDocs",
			VDBType:        "weaviate",
			MaxSamples:     1,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := agent.Execute(ctx, input)
		if err == nil {
			t.Error("Should fail with invalid file path")
		}

		t.Logf("✅ Correctly failed with invalid path: %v", err)
	})
}
