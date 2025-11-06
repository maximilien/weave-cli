// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// updateCmd represents the config update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Interactively update configuration files",
	Long: `Interactively create or update .env and config.yaml files.

This command provides a guided, interactive way to:
- Create .env file from .env.example with your credentials
- Create config.yaml from config.yaml.example with your settings
- Update existing configuration files

The command will:
1. Load existing values (if files exist) or use examples
2. Prompt you for each configuration value
3. Hide secrets (API keys, passwords) during input
4. Allow you to skip values to keep current/default
5. Allow you to quit at any time (Ctrl+C or type 'quit')
6. Confirm before saving changes

Examples:
  # Update .env file interactively
  weave config update --env

  # Update config.yaml file interactively
  weave config update --config-yaml

  # Update both files
  weave config update --env --config-yaml`,
	Run: runConfigUpdate,
}

var (
	updateEnv        bool
	updateConfigYAML bool
)

func init() {
	updateCmd.Flags().BoolVar(&updateEnv, "env", false, "Update .env file")
	updateCmd.Flags().BoolVar(&updateConfigYAML, "config-yaml", false, "Update config.yaml file")
}

func runConfigUpdate(cmd *cobra.Command, args []string) {
	// Check if at least one flag is set
	if !updateEnv && !updateConfigYAML {
		printError("Please specify at least one flag: --env or --config-yaml")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  weave config update --env")
		fmt.Println("  weave config update --config-yaml")
		fmt.Println("  weave config update --env --config-yaml")
		os.Exit(1)
	}

	// Display welcome message
	printHeader("Weave CLI Configuration Updater")
	fmt.Println()
	color.New(color.FgCyan).Println("This tool will help you create or update your configuration files.")
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Tips:")
	fmt.Println("  • Press Enter to keep current/default value")
	fmt.Println("  • Type 'quit' or press Ctrl+C to exit")
	fmt.Println("  • Secrets will be hidden during input")
	fmt.Println()

	// Update .env file
	if updateEnv {
		if err := updateEnvFile(); err != nil {
			printError(fmt.Sprintf("Failed to update .env file: %v", err))
			os.Exit(1)
		}
	}

	// Update config.yaml file
	if updateConfigYAML {
		if err := updateConfigYAMLFile(); err != nil {
			printError(fmt.Sprintf("Failed to update config.yaml file: %v", err))
			os.Exit(1)
		}
	}

	// Success message
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("✅ Configuration update completed!")
	fmt.Println()
	color.New(color.FgCyan).Println("Next steps:")
	fmt.Println("  • Run 'weave config show' to verify your configuration")
	fmt.Println("  • Run 'weave health check' to test your connection")
}

// updateEnvFile interactively updates the .env file
func updateEnvFile() error {
	printHeader("📝 Updating .env File")
	fmt.Println()

	// Check if .env exists
	envFile := ".env"
	envExists := fileExists(envFile)

	if envExists {
		color.New(color.FgGreen).Printf("✓ Found existing %s file\n", envFile)
	} else {
		color.New(color.FgYellow).Printf("ℹ No %s file found - will create from .env.example\n", envFile)
	}
	fmt.Println()

	// Load existing values
	existingValues := make(map[string]string)
	if envExists {
		existingValues = loadEnvFile(envFile)
	}

	// Get all environment variables to configure
	envVars := getEnvVariables()
	updatedValues := make(map[string]string)
	changedCount := 0

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

		// Show current value (masked if secret)
		currentValue := existingValues[envVar.Key]
		if currentValue != "" {
			if envVar.IsSecret {
				color.New(color.FgGreen).Printf("Current: ***hidden***\n")
			} else {
				color.New(color.FgGreen).Printf("Current: %s\n", currentValue)
			}
		} else {
			color.New(color.FgYellow).Printf("Current: (not set)\n")
		}

		// Prompt for new value
		fmt.Println()
		if envVar.IsSecret {
			color.New(color.FgYellow).Print("Enter new value (hidden, press Enter to keep current): ")
		} else {
			color.New(color.FgYellow).Print("Enter new value (press Enter to keep current): ")
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
			color.New(color.FgYellow).Println("Configuration update cancelled by user")
			os.Exit(0)
		}

		// Use new value if provided, otherwise keep current
		if strings.TrimSpace(newValue) != "" {
			updatedValues[envVar.Key] = strings.TrimSpace(newValue)
			if currentValue != strings.TrimSpace(newValue) {
				changedCount++
				color.New(color.FgGreen).Println("✓ Value updated")
			}
		} else if currentValue != "" {
			updatedValues[envVar.Key] = currentValue
			color.New(color.FgBlue).Println("→ Keeping current value")
		} else {
			color.New(color.FgYellow).Println("→ Skipped (will remain empty)")
		}
		fmt.Println()
	}

	// Show summary
	fmt.Println(strings.Repeat("═", 60))
	color.New(color.FgCyan, color.Bold).Println("Summary:")
	if envExists {
		fmt.Printf("• Existing file: %s\n", envFile)
		fmt.Printf("• Variables changed: %d\n", changedCount)
	} else {
		fmt.Printf("• New file: %s\n", envFile)
		fmt.Printf("• Variables configured: %d\n", len(updatedValues))
	}
	fmt.Println()

	// Confirm before saving
	if !confirmSave(envFile) {
		color.New(color.FgYellow).Println("Changes discarded - no files modified")
		return nil
	}

	// Save the file
	if err := saveEnvFile(envFile, updatedValues); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Successfully saved %s\n", envFile)
	return nil
}

// updateConfigYAMLFile interactively updates the config.yaml file
func updateConfigYAMLFile() error {
	printHeader("📝 Updating config.yaml File")
	fmt.Println()

	// Check if config.yaml exists
	configFile := "config.yaml"
	configExists := fileExists(configFile)

	if configExists {
		color.New(color.FgGreen).Printf("✓ Found existing %s file\n", configFile)
	} else {
		color.New(color.FgYellow).Printf("ℹ No %s file found - will create from config.yaml.example\n", configFile)
	}
	fmt.Println()

	// For config.yaml, we'll provide a simpler interface
	// focusing on the most common settings
	color.New(color.FgCyan).Println("💡 The config.yaml file is optional for advanced customization.")
	color.New(color.FgCyan).Println("   Most users only need the .env file!")
	fmt.Println()

	// Ask if user wants to continue
	color.New(color.FgYellow).Print("Do you want to create/update config.yaml? (y/N): ")
	response, err := readLine()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		color.New(color.FgBlue).Println("→ Skipping config.yaml update")
		return nil
	}
	fmt.Println()

	// Copy example file if config doesn't exist
	if !configExists {
		exampleFile := "config.yaml.example"
		if !fileExists(exampleFile) {
			return fmt.Errorf("config.yaml.example not found")
		}

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
	} else {
		color.New(color.FgYellow).Printf("📄 %s already exists\n", configFile)
		fmt.Println()
		color.New(color.FgCyan).Println("💡 To customize, edit the file manually:")
		fmt.Printf("   • Collection names and schemas\n")
		fmt.Printf("   • Batch and PDF processing settings\n")
		fmt.Printf("   • AI agent configuration\n")
	}

	return nil
}
