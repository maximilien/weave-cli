// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

// opensearchDocument represents the document structure in OpenSearch
type opensearchDocument struct {
	DocumentID string                 `json:"document_id"`
	Text       string                 `json:"text"`
	Content    string                 `json:"content"`
	Image      string                 `json:"image,omitempty"`
	ImageData  string                 `json:"image_data,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Vector     []float32              `json:"vector,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SearchSemantic performs k-NN vector search
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

	// Build k-NN query for OpenSearch
	knnQuery := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"knn": map[string]interface{}{
				"vector": map[string]interface{}{
					"vector": queryVector,
					"k":      topK,
				},
			},
		},
	}

	// Execute search
	resp, err := a.client.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{collectionName},
			Body:    opensearchutil.NewJSONReader(knnQuery),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	return a.convertSearchResults(resp)
}

// SearchBM25 performs BM25 text search
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Build BM25 query (multi_match on text and content fields)
	bm25Query := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"text", "content"},
			},
		},
	}

	// Execute search
	resp, err := a.client.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{collectionName},
			Body:    opensearchutil.NewJSONReader(bm25Query),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("BM25 search failed: %w", err)
	}

	return a.convertSearchResults(resp)
}

// SearchHybrid performs hybrid search (vector + BM25)
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

	// Build hybrid query using bool query with should clauses
	hybridQuery := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"knn": map[string]interface{}{
							"vector": map[string]interface{}{
								"vector": queryVector,
								"k":      topK,
							},
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":  query,
							"fields": []string{"text", "content"},
						},
					},
				},
			},
		},
	}

	// Execute search
	resp, err := a.client.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{collectionName},
			Body:    opensearchutil.NewJSONReader(hybridQuery),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	return a.convertSearchResults(resp)
}

// SearchByMetadata searches documents by metadata filters
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Default TopK
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Build metadata filters
	var filters []map[string]interface{}
	for key, value := range metadata {
		fieldName := fmt.Sprintf("metadata.%s", key)
		termQuery := map[string]interface{}{
			"term": map[string]interface{}{
				fieldName: value,
			},
		}
		filters = append(filters, termQuery)
	}

	// Combine filters with bool query
	metadataQuery := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": filters,
			},
		},
	}

	// Execute search
	resp, err := a.client.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{collectionName},
			Body:    opensearchutil.NewJSONReader(metadataQuery),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}

	return a.convertSearchResults(resp)
}

// convertSearchResults converts OpenSearch search response to QueryResults
func (a *Adapter) convertSearchResults(resp *opensearchapi.SearchResp) ([]*vectordb.QueryResult, error) {
	var results []*vectordb.QueryResult

	for _, hit := range resp.Hits.Hits {
		// Decode source
		var osDoc opensearchDocument
		if err := json.Unmarshal(hit.Source, &osDoc); err != nil {
			continue // Skip malformed documents
		}

		// Convert to vectordb.Document
		doc := &vectordb.Document{
			ID:        osDoc.DocumentID,
			Text:      osDoc.Text,
			Content:   osDoc.Content,
			Image:     osDoc.Image,
			ImageData: osDoc.ImageData,
			URL:       osDoc.URL,
			Metadata:  osDoc.Metadata,
		}

		// Get score
		score := float64(hit.Score)

		results = append(results, &vectordb.QueryResult{
			Document: *doc,
			Score:    score,
		})
	}

	return results, nil
}
