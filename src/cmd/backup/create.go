// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/version"
)

var createOpts = &backuppkg.CreateOptions{
	BatchSize: 100,
	Compress:  true,
}

// CreateCmd represents the backup create command
var CreateCmd = &cobra.Command{
	Use:   "create <collection> --output <file>",
	Short: "Backup a vector database collection",
	Long: `Export a vector database collection to a portable .weavebak file.

The backup includes:
  • All document embeddings (vectors)
  • All document content and metadata
  • Collection schema and configuration
  • Images (base64 and external URLs)

The backup file can be used to:
  • Restore the collection to the same VDB
  • Migrate to a different VDB type
  • Create snapshots for disaster recovery`,
	Example: `  # Backup collection to file
  weave backup create MyCollection --output backup.weavebak

  # Backup without compression
  weave backup create MyCollection --output backup.weavebak --no-compress

  # Backup with custom batch size
  weave backup create MyCollection --output backup.weavebak --batch-size 500

  # Backup specific VDB
  weave backup create MyCollection --vdb milvus-local --output backup.weavebak`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupCreate,
}

func init() {
	CreateCmd.Flags().StringVarP(&createOpts.OutputFile, "output", "o", "", "Output backup file path (required)")
	CreateCmd.Flags().BoolVar(&createOpts.Compress, "compress", true, "Compress backup with gzip")
	CreateCmd.Flags().BoolVar(&createOpts.Compress, "no-compress", false, "Disable compression")
	CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 100, "Documents per batch")
	CreateCmd.Flags().BoolVarP(&createOpts.Quiet, "quiet", "q", false, "Suppress progress output")

	CreateCmd.MarkFlagRequired("output")
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	collection := args[0]
	createOpts.Collection = collection

	// Add .gz extension if compressing and not already present
	if createOpts.Compress && !strings.HasSuffix(createOpts.OutputFile, ".gz") {
		createOpts.OutputFile += ".gz"
	}

	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig("", "")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get selected VDB
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		return fmt.Errorf("failed to determine vector databases: %w", err)
	}

	// Use first selected VDB (backup works on single VDB at a time)
	if len(selection.Configs) == 0 {
		return fmt.Errorf("no vector database configured")
	}
	vdbConfig := &selection.Configs[0]

	vdbClient, err := utils.CreateVectorDBClient(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}

	createOpts.VDBType = string(vdbConfig.Type)

	if !createOpts.Quiet {
		utils.PrintInfo(fmt.Sprintf("📦 Creating backup of collection: %s", collection))
		utils.PrintInfo(fmt.Sprintf("   VDB Type: %s", vdbConfig.Type))
		utils.PrintInfo(fmt.Sprintf("   Output: %s", createOpts.OutputFile))
		fmt.Println()
	}

	// Check if collection exists
	exists, err := vdbClient.CollectionExists(ctx, collection)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("collection '%s' does not exist", collection)
	}

	// Get collection info
	collections, err := vdbClient.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	var collectionInfo *vectordb.CollectionInfo
	for _, c := range collections {
		if c.Name == collection {
			collectionInfo = &c
			break
		}
	}

	if collectionInfo == nil {
		return fmt.Errorf("failed to get collection info")
	}

	totalDocs := int(collectionInfo.Count)
	if !createOpts.Quiet {
		utils.PrintInfo(fmt.Sprintf("ℹ️  Collection has %d documents", totalDocs))
		fmt.Println()
	}

	// Create backup format (use defaults for embedding model/dimensions if not available)
	embeddingModel := collectionInfo.Vectorizer
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	// Estimate vector dimensions from model name
	vectorDims := vdbConfig.VectorDimensions
	if vectorDims == 0 {
		vectorDims = 1536 // default
	}

	backup := backuppkg.NewBackupFormat(
		collection,
		string(vdbConfig.Type),
		embeddingModel,
		vectorDims,
	)

	// Start timing
	startTime := time.Now()

	// Export documents in batches
	if !createOpts.Quiet {
		utils.PrintInfo("📥 Exporting documents...")
	}

	exported := 0

	for offset := 0; offset < totalDocs || totalDocs == 0; offset += createOpts.BatchSize {
		// Query batch of documents
		docs, err := vdbClient.ListDocuments(ctx, collection, createOpts.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list documents at offset %d: %w", offset, err)
		}

		if len(docs) == 0 {
			break
		}

		// Convert to backup documents
		for _, doc := range docs {
			backupDoc := backuppkg.BackupDocument{
				ID:             doc.ID,
				Content:        doc.Content,
				Text:           doc.Text,
				Embedding:      doc.Embedding,
				Metadata:       doc.Metadata,
				Image:          doc.Image,
				ImageData:      doc.ImageData,
				URL:            doc.URL,
				ImageThumbnail: doc.ImageThumbnail,
				ImageURL:       doc.ImageURL,
				ImageMetadata:  doc.ImageMetadata,
			}
			backup.Documents = append(backup.Documents, backupDoc)
		}

		exported += len(docs)

		if !createOpts.Quiet && totalDocs > 0 {
			progress := float64(exported) / float64(totalDocs) * 100
			fmt.Printf("\r   Progress: %d/%d documents (%.1f%%)", exported, totalDocs, progress)
		}

		// Break if we got fewer docs than batch size (last batch)
		if len(docs) < createOpts.BatchSize {
			break
		}
	}

	if !createOpts.Quiet && totalDocs > 0 {
		fmt.Println() // New line after progress
	}

	// Update metadata
	backup.Metadata.TotalDocuments = len(backup.Documents)

	// Write backup to file
	if !createOpts.Quiet {
		utils.PrintInfo("💾 Writing backup file...")
	}

	if err := backuppkg.WriteBackup(backup, createOpts.OutputFile, createOpts.Compress); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(createOpts.OutputFile)
	if err == nil {
		backup.Metadata.BackupSizeBytes = fileInfo.Size()
	}

	// Show summary
	duration := time.Since(startTime)
	fmt.Println()
	utils.PrintSuccess(fmt.Sprintf("✅ Backup created successfully!"))
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("   Collection: %s", collection))
	utils.PrintInfo(fmt.Sprintf("   Documents: %d", len(backup.Documents)))
	utils.PrintInfo(fmt.Sprintf("   File size: %.2f MB", float64(backup.Metadata.BackupSizeBytes)/(1024*1024)))
	utils.PrintInfo(fmt.Sprintf("   Duration: %.2fs", duration.Seconds()))
	utils.PrintInfo(fmt.Sprintf("   Weave version: %s", version.Get().Version))
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("📁 Backup saved to: %s", createOpts.OutputFile))

	return nil
}
