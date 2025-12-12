// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/densevectorsimilarity"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Helper function to convert similarity metric string to enum
func getSimilarityMetric(metric string) *densevectorsimilarity.DenseVectorSimilarity {
	var result densevectorsimilarity.DenseVectorSimilarity
	switch metric {
	case "cosine":
		result = densevectorsimilarity.Cosine
	case "dot_product":
		result = densevectorsimilarity.Dotproduct
	case "l2_norm":
		result = densevectorsimilarity.L2norm
	default:
		result = densevectorsimilarity.Cosine // default to cosine
	}
	return &result
}

// CreateCollection creates a new Elasticsearch index with vector field mappings
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Build index mappings with dense_vector field for embeddings
	indexTrue := true
	dims := a.config.VectorDimensions
	similarity := getSimilarityMetric(a.config.SimilarityMetric)

	// Create dense vector property
	vectorProp := types.NewDenseVectorProperty()
	vectorProp.Dims = &dims
	vectorProp.Index = &indexTrue
	vectorProp.Similarity = similarity

	// Create properties map
	properties := map[string]types.Property{
		// Text fields for content
		"text":    types.NewTextProperty(),
		"content": types.NewTextProperty(),

		// Dense vector field for embeddings
		"vector_field": vectorProp,

		// Metadata as object (nested fields)
		"metadata": types.NewObjectProperty(),

		// Additional fields
		"image":       types.NewKeywordProperty(),
		"image_data":  types.NewKeywordProperty(),
		"url":         types.NewKeywordProperty(),
		"document_id": types.NewKeywordProperty(),

		// Timestamps
		"created_at": types.NewDateProperty(),
		"updated_at": types.NewDateProperty(),
	}

	mappings := &types.TypeMapping{
		Properties: properties,
	}

	// Create index with mappings
	req := &create.Request{
		Mappings: mappings,
	}

	res, err := a.client.Indices.Create(name).Request(req).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", name, err)
	}

	if !res.Acknowledged {
		return fmt.Errorf("index creation not acknowledged for %s", name)
	}

	return nil
}

// DeleteCollection deletes an Elasticsearch index
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	res, err := a.client.Indices.Delete(name).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", name, err)
	}

	if !res.Acknowledged {
		return fmt.Errorf("index deletion not acknowledged for %s", name)
	}

	return nil
}

// ListCollections returns a list of all Elasticsearch indices
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Get all indices (excluding system indices that start with .)
	res, err := a.client.Indices.Get("*").Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}

	var collections []vectordb.CollectionInfo
	for indexName := range res {
		// Skip system indices (start with .)
		if strings.HasPrefix(indexName, ".") {
			continue
		}

		// Get document count for this index
		count, err := a.GetCollectionCount(ctx, indexName)
		if err != nil {
			// Log warning but continue
			count = 0
		}

		collections = append(collections, vectordb.CollectionInfo{
			Name:  indexName,
			Count: count,
		})
	}

	return collections, nil
}

// CollectionExists checks if an Elasticsearch index exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	res, err := a.client.Indices.Exists(name).IsSuccess(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if index %s exists: %w", name, err)
	}

	return res, nil
}

// GetCollectionCount returns the number of documents in an Elasticsearch index
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	res, err := a.client.Count().Index(name).Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get count for index %s: %w", name, err)
	}

	return res.Count, nil
}
