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
	Name               string
	Provider           string
	Type               string // "text" or "image" or "multimodal"
	Dimensions         int
	APIKeyEnv          string // Environment variable name for API key
	Description        string
	Module             string   // Weaviate module name
	SupportedDatabases []string // Database types that support this embedding
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
	ListCmd.Flags().StringP("database", "d", "", "Filter embeddings by database type (weaviate-cloud, weaviate-local, supabase, mock)")
	ListCmd.Flags().BoolP("show-compatibility", "c", false, "Show database compatibility for each embedding")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	verbose, _ := cmd.Flags().GetBool("verbose")
	databaseFilter, _ := cmd.Flags().GetString("database")
	showCompat, _ := cmd.Flags().GetBool("show-compatibility")

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
	return showAllEmbeddings(verbose, databaseFilter, showCompat)
}

func showAllEmbeddings(verbose bool, databaseFilter string, showCompat bool) error {
	embeddings := getAllEmbeddingModels()

	// Filter by database type if specified
	if databaseFilter != "" {
		filtered := []EmbeddingModel{}
		for _, emb := range embeddings {
			for _, db := range emb.SupportedDatabases {
				if db == databaseFilter {
					filtered = append(filtered, emb)
					break
				}
			}
		}
		embeddings = filtered

		if len(embeddings) == 0 {
			color.New(color.FgYellow).Printf("⚠️  No embeddings found for database type: %s\n", databaseFilter)
			return nil
		}
	}

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
	if databaseFilter != "" {
		color.New(color.FgCyan, color.Bold).Printf("📊 Available Embedding Models for %s\n", databaseFilter)
	} else {
		color.New(color.FgCyan, color.Bold).Println("📊 Available Embedding Models")
	}
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
			printEmbeddingModel(model, verbose, showCompat)
		}
		fmt.Println()
	}

	// Print usage note
	color.New(color.FgCyan).Println("💡 Usage:")
	fmt.Println("   Use --embedding-model flag when creating collections:")
	fmt.Println("   weave cols create MyCollection --embedding-model text-embedding-3-small")
	fmt.Println()
	fmt.Println("   Filter by database type:")
	fmt.Println("   weave embeddings list --database supabase")
	fmt.Println()
	fmt.Println("   Show compatibility information:")
	fmt.Println("   weave embeddings list --show-compatibility")

	return nil
}

