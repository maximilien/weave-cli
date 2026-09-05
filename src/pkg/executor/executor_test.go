// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package executor

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewExecutorValidatesRequiredEnvironment(t *testing.T) {
	t.Setenv("OPIK_ENABLED", "false")
	t.Setenv("OPIK_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("WEAVE_MCP_STDIO_PATH", "")

	if _, err := NewExecutor(&Config{}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("missing OpenAI key error = %v", err)
	}

	t.Setenv("OPENAI_API_KEY", "test-key")
	if _, err := NewExecutor(&Config{}); err == nil || !strings.Contains(err.Error(), "WEAVE_MCP_STDIO_PATH") {
		t.Fatalf("missing MCP path error = %v", err)
	}
}

func TestNewExecutorContinuesAfterOptionalTracingError(t *testing.T) {
	t.Setenv("OPIK_ENABLED", "true")
	t.Setenv("OPIK_API_KEY", "test-opik-key")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "://invalid")
	t.Setenv("OPENAI_API_KEY", "")

	if _, err := NewExecutor(&Config{}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("NewExecutor() error = %v", err)
	}
}

func TestExecutorHelpers(t *testing.T) {
	executor := newTestExecutor()
	if executor.hasDestructiveOperations(&agents.ExecutionPlan{}) {
		t.Fatal("empty plan reported destructive operations")
	}
	if !executor.hasDestructiveOperations(&agents.ExecutionPlan{Steps: []agents.ExecutionStep{{Destructive: true}}}) {
		t.Fatal("destructive plan was not detected")
	}

	tests := []struct {
		name   string
		output interface{}
		want   interface{}
	}{
		{name: "JSON text", output: []interface{}{map[string]interface{}{"type": "text", "text": `{"count":2}`}}, want: map[string]interface{}{"count": float64(2)}},
		{name: "plain text", output: []interface{}{map[string]interface{}{"type": "text", "text": "plain"}}, want: "plain"},
		{name: "missing text", output: []interface{}{map[string]interface{}{"type": "text"}}, want: []interface{}{map[string]interface{}{"type": "text"}}},
		{name: "empty content", output: []interface{}{}, want: []interface{}{}},
		{name: "other shape", output: "original", want: "original"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.extractMCPTextContent(tt.output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("extractMCPTextContent() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExecuteBashStep(t *testing.T) {
	executor := newTestExecutor()
	tests := []struct {
		name      string
		step      agents.ExecutionStep
		want      string
		wantError string
	}{
		{name: "stdout", step: agents.ExecutionStep{Type: "bash", Command: "echo", Args: []string{"hello"}}, want: "hello"},
		{name: "command failure", step: agents.ExecutionStep{Type: "bash", Command: "grep", Args: []string{"missing", "/dev/null"}}, wantError: "exit code 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.executeBashStep(context.Background(), &tt.step)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("executeBashStep() error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil || !strings.Contains(result.(string), tt.want) {
				t.Fatalf("executeBashStep() = %#v, %v; want text %q", result, err, tt.want)
			}
		})
	}
}

func TestExecuteStep(t *testing.T) {
	executor := newTestExecutor()

	success := executor.executeStep(context.Background(), &agents.ExecutionStep{
		Type: "bash", Command: "echo", Args: []string{"done"},
	}, nil)
	if !success.Success || !strings.Contains(success.Output, "done") || success.Duration < 0 {
		t.Fatalf("successful report = %#v", success)
	}

	errorOutput := executor.executeStep(context.Background(), &agents.ExecutionStep{
		Type: "bash", Command: "echo", Args: []string{"error: reported failure"},
	}, nil)
	if errorOutput.Success || !strings.Contains(errorOutput.Error, "error: reported failure") {
		t.Fatalf("error-like output report = %#v", errorOutput)
	}

	unknown := executor.executeStep(context.Background(), &agents.ExecutionStep{
		Type: "unknown", Command: "noop",
	}, nil)
	if unknown.Success || !strings.Contains(unknown.Error, "unknown step type") {
		t.Fatalf("unknown-step report = %#v", unknown)
	}
}

func TestExecuteConfirmationStep(t *testing.T) {
	executor := newTestExecutor()
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = readEnd
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = readEnd.Close()
	})
	if _, err := writeEnd.WriteString("n\n"); err != nil {
		t.Fatalf("write confirmation: %v", err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close confirmation writer: %v", err)
	}

	report := executor.executeStep(context.Background(), &agents.ExecutionStep{
		Type: "confirm", Description: "continue?",
	}, nil)
	if report.Success || !strings.Contains(report.Error, "declined") {
		t.Fatalf("confirmation report = %#v", report)
	}
}

func TestExecutePlanDependencies(t *testing.T) {
	executor := newTestExecutor()
	plan := &agents.ExecutionPlan{Steps: []agents.ExecutionStep{
		{Type: "bash", Command: "echo", Args: []string{"first"}},
		{Type: "bash", Command: "grep", Args: []string{"missing", "/dev/null"}},
		{Type: "bash", Command: "echo", Args: []string{"optional"}, DependsOn: []int{1}, Optional: true},
		{Type: "bash", Command: "echo", Args: []string{"skipped"}, DependsOn: []int{1}},
		{Type: "bash", Command: "echo", Args: []string{"-n"}},
	}}

	reports, err := executor.executePlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("executePlan() error: %v", err)
	}
	if len(reports) != 5 {
		t.Fatalf("report count = %d, want 5", len(reports))
	}
	if !reports[0].Success || reports[1].Success || !reports[2].Success {
		t.Fatalf("unexpected execution reports: %#v", reports)
	}
	if reports[3].Error != "dependency failed" {
		t.Fatalf("dependent report = %#v", reports[3])
	}

	invalid := &agents.ExecutionPlan{Steps: []agents.ExecutionStep{{
		Type: "bash", Command: "echo", DependsOn: []int{0},
	}}}
	reports, err = executor.executePlan(context.Background(), invalid)
	if err == nil || len(reports) != 0 || !strings.Contains(err.Error(), "future step") {
		t.Fatalf("invalid dependency result = %#v, %v", reports, err)
	}
}

func TestCloseWithoutResources(t *testing.T) {
	if err := (&Executor{}).Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if err := (&Executor{mcpClient: &mcp.Client{}}).Close(); err != nil {
		t.Fatalf("Close() with MCP client error: %v", err)
	}
	provider := sdktrace.NewTracerProvider()
	if err := (&Executor{tracerProvider: provider}).Close(); err != nil {
		t.Fatalf("Close() with tracer provider error: %v", err)
	}
}

func newTestExecutor() *Executor {
	output := agents.NewOutputAgent(agents.OutputConfig{Quiet: true, NoColor: true})
	bash := agents.NewBashAgent()
	bash.SetOutputAgent(output)
	return &Executor{
		bashAgent:   bash,
		outputAgent: output,
		config:      &Config{Quiet: true},
	}
}
