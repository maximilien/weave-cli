// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Dataset represents an evaluation dataset with test cases
type Dataset struct {
	// Metadata
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
	Author      string   `yaml:"author,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`

	// Configuration
	Config DatasetConfig `yaml:"config"`

	// Test cases
	TestCases []TestCase `yaml:"test_cases"`
}

// DatasetConfig contains dataset-level configuration
type DatasetConfig struct {
	// Agent to use (optional - can be overridden at runtime)
	DefaultAgent string `yaml:"default_agent,omitempty"`

	// Collection to query (optional - can be overridden at runtime)
	DefaultCollection string `yaml:"default_collection,omitempty"`

	// Evaluation thresholds
	MinAccuracyScore   float64 `yaml:"min_accuracy_score,omitempty"`
	MinCitationScore   float64 `yaml:"min_citation_score,omitempty"`
	AllowHallucination bool    `yaml:"allow_hallucination,omitempty"`
}

// TestCase represents a single evaluation test case
type TestCase struct {
	// Identification
	ID          string   `yaml:"id"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`

	// Input
	Query      string `yaml:"query"`
	Collection string `yaml:"collection,omitempty"` // Override default collection

	// Expected output
	ExpectedAnswer    string   `yaml:"expected_answer"`
	ExpectedCitations []string `yaml:"expected_citations,omitempty"`
	RequiredConcepts  []string `yaml:"required_concepts,omitempty"`

	// Evaluation criteria
	MinRelevanceScore float64 `yaml:"min_relevance_score,omitempty"`
	MustCite          bool    `yaml:"must_cite,omitempty"`
}

// LoadDataset loads a dataset from a YAML file
func LoadDataset(filepath string) (*Dataset, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dataset file: %w", err)
	}

	var dataset Dataset
	if err := yaml.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("failed to parse dataset YAML: %w", err)
	}

	// Apply defaults
	dataset.applyDefaults()

	// Validate
	if err := dataset.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dataset: %w", err)
	}

	return &dataset, nil
}

// SaveDataset saves a dataset to a YAML file
func SaveDataset(dataset *Dataset, filepath string) error {
	data, err := yaml.Marshal(dataset)
	if err != nil {
		return fmt.Errorf("failed to marshal dataset: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write dataset file: %w", err)
	}

	return nil
}

// applyDefaults applies default values to missing fields
func (d *Dataset) applyDefaults() {
	if d.Version == "" {
		d.Version = "1.0.0"
	}

	// Apply default config values
	if d.Config.MinAccuracyScore == 0 {
		d.Config.MinAccuracyScore = 0.7
	}
	if d.Config.MinCitationScore == 0 {
		d.Config.MinCitationScore = 0.8
	}

	// Assign IDs to test cases if missing
	for i := range d.TestCases {
		if d.TestCases[i].ID == "" {
			d.TestCases[i].ID = fmt.Sprintf("test-%03d", i+1)
		}
	}
}

// Validate validates the dataset structure
func (d *Dataset) Validate() error {
	// Required fields
	if d.Name == "" {
		return fmt.Errorf("dataset name is required")
	}

	// Validate name format (lowercase, alphanumeric, hyphens)
	nameRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !nameRegex.MatchString(d.Name) {
		return fmt.Errorf("dataset name must be lowercase alphanumeric with hyphens")
	}

	// Must have at least one test case
	if len(d.TestCases) == 0 {
		return fmt.Errorf("dataset must contain at least one test case")
	}

	// Validate each test case
	for i, tc := range d.TestCases {
		if err := tc.Validate(); err != nil {
			return fmt.Errorf("test case %d (%s): %w", i, tc.ID, err)
		}
	}

	// Validate config thresholds
	if d.Config.MinAccuracyScore < 0 || d.Config.MinAccuracyScore > 1.0 {
		return fmt.Errorf("min_accuracy_score must be between 0.0 and 1.0")
	}
	if d.Config.MinCitationScore < 0 || d.Config.MinCitationScore > 1.0 {
		return fmt.Errorf("min_citation_score must be between 0.0 and 1.0")
	}

	return nil
}

// Validate validates a test case
func (tc *TestCase) Validate() error {
	if tc.Query == "" {
		return fmt.Errorf("query is required")
	}

	if tc.ExpectedAnswer == "" {
		return fmt.Errorf("expected_answer is required")
	}

	if tc.MinRelevanceScore < 0 || tc.MinRelevanceScore > 1.0 {
		return fmt.Errorf("min_relevance_score must be between 0.0 and 1.0")
	}

	return nil
}

// GetDefaultDatasetDir returns the default directory for evaluation datasets
func GetDefaultDatasetDir() string {
	// Check if in development mode
	if _, err := os.Stat("configs/agents/rag-agent.yaml"); err == nil {
		return "evals/datasets"
	}

	// Deployed mode
	home, err := os.UserHomeDir()
	if err != nil {
		return "evals/datasets"
	}

	return filepath.Join(home, ".weave-cli", "evals", "datasets")
}

// ListDatasets returns all datasets in the default directory
func ListDatasets() ([]*Dataset, error) {
	dir := GetDefaultDatasetDir()

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []*Dataset{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read datasets directory: %w", err)
	}

	var datasets []*Dataset
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process YAML files
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		dataset, err := LoadDataset(path)
		if err != nil {
			// Skip invalid datasets
			continue
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}
