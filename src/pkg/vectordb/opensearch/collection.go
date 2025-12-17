// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a new index with k-NN settings
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Default vector dimensions
	vectorDims := a.config.VectorDimensions
	if vectorDims == 0 {
		vectorDims = 384 // Default embedding dimension
	}

	// Default distance metric
	distanceMetric := a.config.SimilarityMetric
	if distanceMetric == "" {
		distanceMetric = "l2" // Euclidean distance
	}

	// Create index mapping with k-NN settings
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"knn":                      true,
				"knn.algo_param.ef_search": 100,
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"vector": map[string]interface{}{
					"type":      "knn_vector",
					"dimension": vectorDims,
					"method": map[string]interface{}{
						"name":       "hnsw",
						"space_type": distanceMetric,
						"engine":     "lucene",
						"parameters": map[string]interface{}{
							"ef_construction": 128,
							"m":               16,
						},
					},
				},
				"metadata": map[string]interface{}{
					"type":    "object",
					"enabled": true,
				},
			},
		},
	}

	resp, err := a.client.Indices.Create(
		ctx,
		opensearchapi.IndicesCreateReq{
			Index: name,
			Body:  opensearchutil.NewJSONReader(mapping),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if !resp.Acknowledged {
		return fmt.Errorf("index creation not acknowledged")
	}

	return nil
}

// DeleteCollection deletes an index
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	resp, err := a.client.Indices.Delete(
		ctx,
		opensearchapi.IndicesDeleteReq{
			Indices: []string{name},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}

	if !resp.Acknowledged {
		return fmt.Errorf("index deletion not acknowledged")
	}

	return nil
}

// CollectionExists checks if an index exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := a.client.Indices.Exists(
		ctx,
		opensearchapi.IndicesExistsReq{
			Indices: []string{name},
		},
	)
	if err != nil {
		// Exists returns 404 error if not exists
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}

	return true, nil
}

// ListCollections returns a list of all indices
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	// TODO: Implement proper Cat API parsing
	return nil, fmt.Errorf("ListCollections not yet fully implemented")
}

// GetCollectionInfo returns information about an index
func (a *Adapter) GetCollectionInfo(ctx context.Context, name string) (*vectordb.CollectionInfo, error) {
	// TODO: Implement proper stats parsing
	return &vectordb.CollectionInfo{
		Name:  name,
		Count: 0,
	}, nil
}

// GetCollectionCount returns the document count in an index
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	info, err := a.GetCollectionInfo(ctx, name)
	if err != nil {
		return 0, err
	}
	return info.Count, nil
}
