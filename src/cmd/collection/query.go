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
	Use:     "query COLLECTION [COLLECTION...] \"query text\"",
	Aliases: []string{"q"},
	Short:   "Perform semantic search on one or more collections",
	Long: `Perform semantic search on one or more collections using natural language queries.

This command uses vector search capabilities to find the most relevant
documents based on semantic similarity to your query text.

Multi-Collection Queries:
  Query multiple collections at once and aggregate results. Each collection
  returns its top K results, which are then combined and passed to the agent.
  This is useful for searching across different data types (e.g., documents,
  images, metadata) in a unified way.

  For multi-modal queries (text + images), use --top_k_images to ensure image
  results are included. By default, text documents often have higher similarity
  scores than images, causing images to be filtered out. The --top_k_images flag
  guarantees that image collections return their top K results:
    weave cols query WeaveDocs WeaveImages "screenshot" --top_k 5 --top_k_images 2
  This returns top 5 from WeaveDocs + top 2 from WeaveImages, merged by score.

Cross-VDB Queries:
  Query collections from different vector databases in a single command using
  the "Collection:vdb-key" syntax (e.g., "Docs:weaviate-local Images:milvus-cloud").
  Supported VDB keys: weaviate-local, weaviate-cloud, milvus-local, milvus-cloud,
  mongodb-cloud, qdrant-local, qdrant-cloud, neo4j-local, neo4j-cloud,
  chroma-local, chroma-cloud, supabase-cloud, etc.
  Results include collection name and VDB information in citations.

For image collections with multi-vector support, you can search by:
  - Text metadata (captions, OCR, context) using --search-type text (default)
  - Visual similarity using --search-type visual (requires multi2vec-clip)
  - Specific named vector using --vector <name>

Score Interpretation:
  Scores are normalized to spread results across a wider range:
  - < 0.3: No good matches found (try rephrasing your query)
  - 0.3-0.5: Weak/marginal relevance
  - 0.5-0.7: Good semantic relevance
  - > 0.7: Strong semantic relevance

RAG Agents:
  Use the --agent flag to process query results through a custom RAG agent.
  Available agents: rag-agent, summarize-agent, qa-agent
  Agents provide comprehensive answers with citations, summaries, or precise Q&A.
  Requires: OPENAI_API_KEY environment variable

Examples:
  # Single collection search
  weave cols query MyDocs "machine learning algorithms"
  weave cols q MyDocs "artificial intelligence" --top_k 10

  # Multi-collection search (returns top K from EACH collection)
  weave cols query WeaveDocs WeaveImages "weave cli" --agent rag-agent --top_k 3
  weave cols query AuctionListings AuctionResults AuctionImages "vintage cars" --agent rag-agent

  # Multi-modal queries with guaranteed image results
  weave cols query WeaveDocs WeaveImages "screenshot" --agent rag-agent --top_k 5 --top_k_images 2
  weave cols query ProductDocs ProductImages "red vintage car" --agent rag-agent --top_k 10 --top_k_images 3

  # Cross-VDB queries (query collections from different VDBs)
  weave cols query WeaveDocs:weaviate-local WeaveImages:milvus-local "weave cli" --agent rag-agent
  weave cols query AuctionDocs:mongodb AuctionImages:weaviate-cloud "vintage" --agent rag-agent --top_k 3

  # RAG agent query with comprehensive answer
  weave cols query MyDocs "What is machine learning?" --agent rag-agent
  weave cols q MyDocs "Summarize the main topics" --agent summarize-agent --top_k 10

  # Show progress during query execution
  weave cols query MyDocs "complex query" --agent rag-agent --progress

  # JSON output with progress (outputs JSON Lines format)
  weave cols query MyDocs "query" --agent rag-agent --json --progress

  # Image search by text metadata (captions, OCR)
  weave cols query MyImages "sunset over mountains" --top_k 3
  weave cols query MyImages "camera" --search-type text

  # Visual similarity search (CLIP)
  weave cols query MyImages "product photo" --search-type visual
  weave cols query MyImages "" --vector image_vector --top_k 5

  # Keyword search
  weave cols q MyDocs "exact keywords" --bm25
  weave cols q MyDocs "search term" --search-metadata --bm25`,
	Args: cobra.MinimumNArgs(2),
	Run:  runCollectionQuery,
}

