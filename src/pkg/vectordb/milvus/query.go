// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs semantic search using Milvus vector search
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Check if LLM client is available for embedding generation
	if a.llmClient == nil {
		return nil, fmt.Errorf("SearchSemantic requires OpenAI API key for embedding generation. Please set OPENAI_API_KEY environment variable")
	}

	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Generate embedding for query using LLM client
	queryEmbedding64, err := a.llmClient.GenerateEmbedding(ctx, query, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert float64 to float32 for Milvus
	queryEmbedding := make([]float32, len(queryEmbedding64))
	for i, v := range queryEmbedding64 {
		queryEmbedding[i] = float32(v)
	}

	// Build search vectors
	searchVectors := []entity.Vector{
		entity.FloatVector(queryEmbedding),
	}

	// Output fields to retrieve
	outputFields := []string{
		FieldDocumentID, FieldText, FieldContent,
		FieldImage, FieldImageData, FieldURL, FieldMetadata,
	}

	// Search parameters
	sp, _ := entity.NewIndexIvfFlatSearchParam(16) // nprobe=16

	// Perform vector search
	results, err := a.client.Search(
		ctx,
		collectionName,
		nil, // partitions (nil = all)
		"",  // expression filter (empty = no filter)
		outputFields,
		searchVectors,
		FieldEmbedding,
		a.getMetricType(),
		opts.TopK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Parse results
	return a.parseSearchResults(results)
}

// SearchBM25 performs keyword-based search using Milvus BM25 (Milvus 2.4+)
func (c *Client) SearchBM25(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// NOTE: Milvus native BM25 support requires:
	// 1. Milvus 2.4.0 or later
	// 2. BM25 function enabled on text fields
	// 3. Proper index configuration

	// For now, we'll use simple text filtering via expression
	// Full BM25 implementation requires Milvus 2.4+ with BM25 functions

	// Build expression for text search (contains query terms)
	// This is a simplified approach - full BM25 requires function support
	expr := fmt.Sprintf("%s like \"%%%s%%\" or %s like \"%%%s%%\"",
		FieldText, query, FieldContent, query)

	outputFields := []string{
		FieldDocumentID, FieldText, FieldContent,
		FieldImage, FieldImageData, FieldURL, FieldMetadata,
	}

	// Query documents matching the expression
	result, err := c.client.Query(
		ctx,
		collectionName,
		nil, // partitions
		expr,
		outputFields,
		client.WithLimit(int64(opts.TopK)),
	)
	if err != nil {
		return nil, fmt.Errorf("BM25 search failed: %w", err)
	}

	return c.parseQueryResults(result, 1.0) // Use constant score for BM25
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Hybrid search implementation:
	// 1. Perform vector search
	// 2. Perform BM25 search
	// 3. Combine results using RRF (Reciprocal Rank Fusion)

	// Get vector search results
	vectorResults, err := a.SearchSemantic(ctx, collectionName, query, opts)
	if err != nil {
		// If vector search fails, fall back to BM25
		return a.SearchBM25(ctx, collectionName, query, opts)
	}

	// Get BM25 results
	bm25Results, err := a.SearchBM25(ctx, collectionName, query, opts)
	if err != nil {
		// If BM25 fails, return vector results only
		return vectorResults, nil
	}

	// Combine using Reciprocal Rank Fusion
	combined := combineResultsRRF(vectorResults, bm25Results, opts.TopK)
	return combined, nil
}

// SearchByMetadata searches documents by metadata fields
func (c *Client) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Build expression filter for metadata
	// Note: This is simplified - proper JSON querying requires Milvus 2.3+
	expr := ""
	i := 0
	for key, value := range metadata {
		if i > 0 {
			expr += " and "
		}
		expr += fmt.Sprintf("metadata[\"%s\"] == \"%v\"", key, value)
		i++
	}

	if expr == "" {
		expr = "" // No filter
	}

	outputFields := []string{
		FieldDocumentID, FieldText, FieldContent,
		FieldImage, FieldImageData, FieldURL, FieldMetadata,
	}

	// Query documents matching metadata
	result, err := c.client.Query(
		ctx,
		collectionName,
		nil, // partitions
		expr,
		outputFields,
		client.WithLimit(int64(opts.TopK)),
	)
	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}

	return c.parseQueryResults(result, 1.0) // No score for metadata search
}

