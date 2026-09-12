// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"strings"
	"testing"
)

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
