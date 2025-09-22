package cmd

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// createWeaviateClient creates a Weaviate client based on the configuration
func createWeaviateClient(cfg interface{}) (*weaviate.Client, error) {
	if cloudConfig, ok := cfg.(*config.WeaviateCloudConfig); ok {
		return weaviate.NewClient(&weaviate.Config{
			URL:    cloudConfig.URL,
			APIKey: cloudConfig.APIKey,
		})
	}

	if localConfig, ok := cfg.(*config.WeaviateLocalConfig); ok {
		return weaviate.NewClient(&weaviate.Config{
			URL: localConfig.URL,
		})
	}

	return nil, fmt.Errorf("unsupported configuration type: %T", cfg)
}
