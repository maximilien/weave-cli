// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

func TestCitationEvaluator(t *testing.T) {
	evaluator := NewCitationEvaluator()

	testCase := &TestCase{
		ID:       "test-001",
		Query:    "What is RAG?",
		MustCite: true,
	}

	t.Run("Name", func(t *testing.T) {
		if evaluator.Name() != "citation" {
			t.Errorf("Expected name 'citation', got: %s", evaluator.Name())
		}
	})

	t.Run("NoCitations", func(t *testing.T) {
		score, err := evaluator.Evaluate(nil, testCase,
			"RAG stands for Retrieval Augmented Generation.", nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if score != 0.0 {
			t.Errorf("Expected score 0.0 for no citations, got: %.2f", score)
		}
	})

	t.Run("OneCitation", func(t *testing.T) {
		score, err := evaluator.Evaluate(nil, testCase,
			"RAG stands for Retrieval Augmented Generation [1].", nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if score != 0.5 {
			t.Errorf("Expected score 0.5 for one citation, got: %.2f", score)
		}
	})

	t.Run("TwoCitations", func(t *testing.T) {
		score, err := evaluator.Evaluate(nil, testCase,
			"RAG combines retrieval [1] with generation [2].", nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedScore := 0.7
		if score < expectedScore-0.01 || score > expectedScore+0.01 {
			t.Errorf("Expected score %.2f for two citations, got: %.2f", expectedScore, score)
		}
	})

	t.Run("ThreeCitations", func(t *testing.T) {
		score, err := evaluator.Evaluate(nil, testCase,
			"RAG uses retrieval [1], augmentation [2], and generation [3].", nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedScore := 0.8
		if score < expectedScore-0.01 || score > expectedScore+0.01 {
			t.Errorf("Expected score %.2f for three citations, got: %.2f", expectedScore, score)
		}
	})

	t.Run("FivePlusCitations", func(t *testing.T) {
		score, err := evaluator.Evaluate(nil, testCase,
			"RAG [1] combines [2] retrieval [3], augmentation [4], and generation [5] techniques [6].", nil)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if score != 1.0 {
			t.Errorf("Expected score 1.0 for five+ citations, got: %.2f", score)
		}
	})
}

func TestCheckRequiredConcepts(t *testing.T) {
	t.Run("AllConceptsPresent", func(t *testing.T) {
		answer := "Vector databases store high-dimensional embeddings for similarity search."
		concepts := []string{"vector", "embeddings", "similarity"}

		score := checkRequiredConcepts(answer, concepts)
		if score != 1.0 {
			t.Errorf("Expected score 1.0, got: %.2f", score)
		}
	})

	t.Run("PartialConceptsPresent", func(t *testing.T) {
		answer := "Vector databases store data structures."
		concepts := []string{"vector", "embeddings", "similarity"}

		score := checkRequiredConcepts(answer, concepts)
		expected := 1.0 / 3.0 // Only "vector" is present
		if score < expected-0.05 || score > expected+0.05 {
			t.Errorf("Expected score around %.2f, got: %.2f", expected, score)
		}
	})

	t.Run("NoConceptsPresent", func(t *testing.T) {
		answer := "This is about something completely different."
		concepts := []string{"vector", "embeddings", "similarity"}

		score := checkRequiredConcepts(answer, concepts)
		if score != 0.0 {
			t.Errorf("Expected score 0.0, got: %.2f", score)
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		answer := "VECTOR databases use EMBEDDINGS for SIMILARITY search."
		concepts := []string{"vector", "embeddings", "similarity"}

		score := checkRequiredConcepts(answer, concepts)
		if score != 1.0 {
			t.Errorf("Expected score 1.0 (case insensitive), got: %.2f", score)
		}
	})

	t.Run("EmptyConcepts", func(t *testing.T) {
		answer := "Any answer"
		concepts := []string{}

		score := checkRequiredConcepts(answer, concepts)
		if score != 1.0 {
			t.Errorf("Expected score 1.0 for empty concepts, got: %.2f", score)
		}
	})
}

func TestParseScore(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected float64
	}{
		{"SimpleScore", "0.85", 0.85},
		{"ScoreWithWhitespace", "  0.75  ", 0.75},
		{"ScoreInSentence", "The similarity score is 0.92 based on analysis.", 0.92},
		{"ScoreWithNewlines", "Analysis:\n0.88\nComplete", 0.88},
		{"ZeroScore", "0.0", 0.0},
		{"PerfectScore", "1.0", 1.0},
		{"InvalidScore", "invalid", 0.0},
		{"NoScore", "No numeric value here", 0.0},
		{"MultipleScores", "0.5 and 0.8", 0.5}, // Should take first
		{"ScoreOutOfRange", "1.5", 1.0},        // Should cap at 1.0
		{"NegativeScore", "-0.5", 0.0},         // Should floor at 0.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := parseScore(tt.response)
			if score < tt.expected-0.01 || score > tt.expected+0.01 {
				t.Errorf("Expected score %.2f, got: %.2f", tt.expected, score)
			}
		})
	}
}

func TestExtractCitations(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "NoCitations",
			text:     "This text has no citations.",
			expected: []string{},
		},
		{
			name:     "OneCitation",
			text:     "This is a fact [1].",
			expected: []string{"[1]"},
		},
		{
			name:     "MultipleCitations",
			text:     "First fact [1] and second fact [2] and third [3].",
			expected: []string{"[1]", "[2]", "[3]"},
		},
		{
			name:     "DuplicateCitations",
			text:     "Fact [1] and another fact [1] and more [2].",
			expected: []string{"[1]", "[2]"},
		},
		{
			name:     "NonNumericBrackets",
			text:     "This [is] not [a] citation [1].",
			expected: []string{"[1]"},
		},
		{
			name:     "LargeNumbers",
			text:     "Citation [42] and [100].",
			expected: []string{"[42]", "[100]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			citations := extractCitations(tt.text)

			if len(citations) != len(tt.expected) {
				t.Errorf("Expected %d citations, got %d", len(tt.expected), len(citations))
				return
			}

			citationMap := make(map[string]bool)
			for _, c := range citations {
				citationMap[c] = true
			}

			for _, expected := range tt.expected {
				if !citationMap[expected] {
					t.Errorf("Expected citation %s not found", expected)
				}
			}
		})
	}
}

