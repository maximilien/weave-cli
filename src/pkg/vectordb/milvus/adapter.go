// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/reembedding/providers"
	"github.com/maximilien/weave-cli/src/pkg/storage"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Milvus client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
	llmClient    *llm.OpenAIClient
	imageStorage storage.ImageStorage // External storage for images >47KB
	pdfStorage   storage.ImageStorage // External storage for PDFs
}

// NewAdapter creates a new Milvus adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	milvusConfig := &Config{
		Address:          config.Address,
		Username:         config.Username,
		Password:         config.Password,
		APIKey:           config.APIKey,
		Database:         config.Database,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	client, err := NewClient(milvusConfig)
	if err != nil {
		return nil, err
	}

	// Create LLM client for embeddings (optional, only if OpenAI API key is available)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		var err error
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail - embeddings won't work but other operations will
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	// Initialize external storage if configured (v0.10.0+)
	var imageStorage, pdfStorage storage.ImageStorage

	// Image storage
	if config.ImageStorage != nil && config.ImageStorage.Enabled {
		imageStorage, err = storage.NewImageStorage(storage.Config{
			Type:       storage.StorageType(config.ImageStorage.Type),
			Endpoint:   config.ImageStorage.Endpoint,
			AccessKey:  config.ImageStorage.AccessKey,
			SecretKey:  config.ImageStorage.SecretKey,
			Region:     config.ImageStorage.Region,
			Bucket:     config.ImageStorage.Bucket,
			PathPrefix: config.ImageStorage.PathPrefix,
			UseSSL:     config.ImageStorage.UseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize image storage: %w", err)
		}
	}

	// PDF storage
	if config.PDFStorage != nil && config.PDFStorage.Enabled {
		pdfStorage, err = storage.NewImageStorage(storage.Config{
			Type:       storage.StorageType(config.PDFStorage.Type),
			Endpoint:   config.PDFStorage.Endpoint,
			AccessKey:  config.PDFStorage.AccessKey,
			SecretKey:  config.PDFStorage.SecretKey,
			Region:     config.PDFStorage.Region,
			Bucket:     config.PDFStorage.Bucket,
			PathPrefix: config.PDFStorage.PathPrefix,
			UseSSL:     config.PDFStorage.UseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PDF storage: %w", err)
		}
	}

	return &Adapter{
		Client:       client,
		llmClient:    llmClient,
		imageStorage: imageStorage,
		pdfStorage:   pdfStorage,
	}, nil
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	vectorizer := "text-embedding-3-small"
	return &vectordb.CollectionSchema{
		Vectorizer: vectorizer,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	// Milvus requires explicit schemas, so validation is minimal
	// The actual schema is created in CreateCollection
	return nil
}

// UpdateSchema updates a collection's schema
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Milvus doesn't support dynamic schema updates
	// Schema is defined at collection creation time
	return fmt.Errorf("Milvus: Milvus does not support schema updates after collection creation")
}

// Close closes the Milvus adapter and cleans up resources
func (a *Adapter) Close() error {
	if a.Client != nil {
		return a.Client.Close()
	}
	return nil
}

// createEmbeddingProvider creates an embedding provider for the given model
func (a *Adapter) createEmbeddingProvider(ctx context.Context, modelName string) (providers.EmbeddingProvider, error) {
	return providers.CreateProvider(ctx, modelName)
}
