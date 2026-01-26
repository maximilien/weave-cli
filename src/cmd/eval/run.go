// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/spf13/cobra"
)

// NewRunCommand creates the eval run command
func NewRunCommand() *cobra.Command {
	var agentName string
	var datasetPath string
	var collection string
	var outputFormat string
	var saveResults bool
	var useOpik bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an agent evaluation",
		Long: `Run an agent evaluation using a test dataset.

This command:
1. Loads the specified dataset
2. Runs the agent on each test case
3. Evaluates the results (accuracy, citations, hallucinations)
4. Generates a summary report
5. Optionally saves results for later analysis

Evaluator Providers:
  By default, uses local LLM-as-judge evaluators with your OpenAI API key.

  --use-opik: Use Opik's evaluators and send evaluation traces to Opik dashboard
              Requires OPIK_API_KEY environment variable to be set

              Benefits:
              - Better LLM-as-judge evaluators
              - Rich dashboard visualization
              - Cost tracking
              - Historical trends
              - Production monitoring

Examples:
  # Run evaluation with local evaluators (default)
  weave eval run --agent rag-agent --dataset baseline

  # Run with Opik evaluators
  weave eval run --agent rag-agent --dataset baseline --use-opik

  # Run with specific dataset file
  weave eval run --agent qa-agent --dataset evals/datasets/medical.yaml

  # Run with custom collection
  weave eval run --agent rag-agent --dataset baseline --collection MyDocs

  # Save results to file
  weave eval run --agent rag-agent --dataset baseline --save`,
		Run: func(cmd *cobra.Command, args []string) {
			runEvaluation(agentName, datasetPath, collection, outputFormat, saveResults, useOpik)
		},
	}

	cmd.Flags().StringVar(&agentName, "agent", "", "Agent to evaluate (required)")
	cmd.Flags().StringVar(&datasetPath, "dataset", "", "Dataset name or path (required)")
	cmd.Flags().StringVar(&collection, "collection", "", "Collection to query (optional)")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format: text, json, yaml")
	cmd.Flags().BoolVar(&saveResults, "save", true, "Save results to file")
	cmd.Flags().BoolVar(&useOpik, "use-opik", false, "Use Opik evaluators (requires OPIK_API_KEY)")

	cmd.MarkFlagRequired("agent")
	cmd.MarkFlagRequired("dataset")

	return cmd
}

func runEvaluation(agentName, datasetPath, collection, outputFormat string, saveResults bool, useOpik bool) {
	ctx := context.Background()

	// Print header
	color.Cyan("=== Weave Agent Evaluation ===\n\n")

	// Load dataset
	fmt.Printf("Loading dataset: %s\n", datasetPath)
	dataset, err := loadDataset(datasetPath)
	if err != nil {
		color.Red("Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Name: %s\n", dataset.Name)
	fmt.Printf("  Test cases: %d\n", len(dataset.TestCases))
	fmt.Println()

	// Create LLM client (needed for local evaluators and agent execution)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		color.Red("Error: OPENAI_API_KEY environment variable is required\n")
		os.Exit(1)
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		color.Red("Error creating LLM client: %v\n", err)
		os.Exit(1)
	}

	// Create evaluator provider
	var provider evaluation.EvaluatorProvider
	var providerType evaluation.ProviderType

	if useOpik {
		providerType = evaluation.ProviderTypeOpik
	} else {
		providerType = evaluation.ProviderTypeLocal
	}

	provider, err = evaluation.CreateProvider(ctx, providerType, llmClient)
	if err != nil {
		if useOpik {
			color.Yellow("Warning: Failed to create Opik provider: %v\n", err)
			color.Yellow("Falling back to local evaluators...\n\n")
			provider = evaluation.NewLocalProvider(llmClient)
		} else {
			color.Red("Error creating evaluator provider: %v\n", err)
			os.Exit(1)
		}
	}

	// Display provider info
	if provider.Name() == "opik" {
		color.Cyan("Using Opik evaluators\n")
		fmt.Printf("  Workspace: %s\n", os.Getenv("OPIK_WORKSPACE"))
		fmt.Printf("  Project: %s\n", os.Getenv("OPIK_PROJECT_NAME"))
	} else {
		color.Cyan("Using local evaluators\n")
	}
	fmt.Println()

	// Create runner
	runner := evaluation.NewRunner(llmClient)

	// Run evaluation
	fmt.Printf("Running evaluation with agent: %s\n", agentName)
	if collection != "" {
		fmt.Printf("Collection: %s\n", collection)
	}
	fmt.Println()

	startTime := time.Now()

	run, err := runner.RunEvaluationWithProvider(ctx, dataset, agentName, collection, provider)
	if err != nil {
		color.Red("Error running evaluation: %v\n", err)
		os.Exit(1)
	}

	// Cleanup Opik provider if needed
	if opikProvider, ok := provider.(*evaluation.OpikProvider); ok {
		defer opikProvider.Shutdown(ctx)
	}

	elapsed := time.Since(startTime)

	// Print summary
	fmt.Println()
	printSummary(run, elapsed, provider.Name())

	// Print detailed results
	if outputFormat == "text" {
		fmt.Println()
		printDetailedResults(run)
	}

	// Save results
	if saveResults {
		fmt.Println()
		savedPath, err := evaluation.SaveResults(run, "json")
		if err != nil {
			color.Yellow("Warning: Failed to save results: %v\n", err)
		} else {
			color.Green("✅ Results saved to: %s\n", savedPath)
			fmt.Printf("View results: weave eval show %s\n", run.ID)
		}
	}

	// Show Opik dashboard link if using Opik
	if provider.Name() == "opik" {
		fmt.Println()
		color.Cyan("📊 View detailed results in Opik dashboard:\n")
		workspace := os.Getenv("OPIK_WORKSPACE")
		project := os.Getenv("OPIK_PROJECT_NAME")
		if workspace == "" {
			workspace = "default"
		}
		if project == "" {
			project = "weave-cli-queries"
		}
		fmt.Printf("   https://www.comet.com/%s/%s\n\n", workspace, project)
		fmt.Println("   Dashboard includes:")
		fmt.Println("   - Detailed trace of each evaluation")
		fmt.Println("   - Cost breakdown")
		fmt.Println("   - Historical trends")
		fmt.Println("   - Export to CSV/JSON")
	}
}

