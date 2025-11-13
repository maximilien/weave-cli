// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs semantic search using vector embeddings
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// If LLM client is available, use vector similarity search
	if a.llmClient != nil {
		// Generate embedding for the query
		queryEmbedding, err := a.llmClient.GenerateEmbedding(ctx, query, "")
		if err != nil {
			// Fall back to content search if embedding generation fails
			fmt.Fprintf(os.Stderr, "Warning: Failed to generate query embedding: %v. Falling back to content search.\n", err)
			return a.searchByContent(ctx, collectionName, query, options)
		}

		// Use pgvector similarity search
		return a.searchByVectorSimilarity(ctx, collectionName, queryEmbedding, options)
	}

	// Fall back to simple content search if no LLM client
	return a.searchByContent(ctx, collectionName, query, options)
}

// SearchBM25 performs keyword-based search using BM25
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Use PostgreSQL full-text search as a BM25 alternative
	return a.searchByFullText(ctx, collectionName, query, options)
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// For hybrid search, combine semantic and BM25 results
	// This is a simplified implementation
	semanticResults, err := a.SearchSemantic(ctx, collectionName, query, options)
	if err != nil {
		return nil, err
	}

	bm25Results, err := a.SearchBM25(ctx, collectionName, query, options)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate results
	return a.mergeSearchResults(semanticResults, bm25Results, options), nil
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	if len(metadata) == 0 {
		return []*vectordb.QueryResult{}, nil
	}

	tableName := a.getTableName(collectionName)

	// Build WHERE conditions for metadata matching
	var conditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range metadata {
		conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, argIndex))
		args = append(args, fmt.Sprintf("%v", value))
		argIndex++
	}

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	query := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata
		FROM %s
		WHERE %s
		ORDER BY id
		LIMIT $%d
	`, tableName, strings.Join(conditions, " AND "), argIndex)

	args = append(args, limit)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, a.wrapError(err, "search by metadata")
	}
	defer rows.Close()

	var results []*vectordb.QueryResult
	for rows.Next() {
		var id, content, text, image, imageData, url, metadataJSON string
		if err := rows.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON); err != nil {
			continue
		}

		docMetadata, err := a.convertJSONToMetadata(metadataJSON)
		if err != nil {
			continue
		}

		doc := &vectordb.Document{
			ID:        id,
			Content:   content,
			Text:      text,
			Image:     image,
			ImageData: imageData,
			URL:       url,
			Metadata:  docMetadata,
		}

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    1.0, // Metadata matches get a perfect score
		})
	}

	return results, nil
}

// searchByContent performs a simple content-based search
func (a *Adapter) searchByContent(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	tableName := a.getTableName(collectionName)

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	// Use ILIKE for case-insensitive pattern matching
	sqlQuery := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata,
		       CASE 
		           WHEN content ILIKE $1 THEN 1.0
		           WHEN text ILIKE $1 THEN 0.8
		           ELSE 0.5
		       END as score
		FROM %s
		WHERE content ILIKE $1 OR text ILIKE $1
		ORDER BY score DESC, id
		LIMIT $2
	`, tableName)

	searchPattern := "%" + query + "%"
	rows, err := a.db.QueryContext(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, a.wrapError(err, "search by content")
	}
	defer rows.Close()

	var results []*vectordb.QueryResult
	for rows.Next() {
		var id, content, text, image, imageData, url, metadataJSON string
		var score float64

		if err := rows.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON, &score); err != nil {
			continue
		}

		docMetadata, err := a.convertJSONToMetadata(metadataJSON)
		if err != nil {
			continue
		}

		doc := &vectordb.Document{
			ID:        id,
			Content:   content,
			Text:      text,
			Image:     image,
			ImageData: imageData,
			URL:       url,
			Metadata:  docMetadata,
		}

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    score,
		})
	}

	return results, nil
}

