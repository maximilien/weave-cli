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

// UpdateCmd represents the document update command
var UpdateCmd = &cobra.Command{
	Use:     "update COLLECTION_NAME DOCUMENT_ID",
	Aliases: []string{"u"},
	Short:   "Update a document's content or metadata",
	Long: `Update an existing document in a collection.

You can update:
- Document content (using --content or --file)
- Document metadata fields (using --metadata)
- Both content and metadata together

The document is identified by its UUID or by metadata filter (--name, --metadata-filter).

Examples:
  # Update content from string
  weave docs update MyCollection abc-123 --content "New content here"

  # Update content from file
  weave docs update MyCollection abc-123 --file updated-doc.txt

  # Update metadata fields
  weave docs update MyCollection abc-123 --metadata title="Updated Title"
  weave docs update MyCollection abc-123 --metadata version=2,author="John Doe"

  # Update by document name (instead of ID)
  weave docs update MyCollection --name README.md --content "Updated README"

  # Update both content and metadata
  weave docs update MyCollection abc-123 --file new.txt --metadata version=2`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDocumentUpdate,
}

func init() {
	DocumentCmd.AddCommand(UpdateCmd)

	UpdateCmd.Flags().StringP("content", "c", "", "New content for the document")
	UpdateCmd.Flags().StringP("file", "f", "", "Path to file with new content")
	UpdateCmd.Flags().StringSliceP("metadata", "m", []string{}, "Metadata fields to update (key=value format)")
	UpdateCmd.Flags().StringP("name", "n", "", "Document name (alternative to ID)")
	UpdateCmd.Flags().StringSlice("metadata-filter", []string{}, "Filter documents by metadata (key=value)")
}

func runDocumentUpdate(cmd *cobra.Command, args []string) {
	collectionName := args[0]

	// Get flags
	content, _ := cmd.Flags().GetString("content")
	filePath, _ := cmd.Flags().GetString("file")
	metadata, _ := cmd.Flags().GetStringSlice("metadata")
	docName, _ := cmd.Flags().GetString("name")
	metadataFilter, _ := cmd.Flags().GetStringSlice("metadata-filter")

	// Determine document ID or filter
	var documentID string
	if len(args) >= 2 {
		documentID = args[1]
	}

	// Validate inputs
	if documentID == "" && docName == "" && len(metadataFilter) == 0 {
		utils.PrintError("Must specify either document ID, --name, or --metadata-filter")
		os.Exit(1)
	}

	if content == "" && filePath == "" && len(metadata) == 0 {
		utils.PrintError("Must specify at least one of: --content, --file, or --metadata")
		os.Exit(1)
	}

	if content != "" && filePath != "" {
		utils.PrintError("Cannot specify both --content and --file")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		utils.HandleConfigError(err, true)
		os.Exit(1)
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		if err := utils.UpdateWeaviateDocument(ctx, dbConfig, collectionName, documentID, docName, metadataFilter, content, filePath, metadata); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to update document: %v", err))
			os.Exit(1)
		}
	case config.VectorDBTypeSupabase, config.VectorDBTypeMongoDB:
		utils.PrintError(fmt.Sprintf("Document update not yet implemented for %s", dbConfig.Type))
		os.Exit(1)
	case config.VectorDBTypeMock:
		utils.UpdateMockDocument(ctx, dbConfig, collectionName, documentID, docName, metadataFilter, content, filePath, metadata)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
