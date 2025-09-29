package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// DeleteCmd represents the collection delete command
var DeleteCmd = &cobra.Command{
	Use:     "delete COLLECTION_NAME [COLLECTION_NAME...]",
	Aliases: []string{"del", "d"},
	Short:   "Clear one or more collections (delete all documents)",
	Long: `Clear one or more collections by deleting all documents from them.

⚠️  WARNING: This is a destructive operation that will permanently
delete all documents in the specified collection(s). The collection
schema will remain but will be empty. Use with caution!

You can specify collections in multiple ways:
1. By collection name(s): weave cols delete MyCollection
2. By pattern: weave cols delete --pattern "WeaveDocs*"
3. By multiple names: weave cols delete Collection1 Collection2 Collection3

Pattern types (auto-detected):
- Shell glob: WeaveDocs*, Test*, *Docs
- Regex: WeaveDocs.*, ^Test.*$, .*Docs$

Examples:
  weave collection delete MyCollection
  weave collection delete Collection1 Collection2 Collection3
  weave collection delete --pattern "WeaveDocs*"`,
	Args: cobra.MinimumNArgs(1),
	Run:  runCollectionDelete,
}

func init() {
	CollectionCmd.AddCommand(DeleteCmd)

	DeleteCmd.Flags().StringP("pattern", "p", "", "Delete collections matching pattern (auto-detects shell glob vs regex)")
	DeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}

func runCollectionDelete(cmd *cobra.Command, args []string) {
	pattern, _ := cmd.Flags().GetString("pattern")
	force, _ := cmd.Flags().GetBool("force")

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
		if pattern != "" {
			message = fmt.Sprintf("Are you sure you want to delete collections matching pattern '%s'? This action cannot be undone.", pattern)
		} else {
			message = fmt.Sprintf("Are you sure you want to delete collection(s) %v? This action cannot be undone.", args)
		}

		if !utils.ConfirmAction(message) {
			fmt.Println("Operation cancelled")
			return
		}
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		if pattern != "" {
			err = utils.DeleteWeaviateCollectionsByPattern(ctx, dbConfig, pattern)
		} else {
			err = utils.DeleteWeaviateCollections(ctx, dbConfig, args)
		}
	case config.VectorDBTypeMock:
		if pattern != "" {
			err = utils.DeleteMockCollectionsByPattern(ctx, dbConfig, pattern)
		} else {
			err = utils.DeleteMockCollections(ctx, dbConfig, args)
		}
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to delete collection(s): %v", err))
		os.Exit(1)
	}

	if pattern != "" {
		utils.PrintSuccess(fmt.Sprintf("Successfully deleted collections matching pattern: %s", pattern))
	} else {
		utils.PrintSuccess(fmt.Sprintf("Successfully deleted collection(s): %v", args))
	}
}
