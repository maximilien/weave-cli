// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
)

func captureBackupStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	})

	runErr := run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data), runErr
}

func writeCommandBackup(t *testing.T, path, collection string, createdAt time.Time, compressed bool) {
	t.Helper()
	backup := backuppkg.NewBackupFormat(collection, "mock", "test-model", 2)
	backup.Metadata.CreatedAt = createdAt
	backup.Documents = append(backup.Documents, backuppkg.BackupDocument{
		ID: "document-1", Content: "hello", Embedding: []float64{0.25, 0.75},
	})
	backup.Metadata.TotalDocuments = len(backup.Documents)
	if err := backuppkg.WriteBackup(backup, path, compressed); err != nil {
		t.Fatal(err)
	}
}

func TestRunBackupListValidationAndEmptyDirectory(t *testing.T) {
	if err := runBackupList(ListCmd, []string{filepath.Join(t.TempDir(), "missing")}); err == nil || !strings.Contains(err.Error(), "directory not found") {
		t.Fatalf("expected missing directory error, got %v", err)
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runBackupList(ListCmd, []string{file}); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected file type error, got %v", err)
	}

	output, err := captureBackupStdout(t, func() error {
		return runBackupList(ListCmd, []string{t.TempDir()})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "No backup files found") {
		t.Fatalf("unexpected empty directory output: %q", output)
	}

	invalidDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(invalidDir, "bad.weavebak"), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err = captureBackupStdout(t, func() error {
		return runBackupList(ListCmd, []string{invalidDir})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "Failed to read bad.weavebak") || !strings.Contains(output, "No valid backup files found") {
		t.Fatalf("unexpected invalid backup output: %q", output)
	}
}

func TestRunBackupListTableAndJSON(t *testing.T) {
	dir := t.TempDir()
	older := filepath.Join(dir, "older.weavebak")
	newer := filepath.Join(dir, "newer.weavebak.gz")
	writeCommandBackup(t, older, "A very long collection name", time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), false)
	writeCommandBackup(t, newer, "Recent", time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC), true)

	listJSON = false
	t.Cleanup(func() { listJSON = false })
	output, err := captureBackupStdout(t, func() error {
		return runBackupList(ListCmd, []string{dir})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Found 2 backup file(s)", "FILENAME", "newer.weavebak.gz", "Yes", "2026-02-03", "A very long colle..."} {
		if !strings.Contains(output, want) {
			t.Errorf("table output missing %q:\n%s", want, output)
		}
	}
	if strings.Index(output, "newer.weavebak.gz") > strings.Index(output, "older.weavebak") {
		t.Fatalf("expected newest backup first:\n%s", output)
	}

	info, err := getBackupInfo(newer)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Compressed || info.Collection != "Recent" || info.TotalDocuments != 1 || info.BackupSizeBytes == 0 {
		t.Fatalf("unexpected backup info: %#v", info)
	}

	listJSON = true
	output, err = captureBackupStdout(t, func() error {
		return runBackupList(ListCmd, []string{dir})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `"FileName": "newer.weavebak.gz"`) || !strings.Contains(output, `"Compressed": true`) {
		t.Fatalf("unexpected JSON output: %s", output)
	}
}

func TestBackupFormattingHelpers(t *testing.T) {
	for _, tt := range []struct {
		bytes int64
		want  string
	}{
		{bytes: 12, want: "12 B"},
		{bytes: 2 * 1024, want: "2.00 KB"},
		{bytes: 3 * 1024 * 1024, want: "3.00 MB"},
		{bytes: 4 * 1024 * 1024 * 1024, want: "4.00 GB"},
	} {
		if got := formatSize(tt.bytes); got != tt.want {
			t.Errorf("formatSize(%d)=%q, want %q", tt.bytes, got, tt.want)
		}
		if got := formatValidationSize(tt.bytes); got != tt.want {
			t.Errorf("formatValidationSize(%d)=%q, want %q", tt.bytes, got, tt.want)
		}
	}
	if got := truncate("short", 10); got != "short" {
		t.Fatalf("unexpected short string: %q", got)
	}
	if got := truncate("a-long-filename", 10); got != "a-long-..." {
		t.Fatalf("unexpected truncated string: %q", got)
	}
}

func TestRunBackupValidateValidBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "valid.weavebak")
	writeCommandBackup(t, path, "Documents", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), false)

	validateJSON = false
	t.Cleanup(func() { validateJSON = false })
	output, err := captureBackupStdout(t, func() error {
		return runBackupValidate(ValidateCmd, []string{path})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Validating backup", "Collection: Documents", "Documents: 1", "Backup is valid"} {
		if !strings.Contains(output, want) {
			t.Errorf("validation output missing %q:\n%s", want, output)
		}
	}

	validateJSON = true
	output, err = captureBackupStdout(t, func() error {
		return runBackupValidate(ValidateCmd, []string{path})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `"Valid": true`) || !strings.Contains(output, `"Collection": "Documents"`) {
		t.Fatalf("unexpected validation JSON: %s", output)
	}
}

func TestOutputValidationTextIncludesWarnings(t *testing.T) {
	result := &backuppkg.ValidationResult{
		Valid:           true,
		Warnings:        []string{"backup contains no documents"},
		Collection:      "Empty",
		Version:         backuppkg.BackupVersion,
		TotalDocuments:  0,
		BackupSizeBytes: 2 * 1024,
	}
	output, err := captureBackupStdout(t, func() error {
		return outputValidationText(result, "empty.weavebak")
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Validation Warnings", "backup contains no documents", "2.00 KB"} {
		if !strings.Contains(output, want) {
			t.Errorf("warning output missing %q:\n%s", want, output)
		}
	}
}
