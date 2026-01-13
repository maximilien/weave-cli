// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package testing

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WaitForHealth polls VDB until ready or timeout
func WaitForHealth(ctx context.Context, client vectordb.VectorDBClient, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check timeout: %w", ctx.Err())
		case <-ticker.C:
			if err := client.Health(ctx); err == nil {
				return nil
			}
		}
	}
}

// CreateTestCollection creates collection with automatic cleanup
func CreateTestCollection(t *testing.T, client vectordb.VectorDBClient, name string) func() {
	ctx := context.Background()
	schema := GetTestSchema(name)

	err := client.CreateCollection(ctx, name, schema)
	require.NoError(t, err)

	// Return cleanup function
	return func() {
		_ = client.DeleteCollection(ctx, name)
	}
}

// GetTestSchema returns a standard test schema
func GetTestSchema(className string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:       className,
		Description: "Test collection for integration tests",
		Properties: []vectordb.SchemaProperty{
			{
				Name:        "text",
				DataType:    []string{"text"},
				Description: "Text content",
			},
			{
				Name:        "content",
				DataType:    []string{"text"},
				Description: "Document content",
			},
			{
				Name:        "metadata",
				DataType:    []string{"object"},
				Description: "Document metadata",
			},
		},
		VectorIndexConfig: &vectordb.VectorIndexConfig{
			Distance: "cosine",
		},
	}
}

// GenerateTestDocuments creates sample docs with embeddings
func GenerateTestDocuments(count int, dims int) []*vectordb.Document {
	docs := make([]*vectordb.Document, count)
	for i := 0; i < count; i++ {
		docs[i] = &vectordb.Document{
			ID:      fmt.Sprintf("doc-%d", i),
			Text:    fmt.Sprintf("Test document %d", i),
			Content: fmt.Sprintf("This is test content for document %d", i),
			Vector:  GenerateRandomVector(dims),
			Metadata: map[string]interface{}{
				"index": i,
				"type":  "test",
				"even":  i%2 == 0,
			},
		}
	}
	return docs
}

// GenerateRandomVector creates a random unit vector
func GenerateRandomVector(dims int) []float32 {
	vec := make([]float32, dims)
	for i := range vec {
		vec[i] = rand.Float32()*2 - 1 // Range [-1, 1]
	}
	return NormalizeVector(vec)
}

// NormalizeVector normalizes vector to unit length
func NormalizeVector(vec []float32) []float32 {
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	magnitude := float32(math.Sqrt(float64(sum)))
	if magnitude == 0 {
		return vec
	}
	normalized := make([]float32, len(vec))
	for i, v := range vec {
		normalized[i] = v / magnitude
	}
	return normalized
}

// AssertDocumentEqual compares docs (ignoring vector precision)
func AssertDocumentEqual(t *testing.T, expected, actual *vectordb.Document) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Text, actual.Text)
	assert.Equal(t, expected.Content, actual.Content)

	// Metadata comparison (may need to handle type conversions)
	if expected.Metadata != nil {
		assert.NotNil(t, actual.Metadata)
		for k, v := range expected.Metadata {
			assert.Contains(t, actual.Metadata, k)
			// Handle potential type conversions (int -> float64, etc.)
			AssertMetadataValueEqual(t, v, actual.Metadata[k])
		}
	}

	// Vector comparison with tolerance
	if len(expected.Vector) > 0 {
		assert.Len(t, actual.Vector, len(expected.Vector))
		for i := range expected.Vector {
			assert.InDelta(t, expected.Vector[i], actual.Vector[i], 0.0001)
		}
	}
}

// AssertMetadataValueEqual handles type-flexible metadata comparison
func AssertMetadataValueEqual(t *testing.T, expected, actual interface{}) {
	// Handle int -> float64 conversions (common in JSON)
	switch e := expected.(type) {
	case int:
		switch a := actual.(type) {
		case float64:
			assert.Equal(t, float64(e), a)
		case int:
			assert.Equal(t, e, a)
		case int64:
			assert.Equal(t, int64(e), a)
		default:
			assert.Equal(t, expected, actual)
		}
	case bool:
		assert.Equal(t, e, actual)
	case string:
		assert.Equal(t, e, actual)
	default:
		assert.Equal(t, expected, actual)
	}
}

// GetTestCollectionName generates unique collection name
func GetTestCollectionName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// SkipIfNoDocker skips test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	// This is a placeholder - actual implementation would check docker ps
	// For now, we assume Docker is available in integration test environment
}

// AssertSearchResults validates search result properties
func AssertSearchResults(t *testing.T, results []*vectordb.SearchResult, minResults int, maxScore float64) {
	assert.GreaterOrEqual(t, len(results), minResults, "Expected at least %d results", minResults)

	for i, result := range results {
		assert.NotNil(t, result.Document, "Result %d should have a document", i)
		assert.NotEmpty(t, result.Document.ID, "Result %d should have a document ID", i)

		if maxScore > 0 {
			assert.LessOrEqual(t, result.Score, maxScore, "Result %d score should be <= %f", i, maxScore)
		}

		// Verify scores are in descending order (best match first)
		if i > 0 {
			assert.GreaterOrEqual(t, results[i-1].Score, result.Score,
				"Results should be sorted by score descending")
		}
	}
}

// CreateBulkTestDocuments creates documents in bulk with progress
func CreateBulkTestDocuments(t *testing.T, client vectordb.VectorDBClient, collectionName string, count int, dims int) []*vectordb.Document {
	ctx := context.Background()
	docs := GenerateTestDocuments(count, dims)

	err := client.CreateDocuments(ctx, collectionName, docs)
	require.NoError(t, err, "Failed to create bulk documents")

	return docs
}

// WaitForIndexing waits for documents to be indexed
func WaitForIndexing(ctx context.Context, client vectordb.VectorDBClient, collectionName string, expectedCount int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("indexing timeout: %w", ctx.Err())
		case <-ticker.C:
			count, err := client.GetCollectionCount(ctx, collectionName)
			if err != nil {
				continue
			}
			if count >= expectedCount {
				// Extra buffer to ensure indexes are built
				time.Sleep(500 * time.Millisecond)
				return nil
			}
		}
	}
}

// AssertCollectionExists checks if collection exists
func AssertCollectionExists(t *testing.T, client vectordb.VectorDBClient, collectionName string) {
	ctx := context.Background()
	exists, err := client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Failed to check collection existence")
	assert.True(t, exists, "Collection %s should exist", collectionName)
}

// AssertCollectionNotExists checks if collection doesn't exist
func AssertCollectionNotExists(t *testing.T, client vectordb.VectorDBClient, collectionName string) {
	ctx := context.Background()
	exists, err := client.CollectionExists(ctx, collectionName)
	require.NoError(t, err, "Failed to check collection existence")
	assert.False(t, exists, "Collection %s should not exist", collectionName)
}

// CleanupCollections removes test collections
func CleanupCollections(ctx context.Context, client vectordb.VectorDBClient, names []string) {
	for _, name := range names {
		_ = client.DeleteCollection(ctx, name)
	}
}
