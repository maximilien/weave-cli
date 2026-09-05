// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package query

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNewQueryCommand(t *testing.T) {
	resetQueryGlobals(t)
	cmd := NewQueryCommand()

	if cmd.Use != "query [query-text]" || len(cmd.Aliases) != 1 || cmd.Aliases[0] != "q" {
		t.Fatalf("unexpected command metadata: use=%q aliases=%#v", cmd.Use, cmd.Aliases)
	}
	for _, name := range []string{"dry-run", "no-confirm", "output", "model", "max-retries"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag %q is not registered", name)
		}
	}
	if err := cmd.Args(cmd, nil); err == nil {
		t.Fatal("query command accepted no arguments")
	}
	if err := cmd.Args(cmd, []string{"one", "two"}); err == nil {
		t.Fatal("query command accepted multiple arguments")
	}
	if err := cmd.Args(cmd, []string{"one query"}); err != nil {
		t.Fatalf("query command rejected one argument: %v", err)
	}
}

func TestGetModelPriority(t *testing.T) {
	resetQueryGlobals(t)
	t.Setenv("OPENAI_MODEL", "environment-model")
	if got := getModel(); got != "environment-model" {
		t.Fatalf("environment model = %q", got)
	}

	model = "flag-model"
	if got := getModel(); got != "flag-model" {
		t.Fatalf("flag model = %q", got)
	}

	model = ""
	t.Setenv("OPENAI_MODEL", "")
	if got := getModel(); got != "gpt-4o" {
		t.Fatalf("default model = %q", got)
	}
}

func TestRunQueryReportsExecutorConfigurationError(t *testing.T) {
	resetQueryGlobals(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPIK_ENABLED", "false")
	t.Setenv("OPIK_API_KEY", "")
	t.Setenv("WEAVE_MCP_STDIO_PATH", "")
	viper.Set("env", "")
	viper.Set("verbose", false)
	viper.Set("quiet", true)
	viper.Set("no-color", true)
	t.Cleanup(viper.Reset)

	err := runQuery(NewQueryCommand(), []string{"list collections"})
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("runQuery() error = %v", err)
	}
}

func resetQueryGlobals(t *testing.T) {
	t.Helper()
	oldDryRun := dryRun
	oldNoConfirm := noConfirm
	oldOutputFormat := outputFormat
	oldModel := model
	oldMaxRetries := maxRetries
	dryRun = false
	noConfirm = false
	outputFormat = "text"
	model = ""
	maxRetries = 3
	t.Cleanup(func() {
		dryRun = oldDryRun
		noConfirm = oldNoConfirm
		outputFormat = oldOutputFormat
		model = oldModel
		maxRetries = oldMaxRetries
	})
}
