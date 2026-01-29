// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vdb

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ListCmd represents the vdb list command
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured vector databases",
	Long: `List all configured vector databases with their types and connection status.

Shows:
- Database name
- Type (weaviate-cloud, milvus-local, etc.)
- URL/Address
- Status indicator

Examples:
  weave vdb list              # List all databases
  weave vdb ls                # Same (alias)
  weave vdb list --cloud      # List only cloud databases
  weave vdb list --local      # List only local databases`,
	Run: runVDBList,
}

var (
	cloudOnly bool
	localOnly bool
)

func init() {
	VDBCmd.AddCommand(ListCmd)

	ListCmd.Flags().BoolVar(&cloudOnly, "cloud", false, "show only cloud databases")
	ListCmd.Flags().BoolVar(&localOnly, "local", false, "show only local databases")
}

func runVDBList(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		os.Exit(1)
	}

	if cfg == nil || len(cfg.Databases.VectorDatabases) == 0 {
		utils.PrintWarning("No vector databases configured")
		fmt.Println("\nTo add a database:")
		fmt.Println("  1. Run: weave config create --env")
		fmt.Println("  2. Or edit config.yaml manually")
		os.Exit(1)
	}

	// Filter databases
	databases := cfg.Databases.VectorDatabases
	if cloudOnly {
		databases = filterCloudDatabases(databases)
	} else if localOnly {
		databases = filterLocalDatabases(databases)
	}

	if len(databases) == 0 {
		if cloudOnly {
			utils.PrintWarning("No cloud databases configured")
		} else if localOnly {
			utils.PrintWarning("No local databases configured")
		}
		os.Exit(0)
	}

	// Print header
	fmt.Printf("\n📊 Configured Vector Databases (%d total)\n\n", len(databases))

	// Print table
	fmt.Printf("%-20s %-25s %-40s %-10s\n", "NAME", "TYPE", "ENDPOINT", "DEFAULT")
	fmt.Printf("%s\n", "────────────────────────────────────────────────────────────────────────────────────────────")

	for _, db := range databases {
		endpoint := getEndpoint(db)
		isDefault := ""
		if db.Name == cfg.Databases.Default {
			isDefault = "✓"
		}

		fmt.Printf("%-20s %-25s %-40s %-10s\n",
			db.Name,
			db.Type,
			truncate(endpoint, 40),
			isDefault)
	}

	fmt.Println()

	// Show helpful tips
	fmt.Println("💡 Tips:")
	fmt.Println("   • Check health: weave health check")
	fmt.Println("   • Show details: weave vdb info <name>")
	fmt.Println("   • Filter: --cloud or --local")
}

func filterCloudDatabases(databases []config.VectorDBConfig) []config.VectorDBConfig {
	var filtered []config.VectorDBConfig
	for _, db := range databases {
		if isCloudType(string(db.Type)) {
			filtered = append(filtered, db)
		}
	}
	return filtered
}

func filterLocalDatabases(databases []config.VectorDBConfig) []config.VectorDBConfig {
	var filtered []config.VectorDBConfig
	for _, db := range databases {
		if !isCloudType(string(db.Type)) {
			filtered = append(filtered, db)
		}
	}
	return filtered
}

func isCloudType(dbType string) bool {
	cloudTypes := []string{
		"weaviate-cloud",
		"milvus-cloud",
		"mongodb-cloud",
		"supabase-cloud",
		"chroma-cloud",
		"qdrant-cloud",
		"neo4j-cloud",
		"opensearch-cloud",
		"pinecone",
	}
	for _, ct := range cloudTypes {
		if dbType == ct {
			return true
		}
	}
	return false
}

func getEndpoint(db config.VectorDBConfig) string {
	if db.URL != "" {
		return db.URL
	}
	if db.Address != "" {
		return db.Address
	}
	if db.DatabaseURL != "" {
		// Truncate database URLs (they can be long)
		return db.DatabaseURL
	}
	return "N/A"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
