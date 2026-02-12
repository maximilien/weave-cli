// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/logging"
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
	TopKImages     int     `json:"top_k_images"` // Number of results from image collections (0 = use TopK)
	Distance       float64 `json:"distance"`
	SearchMetadata bool    `json:"search_metadata"`
	NoTruncate     bool    `json:"no_truncate"`
	UseBM25        bool    `json:"use_bm25"`
	JSONOutput     bool    `json:"json_output"`
	ImageQuery     string  `json:"image_query"`      // Base64 encoded image for nearImage search
	UseImageVector bool    `json:"use_image_vector"` // Use image_vector instead of text_vector
	IncludeImages  bool    `json:"include_images"`   // Include base64 image data in JSON output
	Verbose        bool    `json:"verbose"`          // Enable debug logging
}

// normalizeScore applies a non-linear transformation to spread scores across a wider range.
// This makes low-relevance results (0.5) appear much lower (0.25) while keeping
// high-relevance results (0.7+) relatively high.
//
// The transformation uses a power function: score^2 which:
// - Maps 0.5 → 0.25 (low, indicates marginal match)
// - Maps 0.6 → 0.36 (moderate match)
// - Maps 0.65 → 0.42 (good match)
// - Maps 0.7 → 0.49 (good relevance)
// - Maps 0.8 → 0.64 (strong relevance)
// - Maps 0.9 → 0.81 (very strong relevance)
// - Maps 1.0 → 1.0 (perfect match)
func normalizeScore(rawScore float64) float64 {
	if rawScore < 0 {
		return 0
	}
	if rawScore > 1 {
		return 1
	}
	// Use quadratic function to amplify differences
	return math.Pow(rawScore, 2.0)
}

// Query performs semantic search on a collection using nearText
func (c *Client) Query(ctx context.Context, collectionName, queryText string, options QueryOptions) ([]QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default top_k if not specified
	if options.TopK <= 0 {
		options.TopK = 5
	}

	logging.Debug("Starting query: collection=%s query=%s top_k=%d", collectionName, queryText, options.TopK)

	// Get the collection schema to determine the content field
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		context := map[string]interface{}{
			"collection": collectionName,
			"query":      queryText,
			"top_k":      options.TopK,
		}
		return nil, logging.WrapErrorWithContext(err, "Query", "weaviate", c.config.URL, context)
	}

	// Determine the content field name and collection type
	contentField := "content"
	hasContent := false
	hasText := false
	hasURL := false

	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
		if prop.Name == "url" {
			hasURL = true
		}
	}

	// Determine if this is an image collection by checking collection name
	// This matches collections created with image schema (dual vectors: text_vector + image_vector)
	isImageColl := isImageCollection(collectionName)

	// Build the field list for the query based on collection type
	var queryFields string
	if isImageColl {
		// For image collections, request url and metadata.filename
		// GraphQL allows accessing nested properties directly by name
		queryFields = "url\n\t\t\t\t\tmetadata {\n\t\t\t\t\t\tfilename\n\t\t\t\t\t}"
		// Include image_data field if requested
		if options.IncludeImages {
			queryFields = "url\n\t\t\t\t\timage_data\n\t\t\t\t\tmetadata {\n\t\t\t\t\t\tfilename\n\t\t\t\t\t}"
		}
		if hasURL {
			contentField = "url" // Use URL for display purposes
		}
	} else {
		// For text collections, request content/text and metadata
		if hasContent {
			contentField = "content"
			queryFields = "content\n\t\t\t\t\tmetadata"
		} else if hasText {
			contentField = "text"
			queryFields = "text\n\t\t\t\t\tmetadata"
		} else {
			queryFields = "metadata"
		}
	}

	// If BM25 flag is set, use BM25 search directly
	if options.UseBM25 {
		return c.queryWithBM25(ctx, collectionName, queryText, options, contentField)
	}

	// If image query is provided, use nearImage search with image_vector
	if options.ImageQuery != "" && isImageColl {
		return c.queryWithNearImage(ctx, collectionName, options.ImageQuery, options, queryFields)
	}

	// Build the GraphQL query for semantic search using nearText
	// Determine which named vector to use based on collection type and options
	var targetVector string
	if isImageColl {
		if options.UseImageVector {
			// Use image_vector for visual similarity search
			targetVector = "image_vector"
		} else {
			// Use text_vector for text metadata search (default)
			targetVector = "text_vector"
		}
	} else {
		// For text collections, use default vector
		targetVector = "default"
	}

	query := fmt.Sprintf(`
		{
			Get {
				%s(
					nearText: {
						concepts: ["%s"]
						targetVectors: ["%s"]
					}
					limit: %d
				) {
					_additional {
						id
						distance
						certainty
					}
					%s
				}
			}
		}`, collectionName, strings.ReplaceAll(queryText, `"`, `\"`), targetVector, options.TopK, queryFields)

	// Debug logging
	if options.Verbose {
		log.Printf("🔍 DEBUG: GraphQL Query (nearText):\n%s\n", query)
		log.Printf("🔍 DEBUG: Target Vector: %s", targetVector)
		log.Printf("🔍 DEBUG: Collection: %s", collectionName)
		log.Printf("🔍 DEBUG: Content field: %s", contentField)
		log.Printf("🔍 DEBUG: Query fields: %s", queryFields)
	}

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to execute semantic search query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// Try fallback query with hybrid search instead of nearText
		return c.queryWithFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse query results: %v", err)
	}

	return results, nil
}

