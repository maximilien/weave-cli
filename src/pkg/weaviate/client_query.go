// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"
	"reflect"
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
	// Try nearText first, fall back to simple search if not supported
	query := fmt.Sprintf(`
		{
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

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// Try fallback query with where clause instead of nearText
		return c.queryWithFallback(ctx, collectionName, queryText, options, contentField)
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
		{
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

	// Access the Data field directly (same as existing code)
	var data map[string]interface{}
	if resultMap, ok := result.(map[string]interface{}); ok {
		data = resultMap
	} else {
		// Try to access Data field using reflection
		v := reflect.ValueOf(result)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		
		dataField := v.FieldByName("Data")
		if !dataField.IsValid() || dataField.IsNil() {
			return nil, fmt.Errorf("invalid result format: %T", result)
		}
		
		if dataMap, ok := dataField.Interface().(map[string]interface{}); ok {
			data = dataMap
		} else {
			// Try to convert JSONObject to interface{}
			if jsonObjectMap, ok := dataField.Interface().(map[string]interface{}); ok {
				data = jsonObjectMap
			} else {
				// Convert the JSONObject to a regular map
				data = convertJSONObjectToMap(dataField.Interface())
			}
		}
	}

	// Extract the Get data
	getData, ok := data["Get"].(map[string]interface{})
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

// getDataField extracts the Data field from a GraphQLResponse using reflection
func getDataField(result interface{}) map[string]interface{} {
	// Use reflection to access the Data field
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	dataField := v.FieldByName("Data")
	if !dataField.IsValid() || dataField.IsNil() {
		return nil
	}
	
	if data, ok := dataField.Interface().(map[string]interface{}); ok {
		return data
	}
	
	return nil
}

// hasGraphQLErrors checks if the GraphQL response contains errors
func hasGraphQLErrors(result interface{}) bool {
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	errorsField := v.FieldByName("Errors")
	if !errorsField.IsValid() || errorsField.IsNil() {
		return false
	}
	
	// Check if errors slice has any elements
	if errorsField.Kind() == reflect.Slice {
		return errorsField.Len() > 0
	}
	
	return false
}

// queryWithFallback performs a fallback text search using where clause
func (c *Client) queryWithFallback(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Build a simple query with where clause for text search
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					where: {
						path: ["%s"]
						operator: Like
						valueText: "*%s*"
					}
					limit: %d
				) {
					_additional {
						id
					}
					%s
					metadata
				}
			}
		}`, collectionName, contentField, strings.ReplaceAll(queryText, `"`, `\"`), options.TopK, contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute fallback query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		return nil, fmt.Errorf("fallback query also returned errors")
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fallback query results: %v", err)
	}

	// Since this is a simple text search, set all scores to 1.0
	for i := range results {
		results[i].Score = 1.0
	}

	return results, nil
}

// convertJSONObjectToMap converts a JSONObject to a regular map[string]interface{}
func convertJSONObjectToMap(jsonObj interface{}) map[string]interface{} {
	// Use reflection to convert the JSONObject
	v := reflect.ValueOf(jsonObj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	result := make(map[string]interface{})
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			value := v.MapIndex(key).Interface()
			result[keyStr] = value
		}
	}
	
	return result
}
