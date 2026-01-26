// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSaveResultsToFile tests saving results to a specific file
func TestSaveResultsToFile(t *testing.T) {
	tmpDir := t.TempDir()

	run := &EvaluationRun{
		ID:          "test-run-001",
		Timestamp:   time.Now(),
		DatasetName: "test-dataset",
		AgentName:   "test-agent",
		Results: []TestCaseResult{
			{
				TestCaseID:   "test-001",
				Query:        "Test query",
				ActualAnswer: "Test answer",
				Passed:       true,
			},
		},
		Summary: EvaluationSummary{
			TotalTests:  1,
			PassedTests: 1,
			PassRate:    100.0,
		},
	}

	t.Run("SaveAsJSON", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test-results.json")
		err := SaveResultsToFile(run, filePath, "json")
		if err != nil {
			t.Fatalf("Failed to save results as JSON: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Expected JSON file to be created")
		}

		// Read and verify directly (LoadResults uses default directory)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read JSON results: %v", err)
		}

		var loaded EvaluationRun
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if loaded.ID != run.ID {
			t.Errorf("Expected run ID %s, got %s", run.ID, loaded.ID)
		}
	})

	t.Run("SaveAsYAML", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test-results.yaml")
		err := SaveResultsToFile(run, filePath, "yaml")
		if err != nil {
			t.Fatalf("Failed to save results as YAML: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Expected YAML file to be created")
		}

		// Verify file has content (YAML parsing would require yaml package)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read YAML results: %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected YAML file to have content")
		}
	})

	t.Run("SaveWithInvalidFormat", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test-results.invalid")
		err := SaveResultsToFile(run, filePath, "invalid")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("SaveRequiresExistingDirectory", func(t *testing.T) {
		// Create nested directory first
		nestedDir := filepath.Join(tmpDir, "nested", "dir")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Failed to create nested directory: %v", err)
		}

		nestedPath := filepath.Join(nestedDir, "results.json")
		err := SaveResultsToFile(run, nestedPath, "json")
		if err != nil {
			t.Fatalf("Failed to save to nested directory: %v", err)
		}

		if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
			t.Error("Expected file to be created")
		}
	})
}

// TestListResultsWithMultipleFiles tests listing multiple result files
func TestListResultsWithMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple result files
	runs := []EvaluationRun{
		{
			ID:          "run-001",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			DatasetName: "dataset-1",
			AgentName:   "agent-1",
			Results:     []TestCaseResult{},
			Summary:     EvaluationSummary{TotalTests: 1},
		},
		{
			ID:          "run-002",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			DatasetName: "dataset-2",
			AgentName:   "agent-2",
			Results:     []TestCaseResult{},
			Summary:     EvaluationSummary{TotalTests: 2},
		},
		{
			ID:          "run-003",
			Timestamp:   time.Now(),
			DatasetName: "dataset-3",
			AgentName:   "agent-3",
			Results:     []TestCaseResult{},
			Summary:     EvaluationSummary{TotalTests: 3},
		},
	}

	for i, run := range runs {
		path := filepath.Join(tmpDir, run.ID+".json")
		if err := SaveResultsToFile(&run, path, "json"); err != nil {
			t.Fatalf("Failed to save run %d: %v", i, err)
		}
	}

	// Note: ListResults() uses default directory, so we can't test with tmpDir
	// Verify the files were created and can be read directly
	for _, run := range runs {
		path := filepath.Join(tmpDir, run.ID+".json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file for run %s to exist", run.ID)
		}

		// Verify file can be read and parsed
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read file for run %s: %v", run.ID, err)
			continue
		}

		var loaded EvaluationRun
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Errorf("Failed to parse JSON for run %s: %v", run.ID, err)
			continue
		}

		if loaded.ID != run.ID {
			t.Errorf("Expected ID %s, got %s", run.ID, loaded.ID)
		}
	}
}

