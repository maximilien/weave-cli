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
You can use the --text or --image flags for standard schemas, use a named
schema from config.yaml/schemas_dir, specify a schema file, or use custom
fields and embedding model.

For --text and --image collections, you must specify how to handle metadata:
- --flat-metadata: Flatten metadata fields into individual properties
- --json-metadata: Keep metadata as a single JSON field

Examples:
  # Create text collection with flattened metadata (recommended)
  weave collection create MyDocsCol --text --flat-metadata

  # Create text collection with JSON metadata
  weave collection create MyDocsCol --text --json-metadata

  # Create image collection with flattened metadata
  weave collection create MyImagesCol --image --flat-metadata

  # Create collection using a named schema from config.yaml or schemas_dir
  weave collection create MyDocsCol --schema RagMeDocs
  weave collection create MyDocsCol --schema TestSchema

  # Create collection from a YAML schema file
  weave collection create MyCollection --schema-yaml-file schema.yaml

  # Create collection with default schema and custom options
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
	CreateCmd.Flags().StringP("schema-yaml-file", "", "", "Create collection from YAML schema file")
	CreateCmd.Flags().String("schema", "", "Use a named schema from config.yaml or schemas_dir (e.g., RagMeDocs, WeaveDocs, WeaveImages)")
	CreateCmd.Flags().Bool("text", false, "Create text collection using WeaveDocs schema (same as --schema WeaveDocs)")
	CreateCmd.Flags().Bool("image", false, "Create image collection using WeaveImages schema (same as --schema WeaveImages)")
	CreateCmd.Flags().Bool("flat-metadata", false, "Flatten metadata fields into individual properties (requires --text or --image)")
	CreateCmd.Flags().Bool("json-metadata", false, "Keep metadata as a single JSON field (requires --text or --image)")
}

func runCollectionCreate(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	embeddingModel, _ := cmd.Flags().GetString("embedding-model")
	fieldsStr, _ := cmd.Flags().GetString("fields")
	schemaYAMLFile, _ := cmd.Flags().GetString("schema-yaml-file")
	schemaName, _ := cmd.Flags().GetString("schema")
	useTextSchema, _ := cmd.Flags().GetBool("text")
	useImageSchema, _ := cmd.Flags().GetBool("image")
	useFlatMetadata, _ := cmd.Flags().GetBool("flat-metadata")
	useJsonMetadata, _ := cmd.Flags().GetBool("json-metadata")

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

	// Validate metadata flags
	if useFlatMetadata && useJsonMetadata {
		utils.PrintError("Cannot use both --flat-metadata and --json-metadata flags")
		os.Exit(1)
	}

	// Handle --text flag (use WeaveDocs schema)
	if useTextSchema {
		if schemaName != "" || useImageSchema {
			utils.PrintError("Cannot use --text with --schema or --image flags")
			os.Exit(1)
		}
		schemaName = "WeaveDocs"
	}

	// Handle --image flag (use WeaveImages schema)
	if useImageSchema {
		if schemaName != "" || useTextSchema {
			utils.PrintError("Cannot use --image with --schema or --text flags")
			os.Exit(1)
		}
		schemaName = "WeaveImages"
	}

	// Validate metadata flags are only used with --text or --image
	if (useFlatMetadata || useJsonMetadata) && !useTextSchema && !useImageSchema {
		utils.PrintError("--flat-metadata and --json-metadata flags require --text or --image")
		os.Exit(1)
	}

	// Require metadata flag when using --text or --image
	if (useTextSchema || useImageSchema) && !useFlatMetadata && !useJsonMetadata {
		utils.PrintError("--text and --image collections require either --flat-metadata or --json-metadata")
		os.Exit(1)
	}

	// Handle named schema from config.yaml or schemas_dir
	if schemaName != "" {
		// Validate that the schema exists
		_, err := cfg.GetSchema(schemaName)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Schema '%s' not found in config.yaml or schemas_dir: %v", schemaName, err))
			utils.PrintInfo("Available schemas:")
			schemaNames := cfg.ListSchemas()
			if len(schemaNames) == 0 {
				utils.PrintInfo("  (none)")
			} else {
				for _, name := range schemaNames {
					utils.PrintInfo(fmt.Sprintf("  - %s", name))
				}
			}
			os.Exit(1)
		}

		// Determine metadata handling mode
		metadataMode := "json" // default
		if useFlatMetadata {
			metadataMode = "flat"
		} else if useJsonMetadata {
			metadataMode = "json"
		}

		switch dbConfig.Type {
		case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
			err = utils.CreateWeaviateCollectionFromConfigSchema(ctx, cfg, dbConfig, collectionName, schemaName, metadataMode)
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
		err = utils.CreateWeaviateCollection(ctx, dbConfig, collectionName, embeddingModel, customFields)
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
