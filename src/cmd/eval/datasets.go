// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	cmd.AddCommand(NewDatasetsCreateCommand())

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

// NewDatasetsCreateCommand creates the datasets create command
func NewDatasetsCreateCommand() *cobra.Command {
	var templateName string
	var fromExample string
	var interactive bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a new evaluation dataset",
		Long: `Create a new evaluation dataset from a template, example, or interactively.

Available Templates:
  baseline         - Basic RAG evaluation with 5 test cases
  medical-qa       - Medical/healthcare knowledge tests
  technical-docs   - Technical documentation and API tests
  simple-qa        - Simple Q&A for getting started
  multi-collection - Multi-collection query tests

Examples:
  # Create from template
  weave eval datasets create my-dataset --template baseline

  # Create from existing dataset
  weave eval datasets create my-dataset --from-example evals/datasets/baseline.yaml

  # Create interactively (default)
  weave eval datasets create my-dataset --interactive

  # Create and save to specific path
  weave eval datasets create my-dataset --template simple-qa --output custom/path.yaml`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

			createDataset(name, templateName, fromExample, interactive, outputPath)
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "", "Use built-in template (baseline, medical-qa, technical-docs, simple-qa, multi-collection)")
	cmd.Flags().StringVarP(&fromExample, "from-example", "f", "", "Copy from existing dataset file")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Create interactively with prompts")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: evals/datasets/<name>.yaml)")

	return cmd
}

func createDataset(name, templateName, fromExample string, interactive bool, outputPath string) {
	var dataset *evaluation.Dataset
	var err error

	// Determine creation mode
	if fromExample != "" {
		// Copy from example
		dataset, err = evaluation.LoadDataset(fromExample)
		if err != nil {
			color.Red("Error loading example dataset: %v\n", err)
			os.Exit(1)
		}
		if name != "" {
			dataset.Name = name
		}
		color.Green("✅ Loaded dataset from example: %s\n", fromExample)

	} else if templateName != "" {
		// Use built-in template
		dataset = getTemplateDataset(templateName)
		if dataset == nil {
			color.Red("Unknown template: %s\n", templateName)
			fmt.Println("\nAvailable templates: baseline, medical-qa, technical-docs, simple-qa, multi-collection")
			os.Exit(1)
		}
		if name != "" {
			dataset.Name = name
		}
		color.Green("✅ Using template: %s\n", templateName)

	} else {
		// Interactive or minimal mode
		if name == "" {
			name = promptString("Dataset name", "my-dataset")
		}
		dataset = createInteractiveDataset(name, interactive)
	}

	// Validate dataset
	if err := dataset.Validate(); err != nil {
		color.Red("Dataset validation failed: %v\n", err)
		os.Exit(1)
	}

	// Determine output path
	if outputPath == "" {
		datasetDir := evaluation.GetDefaultDatasetDir()
		os.MkdirAll(datasetDir, 0755)
		outputPath = filepath.Join(datasetDir, dataset.Name+".yaml")
	}

	// Save dataset
	if err := evaluation.SaveDataset(dataset, outputPath); err != nil {
		color.Red("Error saving dataset: %v\n", err)
		os.Exit(1)
	}

	color.Green("\n✅ Dataset created successfully!\n\n")
	fmt.Printf("Name:       %s\n", dataset.Name)
	fmt.Printf("Version:    %s\n", dataset.Version)
	fmt.Printf("Test cases: %d\n", len(dataset.TestCases))
	fmt.Printf("File:       %s\n", outputPath)
	fmt.Println("\nNext steps:")
	fmt.Printf("  • Review: weave eval datasets show %s\n", dataset.Name)
	fmt.Printf("  • Validate: weave eval datasets validate %s\n", outputPath)
	fmt.Printf("  • Run: weave eval run --agent <agent-name> --dataset %s\n", dataset.Name)
}

func getTemplateDataset(templateName string) *evaluation.Dataset {
	templatesDir := "evals/datasets"
	templateFile := filepath.Join(templatesDir, templateName+".yaml")

	// Try to load from file
	dataset, err := evaluation.LoadDataset(templateFile)
	if err == nil {
		return dataset
	}

	// Return nil if template not found
	return nil
}

func createInteractiveDataset(name string, fullInteractive bool) *evaluation.Dataset {
	dataset := &evaluation.Dataset{
		Name:    name,
		Version: "1.0.0",
	}

	if fullInteractive {
		// Full interactive mode with detailed prompts
		color.Cyan("\n=== Creating Dataset: %s ===\n\n", name)

		dataset.Description = promptString("Description", "Evaluation dataset for "+name)
		dataset.Author = promptString("Author (optional)", "")

		// Config
		fmt.Println("\nConfiguration:")
		dataset.Config.DefaultAgent = promptString("Default agent (optional)", "rag-agent")
		dataset.Config.DefaultCollection = promptString("Default collection (optional)", "")
		dataset.Config.MinAccuracyScore = promptFloat("Min accuracy score (0.0-1.0)", 0.7)
		dataset.Config.MinCitationScore = promptFloat("Min citation score (0.0-1.0)", 0.8)
		dataset.Config.AllowHallucination = promptBool("Allow hallucination", false)

		// Test cases
		fmt.Println("\nTest Cases:")
		numCases := promptInt("Number of test cases to add", 1)

		for i := 0; i < numCases; i++ {
			fmt.Printf("\n--- Test Case %d ---\n", i+1)
			tc := evaluation.TestCase{
				ID:                fmt.Sprintf("test-%03d", i+1),
				Query:             promptString("Query", ""),
				ExpectedAnswer:    promptString("Expected answer", ""),
				MinRelevanceScore: 0.7,
			}

			if promptBool("Add required concepts", false) {
				conceptsStr := promptString("Required concepts (comma-separated)", "")
				if conceptsStr != "" {
					tc.RequiredConcepts = strings.Split(conceptsStr, ",")
					for j := range tc.RequiredConcepts {
						tc.RequiredConcepts[j] = strings.TrimSpace(tc.RequiredConcepts[j])
					}
				}
			}

			tc.MustCite = promptBool("Must cite sources", false)

			dataset.TestCases = append(dataset.TestCases, tc)
		}
	} else {
		// Minimal mode - create basic structure
		dataset.Description = "Evaluation dataset for " + name
		dataset.Config = evaluation.DatasetConfig{
			MinAccuracyScore:  0.7,
			MinCitationScore:  0.8,
			AllowHallucination: false,
		}

		// Add one example test case
		dataset.TestCases = []evaluation.TestCase{
			{
				ID:                "test-001",
				Description:       "Example test case - replace with your own",
				Query:             "What is your test question?",
				ExpectedAnswer:    "Expected answer goes here",
				RequiredConcepts:  []string{"concept1", "concept2"},
				MinRelevanceScore: 0.7,
				MustCite:          false,
			},
		}

		fmt.Println("\nCreated minimal dataset with 1 example test case.")
		fmt.Println("Edit the file to add your own test cases.")
	}

	return dataset
}

func promptString(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func promptInt(prompt string, defaultValue int) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%d]: ", prompt, defaultValue)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	var value int
	fmt.Sscanf(input, "%d", &value)
	return value
}

func promptFloat(prompt string, defaultValue float64) float64 {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%.2f]: ", prompt, defaultValue)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}

	var value float64
	fmt.Sscanf(input, "%f", &value)
	return value
}

func promptBool(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)

	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}
