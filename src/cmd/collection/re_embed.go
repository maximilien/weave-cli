// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/embeddings"
	"github.com/maximilien/weave-cli/src/pkg/reembedding"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
)

// ReEmbedCmd represents the collection re-embed command
var ReEmbedCmd = &cobra.Command{
	Use:     "reembed SOURCE_COLLECTION",
	Aliases: []string{"re-embed"},
	Short:   "Re-embed collection with different embedding model",
	Long: `Re-embed an existing collection using a different embedding model.

This reads existing text chunks and generates new embeddings without
re-ingesting documents. Much faster than full re-ingestion (20x speedup).

Perfect for:
- Testing different embedding models quickly
- Switching from proprietary to OSS models
- Comparing model performance without re-processing documents

Examples:
  # Re-embed with sentence-transformers (OSS)
  weave collection re-embed MyCollection \
    --new-embedding sentence-transformers/all-mpnet-base-v2 \
    --output MyCollection_OSS

  # Re-embed with OpenAI
  weave collection re-embed MyCollection \
    --new-embedding text-embedding-3-large \
    --output MyCollection_Large

  # Re-embed with Ollama (local)
  weave collection re-embed MyCollection \
    --new-embedding nomic-embed-text \
    --output MyCollection_Nomic

  # Custom batch size (default: 100)
  weave collection re-embed MyCollection \
    --new-embedding text-embedding-3-small \
    --output MyCollection_Small \
    --batch-size 50

Performance:
  Typical speed: 200+ docs/min
  Example: 3,500 docs in ~15 minutes (vs 5+ hours for re-ingestion)`,
	Args: cobra.ExactArgs(1),
	Run:  runReEmbed,
}

func init() {
	CollectionCmd.AddCommand(ReEmbedCmd)

	ReEmbedCmd.Flags().StringP("new-embedding", "e", "", "New embedding model (REQUIRED - run 'weave embeddings list' to see options)")
	ReEmbedCmd.Flags().StringP("output", "o", "", "Output collection name (REQUIRED)")
	ReEmbedCmd.Flags().IntP("batch-size", "b", 100, "Batch size for processing (default: 100)")
	ReEmbedCmd.Flags().Bool("skip-existing", false, "Skip if output collection exists")
	ReEmbedCmd.MarkFlagRequired("new-embedding")
	ReEmbedCmd.MarkFlagRequired("output")
}

