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
