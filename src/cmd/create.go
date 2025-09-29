package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"github.com/spf13/cobra"
)

// collectionCreateCmd represents the collection create command
var collectionCreateCmd = &cobra.Command{
	Use:     "create COLLECTION_NAME",
	Aliases: []string{"c"},
	Short:   "Create a collection",
	Long: `Create a new collection in the vector database.

This command creates a collection with the specified name and schema.
You can specify custom fields and embedding model.

Examples:
  weave collection create MyCollection
  weave collection create MyCollection --embedding-model text-embedding-ada-002
  weave collection create MyCollection --fields "title:string,description:text"`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionCreate,
}

func init() {
	collectionCmd.AddCommand(collectionCreateCmd)

	collectionCreateCmd.Flags().StringP("embedding-model", "e", "text-embedding-ada-002", "Embedding model to use")
	collectionCreateCmd.Flags().StringP("fields", "f", "", "Custom fields (format: field1:type1,field2:type2)")
	collectionCreateCmd.Flags().StringP("schema-type", "s", "default", "Schema type (default, ragmedocs, ragmeimages)")
}

func runCollectionCreate(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	embeddingModel, _ := cmd.Flags().GetString("embedding-model")
	fieldsStr, _ := cmd.Flags().GetString("fields")
	schemaType, _ := cmd.Flags().GetString("schema-type")

	// Parse custom fields
	var customFields []weaviate.FieldDefinition
	if fieldsStr != "" {
		fields, err := parseFieldDefinitions(fieldsStr)
		if err != nil {
			printError(fmt.Sprintf("Failed to parse fields: %v", err))
			os.Exit(1)
		}
		customFields = fields
	}

	// Load configuration
	cfg, err := loadConfigWithOverrides()
	if err != nil {
		printError(fmt.Sprintf("Failed to load configuration: %v", err))
		os.Exit(1)
	}

	// Get default database
	dbConfig, err := cfg.GetDefaultDatabase()
	if err != nil {
		printError(fmt.Sprintf("Failed to get default database: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		err = createWeaviateCollection(ctx, dbConfig, collectionName, embeddingModel, customFields, schemaType)
	case config.VectorDBTypeMock:
		err = createMockCollection(ctx, dbConfig, collectionName, embeddingModel, customFields)
	default:
		printError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		printError(fmt.Sprintf("Failed to create collection: %v", err))
		os.Exit(1)
	}

	printSuccess(fmt.Sprintf("Successfully created collection: %s", collectionName))
}
