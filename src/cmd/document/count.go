// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// CountCmd represents the document count command
var CountCmd = &cobra.Command{
	Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
	Aliases: []string{"C"},
	Short:   "Count documents in one or more collections",
	Long: `Count the number of documents in one or more collections.

This command returns the total number of documents in the specified collection(s).
You can specify multiple collections to get counts for each one.

Examples:
  weave docs C MyCollection
  weave docs C RagMeDocs RagMeImages
  weave docs C Collection1 Collection2 Collection3`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDocumentCount,
}

func init() {
	DocumentCmd.AddCommand(CountCmd)
}

func runDocumentCount(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Use vector database selector based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector databases: %v", err))
		os.Exit(1)
	}

	// Document count requires a single database
	if len(selection.Configs) != 1 {
		utils.PrintError("Document count requires a single database. Please specify --weaviate, --supabase, --mongodb, --milvus-local, --milvus-cloud, or --mock")
		os.Exit(1)
	}

	dbConfig := &selection.Configs[0]
	ctx := context.Background()

	type CollectionCount struct {
		Collection string `json:"collection"`
		Count      int    `json:"count"`
	}

	var results []CollectionCount

	// Count documents in each collection
	for _, collectionName := range args {
		var count int
		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			count, err = utils.CountWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeSupabase, config.VectorDBTypeMongoDB:
			count, err = utils.CountDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
			count, err = utils.CountDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
			count, err = utils.CountDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeQdrantLocal, config.VectorDBTypeQdrantCloud:
			count, err = utils.CountDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeNeo4jLocal, config.VectorDBTypeNeo4jCloud:
			count, err = utils.CountDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeMock:
			count, err = utils.CountMockDocuments(ctx, dbConfig, collectionName)
		default:
			utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			if !jsonOutput {
				utils.PrintError(fmt.Sprintf("Failed to count documents in collection '%s': %v", collectionName, err))
			}
			continue
		}

		results = append(results, CollectionCount{
			Collection: collectionName,
			Count:      count,
		})

		if !jsonOutput {
			utils.PrintHeader(fmt.Sprintf("Document Count - %s", collectionName))
			fmt.Printf("Total documents: %d\n", count)
		}
	}

	// Output JSON if requested
	if jsonOutput {
		// For single collection, output simplified JSON
		if len(results) == 1 {
			fmt.Printf(`{"collection": "%s", "count": %d}`+"\n", results[0].Collection, results[0].Count)
		} else {
			// For multiple collections, output array
			jsonBytes, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				utils.PrintError(fmt.Sprintf("Failed to marshal JSON: %v", err))
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
		}
	}
}
