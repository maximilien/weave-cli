// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vdb

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// InfoCmd represents the vdb info command
var InfoCmd = &cobra.Command{
	Use:   "info DATABASE_NAME",
	Short: "Show detailed database information",
	Long: `Show detailed information about a specific vector database configuration.

Displays:
- Database name and type
- Connection endpoint (URL/address)
- Authentication status
- Timeout configuration
- Vector settings (dimensions, similarity metric)
- Configured collections

Examples:
  weave vdb info weaviate-cloud
  weave vdb info milvus-local
  weave vdb info qdrant-cloud`,
	Args: cobra.ExactArgs(1),
	Run:  runVDBInfo,
}

func init() {
	VDBCmd.AddCommand(InfoCmd)
}

func runVDBInfo(cmd *cobra.Command, args []string) {
	dbName := args[0]

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		os.Exit(1)
	}

	// Find database
	var dbConfig *config.VectorDBConfig
	for _, db := range cfg.Databases.VectorDatabases {
		if db.Name == dbName {
			dbCopy := db
			dbConfig = &dbCopy
			break
		}
	}

	if dbConfig == nil {
		utils.PrintError(fmt.Sprintf("Database '%s' not found in configuration", dbName))
		fmt.Println("\nAvailable databases:")
		for _, db := range cfg.Databases.VectorDatabases {
			fmt.Printf("  • %s\n", db.Name)
		}
		os.Exit(1)
	}

	// Print database information
	fmt.Printf("\n")
	color.New(color.FgCyan, color.Bold).Printf("📊 Database Information: %s\n", dbName)
	fmt.Printf("\n")

	// Basic info
	printSection("Basic Information")
	printField("Name", dbConfig.Name)
	printField("Type", string(dbConfig.Type))
	isDefault := dbConfig.Name == cfg.Databases.Default
	if isDefault {
		printField("Default", "Yes ✓")
	} else {
		printField("Default", "No")
	}

	// Connection info
	fmt.Println()
	printSection("Connection")
	endpoint := getEndpoint(*dbConfig)
	printField("Endpoint", endpoint)

	// Authentication
	hasAuth := dbConfig.APIKey != "" || dbConfig.Username != "" || dbConfig.Password != ""
	if hasAuth {
		printField("Authentication", "Configured ✓")
		if dbConfig.APIKey != "" {
			printField("API Key", maskSecret(dbConfig.APIKey))
		}
		if dbConfig.Username != "" {
			printField("Username", dbConfig.Username)
		}
	} else {
		printField("Authentication", "None")
	}

	// Timeout
	if dbConfig.Timeout > 0 {
		printField("Timeout", fmt.Sprintf("%ds", dbConfig.Timeout))
	}

	// Vector settings
	if dbConfig.VectorDimensions > 0 || dbConfig.SimilarityMetric != "" {
		fmt.Println()
		printSection("Vector Configuration")
		if dbConfig.VectorDimensions > 0 {
			printField("Vector Dimensions", fmt.Sprintf("%d", dbConfig.VectorDimensions))
		}
		if dbConfig.SimilarityMetric != "" {
			printField("Similarity Metric", dbConfig.SimilarityMetric)
		}
		if dbConfig.EmbeddingDimension > 0 {
			printField("Embedding Dimension", fmt.Sprintf("%d", dbConfig.EmbeddingDimension))
		}
	}

	// Collections
	if len(dbConfig.Collections) > 0 {
		fmt.Println()
		printSection(fmt.Sprintf("Collections (%d)", len(dbConfig.Collections)))
		for _, col := range dbConfig.Collections {
			desc := col.Description
			if desc == "" {
				desc = "No description"
			}
			fmt.Printf("  • %s (%s) - %s\n",
				color.CyanString(col.Name),
				col.Type,
				desc)
		}
	}

	// Additional settings
	if dbConfig.Database != "" || dbConfig.Tenant != "" {
		fmt.Println()
		printSection("Additional Settings")
		if dbConfig.Database != "" {
			printField("Database", dbConfig.Database)
		}
		if dbConfig.Tenant != "" {
			printField("Tenant", dbConfig.Tenant)
		}
	}

	fmt.Println()

	// Helpful tips
	fmt.Println("💡 Next steps:")
	fmt.Printf("   • Check health: weave health check %s\n", dbName)
	fmt.Printf("   • List collections: weave cols ls --%s\n", dbConfig.Type)
	fmt.Println("   • View all databases: weave vdb list")

	fmt.Println()
}

func printSection(title string) {
	color.New(color.FgYellow, color.Bold).Printf("▸ %s\n", title)
}

func printField(key, value string) {
	fmt.Printf("  %-20s %s\n", color.New(color.Faint).Sprint(key+":"), value)
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	// Show first 4 and last 4 characters
	return s[:4] + "..." + s[len(s)-4:]
}
