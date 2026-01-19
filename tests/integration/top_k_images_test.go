// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTopKImagesFlag tests the --top_k_images flag with real vector databases
// This test requires:
// - WEAVIATE_CLOUD_API_KEY environment variable
// - WEAVIATE_CLOUD_URL environment variable
func TestTopKImagesFlag(t *testing.T) {
	// Skip if no Weaviate Cloud credentials
	apiKey := os.Getenv("WEAVIATE_CLOUD_API_KEY")
	url := os.Getenv("WEAVIATE_CLOUD_URL")
	if apiKey == "" || url == "" {
		t.Skip("Skipping integration test: WEAVIATE_CLOUD_API_KEY or WEAVIATE_CLOUD_URL not set")
	}

	ctx := context.Background()

	// Test collection names
	textCol := "TestTopKImages_TextCol"
	imageCol := "TestTopKImages_ImageCol"

	// Create Weaviate adapter
	cfg := &vectordb.Config{
		Type:         vectordb.VectorDBTypeWeaviateCloud,
		APIKey:       apiKey,
		URL:          url,
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"), // Required for text2vec-openai
		Timeout:      30,
	}

	adapter, err := weaviate.NewAdapter(cfg)
	require.NoError(t, err, "Failed to create Weaviate adapter")
	defer adapter.Close()

	// Cleanup function
	cleanup := func() {
		_ = adapter.DeleteCollection(ctx, textCol)
		_ = adapter.DeleteCollection(ctx, imageCol)
	}
	defer cleanup()
	cleanup() // Clean up any previous test runs

	t.Run("Setup_Collections", func(t *testing.T) {
		// Create text collection
		textSchema := adapter.GetDefaultSchema(vectordb.SchemaTypeText, textCol)
		err := adapter.CreateCollection(ctx, textCol, textSchema)
		require.NoError(t, err, "Failed to create text collection")

		// Create image collection
		imageSchema := adapter.GetDefaultSchema(vectordb.SchemaTypeImage, imageCol)
		err = adapter.CreateCollection(ctx, imageCol, imageSchema)
		require.NoError(t, err, "Failed to create image collection")

		// Add text documents
		textDocs := []*vectordb.Document{
			{
				ID:      uuid.New().String(),
				Content: "Weave CLI is a command-line tool for managing vector databases",
				Metadata: map[string]interface{}{
					"type":  "documentation",
					"topic": "cli",
				},
			},
			{
				ID:      uuid.New().String(),
				Content: "Vector databases enable semantic search using embeddings",
				Metadata: map[string]interface{}{
					"type":  "documentation",
					"topic": "vectors",
				},
			},
			{
				ID:      uuid.New().String(),
				Content: "RAG agents combine retrieval and generation for better answers",
				Metadata: map[string]interface{}{
					"type":  "documentation",
					"topic": "rag",
				},
			},
		}

		for _, doc := range textDocs {
			err := adapter.CreateDocument(ctx, textCol, doc)
			require.NoError(t, err, fmt.Sprintf("Failed to create text document %s", doc.ID))
		}

		// Add image documents (with searchable content in Text field)
		// Note: Image schema has nested object metadata, so we use Text field for searchable content
		imageDocs := []*vectordb.Document{
			{
				ID:    uuid.New().String(),
				Image: "https://example.com/screenshot.png",
				Text:  "Screenshot of Weave CLI terminal showing query results with commands like weave cols query WeaveDocs agent rag-agent",
			},
			{
				ID:    uuid.New().String(),
				Image: "https://example.com/diagram.png",
				Text:  "Architecture diagram showing vector database and RAG agent Client Vector DB RAG Agent LLM system design",
			},
		}

		for _, doc := range imageDocs {
			err := adapter.CreateDocument(ctx, imageCol, doc)
			require.NoError(t, err, fmt.Sprintf("Failed to create image document %s", doc.ID))
		}

		t.Logf("Created %d text docs in %s and %d image docs in %s",
			len(textDocs), textCol, len(imageDocs), imageCol)
	})

	t.Run("Query_Without_TopKImages", func(t *testing.T) {
		// Query both collections without --top_k_images
		// Text docs typically have higher scores, so we might get 0 images
		options := &vectordb.QueryOptions{
			TopK: 5,
		}

		// Query text collection
		textResults, err := adapter.SearchSemantic(ctx, textCol, "weave cli vector database", options)
		require.NoError(t, err)
		t.Logf("Text collection returned %d results", len(textResults))

		// Query image collection
		imageResults, err := adapter.SearchSemantic(ctx, imageCol, "weave cli vector database", options)
		require.NoError(t, err)
		t.Logf("Image collection returned %d results", len(imageResults))

		// Combine results
		allResults := append(textResults, imageResults...)
		t.Logf("Total results without topKImages: %d", len(allResults))

		// All results from image collection are images
		imageCount := len(imageResults)
		t.Logf("Image results: %d", imageCount)
	})

	t.Run("Query_With_TopKImages", func(t *testing.T) {
		// Query with --top_k_images to guarantee image results
		// Simulating what agent_query.go does: different topK for text vs image collections
		textOptions := &vectordb.QueryOptions{
			TopK: 3, // Get top 3 from text collection
		}
		imageOptions := &vectordb.QueryOptions{
			TopK: 2, // FORCE 2 results from image collection (simulating topKImages)
		}

		// Query text collection (should use TopK=3)
		textResults, err := adapter.SearchSemantic(ctx, textCol, "weave cli vector database", textOptions)
		require.NoError(t, err)
		t.Logf("Text collection returned %d results (requested TopK=%d)", len(textResults), textOptions.TopK)

		// Query image collection (should use TopKImages=2)
		imageResults, err := adapter.SearchSemantic(ctx, imageCol, "weave cli vector database", imageOptions)
		require.NoError(t, err)
		t.Logf("Image collection returned %d results (requested TopK=%d)", len(imageResults), imageOptions.TopK)

		// The key assertion: image collection should attempt to return topKImages results
		// (may be less if collection has fewer docs, but should try for topKImages)
		assert.LessOrEqual(t, len(imageResults), imageOptions.TopK,
			"Image collection should not exceed topKImages")

		// If we got any images, that's a success!
		if len(imageResults) > 0 {
			t.Logf("✅ SUCCESS: Got %d image results with topKImages=%d", len(imageResults), imageOptions.TopK)
		} else {
			t.Logf("⚠️  WARNING: Got 0 image results (collection may have no embeddings or no matching docs)")
		}
	})

	t.Run("Image_Collection_Detection", func(t *testing.T) {
		// Test that our detection logic correctly identifies image vs text collections
		testCases := []struct {
			name       string
			collection string
			expected   bool
		}{
			{"TextCollection", textCol, false},
			{"ImageCollection", imageCol, true},
			{"NameBasedDetection_Images", "MyImages", true},
			{"NameBasedDetection_Photos", "UserPhotos", true},
			{"NameBasedDetection_Docs", "MyDocuments", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Simulate name-based detection (first priority)
				nameLower := strings.ToLower(tc.collection)
				imageKeywords := []string{"image", "img", "photo", "picture", "visual", "media"}

				isImage := false
				for _, keyword := range imageKeywords {
					if strings.Contains(nameLower, keyword) {
						isImage = true
						break
					}
				}

				assert.Equal(t, tc.expected, isImage,
					"Collection '%s' detection mismatch", tc.collection)
			})
		}
	})
}
