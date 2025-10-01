// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"github.com/spf13/cobra"
)

// CreateCmd represents the collection create command
var CreateCmd = &cobra.Command{
	Use:     "create COLLECTION_NAME",
	Aliases: []string{"c"},
	Short:   "Create a collection",
	Long: `Create a new collection in the vector database.

This command creates a collection with the specified name and schema.
You can use a named schema from config.yaml, specify a schema file,
or use custom fields and embedding model.

Examples:
  # Create collection using a named schema from config.yaml
  weave collection create MyDocsCol --schema RagMeDocs
  weave collection create MyImagesCol --schema RagMeImages

  # Create collection from a YAML schema file
  weave collection create MyCollection --schema-yaml-file schema.yaml

  # Create collection with default schema
  weave collection create MyCollection
  weave collection create MyCollection --embedding-model text-embedding-ada-002
  weave collection create MyCollection --fields "title:string,description:text"`,
	Args: cobra.ExactArgs(1),
	Run:  runCollectionCreate,
}

func init() {
	CollectionCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().StringP("embedding-model", "e", "text-embedding-ada-002", "Embedding model to use")
	CreateCmd.Flags().StringP("fields", "f", "", "Custom fields (format: field1:type1,field2:type2)")
	CreateCmd.Flags().StringP("schema-type", "s", "default", "Schema type (default, ragmedocs, ragmeimages)")
	CreateCmd.Flags().StringP("schema-yaml-file", "", "", "Create collection from YAML schema file")
	CreateCmd.Flags().String("schema", "", "Use a named schema from config.yaml (e.g., RagMeDocs, RagMeImages)")
}

func runCollectionCreate(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	embeddingModel, _ := cmd.Flags().GetString("embedding-model")
	fieldsStr, _ := cmd.Flags().GetString("fields")
	schemaType, _ := cmd.Flags().GetString("schema-type")
	schemaYAMLFile, _ := cmd.Flags().GetString("schema-yaml-file")
	schemaName, _ := cmd.Flags().GetString("schema")

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

	// Handle named schema from config.yaml
	if schemaName != "" {
		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			err = utils.CreateWeaviateCollectionFromConfigSchema(ctx, cfg, dbConfig, collectionName, schemaName)
		case config.VectorDBTypeMock:
			utils.PrintError("Schema creation from config.yaml not yet supported for mock database")
			os.Exit(1)
		default:
			utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to create collection from schema '%s': %v", schemaName, err))
			os.Exit(1)
		}

		utils.PrintSuccess(fmt.Sprintf("Successfully created collection '%s' using schema '%s'", collectionName, schemaName))
		return
	}

	// Handle schema file creation
	if schemaYAMLFile != "" {
		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			err = utils.CreateWeaviateCollectionFromSchemaFile(ctx, dbConfig, collectionName, schemaYAMLFile)
		case config.VectorDBTypeMock:
			utils.PrintError("Schema file creation not yet supported for mock database")
			os.Exit(1)
		default:
			utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
			os.Exit(1)
		}

		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to create collection from schema file: %v", err))
			os.Exit(1)
		}

		utils.PrintSuccess(fmt.Sprintf("Successfully created collection '%s' from schema file: %s", collectionName, schemaYAMLFile))
		return
	}

	// Parse custom fields
	var customFields []weaviate.FieldDefinition
	if fieldsStr != "" {
		fields, err := utils.ParseFieldDefinitions(fieldsStr)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to parse fields: %v", err))
			os.Exit(1)
		}
		customFields = fields
	}

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		err = utils.CreateWeaviateCollection(ctx, dbConfig, collectionName, embeddingModel, customFields, schemaType)
	case config.VectorDBTypeMock:
		err = utils.CreateMockCollection(ctx, dbConfig, collectionName, embeddingModel, customFields)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}

	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create collection: %v", err))
		os.Exit(1)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully created collection: %s", collectionName))
}