// queryWithNearImage performs visual similarity search using nearImage operator
// This uses the image_vector (multi2vec-clip) to find visually similar images
func (c *Client) queryWithNearImage(ctx context.Context, collectionName, imageBase64 string, options QueryOptions, queryFields string) ([]QueryResult, error) {
	// Build GraphQL query with nearImage operator
	// nearImage requires base64 encoded image data
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					nearImage: {
						image: "%s"
						targetVectors: ["image_vector"]
					}
					limit: %d
				) {
					_additional {
						id
						distance
						certainty
					}
					%s
				}
			}
		}`, collectionName, imageBase64, options.TopK, queryFields)

	// Debug logging
	if options.Verbose {
		log.Printf("🔍 DEBUG: GraphQL Query (nearImage):\n%s\n", query)
		log.Printf("🔍 DEBUG: Target Vector: image_vector")
		log.Printf("🔍 DEBUG: Collection: %s", collectionName)
		log.Printf("🔍 DEBUG: Fields requested: %s", queryFields)
	}

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to execute nearImage search query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		return nil, fmt.Errorf("Weaviate: nearImage query returned errors")
	}

	// Parse the results (reuse existing parser)
	results, err := c.parseQueryResults(result, "url") // Use url as contentField for images
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse nearImage results: %v", err)
	}

	return results, nil
}

// queryWithBM25 performs BM25 keyword search with real similarity scores
func (c *Client) queryWithBM25(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Get schema to check available fields
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to get collection schema for BM25: %w", err)
	}

	// Check available fields
	hasContent := false
	hasText := false
	hasMetadata := false
	hasURL := false
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
		if prop.Name == "url" {
			hasURL = true
		}
	}

	// Build query fields for BM25 search
	var queryFields []string

	// For image collections, search url field only (metadata is object type, can't use in BM25 properties)
	// BM25 will still search through nested text fields automatically
	if isImageCollection(collectionName) {
		if hasURL {
			queryFields = append(queryFields, "url")
		}
	} else {
		// For text collections, use content/text fields
		if hasContent {
			queryFields = append(queryFields, "content")
		}
		if hasText {
			queryFields = append(queryFields, "text")
		}
		if options.SearchMetadata && hasMetadata {
			queryFields = append(queryFields, "metadata")
		}
	}

	if len(queryFields) == 0 {
		return nil, fmt.Errorf("Weaviate: no searchable fields found in collection")
	}

	// Escape query text for GraphQL
	queryTextEscaped := strings.ReplaceAll(queryText, `"`, `\"`)

	// Build properties list for BM25 query
	propertiesList := ""
	if len(queryFields) > 0 {
		// Quote each field name properly
		quotedFields := make([]string, len(queryFields))
		for i, field := range queryFields {
			quotedFields[i] = `"` + field + `"`
		}
		propertiesList = fmt.Sprintf(`properties: [%s]`, strings.Join(quotedFields, ", "))
	}

	// Build the GraphQL query using BM25 for real similarity scores
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					bm25: {
						query: "%s"
						%s
					}
					limit: %d
				) {
					_additional {
						id
						score
						explainScore
					}
					%s
					metadata
				}
			}
		}`, collectionName, queryTextEscaped, propertiesList, options.TopK, contentField)

	// Debug logging
	if options.Verbose {
		log.Printf("🔍 DEBUG: GraphQL Query (BM25):\n%s\n", query)
		log.Printf("🔍 DEBUG: Collection: %s", collectionName)
		log.Printf("🔍 DEBUG: BM25 properties: %v", queryFields)
	}

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to execute BM25 search query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// BM25 might not be supported, fall back to hybrid search
		return c.queryWithFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse BM25 query results: %v", err)
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
		return nil, fmt.Errorf("Weaviate: failed to get collection schema: %w", err)
	}

	// Determine the content field name - prefer content, fallback to text
	contentField := "content"
	hasContent := false
	hasText := false
	hasImage := false

	for _, prop := range schema.Properties {
		if prop.Name == "content" {
			hasContent = true
		}
		if prop.Name == "text" {
			hasText = true
		}
		if prop.Name == "image" || prop.Name == "image_data" {
			hasImage = true
		}
	}

	// Use content if available, otherwise use text
	if hasContent {
		contentField = "content"
	} else if hasText {
		contentField = "text"
	}

	// Determine target vector for multi-vector collections
	targetVector := "default"
	isImageColl := hasImage && !hasContent && !hasText
	if isImageColl {
		targetVector = "text_vector"
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
	// For Weaviate v1.25+, we need to specify the target vector name
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					nearText: {
						concepts: ["%s"]
						targetVectors: ["%s"]
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
		}`, collectionName, strings.ReplaceAll(queryText, `"`, `\"`), targetVector, options.TopK,
		func() string {
			if whereClause != "" {
				return ",\n\t\t\t" + whereClause
			}
			return ""
		}(), contentField)

	// Debug logging
	if options.Verbose {
		log.Printf("🔍 DEBUG: GraphQL Query (QueryWithFilters):\n%s\n", query)
		log.Printf("🔍 DEBUG: Target Vector: %s", targetVector)
		log.Printf("🔍 DEBUG: Collection: %s", collectionName)
		log.Printf("🔍 DEBUG: Content field: %s", contentField)
		log.Printf("🔍 DEBUG: Filters: %v", filters)
	}

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to execute semantic search query with filters: %w", err)
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse query results: %v", err)
	}

	return results, nil
}

