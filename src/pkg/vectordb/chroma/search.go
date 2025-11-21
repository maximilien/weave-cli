// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs semantic search using vector embeddings
// Note: Chroma requires embeddings to be provided or uses its embedding function
func (c *Client) SearchSemantic(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Set default TopK if not specified
	topK := 10
	if opts != nil && opts.TopK > 0 {
		topK = opts.TopK
	}

	// Use Chroma's Query with text (relies on collection's embedding function)
	result, err := collection.Query(ctx, []string{query}, int32(topK), nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Convert results
	var results []*vectordb.QueryResult
	if len(result.Documents) > 0 && len(result.Documents[0]) > 0 {
		for i, doc := range result.Documents[0] {
			qr := &vectordb.QueryResult{
				Document: vectordb.Document{
					Content: doc,
				},
			}

			// Add ID if available
			if len(result.Ids) > 0 && i < len(result.Ids[0]) {
				qr.Document.ID = result.Ids[0][i]
			}

			// Add distance/score if available
			if len(result.Distances) > 0 && i < len(result.Distances[0]) {
				// Chroma returns distances, convert to similarity score
				qr.Score = float64(1.0 - result.Distances[0][i])
			}

			// Add metadata if available
			if len(result.Metadatas) > 0 && i < len(result.Metadatas[0]) && result.Metadatas[0][i] != nil {
				qr.Document.Metadata = make(map[string]interface{})
				for k, v := range result.Metadatas[0][i] {
					switch k {
					case "url":
						if s, ok := v.(string); ok {
							qr.Document.URL = s
						}
					case "image":
						if s, ok := v.(string); ok {
							qr.Document.Image = s
						}
					default:
						qr.Document.Metadata[k] = v
					}
				}
			}

			results = append(results, qr)
		}
	}

	return results, nil
}

// SearchBM25 performs keyword-based search using BM25
// Note: Chroma does not support BM25 search natively
func (c *Client) SearchBM25(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("BM25 search is not supported by Chroma; use SearchSemantic instead")
}

// SearchHybrid performs hybrid search combining vector and keyword search
// Note: Chroma does not support hybrid search natively
func (c *Client) SearchHybrid(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Fall back to semantic search
	return c.SearchSemantic(ctx, collectionName, query, opts)
}

// SearchByMetadata searches documents by metadata fields
func (c *Client) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Get documents matching metadata filter
	result, err := collection.Get(ctx, nil, metadata, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}

	// Convert results
	var results []*vectordb.QueryResult
	for i, id := range result.Ids {
		qr := &vectordb.QueryResult{
			Document: vectordb.Document{
				ID: id,
			},
			Score: 1.0, // No score for metadata-only search
		}

		// Add content if available
		if i < len(result.Documents) {
			qr.Document.Content = result.Documents[i]
		}

		// Add metadata if available
		if i < len(result.Metadatas) && result.Metadatas[i] != nil {
			qr.Document.Metadata = make(map[string]interface{})
			for k, v := range result.Metadatas[i] {
				switch k {
				case "url":
					if s, ok := v.(string); ok {
						qr.Document.URL = s
					}
				case "image":
					if s, ok := v.(string); ok {
						qr.Document.Image = s
					}
				default:
					qr.Document.Metadata[k] = v
				}
			}
		}

		results = append(results, qr)

		// Apply limit if specified
		if opts != nil && opts.TopK > 0 && len(results) >= opts.TopK {
			break
		}
	}

	return results, nil
}
