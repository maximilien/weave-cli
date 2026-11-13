// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
)

func TestBenchmarkPersistsRunsAndSkipsMissingAgents(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "provider fallback"}[fallback], func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			t.Setenv("OPIK_ENABLED", "false")
			t.Setenv("OPIK_API_KEY", "")
			dir := filepath.Join(root, "configs", "agents")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "benchmark-fixture.yaml"), []byte("name: benchmark-fixture\ntype: rag\nllm:\n  model: gpt-4o\nsystem_prompt: Answer from sources.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			dataset := &evaluation.Dataset{Name: "benchmark-dataset", Version: "1.0.0", TestCases: []evaluation.TestCase{{ID: "case-1", Query: "Fixture question?", ExpectedAnswer: "answer"}}}
			path := filepath.Join(root, "dataset.yaml")
			if err := evaluation.SaveDataset(dataset, path); err != nil {
				t.Fatal(err)
			}
			calls := make(chan struct{}, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				calls <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"message": "fixture rejected completion", "type": "invalid_request_error"}})
			}))
			t.Cleanup(server.Close)
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
			output := filepath.Join(root, "comparison.json")
			runBenchmark(" absent-fixture , benchmark-fixture ", path, "Target", "json", output, fallback)
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			var comparison BenchmarkComparison
			if err := json.Unmarshal(data, &comparison); err != nil {
				t.Fatal(err)
			}
			if comparison.DatasetName != dataset.Name || comparison.Collection != "Target" || len(comparison.Results) != 1 {
				t.Fatalf("comparison=%+v", comparison)
			}
			result := comparison.Results[0]
			if result.AgentName != "benchmark-fixture" || result.Summary.TotalTests != 1 || result.Summary.FailedTests != 1 {
				t.Fatalf("agent result=%+v", result)
			}
			run, err := evaluation.LoadResults(result.RunID)
			if err != nil {
				t.Fatal(err)
			}
			if run.Config.Parameters["evaluator_provider"] != "local" || len(run.Results) != 1 || len(run.Results[0].Errors) != 1 {
				t.Fatalf("saved run=%+v", run)
			}
			if len(calls) != 1 {
				t.Fatalf("completion calls=%d; missing agent should make none", len(calls))
			}
		})
	}
}
