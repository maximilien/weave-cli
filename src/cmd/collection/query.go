// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"github.com/spf13/cobra"
)

// QueryCmd represents the collection query command
var QueryCmd = &cobra.Command{
	Use:     "query COLLECTION \"query text\"",
	Aliases: []string{"q"},
	Short:   "Perform semantic search on a collection",
	Long: `Perform semantic search on a collection using natural language queries.

This command uses Weaviate's vector search capabilities to find the most relevant
documents based on semantic similarity to your query text.

Examples:
  weave cols query MyDocs "machine learning algorithms"
  weave cols q MyDocs "artificial intelligence" --top_k 10
  weave cols query WeaveImages "sunset over mountains" --top_k 3
  weave cols q MyDocs "exact keywords" --bm25
  weave cols q MyDocs "search term" --search-metadata --bm25`,
	Args: cobra.ExactArgs(2),
	Run:  runCollectionQuery,
}

func init() {
	QueryCmd.Flags().IntP("top_k", "k", 5, "Number of top results to return (default: 5)")
	QueryCmd.Flags().Float64P("distance", "d", 0.0, "Maximum distance threshold for results")
	QueryCmd.Flags().BoolP("search-metadata", "m", false, "Also search in metadata fields (default: false)")
	QueryCmd.Flags().Bool("bm25", false, "Use BM25 keyword search instead of semantic search (default: false)")
}

func runCollectionQuery(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	queryText := args[1]
	topK, _ := cmd.Flags().GetInt("top_k")
	distance, _ := cmd.Flags().GetFloat64("distance")
	searchMetadata, _ := cmd.Flags().GetBool("search-metadata")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	useBM25, _ := cmd.Flags().GetBool("bm25")

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database config
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	// Create query options
	options := weaviate.QueryOptions{
		TopK:           topK,
		Distance:       distance,
		SearchMetadata: searchMetadata,
		NoTruncate:     noTruncate,
		UseBM25:        useBM25,
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.QueryWeaviateCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMock:
		utils.QueryMockCollection(ctx, dbConfig, collectionName, queryText, options)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
