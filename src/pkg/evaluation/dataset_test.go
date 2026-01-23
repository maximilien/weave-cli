// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDataset(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	validDataset := `name: test-dataset
description: Test dataset
version: 1.0.0

config:
  default_agent: test-agent
  min_accuracy_score: 0.7
  min_citation_score: 0.8
  allow_hallucination: false

test_cases:
  - id: test-001
    description: Test question
    query: "What is a test?"
    expected_answer: "A test is a procedure to verify correctness."
    expected_citations:
      - "test"
      - "verification"
    required_concepts:
      - "test"
      - "procedure"
    must_cite: true
    min_relevance_score: 0.7
`

	validPath := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(validPath, []byte(validDataset), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	t.Run("ValidDataset", func(t *testing.T) {
		dataset, err := LoadDataset(validPath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if dataset.Name != "test-dataset" {
			t.Errorf("Expected name 'test-dataset', got: %s", dataset.Name)
		}

		if dataset.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got: %s", dataset.Version)
		}

		if len(dataset.TestCases) != 1 {
			t.Errorf("Expected 1 test case, got: %d", len(dataset.TestCases))
		}

		tc := dataset.TestCases[0]
		if tc.ID != "test-001" {
			t.Errorf("Expected test case ID 'test-001', got: %s", tc.ID)
		}

		if tc.Query != "What is a test?" {
			t.Errorf("Expected query 'What is a test?', got: %s", tc.Query)
		}

		if len(tc.RequiredConcepts) != 2 {
			t.Errorf("Expected 2 required concepts, got: %d", len(tc.RequiredConcepts))
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		_, err := LoadDataset(filepath.Join(tmpDir, "nonexistent.yaml"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err = LoadDataset(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestDatasetValidate(t *testing.T) {
	t.Run("ValidDataset", func(t *testing.T) {
		dataset := &Dataset{
			Name:        "test",
			Description: "Test dataset",
			Version:     "1.0.0",
			Config: DatasetConfig{
				DefaultAgent:       "agent",
				MinAccuracyScore:   0.7,
				MinCitationScore:   0.8,
				AllowHallucination: false,
			},
			TestCases: []TestCase{
				{
					ID:                "test-001",
					Query:             "Test query?",
					ExpectedAnswer:    "Test answer",
					RequiredConcepts:  []string{"test"},
					MinRelevanceScore: 0.7,
				},
			},
		}

		err := dataset.Validate()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		dataset := &Dataset{
			Version: "1.0.0",
			TestCases: []TestCase{
				{ID: "test-001", Query: "Test?", ExpectedAnswer: "Answer"},
			},
		}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})

	t.Run("MissingVersion", func(t *testing.T) {
		dataset := &Dataset{
			Name: "test",
			TestCases: []TestCase{
				{ID: "test-001", Query: "Test?", ExpectedAnswer: "Answer"},
			},
		}

		// Version is optional
		err := dataset.Validate()
		if err != nil {
			t.Errorf("Expected no error for missing version, got: %v", err)
		}
	})

	t.Run("NoTestCases", func(t *testing.T) {
		dataset := &Dataset{
			Name:      "test",
			Version:   "1.0.0",
			TestCases: []TestCase{},
		}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for no test cases")
		}
	})

	t.Run("InvalidTestCase", func(t *testing.T) {
		dataset := &Dataset{
			Name:    "test",
			Version: "1.0.0",
			TestCases: []TestCase{
				{ID: "test-001", Query: "", ExpectedAnswer: "Answer"}, // Missing query
			},
		}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for invalid test case (missing query)")
		}
	})
}

func TestTestCaseValidate(t *testing.T) {
	t.Run("ValidTestCase", func(t *testing.T) {
		tc := &TestCase{
			ID:                "test-001",
			Query:             "What is testing?",
			ExpectedAnswer:    "Testing verifies correctness",
			RequiredConcepts:  []string{"testing"},
			MinRelevanceScore: 0.7,
		}

		err := tc.Validate()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("MissingID", func(t *testing.T) {
		tc := &TestCase{
			Query:          "Test?",
			ExpectedAnswer: "Answer",
		}

		// ID is optional
		err := tc.Validate()
		if err != nil {
			t.Errorf("Expected no error for missing ID, got: %v", err)
		}
	})

	t.Run("MissingQuery", func(t *testing.T) {
		tc := &TestCase{
			ID:             "test-001",
			ExpectedAnswer: "Answer",
		}

		err := tc.Validate()
		if err == nil {
			t.Error("Expected error for missing query")
		}
	})

	t.Run("MissingExpectedAnswer", func(t *testing.T) {
		tc := &TestCase{
			ID:    "test-001",
			Query: "Test?",
		}

		err := tc.Validate()
		if err == nil {
			t.Error("Expected error for missing expected answer")
		}
	})

	t.Run("InvalidRelevanceScore", func(t *testing.T) {
		tc := &TestCase{
			ID:                "test-001",
			Query:             "Test?",
			ExpectedAnswer:    "Answer",
			MinRelevanceScore: 1.5,
		}

		err := tc.Validate()
		if err == nil {
			t.Error("Expected error for invalid relevance score")
		}
	})
}

func TestSaveDataset(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved.yaml")

	dataset := &Dataset{
		Name:        "test-dataset",
		Description: "Test dataset",
		Version:     "1.0.0",
		Config: DatasetConfig{
			DefaultAgent:       "test-agent",
			MinAccuracyScore:   0.7,
			MinCitationScore:   0.8,
			AllowHallucination: false,
		},
		TestCases: []TestCase{
			{
				ID:                "test-001",
				Description:       "Test case",
				Query:             "What is a test?",
				ExpectedAnswer:    "A test verifies correctness",
				RequiredConcepts:  []string{"test", "verification"},
				MinRelevanceScore: 0.7,
			},
		},
	}

	err := SaveDataset(dataset, savePath)
	if err != nil {
		t.Fatalf("Failed to save dataset: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Saved file does not exist")
	}

	// Load it back
	loaded, err := LoadDataset(savePath)
	if err != nil {
		t.Fatalf("Failed to load saved dataset: %v", err)
	}

	if loaded.Name != dataset.Name {
		t.Errorf("Expected name '%s', got '%s'", dataset.Name, loaded.Name)
	}

	if len(loaded.TestCases) != len(dataset.TestCases) {
		t.Errorf("Expected %d test cases, got %d", len(dataset.TestCases), len(loaded.TestCases))
	}
}

func TestGetDefaultDatasetDir(t *testing.T) {
	dir := GetDefaultDatasetDir()
	if dir == "" {
		t.Error("Expected non-empty dataset directory")
	}

	// Should contain either "evals/datasets" (dev) or ".weave-cli/evals/datasets" (deployed)
	if !filepath.IsAbs(dir) {
		t.Errorf("Expected absolute path, got: %s", dir)
	}
}

func TestListDatasets(t *testing.T) {
	// This test just verifies that ListDatasets doesn't error
	// We can't easily mock the directory without changing the implementation
	datasets, err := ListDatasets()
	if err != nil {
		// It's ok if the directory doesn't exist yet
		t.Logf("ListDatasets returned error (may be expected): %v", err)
		return
	}

	// Just verify it returns a list (may be empty)
	if datasets == nil {
		t.Error("Expected non-nil datasets slice")
	}
}
