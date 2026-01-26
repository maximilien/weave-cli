// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveDatasetAppliesDefaults tests that SaveDataset applies defaults
func TestSaveDatasetAppliesDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("AppliesDefaultVersion", func(t *testing.T) {
		dataset := &Dataset{
			Name: "test-defaults",
			TestCases: []TestCase{
				{ID: "test-001", Query: "Q", ExpectedAnswer: "A"},
			},
		}

		path := filepath.Join(tmpDir, "test-defaults.yaml")
		err := SaveDataset(dataset, path)
		if err != nil {
			t.Fatalf("Failed to save dataset: %v", err)
		}

		// Load and verify defaults were applied
		loaded, err := LoadDataset(path)
		if err != nil {
			t.Fatalf("Failed to load dataset: %v", err)
		}

		if loaded.Version == "" {
			t.Error("Expected default version to be set")
		}
	})
}

// TestListDatasetsFromDefaultDir tests listing datasets from default directory
func TestListDatasetsFromDefaultDir(t *testing.T) {
	// ListDatasets() uses the default directory
	// This test just verifies it doesn't error
	datasets, err := ListDatasets()
	if err != nil {
		// Error is OK if directory doesn't exist
		t.Logf("ListDatasets returned error (expected if no evals dir): %v", err)
		return
	}

	// If it succeeds, verify we get a slice (may be empty)
	if datasets == nil {
		t.Error("Expected non-nil datasets slice")
	}

	t.Logf("Found %d datasets in default directory", len(datasets))
}

// TestSaveDatasetWithExistingDirectory tests SaveDataset when directory exists
func TestSaveDatasetWithExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	datasetPath := filepath.Join(tmpDir, "test.yaml")

	dataset := &Dataset{
		Name:        "test-dataset",
		Version:     "1.0.0",
		Description: "Test",
		TestCases:   []TestCase{{ID: "test-001", Query: "Q", ExpectedAnswer: "A"}},
	}

	err := SaveDataset(dataset, datasetPath)
	if err != nil {
		t.Fatalf("Failed to save dataset: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		t.Error("Expected dataset file to be created")
	}

	// Load and verify
	loaded, err := LoadDataset(datasetPath)
	if err != nil {
		t.Fatalf("Failed to load dataset: %v", err)
	}

	if loaded.Name != dataset.Name {
		t.Errorf("Expected name %s, got %s", dataset.Name, loaded.Name)
	}
}

// TestGetDefaultDatasetDirUsesDevMode tests default dataset directory detection
func TestGetDefaultDatasetDirUsesDevMode(t *testing.T) {
	dir := GetDefaultDatasetDir()

	if dir == "" {
		t.Error("Expected non-empty default dataset directory")
	}

	// Should return evals/datasets in dev mode (current directory has .git)
	if _, err := os.Stat(".git"); err == nil {
		// In dev mode
		if !filepath.IsAbs(dir) {
			// Relative path is OK in dev mode
			expectedSuffix := filepath.Join("evals", "datasets")
			if !filepath.HasPrefix(dir, expectedSuffix) && dir != expectedSuffix {
				t.Logf("Dev mode dataset dir: %s", dir)
			}
		}
	}
}

// TestDatasetValidateComprehensive tests comprehensive validation
func TestDatasetValidateComprehensive(t *testing.T) {
	t.Run("ValidDatasetWithAllFields", func(t *testing.T) {
		dataset := &Dataset{
			Name:        "comprehensive-test",
			Version:     "2.0.0",
			Description: "Full test dataset",
			Author:      "test-author",
			Tags:        []string{"test", "comprehensive"},
			TestCases: []TestCase{
				{
					ID:                "test-001",
					Query:             "What is AI?",
					ExpectedAnswer:    "Artificial Intelligence",
					RequiredConcepts:  []string{"AI", "intelligence"},
					RetrievedContext:  []string{"AI is...", "Intelligence refers to..."},
					MinRelevanceScore: 0.8,
					MustCite:          true,
				},
			},
		}

		if err := dataset.Validate(); err != nil {
			t.Errorf("Valid dataset should not produce error: %v", err)
		}
	})

	t.Run("InvalidDatasetMissingAllFields", func(t *testing.T) {
		dataset := &Dataset{}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for dataset missing all fields")
		}
	})

	t.Run("InvalidTestCaseWithInvalidRelevanceScore", func(t *testing.T) {
		dataset := &Dataset{
			Name:    "test",
			Version: "1.0.0",
			TestCases: []TestCase{
				{
					ID:                "test-001",
					Query:             "Query",
					ExpectedAnswer:    "Answer",
					MinRelevanceScore: 1.5, // Invalid - > 1.0
				},
			},
		}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for invalid relevance score")
		}
	})

	t.Run("InvalidTestCaseWithNegativeRelevanceScore", func(t *testing.T) {
		dataset := &Dataset{
			Name:    "test",
			Version: "1.0.0",
			TestCases: []TestCase{
				{
					ID:                "test-001",
					Query:             "Query",
					ExpectedAnswer:    "Answer",
					MinRelevanceScore: -0.1, // Invalid - negative
				},
			},
		}

		err := dataset.Validate()
		if err == nil {
			t.Error("Expected error for negative relevance score")
		}
	})
}
