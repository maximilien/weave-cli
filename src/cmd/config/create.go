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

// createCmd represents the config create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new configuration files from templates",
	Long: `Create new .env and config.yaml files from templates.

This command provides a guided way to create configuration files:
- Create .env file from .env.example with your credentials
- Create config.yaml from config.yaml.example
- Warn if files already exist before overwriting
- Interactive prompts for all configuration values

The command will:
1. Check if files already exist and warn before overwriting
2. Copy example files as templates
3. Prompt you for each configuration value (for .env)
4. Hide secrets (API keys, passwords) during input
5. Allow you to skip values to use defaults
6. Confirm before saving changes

Examples:
  # Create .env file interactively
  weave config create --env

  # Create config.yaml file
  weave config create --config-yaml

  # Create both files
  weave config create --env --config-yaml`,
	Run: runConfigCreate,
}

var (
	createEnv        bool
	createConfigYAML bool
	createGlobal     bool
)

func init() {
	createCmd.Flags().BoolVar(&createEnv, "env", false, "Create .env file")
	createCmd.Flags().BoolVar(&createConfigYAML, "config-yaml", false, "Create config.yaml file")
	createCmd.Flags().BoolVar(&createGlobal, "global", false, "Create config files in ~/.weave-cli directory (default: current directory)")
}

func runConfigCreate(cmd *cobra.Command, args []string) {
	// Check if at least one flag is set
	if !createEnv && !createConfigYAML {
		printError("Please specify at least one flag: --env or --config-yaml")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  weave config create --env")
		fmt.Println("  weave config create --config-yaml")
		fmt.Println("  weave config create --env --config-yaml")
		fmt.Println("  weave config create --env --global  # Create in ~/.weave-cli")
		os.Exit(1)
	}

	// Determine target directory
	var targetDir string
	var locationDesc string
	if createGlobal {
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

		targetDir = globalDir
		locationDesc = fmt.Sprintf("global directory (%s)", globalDir)
	} else {
		targetDir = "."
		locationDesc = "current directory"
	}

	// Display welcome message
	printHeader("Weave CLI Configuration Creator")
	fmt.Println()
	color.New(color.FgCyan).Println("This tool will help you create new configuration files from templates.")
	color.New(color.FgCyan).Printf("Target location: %s\n", locationDesc)
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Tips:")
	fmt.Println("  • Press Enter to use default/example values")
	fmt.Println("  • Type 'quit' or press Ctrl+C to exit")
	fmt.Println("  • Secrets will be hidden during input")
	fmt.Println()

	// Create .env file
	if createEnv {
		if err := createEnvFileInDir(targetDir); err != nil {
			printError(fmt.Sprintf("Failed to create .env file: %v", err))
			os.Exit(1)
		}
	}

	// Create config.yaml file
	if createConfigYAML {
		if err := createConfigYAMLFileInDir(targetDir); err != nil {
			printError(fmt.Sprintf("Failed to create config.yaml file: %v", err))
			os.Exit(1)
		}
	}

	// Success message
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("✅ Configuration creation completed!")
	fmt.Println()
	color.New(color.FgCyan).Println("Next steps:")
	fmt.Println("  • Run 'weave config show' to verify your configuration")
	fmt.Println("  • Run 'weave health check' to test your connection")
}