func loadDataset(path string) (*evaluation.Dataset, error) {
	// Check if it's a path or a name
	if _, err := os.Stat(path); err == nil {
		// It's a file path
		return evaluation.LoadDataset(path)
	}

	// Try as dataset name in default directory
	datasetDir := evaluation.GetDefaultDatasetDir()
	fullPath := fmt.Sprintf("%s/%s.yaml", datasetDir, path)

	return evaluation.LoadDataset(fullPath)
}

func printSummary(run *evaluation.EvaluationRun, elapsed time.Duration, providerName string) {
	color.Cyan("=== Evaluation Summary ===\n")

	summary := run.Summary

	// Provider info
	fmt.Printf("Provider:      %s\n", providerName)

	// Overall results
	fmt.Printf("Total Tests:   %d\n", summary.TotalTests)

	if summary.PassedTests > 0 {
		color.Green("Passed:        %d (%.1f%%)\n", summary.PassedTests, summary.PassRate)
	} else {
		fmt.Printf("Passed:        %d (%.1f%%)\n", summary.PassedTests, summary.PassRate)
	}

	if summary.FailedTests > 0 {
		color.Red("Failed:        %d (%.1f%%)\n", summary.FailedTests, 100-summary.PassRate)
	} else {
		fmt.Printf("Failed:        %d (%.1f%%)\n", summary.FailedTests, 100-summary.PassRate)
	}

	fmt.Println()

	// Scores
	fmt.Println("Average Scores:")
	fmt.Printf("  Accuracy:       %.2f\n", summary.AvgAccuracy)
	fmt.Printf("  Citation:       %.2f\n", summary.AvgCitation)
	fmt.Printf("  Hallucination:  %.2f (higher is better)\n", summary.AvgHallucination)

	fmt.Println()

	// Performance
	fmt.Println("Performance:")
	fmt.Printf("  Total time:     %.2fs\n", float64(elapsed.Milliseconds())/1000)
	fmt.Printf("  Avg per test:   %.2fs\n", summary.AvgTime/1000)
}

func printDetailedResults(run *evaluation.EvaluationRun) {
	color.Cyan("=== Detailed Results ===\n\n")

	for i, result := range run.Results {
		// Print test case header
		if result.Passed {
			color.Green("✓ Test %d: %s\n", i+1, result.TestCaseID)
		} else {
			color.Red("✗ Test %d: %s\n", i+1, result.TestCaseID)
		}

		fmt.Printf("  Query: %s\n", result.Query)
		fmt.Printf("  Scores: Accuracy=%.2f, Citation=%.2f, Hallucination=%.2f\n",
			result.AccuracyScore, result.CitationScore, result.HallucinationScore)

		// Print errors if any
		if len(result.Errors) > 0 {
			color.Red("  Errors:\n")
			for _, err := range result.Errors {
				fmt.Printf("    - %s\n", err)
			}
		}

		fmt.Println()
	}
}
