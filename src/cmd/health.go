// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	DatabaseName     string   `json:"database_name"`
	DatabaseType     string   `json:"database_type"`
	Healthy          bool     `json:"healthy"`
	Message          string   `json:"message"`
	URL              string   `json:"url,omitempty"`
	Collections      []string `json:"collections,omitempty"`
	CollectionsCount int      `json:"collections_count"`
}

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:        "health",
	Short:      "Health and connectivity management",
	SuggestFor: []string{"check", "status", "ping"},
	Long: `Manage database health and connectivity.

This command provides subcommands to check database health and connectivity.`,
}

// healthCheckCmd represents the health check command
var healthCheckCmd = &cobra.Command{
	Use:     "check [database-name]",
	Aliases: []string{"c"},
	Short:   "Check health of database connections",
	Long: `Check the health of the configured vector database connections.

This command:
- Attempts to connect to the configured database (or all databases)
- Verifies API keys and authentication
- Tests collection access
- Reports connection status and any issues

If no database name is provided, it checks the default database.
Use 'weave config list' to see all available databases.`,
	Run: runHealthCheck,
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.GroupID = "config"
	healthCmd.AddCommand(healthCheckCmd)
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load configuration
	cfg, err := LoadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If a specific database name is provided as argument, check only that
	if len(args) > 0 {
		dbName := args[0]
		dbConfig, err := cfg.GetDatabase(dbName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", dbName, err))
			os.Exit(1)
		}

		result := checkSingleDatabase(ctx, dbName, dbConfig)
		if jsonOutput {
			outputHealthCheckJSON([]HealthCheckResult{result})
		} else {
			printHeader(fmt.Sprintf("Database Health Check: %s", dbName))
			fmt.Println()
			displayHealthCheckResult(result)
		}
		return
	}

	// Get selected databases based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		printError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	// Collect results from all databases
	var results []HealthCheckResult
	for _, dbConfig := range selection.Configs {
		result := checkSingleDatabase(ctx, dbConfig.Name, &dbConfig)
		results = append(results, result)
	}

	// Output results
	if jsonOutput {
		outputHealthCheckJSON(results)
	} else {
		if len(results) == 1 {
			printHeader(fmt.Sprintf("%s Database Health Check", results[0].DatabaseName))
			fmt.Println()
			displayHealthCheckResult(results[0])
		} else {
			printHeader("Multiple Database Health Check")
			fmt.Println()
			for i, result := range results {
				if i > 0 {
					fmt.Println()
					fmt.Println("─────────────────────────────────────────")
					fmt.Println()
				}
				displayHealthCheckResult(result)
			}
		}
	}
}

func checkSingleDatabase(ctx context.Context, dbName string, dbConfig *config.VectorDBConfig) HealthCheckResult {
	result := HealthCheckResult{
		DatabaseName: dbName,
		DatabaseType: string(dbConfig.Type),
		URL:          getDisplayURL(dbConfig),
	}

	// Use vectordb abstraction for all database types
	vdbConfig := &vectordb.Config{
		Type:         vectordb.VectorDBType(dbConfig.Type),
		URL:          dbConfig.URL,
		APIKey:       dbConfig.APIKey,
		OpenAIAPIKey: dbConfig.OpenAIAPIKey,
		DatabaseURL:  dbConfig.DatabaseURL,
		DatabaseKey:  dbConfig.DatabaseKey,
		Timeout:      dbConfig.Timeout,
	}

	client, err := vectordb.CreateClient(vdbConfig)
	if err != nil {
		result.Healthy = false
		result.Message = fmt.Sprintf("Failed to create database client: %v", err)
		return result
	}

	// Test connection
	if err := client.Health(ctx); err != nil {
		result.Healthy = false
		result.Message = fmt.Sprintf("Health check failed: %v", err)
		return result
	}

	result.Healthy = true
	result.Message = fmt.Sprintf("Successfully connected to %s at %s", dbConfig.Type, getDisplayURL(dbConfig))

	// Test collection access
	collectionsInfo, err := client.ListCollections(ctx)
	if err == nil {
		for _, info := range collectionsInfo {
			result.Collections = append(result.Collections, info.Name)
		}
		sort.Strings(result.Collections)
		result.CollectionsCount = len(result.Collections)
	}

	return result
}

func getDisplayURL(dbConfig *config.VectorDBConfig) string {
	if dbConfig.URL != "" {
		return dbConfig.URL
	}
	if dbConfig.DatabaseURL != "" {
		return dbConfig.DatabaseURL
	}
	return "configured location"
}

func displayHealthCheckResult(result HealthCheckResult) {
	color.New(color.FgCyan, color.Bold).Printf("Checking %s database (%s)...\n", result.DatabaseName, result.DatabaseType)
	fmt.Println()
	fmt.Println()

	if result.Healthy {
		printSuccess("Database connection is healthy!")
		color.New(color.FgGreen).Printf("✅ %s\n", result.Message)
	} else {
		printError("Database connection failed!")
		color.New(color.FgRed).Printf("❌ %s\n", result.Message)
		return // Don't show collection info if unhealthy
	}

	// Show collection access test results
	fmt.Println()
	printHeader("Collection Access Test")

	if result.CollectionsCount == 0 {
		printWarning("No collections found in the database")
	} else {
		printSuccess(fmt.Sprintf("Found %d collections:", result.CollectionsCount))
		for _, name := range result.Collections {
			fmt.Printf("  - %s\n", name)
		}
	}
}

func outputHealthCheckJSON(results []HealthCheckResult) {
	output := map[string]interface{}{
		"results": results,
		"total":   len(results),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal JSON: %v", err))
		return
	}

	fmt.Println(string(jsonBytes))
}
