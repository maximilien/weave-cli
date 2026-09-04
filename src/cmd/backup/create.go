// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/version"
)

var (
	createOpts = &backuppkg.CreateOptions{
		BatchSize: 200, // 2x faster based on v0.11.3 profiling (Mar 15, 2026)
		Compress:  true,
	}

	// Remote storage flags
	remoteStorage   string
	s3Bucket        string
	s3Region        string
	s3Endpoint      string
	s3AccessKey     string
	s3SecretKey     string
	s3Prefix        string
	s3NoSSL         bool
	remoteOnly      bool
	remoteKeepLocal bool
)

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
  weave backup create MyCollection --vdb milvus-local --output backup.weavebak

  # Backup to S3
  weave backup create MyCollection --output backup.weavebak \
    --remote-storage s3 \
    --s3-bucket my-backups \
    --s3-region us-east-1

  # Backup to MinIO
  weave backup create MyCollection --output backup.weavebak \
    --remote-storage minio \
    --s3-bucket backups \
    --s3-endpoint localhost:9000 \
    --s3-no-ssl

  # Backup to S3 only (skip local file)
  weave backup create MyCollection --output backup.weavebak \
    --remote-storage s3 \
    --s3-bucket my-backups \
    --remote-only`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupCreate,
}

func init() {
	CreateCmd.Flags().StringVarP(&createOpts.OutputFile, "output", "o", "", "Output backup file path (required)")
	CreateCmd.Flags().BoolVar(&createOpts.Compress, "compress", true, "Compress backup with gzip")
	CreateCmd.Flags().BoolVar(&createOpts.Compress, "no-compress", false, "Disable compression")
	CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 200, "Documents per batch (default optimized for performance)")
	CreateCmd.Flags().BoolVarP(&createOpts.Quiet, "quiet", "q", false, "Suppress progress output")

	// Remote storage flags
	CreateCmd.Flags().StringVar(&remoteStorage, "remote-storage", "", "Remote storage type (s3 or minio)")
	CreateCmd.Flags().StringVar(&s3Bucket, "s3-bucket", "", "S3/MinIO bucket name")
	CreateCmd.Flags().StringVar(&s3Region, "s3-region", "us-east-1", "AWS region (S3 only)")
	CreateCmd.Flags().StringVar(&s3Endpoint, "s3-endpoint", "", "MinIO endpoint (e.g., localhost:9000)")
	CreateCmd.Flags().StringVar(&s3AccessKey, "s3-access-key", "", "S3/MinIO access key (or AWS_ACCESS_KEY_ID env)")
	CreateCmd.Flags().StringVar(&s3SecretKey, "s3-secret-key", "", "S3/MinIO secret key (or AWS_SECRET_ACCESS_KEY env)")
	CreateCmd.Flags().StringVar(&s3Prefix, "s3-prefix", "", "Path prefix in bucket (e.g., 'backups/')")
	CreateCmd.Flags().BoolVar(&s3NoSSL, "s3-no-ssl", false, "Disable SSL/TLS (MinIO only)")
	CreateCmd.Flags().BoolVar(&remoteOnly, "remote-only", false, "Upload to remote storage only (skip local file)")
	CreateCmd.Flags().BoolVar(&remoteKeepLocal, "remote-keep-local", true, "Keep local file after upload")

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

	// We'll detect actual vector dimensions from first document
	// Start with config default, but update after first batch
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

			// Detect actual vector dimensions from first document with embedding
			if exported == 0 && len(doc.Embedding) > 0 {
				actualDims := len(doc.Embedding)
				if actualDims != vectorDims {
					backup.Metadata.VectorDimensions = actualDims
					vectorDims = actualDims
				}
			}
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
	utils.PrintSuccess("✅ Backup created successfully!")
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("   Collection: %s", collection))
	utils.PrintInfo(fmt.Sprintf("   Documents: %d", len(backup.Documents)))
	utils.PrintInfo(fmt.Sprintf("   File size: %.2f MB", float64(backup.Metadata.BackupSizeBytes)/(1024*1024)))
	utils.PrintInfo(fmt.Sprintf("   Duration: %.2fs", duration.Seconds()))
	utils.PrintInfo(fmt.Sprintf("   Weave version: %s", version.Get().Version))
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("📁 Backup saved to: %s", createOpts.OutputFile))

	// Upload to remote storage if configured
	if remoteStorage != "" {
		if err := uploadToRemoteStorage(ctx, createOpts.OutputFile); err != nil {
			return fmt.Errorf("failed to upload to remote storage: %w", err)
		}

		// Delete local file if --remote-only was specified
		if remoteOnly && !remoteKeepLocal {
			if err := os.Remove(createOpts.OutputFile); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to delete local file: %v", err))
			} else {
				utils.PrintInfo("🗑️  Deleted local backup file (remote-only mode)")
			}
		}
	}

	return nil
}

// uploadToRemoteStorage uploads a backup file to S3/MinIO
func uploadToRemoteStorage(ctx context.Context, filePath string) error {
	// Get credentials from env vars if not provided via flags
	accessKey := s3AccessKey
	secretKey := s3SecretKey

	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	// Validate configuration
	if s3Bucket == "" {
		return fmt.Errorf("--s3-bucket is required for remote storage")
	}
	if accessKey == "" || secretKey == "" {
		return fmt.Errorf("S3 credentials required (use --s3-access-key/--s3-secret-key or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars)")
	}

	// Create remote storage config
	remoteCfg := &backuppkg.RemoteStorageConfig{
		Type:            remoteStorage,
		Bucket:          s3Bucket,
		Region:          s3Region,
		Endpoint:        s3Endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UseSSL:          !s3NoSSL,
		PathPrefix:      s3Prefix,
	}

	utils.PrintInfo(fmt.Sprintf("☁️  Uploading to %s...", remoteStorage))
	utils.PrintInfo(fmt.Sprintf("   Bucket: %s", s3Bucket))
	if s3Prefix != "" {
		utils.PrintInfo(fmt.Sprintf("   Prefix: %s", s3Prefix))
	}

	// Create remote storage client
	remote, err := backuppkg.NewRemoteStorage(remoteCfg)
	if err != nil {
		return fmt.Errorf("failed to create remote storage client: %w", err)
	}

	// Read backup file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Upload to remote storage
	// Use filename as the key (within prefix if specified)
	key := filepath.Base(filePath)
	if err := remote.Upload(ctx, key, data); err != nil {
		return err
	}

	fmt.Println()
	utils.PrintSuccess("✅ Upload complete!")
	utils.PrintInfo(fmt.Sprintf("   Remote path: %s/%s", s3Bucket, key))
	if s3Prefix != "" {
		utils.PrintInfo(fmt.Sprintf("   Full path: %s/%s/%s", s3Bucket, s3Prefix, key))
	}
	fmt.Println()

	return nil
}
