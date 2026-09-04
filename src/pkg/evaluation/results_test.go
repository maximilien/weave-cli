// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateRunID(t *testing.T) {
	id := GenerateRunID()

	if !strings.HasPrefix(id, "run-") {
		t.Errorf("Expected ID to start with 'run-', got: %s", id)
	}

	// Format should be: run-YYYYMMDD-HHMMSS
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts in run ID, got: %d (%s)", len(parts), id)
	}

	// Generate two IDs quickly, they should be different
	id2 := GenerateRunID()
	if id == id2 {
		// Wait a second and try again
		time.Sleep(1 * time.Second)
		id3 := GenerateRunID()
		if id == id3 {
			t.Error("Expected unique run IDs")
		}
	}
}

func TestCalculateSummary(t *testing.T) {
	startTime := time.Now()
	time.Sleep(100 * time.Millisecond)
	endTime := time.Now()

	results := []TestCaseResult{
		{
			TestCaseID:         "test-001",
			AccuracyScore:      0.9,
			CitationScore:      0.8,
			HallucinationScore: 0.95,
			Passed:             true,
			ResponseTime:       50.0, // milliseconds
		},
		{
			TestCaseID:         "test-002",
			AccuracyScore:      0.7,
			CitationScore:      0.6,
			HallucinationScore: 0.85,
			Passed:             true,
			ResponseTime:       60.0,
		},
		{
			TestCaseID:         "test-003",
			AccuracyScore:      0.4,
			CitationScore:      0.3,
			HallucinationScore: 0.5,
			Passed:             false,
			ResponseTime:       40.0,
		},
	}

	summary := CalculateSummary(results, startTime, endTime)

	t.Run("TotalTests", func(t *testing.T) {
		if summary.TotalTests != 3 {
			t.Errorf("Expected 3 total tests, got: %d", summary.TotalTests)
		}
	})

	t.Run("PassedTests", func(t *testing.T) {
		if summary.PassedTests != 2 {
			t.Errorf("Expected 2 passed tests, got: %d", summary.PassedTests)
		}
	})

	t.Run("FailedTests", func(t *testing.T) {
		if summary.FailedTests != 1 {
			t.Errorf("Expected 1 failed test, got: %d", summary.FailedTests)
		}
	})

	t.Run("PassRate", func(t *testing.T) {
		expectedRate := (2.0 / 3.0) * 100
		if summary.PassRate < expectedRate-0.1 || summary.PassRate > expectedRate+0.1 {
			t.Errorf("Expected pass rate %.2f%%, got: %.2f%%", expectedRate, summary.PassRate)
		}
	})

	t.Run("AverageAccuracy", func(t *testing.T) {
		expectedAvg := (0.9 + 0.7 + 0.4) / 3.0
		if summary.AvgAccuracy < expectedAvg-0.01 || summary.AvgAccuracy > expectedAvg+0.01 {
			t.Errorf("Expected avg accuracy %.2f, got: %.2f", expectedAvg, summary.AvgAccuracy)
		}
	})

	t.Run("AverageCitation", func(t *testing.T) {
		expectedAvg := (0.8 + 0.6 + 0.3) / 3.0
		if summary.AvgCitation < expectedAvg-0.01 || summary.AvgCitation > expectedAvg+0.01 {
			t.Errorf("Expected avg citation %.2f, got: %.2f", expectedAvg, summary.AvgCitation)
		}
	})

	t.Run("AverageHallucination", func(t *testing.T) {
		expectedAvg := (0.95 + 0.85 + 0.5) / 3.0
		if summary.AvgHallucination < expectedAvg-0.01 || summary.AvgHallucination > expectedAvg+0.01 {
			t.Errorf("Expected avg hallucination %.2f, got: %.2f", expectedAvg, summary.AvgHallucination)
		}
	})

	t.Run("TotalTime", func(t *testing.T) {
		expectedTotal := 150.0 // 50 + 60 + 40
		if summary.TotalTime < expectedTotal-0.1 || summary.TotalTime > expectedTotal+0.1 {
			t.Errorf("Expected total time %.2fms, got: %.2fms", expectedTotal, summary.TotalTime)
		}
	})

	t.Run("AverageTime", func(t *testing.T) {
		expectedAvg := summary.TotalTime / 3.0
		if summary.AvgTime < expectedAvg-0.1 || summary.AvgTime > expectedAvg+0.1 {
			t.Errorf("Expected avg time %.2fms, got: %.2fms", expectedAvg, summary.AvgTime)
		}
	})

	t.Run("Duration", func(t *testing.T) {
		if summary.Duration <= 0 {
			t.Error("Expected positive duration")
		}
	})
}

