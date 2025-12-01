// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"
)

// CreateCollection creates a new vector index for a node label
// In Neo4j, a "collection" maps to a node label with a vector index
func (c *Client) CreateCollection(ctx context.Context, name string, dimensions int, similarity string) error {
	// Use config defaults if not specified
	if dimensions == 0 {
		dimensions = c.config.VectorDimensions
	}
	if similarity == "" {
		similarity = c.config.SimilarityMetric
	}

	// Build the vector index name
	indexName := fmt.Sprintf("%s_vector_idx", name)

	// Create the vector index using Cypher
	// Vector indexes require Neo4j 5.11+
	query := fmt.Sprintf(`
		CREATE VECTOR INDEX %s IF NOT EXISTS
		FOR (n:%s)
		ON (n.embedding)
		OPTIONS {
			indexConfig: {
				` + "`vector.dimensions`" + `: $dimensions,
				` + "`vector.similarity_function`" + `: $similarity
			}
		}
	`, indexName, name)

	params := map[string]interface{}{
		"dimensions": dimensions,
		"similarity": similarity,
	}

	_, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create vector index for collection '%s': %w", name, err)
	}

	return nil
}

// ListCollections returns all vector indexes (collections)
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	query := "SHOW INDEXES YIELD name, type WHERE type = 'VECTOR' RETURN name"

	result, err := c.executeQuery(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector indexes: %w", err)
	}

	var collections []string
	for _, record := range result.Records {
		name, ok := record.Get("name")
		if !ok {
			continue
		}
		indexName, ok := name.(string)
		if !ok {
			continue
		}
		// Extract collection name from index name (remove _vector_idx suffix)
		if len(indexName) > 11 && indexName[len(indexName)-11:] == "_vector_idx" {
			collectionName := indexName[:len(indexName)-11]
			collections = append(collections, collectionName)
		}
	}

	return collections, nil
}

// CollectionExists checks if a vector index exists for the given collection name
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	indexName := fmt.Sprintf("%s_vector_idx", name)
	query := "SHOW INDEXES YIELD name WHERE name = $indexName RETURN name"

	params := map[string]interface{}{
		"indexName": indexName,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return false, fmt.Errorf("failed to check if collection exists: %w", err)
	}

	return len(result.Records) > 0, nil
}

// DeleteCollection deletes a vector index and optionally all nodes with the label
func (c *Client) DeleteCollection(ctx context.Context, name string, deleteNodes bool) error {
	indexName := fmt.Sprintf("%s_vector_idx", name)

	// Delete the vector index
	query := fmt.Sprintf("DROP INDEX %s IF EXISTS", indexName)
	_, err := c.executeQuery(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to delete vector index: %w", err)
	}

	// Optionally delete all nodes with this label
	if deleteNodes {
		deleteQuery := fmt.Sprintf("MATCH (n:%s) DETACH DELETE n", name)
		_, err = c.executeQuery(ctx, deleteQuery, nil)
		if err != nil {
			return fmt.Errorf("failed to delete nodes: %w", err)
		}
	}

	return nil
}

// GetCollectionCount returns the number of nodes with the given label
func (c *Client) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	query := fmt.Sprintf("MATCH (n:%s) RETURN count(n) as count", name)

	result, err := c.executeQuery(ctx, query, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection count: %w", err)
	}

	if len(result.Records) == 0 {
		return 0, nil
	}

	count, ok := result.Records[0].Get("count")
	if !ok {
		return 0, nil
	}

	countInt, ok := count.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected count type: %T", count)
	}

	return countInt, nil
}

// GetCollectionInfo returns metadata about a vector index
func (c *Client) GetCollectionInfo(ctx context.Context, name string) (map[string]interface{}, error) {
	indexName := fmt.Sprintf("%s_vector_idx", name)
	query := `
		SHOW INDEXES
		YIELD name, type, entityType, labelsOrTypes, properties, options
		WHERE name = $indexName
		RETURN name, type, entityType, labelsOrTypes, properties, options
	`

	params := map[string]interface{}{
		"indexName": indexName,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("collection not found: %s", name)
	}

	record := result.Records[0]
	info := make(map[string]interface{})

	if val, ok := record.Get("name"); ok {
		info["name"] = val
	}
	if val, ok := record.Get("type"); ok {
		info["type"] = val
	}
	if val, ok := record.Get("entityType"); ok {
		info["entityType"] = val
	}
	if val, ok := record.Get("labelsOrTypes"); ok {
		info["labelsOrTypes"] = val
	}
	if val, ok := record.Get("properties"); ok {
		info["properties"] = val
	}
	if val, ok := record.Get("options"); ok {
		info["options"] = val
	}

	return info, nil
}
