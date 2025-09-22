package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	envFile string
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
  weave collection delete COLLECTION   # Delete a specific collection
  weave collection delete-all          # Delete all collections
  weave document list COLLECTION      # List documents in collection
  weave document show COLLECTION ID   # Show specific document
  weave document delete COLLECTION ID # Delete specific document
  weave document delete-all COLLECTION # Delete all documents in collection

The tool uses ./config.yaml and ./.env files by default, or you can specify
custom locations with --config and --env flags.`,
	Version: "1.0.0",
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
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", "env file (default is ./.env)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output (minimal messages)")

	// Add version flag
	rootCmd.Flags().BoolP("version", "V", false, "show version information")
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

// printInfo prints a colored info message
func printInfo(message string) {
	color.New(color.FgCyan).Printf("ℹ️  %s\n", message)
}
