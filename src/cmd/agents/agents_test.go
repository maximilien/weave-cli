// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgagents "github.com/maximilien/weave-cli/src/pkg/agents"
	"gopkg.in/yaml.v3"
)

func TestAgentCommandContracts(t *testing.T) {
	root := NewAgentsCommand()
	if root.Use != "agents" || len(root.Commands()) != 7 {
		t.Fatalf("NewAgentsCommand() = %q with %d subcommands", root.Use, len(root.Commands()))
	}

	tests := []struct {
		name string
		cmd  func() interface{ ValidateArgs([]string) error }
		args []string
	}{
		{name: "show", cmd: func() interface{ ValidateArgs([]string) error } { return NewShowCommand() }, args: []string{"one"}},
		{name: "validate", cmd: func() interface{ ValidateArgs([]string) error } { return NewValidateCommand() }, args: []string{"one"}},
		{name: "create", cmd: func() interface{ ValidateArgs([]string) error } { return NewCreateCommand() }, args: []string{"one"}},
		{name: "delete", cmd: func() interface{ ValidateArgs([]string) error } { return NewDeleteCommand() }, args: []string{"one"}},
		{name: "edit", cmd: func() interface{ ValidateArgs([]string) error } { return NewEditCommand() }, args: []string{"one"}},
		{name: "copy", cmd: func() interface{ ValidateArgs([]string) error } { return NewCopyCommand() }, args: []string{"one", "two"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := test.cmd()
			if err := cmd.ValidateArgs(test.args); err != nil {
				t.Fatalf("valid arguments rejected: %v", err)
			}
			if err := cmd.ValidateArgs(nil); err == nil {
				t.Fatal("missing arguments accepted")
			}
		})
	}

	if cmd := NewListCommand(); cmd.Use != "list" {
		t.Fatalf("NewListCommand().Use = %q", cmd.Use)
	}
}

func TestAgentDisplaysAndFileOperations(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	agentDir := filepath.Join(root, "configs", "agents")

	runCreateAgent("test-agent", "rag", false, agentDir)
	path := filepath.Join(agentDir, "test-agent.yaml")
	config, err := pkgagents.LoadCustomAgentConfig(path)
	if err != nil {
		t.Fatalf("load created agent: %v", err)
	}
	config.Author = "test author"
	config.Description = strings.Repeat("long description ", 6)
	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal created agent: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("update created agent: %v", err)
	}

	for _, format := range []string{"text", "json", "yaml"} {
		runListAgents(format)
		runShowAgent("test-agent", format)
	}
	runValidateAgent(path)

	copyDir := filepath.Join(root, "copies")
	runCopyAgent("test-agent", "copied-agent", copyDir)
	if _, err := os.Stat(filepath.Join(copyDir, "copied-agent.yaml")); err != nil {
		t.Fatalf("copied agent missing: %v", err)
	}

	t.Setenv("EDITOR", "true")
	runEditAgent("test-agent")
	runDeleteAgent("test-agent", true)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deleted agent still exists: %v", err)
	}
}

func TestPromptForConfig(t *testing.T) {
	config := pkgagents.GetRAGTemplate()
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
	if _, err := writeEnd.WriteString("Updated description\ngpt-test\n0.25\n2048\n"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close stdin: %v", err)
	}

	got := promptForConfig(config)
	if got.Description != "Updated description" || got.LLM.Model != "gpt-test" || got.LLM.Temperature != 0.25 || got.LLM.MaxTokens != 2048 {
		t.Fatalf("promptForConfig() = %+v", got)
	}
}
