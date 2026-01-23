// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// Evaluator interface for all evaluators
type Evaluator interface {
	Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error)
	Name() string
}

// AccuracyEvaluator evaluates semantic similarity between expected and actual answers
type AccuracyEvaluator struct {
	llmClient llm.Client
}

// NewAccuracyEvaluator creates a new accuracy evaluator
func NewAccuracyEvaluator(llmClient llm.Client) *AccuracyEvaluator {
	return &AccuracyEvaluator{
		llmClient: llmClient,
	}
}

// Name returns the evaluator name
func (e *AccuracyEvaluator) Name() string {
	return "accuracy"
}

// Evaluate computes semantic similarity score
func (e *AccuracyEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Build prompt for LLM-as-judge
	prompt := fmt.Sprintf(`You are an evaluation assistant. Compare the expected and actual answers for semantic similarity.

Question: %s

Expected Answer:
%s

Actual Answer:
%s

Rate the semantic similarity on a scale of 0.0 to 1.0, where:
- 1.0 = Answers are semantically identical, convey the same meaning
- 0.8-0.9 = Answers are very similar, minor differences in wording
- 0.6-0.7 = Answers cover the same concepts but with some differences
- 0.4-0.5 = Answers are somewhat related but miss key points
- 0.2-0.3 = Answers are loosely related
- 0.0-0.1 = Answers are completely different or incorrect

Respond with ONLY a number between 0.0 and 1.0.`, testCase.Query, testCase.ExpectedAnswer, actual)

	// Call LLM
	response, err := e.llmClient.Complete(ctx, prompt, llm.WithMaxTokens(10))
	if err != nil {
		return 0.0, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse score
	score := parseScore(response)
	return score, nil
}

// CitationEvaluator evaluates citation quality
type CitationEvaluator struct{}

// NewCitationEvaluator creates a new citation evaluator
func NewCitationEvaluator() *CitationEvaluator {
	return &CitationEvaluator{}
}

// Name returns the evaluator name
func (e *CitationEvaluator) Name() string {
	return "citation"
}

// Evaluate checks citation presence and format
func (e *CitationEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// If citations aren't required, return perfect score
	if !testCase.MustCite {
		return 1.0, nil
	}

	// Check if answer contains citation markers
	citationRegex := regexp.MustCompile(`\[\d+\]`)
	matches := citationRegex.FindAllString(actual, -1)

	if len(matches) == 0 {
		// No citations found
		return 0.0, nil
	}

	// Score based on number of citations (up to 5)
	citationCount := len(matches)
	baseScore := 0.5 // Base score for having any citations

	// Bonus for multiple citations (shows comprehensive sourcing)
	bonusScore := 0.0
	if citationCount >= 2 {
		bonusScore = 0.2
	}
	if citationCount >= 3 {
		bonusScore = 0.3
	}
	if citationCount >= 5 {
		bonusScore = 0.5
	}

	score := baseScore + bonusScore
	if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// HallucinationDetector detects potential hallucinations
type HallucinationDetector struct {
	llmClient llm.Client
}

// NewHallucinationDetector creates a new hallucination detector
func NewHallucinationDetector(llmClient llm.Client) *HallucinationDetector {
	return &HallucinationDetector{
		llmClient: llmClient,
	}
}

// Name returns the evaluator name
func (e *HallucinationDetector) Name() string {
	return "hallucination"
}

// Evaluate detects hallucinations (returns score where 1.0 = no hallucination)
func (e *HallucinationDetector) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// Check for required concepts
	if len(testCase.RequiredConcepts) > 0 {
		score := checkRequiredConcepts(actual, testCase.RequiredConcepts)
		if score < 1.0 {
			// Missing concepts indicates potential hallucination
			return score, nil
		}
	}

	// Use LLM-as-judge for hallucination detection
	prompt := fmt.Sprintf(`You are an evaluation assistant. Determine if the answer contains hallucinations (made-up or unsupported information).

Question: %s

Expected Answer (ground truth):
%s

Actual Answer:
%s

Rate hallucination on a scale where:
- 1.0 = No hallucination, answer is factual and supported
- 0.8-0.9 = Minor unsupported details but mostly accurate
- 0.6-0.7 = Some unsupported claims
- 0.4-0.5 = Significant unsupported information
- 0.2-0.3 = Mostly hallucinated content
- 0.0-0.1 = Completely hallucinated or false

Respond with ONLY a number between 0.0 and 1.0.`, testCase.Query, testCase.ExpectedAnswer, actual)

	// Call LLM
	response, err := e.llmClient.Complete(ctx, prompt, llm.WithMaxTokens(10))
	if err != nil {
		// Fall back to concept checking
		if len(testCase.RequiredConcepts) > 0 {
			return checkRequiredConcepts(actual, testCase.RequiredConcepts), nil
		}
		return 1.0, nil // Default to no hallucination if can't check
	}

	// Parse score
	score := parseScore(response)
	return score, nil
}

// Helper function to check required concepts
func checkRequiredConcepts(answer string, requiredConcepts []string) float64 {
	if len(requiredConcepts) == 0 {
		return 1.0
	}

	answerLower := strings.ToLower(answer)
	matchedCount := 0

	for _, concept := range requiredConcepts {
		conceptLower := strings.ToLower(concept)
		if strings.Contains(answerLower, conceptLower) {
			matchedCount++
		}
	}

	return float64(matchedCount) / float64(len(requiredConcepts))
}

// Helper function to parse score from LLM response
func parseScore(response string) float64 {
	// Clean response
	response = strings.TrimSpace(response)

	// Try to parse as float
	var score float64
	_, err := fmt.Sscanf(response, "%f", &score)
	if err != nil {
		// Try to extract number from response
		numberRegex := regexp.MustCompile(`\d+\.\d+|\d+`)
		match := numberRegex.FindString(response)
		if match != "" {
			fmt.Sscanf(match, "%f", &score)
		}
	}

	// Clamp to 0.0-1.0 range
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// EvaluateTestCase runs all evaluators on a test case
func EvaluateTestCase(ctx context.Context, testCase *TestCase, actualAnswer string, actualCitations []string, llmClient llm.Client) (*TestCaseResult, error) {
	result := &TestCaseResult{
		TestCaseID:      testCase.ID,
		Query:           testCase.Query,
		ActualAnswer:    actualAnswer,
		ActualCitations: actualCitations,
		Details:         make(map[string]interface{}),
	}

	// Run evaluators
	accuracyEval := NewAccuracyEvaluator(llmClient)
	citationEval := NewCitationEvaluator()
	hallucinationEval := NewHallucinationDetector(llmClient)

	var err error

	// Accuracy
	result.AccuracyScore, err = accuracyEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Accuracy evaluation failed: %v", err))
	}

	// Citation
	result.CitationScore, err = citationEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Citation evaluation failed: %v", err))
	}

	// Hallucination
	result.HallucinationScore, err = hallucinationEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Hallucination evaluation failed: %v", err))
	}

	// Determine pass/fail
	result.Passed = true

	if testCase.MinRelevanceScore > 0 && result.AccuracyScore < testCase.MinRelevanceScore {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Accuracy score %.2f below threshold %.2f", result.AccuracyScore, testCase.MinRelevanceScore))
	}

	if testCase.MustCite && result.CitationScore < 0.5 {
		result.Passed = false
		result.Errors = append(result.Errors, "Missing required citations")
	}

	if result.HallucinationScore < 0.6 {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("High hallucination risk: score %.2f", result.HallucinationScore))
	}

	return result, nil
}
