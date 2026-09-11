// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestPipelineCommandContract(t *testing.T) {
	if PipelineCmd.Use != "pipeline" || len(PipelineCmd.Commands()) != 1 {
		t.Fatalf("PipelineCmd = %q with %d subcommands", PipelineCmd.Use, len(PipelineCmd.Commands()))
	}
	if err := ingestCmd.ValidateArgs([]string{"source"}); err != nil {
		t.Fatalf("ingest command rejected source: %v", err)
	}
	if err := ingestCmd.ValidateArgs(nil); err == nil {
		t.Fatal("ingest command accepted missing source")
	}
}

func TestRunIngestValidationAndEmptyScan(t *testing.T) {
	if err := runIngest(ingestCmd, []string{filepath.Join(t.TempDir(), "missing")}); err == nil || !strings.Contains(err.Error(), "source path does not exist") {
		t.Fatalf("runIngest(missing source) error = %v", err)
	}

	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
	configPath := filepath.Join(root, "config.yaml")
	configData := []byte(`databases:
  default: mock
  vector_databases:
    - name: mock
      type: mock
      simulate_embeddings: true
`)
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	viper.Reset()
	viper.Set("config", configPath)
	t.Cleanup(viper.Reset)

	collection = "Documents"
	glob = "*.txt"
	exclude = nil
	recursive = true
	batchSize = 10
	workers = 1
	metadata = map[string]string{"source": "test"}
	dryRun = true
	resume = false
	stateFile = filepath.Join(root, "state.json")
	outputFormat = "json"
	reportFile = ""
	quiet = true

	t.Setenv("OPENAI_API_KEY", "")
	if err := runIngest(ingestCmd, []string{root}); err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("runIngest(missing API key) error = %v", err)
	}

	t.Setenv("OPENAI_API_KEY", "test-key")
	if err := runIngest(ingestCmd, []string{root}); err != nil {
		t.Fatalf("runIngest(empty scan) error = %v", err)
	}
}
