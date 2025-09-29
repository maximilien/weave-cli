// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ShowCmd represents the document show command
var ShowCmd = &cobra.Command{
	Use:     "show COLLECTION_NAME [DOCUMENT_ID]",
	Aliases: []string{"s"},
	Short:   "Show documents from a collection",
	Long: `Show detailed information about documents from a collection.

You can show documents in three ways:
1. By document ID: weave doc show COLLECTION_NAME DOCUMENT_ID
2. By metadata filter: weave doc show COLLECTION_NAME --metadata key=value
3. By filename/name: weave doc show COLLECTION_NAME --name filename.pdf
   (or use --filename as an alias)

This command displays:
- Full document content
- Complete metadata
- Document ID and collection information

Use --schema to show the document schema including metadata structure.
Use --expand-metadata to show expanded metadata information.`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runDocumentShow,
}

func init() {
	DocumentCmd.AddCommand(ShowCmd)

	ShowCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	ShowCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	ShowCmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter (format: key=value)")
	ShowCmd.Flags().StringP("name", "n", "", "Show document by filename/name")
	ShowCmd.Flags().StringP("filename", "", "", "Show document by filename/name (alias for --name)")
	ShowCmd.Flags().Bool("schema", false, "Show document schema including metadata structure")
	ShowCmd.Flags().Bool("expand-metadata", false, "Show expanded metadata information")
}

func runDocumentShow(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	name, _ := cmd.Flags().GetString("name")
	filename, _ := cmd.Flags().GetString("filename")
	showSchema, _ := cmd.Flags().GetBool("schema")
	expandMetadata, _ := cmd.Flags().GetBool("expand-metadata")

	// Use filename if name is not provided
	if name == "" {
		name = filename
	}

	// Load configuration
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.ShowWeaviateDocument(ctx, dbConfig, collectionName, args[1:], showLong, shortLines, metadataFilters, name, showSchema, expandMetadata)
	case config.VectorDBTypeMock:
		utils.ShowMockDocument(ctx, dbConfig, collectionName, args[1:], showLong, shortLines, metadataFilters, name, showSchema, expandMetadata)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