func init() {
	QueryCmd.Flags().IntP("top_k", "k", 5, "Number of top results to return (default: 5)")
	QueryCmd.Flags().Int("top_k_images", 0, "Number of top results from image collections (0 = use top_k)")
	QueryCmd.Flags().Float64P("distance", "d", 0.0, "Maximum distance threshold for results")
	QueryCmd.Flags().BoolP("search-metadata", "m", false, "Also search in metadata fields (default: false)")
	QueryCmd.Flags().Bool("bm25", false, "Use BM25 keyword search instead of semantic search (default: false)")
	QueryCmd.Flags().StringP("output", "o", "text", "Output format: text, json, yaml")
	QueryCmd.Flags().String("vector", "", "Named vector to search (e.g., 'text_vector', 'image_vector')")
	QueryCmd.Flags().String("search-type", "", "Search type: 'text' (text metadata) or 'visual' (image content)")
	QueryCmd.Flags().String("image", "", "Path to image file for visual similarity search (base64 encoded)")
	QueryCmd.Flags().String("agent", "", "Agent to use for processing results (e.g., 'rag-agent', 'qa-agent', 'summarize-agent')")
	QueryCmd.Flags().StringSlice("agents", []string{}, "Multiple agents to execute in sequence (e.g., 'rag-agent,search-agent')")
	QueryCmd.Flags().BoolP("verbose", "v", false, "Enable verbose debug logging")
	QueryCmd.Flags().Bool("progress", false, "Show progress during query execution")
}

