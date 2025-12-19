// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"fmt"
	"os"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Factory implements the vectordb.ClientFactory interface for Neo4j
type Factory struct {
	llmClient *llm.OpenAIClient
}

// NewFactory creates a new Neo4j client factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates a new Neo4j client from vectordb.Config
func (f *Factory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	// Convert vectordb.Config to Neo4j Config
	neo4jConfig := &Config{
		URI:              config.URL,
		Username:         config.Username,
		Password:         config.Password,
		Database:         config.Database,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	// Set defaults
	if neo4jConfig.URI == "" {
		neo4jConfig.URI = "bolt://localhost:7687"
	}
	if neo4jConfig.Username == "" {
		neo4jConfig.Username = "neo4j"
	}
	if neo4jConfig.Database == "" {
		neo4jConfig.Database = "neo4j"
	}
	if neo4jConfig.MaxConnections == 0 {
		neo4jConfig.MaxConnections = 50
	}
	if neo4jConfig.Timeout == 0 {
		neo4jConfig.Timeout = 30 * time.Second
	}
	if neo4jConfig.VectorDimensions == 0 {
		neo4jConfig.VectorDimensions = 1536
	}
	if neo4jConfig.SimilarityMetric == "" {
		neo4jConfig.SimilarityMetric = "cosine"
	}

	client, err := NewClient(neo4jConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j client: %w", err)
	}

	// Create LLM client on-demand if not already created and API key is available
	llmClient := f.llmClient
	if llmClient == nil {
		if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
			llmClient, err = llm.NewOpenAIClient(openaiKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
			}
		}
	}

	return &Adapter{
		client:    client,
		llmClient: llmClient,
	}, nil
}

// GetSupportedTypes returns the supported database types
func (f *Factory) GetSupportedTypes() []vectordb.VectorDBType {
	return []vectordb.VectorDBType{
		vectordb.VectorDBTypeNeo4jLocal,
		vectordb.VectorDBTypeNeo4jCloud,
	}
}

// ValidateConfig validates the configuration
func (f *Factory) ValidateConfig(config *vectordb.Config) error {
	if config.Type != vectordb.VectorDBTypeNeo4jLocal && config.Type != vectordb.VectorDBTypeNeo4jCloud {
		return fmt.Errorf("Neo4j: invalid type: expected neo4j-local or neo4j-cloud, got %s", config.Type)
	}

	// Validate similarity metric if provided
	if config.SimilarityMetric != "" {
		if config.SimilarityMetric != "cosine" && config.SimilarityMetric != "euclidean" {
			return fmt.Errorf("Neo4j: invalid similarity metric: %s (must be 'cosine' or 'euclidean')", config.SimilarityMetric)
		}
	}

	return nil
}

// Register registers the Neo4j factory with the global registry
func init() {
	factory := NewFactory()
	vectordb.RegisterFactory(vectordb.VectorDBTypeNeo4jLocal, factory)
	vectordb.RegisterFactory(vectordb.VectorDBTypeNeo4jCloud, factory)
}
