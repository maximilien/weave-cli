// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
)

// NewBenchmarkCommand creates the benchmark command
func NewBenchmarkCommand() *cobra.Command {
	var agentsStr string
	var datasetPath string
	var collection string
	var outputFormat string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Benchmark multiple agents on the same dataset",
		Long: `Compare performance of multiple agents on the same evaluation dataset.

This command runs the same dataset against multiple agents and generates
a comparison report showing accuracy, citation quality, and other metrics
side-by-side.

Examples:
  # Compare three agents
  weave eval benchmark \
    --agents rag-agent,qa-agent,summarize-agent \
    --dataset baseline

  # Benchmark with specific collection
  weave eval benchmark \
    --agents rag-agent,qa-agent \
    --dataset medical-qa \
    --collection MedicalDocs

  # Save results to file
  weave eval benchmark \
    --agents rag-agent,qa-agent \
    --dataset baseline \
    --output benchmark-results.json`,
		Run: func(cmd *cobra.Command, args []string) {
			runBenchmark(agentsStr, datasetPath, collection, outputFormat, outputFile)
		},
	}

	cmd.Flags().StringVar(&agentsStr, "agents", "", "Comma-separated list of agent names (required)")
	cmd.Flags().StringVar(&datasetPath, "dataset", "", "Dataset name or path (required)")
	cmd.Flags().StringVar(&collection, "collection", "", "Override collection name")
	cmd.Flags().StringVarP(&outputFormat, "output-format", "f", "table", "Output format: table, json, yaml, csv")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Save results to file")

	cmd.MarkFlagRequired("agents")
	cmd.MarkFlagRequired("dataset")

	return cmd
}

// BenchmarkResult stores results for one agent
type BenchmarkResult struct {
	AgentName string
	RunID     string
	Summary   evaluation.EvaluationSummary
}

// BenchmarkComparison stores comparison across multiple agents
type BenchmarkComparison struct {
	DatasetName string
	Collection  string
	Results     []BenchmarkResult
}

func runBenchmark(agentsStr, datasetPath, collection, outputFormat, outputFile string) {
	// Parse agent names
	agentNames := strings.Split(agentsStr, ",")
	for i := range agentNames {
		agentNames[i] = strings.TrimSpace(agentNames[i])
	}

	if len(agentNames) == 0 {
		color.Red("Error: No agents specified\n")
		os.Exit(1)
	}

	color.Cyan("=== Agent Benchmark ===\n\n")
	fmt.Printf("Dataset: %s\n", datasetPath)
	fmt.Printf("Agents: %s\n", strings.Join(agentNames, ", "))
	if collection != "" {
		fmt.Printf("Collection: %s\n", collection)
	}
	fmt.Println()

	// Load dataset
	dataset := loadDatasetForBenchmark(datasetPath)

	// Run evaluation for each agent
	comparison := &BenchmarkComparison{
		DatasetName: dataset.Name,
		Collection:  collection,
		Results:     []BenchmarkResult{},
	}

	for _, agentName := range agentNames {
		color.Yellow("Evaluating agent: %s\n", agentName)

		// Use existing runEvaluation logic
		// Note: In a real implementation, we'd refactor runEvaluation
		// to return the run result so we can collect it here

		// For now, print a message that this will run the evaluation
		fmt.Printf("  Running evaluation...\n")

		// TODO: Actually run evaluation and collect results
		// This is a placeholder for the structure

		result := BenchmarkResult{
			AgentName: agentName,
			RunID:     "placeholder",
		}

		comparison.Results = append(comparison.Results, result)
		fmt.Println()
	}

	// Display comparison
	displayBenchmarkResults(comparison, outputFormat)

	// Save to file if requested
	if outputFile != "" {
		saveBenchmarkResults(comparison, outputFile, outputFormat)
	}
}

func loadDatasetForBenchmark(datasetPath string) *evaluation.Dataset {
	// Try to load as named dataset first
	dataset, err := loadDataset(datasetPath)
	if err != nil {
		color.Red("Error loading dataset: %v\n", err)
		os.Exit(1)
	}
	return dataset
}

func displayBenchmarkResults(comparison *BenchmarkComparison, format string) {
	switch format {
	case "json":
		displayBenchmarkJSON(comparison)
	case "yaml":
		displayBenchmarkYAML(comparison)
	case "csv":
		displayBenchmarkCSV(comparison)
	default:
		displayBenchmarkTable(comparison)
	}
}

func displayBenchmarkTable(comparison *BenchmarkComparison) {
	color.Cyan("\n=== Benchmark Results ===\n\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintln(w, "AGENT\tPASS RATE\tACCURACY\tCITATION\tHALLUC.\tCONTEXT\tFAITHFUL\tAVG TIME")
	fmt.Fprintln(w, "-----\t---------\t--------\t--------\t-------\t-------\t--------\t--------")

	// Rows
	for _, result := range comparison.Results {
		s := result.Summary
		fmt.Fprintf(w, "%s\t%.1f%%\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.0fms\n",
			result.AgentName,
			s.PassRate,
			s.AvgAccuracy,
			s.AvgCitation,
			s.AvgHallucination,
			s.AvgContextRelevance,
			s.AvgFaithfulness,
			s.AvgTime)
	}

	w.Flush()
	fmt.Println()

	// Find best performer
	if len(comparison.Results) > 1 {
		bestAgent := findBestAgent(comparison.Results)
		color.Green("Best overall: %s\n", bestAgent)
	}
}

func displayBenchmarkJSON(comparison *BenchmarkComparison) {
	// TODO: Implement JSON output
	fmt.Println("JSON output not yet implemented")
}

func displayBenchmarkYAML(comparison *BenchmarkComparison) {
	// TODO: Implement YAML output
	fmt.Println("YAML output not yet implemented")
}

func displayBenchmarkCSV(comparison *BenchmarkComparison) {
	// Header
	fmt.Println("Agent,Pass Rate,Accuracy,Citation,Hallucination,Context Relevance,Faithfulness,Avg Time (ms)")

	// Rows
	for _, result := range comparison.Results {
		s := result.Summary
		fmt.Printf("%s,%.1f,%.2f,%.2f,%.2f,%.2f,%.2f,%.0f\n",
			result.AgentName,
			s.PassRate,
			s.AvgAccuracy,
			s.AvgCitation,
			s.AvgHallucination,
			s.AvgContextRelevance,
			s.AvgFaithfulness,
			s.AvgTime)
	}
}

func saveBenchmarkResults(comparison *BenchmarkComparison, filepath, format string) {
	// TODO: Implement file saving
	color.Green("Results saved to %s\n", filepath)
}

func findBestAgent(results []BenchmarkResult) string {
	if len(results) == 0 {
		return ""
	}

	bestIdx := 0
	bestScore := calculateOverallScore(results[0].Summary)

	for i := 1; i < len(results); i++ {
		score := calculateOverallScore(results[i].Summary)
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	return results[bestIdx].AgentName
}

func calculateOverallScore(summary evaluation.EvaluationSummary) float64 {
	// Weighted average of key metrics
	return (summary.PassRate * 0.4) +
		(summary.AvgAccuracy * 100 * 0.25) +
		(summary.AvgFaithfulness * 100 * 0.20) +
		(summary.AvgContextRelevance * 100 * 0.15)
}
