// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixApplierAppliesAndBacksUpChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	initial := `databases:
  default: first
  vector_databases:
    - name: first
      type: weaviate-cloud
      url: old
    - name: second
      type: mongodb-cloud
    - name: third
      type: qdrant-cloud
`
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	applier, err := NewFixApplier(path)
	if err != nil {
		t.Fatal(err)
	}
	results := []FixResult{
		{Issue: ConfigIssue{Path: "databases.vector_databases[0].url", DatabaseName: "first", DatabaseIdx: 0, Field: "url"}, Action: FixActionSetValue, Value: "https://new.example"},
		{Issue: ConfigIssue{DatabaseName: "second", DatabaseIdx: 1}, Action: FixActionDisable},
		{Issue: ConfigIssue{DatabaseName: "third", DatabaseIdx: 2}, Action: FixActionRemove},
		{Issue: ConfigIssue{DatabaseName: "first", DatabaseIdx: 0}, Action: FixActionSkip},
	}
	if err := applier.ApplyFixes(results, true); err != nil {
		t.Fatalf("ApplyFixes() error: %v", err)
	}
	if applier.GetBackupPath() == "" {
		t.Fatal("backup path was not recorded")
	}
	backup, err := os.ReadFile(applier.GetBackupPath())
	if err != nil || string(backup) != initial {
		t.Fatalf("backup = %q, %v", backup, err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, expected := range []string{"https://new.example", "enabled: false"} {
		if !strings.Contains(text, expected) {
			t.Errorf("updated config missing %q: %s", expected, text)
		}
	}
	if strings.Contains(text, "name: third") {
		t.Errorf("removed database remains: %s", text)
	}

	DryRun(results)
}

func TestFixApplierRejectsInvalidOperations(t *testing.T) {
	if _, err := NewFixApplier(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("NewFixApplier() accepted a missing file")
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("databases: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	applier, err := NewFixApplier(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []FixResult{
		{Issue: ConfigIssue{DatabaseName: "global", DatabaseIdx: -1}, Action: FixActionRemove},
		{Issue: ConfigIssue{DatabaseName: "global", DatabaseIdx: -1}, Action: FixActionDisable},
	} {
		if err := applier.ApplyFixes([]FixResult{result}, false); err == nil {
			t.Fatalf("ApplyFixes(%s) succeeded", result.Action)
		}
	}
}
