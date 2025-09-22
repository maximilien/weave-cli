package cmd

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// createWeaviateClient creates a Weaviate client based on the configuration
func createWeaviateClient(cfg interface{}) (*weaviate.Client, error) {
	switch c := cfg.(type) {
	case *config.WeaviateCloudConfig:
		return weaviate.NewClient(&weaviate.Config{
			URL:    c.URL,
			APIKey: c.APIKey,
		})
	case *config.WeaviateLocalConfig:
		return weaviate.NewClient(&weaviate.Config{
			URL: c.URL,
		})
	default:
		return nil, fmt.Errorf("unsupported configuration type: %T", cfg)
	}
}
