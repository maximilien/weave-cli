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

// CreateCmd represents the document create command
var CreateCmd = &cobra.Command{
	Use:     "create COLLECTION_NAME FILE_PATH",
	Aliases: []string{"c"},
	Short:   "Create a document from a file",
	Long: `Create a document in a collection from a file.

Supported file types:
- Text files (.txt, .md, .json, etc.) - Content goes to 'text' field
- Image files (.jpg, .jpeg, .png, .gif, etc.) - Base64 data goes to 'image_data' field
- PDF files (.pdf) - Text extracted and chunked, images extracted separately

The command will automatically:
- Detect file type and process accordingly
- Generate appropriate metadata
- Chunk text content (default 5000 chars, configurable with --chunk-size)
- Extract images from PDFs with OCR and EXIF data
- Create documents following WeaveDocs/WeaveImages schema (default) or RagMeDocs/RagMeImages (legacy)

For PDF files with images:
- Text chunks go to the main collection
- Extracted images go to a separate collection (use --image-collection)
- Images include OCR text, EXIF data, and captions when available
- Use --skip-all-images for text-only extraction (no image processing)

Examples:
  weave docs create MyCollection document.txt
  weave docs create MyCollection image.jpg
  weave docs create MyCollection document.pdf --chunk-size 500
  weave docs create WeaveDocs document.pdf --image-collection WeaveImages
  weave docs create RagMeDocs document.pdf --image-col RagMeImages
  weave docs create MyDocs document.pdf --skip-all-images  # text only
  weave docs create MyCollection document.txt --embedding text-embedding-3-small`,
	Args: cobra.ExactArgs(2),
	Run:  runDocumentCreate,
}

func init() {
	DocumentCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().IntP("chunk-size", "s", 5000, "Chunk size for text content (default: 5000 characters)")
	CreateCmd.Flags().StringP("image-collection", "", "", "Collection name for extracted PDF images (default: same as main collection)")
	CreateCmd.Flags().StringP("image-col", "", "", "Alias for --image-collection")
	CreateCmd.Flags().StringP("image-cols", "", "", "Alias for --image-collection")
	CreateCmd.Flags().Bool("skip-all-images", false, "Skip all image extraction from PDFs (text-only mode)")
	CreateCmd.Flags().Bool("skip-small-images", true, "Skip small images when extracting from PDFs (default: true)")
	CreateCmd.Flags().Int("min-image-size", 5120, "Minimum image size in bytes (default: 5120 = 5KB)")
	CreateCmd.Flags().Int("batch-size", 10, "Number of images to process before pausing for memory cleanup (default: 10)")
	CreateCmd.Flags().Int("max-metadata-length", 2000, "Maximum length for image metadata text fields (surrounding_text, ocr_content, section_heading). Set to 0 for unlimited. Recommended: 2000 for Milvus, 8000 for Weaviate (default: 2000)")
	CreateCmd.Flags().String("create-report", "", "Create a new CSV report of processing results (default: <filename>.csv in current directory)")
	CreateCmd.Flags().String("append-report", "", "Append to an existing CSV report")
	CreateCmd.Flags().StringP("embedding", "e", "", "Embedding model to use for this document (e.g., text-embedding-3-small, text-embedding-ada-002)")
	CreateCmd.Flags().Bool("json", false, "Output in JSON format")
}

func runDocumentCreate(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	filePath := args[1]
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	imageCollection, _ := cmd.Flags().GetString("image-collection")
	imageCol, _ := cmd.Flags().GetString("image-col")
	imageCols, _ := cmd.Flags().GetString("image-cols")
	skipAllImages, _ := cmd.Flags().GetBool("skip-all-images")
	skipSmallImages, _ := cmd.Flags().GetBool("skip-small-images")
	minImageSize, _ := cmd.Flags().GetInt("min-image-size")
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	maxMetadataLength, _ := cmd.Flags().GetInt("max-metadata-length")
	createReport, _ := cmd.Flags().GetString("create-report")
	appendReport, _ := cmd.Flags().GetString("append-report")
	embeddingModel, _ := cmd.Flags().GetString("embedding")

	// If skip-all-images is set, clear the image collection to skip image processing
	if skipAllImages {
		imageCollection = ""
	} else {
		// Use image collection from flags if provided
		if imageCollection == "" {
			imageCollection = imageCol
		}
		if imageCollection == "" {
			imageCollection = imageCols
		}
	}

	// Determine report path
	reportPath := ""
	reportMode := ""
	if createReport != "" || appendReport != "" {
		if createReport != "" {
			reportPath = createReport
			reportMode = "create"
		} else {
			reportPath = appendReport
			reportMode = "append"
		}

		// If report path is empty, generate default from filename
		if reportPath == "" {
			reportPath = utils.GenerateDefaultReportPath(filePath)
		}
	}

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get selected databases
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	// Smart database selection for single-database operations
	dbConfig := utils.HandleSingleDatabaseSelection(ctx, selection, cfg, collectionName, fmt.Sprintf("weave docs create %s %s", collectionName, filePath))

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		if err := utils.CreateWeaviateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeSupabase, config.VectorDBTypeSupabaseCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeMongoDB, config.VectorDBTypeMongoDBCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeMilvusLocal, config.VectorDBTypeMilvusCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeChromaLocal, config.VectorDBTypeChromaCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeQdrantLocal, config.VectorDBTypeQdrantCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeNeo4jLocal, config.VectorDBTypeNeo4jCloud:
		if err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode, embeddingModel); err != nil {
			utils.PrintError(utils.FormatCreationError("document", err))
			os.Exit(1)
		}
	case config.VectorDBTypeMock:
		utils.CreateMockDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, maxMetadataLength, reportPath, reportMode)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
