// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/config"
)

var (
	backupOutputFile string
	backupCompress   bool
	backupAllCollections bool
	backupBatchSize  int
)

// BackupCmd represents the stack backup command
var BackupCmd = &cobra.Command{
	Use:   "backup <collection> --output <file>",
	Short: "Backup a collection from the stack's vector database",
	Long: `Backup a collection from the weave stack's Milvus instance to a portable .weavebak file.

This command connects to the stack's Milvus instance via port-forwarding and exports
all documents from the specified collection.

The backup includes:
  • All document embeddings (vectors)
  • All document content and metadata
  • Images and external storage references

Use case: Create snapshots of your stack collections for disaster recovery or migration.`,
	Example: `  # Backup a stack collection
  weave stack backup Documents --output backups/docs.weavebak

  # Backup without compression
  weave stack backup Images --output backups/images.weavebak --no-compress

  # Backup all collections
  weave stack backup --all --output backups/

  # Backup with custom batch size
  weave stack backup Documents --output backup.weavebak --batch-size 500`,
	Args: func(cmd *cobra.Command, args []string) error {
		if backupAllCollections {
			return nil // --all flag doesn't need collection argument
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: runStackBackup,
}

func init() {
	BackupCmd.Flags().StringVarP(&backupOutputFile, "output", "o", "", "Output backup file or directory path (required)")
	BackupCmd.Flags().BoolVar(&backupCompress, "compress", true, "Compress backup with gzip")
	BackupCmd.Flags().BoolVar(&backupCompress, "no-compress", false, "Disable compression")
	BackupCmd.Flags().BoolVar(&backupAllCollections, "all", false, "Backup all collections")
	BackupCmd.Flags().IntVar(&backupBatchSize, "batch-size", 100, "Documents per batch")

	BackupCmd.MarkFlagRequired("output")
}

func runStackBackup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var collectionName string
	if !backupAllCollections {
		collectionName = args[0]
	}

	// Load stack configuration
	stackConfig, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	utils.PrintInfo(fmt.Sprintf("📦 Backing up from stack: %s", stackConfig.Name))
	fmt.Println()

	// Get cluster info
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	// Start port-forward to Milvus
	utils.PrintInfo("Setting up connection to stack Milvus...")
	portForward, err := stackpkg.StartMilvusPortForward(clusterInfo)
	if err != nil {
		return fmt.Errorf("failed to start port-forward: %w", err)
	}
	defer portForward.Stop()

	utils.PrintSuccess("Connected to stack Milvus")
	fmt.Println()

	// Create Milvus VDB client
	milvusConfig := &config.VectorDBConfig{
		Type:             "milvus-local",
		Name:             "stack-milvus",
		Address:          fmt.Sprintf("localhost:%d", portForward.LocalPort),
		VectorDimensions: 1536,
		Timeout:          30,
	}

	vdbClient, err := utils.CreateVectorDBClient(milvusConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}

	if backupAllCollections {
		return backupAllStackCollections(ctx, vdbClient, backupOutputFile)
	}

	return backupStackCollection(ctx, vdbClient, collectionName, backupOutputFile)
}

func backupStackCollection(ctx context.Context, vdbClient vectordb.VectorDBClient, collectionName, outputFile string) error {
	// Check if collection exists
	exists, err := vdbClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("collection '%s' does not exist in stack", collectionName)
	}

	// Get collection info
	collections, err := vdbClient.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	var collectionInfo *vectordb.CollectionInfo
	for _, c := range collections {
		if c.Name == collectionName {
			collectionInfo = &c
			break
		}
	}

	if collectionInfo == nil {
		return fmt.Errorf("failed to get collection info")
	}

	totalDocs := int(collectionInfo.Count)
	utils.PrintInfo(fmt.Sprintf("Collection: %s (%d documents)", collectionName, totalDocs))
	fmt.Println()

	// Create backup
	embeddingModel := collectionInfo.Vectorizer
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	backup := backuppkg.NewBackupFormat(
		collectionName,
		"milvus-local",
		embeddingModel,
		1536,
	)

	// Export documents
	utils.PrintInfo("📥 Exporting documents...")
	startTime := time.Now()
	exported := 0

	for offset := 0; offset < totalDocs || totalDocs == 0; offset += backupBatchSize {
		docs, err := vdbClient.ListDocuments(ctx, collectionName, backupBatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to list documents at offset %d: %w", offset, err)
		}

		if len(docs) == 0 {
			break
		}

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

		if totalDocs > 0 {
			progress := float64(exported) / float64(totalDocs) * 100
			fmt.Printf("\r   Progress: %d/%d documents (%.1f%%)", exported, totalDocs, progress)
		}

		if len(docs) < backupBatchSize {
			break
		}
	}

	if totalDocs > 0 {
		fmt.Println() // New line after progress
	}

	backup.Metadata.TotalDocuments = len(backup.Documents)

	// Write backup
	utils.PrintInfo("💾 Writing backup file...")
	if err := backuppkg.WriteBackup(backup, outputFile, backupCompress); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(outputFile)
	if err == nil {
		backup.Metadata.BackupSizeBytes = fileInfo.Size()
	}

	// Show summary
	duration := time.Since(startTime)
	fmt.Println()
	utils.PrintSuccess("✅ Stack backup created successfully!")
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("   Collection: %s", collectionName))
	utils.PrintInfo(fmt.Sprintf("   Documents: %d", len(backup.Documents)))
	utils.PrintInfo(fmt.Sprintf("   File size: %.2f MB", float64(backup.Metadata.BackupSizeBytes)/(1024*1024)))
	utils.PrintInfo(fmt.Sprintf("   Duration: %.2fs", duration.Seconds()))
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("📁 Backup saved to: %s", outputFile))

	return nil
}

func backupAllStackCollections(ctx context.Context, vdbClient vectordb.VectorDBClient, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// List all collections
	collections, err := vdbClient.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) == 0 {
		utils.PrintWarning("No collections found in stack")
		return nil
	}

	utils.PrintInfo(fmt.Sprintf("Found %d collections to backup", len(collections)))
	fmt.Println()

	// Backup each collection
	for i, collection := range collections {
		utils.PrintInfo(fmt.Sprintf("[%d/%d] Backing up: %s", i+1, len(collections), collection.Name))

		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.weavebak", collection.Name))
		if backupCompress {
			outputFile += ".gz"
		}

		if err := backupStackCollection(ctx, vdbClient, collection.Name, outputFile); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to backup %s: %v", collection.Name, err))
			continue
		}

		fmt.Println()
	}

	utils.PrintSuccess(fmt.Sprintf("✅ Backed up %d collections to: %s", len(collections), outputDir))

	return nil
}
