//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"fmt"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs semantic search using vector embeddings
// Note: Chroma requires embeddings to be provided or uses its embedding function
func (c *Client) SearchSemantic(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Get collection
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Set default TopK if not specified
	topK := 10
	if opts != nil && opts.TopK > 0 {
		topK = opts.TopK
	}

	// Use Chroma's Query with text
	result, err := collection.Query(ctx,
		chroma.WithQueryTexts(query),
		chroma.WithNResults(topK),
		chroma.WithIncludeQuery(chroma.IncludeDocuments, chroma.IncludeMetadatas),
	)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Convert results - Query returns groups for each query text
	var results []*vectordb.QueryResult
	recordGroups := result.ToRecordsGroups()
	if len(recordGroups) == 0 {
		return results, nil
	}
	records := recordGroups[0] // First group for our single query
	for _, record := range records {
		qr := &vectordb.QueryResult{
			Document: vectordb.Document{
				ID: string(record.ID()),
			},
		}
		if record.Document() != nil {
			qr.Document.Content = record.Document().ContentString()
		}

		// Default score
		qr.Score = 1.0

		// Add metadata if available
		if record.Metadata() != nil {
			qr.Document.Metadata = make(map[string]interface{})
			if v, ok := record.Metadata().GetString("url"); ok {
				qr.Document.URL = v
			}
			if v, ok := record.Metadata().GetString("image"); ok {
				qr.Document.Image = v
			}
			if v, ok := record.Metadata().GetString("filename"); ok {
				qr.Document.Metadata["filename"] = v
			}
			if v, ok := record.Metadata().GetString("type"); ok {
				qr.Document.Metadata["type"] = v
			}
		}

		results = append(results, qr)
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
	collection, err := c.getCollection(ctx, collectionName)
	if err != nil {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	// Build where clauses from metadata
	var clauses []chroma.WhereClause
	for k, v := range metadata {
		switch val := v.(type) {
		case string:
			clauses = append(clauses, chroma.EqString(k, val))
		case int:
			clauses = append(clauses, chroma.EqInt(k, val))
		case float64:
			clauses = append(clauses, chroma.EqFloat(k, float32(val)))
		case bool:
			clauses = append(clauses, chroma.EqBool(k, val))
		}
	}

	// Build get options
	getOpts := []chroma.CollectionGetOption{
		chroma.WithIncludeGet(chroma.IncludeDocuments, chroma.IncludeMetadatas),
	}

	// Add where filter if we have clauses
	if len(clauses) == 1 {
		getOpts = append(getOpts, chroma.WithWhereGet(clauses[0]))
	} else if len(clauses) > 1 {
		getOpts = append(getOpts, chroma.WithWhereGet(chroma.And(clauses...)))
	}

	result, err := collection.Get(ctx, getOpts...)
	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}

	// Convert results using raw result data
	var results []*vectordb.QueryResult
	ids := result.GetIDs()
	documents := result.GetDocuments()
	metadatas := result.GetMetadatas()

	for i, id := range ids {
		qr := &vectordb.QueryResult{
			Document: vectordb.Document{
				ID: string(id),
			},
			Score: 1.0, // No score for metadata-only search
		}

		// Get document content if available
		if i < len(documents) && documents[i] != nil {
			qr.Document.Content = documents[i].ContentString()
		}

		// Add metadata if available
		if i < len(metadatas) && metadatas[i] != nil {
			qr.Document.Metadata = make(map[string]interface{})
			if v, ok := metadatas[i].GetString("url"); ok {
				qr.Document.URL = v
			}
			if v, ok := metadatas[i].GetString("image"); ok {
				qr.Document.Image = v
			}
			if v, ok := metadatas[i].GetString("filename"); ok {
				qr.Document.Metadata["filename"] = v
			}
			if v, ok := metadatas[i].GetString("type"); ok {
				qr.Document.Metadata["type"] = v
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
