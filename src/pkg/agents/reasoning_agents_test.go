// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

type reasoningLLM struct {
	completion string
	structured interface{}
	err        error
	metrics    *llm.Metrics
	prompt     string
	options    *llm.CompletionOptions
}

func (m *reasoningLLM) Complete(_ context.Context, prompt string, opts ...llm.Option) (string, error) {
	m.capture(prompt, opts)
	return m.completion, m.err
}

func (m *reasoningLLM) CompleteStructured(_ context.Context, prompt string, target interface{}, opts ...llm.Option) (interface{}, error) {
	m.capture(prompt, opts)
	if m.err != nil {
		return nil, m.err
	}
	data, err := json.Marshal(m.structured)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, err
	}
	return target, nil
}

func (m *reasoningLLM) GetMetrics() *llm.Metrics { return m.metrics }

func (m *reasoningLLM) capture(prompt string, opts []llm.Option) {
	m.prompt = prompt
	m.options = llm.DefaultCompletionOptions()
	for _, opt := range opts {
		opt(m.options)
	}
}

func TestQueryAgentStructuredLifecycle(t *testing.T) {
	model := &reasoningLLM{structured: QueryAgentOutput{
		IsWeaveQuery: true, FixedQuery: "list collections", Intent: "list", Confidence: 0.98, Reason: "collection operation",
	}}
	agent := NewQueryAgent(model)
	agent.SetMCPTools([]string{"list_collections", "count_documents"})
	if agent.Name() != "QueryAgent" || !strings.Contains(agent.getSystemMessage(), "valid JSON") {
		t.Fatalf("query agent identity/system prompt is invalid")
	}
	result, err := agent.Execute(context.Background(), &QueryAgentInput{Query: "show my colls"})
	output, ok := result.(*QueryAgentOutput)
	if err != nil || !ok || !output.IsWeaveQuery || output.Intent != "list" {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if !strings.Contains(model.prompt, "list_collections, count_documents") || model.options.Temperature != 0.1 {
		t.Errorf("LLM request = %q, %#v", model.prompt, model.options)
	}
	if !strings.Contains(NewQueryAgent(model).buildPrompt("health"), `User query: "health"`) {
		t.Fatal("buildPrompt() omitted query")
	}

	var parsed QueryAgentOutput
	if err := ParseJSONResponse(`{"intent":"health"}`, &parsed); err != nil || parsed.Intent != "health" {
		t.Fatalf("ParseJSONResponse() = %#v, %v", parsed, err)
	}
	if err := ParseJSONResponse("not-json", &parsed); err == nil {
		t.Fatal("ParseJSONResponse() accepted invalid JSON")
	}
	if result, err := agent.Execute(context.Background(), "bad input"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}
	model.err = errors.New("provider offline")
	if result, err := agent.Execute(context.Background(), &QueryAgentInput{Query: "health"}); err == nil || result != nil {
		t.Fatalf("Execute(provider failure) = %#v, %v", result, err)
	}
}

func TestPlanningAgentStructuredLifecycle(t *testing.T) {
	plan := ExecutionPlan{Summary: "list collections", Steps: []ExecutionStep{{Type: "weave", Command: "list_collections"}}}
	model := &reasoningLLM{structured: plan}
	agent := NewPlanningAgent(model, []string{"list_collections"})
	if agent.Name() != "PlanningAgent" || !strings.Contains(agent.getSystemMessage(), "execution plans") {
		t.Fatal("planning agent identity/system prompt is invalid")
	}
	result, err := agent.Execute(context.Background(), &PlanningAgentInput{FixedQuery: "list collections", Intent: "list"})
	output, ok := result.(*ExecutionPlan)
	if err != nil || !ok || output.Summary != plan.Summary || len(output.Steps) != 1 {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if !strings.Contains(model.prompt, "list_collections") || model.options.MaxTokens != 4096 {
		t.Errorf("LLM request = %q, %#v", model.prompt, model.options)
	}

	detailed := NewPlanningAgentWithTools(model, []mcp.Tool{{
		Name: "delete_collection", Description: "Delete one collection",
		InputSchema: map[string]interface{}{"type": "object", "required": []string{"name"}},
	}})
	prompt := detailed.buildPrompt("delete docs", "delete")
	if !strings.Contains(prompt, "Delete one collection") || !strings.Contains(prompt, `"required"`) {
		t.Fatalf("detailed buildPrompt() = %q", prompt)
	}
	if safe := getSafeBashCommands(); len(safe) < 5 || safe[0] != "ls" {
		t.Fatalf("getSafeBashCommands() = %#v", safe)
	}
	if result, err := agent.Execute(context.Background(), "bad input"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}
	model.err = errors.New("provider offline")
	if result, err := agent.Execute(context.Background(), &PlanningAgentInput{}); err == nil || result != nil {
		t.Fatalf("Execute(provider failure) = %#v, %v", result, err)
	}
}

func TestEvalAgentMetricsAndIntentEvaluation(t *testing.T) {
	model := &reasoningLLM{
		completion: "true",
		metrics:    &llm.Metrics{Invocations: 3, TotalTokens: 120, PromptTokens: 80, CompletionTokens: 40, TotalCost: 0.01},
	}
	agent := NewEvalAgent(model)
	if agent.Name() != "EvalAgent" {
		t.Fatalf("Name() = %q", agent.Name())
	}
	report := &OperationReport{
		QueryIntent: "list collections", ExecutedSteps: 2, SuccessfulSteps: 1, FailedSteps: 1, Duration: 2 * time.Second,
		Commands: []CommandReport{{Command: "weave cols ls", Success: true}, {Command: "weave docs count docs", Error: "offline"}},
	}
	result, err := agent.Execute(context.Background(), report)
	metrics, ok := result.(*EvaluationMetrics)
	if err != nil || !ok || metrics.Success || !metrics.IntentMatched || metrics.ErrorRate != 0.5 || metrics.TotalTokens != 120 {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if metrics.QueryID == "" || metrics.Latency != 2*time.Second || !strings.Contains(model.prompt, "Failed: offline") {
		t.Fatalf("metrics/prompt = %#v, %q", metrics, model.prompt)
	}
	if model.options.Temperature != 0 || model.options.MaxTokens != 10 {
		t.Errorf("evaluation options = %#v", model.options)
	}
	if err := agent.TrackMetrics(metrics); err != nil {
		t.Fatalf("TrackMetrics(): %v", err)
	}
	if err := agent.TrackMetrics(&EvaluationMetrics{}); err == nil {
		t.Fatal("TrackMetrics() accepted missing query ID")
	}

	model.err = errors.New("evaluation unavailable")
	result, err = agent.Execute(context.Background(), report)
	metrics = result.(*EvaluationMetrics)
	if err != nil || metrics.IntentMatched {
		t.Fatalf("Execute(evaluation failure) = %#v, %v", result, err)
	}
	if result, err := agent.Execute(context.Background(), "bad input"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}

	withoutLLM := NewEvalAgent(nil).createMetrics(&OperationReport{ExecutedSteps: 1})
	if withoutLLM.QueryID == "" || !withoutLLM.Success {
		t.Fatalf("createMetrics(no LLM) = %#v", withoutLLM)
	}
}

func TestEvaluationHelpers(t *testing.T) {
	plan := &ExecutionPlan{Steps: make([]ExecutionStep, 2)}
	for _, model := range []string{"gpt-4-turbo", "gpt-4-turbo-preview", "gpt-4", "gpt-3.5-turbo", "unknown"} {
		if cost := GetCostEstimate(plan, model); cost <= 0 {
			t.Errorf("GetCostEstimate(%q) = %f", model, cost)
		}
	}
	if got := CalculateSuccessRate(nil); got != 0 {
		t.Errorf("CalculateSuccessRate(nil) = %f", got)
	}
	metrics := []*EvaluationMetrics{{Success: true, Latency: time.Second}, {Latency: 3 * time.Second}}
	if got := CalculateSuccessRate(metrics); got != 50 {
		t.Errorf("CalculateSuccessRate() = %f", got)
	}
	if got := GetAverageLatency(nil); got != 0 {
		t.Errorf("GetAverageLatency(nil) = %v", got)
	}
	if got := GetAverageLatency(metrics); got != 2*time.Second {
		t.Errorf("GetAverageLatency() = %v", got)
	}
}
