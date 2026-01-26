// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// Runner executes evaluation runs
type Runner struct {
	llmClient llm.Client
}

// NewRunner creates a new evaluation runner
func NewRunner(llmClient llm.Client) *Runner {
	return &Runner{
		llmClient: llmClient,
	}
}

// RunEvaluation executes a complete evaluation using the default local provider
func (r *Runner) RunEvaluation(ctx context.Context, dataset *Dataset, agentName string, collection string) (*EvaluationRun, error) {
	// Use local provider by default
	provider := NewLocalProvider(r.llmClient)
	return r.RunEvaluationWithProvider(ctx, dataset, agentName, collection, provider)
}

// RunEvaluationWithProvider executes a complete evaluation using the specified provider
func (r *Runner) RunEvaluationWithProvider(ctx context.Context, dataset *Dataset, agentName string, collection string, provider EvaluatorProvider) (*EvaluationRun, error) {
	// Load agent
	agentConfig, err := agents.LoadAgent(agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent: %w", err)
	}

	// Create evaluation run
	run := &EvaluationRun{
		ID:          GenerateRunID(),
		Timestamp:   time.Now(),
		DatasetName: dataset.Name,
		AgentName:   agentName,
		Collection:  collection,
		Config: EvaluationConfig{
			AgentPath:  agentName,
			Parameters: make(map[string]string),
		},
		Results: make([]TestCaseResult, 0, len(dataset.TestCases)),
	}

	// Add provider info to config
	run.Config.Parameters["evaluator_provider"] = provider.Name()

	startTime := time.Now()

	// Run each test case
	for i, testCase := range dataset.TestCases {
		fmt.Printf("Running test case %d/%d: %s\n", i+1, len(dataset.TestCases), testCase.ID)

		testStart := time.Now()

		// Generate answer using agent (simulated for now)
		actualAnswer, actualCitations, err := r.generateAnswer(ctx, &testCase, agentConfig, collection)
		if err != nil {
			// Record error but continue
			result := &TestCaseResult{
				TestCaseID:   testCase.ID,
				Query:        testCase.Query,
				ActualAnswer: "",
				Passed:       false,
				Errors:       []string{fmt.Sprintf("Failed to generate answer: %v", err)},
			}
			run.Results = append(run.Results, *result)
			continue
		}

		testEnd := time.Now()

		// Evaluate the result using the specified provider
		result, err := EvaluateTestCase(ctx, &testCase, actualAnswer, actualCitations, provider)
		if err != nil {
			result = &TestCaseResult{
				TestCaseID:   testCase.ID,
				Query:        testCase.Query,
				ActualAnswer: actualAnswer,
				Passed:       false,
				Errors:       []string{fmt.Sprintf("Evaluation failed: %v", err)},
			}
		}

		// Record timing
		result.ResponseTime = float64(testEnd.Sub(testStart).Milliseconds())

		run.Results = append(run.Results, *result)
	}

	endTime := time.Now()

	// Calculate summary
	run.Summary = CalculateSummary(run.Results, startTime, endTime)

	return run, nil
}

// generateAnswer generates an answer using the agent (simplified for Phase 1)
func (r *Runner) generateAnswer(ctx context.Context, testCase *TestCase, agentConfig *agents.CustomAgentConfig, collection string) (string, []string, error) {
	// Build prompt with agent's system prompt
	prompt := fmt.Sprintf(`%s

Question: %s

Please provide a detailed answer with citations.`, agentConfig.SystemPrompt, testCase.Query)

	// Generate response
	response, err := r.llmClient.Complete(ctx, prompt,
		llm.WithModel(agentConfig.LLM.Model),
		llm.WithTemperature(agentConfig.LLM.Temperature),
		llm.WithMaxTokens(agentConfig.LLM.MaxTokens),
	)
	if err != nil {
		return "", nil, err
	}

	// Extract citations from response
	citations := extractCitations(response)

	return response, citations, nil
}

// extractCitations extracts citation markers from text
func extractCitations(text string) []string {
	citationRegex := regexp.MustCompile(`\[\d+\]`)
	matches := citationRegex.FindAllString(text, -1)

	// Remove duplicates
	citationMap := make(map[string]bool)
	var citations []string
	for _, match := range matches {
		if !citationMap[match] {
			citationMap[match] = true
			citations = append(citations, match)
		}
	}

	return citations
}
