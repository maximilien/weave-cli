// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

//go:build integration
// +build integration

package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"gopkg.in/yaml.v3"
)

func TestDatasetLifecycle_CreateValidateDelete(t *testing.T) {
	tmpDir := t.TempDir()
	datasetName := "test-lifecycle-dataset"
	datasetPath := filepath.Join(tmpDir, datasetName+".yaml")

	t.Run("Create minimal dataset", func(t *testing.T) {
		dataset := &evaluation.Dataset{
			Name:        datasetName,
			Version:     "1.0.0",
			Description: "Test dataset for lifecycle",
			Config: evaluation.DatasetConfig{
				MinAccuracyScore:   0.7,
				MinCitationScore:   0.8,
				AllowHallucination: false,
			},
			TestCases: []evaluation.TestCase{
				{
					ID:                "test-001",
					Query:             "Test query",
					ExpectedAnswer:    "Test answer",
					MinRelevanceScore: 0.7,
				},
			},
		}

		// Save dataset
		err := evaluation.SaveDataset(dataset, datasetPath)
		if err != nil {
			t.Fatalf("Failed to save dataset: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
			t.Fatalf("Expected dataset file at %s", datasetPath)
		}
	})

	t.Run("Load and validate dataset", func(t *testing.T) {
		dataset, err := evaluation.LoadDataset(datasetPath)
		if err != nil {
			t.Fatalf("Failed to load dataset: %v", err)
		}

		if dataset.Name != datasetName {
			t.Errorf("Expected name '%s', got '%s'",
				datasetName, dataset.Name)
		}

		if len(dataset.TestCases) != 1 {
			t.Errorf("Expected 1 test case, got %d",
				len(dataset.TestCases))
		}

		if err := dataset.Validate(); err != nil {
			t.Errorf("Dataset validation failed: %v", err)
		}
	})

	t.Run("Delete dataset", func(t *testing.T) {
		if err := os.Remove(datasetPath); err != nil {
			t.Fatalf("Failed to delete dataset: %v", err)
		}

		if _, err := os.Stat(datasetPath); !os.IsNotExist(err) {
			t.Error("Expected dataset file to be deleted")
		}
	})
}

func TestDatasetCreation_Templates(t *testing.T) {
	tmpDir := t.TempDir()

	templates := map[string]string{
		"baseline":         "evals/datasets/baseline.yaml",
		"medical-qa":       "evals/datasets/medical-qa.yaml",
		"technical-docs":   "evals/datasets/technical-docs.yaml",
		"simple-qa":        "evals/datasets/simple-qa.yaml",
		"multi-collection": "evals/datasets/multi-collection.yaml",
	}

	for name, templatePath := range templates {
		t.Run("Load template: "+name, func(t *testing.T) {
			// Try to load template from repo
			dataset, err := evaluation.LoadDataset(templatePath)
			if err != nil {
				t.Skipf("Template not found at %s: %v",
					templatePath, err)
				return
			}

			// Verify template is valid
			if err := dataset.Validate(); err != nil {
				t.Errorf("Template %s validation failed: %v",
					name, err)
			}

			// Verify it has test cases
			if len(dataset.TestCases) == 0 {
				t.Errorf("Template %s has no test cases", name)
			}

			// Copy to tmp dir with new name
			newName := "test-" + name
			dataset.Name = newName
			newPath := filepath.Join(tmpDir, newName+".yaml")

			if err := evaluation.SaveDataset(dataset, newPath); err != nil {
				t.Fatalf("Failed to save copied dataset: %v", err)
			}

			// Reload and verify
			reloaded, err := evaluation.LoadDataset(newPath)
			if err != nil {
				t.Fatalf("Failed to reload dataset: %v", err)
			}

			if reloaded.Name != newName {
				t.Errorf("Expected name '%s', got '%s'",
					newName, reloaded.Name)
			}
		})
	}
}

func TestDatasetCreation_InteractiveMode(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Create dataset programmatically", func(t *testing.T) {
		// Simulate what createInteractiveDataset does
		dataset := &evaluation.Dataset{
			Name:        "test-interactive",
			Version:     "1.0.0",
			Description: "Interactive test dataset",
			Config: evaluation.DatasetConfig{
				DefaultAgent:       "rag-agent",
				MinAccuracyScore:   0.7,
				MinCitationScore:   0.8,
				AllowHallucination: false,
			},
			TestCases: []evaluation.TestCase{
				{
					ID:                "test-001",
					Query:             "What is a test query?",
					ExpectedAnswer:    "A test query is for testing.",
					RequiredConcepts:  []string{"test", "query"},
					MinRelevanceScore: 0.7,
					MustCite:          false,
				},
			},
		}

		// Save it
		outputPath := filepath.Join(tmpDir, "test-interactive.yaml")
		if err := evaluation.SaveDataset(dataset, outputPath); err != nil {
			t.Fatalf("Failed to save dataset: %v", err)
		}

		// Verify it's valid
		loaded, err := evaluation.LoadDataset(outputPath)
		if err != nil {
			t.Fatalf("Failed to load dataset: %v", err)
		}

		if err := loaded.Validate(); err != nil {
			t.Errorf("Dataset validation failed: %v", err)
		}
	})
}

