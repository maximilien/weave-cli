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

Output format:
- Single VDB: Detailed view by default (use --summary for table)
- Multiple VDBs: Summary table by default (use --details for full output)

Filtering options:
- Use --cloud to check only cloud databases
- Use --local to check only local databases

If no database name is provided, it checks the default database.
Use 'weave config list' to see all available databases.`,
	Run: runHealthCheck,
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.GroupID = "config"
	healthCmd.AddCommand(healthCheckCmd)

	// Add flags for output format
	healthCheckCmd.Flags().BoolP("summary", "S", false, "Force summary table output (default for multiple VDBs)")
	healthCheckCmd.Flags().Bool("details", false, "Force detailed output (default for single VDB)")

	// Add flags for filtering by deployment type
	healthCheckCmd.Flags().Bool("cloud", false, "Check only cloud databases")
	healthCheckCmd.Flags().Bool("local", false, "Check only local databases")

	// Add flag for sorting
	healthCheckCmd.Flags().String("sort-by", "name", "Sort by: name or type")
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	summaryOutput, _ := cmd.Flags().GetBool("summary")
	detailsOutput, _ := cmd.Flags().GetBool("details")
	cloudOnly, _ := cmd.Flags().GetBool("cloud")
	localOnly, _ := cmd.Flags().GetBool("local")
	sortBy, _ := cmd.Flags().GetString("sort-by")

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
		} else if summaryOutput {
			displayHealthCheckSummary([]HealthCheckResult{result}, sortBy)
		} else {
			// Single VDB: default to detailed view
			printHeader(fmt.Sprintf("Database Health Check: %s", dbName))
			fmt.Println()
			displayHealthCheckResult(result)
		}
		return
	}

	// When --cloud or --local is used, query all configured databases
	// so we can filter by deployment type, not just the default
	var selection *utils.VectorDBSelection
	if cloudOnly || localOnly {
		selection, err = utils.GetAllConfiguredDatabases(cfg)
	} else {
		selection, err = utils.GetSelectedVectorDBs(cmd, cfg)
	}
	if err != nil {
		printError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	// Filter databases based on cloud/local flags
	var filteredConfigs []config.VectorDBConfig
	for _, dbConfig := range selection.Configs {
		if cloudOnly || localOnly {
			isCloud := isCloudDatabase(dbConfig.Type)
			if (cloudOnly && !isCloud) || (localOnly && isCloud) {
				continue // Skip this database
			}
		}
		filteredConfigs = append(filteredConfigs, dbConfig)
	}

	// Check if filtering resulted in no databases
	if len(filteredConfigs) == 0 {
		if cloudOnly {
			printWarning("No cloud databases found in selection")
		} else if localOnly {
			printWarning("No local databases found in selection")
		} else {
			printWarning("No databases found")
		}
		return
	}

	// For JSON output, collect all results first
	if jsonOutput {
		var results []HealthCheckResult
		for _, dbConfig := range filteredConfigs {
			result := checkSingleDatabase(ctx, dbConfig.Name, &dbConfig)
			results = append(results, result)
		}
		outputHealthCheckJSON(results)
		return
	}

	// For single database, show detailed view (unless --summary specified)
	if len(filteredConfigs) == 1 && !summaryOutput {
		result := checkSingleDatabase(ctx, filteredConfigs[0].Name, &filteredConfigs[0])
		printHeader(fmt.Sprintf("%s Database Health Check", result.DatabaseName))
		fmt.Println()
		displayHealthCheckResult(result)
		return
	}

	// For multiple databases with --details, show detailed view for each
	if detailsOutput {
		printHeader("Multiple Database Health Check")
		fmt.Println()
		for i, dbConfig := range filteredConfigs {
			if i > 0 {
				fmt.Println()
				fmt.Println("─────────────────────────────────────────")
				fmt.Println()
			}
			result := checkSingleDatabase(ctx, dbConfig.Name, &dbConfig)
			displayHealthCheckResult(result)
		}
		return
	}

	// For summary table: print header and rows progressively
	displayHealthCheckSummaryProgressive(ctx, filteredConfigs, sortBy)
}

func checkSingleDatabase(ctx context.Context, dbName string, dbConfig *config.VectorDBConfig) HealthCheckResult {
	result := HealthCheckResult{
		DatabaseName: dbName,
		DatabaseType: string(dbConfig.Type),
		URL:          getDisplayURL(dbConfig),
	}

	// Use vectordb abstraction for all database types
	vdbConfig := &vectordb.Config{
		Type:             vectordb.VectorDBType(dbConfig.Type),
		URL:              dbConfig.URL,
		APIKey:           dbConfig.APIKey,
		OpenAIAPIKey:     dbConfig.OpenAIAPIKey,
		DatabaseURL:      dbConfig.DatabaseURL,
		DatabaseKey:      dbConfig.DatabaseKey,
		Database:         dbConfig.Database,
		Username:         dbConfig.Username,
		Password:         dbConfig.Password,
		Address:          dbConfig.Address,
		Tenant:           dbConfig.Tenant,
		VectorDimensions: dbConfig.VectorDimensions,
		SimilarityMetric: dbConfig.SimilarityMetric,
		Timeout:          dbConfig.Timeout,
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

func displayHealthCheckSummary(results []HealthCheckResult, sortBy string) {
	// Print header
	printHeader("Vector Database Health Check Summary")
	fmt.Println()

	// Table header
	fmt.Printf("%-15s %-15s %-8s %-6s %s\n", "VDB", "TYPE", "STATUS", "COLS", "URL")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")

	// Sort results based on user preference
	switch sortBy {
	case "type":
		// Sort by type, then by name
		sort.Slice(results, func(i, j int) bool {
			if results[i].DatabaseType != results[j].DatabaseType {
				return results[i].DatabaseType < results[j].DatabaseType
			}
			return results[i].DatabaseName < results[j].DatabaseName
		})
	default: // "name" or anything else
		// Sort alphabetically by name only
		sort.Slice(results, func(i, j int) bool {
			return results[i].DatabaseName < results[j].DatabaseName
		})
	}

	// Count healthy/unhealthy
	healthy := 0
	unhealthy := 0

	// Print each result
	for _, result := range results {
		status := "✗ FAIL"
		statusColor := color.New(color.FgRed)
		if result.Healthy {
			status = "✓ OK"
			statusColor = color.New(color.FgGreen)
			healthy++
		} else {
			unhealthy++
		}

		// Truncate URL if too long (more space now with shorter VDB column)
		url := result.URL
		if len(url) > 49 {
			url = url[:46] + "..."
		}

		// Print row
		fmt.Printf("%-15s %-15s ", result.DatabaseName, result.DatabaseType)
		statusColor.Printf("%-8s ", status)
		fmt.Printf("%-6d %s\n", result.CollectionsCount, url)
	}

	// Print footer summary
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d databases  |  ", len(results))
	if healthy > 0 {
		color.New(color.FgGreen).Printf("Healthy: %d", healthy)
	}
	if unhealthy > 0 {
		if healthy > 0 {
			fmt.Print("  |  ")
		}
		color.New(color.FgRed).Printf("Unhealthy: %d", unhealthy)
	}
	fmt.Println()
}

// displayHealthCheckSummaryProgressive shows health check results progressively as they're checked
func displayHealthCheckSummaryProgressive(ctx context.Context, dbConfigs []config.VectorDBConfig, sortBy string) {
	// Print header
	printHeader("Vector Database Health Check Summary")
	fmt.Println()

	// Table header
	fmt.Printf("%-15s %-15s %-8s %-6s %s\n", "VDB", "TYPE", "STATUS", "COLS", "URL")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")

	// Sort databases based on user preference
	switch sortBy {
	case "type":
		// Sort by type, then by name
		sort.Slice(dbConfigs, func(i, j int) bool {
			if dbConfigs[i].Type != dbConfigs[j].Type {
				return dbConfigs[i].Type < dbConfigs[j].Type
			}
			return dbConfigs[i].Name < dbConfigs[j].Name
		})
	default: // "name" or anything else
		// Sort alphabetically by name only
		sort.Slice(dbConfigs, func(i, j int) bool {
			return dbConfigs[i].Name < dbConfigs[j].Name
		})
	}

	// Count healthy/unhealthy
	healthy := 0
	unhealthy := 0

	// Check and print each database progressively
	for _, dbConfig := range dbConfigs {
		result := checkSingleDatabase(ctx, dbConfig.Name, &dbConfig)

		status := "✗ FAIL"
		statusColor := color.New(color.FgRed)
		if result.Healthy {
			status = "✓ OK"
			statusColor = color.New(color.FgGreen)
			healthy++
		} else {
			unhealthy++
		}

		// Truncate URL if too long
		url := result.URL
		if len(url) > 49 {
			url = url[:46] + "..."
		}

		// Print row immediately
		fmt.Printf("%-15s %-15s ", result.DatabaseName, result.DatabaseType)
		statusColor.Printf("%-8s ", status)
		fmt.Printf("%-6d %s\n", result.CollectionsCount, url)
	}

	// Print footer summary
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d databases  |  ", len(dbConfigs))
	if healthy > 0 {
		color.New(color.FgGreen).Printf("Healthy: %d", healthy)
	}
	if unhealthy > 0 {
		if healthy > 0 {
			fmt.Print("  |  ")
		}
		color.New(color.FgRed).Printf("Unhealthy: %d", unhealthy)
	}
	fmt.Println()
}

// isCloudDatabase checks if a database type is a cloud deployment
func isCloudDatabase(dbType config.VectorDBType) bool {
	// Cloud suffix patterns
	cloudTypes := []string{"cloud", "aura", "zilliz"}
	typeStr := string(dbType)
	for _, ct := range cloudTypes {
		if len(typeStr) > len(ct) && typeStr[len(typeStr)-len(ct):] == ct {
			return true
		}
	}

	// Legacy support: MongoDB and Supabase without -cloud suffix are still cloud services
	if dbType == "mongodb" || dbType == "supabase" {
		return true
	}

	return false
}
