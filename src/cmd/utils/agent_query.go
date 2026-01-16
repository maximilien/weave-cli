// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/progress"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

// isImageCollectionBySchema checks if a collection contains image documents
// by checking collection name first, then documents, then schema
// Priority: Name > Documents > Schema (most to least reliable)
func isImageCollectionBySchema(ctx context.Context, client interface{}, collectionName string) bool {
	// PRIORITY 1: Check collection name (most reliable indicator of user intent)
	nameLower := strings.ToLower(collectionName)
	imageKeywords := []string{"image", "img", "photo", "picture", "visual", "media"}
	for _, keyword := range imageKeywords {
		if strings.Contains(nameLower, keyword) {
			fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' detected as IMAGE via name keyword: %s\n", collectionName, keyword)
			return true
		}
	}

	// PRIORITY 2: Sample documents to check for actual image data
	// (more reliable than schema which can have image fields in text collections)
	if vdbClient, ok := client.(vectordb.VectorDBClient); ok {
		docs, err := vdbClient.ListDocuments(ctx, collectionName, 3, 0)
		if err == nil && len(docs) > 0 {
			for _, doc := range docs {
				if doc.Image != "" || doc.ImageData != "" {
					fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' detected as IMAGE via document fields\n", collectionName)
					return true
				}
			}
		}
	}

	// PRIORITY 3: Check schema (least reliable - multi-modal collections can have image fields)
	// Only use this as last resort
	if weaviateClient, ok := client.(*weaviate.Client); ok {
		schema, err := weaviateClient.GetCollectionSchema(ctx, collectionName)
		if err == nil && len(schema) > 0 {
			// Check for image-related fields
			for _, field := range schema {
				fieldLower := strings.ToLower(field)
				// Only consider it an image collection if it has imagedata or image_url
				// (not just "image" which can be present in text collections)
				if fieldLower == "imagedata" || fieldLower == "image_url" {
					fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' detected as IMAGE via schema field: %s\n", collectionName, field)
					return true
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' detected as TEXT (no image indicators found)\n", collectionName)
	return false
}

// QueryMultipleCollectionsWithAgent queries multiple collections and processes aggregated results through an agent
func QueryMultipleCollectionsWithAgent(ctx context.Context, cfg *config.VectorDBConfig, collectionNames []string, queryText string, options weaviate.QueryOptions, agentName string, outputFormat string, showProgress bool) {
	// Create progress reporter
	var reporter *progress.Reporter
	if showProgress && outputFormat == "json" {
		reporter = progress.NewJSONReporter(true)
	} else {
		reporter = progress.NewReporter(showProgress)
	}

	reporter.Start(fmt.Sprintf("Querying %d collections...", len(collectionNames)))
	reporter.SetTotal(len(collectionNames))

	// Aggregate results from all collections
	var allResults []*vectordb.QueryResult

	// Query each collection
	for _, collectionName := range collectionNames {
		reporter.UpdateProgress(fmt.Sprintf("Querying collection '%s'...", collectionName))

		// Determine top_k for this collection based on whether it's an image collection
		collectionTopK := options.TopK

		// Query the collection based on database type
		var results []*vectordb.QueryResult
		var err error

		switch cfg.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			// Weaviate
			client, clientErr := CreateWeaviateClient(cfg)
			if clientErr != nil {
				PrintError(fmt.Sprintf("Failed to create Weaviate client for collection '%s': %v", collectionName, clientErr))
				continue
			}

			// Check if this is an image collection and use topKImages if specified
			isImage := isImageCollectionBySchema(ctx, client, collectionName)
			if options.Verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s': isImage=%v, topK=%d, topKImages=%d\n",
					collectionName, isImage, options.TopK, options.TopKImages)
			}
			if options.TopKImages > 0 && isImage {
				collectionTopK = options.TopKImages
				if options.Verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Using topKImages=%d for image collection '%s'\n",
						collectionTopK, collectionName)
				}
			}

			// Create collection-specific options with adjusted topK
			collectionOptions := options
			collectionOptions.TopK = collectionTopK

			weaviateResults, queryErr := client.Query(ctx, collectionName, queryText, collectionOptions)
			if queryErr != nil {
				PrintError(fmt.Sprintf("Failed to query collection '%s': %v", collectionName, queryErr))
				continue
			}
			// Convert weaviate.QueryResult to vectordb.QueryResult
			results = make([]*vectordb.QueryResult, len(weaviateResults))
			for j, result := range weaviateResults {
				results[j] = &vectordb.QueryResult{
					Document: vectordb.Document{
						ID:       result.ID,
						Content:  result.Content,
						Metadata: result.Metadata,
					},
					Score: result.Score,
				}
			}
		default:
			// All other VDBs: use generic interface
			client, clientErr := CreateVectorDBClient(cfg)
			if clientErr != nil {
				PrintError(fmt.Sprintf("Failed to create client for collection '%s': %v", collectionName, clientErr))
				continue
			}

			// Check if this is an image collection and use topKImages if specified
			if options.TopKImages > 0 && isImageCollectionBySchema(ctx, client, collectionName) {
				collectionTopK = options.TopKImages
			}

			vdbOptions := &vectordb.QueryOptions{
				TopK:           collectionTopK,
				Distance:       options.Distance,
				SearchMetadata: options.SearchMetadata,
				NoTruncate:     options.NoTruncate,
				UseBM25:        options.UseBM25,
			}
			if vdbOptions.UseBM25 {
				results, err = client.SearchBM25(ctx, collectionName, queryText, vdbOptions)
			} else {
				results, err = client.SearchSemantic(ctx, collectionName, queryText, vdbOptions)
			}
			if err != nil {
				PrintError(fmt.Sprintf("Failed to query collection '%s': %v", collectionName, err))
				continue
			}
		}

		// Add collection name to metadata for context
		for _, result := range results {
			if result.Document.Metadata == nil {
				result.Document.Metadata = make(map[string]interface{})
			}
			result.Document.Metadata["_collection"] = collectionName
		}

		// Log results for debugging
		if options.Verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' returned %d results (requested topK=%d)\n",
				collectionName, len(results), collectionTopK)
		}

		// Aggregate results
		allResults = append(allResults, results...)
		reporter.Update(fmt.Sprintf("Found %d results from '%s' (total: %d)", len(results), collectionName, len(allResults)))
	}

	if len(allResults) == 0 {
		reporter.Complete("No results found")
		PrintError("No results found from any collection")
		return
	}

	reporter.Update(fmt.Sprintf("Aggregated %d results from %d collections", len(allResults), len(collectionNames)))

	// Execute through agent with all aggregated results
	ExecuteQueryWithAgent(ctx, agentName, queryText, allResults, outputFormat, showProgress)
}

