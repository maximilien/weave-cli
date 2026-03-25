// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a RediSearch index that acts as a collection.
func (c *Client) CreateCollection(ctx context.Context, name string, dims int, metric string) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeCollection)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if dims == 0 {
		dims = c.config.VectorDimensions
	}
	if metric == "" {
		metric = c.config.SimilarityMetric
	}

	idx := indexName(name)
	prefix := docKeyPrefix(name)

	// FT.CREATE idx ON HASH PREFIX 1 prefix SCHEMA
	//   content TEXT
	//   vector VECTOR HNSW 6 TYPE FLOAT32 DIM dims DISTANCE_METRIC metric
	err := c.rdb.Do(ctx,
		"FT.CREATE", idx,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		defaultContentField, "TEXT",
		"doc_id", "TAG",
		"metadata", "TEXT",
		defaultVectorField, "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", dims,
		"DISTANCE_METRIC", metric,
	).Err()
	if err != nil {
		return fmt.Errorf("Redis: failed to create index %q: %w", idx, err)
	}
	return nil
}

// DeleteCollection drops the RediSearch index and all associated documents.
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeCollection)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// FT.DROPINDEX idx DD — DD flag deletes the underlying hash keys too.
	err := c.rdb.Do(ctx, "FT.DROPINDEX", indexName(name), "DD").Err()
	if err != nil {
		return fmt.Errorf("Redis: failed to drop index %q: %w", name, err)
	}
	return nil
}

// ListCollections returns all weave-managed RediSearch indexes.
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeCollection)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := c.rdb.Do(ctx, "FT._LIST").StringSlice()
	if err != nil {
		return nil, fmt.Errorf("Redis: failed to list indexes: %w", err)
	}

	// Filter to only weave-managed indexes (prefix "weave:")
	var names []string
	for _, idx := range res {
		if strings.HasPrefix(idx, keyPrefix+":") {
			names = append(names, strings.TrimPrefix(idx, keyPrefix+":"))
		}
	}
	return names, nil
}

// CollectionExists checks if a RediSearch index exists for the collection.
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeCollection)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := c.rdb.Do(ctx, "FT.INFO", indexName(name)).Err()
	if err != nil {
		if strings.Contains(err.Error(), "Unknown index name") ||
			strings.Contains(err.Error(), "Unknown Index name") {
			return false, nil
		}
		return false, fmt.Errorf("Redis: failed to check index %q: %w", name, err)
	}
	return true, nil
}

// GetCollectionCount returns the number of documents in a collection.
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeCollection)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// FT.INFO returns a flat list of key-value pairs.
	res, err := c.rdb.Do(ctx, "FT.INFO", indexName(name)).Result()
	if err != nil {
		return 0, fmt.Errorf("Redis: failed to get info for %q: %w", name, err)
	}

	return extractNumDocs(res), nil
}

// extractNumDocs pulls "num_docs" from the FT.INFO response.
func extractNumDocs(raw interface{}) int64 {
	slice, ok := raw.([]interface{})
	if !ok {
		return 0
	}
	for i := 0; i < len(slice)-1; i++ {
		key, ok := slice[i].(string)
		if ok && key == "num_docs" {
			switch v := slice[i+1].(type) {
			case int64:
				return v
			case string:
				var n int64
				fmt.Sscanf(v, "%d", &n)
				return n
			}
		}
	}
	return 0
}
