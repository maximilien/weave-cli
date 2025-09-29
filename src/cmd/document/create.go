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
- Chunk text content (default 1000 chars, configurable with --chunk-size)
- Extract images from PDFs with OCR and EXIF data
- Create documents following RagMeDocs/RagMeImages schema

For PDF files with images:
- Text chunks go to the main collection
- Extracted images go to a separate collection (use --image-collection)
- Images include OCR text, EXIF data, and captions when available

Examples:
  weave docs create MyCollection document.txt
  weave docs create MyCollection image.jpg
  weave docs create MyCollection document.pdf --chunk-size 500
  weave docs create WeaveDocs document.pdf --image-collection WeaveImages
  weave docs create WeaveDocs document.pdf --image-col RagMeImages`,
	Args: cobra.ExactArgs(2),
	Run:  runDocumentCreate,
}

func init() {
	DocumentCmd.AddCommand(CreateCmd)

	CreateCmd.Flags().IntP("chunk-size", "s", 1000, "Chunk size for text content (default: 1000 characters)")
	CreateCmd.Flags().StringP("image-collection", "", "", "Collection name for extracted PDF images (default: same as main collection)")
	CreateCmd.Flags().StringP("image-col", "", "", "Alias for --image-collection")
	CreateCmd.Flags().StringP("image-cols", "", "", "Alias for --image-collection")
	CreateCmd.Flags().Bool("skip-small-images", false, "Skip small images when extracting from PDFs")
	CreateCmd.Flags().Int("min-image-size", 10240, "Minimum image size in bytes (default: 10240 = 10KB)")
}

func runDocumentCreate(cmd *cobra.Command, args []string) {
	collectionName := args[0]
	filePath := args[1]
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	imageCollection, _ := cmd.Flags().GetString("image-collection")
	imageCol, _ := cmd.Flags().GetString("image-col")
	imageCols, _ := cmd.Flags().GetString("image-cols")
	skipSmallImages, _ := cmd.Flags().GetBool("skip-small-images")
	minImageSize, _ := cmd.Flags().GetInt("min-image-size")

	// Use image collection from flags if provided
	if imageCollection == "" {
		imageCollection = imageCol
	}
	if imageCollection == "" {
		imageCollection = imageCols
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

	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		utils.CreateWeaviateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize)
	case config.VectorDBTypeMock:
		utils.CreateMockDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize)
	default:
		utils.PrintError(fmt.Sprintf("Unknown vector database type: %s", dbConfig.Type))
		os.Exit(1)
	}
}
