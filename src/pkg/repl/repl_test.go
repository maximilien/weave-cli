// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package repl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/spf13/viper"
)

type fakeQueryExecutor struct {
	queries    []string
	executeErr error
	closed     bool
}

func (e *fakeQueryExecutor) Execute(_ context.Context, query string) (*agents.OperationReport, error) {
	e.queries = append(e.queries, query)
	return &agents.OperationReport{QueryIntent: query}, e.executeErr
}

func (e *fakeQueryExecutor) Close() error {
	e.closed = true
	return nil
}

func TestLoadQueriesFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queries.txt")
	contents := "# setup\n\n list collections \nsearch for docs\n   # ignored\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write query file: %v", err)
	}

	queries, err := loadQueriesFromFile(path)
	if err != nil {
		t.Fatalf("loadQueriesFromFile() error: %v", err)
	}
	want := []string{"list collections", "search for docs"}
	if !reflect.DeepEqual(queries, want) {
		t.Fatalf("loadQueriesFromFile() = %#v, want %#v", queries, want)
	}
}

func TestLoadQueriesFromFileErrors(t *testing.T) {
	if _, err := loadQueriesFromFile(filepath.Join(t.TempDir(), "missing")); err == nil || !strings.Contains(err.Error(), "failed to open") {
		t.Fatalf("missing file error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "oversized.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 70*1024)), 0o600); err != nil {
		t.Fatalf("write oversized query file: %v", err)
	}
	if _, err := loadQueriesFromFile(path); err == nil || !strings.Contains(err.Error(), "failed to read") {
		t.Fatalf("scanner error = %v", err)
	}
}

func TestNewReportsExecutorConfigurationErrors(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("OPIK_ENABLED", "false")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("WEAVE_MCP_STDIO_PATH", "")

	if _, err := New(); err == nil || !strings.Contains(err.Error(), "failed to create executor") {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := NewWithOptions(Options{NoConfirm: true}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
}

func TestRunBatchMode(t *testing.T) {
	executor := &fakeQueryExecutor{}
	repl := &REPL{
		executor:  executor,
		batchMode: true,
		queries:   []string{"first query", "second query"},
	}

	if err := repl.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !reflect.DeepEqual(executor.queries, repl.queries) {
		t.Fatalf("executed queries = %#v, want %#v", executor.queries, repl.queries)
	}
	if !executor.closed {
		t.Fatal("Run() did not close executor")
	}
}

func TestExecuteQueryHandlesErrorsAndInterruptions(t *testing.T) {
	executor := &fakeQueryExecutor{executeErr: errors.New("execution failed")}
	repl := &REPL{executor: executor}
	repl.executeQuery("failing query")
	repl.interrupted = true
	repl.executeQuery("interrupted query")

	if !reflect.DeepEqual(executor.queries, []string{"failing query", "interrupted query"}) {
		t.Fatalf("executed queries = %#v", executor.queries)
	}
}

func TestHandleInformationalCommands(t *testing.T) {
	t.Setenv("HOME", "/tmp/repl-home")
	repl := &REPL{}
	tests := []struct {
		line    string
		handled bool
	}{
		{line: "", handled: false},
		{line: "plain query", handled: false},
		{line: "/help", handled: true},
		{line: "HELP", handled: true},
		{line: "/history", handled: true},
		{line: "/examples", handled: true},
		{line: "/status", handled: true},
		{line: "/clear", handled: true},
	}

	for _, tt := range tests {
		if got := repl.handleSpecialCommand(tt.line); got != tt.handled {
			t.Errorf("handleSpecialCommand(%q) = %t, want %t", tt.line, got, tt.handled)
		}
	}
}

func TestGetModel(t *testing.T) {
	t.Setenv("OPENAI_MODEL", "")
	if got := getModel(); got != "gpt-4o" {
		t.Fatalf("getModel() = %q", got)
	}
	t.Setenv("OPENAI_MODEL", "gpt-test")
	if got := getModel(); got != "gpt-test" {
		t.Fatalf("getModel() = %q", got)
	}
}
