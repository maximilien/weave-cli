// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/progress"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

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
			weaviateResults, queryErr := client.Query(ctx, collectionName, queryText, options)
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
			vdbOptions := &vectordb.QueryOptions{
				TopK:           options.TopK,
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
			weaviateResults, queryErr := client.Query(ctx, spec.Name, queryText, options)
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
			vdbOptions := &vectordb.QueryOptions{
				TopK:           options.TopK,
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
