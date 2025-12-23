// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Cmd represents the stats command
var Cmd = &cobra.Command{
	Use:   "stats [collection-name]",
	Short: "Show collection statistics and analytics",
	Long: `Show detailed statistics about a collection.

This command displays:
- Document count
- Metadata field distribution
- Collection size and index information
- Top metadata values

Examples:
  weave stats MyCollection
  weave stats MyDocs --output json
  weave stats MyCollection --top 10`,
	Args: cobra.ExactArgs(1),
	Run:  runStats,
}

// CollectionStats represents statistics about a collection
type CollectionStats struct {
	CollectionName string                       `json:"collection_name" yaml:"collection_name"`
	DocumentCount  int                          `json:"document_count" yaml:"document_count"`
	VectorDB       string                       `json:"vector_db" yaml:"vector_db"`
	Metadata       map[string]MetadataFieldStat `json:"metadata_fields,omitempty" yaml:"metadata_fields,omitempty"`
}

// MetadataFieldStat represents statistics for a metadata field
type MetadataFieldStat struct {
	Type         string                 `json:"type" yaml:"type"`
	UniqueValues int                    `json:"unique_values" yaml:"unique_values"`
	TopValues    []ValueCount           `json:"top_values,omitempty" yaml:"top_values,omitempty"`
	SampleValue  interface{}            `json:"sample_value,omitempty" yaml:"sample_value,omitempty"`
}

// ValueCount represents a value and its count
type ValueCount struct {
	Value string `json:"value" yaml:"value"`
	Count int    `json:"count" yaml:"count"`
}

func init() {
	Cmd.Flags().StringP("output", "o", "text", "Output format: text, json, yaml")
	Cmd.Flags().IntP("top", "t", 5, "Number of top values to show for each metadata field")
	Cmd.Flags().IntP("limit", "l", 1000, "Maximum number of documents to analyze for metadata stats")
}

func runStats(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	outputFormat, _ := cmd.Flags().GetString("output")
	topN, _ := cmd.Flags().GetInt("top")
	limit, _ := cmd.Flags().GetInt("limit")

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		os.Exit(1)
	}

	// Get selected database
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector database: %v", err))
		os.Exit(1)
	}

	// Stats requires a single database
	if len(selection.Configs) != 1 {
		utils.PrintError("Stats requires a single database. Please specify --weaviate, --supabase, --mongodb, --milvus-local, --milvus-cloud, --chroma, --qdrant, --neo4j, or --mock")
		os.Exit(1)
	}

	dbConfig := &selection.Configs[0]
	ctx := context.Background()

	// Collect statistics
	stats, err := collectStats(ctx, dbConfig, collectionName, limit, topN)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to collect statistics: %v", err))
		os.Exit(1)
	}

	// Output based on format
	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(stats)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to marshal YAML: %v", err))
			os.Exit(1)
		}
		fmt.Println(string(data))
	default:
		displayStats(stats, topN)
	}
}