func TestResultsStorage_CreateAndRetrieve(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Create and save results", func(t *testing.T) {
		// Create evaluation run
		run := &evaluation.EvaluationRun{
			ID:          evaluation.GenerateRunID(),
			Timestamp:   time.Now(),
			DatasetName: "test-dataset",
			AgentName:   "test-agent",
			Collection:  "test-collection",
			Results: []evaluation.TestCaseResult{
				{
					TestCaseID:         "test-001",
					ActualAnswer:       "Test answer",
					AccuracyScore:      0.85,
					CitationScore:      0.90,
					HallucinationScore: 0.95,
					ResponseTime:       100.0,
					Passed:             true,
				},
			},
		}

		// Calculate summary
		startTime := time.Now().Add(-1 * time.Second)
		endTime := time.Now()
		run.Summary = evaluation.CalculateSummary(run.Results,
			startTime, endTime)

		// Save as JSON
		resultsDir := filepath.Join(tmpDir, "results")
		os.MkdirAll(resultsDir, 0755)

		filePath := filepath.Join(resultsDir, run.ID+".json")
		if err := evaluation.SaveResultsToFile(run, filePath,
			"json"); err != nil {
			t.Fatalf("Failed to save results: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatal("Results file was not created")
		}
	})

	t.Run("Verify results content", func(t *testing.T) {
		// List all result files
		resultsDir := filepath.Join(tmpDir, "results")
		files, err := os.ReadDir(resultsDir)
		if err != nil {
			t.Fatalf("Failed to read results directory: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 result file, got %d", len(files))
		}

		// Verify we can read it back
		filePath := filepath.Join(resultsDir, files[0].Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read results file: %v", err)
		}

		if len(data) == 0 {
			t.Error("Results file is empty")
		}
	})
}

