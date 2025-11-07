// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

// LoadConfigWithInteractiveHelp loads configuration and offers interactive help on errors
func LoadConfigWithInteractiveHelp() (*config.Config, error) {
	cfg, err := LoadConfigWithOverrides()

	// If config loaded successfully, check if we're using mock database as a fallback
	if err == nil {
		// Check if environment variables are missing
		if configErr := config.CheckRequiredEnvVars(); configErr != nil {
			// Get the default database to see if we fell back to mock
			dbConfig, dbErr := cfg.GetDefaultDatabase()
			if dbErr == nil && dbConfig.Type == config.VectorDBTypeMock {
				// We're using mock database as fallback - warn and offer to configure
				color.New(color.FgYellow).Println("\n⚠️  Warning: No configuration found, using mock database for testing")
				fmt.Println()

				// Format and display configuration error with tips
				formattedErr := config.FormatConfigError(configErr)
				fmt.Println(formattedErr)

				// Prompt user to fix configuration
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
						PrintError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
						// Continue with mock database
						return cfg, nil
					}

					// Try loading config again with real configuration
					cfg, err = LoadConfigWithOverrides()
					if err != nil {
						PrintWarning("Failed to load configuration after setup, continuing with mock database")
						return cfg, nil
					}

					PrintSuccess("✅ Configuration updated! Using real database now.")
					return cfg, nil
				}

				// User declined, continue with mock
				fmt.Println()
				PrintInfo("Continuing with mock database. Set environment variables or create .env file to use real database.")
				fmt.Println()
			}
		}
		return cfg, nil
	}

	// Handle configuration loading errors
	if err != nil {
		// Check if this is a configuration error
		if configErr, ok := err.(*config.ConfigError); ok {
			// Format and display the detailed error
			formattedErr := config.FormatConfigError(configErr)
			PrintError(formattedErr)

			// Prompt user to fix configuration interactively
			shouldFix, promptErr := config.PromptToFixConfig(configErr)
			if promptErr != nil {
				return nil, fmt.Errorf("failed to prompt for configuration fix: %w", promptErr)
			}

			if shouldFix {
				// Run interactive configuration fix
				if err := config.InteractiveConfigFix(configErr.EnvFileExists); err != nil {
					return nil, fmt.Errorf("failed to fix configuration: %w", err)
				}

				// Try loading config again
				cfg, err = LoadConfigWithOverrides()
				if err != nil {
					return nil, fmt.Errorf("configuration still invalid after fix attempt: %w", err)
				}

				return cfg, nil
			}

			// User declined to fix, return original error
			return nil, configErr
		}

		// Check if error message suggests configuration issues
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// The error was enhanced with configuration tips
			PrintError(formattedErr)

			// Check if we can prompt for fix
			if configErr := config.CheckRequiredEnvVars(); configErr != nil {
				shouldFix, promptErr := config.PromptToFixConfig(configErr)
				if promptErr == nil && shouldFix {
					if err := config.InteractiveConfigFix(configErr.EnvFileExists); err != nil {
						return nil, fmt.Errorf("failed to fix configuration: %w", err)
					}

					// Try loading config again
					cfg, err = LoadConfigWithOverrides()
					if err != nil {
						return nil, fmt.Errorf("configuration still invalid after fix attempt: %w", err)
					}

					return cfg, nil
				}
			}

			return nil, err
		}

		// Not a configuration error, return as-is
		return nil, err
	}

	return cfg, nil
}

// HandleConfigError handles configuration errors with helpful messages
func HandleConfigError(err error, exitOnError bool) bool {
	if err == nil {
		return false
	}

	// Format the error with helpful tips
	formattedErr := config.FormatConfigError(err)

	// Print the formatted error
	if formattedErr != err.Error() {
		// Enhanced error message
		color.New(color.FgRed).Println(formattedErr)
	} else {
		// Standard error
		PrintError(fmt.Sprintf("Configuration error: %v", err))
	}

	if exitOnError {
		os.Exit(1)
	}

	return true
}

// PrintConfigTips prints helpful tips about configuration
func PrintConfigTips() {
	if viper.GetBool("no-tips") {
		return
	}

	color.New(color.FgCyan).Println("\n💡 Configuration Tips:")
	fmt.Println()
	fmt.Println("  • Use 'weave config show' to see your current configuration")
	fmt.Println("  • Use 'weave config create --env' to create a new .env file")
	fmt.Println("  • Use 'weave config update --env' to update an existing .env file")
	fmt.Println("  • Set VECTOR_DB_TYPE=mock for testing without a real database")
	fmt.Println()
}
