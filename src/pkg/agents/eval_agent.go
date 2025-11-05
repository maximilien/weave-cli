// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// EvalAgent evaluates query execution and tracks metrics
type EvalAgent struct {
	llmClient llm.Client
}

// NewEvalAgent creates a new EvalAgent
func NewEvalAgent(llmClient llm.Client) *EvalAgent {
	return &EvalAgent{
		llmClient: llmClient,
	}
}

// Name returns the agent's name
func (a *EvalAgent) Name() string {
	return "EvalAgent"
}

// Execute evaluates the operation report and generates metrics
func (a *EvalAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	report, ok := input.(*OperationReport)
	if !ok {
		return nil, fmt.Errorf("invalid input type for EvalAgent")
	}

	// Create evaluation metrics
	metrics := a.createMetrics(report)

	// Optionally evaluate if output matches intent
	if a.llmClient != nil {
		intentMatched, err := a.evaluateIntentMatch(ctx, report)
		if err == nil {
			metrics.IntentMatched = intentMatched
		}
	}

	return metrics, nil
}

// createMetrics creates evaluation metrics from the operation report
func (a *EvalAgent) createMetrics(report *OperationReport) *EvaluationMetrics {
	// Get LLM metrics if available
	var llmMetrics *llm.Metrics
	if a.llmClient != nil {
		llmMetrics = a.llmClient.GetMetrics()
	}

	metrics := &EvaluationMetrics{
		QueryID:   uuid.New().String(),
		Success:   report.FailedSteps == 0,
		Latency:   report.Duration,
		ErrorRate: float64(report.FailedSteps) / float64(report.ExecutedSteps),
	}

	if llmMetrics != nil {
		metrics.LLMInvocations = llmMetrics.Invocations
		metrics.TotalTokens = llmMetrics.TotalTokens
		metrics.PromptTokens = llmMetrics.PromptTokens
		metrics.CompletionTokens = llmMetrics.CompletionTokens
		metrics.TotalCost = llmMetrics.TotalCost
	}

	return metrics
}

// evaluateIntentMatch evaluates if the output matches the user's intent
func (a *EvalAgent) evaluateIntentMatch(ctx context.Context, report *OperationReport) (bool, error) {
	prompt := fmt.Sprintf(`Evaluate if the following operation successfully fulfilled the user's intent.

User Intent: %s
Operations Executed: %d
Successful Operations: %d
Failed Operations: %d
Duration: %v

Command Results:
%s

Does the outcome match the user's intent? Respond with only "true" or "false".`,
		report.QueryIntent,
		report.ExecutedSteps,
		report.SuccessfulSteps,
		report.FailedSteps,
		report.Duration,
		a.formatCommandResults(report.Commands),
	)

	response, err := a.llmClient.Complete(ctx, prompt,
		llm.WithTemperature(0.0),
		llm.WithMaxTokens(10),
	)

	if err != nil {
		return false, err
	}

	return response == "true", nil
}

// formatCommandResults formats command results for evaluation
func (a *EvalAgent) formatCommandResults(commands []CommandReport) string {
	result := ""
	for i, cmd := range commands {
		status := "Success"
		if !cmd.Success {
			status = fmt.Sprintf("Failed: %s", cmd.Error)
		}
		result += fmt.Sprintf("%d. %s - %s\n", i+1, cmd.Command, status)
	}
	return result
}

// TrackMetrics tracks metrics (placeholder for Opik integration)
func (a *EvalAgent) TrackMetrics(metrics *EvaluationMetrics) error {
	// TODO: Implement Opik integration
	// For now, we just validate the metrics
	if metrics.QueryID == "" {
		return fmt.Errorf("metrics must have a query ID")
	}

	return nil
}

// GetCostEstimate estimates the cost of a query before execution
func GetCostEstimate(plan *ExecutionPlan, model string) float64 {
	// Rough estimation based on plan complexity
	// This is a simplified version; actual costs depend on LLM usage

	baseTokens := 1000 // Base tokens for query analysis and planning
	perStepTokens := 200

	totalTokens := baseTokens + (len(plan.Steps) * perStepTokens)

	// Cost per 1M tokens (simplified)
	var costPer1M float64
	switch model {
	case "gpt-4-turbo-preview", "gpt-4-turbo":
		costPer1M = 20.0 // Average of prompt and completion
	case "gpt-4":
		costPer1M = 45.0
	case "gpt-3.5-turbo":
		costPer1M = 1.0
	default:
		costPer1M = 20.0
	}

	return (float64(totalTokens) / 1000000) * costPer1M
}

// CalculateSuccessRate calculates success rate from historical metrics
func CalculateSuccessRate(metrics []*EvaluationMetrics) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	successful := 0
	for _, m := range metrics {
		if m.Success {
			successful++
		}
	}

	return float64(successful) / float64(len(metrics)) * 100
}

// GetAverageLatency calculates average latency from historical metrics
func GetAverageLatency(metrics []*EvaluationMetrics) time.Duration {
	if len(metrics) == 0 {
		return 0
	}

	var total time.Duration
	for _, m := range metrics {
		total += m.Latency
	}

	return total / time.Duration(len(metrics))
}
