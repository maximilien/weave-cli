// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewShowCommand creates the eval show command
func NewShowCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show RUN_ID",
		Short: "Show evaluation results",
		Long: `Show detailed results for a specific evaluation run.

Examples:
  # Show results in text format
  weave eval show run-20260123-140530

  # Show as JSON
  weave eval show run-20260123-140530 --output json

  # Show as YAML
  weave eval show run-20260123-140530 -o yaml`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showResults(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

func showResults(runID, outputFormat string) {
	// Load results
	run, err := evaluation.LoadResults(runID)
	if err != nil {
		color.Red("Error loading results: %v\n", err)
		os.Exit(1)
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(run, "", "  ")
		fmt.Println(string(data))

	case "yaml":
		data, _ := yaml.Marshal(run)
		fmt.Println(string(data))

	default:
		// Text format
		printRunDetails(run)
	}
}

func printRunDetails(run *evaluation.EvaluationRun) {
	// Header
	color.Cyan("=== Evaluation Run: %s ===\n\n", run.ID)

	// Metadata
	fmt.Println("Metadata:")
	fmt.Printf("  Timestamp:   %s\n", run.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Dataset:     %s\n", run.DatasetName)
	fmt.Printf("  Agent:       %s\n", run.AgentName)
	if run.Collection != "" {
		fmt.Printf("  Collection:  %s\n", run.Collection)
	}
	fmt.Println()

	// Summary
	summary := run.Summary

	color.Cyan("Summary:\n")
	fmt.Printf("  Total Tests:  %d\n", summary.TotalTests)

	if summary.PassedTests > 0 {
		color.Green("  Passed:       %d (%.1f%%)\n", summary.PassedTests, summary.PassRate)
	} else {
		fmt.Printf("  Passed:       %d (%.1f%%)\n", summary.PassedTests, summary.PassRate)
	}

	if summary.FailedTests > 0 {
		color.Red("  Failed:       %d (%.1f%%)\n", summary.FailedTests, 100-summary.PassRate)
	} else {
		fmt.Printf("  Failed:       %d (%.1f%%)\n", summary.FailedTests, 100-summary.PassRate)
	}

	fmt.Println()

	// Average scores
	fmt.Println("Average Scores:")
	fmt.Printf("  Accuracy:      %.2f\n", summary.AvgAccuracy)
	fmt.Printf("  Citation:      %.2f\n", summary.AvgCitation)
	fmt.Printf("  Hallucination: %.2f\n", summary.AvgHallucination)
	fmt.Println()

	// Performance
	fmt.Println("Performance:")
	fmt.Printf("  Total time:    %.2fs\n", summary.TotalTime/1000)
	fmt.Printf("  Avg per test:  %.2fs\n", summary.AvgTime/1000)
	fmt.Printf("  Duration:      %s\n", summary.Duration)
	fmt.Println()

	// Detailed results
	printDetailedResults(run)
}
