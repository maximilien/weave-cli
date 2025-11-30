// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"
)

// Document represents a document with its vector and metadata (payload)
type Document struct {
	ID       string                 // Unique document ID
	Vector   []float32              // Embedding vector
	Metadata map[string]interface{} // Payload metadata
	Content  string                 // Document content (stored in metadata)
}

// CreateDocument inserts a single document (point) into a collection
func (c *Client) CreateDocument(ctx context.Context, collectionName string, doc *Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}

	// Convert document to Qdrant point
	point, err := documentToPoint(doc)
	if err != nil {
		return fmt.Errorf("failed to convert document to point: %w", err)
	}

	// Upsert the point
	req := &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         []*qdrant.PointStruct{point},
	}

	_, err = c.pointsClient.Upsert(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create document in collection %s: %w", collectionName, err)
	}

	return nil
}

// CreateDocuments inserts multiple documents (points) into a collection in batch
func (c *Client) CreateDocuments(ctx context.Context, collectionName string, docs []*Document) error {
	if len(docs) == 0 {
		return nil
	}

	// Convert documents to Qdrant points
	points := make([]*qdrant.PointStruct, len(docs))
	for i, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}

		point, err := documentToPoint(doc)
		if err != nil {
			return fmt.Errorf("failed to convert document %d to point: %w", i, err)
		}
		points[i] = point
	}

	// Upsert all points
	req := &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	}

	_, err := c.pointsClient.Upsert(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create documents in collection %s: %w", collectionName, err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName string, id string) (*Document, error) {
	// Convert string ID to UUID
	uuidStr, err := stringToUUID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ID to UUID: %w", err)
	}

	// Convert string ID to Qdrant PointId
	pointID := &qdrant.PointId{
		PointIdOptions: &qdrant.PointId_Uuid{
			Uuid: uuidStr,
		},
	}

	req := &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids:            []*qdrant.PointId{pointID},
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	resp, err := c.pointsClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get document %s from collection %s: %w", id, collectionName, err)
	}

	if len(resp.Result) == 0 {
		return nil, fmt.Errorf("document %s not found in collection %s", id, collectionName)
	}

	// Convert Qdrant point to document
	doc, err := pointToDocument(resp.Result[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert point to document: %w", err)
	}

	return doc, nil
}

// UpdateDocument updates a document (same as create, uses upsert)
func (c *Client) UpdateDocument(ctx context.Context, collectionName string, doc *Document) error {
	return c.CreateDocument(ctx, collectionName, doc)
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName string, id string) error {
	// Convert string ID to UUID
	uuidStr, err := stringToUUID(id)
	if err != nil {
		return fmt.Errorf("failed to convert ID to UUID: %w", err)
	}

	// Convert string ID to Qdrant PointId
	pointID := &qdrant.PointId{
		PointIdOptions: &qdrant.PointId_Uuid{
			Uuid: uuidStr,
		},
	}

	req := &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: []*qdrant.PointId{pointID},
				},
			},
		},
	}

	_, delErr := c.pointsClient.Delete(ctx, req)
	if delErr != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: %w", id, collectionName, delErr)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Convert string IDs to Qdrant PointIds
	pointIDs := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		uuidStr, err := stringToUUID(id)
		if err != nil {
			return fmt.Errorf("failed to convert ID %s to UUID: %w", id, err)
		}
		pointIDs[i] = &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: uuidStr,
			},
		}
	}

	req := &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIDs,
				},
			},
		},
	}

	_, err := c.pointsClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete documents from collection %s: %w", collectionName, err)
	}

	return nil
}

// ListDocuments retrieves all documents in a collection using scroll
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int) ([]*Document, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	limitValue := uint32(limit)
	req := &qdrant.ScrollPoints{
		CollectionName: collectionName,
		Limit:          &limitValue,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
		WithVectors:    &qdrant.WithVectorsSelector{SelectorOptions: &qdrant.WithVectorsSelector_Enable{Enable: true}},
	}

	resp, err := c.pointsClient.Scroll(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents from collection %s: %w", collectionName, err)
	}

	// Convert Qdrant points to documents
	docs := make([]*Document, len(resp.Result))
	for i, point := range resp.Result {
		doc, err := pointToDocument(point)
		if err != nil {
			return nil, fmt.Errorf("failed to convert point %d to document: %w", i, err)
		}
		docs[i] = doc
	}

	return docs, nil
}

// documentToPoint converts a Document to a Qdrant PointStruct
func documentToPoint(doc *Document) (*qdrant.PointStruct, error) {
	// Create payload from metadata
	payload := make(map[string]*qdrant.Value)
	for key, value := range doc.Metadata {
		val, err := interfaceToValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert metadata field %s: %w", key, err)
		}
		payload[key] = val
	}

	// Add content to payload if present
	if doc.Content != "" {
		payload["content"] = &qdrant.Value{
			Kind: &qdrant.Value_StringValue{StringValue: doc.Content},
		}
	}

	// Always store original ID in payload for retrieval
	payload["_original_id"] = &qdrant.Value{
		Kind: &qdrant.Value_StringValue{StringValue: doc.ID},
	}

	// Convert ID to valid UUID
	pointID, err := stringToUUID(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ID to UUID: %w", err)
	}

	// Create point
	point := &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: pointID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: doc.Vector,
				},
			},
		},
		Payload: payload,
	}

	return point, nil
}

