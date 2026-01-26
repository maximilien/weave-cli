// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"
	"testing"
)

// TestEvaluateTestCaseErrorHandling tests error paths in EvaluateTestCase
func TestEvaluateTestCaseErrorHandling(t *testing.T) {
	ctx := context.Background()

	testCase := &TestCase{
		ID:                "test-error-001",
		Query:             "Test query",
		ExpectedAnswer:    "Expected answer",
		MinRelevanceScore: 0.7,
		MustCite:          true,
	}

	t.Run("EvaluatorErrorsRecorded", func(t *testing.T) {
		// Use mock client that returns error
		mockClient := &MockLLMClient{response: "invalid", err: fmt.Errorf("mock LLM error")}
		provider := NewLocalProvider(mockClient)

		result, err := EvaluateTestCase(ctx, testCase, "Actual answer", []string{"1"}, provider)
		if err != nil {
			t.Fatalf("EvaluateTestCase should not return error: %v", err)
		}

		// Should have recorded errors from evaluators
		if len(result.Errors) == 0 {
			t.Error("Expected errors to be recorded from failed evaluators")
		}
	})

	t.Run("MinRelevanceScoreThreshold", func(t *testing.T) {
		mockClient := &MockLLMClient{response: "0.5"} // Below threshold of 0.7
		provider := NewLocalProvider(mockClient)

		result, err := EvaluateTestCase(ctx, testCase, "Actual answer", []string{"1"}, provider)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected test to fail due to low accuracy score")
		}

		// Check for threshold error message
		foundThresholdError := false
		for _, errMsg := range result.Errors {
			if len(errMsg) > 0 && errMsg[:8] == "Accuracy" {
				foundThresholdError = true
				break
			}
		}
		if !foundThresholdError {
			t.Error("Expected accuracy threshold error message")
		}
	})

	t.Run("MustCiteRequirement", func(t *testing.T) {
		mockClient := &MockLLMClient{response: "0.85"}
		provider := NewLocalProvider(mockClient)

		// No citations provided, but MustCite is true
		result, err := EvaluateTestCase(ctx, testCase, "Answer without citations", []string{}, provider)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected test to fail due to missing required citations")
		}

		// Check for citation error message
		foundCitationError := false
		for _, errMsg := range result.Errors {
			if len(errMsg) > 0 && (errMsg == "Missing required citations" || errMsg[:7] == "Missing") {
				foundCitationError = true
				break
			}
		}
		if !foundCitationError {
			t.Error("Expected missing citation error message")
		}
	})

	t.Run("HallucinationScoreThreshold", func(t *testing.T) {
		mockClient := &MockLLMClient{response: "0.4"} // Below 0.6 threshold
		provider := NewLocalProvider(mockClient)

		result, err := EvaluateTestCase(ctx, testCase, "Actual answer", []string{"1"}, provider)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected test to fail due to high hallucination risk")
		}

		// Check for hallucination error message
		foundHallucinationError := false
		for _, errMsg := range result.Errors {
			if len(errMsg) > 4 && errMsg[:4] == "High" {
				foundHallucinationError = true
				break
			}
		}
		if !foundHallucinationError {
			t.Error("Expected hallucination error message")
		}
	})
}

// TestEvaluateTestCaseWithLLMClient tests backward compatibility wrapper
func TestEvaluateTestCaseWithLLMClient(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:             "test-compat-001",
		Query:          "Test query",
		ExpectedAnswer: "Expected answer",
	}

	result, err := EvaluateTestCaseWithLLMClient(ctx, testCase, "Actual answer", []string{"1"}, mockClient)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.TestCaseID != testCase.ID {
		t.Errorf("Expected test case ID %s, got %s", testCase.ID, result.TestCaseID)
	}

	// Verify it uses LocalProvider
	if result.Details["evaluator_provider"] != "local" {
		t.Errorf("Expected provider 'local', got %v", result.Details["evaluator_provider"])
	}
}

// TestEvaluatorErrorScenarios tests various error scenarios for evaluators
func TestEvaluatorErrorScenarios(t *testing.T) {
	ctx := context.Background()

	testCase := &TestCase{
		ID:             "test-error-002",
		Query:          "Test query",
		ExpectedAnswer: "Expected answer",
	}

	t.Run("AccuracyEvaluatorWithInvalidResponse", func(t *testing.T) {
		mockClient := &MockLLMClient{response: "not-a-number"}
		eval := NewAccuracyEvaluator(mockClient)

		score, err := eval.Evaluate(ctx, testCase, "Actual answer", []string{})
		if err != nil {
			t.Logf("Error returned (expected): %v", err)
		}

		// Even with error, should return a score (default 0.0)
		if score < 0 || score > 1 {
			t.Errorf("Score should be in range [0,1], got: %.2f", score)
		}
	})

	t.Run("ProviderEvaluatorWithError", func(t *testing.T) {
		mockClient := &MockLLMClient{err: fmt.Errorf("simulated error")}
		provider := NewLocalProvider(mockClient)

		// Test that evaluators handle errors by recording them
		result, err := EvaluateTestCase(ctx, testCase, "Actual answer", []string{"1"}, provider)
		if err != nil {
			t.Fatalf("EvaluateTestCase should not return error: %v", err)
		}

		// May have recorded errors from evaluators (implementation dependent)
		t.Logf("Recorded %d errors from evaluators", len(result.Errors))
	})
}

// TestListDatasetsComprehensive tests ListDatasets with various scenarios
func TestListDatasetsComprehensive(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	t.Run("EmptyDirectory", func(t *testing.T) {
		// This tests against actual default directory (may be empty)
		datasets, err := ListDatasets()
		if err != nil {
			t.Fatalf("Failed to list datasets: %v", err)
		}

		// Should return empty slice, not nil
		if datasets == nil {
			t.Error("Expected non-nil slice")
		}
	})

	// Note: Can't easily test ListDatasets with tmpDir since it uses default directory
	// But we can test the SaveDataset and LoadDataset functions it depends on
	t.Run("SaveAndLoadMultipleDatasets", func(t *testing.T) {
		dataset1 := &Dataset{
			Name:        "dataset-1",
			Version:     "1.0.0",
			Description: "First dataset",
			TestCases: []TestCase{
				{ID: "test-001", Query: "Q1", ExpectedAnswer: "A1"},
			},
		}

		dataset2 := &Dataset{
			Name:        "dataset-2",
			Version:     "2.0.0",
			Description: "Second dataset",
			TestCases: []TestCase{
				{ID: "test-002", Query: "Q2", ExpectedAnswer: "A2"},
			},
		}

		// Save datasets
		path1 := fmt.Sprintf("%s/dataset-1.yaml", tmpDir)
		path2 := fmt.Sprintf("%s/dataset-2.yaml", tmpDir)

		if err := SaveDataset(dataset1, path1); err != nil {
			t.Fatalf("Failed to save dataset1: %v", err)
		}

		if err := SaveDataset(dataset2, path2); err != nil {
			t.Fatalf("Failed to save dataset2: %v", err)
		}

		// Load and verify
		loaded1, err := LoadDataset(path1)
		if err != nil {
			t.Fatalf("Failed to load dataset1: %v", err)
		}

		if loaded1.Name != dataset1.Name {
			t.Errorf("Expected name %s, got %s", dataset1.Name, loaded1.Name)
		}

		loaded2, err := LoadDataset(path2)
		if err != nil {
			t.Fatalf("Failed to load dataset2: %v", err)
		}

		if loaded2.Name != dataset2.Name {
			t.Errorf("Expected name %s, got %s", dataset2.Name, loaded2.Name)
		}
	})
}
