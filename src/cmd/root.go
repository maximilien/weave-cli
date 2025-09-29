// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	envFile        string
	noColor        bool
	noTruncate     bool
	vectorDBType   string
	weaviateAPIKey string
	weaviateURL    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "weave",
	Short: "Weave VDB Management Tool",
	Long: `Weave is a command-line tool for managing Weaviate vector databases.

This tool provides commands following a consistent pattern:
  weave config show                    # Show current configuration
  weave health check                   # Check database health
  weave collection list                # List all collections
  weave collection create COLLECTION [COLLECTION...]  # Create one or more collections
  weave collection delete COLLECTION [COLLECTION...] # Clear one or more collections (delete all documents)
  weave collection delete-all          # Clear all collections (double confirmation)
  weave document list COLLECTION      # List documents in collection
  weave document show COLLECTION ID   # Show specific document
  weave document show COLLECTION --name filename.pdf  # Show document by filename
  weave document delete COLLECTION [ID] [ID...] # Delete one or more documents
  weave document delete COLLECTION --name filename.pdf  # Delete document by filename
  weave document delete-all COLLECTION # Delete all documents in collection (double confirmation)

The tool uses ./config.yaml and ./.env files by default, or you can specify
custom locations with --config and --env flags. Environment variables can be
overridden with --vector-db-type, --weaviate-api-key, and --weaviate-url flags.

Priority order: command flags > --env file > .env file > shell environment.`,
	Version: version.Get().Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initColor)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", "env file (default is ./.env)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output (minimal messages)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noTruncate, "no-truncate", false, "show all data without truncation")

	// Environment variable override flags (highest priority)
	rootCmd.PersistentFlags().StringVar(&vectorDBType, "vector-db-type", "", "override VECTOR_DB_TYPE (weaviate-cloud|weaviate-local|mock)")
	rootCmd.PersistentFlags().StringVar(&weaviateAPIKey, "weaviate-api-key", "", "override WEAVIATE_API_KEY")
	rootCmd.PersistentFlags().StringVar(&weaviateURL, "weaviate-url", "", "override WEAVIATE_URL")

	// Add version flag with custom handler
	rootCmd.Flags().BoolP("version", "V", false, "show version information")

	// Override the default version template
	rootCmd.SetVersionTemplate(version.String())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config.yaml in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	} else {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Warning: Could not read config file: %v\n", err)
		}
	}
}

// loadConfigWithOverrides loads configuration with command-line overrides
func loadConfigWithOverrides() (*config.Config, error) {
	return config.LoadConfigWithOptions(config.LoadConfigOptions{
		ConfigFile:     cfgFile,
		EnvFile:        envFile,
		VectorDBType:   vectorDBType,
		WeaviateAPIKey: weaviateAPIKey,
		WeaviateURL:    weaviateURL,
	})
}

// printHeader prints a colored header message
func printHeader(message string) {
	color.New(color.FgBlue, color.Bold).Printf("🔧 %s\n", message)
}

// printSuccess prints a colored success message
func printSuccess(message string) {
	color.New(color.FgGreen).Printf("✅ %s\n", message)
}

// printWarning prints a colored warning message
func printWarning(message string) {
	color.New(color.FgYellow).Printf("⚠️  %s\n", message)
}

// printError prints a colored error message
func printError(message string) {
	color.New(color.FgRed).Printf("❌ %s\n", message)
}

// initColor initializes color settings based on the no-color flag
func initColor() {
	if noColor {
		color.NoColor = true
	}
}
