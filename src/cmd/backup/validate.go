// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
)

var (
	validateJSON bool
)

// ValidateCmd represents the backup validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate <backup-file>",
	Short: "Validate a backup file",
	Long: `Validate the integrity and structure of a backup file.

Checks performed:
  • File exists and is readable
  • Valid JSON format
  • Correct backup version
  • Collection metadata present
  • All documents have required fields (ID, embedding)
  • Embedding dimensions match metadata
  • No duplicate document IDs

Returns exit code:
  • 0: Backup is valid
  • 1: Backup has errors`,
	Example: `  # Validate a backup
  weave backup validate backup.weavebak

  # Validate with JSON output
  weave backup validate backup.weavebak --json

  # Validate compressed backup
  weave backup validate backup.weavebak.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupValidate,
}

func init() {
	ValidateCmd.Flags().BoolVar(&validateJSON, "json", false, "Output validation results in JSON format")
}

func runBackupValidate(cmd *cobra.Command, args []string) error {
	backupFile := args[0]

	// Validate the backup
	result, err := backuppkg.ValidateBackup(backupFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Output results
	if validateJSON {
		return outputValidationJSON(result)
	}

	return outputValidationText(result, backupFile)
}

func outputValidationJSON(result *backuppkg.ValidationResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	if !result.Valid {
		os.Exit(1)
	}
	return nil
}

func outputValidationText(result *backuppkg.ValidationResult, backupFile string) error {
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("🔍 Validating backup: %s", backupFile))
	fmt.Println()

	// Show backup info
	if result.Collection != "" {
		utils.PrintInfo(fmt.Sprintf("   Collection: %s", result.Collection))
	}
	if result.Version != "" {
		utils.PrintInfo(fmt.Sprintf("   Version: %s", result.Version))
	}
	utils.PrintInfo(fmt.Sprintf("   Documents: %d", result.TotalDocuments))
	if result.BackupSizeBytes > 0 {
		sizeStr := formatValidationSize(result.BackupSizeBytes)
		utils.PrintInfo(fmt.Sprintf("   Size: %s", sizeStr))
	}
	fmt.Println()

	// Show errors
	if len(result.Errors) > 0 {
		utils.PrintError("❌ Validation Errors:")
		for _, err := range result.Errors {
			fmt.Printf("   • %s\n", err)
		}
		fmt.Println()
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		utils.PrintWarning("⚠️  Validation Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
		fmt.Println()
	}

	// Final result
	if result.Valid {
		utils.PrintSuccess("✅ Backup is valid and ready to restore!")
		fmt.Println()
		return nil
	}

	utils.PrintError("❌ Backup validation failed")
	utils.PrintInfo("   Fix the errors above before attempting restore")
	fmt.Println()
	os.Exit(1)
	return nil
}

func formatValidationSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