func runCollectionQuery(cmd *cobra.Command, args []string) {
	// Parse collection specs and query text
	// Last arg is always the query text, all previous args are collection specs
	collectionSpecStrings := args[:len(args)-1]
	queryText := args[len(args)-1]

	// Parse collection specs (supports "Collection" or "Collection:vdb-key" syntax)
	collectionSpecs, err := utils.ParseCollectionSpecs(collectionSpecStrings)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Invalid collection specification: %v", err))
		os.Exit(1)
	}
	topK, _ := cmd.Flags().GetInt("top_k")
	topKImages, _ := cmd.Flags().GetInt("top_k_images")
	distance, _ := cmd.Flags().GetFloat64("distance")
	searchMetadata, _ := cmd.Flags().GetBool("search-metadata")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	useBM25, _ := cmd.Flags().GetBool("bm25")
	outputFormat, _ := cmd.Flags().GetString("output")
	vectorName, _ := cmd.Flags().GetString("vector")
	searchType, _ := cmd.Flags().GetString("search-type")
	imagePath, _ := cmd.Flags().GetString("image")
	agentName, _ := cmd.Flags().GetString("agent")
	agentNames, _ := cmd.Flags().GetStringSlice("agents")
	verbose, _ := cmd.Flags().GetBool("verbose")
	progress, _ := cmd.Flags().GetBool("progress")

	// Handle agent flags - --agents takes precedence over --agent
	if len(agentNames) > 0 {
		// Multi-agent mode
		if agentName != "" {
			utils.PrintError("Cannot specify both --agent and --agents flags")
			os.Exit(1)
		}
	} else if agentName != "" {
		// Single agent mode - convert to slice for uniform handling
		agentNames = []string{agentName}
	}

	// Support legacy --json flag for backward compatibility
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		outputFormat = "json"
	}
	jsonOutput := (outputFormat == "json")

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	ctx := context.Background()

	// Handle search-type flag (maps to UseImageVector)
	useImageVector := false
	if searchType == "visual" {
		useImageVector = true
	} else if searchType == "text" {
		useImageVector = false
	} else if searchType != "" {
		utils.PrintError(fmt.Sprintf("Invalid search-type '%s'. Must be 'text' or 'visual'", searchType))
		os.Exit(1)
	}

	// Handle image file for visual similarity search
	var imageBase64 string
	if imagePath != "" {
		// TODO: Read image file and convert to base64
		// For now, treat imagePath as already base64 encoded
		imageBase64 = imagePath
	}

	// Handle explicit vector name override
	if vectorName != "" {
		// Map common vector names
		switch vectorName {
		case "text", "text_vector":
			vectorName = "text_vector"
			useImageVector = false
		case "image", "image_vector", "visual":
			vectorName = "image_vector"
			useImageVector = true
		case "default":
			vectorName = "default"
			useImageVector = false
		}
		// Note: vectorName is informational for now, actual routing uses useImageVector
	}

	// Create query options
	options := weaviate.QueryOptions{
		TopK:           topK,
		TopKImages:     topKImages,
		Distance:       distance,
		SearchMetadata: searchMetadata,
		NoTruncate:     noTruncate,
		UseBM25:        useBM25,
		JSONOutput:     jsonOutput,
		ImageQuery:     imageBase64,
		UseImageVector: useImageVector,
		Verbose:        verbose,
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

	// Check for cross-VDB multi-collection queries
	hasExplicitVDBs := utils.HasExplicitVDBs(collectionSpecs)

	// Handle multi-collection queries
	if len(collectionSpecs) > 1 {
		if hasExplicitVDBs {
			// Cross-VDB multi-collection query
			resolver := utils.NewVDBConfigResolver(cfg, dbConfig)
			vdbConfigs, err := resolver.ResolveConfigs(collectionSpecs)
			if err != nil {
				utils.PrintError(fmt.Sprintf("Failed to resolve VDB configurations: %v", err))
				os.Exit(1)
			}

			if len(agentNames) > 0 {
				// Cross-VDB query with agent
				utils.QueryMultipleCollectionsWithAgentCrossVDB(ctx, collectionSpecs, vdbConfigs, queryText, options, agentNames, outputFormat, progress)
			} else {
				// Cross-VDB query without agent
				utils.QueryMultipleCollectionsCrossVDB(ctx, collectionSpecs, vdbConfigs, queryText, options)
			}
		} else {
			// Same-VDB multi-collection query (existing behavior)
			collectionNames := make([]string, len(collectionSpecs))
			for i, spec := range collectionSpecs {
				collectionNames[i] = spec.Name
			}

			if len(agentNames) > 0 {
				// Multi-collection query with agent
				utils.QueryMultipleCollectionsWithAgent(ctx, dbConfig, collectionNames, queryText, options, agentNames, outputFormat, progress)
			} else {
				// Multi-collection query without agent
				utils.QueryMultipleCollections(ctx, dbConfig, collectionNames, queryText, options)
			}
		}
		return
	}

	// Single collection query (existing logic)
	collectionSpec := collectionSpecs[0]
	collectionName := collectionSpec.Name

	// If collection has explicit VDB, resolve it
	if collectionSpec.VDBKey != "" {
		resolver := utils.NewVDBConfigResolver(cfg, dbConfig)
		vdbConfigs, err := resolver.ResolveConfigs(collectionSpecs)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to resolve VDB configuration: %v", err))
			os.Exit(1)
		}
		dbConfig = vdbConfigs[collectionName]
	}

	// If agent is specified, use agent-enabled query path
	if len(agentNames) > 0 {
		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			// Weaviate: use specialized function (keeps current behavior)
			utils.QueryWeaviateCollectionWithAgent(ctx, dbConfig, collectionName, queryText, options, agentNames, outputFormat, progress)
		case config.VectorDBTypeMock:
			// Mock: use specialized function
			utils.QueryMockCollectionWithAgent(ctx, dbConfig, collectionName, queryText, options, agentNames, outputFormat, progress)
		default:
			// All other VDBs: use generic function
			utils.QueryCollectionWithAgent(ctx, dbConfig, collectionName, queryText, options, agentNames, outputFormat, progress)
		}
		return
	}

	// Standard query path (no agent)
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.QueryWeaviateCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeSupabase, config.VectorDBTypeSupabaseCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMongoDB, config.VectorDBTypeMongoDBCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeQdrantLocal, config.VectorDBTypeQdrantCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeNeo4jLocal, config.VectorDBTypeNeo4jCloud:
		utils.QueryCollection(ctx, dbConfig, collectionName, queryText, options)
	case config.VectorDBTypeMock:
		utils.QueryMockCollection(ctx, dbConfig, collectionName, queryText, options)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
