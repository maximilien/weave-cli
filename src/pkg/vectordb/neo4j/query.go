// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"
)

// QueryResult represents a vector search result
type QueryResult struct {
	Document *Document
	Score    float64
}

// VectorSearch performs vector similarity search
func (c *Client) VectorSearch(ctx context.Context, collectionName string, vector []float32, limit int) ([]*QueryResult, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	indexName := fmt.Sprintf("%s_vector_idx", collectionName)

	// Use db.index.vector.queryNodes() for vector similarity search
	query := `
		CALL db.index.vector.queryNodes($indexName, $limit, $vector)
		YIELD node, score
		RETURN node.id as id,
		       node.content as content,
		       node.embedding as embedding,
		       properties(node) as props,
		       score
		ORDER BY score DESC
	`

	params := map[string]interface{}{
		"indexName": indexName,
		"limit":     limit,
		"vector":    vector,
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector search: %w", err)
	}

	var results []*QueryResult
	for _, record := range result.Records {
		doc := &Document{
			Metadata: make(map[string]interface{}),
		}

		if val, ok := record.Get("id"); ok {
			if id, ok := val.(string); ok {
				doc.ID = id
			}
		}
		if val, ok := record.Get("content"); ok {
			if content, ok := val.(string); ok {
				doc.Content = content
			}
		}
		if val, ok := record.Get("embedding"); ok {
			if vec, ok := val.([]interface{}); ok {
				doc.Vector = make([]float32, len(vec))
				for i, v := range vec {
					if f, ok := v.(float64); ok {
						doc.Vector[i] = float32(f)
					}
				}
			}
		}
		if val, ok := record.Get("props"); ok {
			if props, ok := val.(map[string]interface{}); ok {
				for k, v := range props {
					if k != "id" && k != "content" && k != "embedding" {
						doc.Metadata[k] = v
					}
				}
			}
		}

		var score float64
		if val, ok := record.Get("score"); ok {
			if s, ok := val.(float64); ok {
				score = s
			}
		}

		results = append(results, &QueryResult{
			Document: doc,
			Score:    score,
		})
	}

	return results, nil
}

// VectorSearchWithFilter performs vector similarity search with metadata filtering
func (c *Client) VectorSearchWithFilter(ctx context.Context, collectionName string, vector []float32, limit int, filter map[string]interface{}) ([]*QueryResult, error) {
	if limit <= 0 {
		limit = 10
	}

	indexName := fmt.Sprintf("%s_vector_idx", collectionName)

	// Build WHERE clause from filter
	whereClause := ""
	if len(filter) > 0 {
		whereClause = "WHERE "
		conditions := []string{}
		for key := range filter {
			conditions = append(conditions, fmt.Sprintf("node.%s = $filter_%s", key, key))
		}
		whereClause += conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	query := fmt.Sprintf(`
		CALL db.index.vector.queryNodes($indexName, $limit, $vector)
		YIELD node, score
		%s
		RETURN node.id as id,
		       node.content as content,
		       node.embedding as embedding,
		       properties(node) as props,
		       score
		ORDER BY score DESC
	`, whereClause)

	params := map[string]interface{}{
		"indexName": indexName,
		"limit":     limit,
		"vector":    vector,
	}

	// Add filter parameters
	for key, value := range filter {
		params[fmt.Sprintf("filter_%s", key)] = value
	}

	result, err := c.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to perform filtered vector search: %w", err)
	}

	var results []*QueryResult
	for _, record := range result.Records {
		doc := &Document{
			Metadata: make(map[string]interface{}),
		}

		if val, ok := record.Get("id"); ok {
			if id, ok := val.(string); ok {
				doc.ID = id
			}
		}
		if val, ok := record.Get("content"); ok {
			if content, ok := val.(string); ok {
				doc.Content = content
			}
		}
		if val, ok := record.Get("embedding"); ok {
			if vec, ok := val.([]interface{}); ok {
				doc.Vector = make([]float32, len(vec))
				for i, v := range vec {
					if f, ok := v.(float64); ok {
						doc.Vector[i] = float32(f)
					}
				}
			}
		}
		if val, ok := record.Get("props"); ok {
			if props, ok := val.(map[string]interface{}); ok {
				for k, v := range props {
					if k != "id" && k != "content" && k != "embedding" {
						doc.Metadata[k] = v
					}
				}
			}
		}

		var score float64
		if val, ok := record.Get("score"); ok {
			if s, ok := val.(float64); ok {
				score = s
			}
		}

		results = append(results, &QueryResult{
			Document: doc,
			Score:    score,
		})
	}

	return results, nil
}
