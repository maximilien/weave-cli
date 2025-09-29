// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// createWeaviateClient creates a Weave client based on the configuration
func createWeaviateClient(cfg *config.VectorDBConfig) (*weaviate.WeaveClient, error) {
	switch cfg.Type {
	case config.VectorDBTypeCloud:
		return weaviate.NewWeaveClient(&weaviate.Config{
			URL:          cfg.URL,
			APIKey:       cfg.APIKey,
			OpenAIAPIKey: cfg.OpenAIAPIKey,
		})
	case config.VectorDBTypeLocal:
		return weaviate.NewWeaveClient(&weaviate.Config{
			URL:          cfg.URL,
			OpenAIAPIKey: cfg.OpenAIAPIKey,
		})
	default:
		return nil, fmt.Errorf("unsupported configuration type: %s", cfg.Type)
	}
}
