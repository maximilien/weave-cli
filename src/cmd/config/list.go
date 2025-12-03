// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// listCmd represents the config list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all configured databases",
	Long: `List all configured vector databases.

By default, displays a summary table showing:
- Database names
- Types (local/cloud)
- Configuration status

Filtering options:
- Use --cloud to show only cloud databases
- Use --local to show only local databases
- Use --details for a detailed list view with all configuration information`,
	Run: runList,
}

func init() {
	listCmd.Flags().Bool("details", false, "Show detailed list view instead of table")
	listCmd.Flags().Bool("cloud", false, "Show only cloud databases")
	listCmd.Flags().Bool("local", false, "Show only local databases")
}

func runList(cobraCmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	databaseNames := cfg.GetDatabaseNames()
	if len(databaseNames) == 0 {
		printWarning("No databases configured")
		return
	}

	// Check flags
	details, _ := cobraCmd.Flags().GetBool("details")
	cloudOnly, _ := cobraCmd.Flags().GetBool("cloud")
	localOnly, _ := cobraCmd.Flags().GetBool("local")

	// Filter databases based on flags
	if cloudOnly || localOnly {
		filteredNames := make(map[string]config.VectorDBType)
		for name, dbType := range databaseNames {
			if name == "default" {
				continue
			}
			isCloud := isCloudDeployment(dbType)
			if (cloudOnly && isCloud) || (localOnly && !isCloud) {
				filteredNames[name] = dbType
			}
		}
		databaseNames = filteredNames

		if len(databaseNames) == 0 {
			if cloudOnly {
				printWarning("No cloud databases configured")
			} else {
				printWarning("No local databases configured")
			}
			return
		}
	}

	if details {
		// Show detailed list view with full configuration
		displayDatabasesDetails(cfg, databaseNames)
	} else {
		// Show table view (default)
		displayDatabasesTable(databaseNames)
	}
}

func displayDatabasesTable(databaseNames map[string]config.VectorDBType) {
	printHeader("Configured Vector Databases")
	fmt.Println()

	// Table header
	fmt.Printf("%-20s %-20s %-15s\n", "NAME", "TYPE", "DEPLOYMENT")
	fmt.Println("─────────────────────────────────────────────────────────")

	// Sort and categorize databases
	var localDBs []string
	var cloudDBs []string

	for name, dbType := range databaseNames {
		if name == "default" {
			continue // Skip default entry
		}

		// Determine if local or cloud based on type suffix
		if isCloudDeployment(dbType) {
			cloudDBs = append(cloudDBs, name)
		} else {
			localDBs = append(localDBs, name)
		}
	}

	// Sort for consistent output
	sort.Strings(localDBs)
	sort.Strings(cloudDBs)

	// Print local databases
	for _, name := range localDBs {
		dbType := string(databaseNames[name])
		fmt.Printf("%-20s %-20s %-15s\n", name, dbType, "Local")
	}

	// Print cloud databases
	for _, name := range cloudDBs {
		dbType := string(databaseNames[name])
		deployment := "Cloud"
		if false {
			deployment += " (default)"
		}
		fmt.Printf("%-20s %-20s %-15s\n", name, dbType, deployment)
	}

	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d databases  |  Local: %d  |  Cloud: %d\n",
		len(localDBs)+len(cloudDBs),
		len(localDBs),
		len(cloudDBs))
}

func isCloudDeployment(dbType config.VectorDBType) bool {
	// Cloud suffix patterns
	cloudTypes := []string{"cloud", "aura", "zilliz"}
	for _, ct := range cloudTypes {
		if len(string(dbType)) > len(ct) && string(dbType)[len(string(dbType))-len(ct):] == ct {
			return true
		}
	}

	// Legacy support: MongoDB Atlas and Supabase are cloud services
	// Old configs may still use "mongodb" and "supabase" without -cloud suffix
	if dbType == "mongodb" || dbType == "supabase" {
		return true
	}

	return false
}

func displayDatabasesDetails(cfg *config.Config, databaseNames map[string]config.VectorDBType) {
	printHeader("Configured Databases - Detailed View")
	fmt.Println()

	// Sort database names for consistent output
	var names []string
	for name := range databaseNames {
		if name != "default" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for i, name := range names {
		if i > 0 {
			fmt.Println()
			fmt.Println("─────────────────────────────────────────")
			fmt.Println()
		}

		dbConfig, err := cfg.GetDatabase(name)
		if err != nil {
			color.New(color.FgRed).Printf("• %s\n", name)
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		// Database name header
		color.New(color.FgCyan, color.Bold).Printf("• %s\n", name)

		// Type and deployment
		fmt.Printf("  Type: %s\n", dbConfig.Type)
		if isCloudDeployment(dbConfig.Type) {
			fmt.Printf("  Deployment: Cloud\n")
		} else {
			fmt.Printf("  Deployment: Local\n")
		}

		// Connection details (non-secret)
		if dbConfig.URL != "" {
			fmt.Printf("  URL: %s\n", dbConfig.URL)
		}
		if dbConfig.DatabaseURL != "" {
			fmt.Printf("  Database URL: %s\n", dbConfig.DatabaseURL)
		}
		if dbConfig.Address != "" {
			fmt.Printf("  Address: %s\n", dbConfig.Address)
		}
		if dbConfig.Database != "" {
			fmt.Printf("  Database: %s\n", dbConfig.Database)
		}
		if dbConfig.Tenant != "" {
			fmt.Printf("  Tenant: %s\n", dbConfig.Tenant)
		}
		if dbConfig.Username != "" {
			fmt.Printf("  Username: %s\n", dbConfig.Username)
		}

		// Vector configuration
		if dbConfig.VectorDimensions > 0 {
			fmt.Printf("  Vector Dimensions: %d\n", dbConfig.VectorDimensions)
		}
		if dbConfig.SimilarityMetric != "" {
			fmt.Printf("  Similarity Metric: %s\n", dbConfig.SimilarityMetric)
		}

		// Authentication status (masked secrets)
		if dbConfig.APIKey != "" {
			fmt.Printf("  API Key: %s\n", maskSecret(dbConfig.APIKey))
		}
		if dbConfig.DatabaseKey != "" {
			fmt.Printf("  Database Key: %s\n", maskSecret(dbConfig.DatabaseKey))
		}
		if dbConfig.Password != "" {
			fmt.Printf("  Password: %s\n", maskSecret(dbConfig.Password))
		}
		if dbConfig.OpenAIAPIKey != "" {
			fmt.Printf("  OpenAI API Key: %s\n", maskSecret(dbConfig.OpenAIAPIKey))
		}

		// Operational settings
		if dbConfig.Timeout > 0 {
			fmt.Printf("  Timeout: %ds\n", dbConfig.Timeout)
		}
		if dbConfig.Enabled {
			fmt.Printf("  Status: Enabled\n")
		}
		if dbConfig.SimulateEmbeddings {
			fmt.Printf("  Simulate Embeddings: true\n")
		}

		// Collections
		if len(dbConfig.Collections) > 0 {
			fmt.Printf("  Collections: %d configured\n", len(dbConfig.Collections))
			for _, coll := range dbConfig.Collections {
				fmt.Printf("    - %s (%s)\n", coll.Name, coll.Type)
			}
		}
	}
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	// Show first 4 and last 4 characters
	return secret[:4] + "..." + secret[len(secret)-4:]
}
