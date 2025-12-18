// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs a semantic (vector) search using kNN
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Generate embedding for the query
	if a.llmClient == nil {
		return nil, fmt.Errorf("LLM client required for semantic search")
	}

	embedding, err := a.llmClient.GenerateEmbedding(ctx, query, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %w", err)
	}

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Convert embedding to float32
	queryVector := make([]float32, len(embedding))
	for i, v := range embedding {
		queryVector[i] = float32(v)
	}

	// Build kNN query
	knnQuery := types.Query{
		Knn: &types.KnnQuery{
			Field:         "vector_field",
			QueryVector:   queryVector,
			K:             &topK,
			NumCandidates: &topK,
		},
	}

	req := &search.Request{
		Query: &knnQuery,
		Size:  &topK,
	}

	// Execute search
	res, err := a.client.Search().
		Index(collectionName).
		Request(req).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	return a.convertSearchResults(res)
}

// SearchBM25 performs a full-text search using BM25
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Build BM25 query (multi_match on text and content fields)
	bm25Query := types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:  query,
			Fields: []string{"text", "content"},
		},
	}

	req := &search.Request{
		Query: &bm25Query,
		Size:  &topK,
	}

	// Execute search
	res, err := a.client.Search().
		Index(collectionName).
		Request(req).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("BM25 search failed: %w", err)
	}

	return a.convertSearchResults(res)
}

// SearchHybrid performs a hybrid search combining semantic and BM25 results
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Generate embedding for the query
	if a.llmClient == nil {
		return nil, fmt.Errorf("LLM client required for hybrid search")
	}

	embedding, err := a.llmClient.GenerateEmbedding(ctx, query, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %w", err)
	}

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Convert embedding to float32
	queryVector := make([]float32, len(embedding))
	for i, v := range embedding {
		queryVector[i] = float32(v)
	}

	// Build hybrid query using kNN with filter
	// This combines vector similarity with text matching
	textFilter := &types.Query{
		MultiMatch: &types.MultiMatchQuery{
			Query:  query,
			Fields: []string{"text", "content"},
		},
	}

	knnQuery := types.Query{
		Knn: &types.KnnQuery{
			Field:         "vector_field",
			QueryVector:   queryVector,
			K:             &topK,
			NumCandidates: &topK,
			Filter:        []types.Query{*textFilter},
		},
	}

	req := &search.Request{
		Query: &knnQuery,
		Size:  &topK,
	}

	// Execute search
	res, err := a.client.Search().
		Index(collectionName).
		Request(req).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	return a.convertSearchResults(res)
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Build metadata filters
	var filters []types.Query
	for key, value := range metadata {
		fieldName := fmt.Sprintf("metadata.%s", key)
		termQuery := types.Query{
			Term: map[string]types.TermQuery{
				fieldName: {Value: value},
			},
		}
		filters = append(filters, termQuery)
	}

	// Combine filters with bool query
	metadataQuery := types.Query{
		Bool: &types.BoolQuery{
			Must: filters,
		},
	}

	req := &search.Request{
		Query: &metadataQuery,
		Size:  &topK,
	}

	// Execute search
	res, err := a.client.Search().
		Index(collectionName).
		Request(req).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}

	return a.convertSearchResults(res)
}

// convertSearchResults converts Elasticsearch search response to QueryResults
func (a *Adapter) convertSearchResults(res *search.Response) ([]*vectordb.QueryResult, error) {
	var results []*vectordb.QueryResult

	for _, hit := range res.Hits.Hits {
		// Decode source
		var esDoc elasticsearchDocument
		if err := json.Unmarshal(hit.Source_, &esDoc); err != nil {
			continue // Skip malformed documents
		}

		// Convert to vectordb.Document
		doc := &vectordb.Document{
			ID:        esDoc.DocumentID,
			Text:      esDoc.Text,
			Content:   esDoc.Content,
			Image:     esDoc.Image,
			ImageData: esDoc.ImageData,
			URL:       esDoc.URL,
			Metadata:  esDoc.Metadata,
		}

		// Get score (default to 0 if nil)
		score := 0.0
		if hit.Score_ != nil {
			score = float64(*hit.Score_)
		}

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    score,
		})
	}

	return results, nil
}
