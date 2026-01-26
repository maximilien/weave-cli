// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestCustomEvaluatorDefValidate tests evaluator definition validation
func TestCustomEvaluatorDefValidate(t *testing.T) {
	t.Run("ValidDefinition", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:        "test_eval",
			Description: "Test evaluator",
			Prompt:      "Evaluate: {{.Answer}}",
			Scoring: ScoringConfig{
				Type:      "llm_judge",
				Threshold: 0.7,
			},
		}

		err := def.Validate()
		if err != nil {
			t.Errorf("Valid definition should not error: %v", err)
		}

		// Check default weight is set
		if def.Scoring.Weight != 1.0 {
			t.Errorf("Expected default weight 1.0, got: %.2f", def.Scoring.Weight)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Prompt: "Test",
			Scoring: ScoringConfig{
				Type:      "llm_judge",
				Threshold: 0.7,
			},
		}

		err := def.Validate()
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})

	t.Run("MissingPrompt", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name: "test",
			Scoring: ScoringConfig{
				Type:      "llm_judge",
				Threshold: 0.7,
			},
		}

		err := def.Validate()
		if err == nil {
			t.Error("Expected error for missing prompt")
		}
	})

	t.Run("Invalid ScoringType", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "test",
			Prompt: "Test",
			Scoring: ScoringConfig{
				Type:      "invalid_type",
				Threshold: 0.7,
			},
		}

		err := def.Validate()
		if err == nil {
			t.Error("Expected error for invalid scoring type")
		}
	})

	t.Run("RegexWithoutPattern", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "test",
			Prompt: "Test",
			Scoring: ScoringConfig{
				Type:      "regex",
				Threshold: 0.7,
			},
		}

		err := def.Validate()
		if err == nil {
			t.Error("Expected error for regex without pattern")
		}
	})

	t.Run("InvalidThreshold", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "test",
			Prompt: "Test",
			Scoring: ScoringConfig{
				Type:      "llm_judge",
				Threshold: 1.5, // Invalid - > 1.0
			},
		}

		err := def.Validate()
		if err == nil {
			t.Error("Expected error for invalid threshold")
		}
	})
}

// TestCustomEvaluatorLoader tests loading evaluators from YAML files
func TestCustomEvaluatorLoader(t *testing.T) {
	// Create temporary directory with test evaluators
	tmpDir := t.TempDir()
	evalDir := filepath.Join(tmpDir, "evaluators")
	os.MkdirAll(evalDir, 0755)

	// Create test evaluator YAML
	yaml1 := `name: test_evaluator
description: Test evaluator for unit tests
version: 1.0.0
prompt: |
  Evaluate the answer: {{.Answer}}
  Expected: {{.ExpectedAnswer}}
scoring:
  type: llm_judge
  threshold: 0.7
  weight: 1.0
tags:
  - test
  - example
author: test-team
required: false
`

	yaml2 := `name: regex_evaluator
description: Regex-based evaluator
prompt: "Not used for regex"
scoring:
  type: regex
  pattern: '\b(correct|accurate)\b'
  threshold: 0.5
`

	os.WriteFile(filepath.Join(evalDir, "test_evaluator.yaml"), []byte(yaml1), 0644)
	os.WriteFile(filepath.Join(evalDir, "regex_evaluator.yaml"), []byte(yaml2), 0644)

	loader := NewCustomEvaluatorLoader(evalDir)

	t.Run("LoadSingleEvaluator", func(t *testing.T) {
		def, err := loader.Load("test_evaluator")
		if err != nil {
			t.Fatalf("Failed to load evaluator: %v", err)
		}

		if def.Name != "test_evaluator" {
			t.Errorf("Expected name 'test_evaluator', got: %s", def.Name)
		}

		if def.Scoring.Type != "llm_judge" {
			t.Errorf("Expected type 'llm_judge', got: %s", def.Scoring.Type)
		}

		if def.Scoring.Threshold != 0.7 {
			t.Errorf("Expected threshold 0.7, got: %.2f", def.Scoring.Threshold)
		}

		if len(def.Tags) != 2 {
			t.Errorf("Expected 2 tags, got: %d", len(def.Tags))
		}
	})

	t.Run("LoadAllEvaluators", func(t *testing.T) {
		defs, err := loader.LoadAll()
		if err != nil {
			t.Fatalf("Failed to load all evaluators: %v", err)
		}

		if len(defs) != 2 {
			t.Errorf("Expected 2 evaluators, got: %d", len(defs))
		}
	})

	t.Run("LoadNonExistentEvaluator", func(t *testing.T) {
		_, err := loader.Load("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent evaluator")
		}
	})

	t.Run("LoadFromEmptyDirectory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.MkdirAll(emptyDir, 0755)

		loader := NewCustomEvaluatorLoader(emptyDir)
		defs, err := loader.LoadAll()
		if err != nil {
			t.Errorf("LoadAll should not error on empty directory: %v", err)
		}

		if len(defs) != 0 {
			t.Errorf("Expected 0 evaluators from empty directory, got: %d", len(defs))
		}
	})
}

