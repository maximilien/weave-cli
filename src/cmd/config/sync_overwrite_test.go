// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncOverwriteConfirmation(t *testing.T) {
	for _, name := range []string{".env", "config.yaml"} {
		for _, answer := range []string{"no\n", "yes\n"} {
			t.Run(name+answer, func(t *testing.T) {
				root := t.TempDir()
				t.Setenv("HOME", root)
				t.Chdir(root)
				oldEnv, oldYAML := syncEnv, syncConfigYAML
				defer func() { syncEnv, syncConfigYAML = oldEnv, oldYAML }()
				syncEnv = name == ".env"
				syncConfigYAML = name == "config.yaml"
				global := filepath.Join(root, ".weave-cli")
				if err := os.MkdirAll(global, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte("new fixture data"), 0600); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(global, name)
				if err := os.WriteFile(target, []byte("original fixture data"), 0600); err != nil {
					t.Fatal(err)
				}
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
				oldInput := os.Stdin
				os.Stdin = input
				defer func() { os.Stdin = oldInput }()
				runConfigSync(syncCmd, nil)
				got, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				want := "original fixture data"
				if answer == "yes\n" {
					want = "new fixture data"
				}
				if string(got) != want {
					t.Fatalf("synced=%q want=%q", got, want)
				}
				source, err := os.ReadFile(name)
				if err != nil || string(source) != "new fixture data" {
					t.Fatalf("source=%q error=%v", source, err)
				}
			})
		}
	}
}
