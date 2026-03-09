// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"path/filepath"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

var (
	ingestEmbedding      string
	ingestChunkSize      int
	ingestParallelWorker int
	ingestBatchSize      int
	ingestType           string
	ingestRestartEvery   int
	ingestRestartDelay   int
	ingestAll            bool
	ingestParallel       int
)

// IngestCmd represents the ingest command
var IngestCmd = &cobra.Command{
	Use:   "ingest [collection-name] [data-path]",
	Short: "Ingest data into stack collection(s)",
	Long: `Ingest data into a collection in the deployed stack.

This command automatically connects to the Milvus instance running in your
stack and ingests documents using the existing weave ingestion pipeline.

The command will:
  1. Detect active stack from .weave-state/
  2. Port-forward to Milvus (if needed)
  3. Create collection (if doesn't exist)
  4. Ingest documents with progress tracking
  5. Show ingestion summary

Examples:
  # Ingest PDFs into Documents collection
  weave stack ingest Documents data/pdfs/

  # Ingest images with custom embedding
  weave stack ingest Images data/images/ \
    --type image \
    --embedding text-embedding-3-small

  # Parallel ingestion with custom chunk size
  weave stack ingest Docs data/ \
    --chunk-size 1000 \
    --parallel-workers 4 \
    --batch-size 20

  # Ingest all collections from weave-stack.yaml
  weave stack ingest --all

  # Ingest all with restart protection
  weave stack ingest --all --restart-every 3`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runIngest,
}

func init() {
	IngestCmd.Flags().StringVar(&ingestEmbedding, "embedding", "text-embedding-3-small", "Embedding model to use")
	IngestCmd.Flags().IntVar(&ingestChunkSize, "chunk-size", 500, "Chunk size for text splitting")
	IngestCmd.Flags().IntVar(&ingestParallelWorker, "parallel-workers", 1, "Number of parallel workers")
	IngestCmd.Flags().IntVar(&ingestBatchSize, "batch-size", 10, "Batch size for ingestion")
	IngestCmd.Flags().StringVar(&ingestType, "type", "text", "Collection type (text or image)")
	IngestCmd.Flags().IntVar(&ingestRestartEvery, "restart-every", 0, "Restart Milvus every N files (0=no restart, helps prevent OOM)")
	IngestCmd.Flags().IntVar(&ingestRestartDelay, "restart-delay", 15, "Seconds to wait after Milvus restart")
	IngestCmd.Flags().BoolVar(&ingestAll, "all", false, "Ingest all collections from weave-stack.yaml")
	IngestCmd.Flags().IntVar(&ingestParallel, "parallel", 0, "Process N collections concurrently (0=sequential)")
}

func runIngest(cmd *cobra.Command, args []string) error {
	// Handle --all flag
	if ingestAll {
		return runIngestAll(cmd, args)
	}

	// Regular single-collection ingestion
	if len(args) != 2 {
		return fmt.Errorf("requires collection name and data path, or use --all flag")
	}

	collectionName := args[0]
	dataPath := args[1]

	// Resolve absolute path
	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	utils.PrintInfo(fmt.Sprintf("Ingesting data from: %s", absPath))
	utils.PrintInfo(fmt.Sprintf("Collection: %s", collectionName))
	utils.PrintInfo(fmt.Sprintf("Type: %s", ingestType))
	utils.PrintInfo(fmt.Sprintf("Embedding: %s", ingestEmbedding))
	fmt.Println()

	// Load cluster info to verify stack is running
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found\n\nRun: weave stack up --runtime kind")
	}

	utils.PrintInfo(fmt.Sprintf("Stack: %s", clusterInfo.Name))
	utils.PrintInfo(fmt.Sprintf("Context: %s", clusterInfo.Context))
	fmt.Println()

	// Start port forwarding to Milvus
	utils.PrintInfo("Setting up connection to Milvus...")

	portForwardCtx, err := stackpkg.StartMilvusPortForward(clusterInfo)
	if err != nil {
		return fmt.Errorf("failed to start port forwarding: %w", err)
	}
	defer portForwardCtx.Stop()

	utils.PrintSuccess("✅ Connected to Milvus")
	fmt.Println()

	// Run ingestion using existing pipeline
	utils.PrintInfo("Starting ingestion pipeline...")

	if ingestRestartEvery > 0 {
		utils.PrintInfo(fmt.Sprintf("🔄 Restart enabled: Milvus will restart every %d files", ingestRestartEvery))
	}

	err = stackpkg.IngestToStack(stackpkg.IngestConfig{
		CollectionName:  collectionName,
		DataPath:        absPath,
		Type:            ingestType,
		EmbeddingModel:  ingestEmbedding,
		ChunkSize:       ingestChunkSize,
		ParallelWorkers: ingestParallelWorker,
		BatchSize:       ingestBatchSize,
		MilvusLocalPort: portForwardCtx.LocalPort,
		RestartEvery:    ingestRestartEvery,
		RestartDelay:    ingestRestartDelay,
		ClusterInfo:     clusterInfo,
		PortForwardCtx:  &portForwardCtx,
	})

	if err != nil {
		return fmt.Errorf("ingestion failed: %w", err)
	}

	fmt.Println()
	utils.PrintSuccess("🎉 Ingestion complete!")

	return nil
}