// parseQueryResults parses the GraphQL response into QueryResult objects
func (c *Client) parseQueryResults(result interface{}, contentField string) ([]QueryResult, error) {
	if result == nil {
		return nil, fmt.Errorf("Weaviate: received nil result from GraphQL query")
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
			return nil, fmt.Errorf("Weaviate: invalid result format: %T", result)
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
		return nil, fmt.Errorf("Weaviate: invalid response format: missing Get data")
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
		return nil, fmt.Errorf("Weaviate: no results found in response")
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

		// Extract similarity score (handle distance, certainty, and score fields)
		var rawScore float64
		if certainty, exists := additional["certainty"]; exists {
			// nearText can provide certainty (0.0 to 1.0, higher is better)
			if cert, ok := certainty.(float64); ok {
				rawScore = cert
			}
		} else if distance, exists := additional["distance"]; exists {
			// nearText can also provide distance (lower is better, convert to score)
			if dist, ok := distance.(float64); ok {
				// Convert distance to similarity score (assuming cosine distance 0-2 range)
				rawScore = 1.0 - (dist / 2.0)
				if rawScore < 0 {
					rawScore = 0
				}
			}
		} else if scoreVal, exists := additional["score"]; exists {
			// BM25/hybrid provides score directly
			if s, ok := scoreVal.(float64); ok {
				rawScore = s
			}
		} else {
			// No score available, default to 0.5
			rawScore = 0.5
		}

		// Apply score normalization to spread scores across wider range
		score := normalizeScore(rawScore)

		// Extract content
		content, _ := resultItem[contentField].(string)

		// Extract metadata
		metadata, _ := resultItem["metadata"].(map[string]interface{})

		// Extract image_data if present and add to metadata
		if imageData, ok := resultItem["image_data"].(string); ok && imageData != "" {
			if metadata == nil {
				metadata = make(map[string]interface{})
			}
			metadata["image_base64"] = imageData
		}

		queryResults = append(queryResults, QueryResult{
			ID:       id,
			Content:  content,
			Metadata: metadata,
			Score:    score,
		})
	}

	return queryResults, nil
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
		return nil, fmt.Errorf("Weaviate: failed to get collection schema for fallback: %w", err)
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
		return nil, fmt.Errorf("Weaviate: no searchable fields found in collection")
	}

	// Escape query text for GraphQL
	queryTextEscaped := strings.ReplaceAll(queryText, `"`, `\"`)

	// Determine target vector for collections with multiple vectors
	targetVector := "default"
	hasImage := false
	for _, prop := range schema.Properties {
		if prop.Name == "image" || prop.Name == "image_data" {
			hasImage = true
			break
		}
	}
	// For image collections with dual vectors, use text_vector
	if hasImage && !hasContent && !hasText {
		targetVector = "text_vector"
	}

	// Build the GraphQL query using hybrid search for real similarity scores
	// Hybrid search combines vector search with keyword search
	query := fmt.Sprintf(`
		{
			Get {
				%s(
					hybrid: {
						query: "%s"
						alpha: 0.75
						targetVectors: ["%s"]
					}
					limit: %d
				) {
					_additional {
						id
						score
						explainScore
					}
					%s
					metadata
				}
			}
		}`, collectionName, queryTextEscaped, targetVector, options.TopK, contentField)

	// Debug logging
	if options.Verbose {
		log.Printf("🔍 DEBUG: GraphQL Query (Hybrid Fallback):\n%s\n", query)
		log.Printf("🔍 DEBUG: Target Vector: %s", targetVector)
		log.Printf("🔍 DEBUG: Collection: %s", collectionName)
		log.Printf("🔍 DEBUG: Content field: %s", contentField)
	}

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to execute hybrid fallback query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		// Hybrid search might not be supported, fall back to simple where clause
		return c.queryWithSimpleFallback(ctx, collectionName, queryText, options, contentField)
	}

	// Parse results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse hybrid fallback query results: %v", err)
	}

	// Hybrid search provides real similarity scores, so we don't need to modify them
	return results, nil
}

// queryWithSimpleFallback performs a simple text search using where clause as final fallback
func (c *Client) queryWithSimpleFallback(ctx context.Context, collectionName, queryText string, options QueryOptions, contentField string) ([]QueryResult, error) {
	// Get schema to check available fields
	schema, err := c.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to get collection schema for simple fallback: %w", err)
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
		return nil, fmt.Errorf("Weaviate: no searchable fields found in collection")
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
		return nil, fmt.Errorf("Weaviate: failed to execute simple fallback query: %w", err)
	}

	// Check for GraphQL errors
	if hasGraphQLErrors(result) {
		return nil, fmt.Errorf("Weaviate: simple fallback query also returned errors")
	}

	// Parse the results
	results, err := c.parseQueryResults(result, contentField)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to parse simple fallback query results: %v", err)
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
