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
	Use:   "check [database-name]",
	Short: "Check health of database connections",
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
	healthCmd.AddCommand(healthCheckCmd)
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If a specific database name is provided, check only that database
	if len(args) > 0 {
		dbName := args[0]
		dbConfig, err := cfg.GetDatabase(dbName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", dbName, err))
			os.Exit(1)
		}

		printHeader(fmt.Sprintf("Database Health Check: %s", dbName))
		fmt.Println()
		checkSingleDatabase(ctx, dbName, dbConfig)
	} else {
		// Check default database
		dbConfig, err := cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}

		printHeader("Default Database Health Check")
		fmt.Println()
		checkSingleDatabase(ctx, "default", dbConfig)
	}
}

func checkSingleDatabase(ctx context.Context, dbName string, dbConfig *config.VectorDBConfig) {
	color.New(color.FgCyan, color.Bold).Printf("Checking %s database...\n", dbConfig.Type)
	fmt.Println()

	var healthStatus bool
	var healthMessage string

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		healthStatus, healthMessage = checkWeaviateCloudHealth(ctx, dbConfig)
	case config.VectorDBTypeLocal:
		healthStatus, healthMessage = checkWeaviateLocalHealth(ctx, dbConfig)
	case config.VectorDBTypeMock:
		healthStatus, healthMessage = checkMockHealth(ctx, dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
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
	testCollectionAccessForDatabase(ctx, dbConfig)
}

func checkWeaviateCloudHealth(ctx context.Context, cfg *config.VectorDBConfig) (bool, string) {
	if cfg.URL == "" {
		return false, "Weaviate Cloud URL is not configured"
	}

	if cfg.APIKey == "" {
		return false, "Weaviate Cloud API key is not configured"
	}

	// Create client
	client, err := weaviate.NewClient(&weaviate.Config{
		URL:          cfg.URL,
		APIKey:       cfg.APIKey,
		OpenAIAPIKey: cfg.OpenAIAPIKey,
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

func checkWeaviateLocalHealth(ctx context.Context, cfg *config.VectorDBConfig) (bool, string) {
	if cfg.URL == "" {
		return false, "Weaviate Local URL is not configured"
	}

	// Create client
	client, err := weaviate.NewClient(&weaviate.Config{
		URL:          cfg.URL,
		OpenAIAPIKey: cfg.OpenAIAPIKey,
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

func checkMockHealth(ctx context.Context, cfg *config.VectorDBConfig) (bool, string) {
	if !cfg.Enabled {
		return false, "Mock database is not enabled"
	}

	// Create mock client - we need to convert to the old MockConfig structure
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

	// Test connection
	if err := client.Health(ctx); err != nil {
		return false, fmt.Sprintf("Mock database health check failed: %v", err)
	}

	return true, "Mock database is working correctly"
}

func testCollectionAccessForDatabase(ctx context.Context, dbConfig *config.VectorDBConfig) {
	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		testWeaviateCollectionAccess(ctx, dbConfig)
	case config.VectorDBTypeLocal:
		testWeaviateCollectionAccess(ctx, dbConfig)
	case config.VectorDBTypeMock:
		testMockCollectionAccess(ctx, dbConfig)
	}
}

func testWeaviateCollectionAccess(ctx context.Context, cfg *config.VectorDBConfig) {
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

func testMockCollectionAccess(ctx context.Context, cfg *config.VectorDBConfig) {
	// Convert to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	client := mock.NewClient(mockConfig)

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
