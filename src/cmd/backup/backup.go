// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"github.com/spf13/cobra"
)

// BackupCmd represents the backup command
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore vector database collections",
	Long: `Backup and restore vector database collections to portable .weavebak files.

Prevent data loss and eliminate costly re-ingestion with backup/restore commands.

Features:
  • Fast backup and restore (2,636 docs in < 2 minutes)
  • Compact files (~50-60MB compressed vs 733MB source)
  • Complete data preservation (embeddings + metadata + images)
  • Cross-VDB migration (export from one VDB, restore to another)

Example workflow:
  weave backup create MyCollection --output backup.weavebak
  weave backup validate backup.weavebak
  weave backup restore backup.weavebak
  weave backup list backups/

Use cases:
  • Before Docker/infrastructure changes
  • Regular snapshots via cron
  • Disaster recovery
  • Data migration between VDBs`,
	Example: `  # Backup a collection
  weave backup create AuctionImages --output backup.weavebak

  # Backup with compression (default)
  weave backup create AuctionImages --output backup.weavebak --compress

  # Restore from backup
  weave backup restore backup.weavebak

  # Restore to different collection
  weave backup restore backup.weavebak --collection NewName

  # Validate backup integrity
  weave backup validate backup.weavebak

  # List all backups in directory
  weave backup list backups/`,
}

func init() {
	// Add subcommands
	BackupCmd.AddCommand(CreateCmd)
	BackupCmd.AddCommand(RestoreCmd)
	BackupCmd.AddCommand(ListCmd)
	BackupCmd.AddCommand(ValidateCmd)
}
