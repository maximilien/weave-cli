// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
)

func TestRunEvaluationRecordsCompletionFailures(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "Opik fallback"}[fallback], func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			t.Setenv("OPIK_API_KEY", "")
			t.Setenv("OPIK_ENABLED", "false")
			dir := filepath.Join(root, "configs", "agents")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "fixture.yaml"), []byte("name: fixture\ntype: rag\nllm:\n  model: gpt-4o\nsystem_prompt: Answer from sources.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			dataset := &evaluation.Dataset{Name: "fixture-dataset", Version: "1.0.0", TestCases: []evaluation.TestCase{{ID: "case-1", Query: "Fixture question?", ExpectedAnswer: "Fixture answer."}}}
			path := filepath.Join(root, "dataset.yaml")
			if err := evaluation.SaveDataset(dataset, path); err != nil {
				t.Fatal(err)
			}
			requests := make(chan struct{}, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				requests <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"message": "fixture completion rejected", "type": "invalid_request_error"}})
			}))
			t.Cleanup(server.Close)
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
			runEvaluation("fixture", path, "Target", "text", true, fallback)
			if len(requests) != 1 {
				t.Fatalf("completion calls=%d", len(requests))
			}
			files, err := filepath.Glob(filepath.Join(evaluation.GetDefaultResultsDir(), "*.json"))
			if err != nil || len(files) != 1 {
				t.Fatalf("result files=%v error=%v", files, err)
			}
			data, err := os.ReadFile(files[0])
			if err != nil {
				t.Fatal(err)
			}
			var run evaluation.EvaluationRun
			if err := json.Unmarshal(data, &run); err != nil {
				t.Fatal(err)
			}
			if run.AgentName != "fixture" || run.DatasetName != dataset.Name || run.Collection != "Target" || run.Config.Parameters["evaluator_provider"] != "local" {
				t.Fatalf("run identity/provider=%+v", run)
			}
			if run.Summary.TotalTests != 1 || run.Summary.FailedTests != 1 || len(run.Results) != 1 || run.Results[0].Passed {
				t.Fatalf("failed evaluation summary=%+v results=%+v", run.Summary, run.Results)
			}
			if len(run.Results[0].Errors) != 1 || !strings.Contains(run.Results[0].Errors[0], "fixture completion rejected") {
				t.Fatalf("recorded errors=%v", run.Results[0].Errors)
			}
		})
	}
}
