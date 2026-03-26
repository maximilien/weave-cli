// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewListCommand creates the eval list command
func NewListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all evaluation runs",
		Long: `List all available evaluation runs with summary information.

Examples:
  # List all runs
  weave eval list

  # List as JSON
  weave eval list --output json`,
		Run: func(cmd *cobra.Command, args []string) {
			listRuns(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

func listRuns(outputFormat string) {
	// Load all runs
	runs, err := evaluation.ListResults()
	if err != nil {
		color.Red("Error loading runs: %v\n", err)
		os.Exit(1)
	}

	if len(runs) == 0 {
		fmt.Println("No evaluation runs found.")
		fmt.Println("\nRun an evaluation:")
		fmt.Println("  weave eval run --agent rag-agent --dataset baseline")
		return
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(runs, "", "  ")
		fmt.Println(string(data))

	case "yaml":
		data, _ := yaml.Marshal(runs)
		fmt.Println(string(data))

	default:
		// Text format with table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "RUN ID\tTIMESTAMP\tDATASET\tAGENT\tPASS RATE\tAVG ACCURACY")
		fmt.Fprintln(w, "------\t---------\t-------\t-----\t---------\t------------")

		for _, run := range runs {
			timestamp := run.Timestamp.Format("2006-01-02 15:04")
			passRate := fmt.Sprintf("%.1f%%", run.Summary.PassRate)
			avgAccuracy := fmt.Sprintf("%.2f", run.Summary.AvgAccuracy)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				run.ID, timestamp, run.DatasetName, run.AgentName, passRate, avgAccuracy)
		}

		w.Flush()

		fmt.Printf("\nTotal: %d runs\n", len(runs))
		fmt.Println("\nView details: weave eval show <run-id>")
	}
}
