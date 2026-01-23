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

// NewDatasetsCommand creates the datasets command
func NewDatasetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "datasets",
		Aliases: []string{"dataset", "ds"},
		Short:   "Manage evaluation datasets",
		Long: `Manage evaluation datasets.

Available Commands:
  list      List all datasets
  show      Show dataset details
  validate  Validate a dataset file

Examples:
  # List all datasets
  weave eval datasets list

  # Show dataset details
  weave eval datasets show baseline

  # Validate a dataset
  weave eval datasets validate evals/datasets/my-dataset.yaml`,
	}

	cmd.AddCommand(NewDatasetsListCommand())
	cmd.AddCommand(NewDatasetsShowCommand())
	cmd.AddCommand(NewDatasetsValidateCommand())

	return cmd
}

// NewDatasetsListCommand creates the datasets list command
func NewDatasetsListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all evaluation datasets",
		Long: `List all available evaluation datasets.

Examples:
  # List datasets
  weave eval datasets list

  # List as JSON
  weave eval datasets list --output json`,
		Run: func(cmd *cobra.Command, args []string) {
			listDatasets(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

// NewDatasetsShowCommand creates the datasets show command
func NewDatasetsShowCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show NAME",
		Short: "Show dataset details",
		Long: `Show detailed information about a dataset.

Examples:
  # Show dataset
  weave eval datasets show baseline

  # Show as YAML
  weave eval datasets show baseline -o yaml`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showDataset(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

// NewDatasetsValidateCommand creates the datasets validate command
func NewDatasetsValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate FILE",
		Short: "Validate a dataset file",
		Long: `Validate a dataset configuration file.

Examples:
  # Validate dataset
  weave eval datasets validate evals/datasets/my-dataset.yaml`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			validateDataset(args[0])
		},
	}

	return cmd
}

func listDatasets(outputFormat string) {
	datasets, err := evaluation.ListDatasets()
	if err != nil {
		color.Red("Error loading datasets: %v\n", err)
		os.Exit(1)
	}

	if len(datasets) == 0 {
		fmt.Println("No datasets found.")
		fmt.Printf("\nDatasets directory: %s\n", evaluation.GetDefaultDatasetDir())
		fmt.Println("\nCreate a dataset by adding a YAML file to this directory.")
		return
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(datasets, "", "  ")
		fmt.Println(string(data))

	case "yaml":
		data, _ := yaml.Marshal(datasets)
		fmt.Println(string(data))

	default:
		// Text format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tTEST CASES\tDESCRIPTION")
		fmt.Fprintln(w, "----\t-------\t----------\t-----------")

		for _, dataset := range datasets {
			desc := dataset.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
				dataset.Name, dataset.Version, len(dataset.TestCases), desc)
		}

		w.Flush()

		fmt.Printf("\nTotal: %d datasets\n", len(datasets))
	}
}

func showDataset(name, outputFormat string) {
	// Load dataset
	dataset, err := loadDataset(name)
	if err != nil {
		color.Red("Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(dataset, "", "  ")
		fmt.Println(string(data))

	case "yaml":
		data, _ := yaml.Marshal(dataset)
		fmt.Println(string(data))

	default:
		// Text format
		color.Cyan("=== Dataset: %s ===\n\n", dataset.Name)

		fmt.Println("Metadata:")
		fmt.Printf("  Name:        %s\n", dataset.Name)
		fmt.Printf("  Version:     %s\n", dataset.Version)
		fmt.Printf("  Description: %s\n", dataset.Description)
		if dataset.Author != "" {
			fmt.Printf("  Author:      %s\n", dataset.Author)
		}
		if len(dataset.Tags) > 0 {
			fmt.Printf("  Tags:        %v\n", dataset.Tags)
		}
		fmt.Println()

		fmt.Println("Configuration:")
		if dataset.Config.DefaultAgent != "" {
			fmt.Printf("  Default Agent:      %s\n", dataset.Config.DefaultAgent)
		}
		if dataset.Config.DefaultCollection != "" {
			fmt.Printf("  Default Collection: %s\n", dataset.Config.DefaultCollection)
		}
		fmt.Printf("  Min Accuracy:       %.2f\n", dataset.Config.MinAccuracyScore)
		fmt.Printf("  Min Citation:       %.2f\n", dataset.Config.MinCitationScore)
		fmt.Printf("  Allow Hallucination: %v\n", dataset.Config.AllowHallucination)
		fmt.Println()

		fmt.Printf("Test Cases: %d\n\n", len(dataset.TestCases))

		for i, tc := range dataset.TestCases {
			fmt.Printf("%d. %s\n", i+1, tc.ID)
			if tc.Description != "" {
				fmt.Printf("   %s\n", tc.Description)
			}
			fmt.Printf("   Query: %s\n", tc.Query)
			fmt.Println()
		}
	}
}

func validateDataset(filepath string) {
	dataset, err := evaluation.LoadDataset(filepath)
	if err != nil {
		color.Red("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	color.Green("✅ Dataset is valid!\n\n")
	fmt.Printf("Dataset: %s\n", dataset.Name)
	fmt.Printf("Version: %s\n", dataset.Version)
	fmt.Printf("Test cases: %d\n", len(dataset.TestCases))
}
