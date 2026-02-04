// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	configpkg "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// fixCmd represents the config fix command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Interactively fix configuration errors and warnings",
	Long: `Interactively fix configuration errors and warnings detected by validation.

This command helps you fix configuration issues by:
1. Parsing validation errors and warnings
2. Prompting you for each issue with helpful context
3. Offering multiple fix options (enter value, skip, remove, disable)
4. Applying fixes atomically with automatic backups
5. Re-validating configuration after fixes

The fix process is interactive and allows you to:
- Enter missing values (with secret masking for API keys)
- Skip issues to fix later
- Remove unused databases from config
- Disable databases by commenting them out
- Quit at any time without making changes

Examples:
  # Interactive fix (default)
  weave config fix

  # Dry run - see what would be fixed without making changes
  weave config fix --dry-run

  # Auto-fix warnings only (skip errors that need input)
  weave config fix --auto-fix --errors-only

  # Fix specific config file
  weave config fix --config-file ~/.weave-cli/config.yaml

  # Skip backup creation (not recommended)
  weave config fix --no-backup

Features:
  • Interactive prompts with helpful hints
  • Secret masking for API keys and tokens
  • Automatic backups before changes
  • Atomic updates (all or nothing)
  • Re-validation after fixes
  • Dry-run mode for preview`,
	Run: runConfigFix,
}

var (
	fixConfigFile string
	fixDryRun     bool
	fixAutoFix    bool
	fixErrorsOnly bool
	fixNoBackup   bool
)

func init() {
	fixCmd.Flags().StringVar(&fixConfigFile, "config-file", "config.yaml", "Config file to fix")
	fixCmd.Flags().BoolVar(&fixDryRun, "dry-run", false, "Show what would be fixed without making changes")
	fixCmd.Flags().BoolVar(&fixAutoFix, "auto-fix", false, "Automatically fix warnings (skip prompts for warnings)")
	fixCmd.Flags().BoolVar(&fixErrorsOnly, "errors-only", false, "Only fix errors, ignore warnings")
	fixCmd.Flags().BoolVar(&fixNoBackup, "no-backup", false, "Skip creating backup before making changes")
}

