// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs KNN vector search via FT.SEARCH.
func (c *Client) SearchSemantic(ctx context.Context, collection string, vector []float32, topK int) ([]SearchResult, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeQuery)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	blob := float32sToBytes(vector)
	idx := indexName(collection)

	// FT.SEARCH idx "*=>[KNN $K @vector $BLOB]"
	//   PARAMS 4 K topK BLOB blob
	//   SORTBY __vector_score ASC
	//   RETURN 3 content metadata __vector_score
	//   DIALECT 2
	res, err := c.rdb.Do(ctx,
		"FT.SEARCH", idx,
		"*=>[KNN $K @"+defaultVectorField+" $BLOB]",
		"PARAMS", "4",
		"K", topK,
		"BLOB", blob,
		"SORTBY", "__"+defaultVectorField+"_score", "ASC",
		"RETURN", "4", defaultContentField, "metadata", "doc_id",
		"__"+defaultVectorField+"_score",
		"LIMIT", "0", topK,
		"DIALECT", "2",
	).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis: semantic search failed: %w", err)
	}

	return parseSearchResultsWithScore(res, collection), nil
}

// SearchByMetadata searches documents by metadata field values.
func (c *Client) SearchByMetadata(ctx context.Context, collection string, metadata map[string]interface{}, limit int) ([]SearchResult, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeQuery)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if limit <= 0 {
		limit = 100
	}

	// Build a full-text query from metadata values
	query := buildMetadataQuery(metadata)
	idx := indexName(collection)

	res, err := c.rdb.Do(ctx,
		"FT.SEARCH", idx, query,
		"LIMIT", "0", limit,
		"RETURN", "3", defaultContentField, "metadata", "doc_id",
	).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis: metadata search failed: %w", err)
	}

	docs := parseSearchResults(res, collection)
	results := make([]SearchResult, len(docs))
	for i, doc := range docs {
		results[i] = SearchResult{Document: doc, Score: 1.0}
	}
	return results, nil
}

// SearchFullText performs full-text search on the content field.
func (c *Client) SearchFullText(ctx context.Context, collection, query string, limit int) ([]SearchResult, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeQuery)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if limit <= 0 {
		limit = 10
	}

	idx := indexName(collection)
	// Escape special characters in query for RediSearch
	escapedQuery := escapeRedisQuery(query)

	res, err := c.rdb.Do(ctx,
		"FT.SEARCH", idx, escapedQuery,
		"LIMIT", "0", limit,
		"RETURN", "3", defaultContentField, "metadata", "doc_id",
	).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis: full-text search failed: %w", err)
	}

	return parseSearchResultsWithScore(res, collection), nil
}

// SearchResult holds a document and its search score.
type SearchResult struct {
	Document *vectordb.Document
	Score    float64
}

// --- helpers ---

// parseSearchResultsWithScore extracts docs + scores from FT.SEARCH.
func parseSearchResultsWithScore(raw interface{}, collection string) []SearchResult {
	slice, ok := raw.([]interface{})
	if !ok || len(slice) < 1 {
		return nil
	}

	prefix := docKeyPrefix(collection)
	var results []SearchResult

	for i := 1; i+1 < len(slice); i += 2 {
		keyRaw, _ := slice[i].(string)
		fieldsRaw, _ := slice[i+1].([]interface{})

		docID := keyRaw
		if len(keyRaw) > len(prefix) {
			docID = keyRaw[len(prefix):]
		}

		fields := flatSliceToMap(fieldsRaw)
		doc := hashToDocument(docID, fields)

		score := 0.0
		if scoreStr, ok := fields["__"+defaultVectorField+"_score"]; ok {
			score, _ = strconv.ParseFloat(scoreStr, 64)
			// Redis returns distance, convert to similarity (1 - distance for cosine)
			if score >= 0 {
				score = 1.0 - score
			}
		}

		results = append(results, SearchResult{Document: doc, Score: score})
	}

	return results
}

// buildMetadataQuery builds a RediSearch query from metadata k/v pairs.
// Searches metadata JSON field for matching values.
func buildMetadataQuery(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return "*"
	}

	var parts []string
	for _, v := range metadata {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return escapeRedisQuery(strings.Join(parts, " "))
}

// escapeRedisQuery escapes RediSearch special characters.
func escapeRedisQuery(q string) string {
	// RediSearch special chars that need escaping
	special := []string{
		",", ".", "<", ">", "{", "}", "[", "]",
		"\"", "'", ":", ";", "!", "@", "#", "$",
		"%", "^", "&", "*", "(", ")", "-", "+",
		"=", "~",
	}
	for _, ch := range special {
		q = strings.ReplaceAll(q, ch, "\\"+ch)
	}
	return q
}
