// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteAgentConfirmation(t *testing.T) {
	for _, answer := range []string{"no\n", " YES \n"} {
		t.Run(answer, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			dir := filepath.Join(root, "custom")
			runCreateAgent("fixture", "rag", false, dir)
			input, err := os.CreateTemp(root, "stdin")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			if _, err = input.WriteString(answer); err != nil {
				t.Fatal(err)
			}
			if _, err = input.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			// Registry discovers project agent definitions under configs/agents.
			target := filepath.Join(root, "configs", "agents")
			if err = os.MkdirAll(target, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(target, "fixture.yaml")
			data, err := os.ReadFile(filepath.Join(dir, "fixture.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			previous := os.Stdin
			os.Stdin = input
			defer func() { os.Stdin = previous }()
			runDeleteAgent("fixture", false)
			_, err = os.Stat(path)
			if answer == "no\n" && err != nil {
				t.Fatal(err)
			}
			if answer != "no\n" && !os.IsNotExist(err) {
				t.Fatalf("agent still exists: %v", err)
			}
		})
	}
}
