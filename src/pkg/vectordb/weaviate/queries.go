// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	weaviateClient "github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// SearchSemantic performs semantic search using vector embeddings
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	weaviateOptions := a.convertQueryOptions(options)

	// Use the main Query method which defaults to semantic search
	results, err := a.client.Query(ctx, collectionName, query, *weaviateOptions)
	if err != nil {
		return nil, a.wrapError(err, "semantic search")
	}

	return a.convertWeaviateQueryResults(results), nil
}

// SearchBM25 performs keyword-based search using BM25
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	weaviateOptions := a.convertQueryOptions(options)
	weaviateOptions.UseBM25 = true

	results, err := a.client.Query(ctx, collectionName, query, *weaviateOptions)
	if err != nil {
		return nil, a.wrapError(err, "BM25 search")
	}

	return a.convertWeaviateQueryResults(results), nil
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// For hybrid search, we'll try semantic first, then BM25 if semantic fails or returns few results
	semanticResults, semanticErr := a.SearchSemantic(ctx, collectionName, query, options)

	// If semantic search works and returns enough results, use it
	if semanticErr == nil && len(semanticResults) >= (options.TopK/2) {
		return semanticResults, nil
	}

	// Otherwise, fall back to BM25
	bm25Results, bm25Err := a.SearchBM25(ctx, collectionName, query, options)
	if bm25Err != nil {
		// If both fail, return the semantic error (primary method)
		if semanticErr != nil {
			return nil, a.wrapError(semanticErr, "hybrid search (semantic failed)")
		}
		return nil, a.wrapError(bm25Err, "hybrid search (BM25 failed)")
	}

	return bm25Results, nil
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	weaviateOptions := a.convertQueryOptions(options)

	// Use QueryWithFilters for metadata search
	results, err := a.client.QueryWithFilters(ctx, collectionName, "", *weaviateOptions, metadata)
	if err != nil {
		return nil, a.wrapError(err, "metadata search")
	}

	return a.convertWeaviateQueryResults(results), nil
}

// convertWeaviateQueryResults converts Weaviate QueryResult to vectordb QueryResult
func (a *Adapter) convertWeaviateQueryResults(results []weaviateClient.QueryResult) []*vectordb.QueryResult {
	if results == nil {
		return nil
	}

	queryResults := make([]*vectordb.QueryResult, len(results))
	for i, result := range results {
		queryResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       result.ID,
				Content:  result.Content,
				Metadata: result.Metadata,
			},
			Score: result.Score,
		}
	}
	return queryResults
}
