// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImageIngestion_AllVDBs tests image ingestion across all supported vector databases
// Addresses Issue #21: Systematic testing of image ingestion
func TestImageIngestion_AllVDBs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping image ingestion integration tests in short mode")
	}

	testCases := []struct {
		name        string
		vdbType     vectordb.VectorDBType
		skipReason  string
		sizeLimit   int // Max image size in bytes, 0 = no limit
		expectIssue string
	}{
		{
			name:    "Weaviate Local",
			vdbType: vectordb.VectorDBTypeWeaviateLocal,
		},
		{
			name:       "Milvus Local",
			vdbType:    vectordb.VectorDBTypeMilvusLocal,
			sizeLimit:  65536, // 64KB JSON field limit
			skipReason: "Known limitation: 64KB JSON field limit (Issue #19)",
		},
		{
			name:    "Chroma Local",
			vdbType: vectordb.VectorDBTypeChromaLocal,
		},
		{
			name:    "Qdrant Local",
			vdbType: vectordb.VectorDBTypeQdrantLocal,
		},
		{
			name:    "Neo4j Local",
			vdbType: vectordb.VectorDBTypeNeo4jLocal,
		},
		{
			name:    "OpenSearch Local",
			vdbType: vectordb.VectorDBTypeOpenSearchLocal,
		},
		{
			name:    "Supabase",
			vdbType: vectordb.VectorDBTypeSupabase,
		},
		{
			name:    "MongoDB",
			vdbType: vectordb.VectorDBTypeMongoDB,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipReason != "" {
				t.Skip(tc.skipReason)
			}

			// Test with small images
			t.Run("SmallImages", func(t *testing.T) {
				testImageIngestion(t, tc.vdbType, "tests/images", tc.sizeLimit)
			})

			// Test with larger images if no size limit
			if tc.sizeLimit == 0 {
				t.Run("LargeImages", func(t *testing.T) {
					// TODO: Add test with larger images (>64KB) when available
					t.Skip("Large image test data not yet available")
				})
			}
		})
	}
}

func testImageIngestion(t *testing.T, vdbType vectordb.VectorDBType, imageDir string, maxSize int) {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")

	// Find VDB config
	var vdbConfig *config.VectorDBConfig
	for _, db := range cfg.Databases.VectorDatabases {
		if db.Type == string(vdbType) {
			vdbConfig = &db
			break
		}
	}
	require.NotNil(t, vdbConfig, "VDB config not found for %s", vdbType)

	// Create VDB client
	client, err := createVDBClient(vdbConfig)
	require.NoError(t, err, "Failed to create VDB client")
	defer client.Close()

	// Test collection name
	collectionName := "test_images_" + string(vdbType)

	// Clean up before test
	if exists, _ := client.CollectionExists(ctx, collectionName); exists {
		err = client.DeleteCollection(ctx, collectionName)
		require.NoError(t, err, "Failed to delete existing collection")
	}

	// Create collection with image schema
	schema := client.GetDefaultSchema(vectordb.SchemaTypeImage, collectionName)
	err = client.CreateCollection(ctx, collectionName, schema)
	require.NoError(t, err, "Failed to create image collection")

	// Ensure cleanup after test
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	// Find test images
	images, err := filepath.Glob(filepath.Join(imageDir, "*.png"))
	require.NoError(t, err, "Failed to find test images")
	require.NotEmpty(t, images, "No test images found in %s", imageDir)

	// Ingest images
	var ingestedCount int
	var skippedCount int

	for _, imagePath := range images {
		// Check image size
		info, err := os.Stat(imagePath)
		require.NoError(t, err, "Failed to stat image %s", imagePath)

		if maxSize > 0 && info.Size() > int64(maxSize) {
			t.Logf("Skipping image %s (size %d > limit %d)", imagePath, info.Size(), maxSize)
			skippedCount++
			continue
		}

		// Read image
		imageData, err := os.ReadFile(imagePath)
		require.NoError(t, err, "Failed to read image %s", imagePath)

		// Create document with image
		doc := &vectordb.Document{
			Text:    filepath.Base(imagePath),
			Content: filepath.Base(imagePath),
			Metadata: map[string]interface{}{
				"source_file": imagePath,
				"file_size":   info.Size(),
				"image_data":  string(imageData), // Store as base64 in real usage
			},
		}

		// Store document
		_, err = client.CreateDocument(ctx, collectionName, doc)
		if err != nil {
			t.Errorf("Failed to create document for %s: %v", imagePath, err)
			continue
		}

		ingestedCount++
	}

	// Verify ingestion
	t.Logf("Ingested %d/%d images (skipped %d due to size limit)",
		ingestedCount, len(images), skippedCount)

	if maxSize > 0 {
		// With size limit, we expect some to be ingested
		assert.Greater(t, ingestedCount, 0, "Should ingest at least some images")
	} else {
		// Without size limit, all should be ingested
		assert.Equal(t, len(images), ingestedCount, "Should ingest all images")
	}

	// Verify collection count
	info, err := client.GetCollectionInfo(ctx, collectionName)
	require.NoError(t, err, "Failed to get collection info")
	assert.Equal(t, int64(ingestedCount), info.Count,
		"Collection count should match ingested count")
}

// TestImageIngestion_CMYK tests CMYK image handling
// Some VDBs may have issues with CMYK color space
func TestImageIngestion_CMYK(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CMYK image test in short mode")
	}

	// TODO: Add CMYK test image to tests/images/
	t.Skip("CMYK test image not yet available")
}

// TestImageIngestion_LargeImages tests handling of large images
// Addresses Issue #21: Testing size limit handling
func TestImageIngestion_LargeImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large image test in short mode")
	}

	testCases := []struct {
		name      string
		vdbType   vectordb.VectorDBType
		maxSize   int
		expectErr bool
	}{
		{
			name:      "Weaviate - No Limit",
			vdbType:   vectordb.VectorDBTypeWeaviateLocal,
			maxSize:   0,
			expectErr: false,
		},
		{
			name:      "Milvus - 64KB Limit",
			vdbType:   vectordb.VectorDBTypeMilvusLocal,
			maxSize:   65536,
			expectErr: true, // Expect error for >64KB
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: Test with 81KB image from Issue #21
			t.Skip("Large image test data (81KB) not yet available")
		})
	}
}

// Helper function to create VDB client
func createVDBClient(cfg *config.VectorDBConfig) (vectordb.VectorDBClient, error) {
	vdbConfig := &vectordb.Config{
		Type:              vectordb.VectorDBType(cfg.Type),
		URL:               cfg.URL,
		APIKey:            cfg.APIKey,
		Database:          cfg.Database,
		Timeout:           cfg.Timeout,
		VectorDimensions:  cfg.VectorDimensions,
		SimilarityMetric:  cfg.SimilarityMetric,
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		Address:           cfg.Address,
		Username:          cfg.Username,
		Password:          cfg.Password,
		DatabaseURL:       cfg.DatabaseURL,
		DatabaseKey:       cfg.DatabaseKey,
		Host:              cfg.Host,
		Port:              cfg.Port,
	}

	return vectordb.NewVectorDBClient(vdbConfig)
}
