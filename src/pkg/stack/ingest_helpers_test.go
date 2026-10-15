// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
