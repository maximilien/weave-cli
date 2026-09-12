// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

type reportLLM struct {
	response string
	err      error
	prompt   string
	options  *llm.CompletionOptions
}

func (m *reportLLM) Complete(_ context.Context, prompt string, opts ...llm.Option) (string, error) {
	m.prompt = prompt
	m.options = llm.DefaultCompletionOptions()
	for _, opt := range opts {
		opt(m.options)
	}
	return m.response, m.err
}

func (m *reportLLM) CompleteStructured(context.Context, string, interface{}, ...llm.Option) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *reportLLM) GetMetrics() *llm.Metrics { return &llm.Metrics{} }

func TestReportAgentEnhancesSuccessfulReport(t *testing.T) {
	model := &reportLLM{response: "• Inspect the collection details\n- Count documents before cleanup\n* Back up important records first\n- This fourth recommendation is deliberately ignored"}
	output := NewOutputAgent(OutputConfig{NoColor: true, Verbose: true})
	agent := NewReportAgent(output, model)
	if agent.Name() != "ReportAgent" {
		t.Fatalf("Name() = %q", agent.Name())
	}
	report := CreateReport("create and list collections with count", time.Now().Add(-2*time.Second), []CommandReport{
		{Command: "weave cols create docs", Success: true, Output: strings.Repeat("created\n", 40), Duration: time.Second},
		{Command: "weave cols ls", Success: true, Duration: 500 * time.Millisecond},
	})
	result, err := agent.Execute(context.Background(), report)
	if err != nil || result != report {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if report.Summary != "Successfully completed all 2 operations." {
		t.Errorf("summary = %q", report.Summary)
	}
	if len(report.Recommendations) != 3 {
		t.Errorf("recommendations = %#v", report.Recommendations)
	}
	if len(report.NextSteps) != 3 {
		t.Errorf("next steps = %#v", report.NextSteps)
	}
	if !strings.Contains(model.prompt, "weave cols create docs") || model.options.Temperature != 0.3 || model.options.MaxTokens != 500 {
		t.Errorf("LLM request = prompt %q, options %#v", model.prompt, model.options)
	}
	agent.PrintReport(report)

	preserved := &OperationReport{Summary: "existing", QueryIntent: "noop"}
	agent.enhanceReport(context.Background(), preserved)
	if preserved.Summary != "existing" {
		t.Errorf("enhanceReport replaced summary: %q", preserved.Summary)
	}
}

func TestReportAgentFailureFallbacks(t *testing.T) {
	model := &reportLLM{err: errors.New("offline")}
	agent := NewReportAgent(NewOutputAgent(OutputConfig{NoColor: true}), model)
	report := CreateReport("list empty collections and count documents", time.Now(), []CommandReport{
		{Command: "weave cols ls", Success: true},
		{Command: "weave docs count docs", Error: "connection refused"},
	})
	result, err := agent.Execute(context.Background(), report)
	if err != nil || result != report {
		t.Fatalf("Execute() = %#v, %v", result, err)
	}
	if report.Summary != "Completed 1 of 2 operations successfully. 1 operations failed." {
		t.Errorf("summary = %q", report.Summary)
	}
	if len(report.Recommendations) != 1 || !strings.Contains(report.Recommendations[0], "error") {
		t.Errorf("fallback recommendations = %#v", report.Recommendations)
	}
	if len(report.NextSteps) != 3 {
		t.Errorf("failure next steps = %#v", report.NextSteps)
	}
	agent.PrintReport(report)

	allSuccessful := &OperationReport{QueryIntent: "list and count", ExecutedSteps: 1, SuccessfulSteps: 1}
	recommendations := agent.generateFallbackRecommendations(allSuccessful)
	if len(recommendations) != 2 {
		t.Errorf("successful fallback recommendations = %#v", recommendations)
	}
}

func TestReportAgentRejectsInvalidInput(t *testing.T) {
	agent := NewReportAgent(NewOutputAgent(OutputConfig{NoColor: true}), &reportLLM{})
	if result, err := agent.Execute(context.Background(), "not a report"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}
}

func TestReportFormattingHelpers(t *testing.T) {
	short := "line one\nline two"
	if got := truncateMultiline(short, 100); got != short {
		t.Errorf("truncateMultiline(short) = %q", got)
	}
	longWithBreak := strings.Repeat("a", 60) + "\n" + strings.Repeat("b", 60)
	if got := truncateMultiline(longWithBreak, 100); got != strings.Repeat("a", 60)+"..." {
		t.Errorf("truncateMultiline(newline) = %q", got)
	}
	longWithoutBreak := strings.Repeat("x", 120)
	if got := truncateMultiline(longWithoutBreak, 100); got != strings.Repeat("x", 100)+"..." {
		t.Errorf("truncateMultiline(no newline) = %q", got)
	}
}
