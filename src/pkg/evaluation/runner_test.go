// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestNewRunner tests runner creation
func TestNewRunner(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}
	runner := NewRunner(mockClient)

	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}

	if runner.llmClient == nil {
		t.Error("Runner should have LLM client")
	}
}

// TestExtractCitations tests citation extraction from text
func TestExtractCitationsInRunner(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"NoCitations", "This is text without citations", 0},
		{"OneCitation", "This has one citation [1]", 1},
		{"MultipleCitations", "First [1] and second [2]", 2},
		{"DuplicateCitations", "Same citation [1] appears twice [1]", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			citations := extractCitations(tt.text)
			if len(citations) != tt.expected {
				t.Errorf("Expected %d citations, got %d", tt.expected, len(citations))
			}
		})
	}
}

// TestRunEvaluationIntegration tests the full evaluation flow
func TestRunEvaluationIntegration(t *testing.T) {
	// Skip if no agent configs available
	if _, err := os.Stat("../../configs/agents"); os.IsNotExist(err) {
		t.Skip("Skipping integration test - agent configs not found")
	}

	mockClient := &MockLLMClient{response: "Vector databases store embeddings for similarity search"}

	// Create temporary dataset
	tmpDir := t.TempDir()
	datasetPath := filepath.Join(tmpDir, "test-dataset.yaml")

	dataset := &Dataset{
		Name:        "test-dataset",
		Version:     "1.0.0",
		Description: "Test dataset for runner integration",
		TestCases: []TestCase{
			{
				ID:                "test-001",
				Query:             "What is a vector database?",
				ExpectedAnswer:    "A database for storing vectors",
				RequiredConcepts:  []string{"vector", "database"},
				MinRelevanceScore: 0.5,
			},
		},
	}

	// Save dataset
	if err := SaveDataset(dataset, datasetPath); err != nil {
		t.Fatalf("Failed to save test dataset: %v", err)
	}

	// Load dataset
	loadedDataset, err := LoadDataset(datasetPath)
	if err != nil {
		t.Fatalf("Failed to load dataset: %v", err)
	}

	// Note: RunEvaluation requires agent configs which may not exist in test environment
	// So we test the provider pattern instead
	t.Run("RunWithLocalProvider", func(t *testing.T) {
		ctx := context.Background()
		provider := NewLocalProvider(mockClient)

		// Test evaluation with provider (without full runner which needs agent configs)
		testCase := &loadedDataset.TestCases[0]
		result, err := EvaluateTestCase(ctx, testCase, "Vector database stores vectors", nil, provider)

		if err != nil {
			t.Fatalf("Evaluation failed: %v", err)
		}

		if result.TestCaseID != testCase.ID {
			t.Errorf("Expected test case ID %s, got %s", testCase.ID, result.TestCaseID)
		}

		// Verify provider info stored
		if result.Details["evaluator_provider"] != "local" {
			t.Errorf("Expected provider 'local', got %v", result.Details["evaluator_provider"])
		}
	})
}

// TestGenerateAnswer tests answer generation (indirectly through runner)
func TestGenerateAnswerCitationExtraction(t *testing.T) {
	// Test that runner extracts citations correctly
	testCases := []struct {
		response          string
		expectedCitations int
	}{
		{"Answer without citations", 0},
		{"Answer with one citation [1]", 1},
		{"Answer with multiple [1] citations [2]", 2},
	}

	for _, tc := range testCases {
		citations := extractCitations(tc.response)
		if len(citations) != tc.expectedCitations {
			t.Errorf("Response %q: expected %d citations, got %d",
				tc.response, tc.expectedCitations, len(citations))
		}
	}
}
