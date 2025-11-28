// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"fmt"

	qdrant "github.com/qdrant/go-client/qdrant"
)

// CreateCollection creates a new collection in Qdrant with the specified vector configuration
func (c *Client) CreateCollection(ctx context.Context, name string, vectorDimensions int, similarityMetric string) error {
	// Map similarity metric to Qdrant's distance type
	distance, err := mapDistanceMetric(similarityMetric)
	if err != nil {
		return err
	}

	// Create collection request
	req := &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(vectorDimensions),
					Distance: distance,
				},
			},
		},
	}

	// Create collection
	_, err = c.collectionsClient.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", name, err)
	}

	return nil
}

// DeleteCollection deletes a collection and all its data
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	req := &qdrant.DeleteCollection{
		CollectionName: name,
	}

	_, err := c.collectionsClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", name, err)
	}

	return nil
}

// ListCollections returns a list of all collection names
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	req := &qdrant.ListCollectionsRequest{}

	resp, err := c.collectionsClient.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	collections := make([]string, len(resp.Collections))
	for i, col := range resp.Collections {
		collections[i] = col.Name
	}

	return collections, nil
}

// CollectionExists checks if a collection exists
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	req := &qdrant.CollectionExistsRequest{
		CollectionName: name,
	}

	resp, err := c.collectionsClient.CollectionExists(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check if collection %s exists: %w", name, err)
	}

	return resp.Result.Exists, nil
}

// GetCollectionInfo retrieves information about a collection
func (c *Client) GetCollectionInfo(ctx context.Context, name string) (*qdrant.GetCollectionInfoResponse, error) {
	req := &qdrant.GetCollectionInfoRequest{
		CollectionName: name,
	}

	resp, err := c.collectionsClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info for %s: %w", name, err)
	}

	return resp, nil
}

// GetCollectionCount returns the number of points (documents) in a collection
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	info, err := c.GetCollectionInfo(ctx, name)
	if err != nil {
		return 0, err
	}

	if info.Result.PointsCount == nil {
		return 0, nil
	}
	return int64(*info.Result.PointsCount), nil
}

// mapDistanceMetric converts a string similarity metric to Qdrant's Distance enum
func mapDistanceMetric(metric string) (qdrant.Distance, error) {
	switch metric {
	case "Cosine", "cosine":
		return qdrant.Distance_Cosine, nil
	case "Dot", "dot":
		return qdrant.Distance_Dot, nil
	case "Euclidean", "euclidean", "L2", "l2":
		return qdrant.Distance_Euclid, nil
	default:
		return qdrant.Distance_UnknownDistance, fmt.Errorf("unsupported similarity metric: %s (must be Cosine, Dot, or Euclidean)", metric)
	}
}
