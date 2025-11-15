// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the MongoDB client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
}

// NewAdapter creates a new MongoDB adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	mongoConfig := &Config{
		URI:              config.URL,
		Database:         config.Database,
		Timeout:          config.Timeout,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
	}

	client, err := NewClient(mongoConfig)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		Client: client,
	}, nil
}