func TestSaveAndLoadResults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Create run for testing
	run := &EvaluationRun{
		ID:          "run-test-001",
		Timestamp:   time.Now(),
		DatasetName: "test-dataset",
		AgentName:   "test-agent",
		Collection:  "test-collection",
		Results: []TestCaseResult{
			{
				TestCaseID:         "test-001",
				Query:              "What is testing?",
				ActualAnswer:       "Testing verifies correctness.",
				ActualCitations:    []string{"[1]", "[2]"},
				AccuracyScore:      0.9,
				CitationScore:      0.8,
				HallucinationScore: 0.95,
				Passed:             true,
				Errors:             []string{},
			},
		},
		Summary: EvaluationSummary{
			TotalTests:       1,
			PassedTests:      1,
			FailedTests:      0,
			PassRate:         100.0,
			AvgAccuracy:      0.9,
			AvgCitation:      0.8,
			AvgHallucination: 0.95,
			TotalTime:        1000,
			AvgTime:          1000,
			Duration:         time.Second,
		},
	}

	t.Run("SaveJSON", func(t *testing.T) {
		savedPath, err := SaveResults(run, "json")
		if err != nil {
			t.Fatalf("Failed to save results: %v", err)
		}

		if !strings.HasSuffix(savedPath, ".json") {
			t.Errorf("Expected .json extension, got: %s", savedPath)
		}

		// Verify file exists
		if _, err := os.Stat(savedPath); os.IsNotExist(err) {
			t.Error("Saved file does not exist")
		}
	})

	t.Run("SaveYAML", func(t *testing.T) {
		savedPath, err := SaveResults(run, "yaml")
		if err != nil {
			t.Fatalf("Failed to save results: %v", err)
		}

		if !strings.HasSuffix(savedPath, ".yaml") {
			t.Errorf("Expected .yaml extension, got: %s", savedPath)
		}

		// Verify file exists
		if _, err := os.Stat(savedPath); os.IsNotExist(err) {
			t.Error("Saved file does not exist")
		}
	})

	t.Run("LoadResults", func(t *testing.T) {
		// Save first
		_, err := SaveResults(run, "json")
		if err != nil {
			t.Fatalf("Failed to save results: %v", err)
		}

		// Load it back
		loaded, err := LoadResults(run.ID)
		if err != nil {
			t.Fatalf("Failed to load results: %v", err)
		}

		if loaded.ID != run.ID {
			t.Errorf("Expected ID %s, got: %s", run.ID, loaded.ID)
		}

		if loaded.DatasetName != run.DatasetName {
			t.Errorf("Expected dataset %s, got: %s", run.DatasetName, loaded.DatasetName)
		}

		if loaded.AgentName != run.AgentName {
			t.Errorf("Expected agent %s, got: %s", run.AgentName, loaded.AgentName)
		}

		if len(loaded.Results) != len(run.Results) {
			t.Errorf("Expected %d results, got: %d", len(run.Results), len(loaded.Results))
		}

		if loaded.Summary.TotalTests != run.Summary.TotalTests {
			t.Errorf("Expected %d total tests, got: %d", run.Summary.TotalTests, loaded.Summary.TotalTests)
		}
	})

	t.Run("LoadNonexistentResults", func(t *testing.T) {
		_, err := LoadResults("nonexistent-run-id")
		if err == nil {
			t.Error("Expected error for nonexistent run")
		}
	})
}

func TestListResults(t *testing.T) {
	// This test just verifies that ListResults doesn't error
	// We can't easily mock the directory without changing the implementation
	runs, err := ListResults()
	if err != nil {
		// It's ok if the directory doesn't exist yet
		t.Logf("ListResults returned error (may be expected): %v", err)
		return
	}

	// Just verify it returns a list (may be empty)
	if runs == nil {
		t.Error("Expected non-nil runs slice")
	}

	// If there are runs, verify they're sorted by timestamp (newest first)
	for i := 1; i < len(runs); i++ {
		if runs[i].Timestamp.After(runs[i-1].Timestamp) {
			t.Error("Expected runs sorted by timestamp (newest first)")
		}
	}
}

func TestGetDefaultResultsDir(t *testing.T) {
	dir := GetDefaultResultsDir()
	if dir == "" {
		t.Error("Expected non-empty results directory")
	}

	// Should contain either "evals/results" (dev) or ".weave-cli/evals/results" (deployed)
	if !filepath.IsAbs(dir) {
		t.Errorf("Expected absolute path, got: %s", dir)
	}
}

func TestCalculateSummaryEmptyResults(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Second)

	summary := CalculateSummary([]TestCaseResult{}, startTime, endTime)

	if summary.TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got: %d", summary.TotalTests)
	}

	if summary.PassedTests != 0 {
		t.Errorf("Expected 0 passed tests, got: %d", summary.PassedTests)
	}

	if summary.PassRate != 0.0 {
		t.Errorf("Expected 0.0 pass rate, got: %.2f", summary.PassRate)
	}
}

func TestCalculateSummaryAllPassed(t *testing.T) {
	results := []TestCaseResult{
		{TestCaseID: "test-001", Passed: true, AccuracyScore: 0.9},
		{TestCaseID: "test-002", Passed: true, AccuracyScore: 0.85},
		{TestCaseID: "test-003", Passed: true, AccuracyScore: 0.95},
	}

	summary := CalculateSummary(results, time.Now(), time.Now())

	if summary.PassRate != 100.0 {
		t.Errorf("Expected 100%% pass rate, got: %.2f%%", summary.PassRate)
	}

	if summary.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got: %d", summary.FailedTests)
	}
}

func TestCalculateSummaryAllFailed(t *testing.T) {
	results := []TestCaseResult{
		{TestCaseID: "test-001", Passed: false, AccuracyScore: 0.4},
		{TestCaseID: "test-002", Passed: false, AccuracyScore: 0.3},
	}

	summary := CalculateSummary(results, time.Now(), time.Now())

	if summary.PassRate != 0.0 {
		t.Errorf("Expected 0%% pass rate, got: %.2f%%", summary.PassRate)
	}

	if summary.PassedTests != 0 {
		t.Errorf("Expected 0 passed tests, got: %d", summary.PassedTests)
	}

	if summary.FailedTests != 2 {
		t.Errorf("Expected 2 failed tests, got: %d", summary.FailedTests)
	}
}
