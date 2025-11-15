// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
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

Score Interpretation:
  Scores are normalized to spread results across a wider range:
  - < 0.3: No good matches found (try rephrasing your query)
  - 0.3-0.5: Weak/marginal relevance
  - 0.5-0.7: Good semantic relevance
  - > 0.7: Strong semantic relevance

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
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
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
		JSONOutput:     jsonOutput,
	}

	// Use vector database selector based on flags to get the appropriate database
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector database: %v", err))
		os.Exit(1)
	}

	// Query operations should use the single selected database
	// If multiple databases selected, use the first one with the collection
	var dbConfig *config.VectorDBConfig
	if len(selection.Configs) == 1 {
		dbConfig = &selection.Configs[0]
	} else {
		// Try to find the collection in one of the selected databases
		for i := range selection.Configs {
			cfg := &selection.Configs[i]
			// For now, use the first database
			// In the future, we could check which database contains the collection
			dbConfig = cfg
			break
		}
	}

	if dbConfig == nil {
		utils.PrintError("No database configuration available")
		os.Exit(1)
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.QueryWeaviateCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeSupabase:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMongoDB:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMock:
		utils.QueryMockCollection(ctx, dbConfig, collectionName, queryText, options)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