// parseSearchResults converts Milvus search results to QueryResults
func (a *Adapter) parseSearchResults(results []client.SearchResult) ([]*vectordb.QueryResult, error) {
	if len(results) == 0 {
		return []*vectordb.QueryResult{}, nil
	}

	// Get first result set (we only searched with one query vector)
	result := results[0]

	queryResults := make([]*vectordb.QueryResult, 0, result.ResultCount)

	for i := 0; i < result.ResultCount; i++ {
		score := result.Scores[i]

		// Extract fields from result
		docID := result.Fields.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[i]
		text := result.Fields.GetColumn(FieldText).(*entity.ColumnVarChar).Data()[i]
		content := result.Fields.GetColumn(FieldContent).(*entity.ColumnVarChar).Data()[i]
		image := result.Fields.GetColumn(FieldImage).(*entity.ColumnVarChar).Data()[i]
		imageData := result.Fields.GetColumn(FieldImageData).(*entity.ColumnVarChar).Data()[i]
		url := result.Fields.GetColumn(FieldURL).(*entity.ColumnVarChar).Data()[i]

		var metadata map[string]interface{}
		if metadataCol := result.Fields.GetColumn(FieldMetadata); metadataCol != nil {
			metadataBytes := metadataCol.(*entity.ColumnJSONBytes).Data()[i]
			metadata = mustUnmarshalJSON(metadataBytes)
		}

		queryResults = append(queryResults, &vectordb.QueryResult{
			Document: *a.fromMilvusDocument(docID, text, content, image, imageData, url, metadata),
			Score:    float64(score),
		})
	}

	return queryResults, nil
}

// parseQueryResults converts Milvus query results to QueryResults
func (c *Client) parseQueryResults(result client.ResultSet, defaultScore float64) ([]*vectordb.QueryResult, error) {
	if result.Len() == 0 {
		return []*vectordb.QueryResult{}, nil
	}

	queryResults := make([]*vectordb.QueryResult, 0, result.Len())

	for i := 0; i < result.Len(); i++ {
		docID := result.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[i]
		text := result.GetColumn(FieldText).(*entity.ColumnVarChar).Data()[i]
		content := result.GetColumn(FieldContent).(*entity.ColumnVarChar).Data()[i]
		image := result.GetColumn(FieldImage).(*entity.ColumnVarChar).Data()[i]
		imageData := result.GetColumn(FieldImageData).(*entity.ColumnVarChar).Data()[i]
		url := result.GetColumn(FieldURL).(*entity.ColumnVarChar).Data()[i]

		var metadata map[string]interface{}
		if metadataCol := result.GetColumn(FieldMetadata); metadataCol != nil {
			metadataBytes := metadataCol.(*entity.ColumnJSONBytes).Data()[i]
			metadata = mustUnmarshalJSON(metadataBytes)
		}

		queryResults = append(queryResults, &vectordb.QueryResult{
			Document: *c.fromMilvusDocument(docID, text, content, image, imageData, url, metadata),
			Score:    defaultScore,
		})
	}

	return queryResults, nil
}

// combineResultsRRF combines results using Reciprocal Rank Fusion
func combineResultsRRF(vectorResults, bm25Results []*vectordb.QueryResult, topK int) []*vectordb.QueryResult {
	// Reciprocal Rank Fusion (RRF) formula:
	// score(d) = Σ 1 / (k + rank(d))
	// where k is typically 60
	const k = 60.0

	// Map document IDs to combined scores
	scores := make(map[string]float64)
	docs := make(map[string]*vectordb.Document)

	// Add vector search results
	for rank, result := range vectorResults {
		scores[result.Document.ID] += 1.0 / (k + float64(rank+1))
		docs[result.Document.ID] = &result.Document
	}

	// Add BM25 results
	for rank, result := range bm25Results {
		scores[result.Document.ID] += 1.0 / (k + float64(rank+1))
		if _, exists := docs[result.Document.ID]; !exists {
			docs[result.Document.ID] = &result.Document
		}
	}

	// Convert to slice and sort by score
	combined := make([]*vectordb.QueryResult, 0, len(scores))
	for id, score := range scores {
		combined = append(combined, &vectordb.QueryResult{
			Document: *docs[id],
			Score:    score,
		})
	}

	// Sort by score descending
	for i := 0; i < len(combined); i++ {
		for j := i + 1; j < len(combined); j++ {
			if combined[j].Score > combined[i].Score {
				combined[i], combined[j] = combined[j], combined[i]
			}
		}
	}

	// Limit to topK
	if len(combined) > topK {
		combined = combined[:topK]
	}

	return combined
}