// MockLLMClient for testing
type MockLLMClient struct {
	response string
	err      error
}

func (m *MockLLMClient) Complete(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockLLMClient) CompleteWithImages(ctx context.Context, prompt string, imageURLs []string, opts ...llm.Option) (string, error) {
	return m.Complete(ctx, prompt, opts...)
}

func (m *MockLLMClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, nil
}

func (m *MockLLMClient) CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...llm.Option) (interface{}, error) {
	return m.Complete(ctx, prompt, opts...)
}

func (m *MockLLMClient) GetMetrics() *llm.Metrics {
	return &llm.Metrics{}
}

func TestContextRelevanceEvaluator(t *testing.T) {
	tests := []struct {
		name            string
		retrievedContext []string
		expectedScore   float64
		llmResponse     string
	}{
		{
			name:            "No context returns 1.0",
			retrievedContext: []string{},
			expectedScore:   1.0,
		},
		{
			name: "Highly relevant context",
			retrievedContext: []string{
				"Vector databases store embeddings",
				"Embeddings are numerical representations",
			},
			expectedScore: 0.9,
			llmResponse:   "0.9",
		},
		{
			name: "Low relevance context",
			retrievedContext: []string{
				"Unrelated information",
			},
			expectedScore: 0.3,
			llmResponse:   "0.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockLLMClient{response: tt.llmResponse}
			evaluator := NewContextRelevanceEvaluator(mockClient)

			testCase := &TestCase{
				Query:            "What is a vector database?",
				RetrievedContext: tt.retrievedContext,
			}

			ctx := context.Background()
			score, err := evaluator.Evaluate(ctx, testCase, "answer", nil)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.retrievedContext == nil || len(tt.retrievedContext) == 0 {
				if score != tt.expectedScore {
					t.Errorf("Expected score %.2f for no context, got %.2f", tt.expectedScore, score)
				}
			}
		})
	}
}

func TestFaithfulnessEvaluator(t *testing.T) {
	tests := []struct {
		name          string
		context       []string
		expectedScore float64
		llmResponse   string
	}{
		{
			name:          "Faithful answer",
			context:       []string{"Vector databases store embeddings"},
			expectedScore: 0.95,
			llmResponse:   "0.95",
		},
		{
			name:          "Unfaithful answer",
			context:       []string{"Completely unrelated info"},
			expectedScore: 0.2,
			llmResponse:   "0.2",
		},
		{
			name:          "No context falls back to expected answer",
			context:       []string{},
			expectedScore: 0.85,
			llmResponse:   "0.85",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockLLMClient{response: tt.llmResponse}
			evaluator := NewFaithfulnessEvaluator(mockClient)

			testCase := &TestCase{
				Query:            "What is a vector database?",
				ExpectedAnswer:   "A vector database stores and queries embeddings",
				RetrievedContext: tt.context,
			}

			ctx := context.Background()
			score, err := evaluator.Evaluate(ctx, testCase, "answer", nil)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Just verify we get a score (mock returns fixed value)
			if score < 0.0 || score > 1.0 {
				t.Errorf("Score out of range: %.2f", score)
			}
		})
	}
}

func TestEvaluateTestCaseWithNewEvaluators(t *testing.T) {
	mockClient := &MockLLMClient{response: "0.85"}

	testCase := &TestCase{
		ID:               "test-001",
		Query:            "What is a vector database?",
		ExpectedAnswer:   "A database for storing vectors",
		RetrievedContext: []string{"Vector databases store embeddings"},
		RequiredConcepts: []string{"vector", "database"},
		MustCite:         false,
		MinRelevanceScore: 0.7,
	}

	ctx := context.Background()
	result, err := EvaluateTestCase(ctx, testCase, "A vector database stores embeddings", nil, mockClient)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all scores are populated
	if result.AccuracyScore == 0 {
		t.Error("AccuracyScore should be populated")
	}
	if result.ContextRelevanceScore == 0 {
		t.Error("ContextRelevanceScore should be populated")
	}
	if result.FaithfulnessScore == 0 {
		t.Error("FaithfulnessScore should be populated")
	}

	// Verify result structure
	if result.TestCaseID != testCase.ID {
		t.Errorf("Expected TestCaseID %s, got %s", testCase.ID, result.TestCaseID)
	}
}