// createEnvFileInDir creates a new .env file from template in the specified directory
func createEnvFileInDir(targetDir string) error {
	printHeader("📝 Creating .env File")
	fmt.Println()

	envFile := filepath.Join(targetDir, ".env")
	exampleFile := ".env.example"

	// Check if .env already exists
	if fileExists(envFile) {
		color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", envFile)
		fmt.Println()
		color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

		response, err := readLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			color.New(color.FgBlue).Printf("→ Skipping %s creation\n", envFile)
			return nil
		}
		fmt.Println()
	}

	// Check if example file exists
	if !fileExists(exampleFile) {
		return fmt.Errorf("%s not found - cannot create .env without template", exampleFile)
	}

	color.New(color.FgGreen).Printf("✓ Found template: %s\n", exampleFile)
	fmt.Println()

	// Get all environment variables to configure
	envVars := getEnvVariables()
	updatedValues := make(map[string]string)

	// Prompt for each variable
	for _, envVar := range envVars {
		// Display variable info
		fmt.Println(strings.Repeat("─", 60))
		color.New(color.FgCyan, color.Bold).Printf("%s\n", envVar.Key)
		fmt.Printf("Description: %s\n", envVar.Description)

		if envVar.Required {
			color.New(color.FgRed).Print("Required: Yes\n")
		} else {
			color.New(color.FgYellow).Print("Required: No (optional)\n")
		}

		if envVar.Example != "" {
			color.New(color.FgBlue).Printf("Example: %s\n", envVar.Example)
		}

		// Prompt for value
		fmt.Println()
		if envVar.IsSecret {
			color.New(color.FgYellow).Print("Enter value (hidden, press Enter to skip): ")
		} else {
			color.New(color.FgYellow).Print("Enter value (press Enter to skip): ")
		}

		// Read input
		var newValue string
		var err error
		if envVar.IsSecret {
			newValue, err = readSecret()
		} else {
			newValue, err = readLine()
		}

		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Check for quit command
		if strings.ToLower(strings.TrimSpace(newValue)) == "quit" {
			fmt.Println()
			color.New(color.FgYellow).Println("Configuration creation cancelled by user")
			os.Exit(0)
		}

		// Store value if provided
		if strings.TrimSpace(newValue) != "" {
			updatedValues[envVar.Key] = strings.TrimSpace(newValue)
			color.New(color.FgGreen).Println("✓ Value set")
		} else {
			color.New(color.FgYellow).Println("→ Skipped (will use example/empty)")
		}
		fmt.Println()
	}

	// Show summary
	fmt.Println(strings.Repeat("═", 60))
	color.New(color.FgCyan, color.Bold).Println("Summary:")
	fmt.Printf("• New file: %s\n", envFile)
	fmt.Printf("• Variables configured: %d\n", len(updatedValues))
	fmt.Println()

	// Confirm before saving
	if !confirmCreate(envFile) {
		color.New(color.FgYellow).Println("Creation cancelled - no files modified")
		return nil
	}

	// Save the file
	if err := saveEnvFile(envFile, updatedValues); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Successfully created %s\n", envFile)
	return nil
}

// createConfigYAMLFileInDir creates a new config.yaml file from template in the specified directory
func createConfigYAMLFileInDir(targetDir string) error {
	printHeader("📝 Creating config.yaml File")
	fmt.Println()

	configFile := filepath.Join(targetDir, "config.yaml")
	exampleFile := "config.yaml.example"

	// Check if config.yaml already exists
	if fileExists(configFile) {
		color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", configFile)
		fmt.Println()
		color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

		response, err := readLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			color.New(color.FgBlue).Printf("→ Skipping %s creation\n", configFile)
			return nil
		}
		fmt.Println()
	}

	// Check if example file exists
	if !fileExists(exampleFile) {
		return fmt.Errorf("%s not found - cannot create config.yaml without template", exampleFile)
	}

	color.New(color.FgGreen).Printf("✓ Found template: %s\n", exampleFile)
	fmt.Println()

	// Info about config.yaml
	color.New(color.FgCyan).Println("💡 The config.yaml file is optional for advanced customization.")
	color.New(color.FgCyan).Println("   Most users only need the .env file!")
	fmt.Println()

	// Ask if user wants to continue
	color.New(color.FgYellow).Print("Do you want to create config.yaml? (y/N): ")
	response, err := readLine()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		color.New(color.FgBlue).Println("→ Skipping config.yaml creation")
		return nil
	}
	fmt.Println()

	// Copy example file
	color.New(color.FgCyan).Printf("Creating %s from %s...\n", configFile, exampleFile)
	if err := copyFile(exampleFile, configFile); err != nil {
		return fmt.Errorf("failed to copy example file: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Successfully created %s\n", configFile)
	fmt.Println()
	color.New(color.FgCyan).Println("💡 You can now edit this file manually to customize:")
	fmt.Println("   • Collection names")
	fmt.Println("   • Batch processing settings")
	fmt.Println("   • PDF processing options")
	fmt.Println("   • AI agent configuration")
	fmt.Println("   • Custom schemas")
	fmt.Println()
	color.New(color.FgCyan).Printf("   Run 'weave config show' to see your configuration\n")

	return nil
}

// confirmCreate asks user to confirm file creation
func confirmCreate(filename string) bool {
	color.New(color.FgYellow, color.Bold).Printf("💾 Create %s with these values? (Y/n): ", filename)
	response, err := readLine()
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	// Default to yes if empty
	return response == "" || response == "y" || response == "yes"
}
