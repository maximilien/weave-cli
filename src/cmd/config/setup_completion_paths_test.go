package config

import "testing"

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
