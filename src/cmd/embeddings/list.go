// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package embeddings

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// EmbeddingModel represents an embedding model with its metadata
type EmbeddingModel struct {
	Name        string
	Provider    string
	Type        string // "text" or "image" or "multimodal"
	Dimensions  int
	APIKeyEnv   string // Environment variable name for API key
	Description string
	Module      string // Weaviate module name
}

// ListCmd represents the embeddings list command
var ListCmd = &cobra.Command{
	Use:     "list [COLLECTION]",
	Aliases: []string{"ls"},
	Short:   "List available embedding models",
	Long: `List all available embedding models for text and image vectorization.

By default, shows all embeddings available for any collection.
If a specific collection is provided, shows embeddings configured for that collection.

Embeddings are grouped by provider (OpenAI, Cohere, Hugging Face, etc.) and type (text/image).`,
	Example: `  # List all available embeddings
  weave embeddings list
  weave emb ls

  # List embeddings for a specific collection
  weave embeddings list MyCollection`,
	RunE: runList,
}

func init() {
	ListCmd.Flags().BoolP("verbose", "v", false, "Show detailed information including API requirements")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	verbose, _ := cmd.Flags().GetBool("verbose")

	var collectionName string
	if len(args) > 0 {
		collectionName = args[0]
	}

	// If collection specified, show collection-specific embeddings
	if collectionName != "" {
		// Load configuration with interactive help
		cfg, err := utils.LoadConfigWithInteractiveHelp()
		if err != nil {
			return err
		}

		// Get default database configuration
		dbConfig, err := cfg.GetDefaultDatabase()
		if err != nil {
			utils.HandleConfigError(err, true)
			return err
		}

		return showCollectionEmbeddings(ctx, dbConfig, collectionName, verbose)
	}

	// Show all available embeddings (no config needed)
	return showAllEmbeddings(verbose)
}

func showAllEmbeddings(verbose bool) error {
	embeddings := getAllEmbeddingModels()

	// Group by provider
	providerGroups := make(map[string][]EmbeddingModel)
	for _, emb := range embeddings {
		providerGroups[emb.Provider] = append(providerGroups[emb.Provider], emb)
	}

	// Sort providers
	providers := make([]string, 0, len(providerGroups))
	for provider := range providerGroups {
		providers = append(providers, provider)
	}
	sort.Strings(providers)

	// Print header
	color.New(color.FgCyan, color.Bold).Println("📊 Available Embedding Models")
	fmt.Println()

	// Print each provider group
	for _, provider := range providers {
		models := providerGroups[provider]

		// Sort models by type (text first, then image, then multimodal)
		sort.Slice(models, func(i, j int) bool {
			typeOrder := map[string]int{"text": 1, "image": 2, "multimodal": 3}
			if typeOrder[models[i].Type] != typeOrder[models[j].Type] {
				return typeOrder[models[i].Type] < typeOrder[models[j].Type]
			}
			return models[i].Name < models[j].Name
		})

		// Print provider header
		color.New(color.FgYellow, color.Bold).Printf("🏢 %s\n", provider)
		fmt.Println()

		// Print models
		for _, model := range models {
			printEmbeddingModel(model, verbose)
		}
		fmt.Println()
	}

	// Print usage note
	color.New(color.FgCyan).Println("💡 Usage:")
	fmt.Println("   Use --embedding-model flag when creating collections:")
	fmt.Println("   weave cols create MyCollection --embedding-model text-embedding-3-small")
	fmt.Println()
	fmt.Println("   Or view embeddings for a specific collection:")
	fmt.Println("   weave embeddings list MyCollection")

	return nil
}

