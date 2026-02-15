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
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

	if a.client == nil {
		return nil, fmt.Errorf("Pinecone client not initialized")
	}

	// Get the collection's schema to determine embedding model
	schema, err := a.GetSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}

	embeddingModel := schema.Vectorizer
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small" // Fallback default
	}

	// Get the index's actual dimensions to verify compatibility
	indexDims, err := a.getIndexDimensions(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get index dimensions: %w", err)
	}

	// Check if this is an OpenAI model or OSS model
	isOpenAI := embeddingModel == "text-embedding-3-small" ||
		embeddingModel == "text-embedding-3-large" ||
		embeddingModel == "text-embedding-ada-002"

	// Generate embedding for query using detected model
	var embeddingFloat64 []float64
	if isOpenAI {
		// Use LLM client for OpenAI models
		if a.llmClient == nil {
			return nil, fmt.Errorf("Pinecone: SearchSemantic requires OpenAI API key for OpenAI embedding models. Please set OPENAI_API_KEY environment variable")
		}
		embeddingFloat64, err = a.llmClient.GenerateEmbedding(ctx, query, embeddingModel)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query embedding with model '%s': %w", embeddingModel, err)
		}
	} else {
		// Use embedding provider factory for OSS models (sentence-transformers, Ollama, etc.)
		provider, err := a.createEmbeddingProvider(ctx, embeddingModel)
		if err != nil {
			return nil, fmt.Errorf("Pinecone: failed to create embedding provider for model '%s': %w", embeddingModel, err)
		}

		embeddingFloat64, err = provider.GenerateEmbedding(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("Pinecone: failed to generate query embedding with model '%s': %w", embeddingModel, err)
		}
	}

	// Verify dimensions match
	if len(embeddingFloat64) != indexDims {
		return nil, fmt.Errorf("embedding dimension mismatch: index has %d dimensions but query embedding has %d dimensions\n\n"+
			"This usually means the index was created with a different embedding model.\n"+
			"Solutions:\n"+
			"  1. Use the same embedding model that was used to create the index\n"+
			"  2. Recreate the index with the current embedding model configuration\n"+
			"  3. Configure PINECONE_VECTOR_DIMENSIONS=%d in your .env file",
			indexDims, len(embeddingFloat64), indexDims)
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
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeQuery))
	defer cancel()

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
