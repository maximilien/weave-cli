// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	configpkg "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// syncCmd represents the config sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync local configuration files to global ~/.weave-cli directory",
	Long: `Sync local configuration files (.env and config.yaml) to the global ~/.weave-cli directory.

This command copies your local configuration files to ~/.weave-cli, allowing you to:
- Use the same configuration across multiple directories
- Have a centralized backup of your configuration
- Easily switch between projects

The sync operation:
1. Checks if local .env and config.yaml files exist
2. Creates ~/.weave-cli directory if it doesn't exist
3. Copies local files to global directory
4. Preserves file permissions
5. Warns before overwriting existing global files

Examples:
  # Sync both .env and config.yaml to global directory
  weave config sync

  # Sync only .env file
  weave config sync --env

  # Sync only config.yaml file
  weave config sync --config-yaml`,
	Run: runConfigSync,
}

var (
	syncEnv        bool
	syncConfigYAML bool
)

func init() {
	syncCmd.Flags().BoolVar(&syncEnv, "env", false, "Sync only .env file")
	syncCmd.Flags().BoolVar(&syncConfigYAML, "config-yaml", false, "Sync only config.yaml file")
}

func runConfigSync(cmd *cobra.Command, args []string) {
	// If no flags are set, sync both files
	syncBoth := !syncEnv && !syncConfigYAML
	if syncBoth {
		syncEnv = true
		syncConfigYAML = true
	}

	// Display header
	printHeader("Weave CLI Configuration Sync")
	fmt.Println()
	color.New(color.FgCyan).Println("This tool will sync your local configuration files to ~/.weave-cli")
	fmt.Println()

	// Get global config directory
	globalDir, err := configpkg.GetGlobalConfigDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get global config directory: %v", err))
		os.Exit(1)
	}

	// Ensure global directory exists
	created, err := configpkg.EnsureGlobalConfigDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to create global config directory: %v", err))
		os.Exit(1)
	}

	if created {
		color.New(color.FgGreen).Printf("✓ Created global config directory: %s\n", globalDir)
		fmt.Println()
	}

	synced := 0
	skipped := 0

	// Sync .env file
	if syncEnv {
		localEnv := ".env"
		globalEnv := filepath.Join(globalDir, ".env")

		if !fileExists(localEnv) {
			color.New(color.FgYellow).Printf("⚠️  Local %s not found - skipping\n", localEnv)
			skipped++
		} else {
			// Check if global file exists
			if fileExists(globalEnv) {
				color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", globalEnv)
				fmt.Println()
				color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

				response, err := readLine()
				if err != nil {
					printError(fmt.Sprintf("Failed to read input: %v", err))
					os.Exit(1)
				}

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					color.New(color.FgBlue).Printf("→ Skipping %s\n", localEnv)
					skipped++
				} else {
					if err := copyFile(localEnv, globalEnv); err != nil {
						printError(fmt.Sprintf("Failed to copy %s: %v", localEnv, err))
						os.Exit(1)
					}
					color.New(color.FgGreen).Printf("✓ Synced %s → %s\n", localEnv, globalEnv)
					synced++
				}
			} else {
				if err := copyFile(localEnv, globalEnv); err != nil {
					printError(fmt.Sprintf("Failed to copy %s: %v", localEnv, err))
					os.Exit(1)
				}
				color.New(color.FgGreen).Printf("✓ Synced %s → %s\n", localEnv, globalEnv)
				synced++
			}
		}
		fmt.Println()
	}

	// Sync config.yaml file
	if syncConfigYAML {
		localConfig := "config.yaml"
		globalConfig := filepath.Join(globalDir, "config.yaml")

		if !fileExists(localConfig) {
			color.New(color.FgYellow).Printf("⚠️  Local %s not found - skipping\n", localConfig)
			skipped++
		} else {
			// Check if global file exists
			if fileExists(globalConfig) {
				color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", globalConfig)
				fmt.Println()
				color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

				response, err := readLine()
				if err != nil {
					printError(fmt.Sprintf("Failed to read input: %v", err))
					os.Exit(1)
				}

				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					color.New(color.FgBlue).Printf("→ Skipping %s\n", localConfig)
					skipped++
				} else {
					if err := copyFile(localConfig, globalConfig); err != nil {
						printError(fmt.Sprintf("Failed to copy %s: %v", localConfig, err))
						os.Exit(1)
					}
					color.New(color.FgGreen).Printf("✓ Synced %s → %s\n", localConfig, globalConfig)
					synced++
				}
			} else {
				if err := copyFile(localConfig, globalConfig); err != nil {
					printError(fmt.Sprintf("Failed to copy %s: %v", localConfig, err))
					os.Exit(1)
				}
				color.New(color.FgGreen).Printf("✓ Synced %s → %s\n", localConfig, globalConfig)
				synced++
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("═", 60))
	color.New(color.FgCyan, color.Bold).Println("Sync Summary:")
	fmt.Printf("• Files synced: %d\n", synced)
	fmt.Printf("• Files skipped: %d\n", skipped)
	fmt.Printf("• Global directory: %s\n", globalDir)
	fmt.Println()

	if synced > 0 {
		color.New(color.FgGreen, color.Bold).Println("✅ Sync completed successfully!")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 Your configuration is now available in ~/.weave-cli")
		color.New(color.FgCyan).Println("   Run 'weave config show' from any directory to use it")
	} else {
		color.New(color.FgYellow).Println("ℹ No files were synced")
	}
}
