// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
)

var (
	listJSON bool
)

// ListCmd represents the backup list command
var ListCmd = &cobra.Command{
	Use:   "list [directory]",
	Short: "List backup files in a directory",
	Long: `List all .weavebak backup files in a directory with metadata.

Displays information about each backup:
  • Filename
  • Collection name
  • Document count
  • File size
  • Creation date
  • Weave version
  • Compression status

Useful for:
  • Finding available backups
  • Checking backup metadata before restore
  • Inventory management`,
	Example: `  # List backups in current directory
  weave backup list

  # List backups in specific directory
  weave backup list backups/

  # List with JSON output
  weave backup list backups/ --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBackupList,
}

func init() {
	ListCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
}

func runBackupList(cmd *cobra.Command, args []string) error {
	// Default to current directory
	directory := "."
	if len(args) > 0 {
		directory = args[0]
	}

	// Check if directory exists
	if info, err := os.Stat(directory); err != nil {
		return fmt.Errorf("directory not found: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", directory)
	}

	// Find all .weavebak files
	pattern1 := filepath.Join(directory, "*.weavebak")
	pattern2 := filepath.Join(directory, "*.weavebak.gz")

	files1, _ := filepath.Glob(pattern1)
	files2, _ := filepath.Glob(pattern2)

	files := append(files1, files2...)

	if len(files) == 0 {
		utils.PrintWarning(fmt.Sprintf("No backup files found in: %s", directory))
		return nil
	}

	// Read metadata from each backup
	backups := make([]*backuppkg.BackupInfo, 0, len(files))

	for _, file := range files {
		info, err := getBackupInfo(file)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to read %s: %v", filepath.Base(file), err))
			continue
		}
		backups = append(backups, info)
	}

	if len(backups) == 0 {
		utils.PrintWarning("No valid backup files found")
		return nil
	}

	// Sort by creation date (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	// Output
	if listJSON {
		return outputJSON(backups)
	}

	return outputTable(backups)
}

func getBackupInfo(filePath string) (*backuppkg.BackupInfo, error) {
	// Read backup metadata
	backup, err := backuppkg.ReadBackup(filePath)
	if err != nil {
		return nil, err
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Check if compressed
	compressed := strings.HasSuffix(filePath, ".gz")

	info := &backuppkg.BackupInfo{
		FilePath:        filePath,
		FileName:        filepath.Base(filePath),
		Collection:      backup.Metadata.Collection,
		VDBType:         backup.Metadata.VDBType,
		TotalDocuments:  backup.Metadata.TotalDocuments,
		BackupSizeBytes: fileInfo.Size(),
		CreatedAt:       backup.Metadata.CreatedAt,
		WeaveVersion:    backup.Metadata.WeaveVersion,
		Compressed:      compressed,
	}

	return info, nil
}

func outputJSON(backups []*backuppkg.BackupInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(backups)
}

func outputTable(backups []*backuppkg.BackupInfo) error {
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("Found %d backup file(s)", len(backups)))
	fmt.Println()

	// Print header
	fmt.Printf("%-35s %-20s %8s %10s %12s %s\n",
		"FILENAME", "COLLECTION", "DOCS", "SIZE", "COMPRESSED", "CREATED")
	fmt.Println(strings.Repeat("-", 120))

	// Print backups
	for _, backup := range backups {
		// Format size
		sizeStr := formatSize(backup.BackupSizeBytes)

		// Format date
		dateStr := backup.CreatedAt.Format("2006-01-02")

		// Compression status
		compressedStr := "No"
		if backup.Compressed {
			compressedStr = "Yes"
		}

		fmt.Printf("%-35s %-20s %8d %10s %12s %s\n",
			truncate(backup.FileName, 35),
			truncate(backup.Collection, 20),
			backup.TotalDocuments,
			sizeStr,
			compressedStr,
			dateStr)
	}

	fmt.Println()

	return nil
}

func formatSize(bytes int64) string {
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
