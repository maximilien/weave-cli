// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// QueryResult represents the result of a semantic search query
type QueryResult struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Score    float64                `json:"score"`
}

// QueryOptions holds options for semantic search queries
type QueryOptions struct {
	TopK     int     `json:"top_k"`
	Distance float64 `json:"distance"`
}

// Query performs semantic search on a collection using nearText
func (c *Client) Query(ctx context.Context, collectionName, queryText string, options QueryOptions) ([]QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default top_k if not specified
	if options.TopK <= 0 {
		options.TopK = 5
	}

	// Get the collection schema to determine the content field
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}

	// Determine the content field name
	contentField := "content"
	for _, prop := range schema.Properties {
		if prop.Name == "content" || prop.Name == "text" {
			contentField = prop.Name
			break
		}
	}

	// Build the GraphQL query for semantic search
	query := fmt.Sprintf(`
		query {
			Get {
				%s(
					nearText: {
						concepts: ["%s"]
						limit: %d
					}
				) {
					_additional {
						id
						distance
					}
					%s
					metadata
				}
			}
		}`, collectionName, strings.ReplaceAll(queryText, `"`, `\"`), options.TopK, contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute semantic search query: %w", err)
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query results: %v", err)
	}

	return results, nil
}

// QueryWithFilters performs semantic search with additional metadata filters
func (c *Client) QueryWithFilters(ctx context.Context, collectionName, queryText string, options QueryOptions, filters map[string]interface{}) ([]QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default top_k if not specified
	if options.TopK <= 0 {
		options.TopK = 5
	}

	// Get the collection schema to determine the content field
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}

	// Determine the content field name
	contentField := "content"
	for _, prop := range schema.Properties {
		if prop.Name == "content" || prop.Name == "text" {
			contentField = prop.Name
			break
		}
	}

	// Build where clause for filters
	whereClause := ""
	if len(filters) > 0 {
		whereClause = "where: {\n"
		for key, value := range filters {
			whereClause += fmt.Sprintf("\t\t\t\t%s: {\n\t\t\t\t\tequal: \"%v\"\n\t\t\t\t}\n", key, value)
		}
		whereClause += "\t\t\t}"
	}

	// Build the GraphQL query for semantic search with filters
	query := fmt.Sprintf(`
		query {
			Get {
				%s(
					nearText: {
						concepts: ["%s"]
						limit: %d
					}%s
				) {
					_additional {
						id
						distance
					}
					%s
					metadata
				}
			}
		}`, collectionName, strings.ReplaceAll(queryText, `"`, `\"`), options.TopK,
		func() string {
			if whereClause != "" {
				return ",\n\t\t\t" + whereClause
			}
			return ""
		}(), contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute semantic search query with filters: %w", err)
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query results: %v", err)
	}

	return results, nil
}

// parseQueryResults parses the GraphQL response into QueryResult objects
func (c *Client) parseQueryResults(result interface{}, contentField string) ([]QueryResult, error) {
	if result == nil {
		return nil, fmt.Errorf("received nil result from GraphQL query")
	}

	// Cast result to the expected type
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	// Extract the Get data
	getData, ok := resultMap["Get"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing Get data")
	}

	// Get the collection results (the key is the collection name)
	var collectionResults []interface{}
	for _, results := range getData {
		if resultsArray, ok := results.([]interface{}); ok {
			collectionResults = resultsArray
			break
		}
	}

	if collectionResults == nil {
		return nil, fmt.Errorf("no results found in response")
	}

	// Parse each result
	var queryResults []QueryResult
	for _, item := range collectionResults {
		resultItem, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract additional information (id, distance)
		additional, ok := resultItem["_additional"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract ID
		id, _ := additional["id"].(string)

		// Extract distance (similarity score)
		distance, _ := additional["distance"].(float64)
		score := 1.0 - distance // Convert distance to similarity score

		// Extract content
		content, _ := resultItem[contentField].(string)

		// Extract metadata
		metadata, _ := resultItem["metadata"].(map[string]interface{})

		queryResults = append(queryResults, QueryResult{
			ID:       id,
			Content:  content,
			Metadata: metadata,
			Score:    score,
		})
	}

	return queryResults, nil
}
