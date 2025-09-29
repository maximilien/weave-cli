package document

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
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
		utils.ListWeaviateDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
	case config.VectorDBTypeMock:
		utils.ListMockDocuments(ctx, dbConfig, collectionName, limit, showLong, shortLines, virtual, summary)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
