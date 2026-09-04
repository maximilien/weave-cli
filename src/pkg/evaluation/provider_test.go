// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"testing"
)

// TestLocalProvider tests the LocalProvider implementation
func TestLocalProvider(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}
	provider := NewLocalProvider(mockClient)

	t.Run("Name", func(t *testing.T) {
		if provider.Name() != "local" {
			t.Errorf("Expected name 'local', got: %s", provider.Name())
		}
	})

	t.Run("IsAvailable", func(t *testing.T) {
		ctx := context.Background()
		if !provider.IsAvailable(ctx) {
			t.Error("LocalProvider should be available with valid LLM client")
		}
	})

	t.Run("IsAvailableWithNilClient", func(t *testing.T) {
		ctx := context.Background()
		nilProvider := NewLocalProvider(nil)
		if nilProvider.IsAvailable(ctx) {
			t.Error("LocalProvider should not be available with nil client")
		}
	})

	t.Run("GetAccuracyEvaluator", func(t *testing.T) {
		eval := provider.GetAccuracyEvaluator()
		if eval == nil {
			t.Error("Expected non-nil accuracy evaluator")
		}
		if eval.Name() != "accuracy" {
			t.Errorf("Expected evaluator name 'accuracy', got: %s", eval.Name())
		}
	})

	t.Run("GetFaithfulnessEvaluator", func(t *testing.T) {
		eval := provider.GetFaithfulnessEvaluator()
		if eval == nil {
			t.Error("Expected non-nil faithfulness evaluator")
		}
		if eval.Name() != "faithfulness" {
			t.Errorf("Expected evaluator name 'faithfulness', got: %s", eval.Name())
		}
	})

	t.Run("GetHallucinationEvaluator", func(t *testing.T) {
		eval := provider.GetHallucinationEvaluator()
		if eval == nil {
			t.Error("Expected non-nil hallucination evaluator")
		}
		if eval.Name() != "hallucination" {
			t.Errorf("Expected evaluator name 'hallucination', got: %s", eval.Name())
		}
	})

	t.Run("GetContextRelevanceEvaluator", func(t *testing.T) {
		eval := provider.GetContextRelevanceEvaluator()
		if eval == nil {
			t.Error("Expected non-nil context relevance evaluator")
		}
		if eval.Name() != "context_relevance" {
			t.Errorf("Expected evaluator name 'context_relevance', got: %s", eval.Name())
		}
	})
}

// TestProviderFactory tests the provider factory
func TestProviderFactory(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	t.Run("CreateLocalProvider", func(t *testing.T) {
		ctx := context.Background()
		provider, err := CreateProvider(ctx, ProviderTypeLocal, mockClient)

		if err != nil {
			t.Fatalf("Failed to create local provider: %v", err)
		}

		if provider.Name() != "local" {
			t.Errorf("Expected provider name 'local', got: %s", provider.Name())
		}

		if !provider.IsAvailable(ctx) {
			t.Error("Local provider should be available")
		}
	})

	t.Run("CreateLocalProviderWithNilClient", func(t *testing.T) {
		ctx := context.Background()
		_, err := CreateProvider(ctx, ProviderTypeLocal, nil)

		if err == nil {
			t.Error("Expected error when creating local provider with nil client")
		}
	})

	t.Run("CreateOpikProviderWithoutAPIKey", func(t *testing.T) {
		t.Setenv("OPIK_API_KEY", "")
		ctx := context.Background()
		_, err := CreateProvider(ctx, ProviderTypeOpik, mockClient)

		if err == nil {
			t.Error("Expected error when creating Opik provider without API key")
		}
	})

	t.Run("UnknownProviderType", func(t *testing.T) {
		ctx := context.Background()
		_, err := CreateProvider(ctx, ProviderType("invalid"), mockClient)

		if err == nil {
			t.Error("Expected error for unknown provider type")
		}
	})

	t.Run("GetDefaultProvider", func(t *testing.T) {
		defaultType := GetDefaultProvider()
		if defaultType != ProviderTypeLocal {
			t.Errorf("Expected default provider 'local', got: %s", defaultType)
		}
	})
}

// TestGetAvailableProviders tests available provider detection
func TestGetAvailableProviders(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	t.Run("WithLLMClient", func(t *testing.T) {
		ctx := context.Background()
		providers := GetAvailableProviders(ctx, mockClient)

		if len(providers) == 0 {
			t.Error("Expected at least one available provider")
		}

		// Local should always be available with valid client
		foundLocal := false
		for _, p := range providers {
			if p == ProviderTypeLocal {
				foundLocal = true
				break
			}
		}

		if !foundLocal {
			t.Error("Expected local provider to be available")
		}
	})

	t.Run("WithoutLLMClient", func(t *testing.T) {
		ctx := context.Background()
		providers := GetAvailableProviders(ctx, nil)

		// Should not include local without client
		for _, p := range providers {
			if p == ProviderTypeLocal {
				t.Error("Local provider should not be available without LLM client")
			}
		}
	})
}

