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

// DeleteCmd represents the document delete command
var DeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME [DOCUMENT_ID] [DOCUMENT_ID...]",
	Aliases: []string{"del", "d"},
	Short:   "Delete documents from a collection",
	Long: `Delete documents from a collection.

You can delete documents in six ways:
1. By single document ID: weave doc delete COLLECTION_NAME DOCUMENT_ID
2. By multiple document IDs: weave doc delete COLLECTION_NAME DOC_ID1 DOC_ID2 DOC_ID3
3. By metadata filter: weave doc delete COLLECTION_NAME --metadata key=value
4. By filename/name: weave doc delete COLLECTION_NAME --name filename.pdf
   (or use --filename as an alias)
5. By original filename (virtual): weave doc delete COLLECTION_NAME ORIGINAL_FILENAME --virtual
6. By pattern: weave doc delete COLLECTION_NAME --pattern "tmp*.png"

Pattern types (auto-detected):
- Shell glob: tmp*.png, tmp?.png, tmp[0-9].png
- Regex: tmp.*\.png, ^tmp.*\.png$, .*\.(png|jpg)$

When using --virtual flag, all chunks and images associated with the original filename
will be deleted in one operation.

Examples:
  weave docs delete MyCollection doc123
  weave docs d MyCollection doc123 doc456 doc789
  weave docs delete MyCollection --metadata filename=test.pdf
  weave docs delete MyCollection --name test_image.png
  weave docs delete MyCollection --filename test_image.png
  weave docs delete MyCollection test.pdf --virtual
  weave docs delete MyCollection --pattern "tmp*.png"
  weave docs delete MyCollection --pattern "tmp.*\.png"

⚠️  WARNING: This is a destructive operation that will permanently
delete the specified documents. Use with caution!`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDocumentDelete,
}

func init() {
	DocumentCmd.AddCommand(DeleteCmd)

	DeleteCmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter (format: key=value)")
	DeleteCmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")
	DeleteCmd.Flags().StringP("pattern", "p", "", "Delete documents matching pattern (auto-detects shell glob vs regex)")
	DeleteCmd.Flags().StringP("name", "n", "", "Delete document by filename/name")
	DeleteCmd.Flags().StringP("filename", "", "", "Delete document by filename/name (alias for --name)")
	DeleteCmd.Flags().BoolP("force", "", false, "Skip confirmation prompt")
}

func runDocumentDelete(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
	virtual, _ := cmd.Flags().GetBool("virtual")
	pattern, _ := cmd.Flags().GetString("pattern")
	name, _ := cmd.Flags().GetString("name")
	filename, _ := cmd.Flags().GetString("filename")
	force, _ := cmd.Flags().GetBool("force")

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

	// Confirmation prompt
	if !force {
		var message string
		if len(args) > 1 {
			message = fmt.Sprintf("Are you sure you want to delete document(s) %v from collection '%s'?", args[1:], collectionName)
		} else if pattern != "" {
			message = fmt.Sprintf("Are you sure you want to delete documents matching pattern '%s' from collection '%s'?", pattern, collectionName)
		} else if name != "" {
			message = fmt.Sprintf("Are you sure you want to delete document '%s' from collection '%s'?", name, collectionName)
		} else if len(metadataFilters) > 0 {
			message = fmt.Sprintf("Are you sure you want to delete documents matching metadata %v from collection '%s'?", metadataFilters, collectionName)
		} else {
			message = fmt.Sprintf("Are you sure you want to delete documents from collection '%s'?", collectionName)
		}

		if !utils.ConfirmAction(message) {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.DeleteWeaviateDocuments(ctx, dbConfig, collectionName, args[1:], metadataFilters, virtual, pattern, name)
	case config.VectorDBTypeMock:
		utils.DeleteMockDocuments(ctx, dbConfig, collectionName, args[1:], metadataFilters, virtual, pattern, name)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	utils.PrintSuccess("Successfully deleted document(s)")
}
