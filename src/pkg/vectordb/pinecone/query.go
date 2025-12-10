// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

// SearchSemantic performs a semantic search in Pinecone
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	if a.llmClient == nil {
		return nil, fmt.Errorf("OpenAI client not configured for embedding generation")
	}

	// Generate embedding for query
	embeddingFloat64, err := a.llmClient.GenerateEmbedding(ctx, query, "text-embedding-3-small")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to float32
	embedding := make([]float32, len(embeddingFloat64))
	for i, v := range embeddingFloat64 {
		embedding[i] = float32(v)
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Prepare query parameters
	topK := uint32(10) // Default
	if options != nil && options.TopK > 0 {
		topK = uint32(options.TopK)
	}

	// Build metadata filter if provided (currently not in QueryOptions interface)
	var metadataFilter *pinecone.MetadataFilter
	// Note: vectordb.QueryOptions doesn't have Filters field yet

	// Perform query
	queryResp, err := idxConn.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          embedding,
		TopK:            topK,
		MetadataFilter:  metadataFilter,
		IncludeValues:   true,
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	// Convert to QueryResults
	results := make([]*vectordb.QueryResult, 0, len(queryResp.Matches))
	if queryResp.Matches != nil {
		for _, match := range queryResp.Matches {
			doc := vectordb.Document{
				ID: match.Vector.Id,
			}

			// Extract content and metadata
			if match.Vector.Metadata != nil {
				metadataMap := match.Vector.Metadata.AsMap()
				doc.Metadata = make(map[string]interface{})
				for k, v := range metadataMap {
					if k == "content" {
						if content, ok := v.(string); ok {
							doc.Content = content
						}
					} else {
						doc.Metadata[k] = v
					}
				}
			}

			result := &vectordb.QueryResult{
				Document: doc,
				Score:    float64(match.Score),
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SearchByMetadata searches documents by metadata filters
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Pinecone requires a query vector for search
	// For metadata-only search, we convert GetDocumentsByMetadata results to QueryResults
	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	docs, err := a.GetDocumentsByMetadata(ctx, collectionName, metadata, limit)
	if err != nil {
		return nil, err
	}

	// Convert documents to query results
	results := make([]*vectordb.QueryResult, 0, len(docs))
	for _, doc := range docs {
		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    1.0, // No similarity score for metadata-only search
		})
	}

	return results, nil
}

// SearchBM25 performs keyword-based search
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Pinecone does not support native BM25/keyword search
	// It's a pure vector database
	_ = options // Will be used for topK, etc.
	return nil, fmt.Errorf("BM25 search not supported by Pinecone (vector-only database)")
}

// SearchHybrid performs hybrid search combining vector and keyword
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Pinecone supports sparse-dense vectors for hybrid search
	// This requires special index configuration and sparse vector generation
	// For now, fall back to pure vector search
	return a.SearchSemantic(ctx, collectionName, query, options)
}
