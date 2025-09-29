package document

import (
	"context"
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

	// Count documents in each collection
	for _, collectionName := range args {
		var count int
		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			count, err = utils.CountWeaviateDocuments(ctx, dbConfig, collectionName)
		case config.VectorDBTypeMock:
			count, err = utils.CountMockDocuments(ctx, dbConfig, collectionName)
		default:
			utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to count documents in collection '%s': %v", collectionName, err))
			continue
		}

		utils.PrintHeader(fmt.Sprintf("Document Count - %s", collectionName))
		fmt.Printf("Total documents: %d\n", count)
	}
}