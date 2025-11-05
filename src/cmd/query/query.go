// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package query

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/maximilien/weave-cli/src/pkg/executor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun       bool
	noConfirm    bool
	outputFormat string
	model        string
	maxRetries   int
)

// NewQueryCommand creates the query command
func NewQueryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query [query-text]",
		Aliases: []string{"q"},
		Short:   "Execute natural language queries against weave-cli",
		Long: `Execute natural language queries against weave-cli using AI agents.

The query command uses AI agents to:
1. Understand your natural language query
2. Plan the appropriate weave-cli and bash commands
3. Execute the plan and provide comprehensive results

Examples:
  weave query "show me all my collections"
  weave q "find all empty collections"
  weave query "add files in /tmp/docs to MyCollection"
  weave q "convert CMYK PDFs in ./pdfs to RGB"

Flags:
  --dry-run              Show the execution plan without running it
  --no-confirm           Skip confirmation for destructive operations
  --verbose, -v          Show detailed execution information
  --output, -o string    Output format: text (default), json, yaml
  --model string         LLM model to use (default: from env OPENAI_MODEL or gpt-4o)
  --max-retries int      Maximum retries for failed operations (default: 3)`,
		Args: cobra.ExactArgs(1),
		RunE: runQuery,
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show execution plan without running it")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation for destructive operations")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")
	cmd.Flags().StringVar(&model, "model", "", "LLM model to use")
	cmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum retries for failed operations")

	return cmd
}

// runQuery executes the query command
func runQuery(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Load .env file if it exists
	envFile := viper.GetString("env")
	if envFile != "" {
		_ = godotenv.Load(envFile)
	} else {
		_ = godotenv.Load() // Try loading .env from current directory
	}

	// Get configuration from environment and flags
	config := &executor.Config{
		DryRun:       dryRun,
		NoConfirm:    noConfirm,
		Verbose:      viper.GetBool("verbose"),
		Quiet:        viper.GetBool("quiet"),
		NoColor:      viper.GetBool("no-color"),
		OutputFormat: outputFormat,
		Model:        getModel(),
		StdioPath:    os.Getenv("WEAVE_MCP_STDIO_PATH"),
		MaxRetries:   maxRetries,
	}

	// Create executor
	exec, err := executor.NewExecutor(config)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer func() {
		if closeErr := exec.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close executor: %v\n", closeErr)
		}
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute query
	if dryRun {
		_, err = exec.DryRun(ctx, query)
	} else {
		_, err = exec.Execute(ctx, query)
	}

	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	return nil
}

// getModel returns the model to use, with priority: flag > env > default
func getModel() string {
	if model != "" {
		return model
	}

	envModel := os.Getenv("OPENAI_MODEL")
	if envModel != "" {
		return envModel
	}

	return "gpt-4o"
}