func printEmbeddingModel(model EmbeddingModel, verbose bool) {
	// Type icon
	typeIcon := "📝"
	if model.Type == "image" {
		typeIcon = "🖼️ "
	} else if model.Type == "multimodal" {
		typeIcon = "🔄"
	}

	// Print model name
	color.New(color.FgGreen).Printf("  %s %s", typeIcon, model.Name)
	if model.Type != "text" {
		color.New(color.FgWhite).Printf(" (%s)", model.Type)
	}
	fmt.Println()

	if verbose {
		// Print description
		if model.Description != "" {
			fmt.Printf("     %s\n", model.Description)
		}

		// Print dimensions
		if model.Dimensions > 0 {
			color.New(color.FgWhite).Printf("     Dimensions: %d\n", model.Dimensions)
		}

		// Print module
		if model.Module != "" {
			color.New(color.FgWhite).Printf("     Weaviate Module: %s\n", model.Module)
		}

		// Print API key requirement
		if model.APIKeyEnv != "" {
			color.New(color.FgYellow).Printf("     ⚠️  Requires: %s\n", model.APIKeyEnv)
		}
		fmt.Println()
	}
}

func showCollectionEmbeddings(ctx context.Context, dbConfig *config.VectorDBConfig, collectionName string, verbose bool) error {
	// Print collection header
	color.New(color.FgCyan, color.Bold).Printf("📊 Embeddings for Collection: %s\n", collectionName)
	fmt.Println()

	// For now, show all available embeddings for the collection's database type
	// Future enhancement: query actual collection schema to get specific vectorizer
	color.New(color.FgYellow).Println("Note: Showing all available embeddings for this database type.")
	color.New(color.FgYellow).Printf("Database Type: %s\n", dbConfig.Type)
	fmt.Println()

	// Show commonly used embeddings based on database type
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		color.New(color.FgGreen).Println("Recommended Embeddings:")
		fmt.Println()
		recommendedModels := []string{
			"text-embedding-3-small",
			"text-embedding-3-large",
			"text-embedding-ada-002",
		}
		allModels := getAllEmbeddingModels()
		for _, modelName := range recommendedModels {
			for _, model := range allModels {
				if model.Name == modelName {
					printEmbeddingModel(model, verbose)
				}
			}
		}

		fmt.Println()
		color.New(color.FgCyan).Println("💡 To see all available embeddings:")
		fmt.Println("   weave embeddings list")

	case config.VectorDBTypeMock:
		color.New(color.FgGreen).Println("Mock Database:")
		fmt.Println("   Mock database accepts any embedding model name")
		fmt.Println("   Use --embedding-model flag when creating collections")

	default:
		return fmt.Errorf("unknown vector database type: %s", dbConfig.Type)
	}

	return nil
}

// GetAllEmbeddingModels returns all available embedding models
func GetAllEmbeddingModels() []EmbeddingModel {
	return getAllEmbeddingModels()
}

