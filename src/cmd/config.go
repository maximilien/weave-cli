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
	Use:   "show",
	Short: "Show currently configured VDB",
	Long: `Show the currently configured vector database settings.

This command displays:
- Vector database type (weaviate-cloud, weaviate-local, mock)
- Connection details (URL, API key status)
- Collection names
- Configuration source files`,
	Run: runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
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

	printHeader("Vector Database Configuration")
	fmt.Println()

	// Display vector database type
	dbType := cfg.Database.VectorDB.Type
	color.New(color.FgCyan, color.Bold).Printf("Type: %s\n", dbType)
	fmt.Println()

	switch dbType {
	case config.VectorDBTypeCloud:
		displayWeaviateCloudConfig(&cfg.Database.VectorDB.WeaviateCloud)
	case config.VectorDBTypeLocal:
		displayWeaviateLocalConfig(&cfg.Database.VectorDB.WeaviateLocal)
	case config.VectorDBTypeMock:
		displayMockConfig(&cfg.Database.VectorDB.Mock)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbType))
		os.Exit(1)
	}

	// Display configuration sources
	fmt.Println()
	printHeader("Configuration Sources")
	fmt.Printf("Config file: %s\n", config.GetConfigFile())
	fmt.Printf("Env file: %s\n", config.GetEnvFile())
}

func displayWeaviateCloudConfig(cfg *config.WeaviateCloudConfig) {
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

	fmt.Printf("  Collection: %s\n", cfg.CollectionName)
	fmt.Printf("  Test Collection: %s\n", cfg.CollectionNameTest)
}

func displayWeaviateLocalConfig(cfg *config.WeaviateLocalConfig) {
	color.New(color.FgBlue).Printf("🏠 Weaviate Local Configuration\n")
	fmt.Printf("  URL: %s\n", cfg.URL)
	fmt.Printf("  Collection: %s\n", cfg.CollectionName)
	fmt.Printf("  Test Collection: %s\n", cfg.CollectionNameTest)
}

func displayMockConfig(cfg *config.MockConfig) {
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
