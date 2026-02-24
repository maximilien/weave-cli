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
)

// IngestCmd represents the ingest command
var IngestCmd = &cobra.Command{
	Use:   "ingest <collection-name> <data-path>",
	Short: "Ingest data into stack collection",
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
    --batch-size 20`,
	Args: cobra.ExactArgs(2),
	RunE: runIngest,
}

func init() {
	IngestCmd.Flags().StringVar(&ingestEmbedding, "embedding", "text-embedding-3-small", "Embedding model to use")
	IngestCmd.Flags().IntVar(&ingestChunkSize, "chunk-size", 500, "Chunk size for text splitting")
	IngestCmd.Flags().IntVar(&ingestParallelWorker, "parallel-workers", 1, "Number of parallel workers")
	IngestCmd.Flags().IntVar(&ingestBatchSize, "batch-size", 10, "Batch size for ingestion")
	IngestCmd.Flags().StringVar(&ingestType, "type", "text", "Collection type (text or image)")
}

func runIngest(cmd *cobra.Command, args []string) error {
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

	err = stackpkg.IngestToStack(stackpkg.IngestConfig{
		CollectionName:  collectionName,
		DataPath:        absPath,
		Type:            ingestType,
		EmbeddingModel:  ingestEmbedding,
		ChunkSize:       ingestChunkSize,
		ParallelWorkers: ingestParallelWorker,
		BatchSize:       ingestBatchSize,
		MilvusLocalPort: portForwardCtx.LocalPort,
	})

	if err != nil {
		return fmt.Errorf("ingestion failed: %w", err)
	}

	fmt.Println()
	utils.PrintSuccess("🎉 Ingestion complete!")

	return nil
}