// searchByFullText performs full-text search using PostgreSQL's built-in capabilities
func (a *Adapter) searchByFullText(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	tableName := a.getTableName(collectionName)

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	// Use PostgreSQL's full-text search with improved ranking
	// ts_rank_cd (cover density) provides better ranking than basic ts_rank
	// Normalization flags: 1 (divides by document length) approximates BM25's length normalization
	sqlQuery := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata,
		       ts_rank_cd(
		           to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, '')),
		           plainto_tsquery('english', $1),
		           1  -- Normalize by document length (BM25-like)
		       ) as score
		FROM %s
		WHERE to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, '')) @@ plainto_tsquery('english', $1)
		ORDER BY score DESC, id
		LIMIT $2
	`, tableName)

	rows, err := a.db.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, a.wrapError(err, "search by full text")
	}
	defer rows.Close()

	var results []*vectordb.QueryResult
	for rows.Next() {
		var id, content, text, image, imageData, url, metadataJSON string
		var score float64

		if err := rows.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON, &score); err != nil {
			continue
		}

		docMetadata, err := a.convertJSONToMetadata(metadataJSON)
		if err != nil {
			continue
		}

		doc := &vectordb.Document{
			ID:        id,
			Content:   content,
			Text:      text,
			Image:     image,
			ImageData: imageData,
			URL:       url,
			Metadata:  docMetadata,
		}

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    score,
		})
	}

	return results, nil
}

// mergeSearchResults merges and deduplicates search results from multiple sources
func (a *Adapter) mergeSearchResults(semanticResults, bm25Results []*vectordb.QueryResult, options *vectordb.QueryOptions) []*vectordb.QueryResult {
	// Create a map to track documents by ID and combine scores
	resultMap := make(map[string]*vectordb.QueryResult)

	// Add semantic results with weight
	for _, result := range semanticResults {
		resultMap[result.Document.ID] = &vectordb.QueryResult{
			Document: result.Document,
			Score:    result.Score * 0.6, // 60% weight for semantic
		}
	}

	// Add BM25 results with weight, combining scores if document already exists
	for _, result := range bm25Results {
		if existing, exists := resultMap[result.Document.ID]; exists {
			existing.Score += result.Score * 0.4 // 40% weight for BM25
		} else {
			resultMap[result.Document.ID] = &vectordb.QueryResult{
				Document: result.Document,
				Score:    result.Score * 0.4,
			}
		}
	}

	// Convert map back to slice and sort by score
	var mergedResults []*vectordb.QueryResult
	for _, result := range resultMap {
		mergedResults = append(mergedResults, result)
	}

	// Sort by score descending
	for i := 0; i < len(mergedResults)-1; i++ {
		for j := i + 1; j < len(mergedResults); j++ {
			if mergedResults[i].Score < mergedResults[j].Score {
				mergedResults[i], mergedResults[j] = mergedResults[j], mergedResults[i]
			}
		}
	}

	// Apply limit
	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	if len(mergedResults) > limit {
		mergedResults = mergedResults[:limit]
	}

	return mergedResults
}

// searchByVectorSimilarity performs vector similarity search using pgvector
func (a *Adapter) searchByVectorSimilarity(ctx context.Context, collectionName string, queryEmbedding []float64, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	tableName := a.getTableName(collectionName)

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	// Convert query embedding to pgvector format
	queryVector := floatsToVector(queryEmbedding)

	// Disable index scans for vector similarity search
	// The ivfflat index can cause issues with small datasets or when not properly trained
	// Using sequential scan ensures accurate results
	_, _ = a.db.ExecContext(ctx, "SET enable_indexscan = OFF")
	_, _ = a.db.ExecContext(ctx, "SET enable_bitmapscan = OFF")

	// Note: We embed the vector directly in the SQL query instead of using a parameter
	// because lib/pq driver doesn't properly support vector type parameters
	sqlQuery := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata,
		       COALESCE(1 - (embedding <=> '%s'::vector), 0) as score
		FROM %s
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> '%s'::vector
		LIMIT $1
	`, queryVector, tableName, queryVector)

	rows, err := a.db.QueryContext(ctx, sqlQuery, limit)
	if err != nil {
		return nil, a.wrapError(err, "search by vector similarity")
	}
	defer rows.Close()

	var results []*vectordb.QueryResult
	for rows.Next() {
		var id, content, text, image, imageData, url, metadataJSON string
		var score float64

		if err := rows.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON, &score); err != nil {
			continue
		}

		docMetadata, err := a.convertJSONToMetadata(metadataJSON)
		if err != nil {
			continue
		}

		doc := &vectordb.Document{
			ID:        id,
			Content:   content,
			Text:      text,
			Image:     image,
			ImageData: imageData,
			URL:       url,
			Metadata:  docMetadata,
		}

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    score,
		})
	}

	// Re-enable indexes after query
	_, _ = a.db.ExecContext(ctx, "SET enable_indexscan = ON")
	_, _ = a.db.ExecContext(ctx, "SET enable_bitmapscan = ON")

	return results, nil
}
