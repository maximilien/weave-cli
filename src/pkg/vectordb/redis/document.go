// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument stores a single document as a Redis HASH.
func (c *Client) CreateDocument(ctx context.Context, collection string, doc *vectordb.Document) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeDocument)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	id := doc.ID
	if id == "" {
		id = uuid.New().String()
		doc.ID = id // Write back so caller sees the generated ID
	}

	fields := buildHashFields(doc)
	key := docKey(collection, id)

	if err := c.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("Redis: failed to create document %s: %w", id, err)
	}
	return nil
}

// CreateDocuments stores multiple documents using a pipeline.
func (c *Client) CreateDocuments(ctx context.Context, collection string, docs []*vectordb.Document) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeBulk)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pipe := c.rdb.Pipeline()
	for _, doc := range docs {
		id := doc.ID
		if id == "" {
			id = uuid.New().String()
			doc.ID = id // Write back generated ID
		}
		fields := buildHashFields(doc)
		pipe.HSet(ctx, docKey(collection, id), fields)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("Redis: failed to create %d documents: %w", len(docs), err)
	}
	return nil
}

// GetDocument retrieves a document by ID.
func (c *Client) GetDocument(ctx context.Context, collection, docID string) (*vectordb.Document, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeDocument)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	key := docKey(collection, docID)
	fields, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis: failed to get document %s: %w", docID, err)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("Redis: document %s not found", docID)
	}

	return hashToDocument(docID, fields), nil
}

// UpdateDocument updates (upserts) a document.
func (c *Client) UpdateDocument(ctx context.Context, collection string, doc *vectordb.Document) error {
	return c.CreateDocument(ctx, collection, doc)
}

// DeleteDocument deletes a single document.
func (c *Client) DeleteDocument(ctx context.Context, collection, docID string) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeDocument)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	key := docKey(collection, docID)
	n, err := c.rdb.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("Redis: failed to delete document %s: %w", docID, err)
	}
	if n == 0 {
		return fmt.Errorf("Redis: document %s not found", docID)
	}
	return nil
}

// DeleteDocuments deletes multiple documents.
func (c *Client) DeleteDocuments(ctx context.Context, collection string, ids []string) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeBulk)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = docKey(collection, id)
	}

	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("Redis: failed to delete %d documents: %w", len(ids), err)
	}
	return nil
}

// ListDocuments returns documents using FT.SEARCH with wildcard.
func (c *Client) ListDocuments(ctx context.Context, collection string, limit, offset int) ([]*vectordb.Document, error) {
	timeout := c.getTimeoutFor(vectordb.OperationTypeQuery)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if limit <= 0 {
		limit = 100
	}

	// FT.SEARCH idx "*" LIMIT offset limit RETURN 2 content metadata
	res, err := c.rdb.Do(ctx,
		"FT.SEARCH", indexName(collection), "*",
		"LIMIT", offset, limit,
		"RETURN", "3", defaultContentField, "metadata", "doc_id",
	).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis: failed to list documents: %w", err)
	}

	return parseSearchResults(res, collection), nil
}

// --- helpers ---

// buildHashFields converts a Document into map[string]interface{} for HSET.
func buildHashFields(doc *vectordb.Document) map[string]interface{} {
	fields := map[string]interface{}{
		defaultContentField: doc.Text,
		"doc_id":            doc.ID,
	}

	// Store metadata as JSON string
	if doc.Metadata != nil {
		if b, err := json.Marshal(doc.Metadata); err == nil {
			fields["metadata"] = string(b)
		}
	}

	// Store vector as binary blob (little-endian float32)
	if doc.Embedding != nil {
		fields[defaultVectorField] = float64sToBytes(doc.Embedding)
	}

	return fields
}

// hashToDocument converts Redis HASH fields back to a Document.
func hashToDocument(id string, fields map[string]string) *vectordb.Document {
	doc := &vectordb.Document{
		ID:   id,
		Text: fields[defaultContentField],
	}

	// Restore doc_id if stored
	if storedID, ok := fields["doc_id"]; ok && storedID != "" {
		doc.ID = storedID
	}

	// Parse metadata JSON
	if raw, ok := fields["metadata"]; ok && raw != "" {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &meta); err == nil {
			doc.Metadata = meta
		}
	}

	return doc
}

// float64sToBytes converts []float64 to little-endian []byte (float32).
func float64sToBytes(vec []float64) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(float32(v)))
	}
	return buf
}

// float32sToBytes converts []float32 to little-endian []byte.
func float32sToBytes(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

// parseSearchResults converts FT.SEARCH raw output to Documents.
// FT.SEARCH returns: [totalResults, docKey1, [field1, val1, ...], docKey2, ...]
func parseSearchResults(raw interface{}, collection string) []*vectordb.Document {
	slice, ok := raw.([]interface{})
	if !ok || len(slice) < 1 {
		return nil
	}

	// First element is total count (int64)
	var docs []*vectordb.Document
	prefix := docKeyPrefix(collection)

	for i := 1; i+1 < len(slice); i += 2 {
		keyRaw, _ := slice[i].(string)
		fieldsRaw, _ := slice[i+1].([]interface{})

		// Extract doc ID from key (strip prefix)
		docID := keyRaw
		if len(keyRaw) > len(prefix) {
			docID = keyRaw[len(prefix):]
		}

		fields := flatSliceToMap(fieldsRaw)
		doc := hashToDocument(docID, fields)
		docs = append(docs, doc)
	}

	return docs
}

// flatSliceToMap converts [k1, v1, k2, v2, ...] to map[string]string.
func flatSliceToMap(s []interface{}) map[string]string {
	m := make(map[string]string, len(s)/2)
	for i := 0; i+1 < len(s); i += 2 {
		k, _ := s[i].(string)
		v, _ := s[i+1].(string)
		m[k] = v
	}
	return m
}
