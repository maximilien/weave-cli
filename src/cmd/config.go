package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Manage Weave CLI configuration.

This command provides subcommands to view and manage configuration settings.`,
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show [database-name]",
	Short: "Show currently configured databases",
	Long: `Show the currently configured vector database settings.

This command displays:
- All configured databases or a specific database
- Vector database type (weaviate-cloud, weaviate-local, mock)
- Connection details (URL, API key status)
- Collection names
- Configuration source files

If no database name is provided, it shows the default database.
Use 'weave config list' to see all available databases.`,
	Run: runConfigShow,
}

// configListCmd represents the config list command
var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all configured databases",
	Long: `List all configured vector databases.

This command displays:
- All configured database names
- Database types
- Which database is the default`,
	Run: runConfigList,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// If a specific database name is provided, show only that database
	if len(args) > 0 {
		dbName := args[0]
		dbConfig, err := cfg.GetDatabase(dbName)
		if err != nil {
			printError(fmt.Sprintf("Failed to get database '%s': %v", dbName, err))
			os.Exit(1)
		}

		printHeader(fmt.Sprintf("Database Configuration: %s", dbName))
		fmt.Println()
		displayDatabaseConfig(dbName, dbConfig)
	} else {
		// Show default database
		dbConfig, err := cfg.GetDefaultDatabase()
		if err != nil {
			printError(fmt.Sprintf("Failed to get default database: %v", err))
			os.Exit(1)
		}

		printHeader("Default Database Configuration")
		fmt.Println()
		displayDatabaseConfig("default", dbConfig)
	}

	// Display configuration sources
	fmt.Println()
	printHeader("Configuration Sources")
	fmt.Printf("Config file: %s\n", config.GetConfigFile())
	fmt.Printf("Env file: %s\n", config.GetEnvFile())
}

func runConfigList(cmd *cobra.Command, args []string) {
	cfgFile, _ := cmd.Flags().GetString("config")
	envFile, _ := cmd.Flags().GetString("env")

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile, envFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	printHeader("Configured Databases")
	fmt.Println()

	databaseNames := cfg.GetDatabaseNames()
	if len(databaseNames) == 0 {
		printWarning("No databases configured")
		return
	}

	for name, dbType := range databaseNames {
		isDefault := name == "default"
		if isDefault {
			color.New(color.FgGreen, color.Bold).Printf("• %s (default)\n", name)
		} else {
			fmt.Printf("• %s\n", name)
		}
		fmt.Printf("  Type: %s\n", dbType)
		fmt.Println()
	}
}

func displayDatabaseConfig(name string, dbConfig *config.VectorDBConfig) {
	color.New(color.FgCyan, color.Bold).Printf("Type: %s\n", dbConfig.Type)
	fmt.Println()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud:
		displayWeaviateCloudConfig(dbConfig)
	case config.VectorDBTypeLocal:
		displayWeaviateLocalConfig(dbConfig)
	case config.VectorDBTypeMock:
		displayMockConfig(dbConfig)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}

func displayWeaviateCloudConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgGreen).Printf("🌐 Weaviate Cloud Configuration\n")
	fmt.Printf("  URL: %s\n", cfg.URL)

	// Mask API key for security
	apiKeyDisplay := "***hidden***"
	if cfg.APIKey == "" {
		apiKeyDisplay = "❌ not set"
		color.New(color.FgRed).Printf("  API Key: %s\n", apiKeyDisplay)
	} else {
		color.New(color.FgGreen).Printf("  API Key: %s\n", apiKeyDisplay)
	}

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s)\n", collection.Name, collection.Type)
		}
	}
}

func displayWeaviateLocalConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgBlue).Printf("🏠 Weaviate Local Configuration\n")
	fmt.Printf("  URL: %s\n", cfg.URL)

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s)\n", collection.Name, collection.Type)
		}
	}
}

func displayMockConfig(cfg *config.VectorDBConfig) {
	color.New(color.FgYellow).Printf("🎭 Mock Database Configuration\n")
	fmt.Printf("  Enabled: %t\n", cfg.Enabled)
	fmt.Printf("  Simulate Embeddings: %t\n", cfg.SimulateEmbeddings)
	fmt.Printf("  Embedding Dimension: %d\n", cfg.EmbeddingDimension)

	if len(cfg.Collections) > 0 {
		fmt.Printf("  Collections:\n")
		for _, collection := range cfg.Collections {
			fmt.Printf("    - %s (%s): %s\n", collection.Name, collection.Type, collection.Description)
		}
	}
}
