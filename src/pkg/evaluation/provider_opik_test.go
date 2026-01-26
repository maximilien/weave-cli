// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// TestNewOpikProvider tests Opik provider creation
func TestNewOpikProvider(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	t.Run("ValidConfig", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: true,
			APIKey:  "test-key",
		}

		provider, err := NewOpikProvider(config, mockClient)
		if err != nil {
			t.Fatalf("Failed to create Opik provider: %v", err)
		}

		if provider.Name() != "opik" {
			t.Errorf("Expected provider name 'opik', got: %s", provider.Name())
		}

		// Verify provider has LLM client
		if provider.llmClient == nil {
			t.Error("Expected provider to have LLM client")
		}
	})

	t.Run("MissingAPIKey", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: true,
			APIKey:  "",
		}

		_, err := NewOpikProvider(config, mockClient)
		if err == nil {
			t.Error("Expected error for missing API key")
		}
	})

	t.Run("MissingLLMClient", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: true,
			APIKey:  "test-key",
		}

		_, err := NewOpikProvider(config, nil)
		if err == nil {
			t.Error("Expected error for missing LLM client")
		}
	})

	t.Run("DisabledConfig", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: false,
			APIKey:  "test-key",
		}

		_, err := NewOpikProvider(config, mockClient)
		if err == nil {
			t.Error("Expected error for disabled config")
		}
	})
}

// TestOpikProviderEvaluators tests that Opik provider returns correct evaluators
func TestOpikProviderEvaluators(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}
	config := &llm.OpikConfig{
		Enabled: true,
		APIKey:  "test-key",
	}

	provider, err := NewOpikProvider(config, mockClient)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	t.Run("GetAccuracyEvaluator", func(t *testing.T) {
		eval := provider.GetAccuracyEvaluator()
		if eval == nil {
			t.Fatal("Expected non-nil evaluator")
		}
		if eval.Name() != "accuracy" {
			t.Errorf("Expected name 'accuracy', got: %s", eval.Name())
		}
	})

	t.Run("GetFaithfulnessEvaluator", func(t *testing.T) {
		eval := provider.GetFaithfulnessEvaluator()
		if eval == nil {
			t.Fatal("Expected non-nil evaluator")
		}
		if eval.Name() != "faithfulness" {
			t.Errorf("Expected name 'faithfulness', got: %s", eval.Name())
		}
	})

	t.Run("GetHallucinationEvaluator", func(t *testing.T) {
		eval := provider.GetHallucinationEvaluator()
		if eval == nil {
			t.Fatal("Expected non-nil evaluator")
		}
		if eval.Name() != "hallucination" {
			t.Errorf("Expected name 'hallucination', got: %s", eval.Name())
		}
	})

	t.Run("GetContextRelevanceEvaluator", func(t *testing.T) {
		eval := provider.GetContextRelevanceEvaluator()
		if eval == nil {
			t.Fatal("Expected non-nil evaluator")
		}
		if eval.Name() != "context_relevance" {
			t.Errorf("Expected name 'context_relevance', got: %s", eval.Name())
		}
	})
}

// TestOpikAccuracyEvaluator tests Opik accuracy evaluator returns real scores
func TestOpikAccuracyEvaluator(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}
	config := &llm.OpikConfig{
		Enabled: true,
		APIKey:  "test-key",
	}

	provider, err := NewOpikProvider(config, mockClient)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	testCase := &TestCase{
		ID:             "test-001",
		Query:          "What is a vector database?",
		ExpectedAnswer: "A database for storing vectors",
	}

	eval := provider.GetAccuracyEvaluator()
	score, err := eval.Evaluate(ctx, testCase, "A database that stores vector embeddings", []string{})

	// Should return real score, not 0.0
	if score == 0.0 {
		t.Error("Expected non-zero score from Opik accuracy evaluator")
	}

	// Should not return error (unlike old implementation)
	if err != nil {
		t.Logf("Evaluation returned error: %v", err)
	}

	// Score should be in valid range
	if score < 0.0 || score > 1.0 {
		t.Errorf("Score out of range [0,1]: %.2f", score)
	}
}

