// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestOutputAgentPresentationPaths(t *testing.T) {
	agent := NewOutputAgent(OutputConfig{NoColor: true, Verbose: true, OutputFormat: "text"})
	if agent.Name() != "OutputAgent" {
		t.Fatalf("Name() = %q", agent.Name())
	}
	input := map[string]string{"status": "ok"}
	got, err := agent.Execute(context.Background(), input)
	if err != nil || got == nil {
		t.Fatalf("Execute() = %#v, %v", got, err)
	}

	plan := &ExecutionPlan{
		Summary:  "inspect collections",
		Warnings: []string{"read-only plan"},
		Steps: []ExecutionStep{
			{Type: "bash", Description: "check tools"},
			{Type: "weave", Description: "list collections"},
			{Type: "confirm", Description: "remove collection", Destructive: true},
		},
	}
	plan.Estimations.Duration = "2s"
	plan.Estimations.Risk = "medium"
	agent.PrintPlan(plan)
	agent.PrintStepProgress(1, &plan.Steps[0], "running")
	agent.PrintStepCompletion(1, 1500*time.Millisecond)
	agent.PrintStepError(2, "boom")
	agent.PrintSuccess("done")
	agent.PrintError("failed")
	agent.PrintWarning("careful")
	agent.PrintInfo("details")
	agent.PrintCommandOutput(`[map[text:{"status":"ok","count":2} type:text]]`)
	agent.PrintCommandOutput("plain output\nsecond line")

	agent.PrintCommandResult(&CommandReport{Command: "weave cols ls", Success: true, Output: strings.Repeat("x", 120), Duration: time.Second})
	agent.PrintCommandResult(&CommandReport{Command: "weave cols del", Error: "denied"})
	satisfaction := 4.5
	agent.PrintMetrics(&EvaluationMetrics{
		QueryID: "query-1", Success: true, IntentMatched: true, LLMInvocations: 2,
		TotalTokens: 1500, PromptTokens: 1000, CompletionTokens: 500,
		TotalCost: 0.005, Latency: 250 * time.Millisecond, ErrorRate: 0.1,
		UserSatisfaction: &satisfaction,
	})
	agent.PrintRejectionMessage("unsupported request")
	if err := agent.CreateProgressBar(2, "testing").Add(1); err != nil {
		t.Fatalf("progress Add(): %v", err)
	}
}

func TestOutputAgentColorAndQuietPaths(t *testing.T) {
	colored := NewOutputAgent(OutputConfig{Verbose: true})
	for _, risk := range []string{"low", "medium", "high", "unknown"} {
		if colored.colorizeRisk(risk) == "" {
			t.Fatalf("colorizeRisk(%q) returned empty", risk)
		}
	}
	colored.PrintStepCompletion(1, 500*time.Millisecond)
	colored.PrintStepError(1, "error")
	colored.PrintSuccess("success")
	colored.PrintError("error")
	colored.PrintWarning("warning")
	colored.PrintCommandOutput("{\n\"string\": \"value\",\n\"number\": 2,\n\"bool\": true,\n\"nothing\": null,\n\"array\": [\n\"item\"\n]\n}")
	for _, cost := range []float64{0.005, 0.05, 0.5} {
		colored.PrintMetrics(&EvaluationMetrics{LLMInvocations: 1, TotalTokens: 10, PromptTokens: 6, CompletionTokens: 4, TotalCost: cost})
	}

	quiet := NewOutputAgent(OutputConfig{Quiet: true, NoColor: true})
	quiet.PrintPlan(&ExecutionPlan{})
	quiet.PrintStepProgress(1, &ExecutionStep{}, "running")
	quiet.PrintStepCompletion(1, time.Second)
	quiet.PrintStepError(1, "error")
	quiet.PrintInfo("hidden")
	quiet.PrintCommandOutput("hidden")
	quiet.PrintCommandResult(&CommandReport{Success: true})
	quiet.PrintMetrics(&EvaluationMetrics{})
	if err := quiet.CreateProgressBar(1, "silent").Add(1); err != nil {
		t.Fatalf("silent progress Add(): %v", err)
	}
}

func TestOutputFormattingHelpers(t *testing.T) {
	jsonCases := map[string]bool{
		`{"ok":true}`: true,
		`[1,2]`:       true,
		`plain`:       false,
		`{unfinished`: false,
	}
	for input, want := range jsonCases {
		if got := isJSON(input); got != want {
			t.Errorf("isJSON(%q) = %t, want %t", input, got, want)
		}
	}

	agent := NewOutputAgent(OutputConfig{NoColor: true})
	if got := agent.extractMCPContent(`[map[text:{"nested":{"ok":true}} type:text]]`); got != `{"nested":{"ok":true}}` {
		t.Errorf("extractMCPContent() = %q", got)
	}
	for _, input := range []string{"plain", "text:{unfinished"} {
		if got := agent.extractMCPContent(input); got != input {
			t.Errorf("extractMCPContent(%q) = %q", input, got)
		}
	}
	if got := colorizeJSON("plain\n\"array item\""); got == "" {
		t.Fatal("colorizeJSON() returned empty")
	}

	for input, want := range map[int]string{12: "12", 1234: "1,234", 1234567: "1,234,567"} {
		if got := formatNumber(input); got != want {
			t.Errorf("formatNumber(%d) = %q, want %q", input, got, want)
		}
	}
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate(short) = %q", got)
	}
	if got := truncate("long value", 4); got != "long..." {
		t.Errorf("truncate(long) = %q", got)
	}
	for duration, want := range map[time.Duration]string{
		500 * time.Millisecond:  "500ms",
		1500 * time.Millisecond: "1.5s",
		90 * time.Second:        "1.5m",
	} {
		if got := FormatDuration(duration); got != want {
			t.Errorf("FormatDuration(%v) = %q, want %q", duration, got, want)
		}
	}
}