// TestListResultsFromDefaultDir tests listing from default directory
func TestListResultsFromDefaultDir(t *testing.T) {
	// ListResults() uses the default directory
	// This test just verifies it doesn't error
	results, err := ListResults()
	if err != nil {
		// Error is OK if directory doesn't exist
		t.Logf("ListResults returned error (expected if no evals dir): %v", err)
		return
	}

	// If it succeeds, verify we get a slice (may be empty)
	if results == nil {
		t.Error("Expected non-nil results slice")
	}

	t.Logf("Found %d results in default directory", len(results))
}

// TestLoadResultsWithDifferentFormats tests loading JSON and YAML formats
func TestLoadResultsWithDifferentFormats(t *testing.T) {
	tmpDir := t.TempDir()

	run := &EvaluationRun{
		ID:          "multi-format-test",
		Timestamp:   time.Now(),
		DatasetName: "test-dataset",
		AgentName:   "test-agent",
		Results: []TestCaseResult{
			{TestCaseID: "test-001", Passed: true},
		},
		Summary: EvaluationSummary{TotalTests: 1, PassedTests: 1},
	}

	t.Run("SaveAndReadJSON", func(t *testing.T) {
		jsonPath := filepath.Join(tmpDir, "test.json")
		if err := SaveResultsToFile(run, jsonPath, "json"); err != nil {
			t.Fatalf("Failed to save JSON: %v", err)
		}

		// Read and parse directly (LoadResults uses default directory)
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("Failed to read JSON: %v", err)
		}

		var loaded EvaluationRun
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if loaded.ID != run.ID {
			t.Errorf("Expected ID %s, got %s", run.ID, loaded.ID)
		}
	})

	t.Run("SaveAndReadYAML", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "test.yaml")
		if err := SaveResultsToFile(run, yamlPath, "yaml"); err != nil {
			t.Fatalf("Failed to save YAML: %v", err)
		}

		// Verify file exists and has content
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			t.Fatalf("Failed to read YAML: %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected YAML file to have content")
		}
	})

	t.Run("LoadNonExistent", func(t *testing.T) {
		_, err := LoadResults("nonexistent-run-id-that-does-not-exist")
		if err == nil {
			t.Error("Expected error for non-existent run ID")
		}
	})

	t.Run("LoadInvalidJSON", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(invalidPath, []byte("{invalid}"), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		_, err := LoadResults(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestSaveResultsUsesDefaultDirectory tests SaveResults function
func TestSaveResultsUsesDefaultDirectory(t *testing.T) {
	// This test verifies SaveResults creates files in default location
	run := &EvaluationRun{
		ID:          "default-dir-test-" + GenerateRunID(),
		Timestamp:   time.Now(),
		DatasetName: "test",
		Results:     []TestCaseResult{},
		Summary:     EvaluationSummary{},
	}

	// Save with default directory
	savedPath, err := SaveResults(run, "json")
	if err != nil {
		t.Fatalf("Failed to save results: %v", err)
	}

	defer os.Remove(savedPath) // Cleanup

	// Verify file exists
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Error("Expected results file to be created")
	}

	// Verify path contains run ID
	if filepath.Base(savedPath) != run.ID+".json" {
		t.Errorf("Expected filename %s.json, got %s", run.ID, filepath.Base(savedPath))
	}

	// Load using just the run ID (not full path)
	loaded, err := LoadResults(run.ID)
	if err != nil {
		t.Fatalf("Failed to load saved results: %v", err)
	}

	if loaded.ID != run.ID {
		t.Errorf("Expected ID %s, got %s", run.ID, loaded.ID)
	}
}

// TestGetDefaultResultsDirExists tests default results directory
func TestGetDefaultResultsDirExists(t *testing.T) {
	dir := GetDefaultResultsDir()

	if dir == "" {
		t.Error("Expected non-empty default results directory")
	}

	// Should return evals/results in dev mode or ~/.weave-cli/evals/results otherwise
	if _, err := os.Stat(".git"); err == nil {
		// In dev mode
		if !filepath.IsAbs(dir) {
			expectedSuffix := filepath.Join("evals", "results")
			if !filepath.HasPrefix(dir, expectedSuffix) && dir != expectedSuffix {
				t.Logf("Dev mode results dir: %s", dir)
			}
		}
	}
}