// ExecuteQueryWithAgent executes a query and processes results through an agent
func ExecuteQueryWithAgent(ctx context.Context, agentName string, query string, results []*vectordb.QueryResult, outputFormat string, showProgress bool) {
	// Create progress reporter (use JSON reporter if output format is JSON)
	var reporter *progress.Reporter
	if showProgress && outputFormat == "json" {
		reporter = progress.NewJSONReporter(true)
	} else {
		reporter = progress.NewReporter(showProgress)
	}
	reporter.Start(fmt.Sprintf("Processing %d results with %s...", len(results), agentName))
	// Load agent configuration
	reporter.Update("Loading agent configuration...")
	agentConfig, err := agents.LoadAgent(agentName)
	if err != nil {
		if agents.IsAgentNotFoundError(err) {
			PrintError(fmt.Sprintf("Agent '%s' not found. Use 'weave agents list' to see available agents.", agentName))
		} else {
			PrintError(fmt.Sprintf("Failed to load agent '%s': %v", agentName, err))
		}
		os.Exit(1)
	}

	// Override output format if specified via command-line flag
	if outputFormat != "" {
		agentConfig.Output.Format = outputFormat
	}

	// Create LLM client
	reporter.Update("Initializing LLM client...")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		PrintError("OPENAI_API_KEY environment variable is required for agent execution")
		os.Exit(1)
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create LLM client: %v", err))
		os.Exit(1)
	}

	// Create RAG agent
	reporter.Update("Creating RAG agent...")
	ragAgent, err := agents.NewRAGAgent(agentConfig, llmClient)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create RAG agent: %v", err))
		os.Exit(1)
	}

	// Execute agent
	reporter.Update(fmt.Sprintf("Building context from %d sources...", len(results)))
	input := &agents.RAGInput{
		Query:   query,
		Results: results,
	}

	reporter.Update("Generating response...")
	output, err := ragAgent.Execute(ctx, input)
	if err != nil {
		PrintError(fmt.Sprintf("Agent execution failed: %v", err))
		os.Exit(1)
	}

	ragOutput, ok := output.(*agents.RAGOutput)
	if !ok {
		PrintError(fmt.Sprintf("Unexpected agent output type: %T", output))
		os.Exit(1)
	}

	// Format and display output
	reporter.Update("Formatting output...")
	formatted, err := ragAgent.FormatOutput(ragOutput)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to format agent output: %v", err))
		os.Exit(1)
	}

	reporter.Complete("Done")
	fmt.Println(formatted)
}