func collectStats(ctx context.Context, dbConfig *config.VectorDBConfig, collectionName string, limit int, topN int) (*CollectionStats, error) {
	// Create vector DB client
	vdbConfig := &vectordb.Config{
		Type:             vectordb.VectorDBType(dbConfig.Type),
		URL:              dbConfig.URL,
		APIKey:           dbConfig.APIKey,
		OpenAIAPIKey:     dbConfig.OpenAIAPIKey,
		DatabaseURL:      dbConfig.DatabaseURL,
		DatabaseKey:      dbConfig.DatabaseKey,
		Database:         dbConfig.Database,
		Username:         dbConfig.Username,
		Password:         dbConfig.Password,
		Address:          dbConfig.Address,
		Tenant:           dbConfig.Tenant,
		VectorDimensions: dbConfig.VectorDimensions,
		SimilarityMetric: dbConfig.SimilarityMetric,
		Timeout:          dbConfig.Timeout,
	}

	client, err := vectordb.CreateClient(vdbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Get document count
	collections, err := client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	docCount := int64(0)
	for _, coll := range collections {
		if coll.Name == collectionName {
			docCount = coll.Count
			break
		}
	}

	stats := &CollectionStats{
		CollectionName: collectionName,
		DocumentCount:  int(docCount),
		VectorDB:       string(dbConfig.Type),
		Metadata:       make(map[string]MetadataFieldStat),
	}

	// If we have documents, analyze metadata
	if docCount > 0 {
		// Sample documents for metadata analysis
		sampleSize := limit
		if int64(sampleSize) > docCount {
			sampleSize = int(docCount)
		}

		// Get sample documents (offset 0 for first documents)
		docs, err := client.ListDocuments(ctx, collectionName, sampleSize, 0)
		if err != nil {
			// Metadata analysis is optional, so don't fail if we can't get documents
			return stats, nil
		}

		// Analyze metadata fields
		stats.Metadata = analyzeMetadata(docs, topN)
	}

	return stats, nil
}

func analyzeMetadata(docs []*vectordb.Document, topN int) map[string]MetadataFieldStat {
	fieldStats := make(map[string]MetadataFieldStat)
	fieldValues := make(map[string]map[string]int)

	// Collect all metadata fields and their values
	for _, doc := range docs {
		for field, value := range doc.Metadata {
			if _, exists := fieldValues[field]; !exists {
				fieldValues[field] = make(map[string]int)
			}

			// Convert value to string for counting
			valueStr := fmt.Sprintf("%v", value)
			fieldValues[field][valueStr]++

			// Store sample value if not already stored
			if stat, exists := fieldStats[field]; !exists {
				fieldStats[field] = MetadataFieldStat{
					Type:        fmt.Sprintf("%T", value),
					SampleValue: value,
				}
			} else if stat.SampleValue == nil {
				stat.SampleValue = value
				fieldStats[field] = stat
			}
		}
	}

	// Calculate statistics for each field
	for field, values := range fieldValues {
		stat := fieldStats[field]
		stat.UniqueValues = len(values)

		// Get top N values
		type kv struct {
			Key   string
			Value int
		}
		var sortedValues []kv
		for k, v := range values {
			sortedValues = append(sortedValues, kv{k, v})
		}
		sort.Slice(sortedValues, func(i, j int) bool {
			return sortedValues[i].Value > sortedValues[j].Value
		})

		// Store top N
		topValues := []ValueCount{}
		for i := 0; i < len(sortedValues) && i < topN; i++ {
			topValues = append(topValues, ValueCount{
				Value: sortedValues[i].Key,
				Count: sortedValues[i].Value,
			})
		}
		stat.TopValues = topValues

		fieldStats[field] = stat
	}

	return fieldStats
}

func displayStats(stats *CollectionStats, topN int) {
	// Header
	utils.PrintHeader(fmt.Sprintf("Collection Statistics: %s", stats.CollectionName))
	fmt.Println()

	// Basic info
	color.New(color.FgCyan).Printf("📊 Overview\n")
	fmt.Printf("  Collection: %s\n", stats.CollectionName)
	fmt.Printf("  Vector DB: %s\n", stats.VectorDB)
	fmt.Printf("  Documents: %s\n", color.New(color.FgGreen).Sprintf("%d", stats.DocumentCount))
	fmt.Println()

	// Metadata fields
	if len(stats.Metadata) > 0 {
		color.New(color.FgCyan).Printf("📋 Metadata Fields (%d total)\n", len(stats.Metadata))
		fmt.Println()

		// Sort fields alphabetically
		fields := make([]string, 0, len(stats.Metadata))
		for field := range stats.Metadata {
			fields = append(fields, field)
		}
		sort.Strings(fields)

		for _, field := range fields {
			stat := stats.Metadata[field]
			color.New(color.FgYellow).Printf("  %s\n", field)
			fmt.Printf("    Type: %s\n", stat.Type)
			fmt.Printf("    Unique values: %d\n", stat.UniqueValues)

			if len(stat.TopValues) > 0 {
				fmt.Printf("    Top %d values:\n", len(stat.TopValues))
				for _, vc := range stat.TopValues {
					fmt.Printf("      • %s (%d occurrences)\n", vc.Value, vc.Count)
				}
			}
			fmt.Println()
		}
	} else {
		color.New(color.FgYellow).Printf("ℹ️  No metadata fields found\n")
	}
}