// TestEvaluateTestCaseWithProvider tests evaluation with different providers
func TestEvaluateTestCaseWithProvider(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:                "test-provider-001",
		Query:             "What is a vector database?",
		ExpectedAnswer:    "A database for storing vectors",
		RetrievedContext:  []string{"Vector databases store embeddings"},
		RequiredConcepts:  []string{"vector", "database"},
		MustCite:          false,
		MinRelevanceScore: 0.7,
	}

	t.Run("EvaluateWithLocalProvider", func(t *testing.T) {
		ctx := context.Background()
		provider := NewLocalProvider(mockClient)

		result, err := EvaluateTestCase(ctx, testCase, "A vector database stores embeddings", nil, provider)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify provider info is stored
		if result.Details["evaluator_provider"] != "local" {
			t.Errorf("Expected provider 'local', got: %v", result.Details["evaluator_provider"])
		}

		// Verify all scores are populated
		if result.AccuracyScore == 0 {
			t.Error("AccuracyScore should be populated")
		}
		if result.CitationScore == 0 {
			t.Error("CitationScore should be populated (rule-based)")
		}
		if result.HallucinationScore == 0 {
			t.Error("HallucinationScore should be populated")
		}
		if result.ContextRelevanceScore == 0 {
			t.Error("ContextRelevanceScore should be populated")
		}
		if result.FaithfulnessScore == 0 {
			t.Error("FaithfulnessScore should be populated")
		}
	})

	t.Run("EvaluateWithDifferentProviderReturnsDifferentProviderName", func(t *testing.T) {
		ctx := context.Background()

		// Create a mock provider that returns different name
		mockProvider := &MockProvider{name: "mock-provider"}

		result, err := EvaluateTestCase(ctx, testCase, "A vector database stores embeddings", nil, mockProvider)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify provider info matches mock
		if result.Details["evaluator_provider"] != "mock-provider" {
			t.Errorf("Expected provider 'mock-provider', got: %v", result.Details["evaluator_provider"])
		}
	})
}

// MockProvider is a test provider for testing provider selection
type MockProvider struct {
	name string
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) GetAccuracyEvaluator() Evaluator {
	return &MockEvaluator{name: "accuracy", score: 0.85}
}

func (p *MockProvider) GetFaithfulnessEvaluator() Evaluator {
	return &MockEvaluator{name: "faithfulness", score: 0.90}
}

func (p *MockProvider) GetHallucinationEvaluator() Evaluator {
	return &MockEvaluator{name: "hallucination", score: 0.95}
}

func (p *MockProvider) GetContextRelevanceEvaluator() Evaluator {
	return &MockEvaluator{name: "context_relevance", score: 0.88}
}

func (p *MockProvider) IsAvailable(ctx context.Context) bool {
	return true
}

// MockEvaluator is a test evaluator that returns fixed scores
type MockEvaluator struct {
	name  string
	score float64
}

func (e *MockEvaluator) Name() string {
	return e.name
}

func (e *MockEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	return e.score, nil
}

// TestProviderSwitching tests switching between providers
func TestProviderSwitching(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:             "test-switch-001",
		Query:          "Test query",
		ExpectedAnswer: "Test answer",
	}

	ctx := context.Background()

	// Test with local provider
	localProvider := NewLocalProvider(mockClient)
	localResult, err := EvaluateTestCase(ctx, testCase, "Test response", nil, localProvider)
	if err != nil {
		t.Fatalf("Local evaluation failed: %v", err)
	}

	// Test with mock provider
	mockProvider := &MockProvider{name: "test-provider"}
	mockResult, err := EvaluateTestCase(ctx, testCase, "Test response", nil, mockProvider)
	if err != nil {
		t.Fatalf("Mock evaluation failed: %v", err)
	}

	// Verify different providers are used
	if localResult.Details["evaluator_provider"] == mockResult.Details["evaluator_provider"] {
		t.Error("Expected different provider names for different providers")
	}

	if localResult.Details["evaluator_provider"] != "local" {
		t.Errorf("Expected local provider, got: %v", localResult.Details["evaluator_provider"])
	}

	if mockResult.Details["evaluator_provider"] != "test-provider" {
		t.Errorf("Expected test-provider, got: %v", mockResult.Details["evaluator_provider"])
	}
}

// TestBackwardCompatibility tests that old code still works
func TestBackwardCompatibility(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:                "test-compat-001",
		Query:             "What is RAG?",
		ExpectedAnswer:    "Retrieval Augmented Generation",
		RequiredConcepts:  []string{"retrieval"},
		MinRelevanceScore: 0.7,
	}

	ctx := context.Background()

	// Test using old wrapper function
	result, err := EvaluateTestCaseWithLLMClient(ctx, testCase, "RAG is Retrieval Augmented Generation", nil, mockClient)

	if err != nil {
		t.Fatalf("Backward compatibility test failed: %v", err)
	}

	// Should use local provider by default
	if result.Details["evaluator_provider"] != "local" {
		t.Errorf("Expected default provider 'local', got: %v", result.Details["evaluator_provider"])
	}

	// Verify all scores work
	if result.AccuracyScore == 0 {
		t.Error("AccuracyScore should be populated")
	}
}
