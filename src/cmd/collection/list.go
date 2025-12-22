// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
)

// ListCmd represents the collection list command
var ListCmd = &cobra.Command{
	Use:     "list [database-name]",
	Aliases: []string{"ls", "l"},
	Short:   "List all collections",
	Long: `List all collections in the configured vector database.

This command shows:
- Collection names
- Document counts (if available)
- Collection metadata (if available)

If no database name is provided, it uses the default database.
Use 'weave config list' to see all available databases.`,
	Run: runCollectionList,
}

func init() {
	CollectionCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntP("limit", "l", 100, "Maximum number of collections to show")
	ListCmd.Flags().BoolP("virtual", "", false, "Show collections in virtual structure")
	ListCmd.Flags().BoolP("summary", "S", false, "Show summary table (default for multiple VDBs)")
	ListCmd.Flags().Bool("details", false, "Show detailed list (overrides default summary for multiple VDBs)")
	ListCmd.Flags().StringP("output", "o", "text", "Output format: text, json, yaml")
}

func runCollectionList(cmd *cobra.Command, args []string) {
	limit, _ := cmd.Flags().GetInt("limit")
	virtual, _ := cmd.Flags().GetBool("virtual")
	outputFormat, _ := cmd.Flags().GetString("output")
	summaryOutput, _ := cmd.Flags().GetBool("summary")
	detailsOutput, _ := cmd.Flags().GetBool("details")

	// Support legacy --json flag for backward compatibility
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		outputFormat = "json"
	}
	jsonOutput := (outputFormat == "json")

	// Load configuration with interactive help
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	ctx := context.Background()

	// Check if a specific database name is provided (legacy behavior)
	if len(args) > 0 {
		// Use specified database (legacy behavior)
		dbConfig, err := cfg.GetDatabase(args[0])
		if err != nil {
			utils.HandleConfigError(err, true)
			os.Exit(1)
		}
		listCollectionsForDatabase(ctx, dbConfig, limit, virtual, jsonOutput, false)
		return
	}

	// Use new vector database selector based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector databases: %v", err))
		os.Exit(1)
	}

	// Decide whether to show summary:
	// - Explicit --summary/-S flag: always show summary
	// - Explicit --details flag: never show summary (force detailed view)
	// - Default: summary for multiple VDBs, detailed for single VDB
	showSummary := summaryOutput || (len(selection.Configs) > 1 && !jsonOutput && !detailsOutput)

	// If summary mode, collect all results first
	if showSummary && !jsonOutput {
		displayCollectionsSummary(ctx, selection.Configs, limit)
		return
	}

	// List collections for each selected database (detailed view)
	for i, dbConfig := range selection.Configs {
		// If multiple databases, show which database we're listing
		if len(selection.Configs) > 1 {
			dbType := utils.GetVectorDBTypeFromConfig(&dbConfig)
			if !jsonOutput {
				utils.PrintInfo(fmt.Sprintf("\n=== %s ===", dbType))
			}
		}

		listCollectionsForDatabase(ctx, &dbConfig, limit, virtual, jsonOutput, false)

		// Add spacing between databases (except for the last one)
		if i < len(selection.Configs)-1 && !jsonOutput {
			fmt.Println()
		}
	}
}

// listCollectionsForDatabase lists collections for a specific database configuration
func listCollectionsForDatabase(ctx context.Context, dbConfig *config.VectorDBConfig, limit int, virtual bool, jsonOutput bool, summaryOnly bool) {
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.ListWeaviateCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMock:
		utils.ListMockCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeSupabase, config.VectorDBTypeSupabaseCloud:
		utils.ListSupabaseCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMongoDB, config.VectorDBTypeMongoDBCloud:
		utils.ListMongoDBCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
		utils.ListMilvusCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
		utils.ListChromaCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeQdrantLocal, config.VectorDBTypeQdrantCloud:
		utils.ListQdrantCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	case config.VectorDBTypeNeo4jLocal, config.VectorDBTypeNeo4jCloud:
		utils.ListNeo4jCollections(ctx, dbConfig, limit, virtual, jsonOutput)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
	}
}

// displayCollectionsSummary shows a summary table of collections across databases
func displayCollectionsSummary(ctx context.Context, dbConfigs []config.VectorDBConfig, limit int) {
	utils.PrintHeader("Collections Summary")
	fmt.Println()

	// Table header
	fmt.Printf("%-20s %-20s %-10s %s\n", "VDB", "TYPE", "COLS", "STATUS")
	fmt.Println("────────────────────────────────────────────────────────────────")

	totalCollections := 0
	successfulDBs := 0
	failedDBs := 0

	// Collect data for each database
	for _, dbConfig := range dbConfigs {
		dbName := dbConfig.Name
		dbType := string(dbConfig.Type)

		// Get collection count
		count, err := getCollectionCount(ctx, &dbConfig)

		status := "✓ OK"
		statusColor := color.New(color.FgGreen)
		if err != nil {
			status = "✗ FAIL"
			statusColor = color.New(color.FgRed)
			count = 0
			failedDBs++
		} else {
			successfulDBs++
			totalCollections += count
		}

		// Print row
		fmt.Printf("%-20s %-20s %-10d ", dbName, dbType, count)
		statusColor.Printf("%s\n", status)
	}

	// Print footer summary
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d databases  |  Collections: %d  |  ",
		len(dbConfigs), totalCollections)
	if successfulDBs > 0 {
		color.New(color.FgGreen).Printf("OK: %d", successfulDBs)
	}
	if failedDBs > 0 {
		if successfulDBs > 0 {
			fmt.Print("  |  ")
		}
		color.New(color.FgRed).Printf("Failed: %d", failedDBs)
	}
	fmt.Println()
}

// getCollectionCount returns the number of collections in a database
func getCollectionCount(ctx context.Context, dbConfig *config.VectorDBConfig) (int, error) {
	// Create vector DB client
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
		return 0, err
	}

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return 0, err
	}

	return len(collections), nil
}
