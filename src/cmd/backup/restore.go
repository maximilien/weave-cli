// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

var restoreOpts = &backuppkg.RestoreOptions{
	Overwrite: false,
	Quiet:     false,
}

// RestoreCmd represents the backup restore command
var RestoreCmd = &cobra.Command{
	Use:   "restore <backup-file> [flags]",
	Short: "Restore a collection from a backup file",
	Long: `Restore a vector database collection from a portable .weavebak file.

The restore command:
  • Reads backup file (auto-detects compression)
  • Validates backup format and version
  • Creates collection if it doesn't exist
  • Batch inserts all documents with progress tracking
  • Preserves embeddings, metadata, and images

Use cases:
  • Disaster recovery after data loss
  • Restore after infrastructure changes
  • Migrate data between VDB instances
  • Clone collections with new names`,
	Example: `  # Restore to original collection name
  weave backup restore backup.weavebak

  # Restore to different collection
  weave backup restore backup.weavebak --collection NewName

  # Overwrite existing collection
  weave backup restore backup.weavebak --overwrite

  # Restore to specific VDB
  weave backup restore backup.weavebak --vdb milvus-local

  # Quiet mode (no progress output)
  weave backup restore backup.weavebak --quiet`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupRestore,
}

func init() {
	RestoreCmd.Flags().StringVarP(&restoreOpts.Collection, "collection", "c", "", "Collection name (default: use name from backup)")
	RestoreCmd.Flags().BoolVar(&restoreOpts.Overwrite, "overwrite", false, "Overwrite existing collection")
	RestoreCmd.Flags().BoolVarP(&restoreOpts.Quiet, "quiet", "q", false, "Suppress progress output")
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupFile := args[0]
	restoreOpts.BackupFile = backupFile

	ctx := context.Background()

	if !restoreOpts.Quiet {
		utils.PrintInfo(fmt.Sprintf("📦 Restoring from backup: %s", backupFile))
		fmt.Println()
	}

	// Read backup file
	if !restoreOpts.Quiet {
		utils.PrintInfo("📥 Reading backup file...")
	}

	backup, err := backuppkg.ReadBackup(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if !restoreOpts.Quiet {
		utils.PrintSuccess("Backup file loaded successfully")
		utils.PrintInfo(fmt.Sprintf("   Version: %s", backup.Version))
		utils.PrintInfo(fmt.Sprintf("   Collection: %s", backup.Metadata.Collection))
		utils.PrintInfo(fmt.Sprintf("   Documents: %d", len(backup.Documents)))
		utils.PrintInfo(fmt.Sprintf("   VDB Type: %s", backup.Metadata.VDBType))
		utils.PrintInfo(fmt.Sprintf("   Created: %s", backup.Metadata.CreatedAt.Format(time.RFC3339)))
		fmt.Println()
	}

	// Determine collection name
	collectionName := restoreOpts.Collection
	if collectionName == "" {
		collectionName = backup.Metadata.Collection
	}

	if !restoreOpts.Quiet {
		if restoreOpts.Collection != "" {
			utils.PrintInfo(fmt.Sprintf("🔄 Restoring to collection: %s (renamed from %s)", collectionName, backup.Metadata.Collection))
		} else {
			utils.PrintInfo(fmt.Sprintf("🔄 Restoring to collection: %s", collectionName))
		}
	}

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		return err
	}

	// Get selected VDB
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		return fmt.Errorf("failed to determine vector databases: %w", err)
	}

	// Use first selected VDB (restore works on single VDB at a time)
	if len(selection.Configs) == 0 {
		return fmt.Errorf("no vector database configured")
	}
	vdbConfig := &selection.Configs[0]

	vdbClient, err := utils.CreateVectorDBClient(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}

	restoreOpts.VDBType = string(vdbConfig.Type)

	if !restoreOpts.Quiet {
		utils.PrintInfo(fmt.Sprintf("   Target VDB: %s", vdbConfig.Type))
		fmt.Println()
	}

	// Check if collection exists
	exists, err := vdbClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		if !restoreOpts.Overwrite {
			return fmt.Errorf("collection '%s' already exists (use --overwrite to replace)", collectionName)
		}

		if !restoreOpts.Quiet {
			utils.PrintWarning(fmt.Sprintf("⚠️  Collection '%s' exists - will be overwritten", collectionName))
		}

		// Delete and recreate collection
		if !restoreOpts.Quiet {
			utils.PrintInfo("🗑️  Deleting existing collection...")
		}

		if err := vdbClient.DeleteCollection(ctx, collectionName); err != nil {
			return fmt.Errorf("failed to delete existing collection: %w", err)
		}

		if !restoreOpts.Quiet {
			utils.PrintSuccess("Existing collection deleted")
		}

		// Now create it fresh
		exists = false
	}

	if !exists {
		// Create collection
		if !restoreOpts.Quiet {
			utils.PrintInfo(fmt.Sprintf("📝 Creating collection: %s", collectionName))
		}

		schema := &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: backup.Metadata.EmbeddingModel,
		}

		if err := vdbClient.CreateCollection(ctx, collectionName, schema); err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}

		if !restoreOpts.Quiet {
			utils.PrintSuccess("Collection created")
		}
	}

	// Restore documents in batches
	if !restoreOpts.Quiet {
		utils.PrintInfo("📤 Restoring documents...")
		fmt.Println()
	}

	startTime := time.Now()
	batchSize := 100 // Default batch size
	totalDocs := len(backup.Documents)
	restored := 0

	for i := 0; i < totalDocs; i += batchSize {
		end := i + batchSize
		if end > totalDocs {
			end = totalDocs
		}

		// Convert backup documents to vectordb documents
		batch := make([]*vectordb.Document, 0, end-i)
		for _, backupDoc := range backup.Documents[i:end] {
			doc := &vectordb.Document{
				ID:             backupDoc.ID,
				Content:        backupDoc.Content,
				Text:           backupDoc.Text,
				Embedding:      backupDoc.Embedding,
				Metadata:       backupDoc.Metadata,
				Image:          backupDoc.Image,
				ImageData:      backupDoc.ImageData,
				URL:            backupDoc.URL,
				ImageThumbnail: backupDoc.ImageThumbnail,
				ImageURL:       backupDoc.ImageURL,
				ImageMetadata:  backupDoc.ImageMetadata,
			}
			batch = append(batch, doc)
		}

		// Insert batch
		if err := vdbClient.CreateDocuments(ctx, collectionName, batch); err != nil {
			return fmt.Errorf("failed to insert batch at offset %d: %w", i, err)
		}

		restored += len(batch)

		if !restoreOpts.Quiet {
			progress := float64(restored) / float64(totalDocs) * 100
			fmt.Printf("\r   Progress: %d/%d documents (%.1f%%)", restored, totalDocs, progress)
		}
	}

	if !restoreOpts.Quiet {
		fmt.Println() // New line after progress
	}

	// Show summary
	duration := time.Since(startTime)

	if !restoreOpts.Quiet {
		fmt.Println()
		utils.PrintSuccess("✅ Restore completed successfully!")
		fmt.Println()
		utils.PrintInfo(fmt.Sprintf("   Collection: %s", collectionName))
		utils.PrintInfo(fmt.Sprintf("   Documents restored: %d", restored))
		utils.PrintInfo(fmt.Sprintf("   Duration: %.2fs", duration.Seconds()))
		utils.PrintInfo(fmt.Sprintf("   Throughput: %.0f docs/sec", float64(restored)/duration.Seconds()))
		fmt.Println()
		utils.PrintInfo(fmt.Sprintf("🎉 Collection '%s' is ready to use!", collectionName))
	}

	return nil
}