// pointToDocument converts a Qdrant RetrievedPoint to a Document
func pointToDocument(point *qdrant.RetrievedPoint) (*Document, error) {
	doc := &Document{
		Metadata: make(map[string]interface{}),
	}

	// Extract ID - prefer _original_id from payload if available
	if point.Payload != nil {
		if originalID, ok := point.Payload["_original_id"]; ok {
			if strVal, ok := originalID.Kind.(*qdrant.Value_StringValue); ok {
				doc.ID = strVal.StringValue
			}
		}
	}

	// Fallback to UUID if no original ID found
	if doc.ID == "" && point.Id != nil {
		if uuid, ok := point.Id.PointIdOptions.(*qdrant.PointId_Uuid); ok {
			doc.ID = uuid.Uuid
		}
	}

	// Extract vector - VectorsOutput vs Vectors for input/output
	if point.Vectors != nil {
		if vec, ok := point.Vectors.VectorsOptions.(*qdrant.VectorsOutput_Vector); ok {
			// Use GetDense().GetData() for the new Vector structure
			if dense := vec.Vector.GetDense(); dense != nil {
				doc.Vector = dense.GetData()
			}
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

// interfaceToValue converts a Go interface{} to a Qdrant Value
func interfaceToValue(v interface{}) (*qdrant.Value, error) {
	switch val := v.(type) {
	case string:
		return &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: val}}, nil
	case int:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(val)}}, nil
	case int64:
		return &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: val}}, nil
	case float64:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: val}}, nil
	case float32:
		return &qdrant.Value{Kind: &qdrant.Value_DoubleValue{DoubleValue: float64(val)}}, nil
	case bool:
		return &qdrant.Value{Kind: &qdrant.Value_BoolValue{BoolValue: val}}, nil
	case []int:
		// Convert []int to ListValue
		values := make([]*qdrant.Value, len(val))
		for i, item := range val {
			values[i] = &qdrant.Value{Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(item)}}
		}
		return &qdrant.Value{Kind: &qdrant.Value_ListValue{ListValue: &qdrant.ListValue{Values: values}}}, nil
	case []string:
		// Convert []string to ListValue
		values := make([]*qdrant.Value, len(val))
		for i, item := range val {
			values[i] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: item}}
		}
		return &qdrant.Value{Kind: &qdrant.Value_ListValue{ListValue: &qdrant.ListValue{Values: values}}}, nil
	case []interface{}:
		// Convert []interface{} to ListValue recursively
		values := make([]*qdrant.Value, len(val))
		for i, item := range val {
			convertedVal, err := interfaceToValue(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert array element at index %d: %w", i, err)
			}
			values[i] = convertedVal
		}
		return &qdrant.Value{Kind: &qdrant.Value_ListValue{ListValue: &qdrant.ListValue{Values: values}}}, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %T", v)
	}
}

// valueToInterface converts a Qdrant Value to a Go interface{}
func valueToInterface(v *qdrant.Value) interface{} {
	switch val := v.Kind.(type) {
	case *qdrant.Value_StringValue:
		return val.StringValue
	case *qdrant.Value_IntegerValue:
		return val.IntegerValue
	case *qdrant.Value_DoubleValue:
		return val.DoubleValue
	case *qdrant.Value_BoolValue:
		return val.BoolValue
	case *qdrant.Value_ListValue:
		// Convert ListValue back to []interface{}
		if val.ListValue == nil {
			return []interface{}{}
		}
		result := make([]interface{}, len(val.ListValue.Values))
		for i, item := range val.ListValue.Values {
			result[i] = valueToInterface(item)
		}
		return result
	default:
		return nil
	}
}

// stringToUUID converts a string ID to a valid UUID format
// If the string is already a valid UUID, it returns it as-is
// Otherwise, it generates a deterministic UUID v5 based on the string
func stringToUUID(id string) (string, error) {
	// Try to parse as UUID first
	if _, err := uuid.Parse(id); err == nil {
		return id, nil
	}

	// Generate deterministic UUID v5 from string
	// Using DNS namespace as it's standard and collision-resistant
	namespace := uuid.NameSpaceDNS
	deterministicUUID := uuid.NewSHA1(namespace, []byte(id))
	return deterministicUUID.String(), nil
}

// DeleteAllDocuments deletes all documents from a collection
func (c *Client) DeleteAllDocuments(ctx context.Context, collectionName string) error {
	// Create delete request with empty filter to match all points
	_, err := c.pointsClient.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: &qdrant.Filter{
					// Empty filter matches all points
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete all documents: %w", err)
	}

	return nil
}