// TestOpikEvaluatorScores tests all Opik evaluators return real scores
func TestOpikEvaluatorScores(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.75"}
	config := &llm.OpikConfig{
		Enabled: true,
		APIKey:  "test-key",
	}

	provider, err := NewOpikProvider(config, mockClient)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	testCase := &TestCase{
		ID:               "test-002",
		Query:            "Explain RAG",
		ExpectedAnswer:   "Retrieval Augmented Generation",
		RetrievedContext: []string{"RAG combines retrieval with generation"},
	}

	t.Run("AccuracyReturnsRealScore", func(t *testing.T) {
		eval := provider.GetAccuracyEvaluator()
		score, _ := eval.Evaluate(ctx, testCase, "RAG retrieves and generates", []string{})

		if score == 0.0 {
			t.Error("Expected non-zero score")
		}
	})

	t.Run("FaithfulnessReturnsRealScore", func(t *testing.T) {
		eval := provider.GetFaithfulnessEvaluator()
		score, _ := eval.Evaluate(ctx, testCase, "RAG retrieves and generates", []string{})

		if score == 0.0 {
			t.Error("Expected non-zero score")
		}
	})

	t.Run("HallucinationReturnsRealScore", func(t *testing.T) {
		eval := provider.GetHallucinationEvaluator()
		score, _ := eval.Evaluate(ctx, testCase, "RAG retrieves and generates", []string{})

		if score == 0.0 {
			t.Error("Expected non-zero score")
		}
	})

	t.Run("ContextRelevanceReturnsRealScore", func(t *testing.T) {
		eval := provider.GetContextRelevanceEvaluator()
		score, _ := eval.Evaluate(ctx, testCase, "RAG retrieves and generates", []string{})

		// Context relevance may return 0.0 for invalid context, so check range
		if score < 0.0 || score > 1.0 {
			t.Errorf("Score out of range: %.2f", score)
		}
	})
}

// TestOpikProviderIntegration tests full evaluation with Opik provider
func TestOpikProviderIntegration(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.80"}
	config := &llm.OpikConfig{
		Enabled: true,
		APIKey:  "test-key",
	}

	provider, err := NewOpikProvider(config, mockClient)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	testCase := &TestCase{
		ID:               "integration-001",
		Query:            "What is embeddings?",
		ExpectedAnswer:   "Vector representations of text",
		RequiredConcepts: []string{"vector", "representation"},
		RetrievedContext: []string{"Embeddings are vector representations"},
	}

	// Run full evaluation
	result, err := EvaluateTestCase(ctx, testCase, "Embeddings are vectors representing text", []string{"1"}, provider)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Verify result structure
	if result.TestCaseID != testCase.ID {
		t.Errorf("Expected test case ID %s, got %s", testCase.ID, result.TestCaseID)
	}

	// Verify provider info
	if result.Details["evaluator_provider"] != "opik" {
		t.Errorf("Expected provider 'opik', got %v", result.Details["evaluator_provider"])
	}

	// Verify all scores are present and non-zero
	scores := map[string]float64{
		"accuracy":          result.AccuracyScore,
		"citation":          result.CitationScore,
		"hallucination":     result.HallucinationScore,
		"faithfulness":      result.FaithfulnessScore,
		"context_relevance": result.ContextRelevanceScore,
	}

	for name, score := range scores {
		// Citation is rule-based and may be 0, others should be non-zero
		if name != "citation" && score == 0.0 {
			t.Errorf("Expected non-zero %s score", name)
		}

		if score < 0.0 || score > 1.0 {
			t.Errorf("%s score out of range: %.2f", name, score)
		}
	}
}

// TestOpikProviderIsAvailable tests availability check
func TestOpikProviderIsAvailable(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	t.Run("AvailableWithValidConfig", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: true,
			APIKey:  "test-key",
		}

		provider, err := NewOpikProvider(config, mockClient)
		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}
		defer provider.Shutdown(context.Background())

		if !provider.IsAvailable(context.Background()) {
			t.Error("Expected provider to be available")
		}
	})

	t.Run("UnavailableWithMissingAPIKey", func(t *testing.T) {
		config := &llm.OpikConfig{
			Enabled: true,
			APIKey:  "",
		}

		// Can't create provider without API key
		_, err := NewOpikProvider(config, mockClient)
		if err == nil {
			t.Error("Expected error for missing API key")
		}
	})
}

// TestOpikProviderShutdown tests graceful shutdown
func TestOpikProviderShutdown(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}
	config := &llm.OpikConfig{
		Enabled: true,
		APIKey:  "test-key",
	}

	provider, err := NewOpikProvider(config, mockClient)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Shutdown should not error
	err = provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}
