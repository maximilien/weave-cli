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

// ContextRelevanceEvaluator evaluates how relevant retrieved context is to the query
type ContextRelevanceEvaluator struct {
	llmClient llm.Client
}

// NewContextRelevanceEvaluator creates a new context relevance evaluator
func NewContextRelevanceEvaluator(llmClient llm.Client) *ContextRelevanceEvaluator {
	return &ContextRelevanceEvaluator{
		llmClient: llmClient,
	}
}

// Name returns the evaluator name
func (e *ContextRelevanceEvaluator) Name() string {
	return "context_relevance"
}

// Evaluate scores how relevant the retrieved context is to the query
func (e *ContextRelevanceEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// If no context provided, return perfect score (not applicable)
	if len(testCase.RetrievedContext) == 0 {
		return 1.0, nil
	}

	// Evaluate each context chunk for relevance
	totalScore := 0.0
	evaluatedChunks := 0

	for i, chunk := range testCase.RetrievedContext {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		prompt := fmt.Sprintf(`You are an evaluation assistant. Rate how relevant this context is to answering the query.

Query: %s

Context Chunk %d:
%s

Rate the relevance on a scale of 0.0 to 1.0, where:
- 1.0 = Highly relevant, directly answers or helps answer the query
- 0.8-0.9 = Very relevant, contains useful information
- 0.6-0.7 = Somewhat relevant, tangentially related
- 0.4-0.5 = Marginally relevant, weak connection
- 0.2-0.3 = Barely relevant, mostly unrelated
- 0.0-0.1 = Not relevant, completely unrelated

Respond with ONLY a number between 0.0 and 1.0.`, testCase.Query, i+1, chunk)

		response, err := e.llmClient.Complete(ctx, prompt, llm.WithMaxTokens(10))
		if err != nil {
			// Skip this chunk if evaluation fails
			continue
		}

		score := parseScore(response)
		totalScore += score
		evaluatedChunks++
	}

	if evaluatedChunks == 0 {
		return 1.0, nil // Default if no chunks could be evaluated
	}

	// Return average relevance score
	avgScore := totalScore / float64(evaluatedChunks)
	return avgScore, nil
}

// FaithfulnessEvaluator evaluates whether the answer is supported by the retrieved context
type FaithfulnessEvaluator struct {
	llmClient llm.Client
}

// NewFaithfulnessEvaluator creates a new faithfulness evaluator
func NewFaithfulnessEvaluator(llmClient llm.Client) *FaithfulnessEvaluator {
	return &FaithfulnessEvaluator{
		llmClient: llmClient,
	}
}

// Name returns the evaluator name
func (e *FaithfulnessEvaluator) Name() string {
	return "faithfulness"
}

// Evaluate verifies that the answer is supported by the retrieved context
func (e *FaithfulnessEvaluator) Evaluate(ctx context.Context, testCase *TestCase, actual string, actualCitations []string) (float64, error) {
	// If no context provided, use expected answer as reference
	context := testCase.RetrievedContext
	if len(context) == 0 {
		// Fall back to comparing with expected answer
		return e.evaluateAgainstExpected(ctx, testCase, actual)
	}

	// Combine all context chunks
	combinedContext := strings.Join(context, "\n\n")

	prompt := fmt.Sprintf(`You are an evaluation assistant. Determine if the answer is fully supported by the provided context.

Query: %s

Context:
%s

Answer:
%s

Rate the faithfulness on a scale of 0.0 to 1.0, where:
- 1.0 = All claims in the answer are fully supported by context
- 0.8-0.9 = Answer is mostly supported, minor unsupported details
- 0.6-0.7 = Answer has some unsupported claims
- 0.4-0.5 = Significant portions are not supported by context
- 0.2-0.3 = Most of answer is unsupported
- 0.0-0.1 = Answer is completely unsupported or contradicts context

Consider:
- Are all factual claims backed by the context?
- Does the answer introduce information not in context?
- Does the answer misrepresent or distort context?

Respond with ONLY a number between 0.0 and 1.0.`, testCase.Query, combinedContext, actual)

	response, err := e.llmClient.Complete(ctx, prompt, llm.WithMaxTokens(10))
	if err != nil {
		return 0.0, fmt.Errorf("LLM call failed: %w", err)
	}

	score := parseScore(response)
	return score, nil
}

// evaluateAgainstExpected is a fallback when no context is provided
func (e *FaithfulnessEvaluator) evaluateAgainstExpected(ctx context.Context, testCase *TestCase, actual string) (float64, error) {
	prompt := fmt.Sprintf(`You are an evaluation assistant. Determine if the actual answer stays faithful to the expected answer (no hallucinations or unsupported claims).

Query: %s

Expected Answer (ground truth):
%s

Actual Answer:
%s

Rate the faithfulness on a scale of 0.0 to 1.0, where:
- 1.0 = Actual answer is faithful to expected, no unsupported claims
- 0.8-0.9 = Mostly faithful, minor deviations
- 0.6-0.7 = Some unsupported information added
- 0.4-0.5 = Significant unfaithful content
- 0.2-0.3 = Mostly unfaithful
- 0.0-0.1 = Completely unfaithful or contradictory

Respond with ONLY a number between 0.0 and 1.0.`, testCase.Query, testCase.ExpectedAnswer, actual)

	response, err := e.llmClient.Complete(ctx, prompt, llm.WithMaxTokens(10))
	if err != nil {
		return 0.0, fmt.Errorf("LLM call failed: %w", err)
	}

	score := parseScore(response)
	return score, nil
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
	contextRelEval := NewContextRelevanceEvaluator(llmClient)
	faithfulnessEval := NewFaithfulnessEvaluator(llmClient)

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

	// Context Relevance
	result.ContextRelevanceScore, err = contextRelEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Context relevance evaluation failed: %v", err))
	}

	// Faithfulness
	result.FaithfulnessScore, err = faithfulnessEval.Evaluate(ctx, testCase, actualAnswer, actualCitations)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Faithfulness evaluation failed: %v", err))
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