func getAllEmbeddingModels() []EmbeddingModel {
	return []EmbeddingModel{
		// OpenAI Text Embeddings
		{
			Name:        "text-embedding-3-small",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  1536,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Latest small embedding model, best performance/cost ratio",
			Module:      "text2vec-openai",
		},
		{
			Name:        "text-embedding-3-large",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  3072,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Latest large embedding model, highest accuracy",
			Module:      "text2vec-openai",
		},
		{
			Name:        "text-embedding-ada-002",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  1536,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Previous generation embedding model",
			Module:      "text2vec-openai",
		},
		{
			Name:        "ada",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  1024,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Legacy Ada model",
			Module:      "text2vec-openai",
		},
		{
			Name:        "babbage",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  2048,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Legacy Babbage model",
			Module:      "text2vec-openai",
		},
		{
			Name:        "curie",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  4096,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Legacy Curie model",
			Module:      "text2vec-openai",
		},
		{
			Name:        "davinci",
			Provider:    "OpenAI",
			Type:        "text",
			Dimensions:  12288,
			APIKeyEnv:   "OPENAI_API_KEY",
			Description: "Legacy Davinci model",
			Module:      "text2vec-openai",
		},

		// Cohere Embeddings
		{
			Name:        "embed-english-v3.0",
			Provider:    "Cohere",
			Type:        "text",
			Dimensions:  1024,
			APIKeyEnv:   "COHERE_API_KEY",
			Description: "English text embeddings v3",
			Module:      "text2vec-cohere",
		},
		{
			Name:        "embed-multilingual-v3.0",
			Provider:    "Cohere",
			Type:        "text",
			Dimensions:  1024,
			APIKeyEnv:   "COHERE_API_KEY",
			Description: "Multilingual text embeddings v3",
			Module:      "text2vec-cohere",
		},
		{
			Name:        "embed-english-light-v3.0",
			Provider:    "Cohere",
			Type:        "text",
			Dimensions:  384,
			APIKeyEnv:   "COHERE_API_KEY",
			Description: "Lightweight English embeddings",
			Module:      "text2vec-cohere",
		},

		// Hugging Face Embeddings
		{
			Name:        "sentence-transformers/all-MiniLM-L6-v2",
			Provider:    "Hugging Face",
			Type:        "text",
			Dimensions:  384,
			APIKeyEnv:   "",
			Description: "Fast and efficient sentence embeddings",
			Module:      "text2vec-huggingface",
		},
		{
			Name:        "sentence-transformers/all-mpnet-base-v2",
			Provider:    "Hugging Face",
			Type:        "text",
			Dimensions:  768,
			APIKeyEnv:   "",
			Description: "High-quality sentence embeddings",
			Module:      "text2vec-huggingface",
		},
		{
			Name:        "sentence-transformers/paraphrase-MiniLM-L6-v2",
			Provider:    "Hugging Face",
			Type:        "text",
			Dimensions:  384,
			APIKeyEnv:   "",
			Description: "Paraphrase detection model",
			Module:      "text2vec-huggingface",
		},

		// Weaviate Embeddings (built-in)
		{
			Name:        "weaviate-default",
			Provider:    "Weaviate",
			Type:        "text",
			Dimensions:  384,
			APIKeyEnv:   "",
			Description: "Built-in Weaviate text embeddings (free)",
			Module:      "text2vec-weaviate",
		},

		// Google PaLM Embeddings
		{
			Name:        "textembedding-gecko@001",
			Provider:    "Google PaLM",
			Type:        "text",
			Dimensions:  768,
			APIKeyEnv:   "PALM_API_KEY",
			Description: "Google PaLM text embeddings",
			Module:      "text2vec-palm",
		},

		// Image Embeddings
		{
			Name:        "clip-vit-base-patch32",
			Provider:    "OpenAI CLIP",
			Type:        "multimodal",
			Dimensions:  512,
			APIKeyEnv:   "",
			Description: "Vision-language model for images and text",
			Module:      "multi2vec-clip",
		},
		{
			Name:        "resnet50",
			Provider:    "Image Recognition",
			Type:        "image",
			Dimensions:  2048,
			APIKeyEnv:   "",
			Description: "ResNet-50 image embeddings",
			Module:      "img2vec-neural",
		},

		// AWS Embeddings
		{
			Name:        "amazon.titan-embed-text-v1",
			Provider:    "AWS Bedrock",
			Type:        "text",
			Dimensions:  1536,
			APIKeyEnv:   "AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY",
			Description: "Amazon Titan text embeddings",
			Module:      "text2vec-aws",
		},

		// Jina AI Embeddings
		{
			Name:        "jina-embeddings-v2-base-en",
			Provider:    "Jina AI",
			Type:        "text",
			Dimensions:  768,
			APIKeyEnv:   "JINA_API_KEY",
			Description: "Jina AI English embeddings",
			Module:      "text2vec-jinaai",
		},
	}
}

// IsAPIKeySet checks if required API key is set in environment
func IsAPIKeySet(envVars string) bool {
	if envVars == "" {
		return true // No API key required
	}

	// Split multiple env vars (comma-separated)
	vars := strings.Split(envVars, ",")
	for _, v := range vars {
		v = strings.TrimSpace(v)
		if os.Getenv(v) == "" {
			return false
		}
	}
	return true
}
