package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health and connectivity management",
	Long: `Manage database health and connectivity.

This command provides subcommands to check database health and connectivity.`,
}

// healthCheckCmd represents the health check command
var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check health of VDB connections",
	Long: `Check the health of the configured vector database connections.

This command:
- Attempts to connect to the configured VDB
- Verifies API keys and authentication
- Tests collection access
- Reports connection status and any issues`,
	Run: runHealthCheck,
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.AddCommand(healthCheckCmd)
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Vector Database Health Check")
	fmt.Println()

	dbType := cfg.Database.VectorDB.Type
	color.New(color.FgCyan, color.Bold).Printf("Checking %s database...\n", dbType)
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var healthStatus bool
	var healthMessage string

	switch dbType {
	case config.VectorDBTypeCloud:
		healthStatus, healthMessage = checkWeaviateCloudHealth(ctx, &cfg.Database.VectorDB.WeaviateCloud)
	case config.VectorDBTypeLocal:
		healthStatus, healthMessage = checkWeaviateLocalHealth(ctx, &cfg.Database.VectorDB.WeaviateLocal)
	case config.VectorDBTypeMock:
		healthStatus, healthMessage = checkMockHealth(ctx, &cfg.Database.VectorDB.Mock)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbType))
		os.Exit(1)
	}

	// Display health status
	fmt.Println()
	if healthStatus {
		printSuccess("Database connection is healthy!")
		color.New(color.FgGreen).Printf("✅ %s\n", healthMessage)
	} else {
		printError("Database connection failed!")
		color.New(color.FgRed).Printf("❌ %s\n", healthMessage)
		os.Exit(1)
	}

	// Test collection access
	fmt.Println()
	printHeader("Collection Access Test")
	testCollectionAccess(ctx, cfg)
}

func checkWeaviateCloudHealth(ctx context.Context, cfg *config.WeaviateCloudConfig) (bool, string) {
	if cfg.URL == "" {
		return false, "Weaviate Cloud URL is not configured"
	}

	if cfg.APIKey == "" {
		return false, "Weaviate Cloud API key is not configured"
	}

	// Create client
	client, err := weaviate.NewClient(&weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return false, fmt.Sprintf("Failed to create Weaviate client: %v", err)
	}

	// Test connection
	if err := client.Health(ctx); err != nil {
		return false, fmt.Sprintf("Failed to connect to Weaviate Cloud: %v", err)
	}

	return true, fmt.Sprintf("Successfully connected to Weaviate Cloud at %s", cfg.URL)
}

func checkWeaviateLocalHealth(ctx context.Context, cfg *config.WeaviateLocalConfig) (bool, string) {
	if cfg.URL == "" {
		return false, "Weaviate Local URL is not configured"
	}

	// Create client
	client, err := weaviate.NewClient(&weaviate.Config{
		URL: cfg.URL,
	})
	if err != nil {
		return false, fmt.Sprintf("Failed to create Weaviate client: %v", err)
	}

	// Test connection
	if err := client.Health(ctx); err != nil {
		return false, fmt.Sprintf("Failed to connect to Weaviate Local: %v", err)
	}

	return true, fmt.Sprintf("Successfully connected to Weaviate Local at %s", cfg.URL)
}

func checkMockHealth(ctx context.Context, cfg *config.MockConfig) (bool, string) {
	if !cfg.Enabled {
		return false, "Mock database is not enabled"
	}

	// Create mock client
	client := mock.NewClient(cfg)

	// Test connection
	if err := client.Health(ctx); err != nil {
		return false, fmt.Sprintf("Mock database health check failed: %v", err)
	}

	return true, "Mock database is working correctly"
}

func testCollectionAccess(ctx context.Context, cfg *config.Config) {
	dbType := cfg.Database.VectorDB.Type

	switch dbType {
	case config.VectorDBTypeCloud:
		testWeaviateCollectionAccess(ctx, &cfg.Database.VectorDB.WeaviateCloud)
	case config.VectorDBTypeLocal:
		testWeaviateCollectionAccess(ctx, &cfg.Database.VectorDB.WeaviateLocal)
	case config.VectorDBTypeMock:
		testMockCollectionAccess(ctx, &cfg.Database.VectorDB.Mock)
	}
}

func testWeaviateCollectionAccess(ctx context.Context, cfg interface{}) {
	client, err := createWeaviateClient(cfg)

	if err != nil {
		printError(fmt.Sprintf("Failed to create client for collection test: %v", err))
		return
	}

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printWarning("No collections found in the database")
	} else {
		printSuccess(fmt.Sprintf("Found %d collections:", len(collections)))
		// Sort collections alphabetically
		sort.Strings(collections)
		for _, collection := range collections {
			fmt.Printf("  - %s\n", collection)
		}
	}
}

func testMockCollectionAccess(ctx context.Context, cfg *config.MockConfig) {
	client := mock.NewClient(cfg)

	// List collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to list mock collections: %v", err))
		return
	}

	if len(collections) == 0 {
		printWarning("No collections found in the mock database")
	} else {
		printSuccess(fmt.Sprintf("Found %d mock collections:", len(collections)))
		// Sort collections alphabetically
		sort.Strings(collections)
		for _, collection := range collections {
			fmt.Printf("  - %s\n", collection)
		}
	}
}
