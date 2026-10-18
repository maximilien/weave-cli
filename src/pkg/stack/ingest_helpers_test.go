// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/pipeline"
)

func TestGetEmbeddingDimensions(t *testing.T) {
	tests := map[string]int{
		"text-embedding-3-small":                  1536,
		"text-embedding-3-large":                  3072,
		"text-embedding-ada-002":                  1536,
		"sentence-transformers/all-mpnet-base-v2": 768,
		"sentence-transformers/all-MiniLM-L6-v2":  384,
		"nomic-embed-text":                        768,
		"mxbai-embed-large":                       1024,
		"unknown-model":                           1536,
	}
	for model, want := range tests {
		t.Run(model, func(t *testing.T) {
			if got := getEmbeddingDimensions(model); got != want {
				t.Fatalf("getEmbeddingDimensions(%q) = %d, want %d", model, got, want)
			}
		})
	}
}

func TestStartMilvusPortForwardReportsConfigurationAndExecutableErrors(t *testing.T) {
	t.Run("missing stack configuration", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		if _, err := StartMilvusPortForward(&ClusterInfo{Context: "fixture"}); err == nil || !strings.Contains(err.Error(), "failed to load stack config") {
			t.Fatalf("StartMilvusPortForward() error = %v, want missing-config error", err)
		}
	})

	t.Run("kubectl unavailable", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		config := &StackConfig{
			Version: "1.0", Name: "fixture",
			Runtime:        RuntimeConfig{Kubernetes: KubernetesConfig{Provider: "kind"}},
			Infrastructure: InfrastructureConfig{VectorDB: VectorDBConfig{Type: "milvus", Version: "2.3.0"}},
		}
		if err := SaveStackConfig(config, filepath.Join(workDir, DefaultStackFile)); err != nil {
			t.Fatalf("save stack config: %v", err)
		}
		binDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(binDir, "unrelated"), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir)
		if _, err := StartMilvusPortForward(&ClusterInfo{Context: "fixture"}); err == nil || !strings.Contains(err.Error(), "failed to start port forwarding") {
			t.Fatalf("StartMilvusPortForward() error = %v, want missing-kubectl error", err)
		}
	})
}

