// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBashAgentExecuteOutcomes(t *testing.T) {
	agent := NewBashAgent()
	if result, err := agent.Execute(context.Background(), "invalid"); err == nil || result != nil {
		t.Fatalf("Execute(invalid) = %#v, %v", result, err)
	}

	result, err := agent.Execute(context.Background(), &BashCommand{Command: "echo", Args: []string{"hello"}})
	bashResult, ok := result.(*BashResult)
	if err != nil || !ok || !bashResult.Success || strings.TrimSpace(bashResult.Stdout) != "hello" || bashResult.ExitCode != 0 {
		t.Fatalf("Execute(echo) = %#v, %v", result, err)
	}

	result, err = agent.Execute(context.Background(), &BashCommand{Command: "pwd", WorkingDir: t.TempDir()})
	bashResult = result.(*BashResult)
	if err != nil || !bashResult.Success || strings.TrimSpace(bashResult.Stdout) == "" {
		t.Fatalf("Execute(pwd) = %#v, %v", result, err)
	}

	agent.allowedCommands = append(agent.allowedCommands, "env")
	result, err = agent.Execute(context.Background(), &BashCommand{Command: "env", Environment: map[string]string{"WEAVE_DAY": "10"}})
	bashResult = result.(*BashResult)
	if err != nil || !strings.Contains(bashResult.Stdout, "WEAVE_DAY=10") {
		t.Fatalf("Execute(env) = %#v, %v", result, err)
	}

	result, err = agent.Execute(context.Background(), &BashCommand{Command: "which", Args: []string{"weave-command-that-does-not-exist"}})
	bashResult = result.(*BashResult)
	if err != nil || bashResult.Success || bashResult.ExitCode == 0 {
		t.Fatalf("Execute(which missing) = %#v, %v", result, err)
	}

	inputPath := filepath.Join(t.TempDir(), "confirmation.txt")
	if err := os.WriteFile(inputPath, []byte("yes\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(confirmation): %v", err)
	}
	input, err := os.Open(inputPath)
	if err != nil {
		t.Fatalf("Open(confirmation): %v", err)
	}
	defer input.Close()
	originalStdin := os.Stdin
	os.Stdin = input
	t.Cleanup(func() { os.Stdin = originalStdin })
	agent.SetOutputAgent(NewOutputAgent(OutputConfig{NoColor: true}))
	result, err = agent.Execute(context.Background(), &BashCommand{Command: "echo shell | tr a-z A-Z"})
	bashResult = result.(*BashResult)
	if err != nil || !bashResult.Success || strings.TrimSpace(bashResult.Stdout) != "SHELL" {
		t.Fatalf("Execute(shell) = %#v, %v", result, err)
	}

	tailFile := filepath.Join(t.TempDir(), "tail.txt")
	if err := os.WriteFile(tailFile, []byte("line\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(tail): %v", err)
	}
	result, err = agent.Execute(context.Background(), &BashCommand{Command: "tail", Args: []string{"-f", tailFile}, Timeout: 20 * time.Millisecond})
	bashResult = result.(*BashResult)
	if err != nil || bashResult.Success {
		t.Fatalf("Execute(timeout) = %#v, %v", result, err)
	}
}

func TestBashAgentCommandValidation(t *testing.T) {
	agent := NewBashAgent()
	if agent.Name() != "BashAgent" {
		t.Fatalf("Name() = %q", agent.Name())
	}
	if err := agent.validateCommand("ls"); err != nil {
		t.Fatalf("validateCommand(ls): %v", err)
	}
	if err := agent.validateCommand(`export PATH="/tmp/bin:$PATH" && weave cols ls`); err != nil {
		t.Fatalf("validateCommand(weave): %v", err)
	}
	for _, command := range []string{"rm records", "echo ok | wc", "custom-tool"} {
		if err := agent.validateCommand(command); err == nil {
			t.Errorf("validateCommand(%q) unexpectedly succeeded", command)
		}
	}

	for command, want := range map[string]bool{
		"weave cols ls": true,
		`export PATH="/tmp/bin" && weave docs ls docs`:      true,
		"weave cols ls | head":                              false,
		`export PATH="/tmp/bin" && weave cols ls && echo x`: false,
		"echo weave cols ls":                                false,
	} {
		if got := isSimpleWeaveCommand(command); got != want {
			t.Errorf("isSimpleWeaveCommand(%q) = %t, want %t", command, got, want)
		}
	}
	if !agent.isShellScript("for file in docs; do echo $file; done") || agent.isShellScript("ls") {
		t.Error("isShellScript() misclassified command")
	}
}

func TestBashAgentFileHelpers(t *testing.T) {
	files := ParseFileList(" .\nreport.PDF\n\nimage.PNG\n..\nnotes.md ")
	if strings.Join(files, ",") != "report.PDF,image.PNG,notes.md" {
		t.Fatalf("ParseFileList() = %#v", files)
	}
	for filename, want := range map[string]string{
		"photo.JPEG":  "image",
		"notes.yaml":  "text",
		"report.PDF":  "pdf",
		"archive.zip": "unknown",
	} {
		if got := ClassifyFileType(filename); got != want {
			t.Errorf("ClassifyFileType(%q) = %q, want %q", filename, got, want)
		}
	}
	if suggestion := NewBashAgent().SuggestInstallation("tesseract"); suggestion == "" {
		t.Fatal("SuggestInstallation(tesseract) returned empty")
	}
	if suggestion := NewBashAgent().SuggestInstallation("custom-tool"); !strings.Contains(suggestion, "custom-tool") {
		t.Fatalf("SuggestInstallation(custom-tool) = %q", suggestion)
	}
}

func TestProgressWriterBuffersCompleteLines(t *testing.T) {
	writer := newProgressWriter("  ")
	if n, err := writer.Write([]byte("first\n\npartial")); err != nil || n != len("first\n\npartial") {
		t.Fatalf("Write() = %d, %v", n, err)
	}
	if got := writer.GetLineCount(); got != 1 {
		t.Fatalf("GetLineCount() = %d", got)
	}
	writer.Flush()
	writer.Flush()
}
