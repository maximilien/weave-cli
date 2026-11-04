// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeUserAgentsCopiesAndPreservesConfigurations(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", filepath.Join(root, "home"))
	source := filepath.Join(root, "installation", "configs", "agents")
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"rag-agent.yaml", "qa-agent.yaml", "summarize-agent.yaml"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte("built-in "+name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	previousArg := os.Args[0]
	os.Args[0] = filepath.Join(root, "installation", "weave")
	t.Cleanup(func() { os.Args[0] = previousArg })
	dest := filepath.Join(root, "home", ".weave-cli", "agents")
	if err := os.MkdirAll(dest, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "qa-agent.yaml"), []byte("custom"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := InitializeUserAgents(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"rag-agent.yaml": "built-in rag-agent.yaml", "qa-agent.yaml": "custom", "summarize-agent.yaml": "built-in summarize-agent.yaml", ".initialized": ""} {
		got, err := os.ReadFile(filepath.Join(dest, name))
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v", name, got, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dest, "rag-agent.yaml"), []byte("edited"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := InitializeUserAgents(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(dest, "rag-agent.yaml")); err != nil || string(got) != "edited" {
		t.Fatalf("reinitialization: %q, %v", got, err)
	}
}

func TestInitializeUserAgentsReportsMissingSourceAgent(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", filepath.Join(root, "home"))
	source := filepath.Join(root, "installation", "configs", "agents")
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "rag-agent.yaml"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	previousArg := os.Args[0]
	os.Args[0] = filepath.Join(root, "installation", "weave")
	t.Cleanup(func() { os.Args[0] = previousArg })
	if err := InitializeUserAgents(); err == nil {
		t.Fatal("expected missing qa agent error")
	}
	if _, err := os.Stat(filepath.Join(root, "home", ".weave-cli", "agents", ".initialized")); !os.IsNotExist(err) {
		t.Fatalf("marker created after failure: %v", err)
	}
}