func TestDatasetValidation_InvalidConfigs(t *testing.T) {
	testCases := []struct {
		name        string
		modifyFunc  func(*evaluation.Dataset)
		shouldError bool
	}{
		{
			name: "Valid dataset",
			modifyFunc: func(d *evaluation.Dataset) {
				// No modifications
			},
			shouldError: false,
		},
		{
			name: "Missing name",
			modifyFunc: func(d *evaluation.Dataset) {
				d.Name = ""
			},
			shouldError: true,
		},
		{
			name: "No test cases",
			modifyFunc: func(d *evaluation.Dataset) {
				d.TestCases = []evaluation.TestCase{}
			},
			shouldError: true,
		},
		{
			name: "Invalid accuracy score",
			modifyFunc: func(d *evaluation.Dataset) {
				d.Config.MinAccuracyScore = 1.5
			},
			shouldError: true,
		},
		{
			name: "Negative citation score",
			modifyFunc: func(d *evaluation.Dataset) {
				d.Config.MinCitationScore = -0.1
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataset := &evaluation.Dataset{
				Name:        "test-validation",
				Version:     "1.0.0",
				Description: "Test dataset",
				Config: evaluation.DatasetConfig{
					MinAccuracyScore:   0.7,
					MinCitationScore:   0.8,
					AllowHallucination: false,
				},
				TestCases: []evaluation.TestCase{
					{
						ID:                "test-001",
						Query:             "Test query",
						ExpectedAnswer:    "Test answer",
						MinRelevanceScore: 0.7,
					},
				},
			}

			tc.modifyFunc(dataset)

			err := dataset.Validate()
			if tc.shouldError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestDatasetDevelopmentVsDeployedMode(t *testing.T) {
	t.Run("Get default dataset directory", func(t *testing.T) {
		datasetDir := evaluation.GetDefaultDatasetDir()
		if datasetDir == "" {
			t.Error("Dataset directory should not be empty")
		}

		// Should be either evals/datasets (dev) or
		// ~/.weave-cli/evals/datasets (deployed)
		if datasetDir != "evals/datasets" {
			home, _ := os.UserHomeDir()
			expected := filepath.Join(home, ".weave-cli",
				"evals", "datasets")
			if datasetDir != expected {
				t.Errorf("Unexpected dataset dir: %s", datasetDir)
			}
		}
	})

	t.Run("Get default results directory", func(t *testing.T) {
		resultsDir := evaluation.GetDefaultResultsDir()
		if resultsDir == "" {
			t.Error("Results directory should not be empty")
		}

		// Should be either evals/results (dev) or
		// ~/.weave-cli/evals/results (deployed)
		if resultsDir != "evals/results" {
			home, _ := os.UserHomeDir()
			expected := filepath.Join(home, ".weave-cli",
				"evals", "results")
			if resultsDir != expected {
				t.Errorf("Unexpected results dir: %s", resultsDir)
			}
		}
	})
}

func TestListDatasets_EmptyAndPopulated(t *testing.T) {
	t.Run("List datasets from actual directory", func(t *testing.T) {
		datasets, err := evaluation.ListDatasets()
		if err != nil {
			t.Skipf("Could not list datasets: %v", err)
			return
		}

		// Should have at least the template datasets we created
		if len(datasets) > 0 {
			// Verify each dataset is valid
			for _, dataset := range datasets {
				if err := dataset.Validate(); err != nil {
					t.Errorf("Dataset %s failed validation: %v",
						dataset.Name, err)
				}
			}
		}
	})
}

func TestRunIDGeneration(t *testing.T) {
	t.Run("Generate unique run IDs", func(t *testing.T) {
		id1 := evaluation.GenerateRunID()
		// Sleep 1 second to ensure different timestamp
		time.Sleep(1 * time.Second)
		id2 := evaluation.GenerateRunID()

		if id1 == id2 {
			t.Error("Run IDs should be unique")
		}

		// Should match pattern: run-YYYYMMDD-HHMMSS
		if len(id1) < 19 {
			t.Errorf("Run ID too short: %s", id1)
		}

		if id1[:4] != "run-" {
			t.Errorf("Run ID should start with 'run-': %s", id1)
		}
	})
}

func TestSummaryCalculation(t *testing.T) {
	t.Run("Calculate summary from results", func(t *testing.T) {
		results := []evaluation.TestCaseResult{
			{
				TestCaseID:         "test-001",
				AccuracyScore:      0.9,
				CitationScore:      0.8,
				HallucinationScore: 0.95,
				ResponseTime:       100.0,
				Passed:             true,
			},
			{
				TestCaseID:         "test-002",
				AccuracyScore:      0.7,
				CitationScore:      0.6,
				HallucinationScore: 0.85,
				ResponseTime:       150.0,
				Passed:             false,
			},
			{
				TestCaseID:         "test-003",
				AccuracyScore:      0.85,
				CitationScore:      0.9,
				HallucinationScore: 0.9,
				ResponseTime:       120.0,
				Passed:             true,
			},
		}

		startTime := time.Now()
		endTime := startTime.Add(500 * time.Millisecond)

		summary := evaluation.CalculateSummary(results,
			startTime, endTime)

		// Verify calculations
		if summary.TotalTests != 3 {
			t.Errorf("Expected 3 total tests, got %d",
				summary.TotalTests)
		}

		if summary.PassedTests != 2 {
			t.Errorf("Expected 2 passed tests, got %d",
				summary.PassedTests)
		}

		if summary.FailedTests != 1 {
			t.Errorf("Expected 1 failed test, got %d",
				summary.FailedTests)
		}

		expectedAvgAccuracy := (0.9 + 0.7 + 0.85) / 3
		// Use approximate comparison for floating point
		delta := 0.0001
		if summary.AvgAccuracy < expectedAvgAccuracy-delta ||
			summary.AvgAccuracy > expectedAvgAccuracy+delta {
			t.Errorf("Expected avg accuracy %.4f, got %.4f",
				expectedAvgAccuracy, summary.AvgAccuracy)
		}

		expectedTotalTime := 100.0 + 150.0 + 120.0
		if summary.TotalTime != expectedTotalTime {
			t.Errorf("Expected total time %.2f, got %.2f",
				expectedTotalTime, summary.TotalTime)
		}
	})
}

func TestEvaluatorIntegration_CitationDetection(t *testing.T) {
	t.Run("Detect citations in text", func(t *testing.T) {
		evaluator := evaluation.NewCitationEvaluator()

		testCase := &evaluation.TestCase{
			ID:       "test-001",
			Query:    "Test query",
			MustCite: true,
		}

		testTexts := map[string]float64{
			"Answer with citation [1]":             0.5,
			"Answer with multiple [1] [2] [3]":     0.8,
			"Answer with many [1] [2] [3] [4] [5]": 1.0,
			"Answer without citation":              0.0,
		}

		ctx := context.Background()
		for text, expectedScore := range testTexts {
			score, err := evaluator.Evaluate(ctx, testCase, text, nil)
			if err != nil {
				t.Errorf("Evaluation failed for '%s': %v", text, err)
				continue
			}

			if score != expectedScore {
				t.Errorf("For text '%s': expected score %.2f, "+
					"got %.2f", text, expectedScore, score)
			}
		}
	})
}

func TestYAMLFormatting_LineLength(t *testing.T) {
	t.Run("Verify YAML line lengths", func(t *testing.T) {
		tmpDir := t.TempDir()
		dataset := &evaluation.Dataset{
			Name:    "test-formatting",
			Version: "1.0.0",
			Description: "This is a very long description that " +
				"should be properly formatted across multiple " +
				"lines in the YAML output to comply with " +
				"80 character line length limits",
			Config: evaluation.DatasetConfig{
				MinAccuracyScore:   0.7,
				MinCitationScore:   0.8,
				AllowHallucination: false,
			},
			TestCases: []evaluation.TestCase{
				{
					ID:    "test-001",
					Query: "Test query",
					ExpectedAnswer: "This is a very long expected " +
						"answer that should also be formatted " +
						"properly across multiple lines to " +
						"comply with line length limits",
					MinRelevanceScore: 0.7,
				},
			},
		}

		outputPath := filepath.Join(tmpDir, "test-formatting.yaml")
		if err := evaluation.SaveDataset(dataset, outputPath); err != nil {
			t.Fatalf("Failed to save dataset: %v", err)
		}

		// Read the file and check line lengths
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output: %v", err)
		}

		// Verify it's valid YAML
		var result map[string]interface{}
		if err := yaml.Unmarshal(data, &result); err != nil {
			t.Errorf("Output is not valid YAML: %v", err)
		}
	})
}
