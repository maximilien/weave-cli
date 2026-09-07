// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
)

func TestEvaluationCommandConstruction(t *testing.T) {
	commands := []*cobra.Command{
		NewEvalCommand(),
		NewRunCommand(),
		NewBenchmarkCommand(),
		NewShowCommand(),
		NewListCommand(),
		NewDatasetsCommand(),
		NewDatasetsUploadOpikCommand(),
		NewDatasetsListCommand(),
		NewDatasetsShowCommand(),
		NewDatasetsValidateCommand(),
		NewDatasetsCreateCommand(),
		NewListEvaluatorsCommand(),
		NewValidateEvaluatorCommand(),
		NewCreateEvaluatorCommand(),
	}
	for _, command := range commands {
		if command == nil || command.Use == "" {
			t.Fatalf("invalid command: %#v", command)
		}
	}

	if got := len(commands[0].Commands()); got != 8 {
		t.Fatalf("eval subcommand count = %d, want 8", got)
	}
	if err := NewShowCommand().Args(nil, nil); err == nil {
		t.Fatal("show command accepted missing run ID")
	}
	if err := NewDatasetsCreateCommand().Args(nil, []string{"one", "two"}); err == nil {
		t.Fatal("dataset create accepted too many arguments")
	}
	if flag := NewBenchmarkCommand().Flag("agents"); flag == nil {
		t.Fatal("benchmark command is missing agents flag")
	}
}

func TestEvaluationReportFormatting(t *testing.T) {
	run := sampleEvaluationRun()
	printSummary(run, 2500*time.Millisecond, "local")
	printDetailedResults(run)
	printRunDetails(run)

	empty := &evaluation.EvaluationRun{Summary: evaluation.EvaluationSummary{}}
	printSummary(empty, 0, "local")
	printDetailedResults(empty)
	printRunDetails(empty)
}

func TestBenchmarkFormattingAndScoring(t *testing.T) {
	comparison := &BenchmarkComparison{
		DatasetName: "baseline",
		Collection:  "docs",
		Results: []BenchmarkResult{
			{AgentName: "steady", RunID: "run-1", Summary: evaluation.EvaluationSummary{PassRate: 80, AvgAccuracy: 0.8, AvgCitation: 0.7, AvgHallucination: 0.9, AvgContextRelevance: 0.8, AvgFaithfulness: 0.9, AvgTime: 15}},
			{AgentName: "best", RunID: "run-2", Summary: evaluation.EvaluationSummary{PassRate: 95, AvgAccuracy: 0.95, AvgCitation: 0.9, AvgHallucination: 0.95, AvgContextRelevance: 0.9, AvgFaithfulness: 0.95, AvgTime: 10}},
		},
	}

	for _, format := range []string{"table", "json", "yaml", "csv", "unknown"} {
		displayBenchmarkResults(comparison, format)
	}
	if got := findBestAgent(comparison.Results); got != "best" {
		t.Fatalf("findBestAgent() = %q", got)
	}
	if got := findBestAgent(nil); got != "" {
		t.Fatalf("findBestAgent(nil) = %q", got)
	}
	if score := calculateOverallScore(comparison.Results[1].Summary); score <= 90 {
		t.Fatalf("calculateOverallScore() = %f", score)
	}
	if got := filepathDir("results.json"); got != "." {
		t.Fatalf("filepathDir() = %q", got)
	}

	path := filepath.Join(t.TempDir(), "nested", "benchmark.json")
	saveBenchmarkResults(comparison, path, "json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read benchmark: %v", err)
	}
	if !strings.Contains(string(data), "best") {
		t.Fatalf("benchmark output = %q", data)
	}

	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	saveBenchmarkResults(comparison, filepath.Join(blocked, "benchmark.json"), "yaml")
}

func TestLoadDatasetByPathAndName(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join("configs", "agents"), 0o755); err != nil {
		t.Fatalf("create config marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join("configs", "agents", "rag-agent.yaml"), nil, 0o600); err != nil {
		t.Fatalf("write config marker: %v", err)
	}
	if err := os.MkdirAll(filepath.Join("evals", "datasets"), 0o755); err != nil {
		t.Fatalf("create datasets directory: %v", err)
	}
	dataset := &evaluation.Dataset{
		Name:    "baseline",
		Version: "1.0.0",
		TestCases: []evaluation.TestCase{{
			ID: "test-001", Query: "question", ExpectedAnswer: "answer",
		}},
	}
	path := filepath.Join("evals", "datasets", "baseline.yaml")
	if err := evaluation.SaveDataset(dataset, path); err != nil {
		t.Fatalf("save dataset: %v", err)
	}

	for _, input := range []string{path, "baseline"} {
		loaded, err := loadDataset(input)
		if err != nil {
			t.Fatalf("loadDataset(%q) error: %v", input, err)
		}
		if loaded.Name != "baseline" {
			t.Fatalf("loadDataset(%q) = %#v", input, loaded)
		}
	}
	if _, err := loadDataset("missing"); err == nil {
		t.Fatal("loadDataset() accepted missing dataset")
	}
	if got := loadDatasetForBenchmark(path); got.Name != "baseline" {
		t.Fatalf("loadDatasetForBenchmark() = %#v", got)
	}
}

func sampleEvaluationRun() *evaluation.EvaluationRun {
	return &evaluation.EvaluationRun{
		ID:          "run-test",
		Timestamp:   time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC),
		DatasetName: "baseline",
		AgentName:   "rag-agent",
		Collection:  "docs",
		Summary: evaluation.EvaluationSummary{
			TotalTests: 2, PassedTests: 1, FailedTests: 1, PassRate: 50,
			AvgAccuracy: 0.8, AvgCitation: 0.7, AvgHallucination: 0.9,
			TotalTime: 3000, AvgTime: 1500, Duration: 3 * time.Second,
		},
		Results: []evaluation.TestCaseResult{
			{TestCaseID: "pass", Query: "first", Passed: true, AccuracyScore: 0.9, CitationScore: 0.8, HallucinationScore: 1, CustomScores: map[string]float64{"quality": 0.9}},
			{TestCaseID: "fail", Query: "second", Passed: false, AccuracyScore: 0.7, CitationScore: 0.6, HallucinationScore: 0.8, CustomScores: map[string]float64{"quality": 0.5}, Errors: []string{"incorrect"}},
		},
	}
}
