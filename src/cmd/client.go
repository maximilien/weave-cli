package cmd

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// createWeaviateClient creates a Weaviate client based on the configuration
func createWeaviateClient(cfg *config.VectorDBConfig) (*weaviate.Client, error) {
	switch cfg.Type {
	case config.VectorDBTypeCloud:
		return weaviate.NewClient(&weaviate.Config{
			URL:    cfg.URL,
			APIKey: cfg.APIKey,
		})
	case config.VectorDBTypeLocal:
		return weaviate.NewClient(&weaviate.Config{
			URL: cfg.URL,
		})
	default:
		return nil, fmt.Errorf("unsupported configuration type: %s", cfg.Type)
	}
}