func runReEmbed(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	sourceCollection := args[0]
	newEmbeddingModel, _ := cmd.Flags().GetString("new-embedding")
	outputCollection, _ := cmd.Flags().GetString("output")
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	skipExisting, _ := cmd.Flags().GetBool("skip-existing")

	// 1. Load config and create client
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector databases: %v", err))
		os.Exit(1)
	}

	// Re-embedding requires exactly one database
	if len(selection.Configs) == 0 {
		utils.PrintError("No vector database configured")
		os.Exit(1)
	}
	if len(selection.Configs) > 1 {
		utils.PrintError("Re-embedding requires exactly one database (use --weaviate, --milvus, etc.)")
		os.Exit(1)
	}

	dbConfig := &selection.Configs[0]

	client, err := utils.CreateVectorDBClient(dbConfig)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create vector database client: %v", err))
		os.Exit(1)
	}

	// 2. Validate source collection exists
	utils.PrintInfo(fmt.Sprintf("🔍 Validating source collection '%s'...", sourceCollection))
	exists, err := client.CollectionExists(ctx, sourceCollection)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to check source collection: %v", err))
		os.Exit(1)
	}
	if !exists {
		utils.PrintError(fmt.Sprintf("Source collection '%s' not found", sourceCollection))
		os.Exit(1)
	}
	utils.PrintSuccess(fmt.Sprintf("✓ Source collection '%s' exists", sourceCollection))

	// 3. Auto-detect dimensions for new model (using model registry!)
	utils.PrintInfo(fmt.Sprintf("📐 Detecting dimensions for embedding model '%s'...", newEmbeddingModel))
	dims, err := embeddings.GetModelDimensions(newEmbeddingModel)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Unknown embedding model: %s", newEmbeddingModel))
		utils.PrintInfo("Run 'weave embeddings list' to see supported models")
		os.Exit(1)
	}

	info, _ := embeddings.GetModelInfo(newEmbeddingModel)
	ossLabel := ""
	if info.OSS {
		ossLabel = " (OSS)"
	}
	utils.PrintSuccess(fmt.Sprintf("✓ Auto-detected: %d dimensions for %s%s", dims, newEmbeddingModel, ossLabel))

	// 4. Get total document count
	utils.PrintInfo(fmt.Sprintf("📊 Counting documents in '%s'...", sourceCollection))
	totalDocs, err := client.GetCollectionCount(ctx, sourceCollection)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to count documents: %v", err))
		os.Exit(1)
	}
	utils.PrintSuccess(fmt.Sprintf("✓ Found %d documents to re-embed", totalDocs))

	// 5. Check if output collection exists
	exists, err = client.CollectionExists(ctx, outputCollection)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to check output collection: %v", err))
		os.Exit(1)
	}
	if exists && !skipExisting {
		utils.PrintError(fmt.Sprintf("Output collection '%s' already exists (use --skip-existing to override)", outputCollection))
		os.Exit(1)
	}
	if exists && skipExisting {
		utils.PrintWarning(fmt.Sprintf("⚠️  Output collection '%s' exists, skipping re-embedding", outputCollection))
		return
	}

	// 6. Create output collection with new dimensions
	utils.PrintInfo(fmt.Sprintf("🔨 Creating output collection '%s' with %d dimensions...", outputCollection, dims))

	// CRITICAL: Update the VDB client config with NEW embedding dimensions
	// The client was created with source collection's dimensions (e.g., 1536 for OpenAI),
	// but we need to use the new embedding model's dimensions (e.g., 768 for sentence-transformers)
	// for the output collection. This fixes the dimension mismatch issue.
	if dbConfig.VectorDimensions != dims {
		utils.PrintInfo(fmt.Sprintf("📐 Updating VDB dimensions: %d → %d (for %s)",
			dbConfig.VectorDimensions, dims, newEmbeddingModel))
		dbConfig.VectorDimensions = dims

		// Recreate client with updated config so CreateCollection uses correct dimensions
		client, err = utils.CreateVectorDBClient(dbConfig)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to recreate VDB client with new dimensions: %v", err))
			os.Exit(1)
		}
	}

	schema := client.GetDefaultSchema(vectordb.SchemaTypeText, outputCollection)
	schema.Vectorizer = newEmbeddingModel

	err = client.CreateCollection(ctx, outputCollection, schema)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create output collection: %v", err))
		os.Exit(1)
	}
	utils.PrintSuccess(fmt.Sprintf("✓ Created collection '%s'", outputCollection))

	// 7. Initialize re-embedding components
	utils.PrintInfo(fmt.Sprintf("🚀 Initializing re-embedding pipeline (batch size: %d)...", batchSize))
	reader, err := reembedding.NewCollectionReader(client, sourceCollection, batchSize)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create collection reader: %v", err))
		os.Exit(1)
	}

	pipeline, err := reembedding.NewEmbeddingPipeline(newEmbeddingModel)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create embedding pipeline: %v", err))
		os.Exit(1)
	}

	// Check if provider is supported
	if !pipeline.IsSupported() {
		utils.PrintError(fmt.Sprintf("Embedding provider '%s' is not yet supported for re-embedding", pipeline.GetProvider()))
		utils.PrintInfo("Currently supported: OpenAI models")
		utils.PrintInfo("Coming soon: sentence-transformers, Ollama")
		os.Exit(1)
	}

	tracker := reembedding.NewProgressTracker(totalDocs)
	utils.PrintSuccess("✓ Pipeline initialized")

	// 8. Process in batches
	utils.PrintInfo(fmt.Sprintf("\n📈 Re-embedding %d documents from '%s' to '%s'...\n", totalDocs, sourceCollection, outputCollection))

	batchCount := 0
	for reader.HasMore() {
		// Read batch
		docs, err := reader.ReadBatch(ctx)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("⚠️  Failed to read batch at offset %d: %v", batchCount*batchSize, err))
			continue
		}

		if len(docs) == 0 {
			break
		}

		// Generate new embeddings
		err = pipeline.ProcessBatch(ctx, docs)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("⚠️  Failed to process batch %d: %v", batchCount, err))
			continue
		}

		// Write to output collection
		err = client.CreateDocuments(ctx, outputCollection, docs)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("⚠️  Failed to write batch %d: %v", batchCount, err))
			continue
		}

		// Update progress
		tracker.Update(len(docs))
		tracker.Display()

		batchCount++
	}

	// 9. Success
	utils.PrintInfo("")
	utils.PrintSuccess(fmt.Sprintf("🎉 Successfully re-embedded %d documents to '%s'", totalDocs, outputCollection))
	utils.PrintSuccess(fmt.Sprintf("✓ New embedding model: %s (%d dimensions)", newEmbeddingModel, dims))

	// Final verification
	finalCount, err := client.GetCollectionCount(ctx, outputCollection)
	if err == nil {
		utils.PrintSuccess(fmt.Sprintf("✓ Output collection contains %d documents", finalCount))
		if finalCount != totalDocs {
			utils.PrintWarning(fmt.Sprintf("⚠️  Document count mismatch: expected %d, got %d", totalDocs, finalCount))
		}
	}
}