// runIngestAll ingests all collections from weave-stack.yaml
func runIngestAll(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Loading stack configuration...")

	// Load stack config
	stackConfig, err := stackpkg.LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	if stackConfig.Ingestion == nil || len(stackConfig.Ingestion.Collections) == 0 {
		return fmt.Errorf("no collections defined in weave-stack.yaml ingestion section")
	}

	utils.PrintInfo(fmt.Sprintf("Found %d collections to ingest", len(stackConfig.Ingestion.Collections)))
	fmt.Println()

	// Load cluster info to verify stack is running
	clusterInfo, err := stackpkg.LoadClusterInfo()
	if err != nil {
		return fmt.Errorf("failed to load cluster info: %w", err)
	}

	if clusterInfo == nil {
		return fmt.Errorf("no active stack found\n\nRun: weave stack up --runtime kind")
	}

	utils.PrintInfo(fmt.Sprintf("Stack: %s", clusterInfo.Name))
	utils.PrintInfo(fmt.Sprintf("Context: %s", clusterInfo.Context))
	fmt.Println()

	// Start port forwarding to Milvus
	utils.PrintInfo("Setting up connection to Milvus...")

	portForwardCtx, err := stackpkg.StartMilvusPortForward(clusterInfo)
	if err != nil {
		return fmt.Errorf("failed to start port forwarding: %w", err)
	}
	defer portForwardCtx.Stop()

	utils.PrintSuccess("✅ Connected to Milvus")
	fmt.Println()

	// Process collections
	totalCollections := len(stackConfig.Ingestion.Collections)
	processedCount := 0
	failedCount := 0

	for collectionName, collectionConfig := range stackConfig.Ingestion.Collections {
		processedCount++

		utils.PrintInfo(fmt.Sprintf("[%d/%d] Processing collection: %s", processedCount, totalCollections, collectionName))
		utils.PrintInfo(fmt.Sprintf("  Source: %s", collectionConfig.Source))
		fmt.Println()

		// Resolve absolute path
		absPath, err := filepath.Abs(collectionConfig.Source)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to resolve path for %s: %v", collectionName, err))
			failedCount++
			continue
		}

		// Build ingestion config from YAML
		ingestCfg := stackpkg.IngestConfig{
			CollectionName:  collectionName,
			DataPath:        absPath,
			Type:            getConfigValue(collectionConfig.Type, "text"),
			EmbeddingModel:  getConfigValue(collectionConfig.Embedding, "text-embedding-3-small"),
			ChunkSize:       getConfigIntValue(collectionConfig.ChunkSize, 500),
			ParallelWorkers: getConfigIntValue(collectionConfig.Workers, 1),
			BatchSize:       getConfigIntValue(collectionConfig.BatchSize, 10),
			MilvusLocalPort: portForwardCtx.LocalPort,
			RestartEvery:    getConfigIntValue(collectionConfig.RestartEvery, ingestRestartEvery),
			RestartDelay:    ingestRestartDelay,
			ClusterInfo:     clusterInfo,
			PortForwardCtx:  &portForwardCtx,
		}

		// Run ingestion
		err = stackpkg.IngestToStack(ingestCfg)
		if err != nil {
			utils.PrintError(fmt.Sprintf("❌ Failed to ingest %s: %v", collectionName, err))
			failedCount++
			continue
		}

		utils.PrintSuccess(fmt.Sprintf("✅ Completed: %s", collectionName))
		fmt.Println()
	}

	// Summary
	fmt.Println()
	utils.PrintInfo("═══════════════════════════════════════")
	utils.PrintInfo("  INGESTION SUMMARY")
	utils.PrintInfo("═══════════════════════════════════════")
	utils.PrintInfo(fmt.Sprintf("Total collections: %d", totalCollections))
	utils.PrintInfo(fmt.Sprintf("Successful: %d", processedCount-failedCount))
	if failedCount > 0 {
		utils.PrintWarning(fmt.Sprintf("Failed: %d", failedCount))
	}
	fmt.Println()

	if failedCount > 0 {
		return fmt.Errorf("%d collections failed to ingest", failedCount)
	}

	utils.PrintSuccess("🎉 All collections ingested successfully!")
	return nil
}

// Helper functions to get config values with defaults
func getConfigValue(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func getConfigIntValue(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}
