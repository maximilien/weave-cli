// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/evaluation"
	"github.com/spf13/cobra"
)

// NewListEvaluatorsCommand creates a command to list available custom evaluators
func NewListEvaluatorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-evaluators",
		Short: "List available custom evaluators",
		Long: `List all custom evaluators defined in the evaluators directory.

Custom evaluators allow you to define domain-specific evaluation metrics
using YAML configuration files.`,
		RunE: runListEvaluators,
	}

	return cmd
}

// NewValidateEvaluatorCommand creates a command to validate an evaluator definition
func NewValidateEvaluatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-evaluator <file>",
		Short: "Validate a custom evaluator definition",
		Long: `Validate a custom evaluator YAML file for correctness.

Checks:
- Required fields present
- Valid scoring type
- Threshold in valid range (0.0-1.0)
- Regex pattern valid (for regex type)`,
		Args: cobra.ExactArgs(1),
		RunE: runValidateEvaluator,
	}

	return cmd
}

// NewCreateEvaluatorCommand creates a command to create a new evaluator from template
func NewCreateEvaluatorCommand() *cobra.Command {
	var scoringType string

	cmd := &cobra.Command{
		Use:   "create-evaluator <name>",
		Short: "Create a new custom evaluator from template",
		Long: `Create a new custom evaluator YAML file from a template.

The evaluator will be created in evals/evaluators/<name>.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateEvaluator(args[0], scoringType)
		},
	}

	cmd.Flags().StringVar(&scoringType, "type", "llm_judge", "Scoring type: llm_judge, regex, exact_match, contains")

	return cmd
}

// runListEvaluators lists all available custom evaluators
func runListEvaluators(cmd *cobra.Command, args []string) error {
	evalDir := evaluation.GetDefaultCustomEvaluatorDir()

	color.Cyan("📋 Custom Evaluators\n")
	fmt.Printf("Directory: %s\n\n", evalDir)

	loader := evaluation.NewCustomEvaluatorLoader(evalDir)
	evaluators, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load evaluators: %w", err)
	}

	if len(evaluators) == 0 {
		color.Yellow("No custom evaluators found.\n")
		fmt.Printf("\nCreate your first evaluator:\n")
		fmt.Printf("  weave eval create-evaluator my_evaluator\n\n")
		fmt.Printf("See examples in: %s/examples/\n", evalDir)
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tTHRESHOLD\tDESCRIPTION")
	fmt.Fprintln(w, "────\t────\t─────────\t───────────")

	for _, eval := range evaluators {
		threshold := fmt.Sprintf("%.2f", eval.Scoring.Threshold)
		description := eval.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			eval.Name,
			eval.Scoring.Type,
			threshold,
			description,
		)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d evaluators\n", len(evaluators))
	fmt.Printf("\nUse in dataset:\n")
	fmt.Printf("  custom_evaluators:\n")
	fmt.Printf("    - %s\n", evaluators[0].Name)

	return nil
}

// runValidateEvaluator validates an evaluator definition
func runValidateEvaluator(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	color.Cyan("🔍 Validating evaluator: %s\n\n", filePath)

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		color.Red("✗ File not found: %s\n", filePath)
		return fmt.Errorf("file not found")
	}

	// Load evaluator definition
	data, err := os.ReadFile(filePath)
	if err != nil {
		color.Red("✗ Failed to read file: %v\n", err)
		return err
	}

	// Parse YAML
	var def evaluation.CustomEvaluatorDef
	if err := evaluation.UnmarshalYAML(data, &def); err != nil {
		color.Red("✗ Invalid YAML: %v\n", err)
		return err
	}

	// Validate definition
	if err := def.Validate(); err != nil {
		color.Red("✗ Validation failed: %v\n", err)
		return err
	}

	// Success!
	color.Green("✓ Evaluator definition is valid\n\n")

	// Print summary
	fmt.Printf("Name:        %s\n", def.Name)
	fmt.Printf("Description: %s\n", def.Description)
	fmt.Printf("Type:        %s\n", def.Scoring.Type)
	fmt.Printf("Threshold:   %.2f\n", def.Scoring.Threshold)

	if def.Scoring.Type == "regex" {
		fmt.Printf("Pattern:     %s\n", def.Scoring.Pattern)
	}

	if len(def.Tags) > 0 {
		fmt.Printf("Tags:        %v\n", def.Tags)
	}

	return nil
}

// runCreateEvaluator creates a new evaluator from template
func runCreateEvaluator(name string, scoringType string) error {
	evalDir := evaluation.GetDefaultCustomEvaluatorDir()

	// Ensure directory exists
	if err := os.MkdirAll(evalDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if evaluator already exists
	filePath := filepath.Join(evalDir, name+".yaml")
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("evaluator already exists: %s", filePath)
	}

	// Generate template based on type
	template := generateEvaluatorTemplate(name, scoringType)

	// Write file
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	color.Green("✓ Created evaluator: %s\n\n", filePath)
	fmt.Printf("Next steps:\n")
	fmt.Printf("1. Edit the evaluator: %s\n", filePath)
	fmt.Printf("2. Validate it: weave eval validate-evaluator %s\n", filePath)
	fmt.Printf("3. Reference in dataset:\n")
	fmt.Printf("   custom_evaluators:\n")
	fmt.Printf("     - %s\n", name)

	return nil
}

// generateEvaluatorTemplate generates an evaluator template based on type
func generateEvaluatorTemplate(name string, scoringType string) string {
	switch scoringType {
	case "regex":
		return fmt.Sprintf(`name: %s
description: Describe what this evaluator checks
version: 1.0.0

# Regex evaluators match patterns in responses
prompt: "Not used for regex evaluators"

scoring:
  type: regex
  pattern: '\b(keyword|phrase)\b'  # Update this pattern
  threshold: 0.7

tags:
  - regex
  - custom

author: your-team
required: false
`, name)

	case "exact_match":
		return fmt.Sprintf(`name: %s
description: Requires exact match with expected answer
version: 1.0.0

# Exact match evaluators check for exact string equality
prompt: "Not used for exact match evaluators"

scoring:
  type: exact_match
  threshold: 0.9

tags:
  - exact-match
  - strict

author: your-team
required: false
`, name)

	case "contains":
		return fmt.Sprintf(`name: %s
description: Checks if response contains expected terms
version: 1.0.0

# Contains evaluators check for substring presence
prompt: "Not used for contains evaluators"

scoring:
  type: contains
  threshold: 0.5

tags:
  - contains
  - flexible

author: your-team
required: false
`, name)

	default: // llm_judge
		return fmt.Sprintf(`name: %s
description: Describe what this evaluator checks
version: 1.0.0

prompt: |
  You are evaluating a response for quality.

  Query: {{.Query}}
  Expected Answer: {{.ExpectedAnswer}}
  Actual Answer: {{.Answer}}
  Context: {{.Context}}

  Evaluate the response considering:
  1. [Add your criteria here]
  2. [Add more criteria]
  3. [And more...]

  Rate from 0.0 (poor) to 1.0 (excellent).
  Provide only a number between 0.0 and 1.0.

scoring:
  type: llm_judge
  threshold: 0.7
  weight: 1.0

# Optional: Override default LLM model
# model:
#   provider: openai
#   name: gpt-4o-mini
#   temperature: 0.1

tags:
  - quality
  - custom

author: your-team
required: false
`, name)
	}
}