func printEmbeddingModel(model EmbeddingModel, verbose bool, showCompat bool) {
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

	// Show compatibility if requested (or if verbose)
	if showCompat || verbose {
		if len(model.SupportedDatabases) > 0 {
			dbIcons := map[string]string{
				"weaviate-cloud": "☁️",
				"weaviate-local": "🏠",
				"supabase":       "🐘",
				"mock":           "🧪",
			}

			dbDisplayNames := map[string]string{
				"weaviate-cloud": "Weaviate Cloud",
				"weaviate-local": "Weaviate Local",
				"supabase":       "Supabase",
				"mock":           "Mock",
			}

			color.New(color.FgCyan).Printf("     Supported: ")
			for i, db := range model.SupportedDatabases {
				icon := dbIcons[db]
				if icon == "" {
					icon = "📦"
				}
				displayName := dbDisplayNames[db]
				if displayName == "" {
					displayName = db
				}
				fmt.Printf("%s %s", icon, displayName)
				if i < len(model.SupportedDatabases)-1 {
					fmt.Print(", ")
				}
			}
			fmt.Println()
		}
	}

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

	// Get database type as string
	dbType := string(dbConfig.Type)

	// Show database type
	dbIcons := map[string]string{
		"weaviate-cloud": "☁️",
		"weaviate-local": "🏠",
		"supabase":       "🐘",
		"mock":           "🧪",
	}
	icon := dbIcons[dbType]
	if icon == "" {
		icon = "📦"
	}

	color.New(color.FgYellow).Printf("%s Database Type: %s\n", icon, dbType)
	fmt.Println()

	// Get all embeddings for this database type
	allModels := getAllEmbeddingModels()
	supportedModels := []EmbeddingModel{}
	for _, model := range allModels {
		for _, db := range model.SupportedDatabases {
			if db == dbType {
				supportedModels = append(supportedModels, model)
				break
			}
		}
	}

	if len(supportedModels) == 0 {
		color.New(color.FgYellow).Printf("⚠️  No embeddings found for database type: %s\n", dbType)
		return nil
	}

	// Show recommended embeddings based on database type
	switch dbConfig.Type {
	case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
		color.New(color.FgGreen).Println("Recommended Embeddings (OpenAI):")
		fmt.Println()
		recommendedModels := []string{
			"text-embedding-3-small",
			"text-embedding-3-large",
			"text-embedding-ada-002",
		}
		for _, modelName := range recommendedModels {
			for _, model := range supportedModels {
				if model.Name == modelName {
					printEmbeddingModel(model, verbose, true)
				}
			}
		}

		fmt.Println()
		color.New(color.FgCyan).Printf("💡 %d total embeddings available for %s\n", len(supportedModels), dbType)
		fmt.Println("   To see all:")
		fmt.Printf("   weave embeddings list --database %s\n", dbType)

	case config.VectorDBTypeSupabase:
		color.New(color.FgGreen).Println("Supported Embeddings:")
		fmt.Println()

		// For Supabase, show all supported embeddings (limited set)
		for _, model := range supportedModels {
			printEmbeddingModel(model, verbose, true)
		}

		fmt.Println()
		color.New(color.FgYellow).Printf("ℹ️  Note: Supabase currently supports OpenAI embeddings and manual embeddings.\n")
		color.New(color.FgYellow).Println("   Other vectorizer types (Cohere, Hugging Face, etc.) are not yet supported.")

	case config.VectorDBTypeMock:
		color.New(color.FgGreen).Println("Mock Database:")
		fmt.Println("   Mock database accepts any embedding model name")
		fmt.Println("   Embeddings are simulated for testing purposes")
		fmt.Println()
		fmt.Printf("   %d embeddings available\n", len(supportedModels))

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
	// Database type constants for clarity
	weaviateCloud := "weaviate-cloud"
	weaviateLocal := "weaviate-local"
	supabase := "supabase"
	mock := "mock"

	// All databases for wide support
	allDatabases := []string{weaviateCloud, weaviateLocal, supabase, mock}
	// Weaviate only (cloud and local)
	weaviateOnly := []string{weaviateCloud, weaviateLocal, mock}

	return []EmbeddingModel{
		// OpenAI Text Embeddings - Supported by all databases
		{
			Name:               "text-embedding-3-small",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         1536,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Latest small embedding model, best performance/cost ratio",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "text-embedding-3-large",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         3072,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Latest large embedding model, highest accuracy",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "text-embedding-ada-002",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         1536,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Previous generation embedding model",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "ada",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         1024,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Legacy Ada model",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "babbage",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         2048,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Legacy Babbage model",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "curie",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         4096,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Legacy Curie model",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},
		{
			Name:               "davinci",
			Provider:           "OpenAI",
			Type:               "text",
			Dimensions:         12288,
			APIKeyEnv:          "OPENAI_API_KEY",
			Description:        "Legacy Davinci model",
			Module:             "text2vec-openai",
			SupportedDatabases: allDatabases,
		},

		// Cohere Embeddings - Weaviate only
		{
			Name:               "embed-english-v3.0",
			Provider:           "Cohere",
			Type:               "text",
			Dimensions:         1024,
			APIKeyEnv:          "COHERE_API_KEY",
			Description:        "English text embeddings v3",
			Module:             "text2vec-cohere",
			SupportedDatabases: weaviateOnly,
		},
		{
			Name:               "embed-multilingual-v3.0",
			Provider:           "Cohere",
			Type:               "text",
			Dimensions:         1024,
			APIKeyEnv:          "COHERE_API_KEY",
			Description:        "Multilingual text embeddings v3",
			Module:             "text2vec-cohere",
			SupportedDatabases: weaviateOnly,
		},
		{
			Name:               "embed-english-light-v3.0",
			Provider:           "Cohere",
			Type:               "text",
			Dimensions:         384,
			APIKeyEnv:          "COHERE_API_KEY",
			Description:        "Lightweight English embeddings",
			Module:             "text2vec-cohere",
			SupportedDatabases: weaviateOnly,
		},

		// Hugging Face Embeddings - Weaviate only
		{
			Name:               "sentence-transformers/all-MiniLM-L6-v2",
			Provider:           "Hugging Face",
			Type:               "text",
			Dimensions:         384,
			APIKeyEnv:          "",
			Description:        "Fast and efficient sentence embeddings",
			Module:             "text2vec-huggingface",
			SupportedDatabases: weaviateOnly,
		},
		{
			Name:               "sentence-transformers/all-mpnet-base-v2",
			Provider:           "Hugging Face",
			Type:               "text",
			Dimensions:         768,
			APIKeyEnv:          "",
			Description:        "High-quality sentence embeddings",
			Module:             "text2vec-huggingface",
			SupportedDatabases: weaviateOnly,
		},
		{
			Name:               "sentence-transformers/paraphrase-MiniLM-L6-v2",
			Provider:           "Hugging Face",
			Type:               "text",
			Dimensions:         384,
			APIKeyEnv:          "",
			Description:        "Paraphrase detection model",
			Module:             "text2vec-huggingface",
			SupportedDatabases: weaviateOnly,
		},

		// Weaviate Embeddings (built-in) - Weaviate only
		{
			Name:               "weaviate-default",
			Provider:           "Weaviate",
			Type:               "text",
			Dimensions:         384,
			APIKeyEnv:          "",
			Description:        "Built-in Weaviate text embeddings (free)",
			Module:             "text2vec-weaviate",
			SupportedDatabases: weaviateOnly,
		},

		// Google PaLM Embeddings - Weaviate only
		{
			Name:               "textembedding-gecko@001",
			Provider:           "Google PaLM",
			Type:               "text",
			Dimensions:         768,
			APIKeyEnv:          "PALM_API_KEY",
			Description:        "Google PaLM text embeddings",
			Module:             "text2vec-palm",
			SupportedDatabases: weaviateOnly,
		},

		// Image Embeddings - Weaviate only
		{
			Name:               "clip-vit-base-patch32",
			Provider:           "OpenAI CLIP",
			Type:               "multimodal",
			Dimensions:         512,
			APIKeyEnv:          "",
			Description:        "Vision-language model for images and text",
			Module:             "multi2vec-clip",
			SupportedDatabases: weaviateOnly,
		},
		{
			Name:               "resnet50",
			Provider:           "Image Recognition",
			Type:               "image",
			Dimensions:         2048,
			APIKeyEnv:          "",
			Description:        "ResNet-50 image embeddings",
			Module:             "img2vec-neural",
			SupportedDatabases: weaviateOnly,
		},

		// AWS Embeddings - Weaviate only
		{
			Name:               "amazon.titan-embed-text-v1",
			Provider:           "AWS Bedrock",
			Type:               "text",
			Dimensions:         1536,
			APIKeyEnv:          "AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY",
			Description:        "Amazon Titan text embeddings",
			Module:             "text2vec-aws",
			SupportedDatabases: weaviateOnly,
		},

		// Jina AI Embeddings - Weaviate only
		{
			Name:               "jina-embeddings-v2-base-en",
			Provider:           "Jina AI",
			Type:               "text",
			Dimensions:         768,
			APIKeyEnv:          "JINA_API_KEY",
			Description:        "Jina AI English embeddings",
			Module:             "text2vec-jinaai",
			SupportedDatabases: weaviateOnly,
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
