package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// CreateWeaviateClient creates a Weaviate client from configuration
func CreateWeaviateClient(cfg *config.VectorDBConfig) (*weaviate.Client, error) {
	// Convert VectorDBConfig to weaviate.Config
	weaviateConfig := &weaviate.Config{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
	}

	client, err := weaviate.NewClient(weaviateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %v", err)
	}

	return client, nil
}

// CreateMockClient creates a mock client from configuration
func CreateMockClient(cfg *config.VectorDBConfig) *mock.Client {
	// Convert VectorDBConfig to MockConfig for backward compatibility
	mockConfig := &config.MockConfig{
		Enabled:            cfg.Enabled,
		SimulateEmbeddings: cfg.SimulateEmbeddings,
		EmbeddingDimension: cfg.EmbeddingDimension,
		Collections:        make([]config.MockCollection, len(cfg.Collections)),
	}

	for i, col := range cfg.Collections {
		mockConfig.Collections[i] = config.MockCollection(col)
	}

	return mock.NewClient(mockConfig)
}
