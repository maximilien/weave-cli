// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package executor

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type fakeAgent struct {
	name    string
	execute func(context.Context, interface{}) (interface{}, error)
}

func (a *fakeAgent) Name() string {
	if a.name == "" {
		return "fake-agent"
	}
	return a.name
}

func (a *fakeAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return a.execute(ctx, input)
}

type fakeReportAgent struct {
	fakeAgent
	printed []*agents.OperationReport
}

func (a *fakeReportAgent) PrintReport(report *agents.OperationReport) {
	a.printed = append(a.printed, report)
}

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

func TestDryRun(t *testing.T) {
	plan := &agents.ExecutionPlan{Summary: "inspect collections"}
	executor := newAgentTestExecutor(
		&fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{
			IsWeaveQuery: true,
			FixedQuery:   "list collections",
			Intent:       "list",
		})},
		&fakeAgent{name: "planner", execute: returning(plan)},
	)

	got, err := executor.DryRun(context.Background(), "show collections")
	if err != nil {
		t.Fatalf("DryRun() error: %v", err)
	}
	if got != plan {
		t.Fatalf("DryRun() = %#v, want %#v", got, plan)
	}
}

func TestDryRunErrors(t *testing.T) {
	tests := []struct {
		name      string
		query     agents.Agent
		planner   agents.Agent
		wantError string
	}{
		{
			name:      "query failure",
			query:     &fakeAgent{execute: failing("query unavailable")},
			planner:   &fakeAgent{execute: returning(nil)},
			wantError: "failed to analyze query",
		},
		{
			name: "unrelated query",
			query: &fakeAgent{execute: returning(&agents.QueryAgentOutput{
				Reason: "unrelated",
			})},
			planner:   &fakeAgent{execute: returning(nil)},
			wantError: "not weave-related",
		},
		{
			name: "planning failure",
			query: &fakeAgent{execute: returning(&agents.QueryAgentOutput{
				IsWeaveQuery: true,
				FixedQuery:   "list",
			})},
			planner:   &fakeAgent{execute: failing("planner unavailable")},
			wantError: "failed to create execution plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newAgentTestExecutor(tt.query, tt.planner)
			if _, err := executor.DryRun(context.Background(), "query"); err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("DryRun() error = %v, want %q", err, tt.wantError)
			}
		})
	}
}

func TestExecuteSuccess(t *testing.T) {
	t.Setenv("OPIK_ENABLED", "false")
	reporter := &fakeReportAgent{fakeAgent: fakeAgent{name: "report", execute: returning(nil)}}
	executor := newAgentTestExecutor(
		&fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{
			IsWeaveQuery: true,
			FixedQuery:   "echo complete",
			Intent:       "query",
		})},
		&fakeAgent{name: "planner", execute: returning(&agents.ExecutionPlan{
			Steps: []agents.ExecutionStep{{Type: "bash", Command: "echo", Args: []string{"complete"}}},
		})},
	)
	executor.reportAgent = reporter
	executor.evalAgent = &fakeAgent{name: "eval", execute: returning(&agents.EvaluationMetrics{Success: true})}

	report, err := executor.Execute(context.Background(), "do the thing")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if report.ExecutedSteps != 1 || report.SuccessfulSteps != 1 || report.QueryIntent != "query" {
		t.Fatalf("Execute() report = %#v", report)
	}
	if len(reporter.printed) != 1 || reporter.printed[0] != report {
		t.Fatalf("printed reports = %#v", reporter.printed)
	}
}

func TestExecuteBoundaryErrors(t *testing.T) {
	t.Setenv("OPIK_ENABLED", "false")
	tests := []struct {
		name      string
		query     agents.Agent
		planner   agents.Agent
		wantError string
	}{
		{
			name:      "query failure",
			query:     &fakeAgent{name: "query", execute: failing("query unavailable")},
			planner:   &fakeAgent{name: "planner", execute: returning(nil)},
			wantError: "failed to analyze query",
		},
		{
			name: "unrelated query",
			query: &fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{
				Reason: "not a database request",
			})},
			planner:   &fakeAgent{name: "planner", execute: returning(nil)},
			wantError: "not weave-related",
		},
		{
			name: "planning failure",
			query: &fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{
				IsWeaveQuery: true,
			})},
			planner:   &fakeAgent{name: "planner", execute: failing("planner unavailable")},
			wantError: "failed to create execution plan",
		},
		{
			name: "invalid plan dependency",
			query: &fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{
				IsWeaveQuery: true,
			})},
			planner: &fakeAgent{name: "planner", execute: returning(&agents.ExecutionPlan{
				Steps: []agents.ExecutionStep{{Type: "bash", Command: "echo", DependsOn: []int{0}}},
			})},
			wantError: "failed to execute plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newAgentTestExecutor(tt.query, tt.planner)
			if _, err := executor.Execute(context.Background(), "query"); err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("Execute() error = %v, want %q", err, tt.wantError)
			}
		})
	}
}

