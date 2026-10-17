// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWeaveMCPUpdateCanDeclineReinstall(t *testing.T) {
	installDir := t.TempDir()
	binary := filepath.Join(installDir, "weave-mcp-stdio")
	if err := os.WriteFile(binary, []byte("fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEAVE_MCP_STDIO_PATH", binary)

	previousForce := weaveMCPForce
	weaveMCPForce = false
	t.Cleanup(func() { weaveMCPForce = previousForce })

	oldStdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString("n\n"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = reader.Close()
	})

	if err := runWeaveMCPUpdate(); err != nil {
		t.Fatalf("runWeaveMCPUpdate() after declining reinstall = %v", err)
	}
}
