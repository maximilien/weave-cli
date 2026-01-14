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
