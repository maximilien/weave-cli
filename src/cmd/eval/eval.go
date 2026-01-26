// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"github.com/spf13/cobra"
)

// NewEvalCommand creates the eval command
func NewEvalCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate RAG agents and pipelines",
		Long: `Evaluate RAG agents and pipelines using test datasets.

The eval command provides tools for:
- Running agent evaluations with test datasets
- Viewing evaluation results and metrics
- Comparing different agent configurations
- Tracking evaluation history

Available Commands:
  run        Run an evaluation with a dataset
  benchmark  Compare multiple agents on the same dataset
  show       Show evaluation results
  list       List all evaluation runs
  datasets   Manage evaluation datasets

Examples:
  # Run evaluation
  weave eval run --agent rag-agent --dataset baseline

  # Benchmark multiple agents
  weave eval benchmark --agents rag-agent,qa-agent --dataset baseline

  # Show results
  weave eval show run-20260123-140530

  # List all runs
  weave eval list

  # List datasets
  weave eval datasets list`,
	}

	// Add subcommands
	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewBenchmarkCommand())
	cmd.AddCommand(NewShowCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDatasetsCommand())

	return cmd
}
