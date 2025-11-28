// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"fmt"

	qdrant "github.com/qdrant/go-client/qdrant"
)

// SearchResult represents a search result with a document and score
type SearchResult struct {
	Document *Document
	Score    float32
}

// SearchSemantic performs a vector similarity search
func (c *Client) SearchSemantic(ctx context.Context, collectionName string, vector []float32, topK int) ([]*SearchResult, error) {
	if topK <= 0 {
		topK = 10 // Default limit
	}

	req := &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Limit:          uint64(topK),
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	resp, err := c.pointsClient.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search in collection %s: %w", collectionName, err)
	}

	// Convert search results to SearchResult objects
	results := make([]*SearchResult, len(resp.Result))
	for i, point := range resp.Result {
		doc, err := scoredPointToDocument(point)
		if err != nil {
			return nil, fmt.Errorf("failed to convert search result %d: %w", i, err)
		}
		results[i] = &SearchResult{
			Document: doc,
			Score:    point.Score,
		}
	}

	return results, nil
}

// SearchByMetadata searches using metadata filters only
func (c *Client) SearchByMetadata(ctx context.Context, collectionName string, filter *qdrant.Filter, limit int) ([]*Document, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	limitValue := uint32(limit)
	req := &qdrant.ScrollPoints{
		CollectionName: collectionName,
		Filter:         filter,
		Limit:          &limitValue,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	resp, err := c.pointsClient.Scroll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search by metadata in collection %s: %w", collectionName, err)
	}

	// Convert results to documents
	docs := make([]*Document, len(resp.Result))
	for i, point := range resp.Result {
		doc, err := pointToDocument(point)
		if err != nil {
			return nil, fmt.Errorf("failed to convert result %d: %w", i, err)
		}
		docs[i] = doc
	}

	return docs, nil
}

// SearchHybrid performs a hybrid search combining vector similarity and metadata filtering
func (c *Client) SearchHybrid(ctx context.Context, collectionName string, vector []float32, filter *qdrant.Filter, topK int) ([]*SearchResult, error) {
	if topK <= 0 {
		topK = 10 // Default limit
	}

	req := &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Filter:         filter,
		Limit:          uint64(topK),
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	resp, err := c.pointsClient.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform hybrid search in collection %s: %w", collectionName, err)
	}

	// Convert search results to SearchResult objects
	results := make([]*SearchResult, len(resp.Result))
	for i, point := range resp.Result {
		doc, err := scoredPointToDocument(point)
		if err != nil {
			return nil, fmt.Errorf("failed to convert search result %d: %w", i, err)
		}
		results[i] = &SearchResult{
			Document: doc,
			Score:    point.Score,
		}
	}

	return results, nil
}

// BuildFilter is a helper to construct Qdrant filters from key-value pairs
func BuildFilter(conditions map[string]interface{}) (*qdrant.Filter, error) {
	must := make([]*qdrant.Condition, 0, len(conditions))

	for key, value := range conditions {
		condition, err := buildCondition(key, value)
		if err != nil {
			return nil, err
		}
		must = append(must, condition)
	}

	return &qdrant.Filter{
		Must: must,
	}, nil
}

// buildCondition creates a Qdrant condition for a single key-value pair
func buildCondition(key string, value interface{}) (*qdrant.Condition, error) {
	switch val := value.(type) {
	case string:
		return &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: key,
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keyword{
							Keyword: val,
						},
					},
				},
			},
		}, nil
	case int, int64:
		intVal := int64(0)
		switch v := val.(type) {
		case int:
			intVal = int64(v)
		case int64:
			intVal = v
		}
		return &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: key,
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Integer{
							Integer: intVal,
						},
					},
				},
			},
		}, nil
	case bool:
		return &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: key,
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Boolean{
							Boolean: val,
						},
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported filter value type for key %s: %T", key, value)
	}
}

// scoredPointToDocument converts a Qdrant ScoredPoint to a Document
func scoredPointToDocument(point *qdrant.ScoredPoint) (*Document, error) {
	doc := &Document{
		Metadata: make(map[string]interface{}),
	}

	// Extract ID
	if point.Id != nil {
		if uuid, ok := point.Id.PointIdOptions.(*qdrant.PointId_Uuid); ok {
			doc.ID = uuid.Uuid
		}
	}

	// Extract vector - VectorsOutput vs Vectors for input/output
	if point.Vectors != nil {
		if vec, ok := point.Vectors.VectorsOptions.(*qdrant.VectorsOutput_Vector); ok {
			// Use GetData() instead of deprecated Data field
			doc.Vector = vec.Vector.GetData()
		}
	}

	// Extract payload
	for key, value := range point.Payload {
		if key == "content" {
			if strVal, ok := value.Kind.(*qdrant.Value_StringValue); ok {
				doc.Content = strVal.StringValue
			}
		} else {
			doc.Metadata[key] = valueToInterface(value)
		}
	}

	return doc, nil
}