// TestCustomEvaluatorExecution tests custom evaluator execution
func TestCustomEvaluatorExecution(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:             "test-001",
		Query:          "What is AI?",
		ExpectedAnswer: "Artificial Intelligence",
	}

	t.Run("LLMJudgeEvaluator", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "llm_test",
			Prompt: "Rate this: {{.Answer}}",
			Scoring: ScoringConfig{
				Type:      "llm_judge",
				Threshold: 0.7,
			},
		}

		eval := NewCustomEvaluator(def, mockClient)

		score, err := eval.Evaluate(context.Background(), testCase, "AI stands for Artificial Intelligence", []string{})
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if score == 0.0 {
			t.Error("Expected non-zero score from LLM judge")
		}

		if score < 0.0 || score > 1.0 {
			t.Errorf("Score out of range: %.2f", score)
		}
	})

	t.Run("RegexEvaluator", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "regex_test",
			Prompt: "Not used",
			Scoring: ScoringConfig{
				Type:      "regex",
				Pattern:   `\b(AI|Artificial Intelligence)\b`,
				Threshold: 0.5,
			},
		}

		eval := NewCustomEvaluator(def, mockClient)

		// Should match
		score, err := eval.Evaluate(context.Background(), testCase, "AI is amazing", []string{})
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if score != 1.0 {
			t.Errorf("Expected score 1.0 for matching regex, got: %.2f", score)
		}

		// Should not match
		score, err = eval.Evaluate(context.Background(), testCase, "Machine learning is cool", []string{})
		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if score != 0.0 {
			t.Errorf("Expected score 0.0 for non-matching regex, got: %.2f", score)
		}
	})

	t.Run("ExactMatchEvaluator", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "exact_test",
			Prompt: "Not used",
			Scoring: ScoringConfig{
				Type:      "exact_match",
				Threshold: 0.9,
			},
		}

		eval := NewCustomEvaluator(def, mockClient)

		// Exact match
		score, _ := eval.Evaluate(context.Background(), testCase, "Artificial Intelligence", []string{})
		if score != 1.0 {
			t.Errorf("Expected 1.0 for exact match, got: %.2f", score)
		}

		// Whitespace normalization
		score, _ = eval.Evaluate(context.Background(), testCase, "  Artificial Intelligence  ", []string{})
		if score != 1.0 {
			t.Errorf("Expected 1.0 with whitespace normalization, got: %.2f", score)
		}

		// No match
		score, _ = eval.Evaluate(context.Background(), testCase, "Machine Learning", []string{})
		if score != 0.0 {
			t.Errorf("Expected 0.0 for no match, got: %.2f", score)
		}
	})

	t.Run("ContainsEvaluator", func(t *testing.T) {
		def := &CustomEvaluatorDef{
			Name:   "contains_test",
			Prompt: "Not used",
			Scoring: ScoringConfig{
				Type:      "contains",
				Threshold: 0.5,
			},
		}

		eval := NewCustomEvaluator(def, mockClient)

		// Contains
		score, _ := eval.Evaluate(context.Background(), testCase, "The term Artificial Intelligence was coined...", []string{})
		if score != 1.0 {
			t.Errorf("Expected 1.0 for contains, got: %.2f", score)
		}

		// Case insensitive
		score, _ = eval.Evaluate(context.Background(), testCase, "artificial intelligence is cool", []string{})
		if score != 1.0 {
			t.Errorf("Expected 1.0 for case-insensitive contains, got: %.2f", score)
		}

		// Not contains
		score, _ = eval.Evaluate(context.Background(), testCase, "Machine learning", []string{})
		if score != 0.0 {
			t.Errorf("Expected 0.0 for not contains, got: %.2f", score)
		}
	})
}

// TestPromptTemplateRendering tests template rendering with test data
func TestPromptTemplateRendering(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.90"}

	testCase := &TestCase{
		ID:               "test-001",
		Query:            "What is RAG?",
		ExpectedAnswer:   "Retrieval Augmented Generation",
		RetrievedContext: []string{"RAG combines retrieval", "Generation uses LLMs"},
	}

	def := &CustomEvaluatorDef{
		Name:        "template_test",
		Description: "Tests template rendering",
		Prompt: `Query: {{.Query}}
Expected: {{.ExpectedAnswer}}
Actual: {{.Answer}}
Context: {{.Context}}
Description: {{.Description}}`,
		Scoring: ScoringConfig{
			Type:      "llm_judge",
			Threshold: 0.7,
		},
	}

	eval := NewCustomEvaluator(def, mockClient)

	// Render the prompt
	prompt, err := eval.renderPrompt(testCase, "RAG is Retrieval Augmented Generation")
	if err != nil {
		t.Fatalf("Failed to render prompt: %v", err)
	}

	// Check that all template variables were substituted
	if !contains(prompt, "What is RAG?") {
		t.Error("Prompt should contain query")
	}

	if !contains(prompt, "Retrieval Augmented Generation") {
		t.Error("Prompt should contain expected answer")
	}

	if !contains(prompt, "RAG combines retrieval") {
		t.Error("Prompt should contain context")
	}

	if !contains(prompt, "Tests template rendering") {
		t.Error("Prompt should contain description")
	}
}

// TestGetDefaultCustomEvaluatorDir tests default directory detection
func TestGetDefaultCustomEvaluatorDir(t *testing.T) {
	dir := GetDefaultCustomEvaluatorDir()

	if dir == "" {
		t.Error("Expected non-empty default directory")
	}

	// In dev mode (with .git), should return relative path
	if _, err := os.Stat(".git"); err == nil {
		if dir != "evals/evaluators" {
			t.Logf("Dev mode directory: %s", dir)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
