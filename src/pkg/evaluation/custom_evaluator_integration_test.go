package evaluation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCustomEvaluatorIntegration(t *testing.T) {
	// Create a temp directory for test evaluators
	tmpDir := t.TempDir()
	testEvalDir := filepath.Join(tmpDir, "evaluators")
	if err := os.MkdirAll(testEvalDir, 0755); err != nil {
		t.Fatalf("Failed to create evaluator directory: %v", err)
	}

	// Create a simple regex evaluator
	regexEval := `---
name: test_regex
description: Test regex evaluator
version: 1.0.0
prompt: "Not used"
scoring:
  type: regex
  pattern: '\d+'
  threshold: 0.8
  weight: 1.0
tags:
  - test
author: test
required: false
`
	evalPath := filepath.Join(testEvalDir, "test_regex.yaml")
	if err := os.WriteFile(evalPath, []byte(regexEval), 0644); err != nil {
		t.Fatalf("Failed to write evaluator: %v", err)
	}

	// Create dataset with custom evaluator
	dataset := &Dataset{
		Name:             "test-dataset",
		Version:          "1.0.0",
		CustomEvaluators: []string{"test_regex"},
		TestCases: []TestCase{
			{
				ID:             "test-001",
				Query:          "Test query",
				ExpectedAnswer: "Test answer",
			},
		},
	}

	// Create mock LLM client
	mockClient := &MockLLMClient{response: "0.85"}

	// Create provider
	provider := NewLocalProvider(mockClient)

	// Test without custom evaluators first
	ctx := context.Background()
	result1, err := EvaluateTestCaseWithDataset(ctx, &dataset.TestCases[0], "Answer 123", []string{}, provider, nil)
	if err != nil {
		t.Fatalf("Evaluation without custom evaluators failed: %v", err)
	}
	if len(result1.CustomScores) != 0 {
		t.Error("Expected no custom scores when dataset is nil")
	}

	// Now test by manually loading and executing the custom evaluator
	// (bypassing GetDefaultCustomEvaluatorDir since we can't override it)
	loader := NewCustomEvaluatorLoader(testEvalDir)
	evalDef, err := loader.Load("test_regex")
	if err != nil {
		t.Fatalf("Failed to load test evaluator: %v", err)
	}

	customEval := NewCustomEvaluator(evalDef, mockClient)
	score, err := customEval.Evaluate(ctx, &dataset.TestCases[0], "Answer with 123", []string{})
	if err != nil {
		t.Fatalf("Custom evaluator execution failed: %v", err)
	}

	// Regex evaluator should find the number "123"
	if score < 0.8 {
		t.Errorf("Expected score >= 0.8, got %.2f", score)
	}

	t.Logf("✓ Custom evaluator integration test passed")
	t.Logf("  Custom score (test_regex): %.2f", score)
}
