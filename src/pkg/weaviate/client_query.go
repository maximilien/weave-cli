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
	TopK           int     `json:"top_k"`
	Distance       float64 `json:"distance"`
	SearchMetadata bool    `json:"search_metadata"`
	NoTruncate     bool    `json:"no_truncate"`
	UseBM25        bool    `json:"use_bm25"`
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

	// Determine the content field name - prefer content, fallback to text
	contentField := "content"
	hasContent := false
	hasText := false

	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
	}

	// Use content if available, otherwise use text
	if hasContent {
		contentField = "content"
	} else if hasText {
		contentField = "text"
	}

	// If BM25 flag is set, use BM25 search directly
	if options.UseBM25 {
		return c.queryWithBM25(ctx, collectionName, queryText, options, contentField)
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
		// Try fallback query with hybrid search instead of nearText
		return c.queryWithFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query results: %v", err)
	}

	return results, nil
}

// queryWithBM25 performs BM25 keyword search with real similarity scores
func (c *Client) queryWithBM25(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Get schema to check available fields
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema for BM25: %w", err)
	}

	// Check available fields
	hasContent := false
	hasText := false
	hasMetadata := false
	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
		if prop.Name == "metadata" {
			hasMetadata = true
		}
	}

	// Build query fields for BM25 search
	var queryFields []string
	if hasContent {
		queryFields = append(queryFields, "content")
	}
	if hasText {
		queryFields = append(queryFields, "text")
	}
	if options.SearchMetadata && hasMetadata {
		queryFields = append(queryFields, "metadata")
	}

	if len(queryFields) == 0 {
		return nil, fmt.Errorf("no searchable fields found in collection")
	}

	// Escape query text for GraphQL
	queryTextEscaped := strings.ReplaceAll(queryText, `"`, `\"`)

	// Build the GraphQL query using BM25 for real similarity scores
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					bm25: {
						query: "%s"
						properties: [%s]
						limit: %d
					}
				) {
					_additional {
						id
						score
					}
					%s
					metadata
				}
			}
		}`, collectionName, queryTextEscaped, strings.Join(queryFields, ","), options.TopK, contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute BM25 search query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// BM25 might not be supported, fall back to hybrid search
		return c.queryWithFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BM25 query results: %v", err)
	}

	// BM25 provides real similarity scores, so we don't need to modify them
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

	// Determine the content field name - prefer content, fallback to text
	contentField := "content"
	hasContent := false
	hasText := false

	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
	}

	// Use content if available, otherwise use text
	if hasContent {
		contentField = "content"
	} else if hasText {
		contentField = "text"
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

		// Extract similarity score (handle both distance and score fields)
		var score float64
		if distance, exists := additional["distance"]; exists {
			// nearText provides distance, convert to similarity score
			if dist, ok := distance.(float64); ok {
				score = 1.0 - dist
			}
		} else if scoreVal, exists := additional["score"]; exists {
			// BM25/hybrid provides score directly
			if s, ok := scoreVal.(float64); ok {
				score = s
			}
		}

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
	if result == nil {
		return true
	}

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

// queryWithFallback performs a fallback search using hybrid search for real similarity scores
func (c *Client) queryWithFallback(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Get schema to check available fields
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema for fallback: %w", err)
	}

	// Check available fields
	hasContent := false
	hasText := false
	hasMetadata := false
	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
		if prop.Name == "metadata" {
			hasMetadata = true
		}
	}

	// Build query fields for hybrid search
	var queryFields []string
	if hasContent {
		queryFields = append(queryFields, "content")
	}
	if hasText {
		queryFields = append(queryFields, "text")
	}
	if options.SearchMetadata && hasMetadata {
		queryFields = append(queryFields, "metadata")
	}

	if len(queryFields) == 0 {
		return nil, fmt.Errorf("no searchable fields found in collection")
	}

	// Escape query text for GraphQL
	queryTextEscaped := strings.ReplaceAll(queryText, `"`, `\"`)

	// Build the GraphQL query using hybrid search for real similarity scores
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					hybrid: {
						query: "%s"
						properties: [%s]
						limit: %d
					}
				) {
					_additional {
						id
						score
					}
					%s
					metadata
				}
			}
		}`, collectionName, queryTextEscaped, strings.Join(queryFields, ","), options.TopK, contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute hybrid fallback query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// Hybrid search might not be supported, fall back to simple where clause
		return c.queryWithSimpleFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hybrid fallback query results: %v", err)
	}

	// Hybrid search provides real similarity scores, so we don't need to modify them
	return results, nil
}

// queryWithSimpleFallback performs a simple text search using where clause as final fallback
func (c *Client) queryWithSimpleFallback(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Get schema to check available fields
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema for simple fallback: %w", err)
	}

	// Check available fields
	hasContent := false
	hasText := false
	hasMetadata := false
	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
		if prop.Name == "metadata" {
			hasMetadata = true
		}
	}

	// Build query based on available fields and search options
	var operands []string
	queryTextEscaped := strings.ReplaceAll(queryText, `"`, `\"`)

	// Always search content/text fields
	if hasContent && hasText {
		operands = append(operands, fmt.Sprintf(`{
							path: ["content"]
							operator: Equal
							valueText: "%s"
						}`, queryTextEscaped))
		operands = append(operands, fmt.Sprintf(`{
							path: ["text"]
							operator: Equal
							valueText: "%s"
						}`, queryTextEscaped))
	} else if hasContent {
		operands = append(operands, fmt.Sprintf(`{
							path: ["content"]
							operator: Equal
							valueText: "%s"
						}`, queryTextEscaped))
	} else if hasText {
		operands = append(operands, fmt.Sprintf(`{
							path: ["text"]
							operator: Equal
							valueText: "%s"
						}`, queryTextEscaped))
	}

	// Add metadata search if enabled and available
	if options.SearchMetadata && hasMetadata {
		operands = append(operands, fmt.Sprintf(`{
							path: ["metadata"]
							operator: Equal
							valueText: "%s"
						}`, queryTextEscaped))
	}

	// Build the where clause
	var whereClause string
	if len(operands) == 1 {
		// Single field search
		whereClause = fmt.Sprintf(`
					where: %s`, operands[0])
	} else if len(operands) > 1 {
		// Multiple fields search with OR
		operandsStr := strings.Join(operands, ",\n\t\t\t\t\t")
		whereClause = fmt.Sprintf(`
					where: {
						operator: Or
						operands: [
							%s
						]
					}`, operandsStr)
	} else {
		return nil, fmt.Errorf("no searchable fields found in collection")
	}

	// Build the complete query
	query := fmt.Sprintf(`
		{
			Get {
				%s(%s
					limit: %d
				) {
					_additional {
						id
					}
					%s
					metadata
				}
			}
		}`, collectionName, whereClause, options.TopK, contentField)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute simple fallback query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		return nil, fmt.Errorf("simple fallback query also returned errors")
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("failed to parse simple fallback query results: %v", err)
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
