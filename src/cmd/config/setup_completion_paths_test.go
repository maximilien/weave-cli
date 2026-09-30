package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellSetupPathsDay27(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell", "pwsh"} {
		setup := getShellSetup(shell)
		if setup == nil || setup.ConfigFile == "" || setup.CompletionLine == "" {
			t.Fatalf("missing setup for %s: %#v", shell, setup)
		}
	}
	if getShellSetup("unknown") != nil {
		t.Fatal("unknown shell should have no setup")
	}
	t.Setenv("SHELL", "/bin/zsh")
	if detectShell() != "zsh" {
		t.Fatalf("detectShell = %q", detectShell())
	}
	t.Setenv("SHELL", "")
	_ = detectShell()
}

func TestPerformSetupFileBranches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shellrc")
	oldUpdate := updateCompletion
	t.Cleanup(func() { updateCompletion = oldUpdate })
	updateCompletion = false
	setup := &ShellSetup{ConfigFile: path, CompletionLine: "source <(weave completion bash)"}
	if err := performSetup(setup); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "weave completion bash") {
		t.Fatalf("setup file: %q err=%v", data, err)
	}
	if err := performSetup(setup); err != nil {
		t.Fatal(err)
	}
	updateCompletion = true
	if err := performSetup(setup); err != nil {
		t.Fatal(err)
	}
}
