// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"testing"
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