func TestRestartIngestionHelpersHandleEmptyAndExistingFiles(t *testing.T) {
	cfg := IngestConfig{RestartEvery: 2, MaxRetries: 1}
	options := &pipeline.IngestOptions{Workers: 1, BatchSize: 1}
	progress := pipeline.NewProgressTracker(false, true)

	batched, err := processFilesWithRestarts(context.Background(), nil, nil, cfg, nil, nil, options, progress)
	if err != nil || batched.FilesScanned != 0 || batched.FilesProcessed != 0 || batched.Status != "success" {
		t.Fatalf("processFilesWithRestarts(empty) = (%+v, %v)", batched, err)
	}

	auto, err := processFilesWithAutoRestart(context.Background(), nil, nil, cfg, nil, nil, options, progress, nil)
	if err != nil || auto.FilesScanned != 0 || auto.FilesProcessed != 0 || auto.Status != "success" {
		t.Fatalf("processFilesWithAutoRestart(empty) = (%+v, %v)", auto, err)
	}

	stateFile := filepath.Join(t.TempDir(), "resume.json")
	if err := os.WriteFile(stateFile, []byte(`{"processed_files":{"already-done":true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	options.Resume = true
	options.StateFile = stateFile
	processor := pipeline.NewProcessor(nil, nil, options, progress)
	files := []pipeline.FileInfo{{Path: "already-done.txt", Hash: "already-done"}}
	batched, err = processFilesWithRestarts(context.Background(), processor, files, cfg, nil, nil, options, progress)
	if err != nil || batched.FilesScanned != 1 || batched.FilesProcessed != 0 {
		t.Fatalf("processFilesWithRestarts(resumed file) = (%+v, %v)", batched, err)
	}

	auto, err = processFilesWithAutoRestart(context.Background(), processor, files, cfg, nil, nil, options, progress, nil)
	if err != nil || auto.FilesScanned != 1 || auto.FilesFailed != 0 || auto.Status != "success" {
		t.Fatalf("processFilesWithAutoRestart(resumed file) = (%+v, %v)", auto, err)
	}
}

func TestAutoRestartRecordsAndAbortsFailedFiles(t *testing.T) {
	progress := pipeline.NewProgressTracker(false, true)
	options := &pipeline.IngestOptions{Workers: 1, BatchSize: 1}
	processor := pipeline.NewProcessor(nil, nil, options, progress)
	files := []pipeline.FileInfo{{Path: "unsupported.fixture", Type: pipeline.FileType("unsupported")}}

	t.Run("skip and checkpoint failed file", func(t *testing.T) {
		checkpointPath := filepath.Join(t.TempDir(), "nested", "checkpoint.json")
		checkpoint := InitCheckpoint("collection", []string{files[0].Path}, nil)
		cfg := IngestConfig{MaxRetries: 0, SkipFailed: true, CheckpointEvery: 1, CheckpointFile: checkpointPath}

		report, err := processFilesWithAutoRestart(context.Background(), processor, files, cfg, nil, nil, options, progress, checkpoint)
		if err != nil {
			t.Fatalf("processFilesWithAutoRestart() error = %v", err)
		}
		if report.FilesScanned != 1 || report.FilesFailed != 1 || report.FilesProcessed != 0 {
			t.Fatalf("report = %+v", report)
		}
		if len(checkpoint.FailedFiles) != 1 || checkpoint.FailedFiles[0].File != files[0].Path {
			t.Fatalf("checkpoint failed files = %+v", checkpoint.FailedFiles)
		}
		if _, err := os.Stat(checkpointPath); err != nil {
			t.Fatalf("checkpoint was not saved: %v", err)
		}
	})

	t.Run("abort after failure", func(t *testing.T) {
		checkpoint := InitCheckpoint("collection", []string{files[0].Path}, nil)
		checkpointPath := filepath.Join(t.TempDir(), "checkpoint.json")
		cfg := IngestConfig{MaxRetries: 0, CheckpointFile: checkpointPath}

		report, err := processFilesWithAutoRestart(context.Background(), processor, files, cfg, nil, nil, options, progress, checkpoint)
		if err == nil || !strings.Contains(err.Error(), "file failed after 1 attempts") {
			t.Fatalf("processFilesWithAutoRestart() error = %v", err)
		}
		if report != nil {
			t.Fatalf("report = %+v, want nil on abort", report)
		}
		if _, err := os.Stat(checkpointPath); err != nil {
			t.Fatalf("checkpoint was not saved before abort: %v", err)
		}
	})
}

func TestProcessFilesWithRestartsAggregatesFailedBatch(t *testing.T) {
	options := &pipeline.IngestOptions{Workers: 1, BatchSize: 1}
	progress := pipeline.NewProgressTracker(false, true)
	processor := pipeline.NewProcessor(nil, nil, options, progress)
	files := []pipeline.FileInfo{
		{Path: "first.unsupported", Type: pipeline.FileType("unsupported")},
		{Path: "second.unsupported", Type: pipeline.FileType("unsupported")},
	}

	report, err := processFilesWithRestarts(context.Background(), processor, files,
		IngestConfig{RestartEvery: 3}, nil, nil, options, progress)
	if err != nil {
		t.Fatalf("processFilesWithRestarts() error = %v", err)
	}
	if report.FilesScanned != 2 || report.FilesFailed != 2 || report.FilesProcessed != 0 || report.Status != "success" {
		t.Fatalf("report = %+v", report)
	}
}

func TestPortForwardContextStopCancels(t *testing.T) {
	canceled := false
	portForward := &PortForwardContext{Cancel: func() { canceled = true }}
	portForward.Stop()
	if !canceled {
		t.Fatal("Stop() did not cancel port forwarding")
	}
	(&PortForwardContext{}).Stop()
}

func TestRestartMilvusReportsConfigurationAndCommandErrors(t *testing.T) {
	t.Run("missing stack config", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := restartMilvus(&ClusterInfo{Context: "fixture"}); err == nil || !strings.Contains(err.Error(), "failed to load stack config") {
			t.Fatalf("restartMilvus() error = %v", err)
		}
	})

	t.Run("kubectl unavailable", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		config := &StackConfig{
			Version: "1.0", Name: "fixture",
			Runtime:        RuntimeConfig{Kubernetes: KubernetesConfig{Provider: "kind"}},
			Infrastructure: InfrastructureConfig{VectorDB: VectorDBConfig{Type: "milvus", Version: "2.3.0"}},
		}
		if err := SaveStackConfig(config, filepath.Join(workDir, DefaultStackFile)); err != nil {
			t.Fatalf("save stack config: %v", err)
		}
		binDir := t.TempDir()
		t.Setenv("PATH", binDir)
		if err := restartMilvus(&ClusterInfo{Context: "fixture"}); err == nil || !strings.Contains(err.Error(), "failed to restart deployment") {
			t.Fatalf("restartMilvus() error = %v", err)
		}
	})

	t.Run("kubectl succeeds", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		config := &StackConfig{
			Version: "1.0", Name: "fixture",
			Runtime:        RuntimeConfig{Kubernetes: KubernetesConfig{Provider: "kind"}},
			Infrastructure: InfrastructureConfig{VectorDB: VectorDBConfig{Type: "milvus", Version: "2.3.0"}},
		}
		if err := SaveStackConfig(config, filepath.Join(workDir, DefaultStackFile)); err != nil {
			t.Fatalf("save stack config: %v", err)
		}
		binDir := t.TempDir()
		kubectl := filepath.Join(binDir, "kubectl")
		if err := os.WriteFile(kubectl, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatalf("write kubectl fixture: %v", err)
		}
		t.Setenv("PATH", binDir)
		if err := restartMilvus(&ClusterInfo{Context: "fixture"}); err != nil {
			t.Fatalf("restartMilvus() error = %v", err)
		}
	})
}