func TestExecuteDryRunAndCancellation(t *testing.T) {
	t.Setenv("OPIK_ENABLED", "false")
	newExecutor := func() *Executor {
		executor := newAgentTestExecutor(
			&fakeAgent{name: "query", execute: returning(&agents.QueryAgentOutput{IsWeaveQuery: true})},
			&fakeAgent{name: "planner", execute: returning(&agents.ExecutionPlan{
				Steps: []agents.ExecutionStep{{Type: "bash", Command: "echo", Destructive: true}},
			})},
		)
		executor.config.NoConfirm = false
		return executor
	}

	dryRun := newExecutor()
	dryRun.config.DryRun = true
	if report, err := dryRun.Execute(context.Background(), "query"); err != nil || report != nil {
		t.Fatalf("dry-run Execute() = %#v, %v", report, err)
	}

	cancelled := newExecutor()
	withStdin(t, "n\n", func() {
		if _, err := cancelled.Execute(context.Background(), "query"); err == nil || !strings.Contains(err.Error(), "cancelled") {
			t.Fatalf("cancelled Execute() error = %v", err)
		}
	})
}

func TestExecuteWeaveStep(t *testing.T) {
	executor := newTestExecutor()
	executor.weaveAgent = &fakeAgent{execute: returning(&agents.WeaveAgentResult{
		Success: true,
		Output:  []interface{}{map[string]interface{}{"text": `{"count": 3}`}},
	})}
	result, err := executor.executeWeaveStep(context.Background(), &agents.ExecutionStep{Command: "list"})
	if err != nil {
		t.Fatalf("executeWeaveStep() error: %v", err)
	}
	if !reflect.DeepEqual(result, map[string]interface{}{"count": float64(3)}) {
		t.Fatalf("executeWeaveStep() = %#v", result)
	}

	executor.weaveAgent = &fakeAgent{execute: returning(&agents.WeaveAgentResult{Success: false, Error: "tool failed"})}
	if _, err := executor.executeWeaveStep(context.Background(), &agents.ExecutionStep{}); err == nil || !strings.Contains(err.Error(), "tool failed") {
		t.Fatalf("executeWeaveStep() error = %v", err)
	}

	executor.weaveAgent = &fakeAgent{execute: failing("transport failed")}
	if _, err := executor.executeWeaveStep(context.Background(), &agents.ExecutionStep{}); err == nil || !strings.Contains(err.Error(), "transport failed") {
		t.Fatalf("executeWeaveStep() error = %v", err)
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

func newAgentTestExecutor(queryAgent, planningAgent agents.Agent) *Executor {
	executor := newTestExecutor()
	executor.queryAgent = queryAgent
	executor.planningAgent = planningAgent
	executor.reportAgent = &fakeReportAgent{fakeAgent: fakeAgent{name: "report", execute: returning(nil)}}
	executor.evalAgent = &fakeAgent{name: "eval", execute: returning(&agents.EvaluationMetrics{})}
	executor.config.NoConfirm = true
	return executor
}

func returning(value interface{}) func(context.Context, interface{}) (interface{}, error) {
	return func(context.Context, interface{}) (interface{}, error) {
		return value, nil
	}
}

func failing(message string) func(context.Context, interface{}) (interface{}, error) {
	return func(context.Context, interface{}) (interface{}, error) {
		return nil, errors.New(message)
	}
}

func withStdin(t *testing.T, input string, run func()) {
	t.Helper()
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
	if _, err := writeEnd.WriteString(input); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	run()
}