func runConfigFix(cmd *cobra.Command, args []string) {
	// Check if config file exists
	if _, err := os.Stat(fixConfigFile); os.IsNotExist(err) {
		printError(fmt.Sprintf("Config file not found: %s", fixConfigFile))
		fmt.Println()
		fmt.Println("Create a config file first:")
		fmt.Println("  weave config create --config-yaml")
		os.Exit(1)
	}

	// Load and validate config
	color.Cyan("🔍 Validating configuration...")
	fmt.Println()

	cfg, validationOutput, err := validateConfig(fixConfigFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Check if there are any issues
	if validationOutput == "" {
		color.Green("✅ Configuration is valid - no issues to fix!")
		os.Exit(0)
	}

	// Parse validation output
	issues, err := configpkg.ParseValidationOutput(validationOutput)
	if err != nil {
		printError(fmt.Sprintf("Failed to parse validation output: %v", err))
		os.Exit(1)
	}

	// Filter issues based on flags
	if fixErrorsOnly {
		var errors []configpkg.ConfigIssue
		for _, issue := range issues {
			if issue.Type == configpkg.IssueTypeError {
				errors = append(errors, issue)
			}
		}
		issues = errors
	}

	if len(issues) == 0 {
		color.Green("✅ No issues to fix!")
		os.Exit(0)
	}

	// Run interactive fixer
	fixer := configpkg.NewInteractiveFixer(issues)
	results, err := fixer.Run()
	if err != nil {
		if strings.Contains(err.Error(), "user cancelled") || strings.Contains(err.Error(), "user quit") {
			color.Yellow("\n👋 No changes made")
			os.Exit(0)
		}
		printError(fmt.Sprintf("Failed to run interactive fixer: %v", err))
		os.Exit(1)
	}

	// Dry run mode
	if fixDryRun {
		configpkg.DryRun(results)
		os.Exit(0)
	}

	// Apply fixes
	applier, err := configpkg.NewFixApplier(fixConfigFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to create applier: %v", err))
		os.Exit(1)
	}

	fmt.Println()
	color.Cyan("💾 Applying changes...")
	fmt.Println()

	if err := applier.ApplyFixes(results, !fixNoBackup); err != nil {
		printError(fmt.Sprintf("Failed to apply fixes: %v", err))
		if backupPath := applier.GetBackupPath(); backupPath != "" {
			fmt.Println()
			color.Yellow("💡 Restore from backup: cp %s %s", backupPath, fixConfigFile)
		}
		os.Exit(1)
	}

	color.Green("✅ Configuration updated successfully!")
	fmt.Println()

	// Re-validate
	color.Cyan("🔍 Re-validating configuration...")
	fmt.Println()

	_, validationOutput, err = validateConfig(fixConfigFile)
	if err != nil {
		color.Yellow("⚠️  Warning: Failed to re-validate: %v", err)
	} else if validationOutput != "" {
		color.Yellow("⚠️  Note: Some issues remain:")
		fmt.Println(validationOutput)
		fmt.Println()
		color.Cyan("Run 'weave config fix' again to address remaining issues")
	} else {
		color.Green("✅ Configuration is now valid!")
	}

	// Show what to do next
	fmt.Println()
	color.Cyan("Next steps:")
	fmt.Println("  • Test your configuration: weave config show")
	fmt.Println("  • List databases: weave config list")
	if cfg != nil {
		fmt.Printf("  • Try it out: weave cols ls --%s\n", cfg.Databases.Default)
	}
}

// validateConfig validates a config file and returns validation output
func validateConfig(configPath string) (*configpkg.Config, string, error) {
	// Set config file path for viper
	origConfigFile := os.Getenv("CONFIG_FILE")
	os.Setenv("CONFIG_FILE", configPath)
	defer func() {
		if origConfigFile != "" {
			os.Setenv("CONFIG_FILE", origConfigFile)
		} else {
			os.Unsetenv("CONFIG_FILE")
		}
	}()

	// Load config
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		return nil, "", err
	}

	// Run validation by checking each database config
	var validationIssues []string

	for i, db := range cfg.Databases.VectorDatabases {
		// Check required fields based on database type
		issues := validateDatabase(i, db)
		validationIssues = append(validationIssues, issues...)
	}

	if len(validationIssues) == 0 {
		return cfg, "", nil
	}

	// Format output like existing validation
	var output strings.Builder
	output.WriteString("⚠️  Configuration Errors:\n")
	for _, issue := range validationIssues {
		output.WriteString(issue)
		output.WriteString("\n")
	}

	return cfg, output.String(), nil
}

// validateDatabase checks a database configuration for errors
func validateDatabase(index int, db configpkg.VectorDBConfig) []string {
	var issues []string

	switch db.Type {
	case configpkg.VectorDBTypeMongoDB:
		if db.DatabaseURL == "" {
			issues = append(issues, fmt.Sprintf("  • databases.vector_databases[%d] (%s).database_url: MongoDB database_url is required\n    💡 Set 'database_url' field with MongoDB Atlas connection string", index, db.Name))
		}
	case configpkg.VectorDBTypeMilvusCloud:
		if db.APIKey == "" {
			issues = append(issues, fmt.Sprintf("  • databases.vector_databases[%d] (%s).api_key: Milvus Cloud (Zilliz) requires an API key\n    💡 Set 'api_key' field for authentication", index, db.Name))
		}
	case configpkg.VectorDBTypeChromaCloud:
		if db.Tenant == "" {
			issues = append(issues, fmt.Sprintf("  • databases.vector_databases[%d] (%s).tenant: Chroma Cloud requires a tenant ID\n    💡 Set 'tenant' field to your Chroma Cloud tenant ID", index, db.Name))
		}
	case configpkg.VectorDBTypeQdrantCloud:
		if db.APIKey == "" {
			issues = append(issues, fmt.Sprintf("  • databases.vector_databases[%d] (%s).api_key: Qdrant Cloud requires an API key\n    💡 Set 'api_key' field for authentication", index, db.Name))
		}
	}

	return issues
}
