// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// SearchSemantic performs semantic search using MongoDB Atlas Vector Search
func (c *Client) SearchSemantic(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Note: This requires embeddings to be generated from the query
	// In production, you would call an embedding service here
	// For now, we return an error indicating embeddings are needed
	return nil, fmt.Errorf("SearchSemantic requires embedding generation - not yet implemented")

	// Example implementation when embeddings are available:
	/*
		collection := c.getCollection(collectionName)

		// Generate embedding for query (placeholder)
		queryEmbedding := []float64{} // TODO: Generate from embedding service

		// Build vector search aggregation pipeline
		pipeline := bson.A{
			bson.D{{Key: "$vectorSearch", Value: bson.D{
				{Key: "index", Value: "vector_index"},
				{Key: "path", Value: "embedding"},
				{Key: "queryVector", Value: queryEmbedding},
				{Key: "numCandidates", Value: opts.TopK * 10},
				{Key: "limit", Value: opts.TopK},
			}}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "document_id", Value: 1},
				{Key: "text", Value: 1},
				{Key: "content", Value: 1},
				{Key: "metadata", Value: 1},
				{Key: "score", Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}},
			}}},
		}

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, fmt.Errorf("vector search failed: %w", err)
		}
		defer cursor.Close(ctx)

		var results []*vectordb.QueryResult
		for cursor.Next(ctx) {
			var result struct {
				DocumentID string                 `bson:"document_id"`
				Text       string                 `bson:"text"`
				Content    string                 `bson:"content"`
				Metadata   map[string]interface{} `bson:"metadata"`
				Score      float64                `bson:"score"`
			}
			if err := cursor.Decode(&result); err != nil {
				return nil, fmt.Errorf("failed to decode result: %w", err)
			}

			results = append(results, &vectordb.QueryResult{
				Document: vectordb.Document{
					ID:       result.DocumentID,
					Text:     result.Text,
					Content:  result.Content,
					Metadata: result.Metadata,
				},
				Score: result.Score,
			})
		}

		return results, cursor.Err()
	*/
}

// SearchBM25 performs keyword-based search using MongoDB text search
func (c *Client) SearchBM25(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)

	// Use aggregation pipeline for better control over text search scoring
	pipeline := bson.A{
		// Match documents with text search
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$text", Value: bson.D{{Key: "$search", Value: query}}},
		}}},
		// Add text score as a field
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "textScore", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
		}}},
		// Sort by text score descending
		bson.D{{Key: "$sort", Value: bson.D{{Key: "textScore", Value: -1}}}},
		// Limit results
		bson.D{{Key: "$limit", Value: opts.TopK}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("BM25 search failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*vectordb.QueryResult
	for cursor.Next(ctx) {
		// Decode into a raw BSON document first to debug
		var raw bson.M
		if err := cursor.Decode(&raw); err != nil {
			return nil, fmt.Errorf("failed to decode raw result: %w", err)
		}

		// Extract fields manually
		mdoc := &mongoDocument{}
		if docID, ok := raw["document_id"].(string); ok {
			mdoc.DocumentID = docID
		}
		if text, ok := raw["text"].(string); ok {
			mdoc.Text = text
		}
		if content, ok := raw["content"].(string); ok {
			mdoc.Content = content
		}
		if metadata, ok := raw["metadata"].(bson.M); ok {
			mdoc.Metadata = make(map[string]interface{})
			for k, v := range metadata {
				mdoc.Metadata[k] = v
			}
		}

		score := 0.0
		if textScore, ok := raw["textScore"].(float64); ok {
			score = textScore
		}

		results = append(results, &vectordb.QueryResult{
			Document: *c.fromMongoDocument(mdoc),
			Score:    score,
		})
	}

	return results, cursor.Err()
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (c *Client) SearchHybrid(ctx context.Context, collectionName, query string, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Hybrid search implementation would:
	// 1. Perform vector search
	// 2. Perform BM25 search
	// 3. Combine results using RRF (Reciprocal Rank Fusion) or weighted scoring

	// For now, fall back to BM25 only
	return c.SearchBM25(ctx, collectionName, query, opts)

	// Example hybrid implementation:
	/*
		// Get vector search results
		vectorResults, err := c.SearchSemantic(ctx, collectionName, query, opts)
		if err != nil {
			return nil, err
		}

		// Get BM25 results
		bm25Results, err := c.SearchBM25(ctx, collectionName, query, opts)
		if err != nil {
			return nil, err
		}

		// Combine using RRF
		combined := combineResultsRRF(vectorResults, bm25Results, opts.TopK)
		return combined, nil
	*/
}

// SearchByMetadata searches documents by metadata fields
func (c *Client) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, opts *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collection := c.getCollection(collectionName)

	// Build filter for metadata fields
	filter := bson.D{}
	for key, value := range metadata {
		filter = append(filter, bson.E{Key: "metadata." + key, Value: value})
	}

	findOpts := options.Find().
		SetLimit(int64(opts.TopK)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("metadata search failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*vectordb.QueryResult
	for cursor.Next(ctx) {
		var mdoc mongoDocument
		if err := cursor.Decode(&mdoc); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}

		results = append(results, &vectordb.QueryResult{
			Document: *c.fromMongoDocument(&mdoc),
			Score:    1.0, // No score for metadata-only search
		})
	}

	return results, cursor.Err()
}

// combineResultsRRF combines search results using Reciprocal Rank Fusion
func combineResultsRRF(results1, results2 []*vectordb.QueryResult, topK int) []*vectordb.QueryResult {
	const k = 60.0 // RRF constant

	// Build score map
	scores := make(map[string]float64)
	docs := make(map[string]*vectordb.Document)

	// Add scores from first result set
	for rank, result := range results1 {
		docID := result.Document.ID
		scores[docID] += 1.0 / (float64(rank+1) + k)
		docs[docID] = &result.Document
	}

	// Add scores from second result set
	for rank, result := range results2 {
		docID := result.Document.ID
		scores[docID] += 1.0 / (float64(rank+1) + k)
		if _, exists := docs[docID]; !exists {
			docs[docID] = &result.Document
		}
	}

	// Convert to results and sort by score
	combined := make([]*vectordb.QueryResult, 0, len(scores))
	for docID, score := range scores {
		combined = append(combined, &vectordb.QueryResult{
			Document: *docs[docID],
			Score:    score,
		})
	}

	// Sort by score descending
	// (Would use sort.Slice in real implementation)

	// Limit to topK
	if len(combined) > topK {
		combined = combined[:topK]
	}

	return combined
}