// QueryWeaviateCollectionWithAgent queries a collection and processes results through an agent
func QueryWeaviateCollectionWithAgent(ctx context.Context, cfg *config.VectorDBConfig, collectionName, queryText string, options weaviate.QueryOptions, agentName string, outputFormat string, showProgress bool) {
	// Create progress reporter (use JSON reporter if output format is JSON)
	var reporter *progress.Reporter
	if showProgress && outputFormat == "json" {
		reporter = progress.NewJSONReporter(true)
	} else {
		reporter = progress.NewReporter(showProgress)
	}
	reporter.Start("Searching collection...")
	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to create Weaviate client: %v", err))
		return
	}

	// Perform the semantic search
	results, err := client.Query(ctx, collectionName, queryText, options)
	if err != nil {
		PrintError(fmt.Sprintf("Failed to query collection '%s': %v", collectionName, err))
		return
	}

	reporter.Update(fmt.Sprintf("Found %d results", len(results)))

	// Convert weaviate.QueryResult to vectordb.QueryResult
	vdbResults := make([]*vectordb.QueryResult, len(results))
	for i, result := range results {
		vdbResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       result.ID,
				Content:  result.Content,
				Metadata: result.Metadata,
			},
			Score: result.Score,
		}
	}

	// Execute through agent
	ExecuteQueryWithAgent(ctx, agentName, queryText, vdbResults, outputFormat, showProgress)
}

