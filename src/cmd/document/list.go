// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/spf13/cobra"
)

// ListCmd represents the document list command
var ListCmd = &cobra.Command{
	Use:     "list COLLECTION_NAME",
	Aliases: []string{"ls", "l"},
	Short:   "List documents in a collection",
	Long: `List documents in a specific collection.

This command shows:
- Document IDs
- Content previews (truncated)
- Metadata information
- Document counts`,
	Args: cobra.ExactArgs(1),
	Run:  runDocumentList,
}

func init() {
	DocumentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntP("limit", "l", 50, "Maximum number of documents to show")
	ListCmd.Flags().BoolP("long", "L", false, "Show full content instead of preview")
	ListCmd.Flags().IntP("short", "s", 5, "Show only first N lines of content (default: 5)")
	ListCmd.Flags().BoolP("virtual", "w", false, "Show documents in virtual structure (aggregate chunks by original document)")
	ListCmd.Flags().BoolP("summary", "S", false, "Show a clean summary of documents (works with --virtual)")
}

func runDocumentList(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	showLong, _ := cmd.Flags().GetBool("long")
	shortLines, _ := cmd.Flags().GetInt("short")
	virtual, _ := cmd.Flags().GetBool("virtual")
	summary, _ := cmd.Flags().GetBool("summary")

	// Load configuration with interactive help
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get selected databases based on flags
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	// Validate that only one database is selected for read operations
	if len(selection.Configs) > 1 {
		utils.PrintError("Document list requires a single database. Please specify --weaviate, --supabase, or --mock")
		os.Exit(1)
	}

	dbConfig := &selection.Configs[0]
	ctx := context.Background()

	// Use generic ListDocuments that works with all database types via vectordb abstraction
	utils.ListDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
}
