// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// EvaluationRun represents a complete evaluation run
type EvaluationRun struct {
	// Metadata
	ID          string    `json:"id" yaml:"id"`
	Timestamp   time.Time `json:"timestamp" yaml:"timestamp"`
	DatasetName string    `json:"dataset_name" yaml:"dataset_name"`
	AgentName   string    `json:"agent_name" yaml:"agent_name"`
	Collection  string    `json:"collection,omitempty" yaml:"collection,omitempty"`

	// Configuration
	Config EvaluationConfig `json:"config" yaml:"config"`

	// Results
	Results []TestCaseResult `json:"results" yaml:"results"`

	// Summary
	Summary EvaluationSummary `json:"summary" yaml:"summary"`
}

// EvaluationConfig stores the configuration used for the evaluation
type EvaluationConfig struct {
	AgentPath  string            `json:"agent_path" yaml:"agent_path"`
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// TestCaseResult represents the result of evaluating a single test case
type TestCaseResult struct {
	// Test case info
	TestCaseID string `json:"test_case_id" yaml:"test_case_id"`
	Query      string `json:"query" yaml:"query"`

	// Actual output
	ActualAnswer    string   `json:"actual_answer" yaml:"actual_answer"`
	ActualCitations []string `json:"actual_citations,omitempty" yaml:"actual_citations,omitempty"`
	ResponseTime    float64  `json:"response_time_ms" yaml:"response_time_ms"`

	// Evaluation scores
	AccuracyScore         float64 `json:"accuracy_score" yaml:"accuracy_score"`
	CitationScore         float64 `json:"citation_score" yaml:"citation_score"`
	HallucinationScore    float64 `json:"hallucination_score" yaml:"hallucination_score"`
	ContextRelevanceScore float64 `json:"context_relevance_score" yaml:"context_relevance_score"`
	FaithfulnessScore     float64 `json:"faithfulness_score" yaml:"faithfulness_score"`

	// Custom evaluator scores
	CustomScores map[string]float64 `json:"custom_scores,omitempty" yaml:"custom_scores,omitempty"`

	// Pass/fail
	Passed bool     `json:"passed" yaml:"passed"`
	Errors []string `json:"errors,omitempty" yaml:"errors,omitempty"`

	// Details
	Details map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

// EvaluationSummary contains aggregate statistics
type EvaluationSummary struct {
	TotalTests  int     `json:"total_tests" yaml:"total_tests"`
	PassedTests int     `json:"passed_tests" yaml:"passed_tests"`
	FailedTests int     `json:"failed_tests" yaml:"failed_tests"`
	PassRate    float64 `json:"pass_rate" yaml:"pass_rate"`

	// Average scores
	AvgAccuracy         float64 `json:"avg_accuracy" yaml:"avg_accuracy"`
	AvgCitation         float64 `json:"avg_citation" yaml:"avg_citation"`
	AvgHallucination    float64 `json:"avg_hallucination" yaml:"avg_hallucination"`
	AvgContextRelevance float64 `json:"avg_context_relevance" yaml:"avg_context_relevance"`
	AvgFaithfulness     float64 `json:"avg_faithfulness" yaml:"avg_faithfulness"`

	// Performance
	TotalTime float64 `json:"total_time_ms" yaml:"total_time_ms"`
	AvgTime   float64 `json:"avg_time_ms" yaml:"avg_time_ms"`

	// Duration
	StartTime time.Time     `json:"start_time" yaml:"start_time"`
	EndTime   time.Time     `json:"end_time" yaml:"end_time"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
}

// GenerateRunID generates a unique run ID
func GenerateRunID() string {
	return fmt.Sprintf("run-%s", time.Now().Format("20060102-150405"))
}

// CalculateSummary calculates summary statistics from results
func CalculateSummary(results []TestCaseResult, startTime, endTime time.Time) EvaluationSummary {
	summary := EvaluationSummary{
		TotalTests: len(results),
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   endTime.Sub(startTime),
	}

	if len(results) == 0 {
		return summary
	}

	var totalAccuracy, totalCitation, totalHallucination, totalContextRel, totalFaithfulness, totalTime float64

	for _, result := range results {
		if result.Passed {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}

		totalAccuracy += result.AccuracyScore
		totalCitation += result.CitationScore
		totalHallucination += result.HallucinationScore
		totalContextRel += result.ContextRelevanceScore
		totalFaithfulness += result.FaithfulnessScore
		totalTime += result.ResponseTime
	}

	summary.PassRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	summary.AvgAccuracy = totalAccuracy / float64(summary.TotalTests)
	summary.AvgCitation = totalCitation / float64(summary.TotalTests)
	summary.AvgHallucination = totalHallucination / float64(summary.TotalTests)
	summary.AvgContextRelevance = totalContextRel / float64(summary.TotalTests)
	summary.AvgFaithfulness = totalFaithfulness / float64(summary.TotalTests)
	summary.TotalTime = totalTime
	summary.AvgTime = totalTime / float64(summary.TotalTests)

	return summary
}

// SaveResults saves evaluation results to a file
func SaveResults(run *EvaluationRun, format string) (string, error) {
	// Get results directory
	resultsDir := GetDefaultResultsDir()

	// Ensure directory exists
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create results directory: %w", err)
	}

	// Generate filename
	var filename string
	var data []byte
	var err error

	switch format {
	case "json":
		filename = filepath.Join(resultsDir, run.ID+".json")
		data, err = json.MarshalIndent(run, "", "  ")
	case "yaml":
		filename = filepath.Join(resultsDir, run.ID+".yaml")
		data, err = yaml.Marshal(run)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write results file: %w", err)
	}

	return filename, nil
}

// LoadResults loads evaluation results from a file
func LoadResults(runID string) (*EvaluationRun, error) {
	resultsDir := GetDefaultResultsDir()

	// Try JSON first
	jsonPath := filepath.Join(resultsDir, runID+".json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var run EvaluationRun
		if err := json.Unmarshal(data, &run); err == nil {
			return &run, nil
		}
	}

	// Try YAML
	yamlPath := filepath.Join(resultsDir, runID+".yaml")
	if data, err := os.ReadFile(yamlPath); err == nil {
		var run EvaluationRun
		if err := yaml.Unmarshal(data, &run); err == nil {
			return &run, nil
		}
	}

	return nil, fmt.Errorf("results not found for run ID: %s", runID)
}

// ListResults returns all available evaluation runs
func ListResults() ([]*EvaluationRun, error) {
	resultsDir := GetDefaultResultsDir()

	// Check if directory exists
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		return []*EvaluationRun{}, nil
	}

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read results directory: %w", err)
	}

	var runs []*EvaluationRun
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".json" && ext != ".yaml" {
			continue
		}

		runID := entry.Name()[:len(entry.Name())-len(ext)]
		run, err := LoadResults(runID)
		if err != nil {
			continue
		}

		runs = append(runs, run)
	}

	return runs, nil
}

// SaveResultsToFile saves evaluation results to a specific file path
func SaveResultsToFile(run *EvaluationRun, filePath, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(run, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(run)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}

// GetDefaultResultsDir returns the default directory for evaluation results
func GetDefaultResultsDir() string {
	// Check if in development mode
	if _, err := os.Stat("configs/agents/rag-agent.yaml"); err == nil {
		return "evals/results"
	}

	// Deployed mode
	home, err := os.UserHomeDir()
	if err != nil {
		return "evals/results"
	}

	return filepath.Join(home, ".weave-cli", "evals", "results")
}
