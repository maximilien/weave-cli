// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ValidArgsFunction for database types
func validDatabaseTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"weaviate-cloud\tWeaviate Cloud instance",
		"weaviate-local\tLocal Weaviate instance",
		"supabase-cloud\tSupabase cloud instance",
		"supabase-local\tLocal Supabase instance",
		"mongodb-cloud\tMongoDB Atlas cloud",
		"mongodb-local\tLocal MongoDB instance",
		"milvus-cloud\tMilvus cloud (Zilliz)",
		"milvus-local\tLocal Milvus instance",
		"chroma-cloud\tChroma cloud instance",
		"chroma-local\tLocal Chroma instance",
		"qdrant-cloud\tQdrant cloud instance",
		"qdrant-local\tLocal Qdrant instance",
		"neo4j-cloud\tNeo4j cloud (AuraDB)",
		"neo4j-local\tLocal Neo4j instance",
		"opensearch-cloud\tOpenSearch cloud",
		"opensearch-local\tLocal OpenSearch",
		"mock\tMock database for testing",
	}, cobra.ShellCompDirectiveNoFileComp
}

// ValidArgsFunction for database names from config
func validDatabaseNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Try to load config
	cfg, err := config.LoadConfig("", "")
	if err != nil {
		// If config can't be loaded, return empty
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// Extract database names with descriptions
	var names []string
	for _, db := range cfg.Databases.VectorDatabases {
		names = append(names, db.Name+"\t"+string(db.Type))
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// ValidArgsFunction for config fields
func validConfigFields(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"url\tVector database URL",
		"api_key\tAPI key for authentication",
		"database_url\tDatabase connection string",
		"tenant\tTenant ID",
		"database\tDatabase name",
		"host\tHost address",
		"port\tPort number",
	}, cobra.ShellCompDirectiveNoFileComp
}

// ValidArgsFunction for log levels
func validLogLevels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"debug\tDetailed debug information",
		"info\tGeneral informational messages",
		"warn\tWarning messages",
		"error\tError messages only",
	}, cobra.ShellCompDirectiveNoFileComp
}

// ValidArgsFunction for log formats
func validLogFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"json\tStructured JSON logs",
		"text\tHuman-readable text logs",
		"pretty\tColorized pretty logs",
	}, cobra.ShellCompDirectiveNoFileComp
}

// ValidArgsFunction for timeout durations
func validTimeouts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"5s\t5 seconds",
		"10s\t10 seconds (default)",
		"30s\t30 seconds",
		"1m\t1 minute",
		"5m\t5 minutes",
	}, cobra.ShellCompDirectiveNoFileComp
}

// Register completions for common flags
func registerFlagCompletions(cmd *cobra.Command) {
	// Database type completion
	cmd.RegisterFlagCompletionFunc("vector-db-type", validDatabaseTypes)
	cmd.RegisterFlagCompletionFunc("vdb", validDatabaseTypes)

	// Log level completion
	cmd.RegisterFlagCompletionFunc("log-level", validLogLevels)

	// Log format completion
	cmd.RegisterFlagCompletionFunc("log-format", validLogFormats)

	// Timeout completion
	cmd.RegisterFlagCompletionFunc("timeout", validTimeouts)
}