// QueryMultipleCollectionsWithAgentCrossVDB queries multiple collections across different VDBs and processes aggregated results through an agent
func QueryMultipleCollectionsWithAgentCrossVDB(ctx context.Context, collectionSpecs []CollectionSpec, vdbConfigs map[string]*config.VectorDBConfig, queryText string, options weaviate.QueryOptions, agentName string, outputFormat string, showProgress bool) {
	// Create progress reporter
	var reporter *progress.Reporter
	if showProgress && outputFormat == "json" {
		reporter = progress.NewJSONReporter(true)
	} else {
		reporter = progress.NewReporter(showProgress)
	}

	reporter.Start(fmt.Sprintf("Querying %d collections across multiple VDBs...", len(collectionSpecs)))
	reporter.SetTotal(len(collectionSpecs))

	// Aggregate results from all collections
	var allResults []*vectordb.QueryResult

	// Query each collection from its respective VDB
	for _, spec := range collectionSpecs {
		reporter.UpdateProgress(fmt.Sprintf("Querying collection '%s' from %s...", spec.Name, spec.VDBKey))

		vdbConfig := vdbConfigs[spec.Name]
		if vdbConfig == nil {
			PrintError(fmt.Sprintf("No VDB configuration found for collection '%s'", spec.Name))
			continue
		}

		// Determine top_k for this collection based on whether it's an image collection
		collectionTopK := options.TopK

		// Query the collection based on database type
		var results []*vectordb.QueryResult
		var err error

		switch vdbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			// Weaviate
			client, clientErr := CreateWeaviateClient(vdbConfig)
			if clientErr != nil {
				PrintError(fmt.Sprintf("Failed to create Weaviate client for collection '%s': %v", spec.Name, clientErr))
				continue
			}

			// Check if this is an image collection and use topKImages if specified
			isImage := isImageCollectionBySchema(ctx, client, spec.Name)
			if options.Verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' (%s): isImage=%v, topK=%d, topKImages=%d\n",
					spec.Name, spec.VDBKey, isImage, options.TopK, options.TopKImages)
			}
			if options.TopKImages > 0 && isImage {
				collectionTopK = options.TopKImages
				if options.Verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Using topKImages=%d for image collection '%s'\n",
						collectionTopK, spec.Name)
				}
			}

			// Create collection-specific options with adjusted topK
			collectionOptions := options
			collectionOptions.TopK = collectionTopK

			weaviateResults, queryErr := client.Query(ctx, spec.Name, queryText, collectionOptions)
			if queryErr != nil {
				PrintError(fmt.Sprintf("Failed to query collection '%s': %v", spec.Name, queryErr))
				continue
			}
			// Convert weaviate.QueryResult to vectordb.QueryResult
			results = make([]*vectordb.QueryResult, len(weaviateResults))
			for j, result := range weaviateResults {
				results[j] = &vectordb.QueryResult{
					Document: vectordb.Document{
						ID:       result.ID,
						Content:  result.Content,
						Metadata: result.Metadata,
					},
					Score: result.Score,
				}
			}
		default:
			// All other VDBs: use generic interface
			client, clientErr := CreateVectorDBClient(vdbConfig)
			if clientErr != nil {
				PrintError(fmt.Sprintf("Failed to create client for collection '%s': %v", spec.Name, clientErr))
				continue
			}

			// Check if this is an image collection and use topKImages if specified
			if options.TopKImages > 0 && isImageCollectionBySchema(ctx, client, spec.Name) {
				collectionTopK = options.TopKImages
			}

			vdbOptions := &vectordb.QueryOptions{
				TopK:           collectionTopK,
				Distance:       options.Distance,
				SearchMetadata: options.SearchMetadata,
				NoTruncate:     options.NoTruncate,
				UseBM25:        options.UseBM25,
			}
			if vdbOptions.UseBM25 {
				results, err = client.SearchBM25(ctx, spec.Name, queryText, vdbOptions)
			} else {
				results, err = client.SearchSemantic(ctx, spec.Name, queryText, vdbOptions)
			}
			if err != nil {
				PrintError(fmt.Sprintf("Failed to query collection '%s': %v", spec.Name, err))
				continue
			}
		}

		// Add collection name and VDB info to metadata for context
		vdbKey := spec.VDBKey
		if vdbKey == "" {
			vdbKey = string(vdbConfig.Type)
		}
		for _, result := range results {
			if result.Document.Metadata == nil {
				result.Document.Metadata = make(map[string]interface{})
			}
			result.Document.Metadata["_collection"] = spec.Name
			result.Document.Metadata["_vdb"] = vdbKey
			result.Document.Metadata["_vdb_type"] = string(vdbConfig.Type)
		}

		// Log results for debugging
		if options.Verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Collection '%s' (%s) returned %d results (requested topK=%d)\n",
				spec.Name, vdbKey, len(results), collectionTopK)
		}

		// Aggregate results
		allResults = append(allResults, results...)
		reporter.Update(fmt.Sprintf("Found %d results from '%s' (%s) (total: %d)", len(results), spec.Name, vdbKey, len(allResults)))
	}

	if len(allResults) == 0 {
		reporter.Complete("No results found")
		PrintError("No results found from any collection")
		return
	}

	reporter.Update(fmt.Sprintf("Aggregated %d results from %d collections across multiple VDBs", len(allResults), len(collectionSpecs)))

	// Execute through agent with all aggregated results
	ExecuteQueryWithAgent(ctx, agentName, queryText, allResults, outputFormat, showProgress)
}
