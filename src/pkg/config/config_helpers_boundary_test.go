// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withConfigInput(t *testing.T, input string, run func()) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	oldStdin := os.Stdin
	os.Stdin = reader
	run()
	os.Stdin = oldStdin
	_ = reader.Close()
}

func TestSimpleEnvironmentFileHelpers(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	values := map[string]string{
		"WEAVIATE_URL":     "https://fixture.test",
		"WEAVIATE_API_KEY": "weaviate-secret",
		"OPENAI_API_KEY":   "openai-secret",
	}

	withConfigInput(t, "  piped-secret  \n", func() {
		value, err := readSecretSimple()
		if err != nil || value != "piped-secret" {
			t.Fatalf("readSecretSimple() = (%q, %v)", value, err)
		}
	})

	if got := loadEnvFileSimple("missing.env"); len(got) != 0 {
		t.Fatalf("missing env values = %#v", got)
	}
	if err := createSimpleEnvFile("simple.env", values); err != nil {
		t.Fatal(err)
	}
	loaded := loadEnvFileSimple("simple.env")
	if loaded["WEAVIATE_URL"] != values["WEAVIATE_URL"] || loaded["OPENAI_API_KEY"] != values["OPENAI_API_KEY"] {
		t.Fatalf("loaded simple env = %#v", loaded)
	}

	example := "# Header\n\nWEAVIATE_URL=old\nWEAVIATE_API_KEY=keep\nMALFORMED\n"
	if err := os.WriteFile(".env.example", []byte(example), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := saveEnvFileSimple("updated.env", map[string]string{"WEAVIATE_URL": "https://new.test"}); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile("updated.env")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`WEAVIATE_URL="https://new.test"`, "WEAVIATE_API_KEY=keep", "MALFORMED"} {
		if !strings.Contains(string(updated), want) {
			t.Errorf("updated env missing %q:\n%s", want, updated)
		}
	}

	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := saveEnvFileSimple(filepath.Join(blocker, "child.env"), values); err == nil {
		t.Fatal("expected env write error")
	}
}

func TestMinimalConfigAndEnvPathHelpers(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	if content := getMinimalConfigYAMLContent(); !strings.Contains(content, "${WEAVIATE_URL}") {
		t.Fatalf("unexpected minimal config: %s", content)
	}
	if err := createMinimalConfigYAML(); err != nil {
		t.Fatal(err)
	}
	if err := createMinimalConfigYAML(); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected existing config error, got %v", err)
	}

	previousLoaded := loadedEnvFile
	loadedEnvFile = "selected.env"
	t.Cleanup(func() { loadedEnvFile = previousLoaded })
	if got := GetEnvFile(); got != filepath.Join(root, "selected.env") {
		t.Fatalf("GetEnvFile(loaded) = %q", got)
	}
	loadedEnvFile = ""
	t.Setenv("ENV_FILE", "/tmp/explicit.env")
	if got := GetEnvFile(); got != "/tmp/explicit.env" {
		t.Fatalf("GetEnvFile(explicit) = %q", got)
	}

	if proceed, err := PromptToFixConfig(&ConfigError{Message: "fixture"}); err != nil || proceed {
		t.Fatalf("PromptToFixConfig(non-terminal) = (%v, %v)", proceed, err)
	}
	vars := getRequiredEnvVars()
	if len(vars) != 3 || maskSecret("short") != "***" || maskSecret("123456789") != "1234...6789" {
		t.Fatalf("unexpected helper values: %#v", vars)
	}
}
