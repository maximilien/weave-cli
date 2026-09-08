// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckpointLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "checkpoint.json")
	checkpoint, err := LoadCheckpoint(path)
	if err != nil || checkpoint != nil {
		t.Fatalf("LoadCheckpoint(missing) = %#v, %v", checkpoint, err)
	}

	start := time.Now().Add(-time.Hour).Round(time.Second)
	existing := &IngestCheckpoint{
		StartTime:      start,
		CompletedFiles: []string{"done.txt"},
		FailedFiles:    []FailedFile{{File: "old.txt", Attempts: 2, Error: "old failure"}},
	}
	checkpoint = InitCheckpoint("documents", []string{"done.txt", "next.txt", "bad.txt"}, existing)
	if checkpoint.Collection != "documents" || !checkpoint.StartTime.Equal(start) {
		t.Fatalf("InitCheckpoint() = %#v", checkpoint)
	}
	if len(checkpoint.CompletedFiles) != 1 || len(checkpoint.FailedFiles) != 1 || len(checkpoint.PendingFiles) != 3 {
		t.Fatalf("InitCheckpoint() lists = %#v", checkpoint)
	}

	checkpoint.MarkFileCompleted("next.txt")
	checkpoint.MarkFileFailed("bad.txt", 3, "parse failed")
	if failed := checkpoint.GetFailedFile("bad.txt"); failed == nil || failed.Attempts != 3 || failed.Error != "parse failed" {
		t.Fatalf("GetFailedFile() = %#v", failed)
	}
	if checkpoint.GetFailedFile("missing.txt") != nil {
		t.Fatal("GetFailedFile() found a missing file")
	}
	if len(checkpoint.PendingFiles) != 1 || checkpoint.PendingFiles[0] != "done.txt" {
		t.Fatalf("PendingFiles = %v", checkpoint.PendingFiles)
	}

	beforeSave := time.Now()
	if err := SaveCheckpoint(checkpoint, path); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}
	if checkpoint.LastUpdate.Before(beforeSave) {
		t.Fatalf("LastUpdate = %s, want after %s", checkpoint.LastUpdate, beforeSave)
	}
	loaded, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("LoadCheckpoint() error = %v", err)
	}
	if loaded.Collection != checkpoint.Collection || len(loaded.CompletedFiles) != 2 || len(loaded.FailedFiles) != 2 {
		t.Fatalf("LoadCheckpoint() = %#v", loaded)
	}
}

func TestCheckpointFailures(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "checkpoint.json")
		if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadCheckpoint(path)
		if err == nil || !strings.Contains(err.Error(), "failed to parse checkpoint") {
			t.Fatalf("LoadCheckpoint() error = %v", err)
		}
	})

	t.Run("read directory", func(t *testing.T) {
		_, err := LoadCheckpoint(t.TempDir())
		if err == nil || !strings.Contains(err.Error(), "failed to read checkpoint") {
			t.Fatalf("LoadCheckpoint() error = %v", err)
		}
	})

	t.Run("create directory", func(t *testing.T) {
		root := t.TempDir()
		parent := filepath.Join(root, "file")
		if err := os.WriteFile(parent, nil, 0600); err != nil {
			t.Fatal(err)
		}
		err := SaveCheckpoint(&IngestCheckpoint{}, filepath.Join(parent, "checkpoint.json"))
		if err == nil || !strings.Contains(err.Error(), "failed to create checkpoint directory") {
			t.Fatalf("SaveCheckpoint() error = %v", err)
		}
	})
}

func TestFilterFilesFromCheckpoint(t *testing.T) {
	files := []string{"one.txt", "/docs/two.txt", "three.txt", "four.txt"}
	tests := []struct {
		name       string
		checkpoint *IngestCheckpoint
		resume     string
		want       []string
		skipped    int
		wantErr    string
	}{
		{name: "unchanged", want: files},
		{name: "completed", checkpoint: &IngestCheckpoint{CompletedFiles: []string{"one.txt", "three.txt"}}, want: []string{"/docs/two.txt", "four.txt"}, skipped: 2},
		{name: "resume basename", resume: "two.txt", want: []string{"/docs/two.txt", "three.txt", "four.txt"}, skipped: 1},
		{name: "resume full path", resume: "/docs/two.txt", want: []string{"/docs/two.txt", "three.txt", "four.txt"}, skipped: 1},
		{name: "completed before resume", checkpoint: &IngestCheckpoint{CompletedFiles: []string{"one.txt"}}, resume: "three.txt", want: []string{"three.txt", "four.txt"}, skipped: 2},
		{name: "resume missing", resume: "missing.txt", wantErr: "resume file not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, skipped, err := FilterFilesFromCheckpoint(files, tt.checkpoint, tt.resume)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FilterFilesFromCheckpoint() error = %v", err)
			}
			if strings.Join(got, "|") != strings.Join(tt.want, "|") || skipped != tt.skipped {
				t.Fatalf("FilterFilesFromCheckpoint() = %v, %d; want %v, %d", got, skipped, tt.want, tt.skipped)
			}
		})
	}
}

func TestInitCheckpointWithoutExistingState(t *testing.T) {
	checkpoint := InitCheckpoint("new", []string{"one"}, nil)
	if checkpoint.Collection != "new" || len(checkpoint.PendingFiles) != 1 {
		t.Fatalf("InitCheckpoint() = %#v", checkpoint)
	}
	if checkpoint.StartTime.IsZero() || checkpoint.LastUpdate.IsZero() {
		t.Fatal("InitCheckpoint() did not set timestamps")
	}
}
