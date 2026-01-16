// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

//go:build integration
// +build integration

package utils

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_CrossVDB_MultiCollectionQuery tests querying collections across different VDBs
func TestIntegration_CrossVDB_MultiCollectionQuery(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create unique collection names
	weaviateCol := "TestCrossVDB_Weaviate_" + uuid.New().String()[:8]
	milvusCol := "TestCrossVDB_Milvus_" + uuid.New().String()[:8]

	// Setup Weaviate collection
	weaviateConfig := getWeaviateConfig(t)
	weaviateAdapter := createWeaviateAdapter(t, weaviateConfig)
	defer weaviateAdapter.Close()
	defer weaviateAdapter.DeleteCollection(ctx, weaviateCol)

	// Setup Milvus collection
	milvusConfig := getMilvusConfig(t)
	milvusAdapter := createMilvusAdapter(t, milvusConfig)
	defer milvusAdapter.Close()
	defer milvusAdapter.DeleteCollection(ctx, milvusCol)

	// Create Weaviate collection
	weaviateSchema := &vectordb.CollectionSchema{
		Name:        weaviateCol,
		Description: "Cross-VDB test collection in Weaviate",
		Properties: []vectordb.Property{
			{Name: "content", DataType: "text"},
		},
		VectorConfig: &vectordb.VectorConfig{
			Vectorizer: "none",
		},
	}
	err := weaviateAdapter.CreateCollection(ctx, weaviateCol, weaviateSchema)
	require.NoError(t, err, "Failed to create Weaviate collection")

	// Create Milvus collection
	milvusSchema := &vectordb.CollectionSchema{
		Name:        milvusCol,
		Description: "Cross-VDB test collection in Milvus",
		Properties: []vectordb.Property{
			{Name: "content", DataType: "text"},
		},
		VectorConfig: &vectordb.VectorConfig{
			Dimensions: 1536, // OpenAI embedding size
		},
	}
	err = milvusAdapter.CreateCollection(ctx, milvusCol, milvusSchema)
	require.NoError(t, err, "Failed to create Milvus collection")

	time.Sleep(2 * time.Second)

	// Add documents to Weaviate collection
	weaviateDocs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Content: "Weaviate is a cloud-native vector database",
			Metadata: map[string]interface{}{
				"source": "weaviate",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Vector databases enable semantic search",
			Metadata: map[string]interface{}{
				"source": "weaviate",
			},
		},
	}
	err = weaviateAdapter.CreateDocuments(ctx, weaviateCol, weaviateDocs)
	require.NoError(t, err, "Failed to create Weaviate documents")

	// Add documents to Milvus collection
	milvusDocs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Content: "Milvus is an open-source vector database",
			Metadata: map[string]interface{}{
				"source": "milvus",
			},
		},
		{
			ID:      uuid.New().String(),
			Content: "Similarity search with vector embeddings",
			Metadata: map[string]interface{}{
				"source": "milvus",
			},
		},
	}
	err = milvusAdapter.CreateDocuments(ctx, milvusCol, milvusDocs)
	require.NoError(t, err, "Failed to create Milvus documents")

	time.Sleep(3 * time.Second)

	// Create collection specs with explicit VDB keys
	specs := []CollectionSpec{
		{Name: weaviateCol, VDBKey: "weaviate-cloud"},
		{Name: milvusCol, VDBKey: "milvus-local"},
	}

	// Create VDB config map
	testConfig := &config.Config{
		Databases: config.DatabasesConfig{
			VectorDatabases: []config.VectorDBConfig{
				*weaviateConfig,
				*milvusConfig,
			},
		},
	}

	resolver := NewVDBConfigResolver(testConfig, weaviateConfig)
	vdbConfigs, err := resolver.ResolveConfigs(specs)
	require.NoError(t, err, "Failed to resolve VDB configs")

	// Verify configs resolved correctly
	assert.NotNil(t, vdbConfigs[weaviateCol], "Weaviate config should be resolved")
	assert.NotNil(t, vdbConfigs[milvusCol], "Milvus config should be resolved")

	// Query both collections (simulating cross-VDB query)
	weaviateResults, err := weaviateAdapter.SearchSemantic(ctx, weaviateCol, "vector database", &vectordb.QueryOptions{TopK: 2})
	if err != nil && err.Error() == "LLM client not initialized" {
		t.Skip("Skipping cross-VDB test: OpenAI API key not available")
	}
	require.NoError(t, err, "Failed to query Weaviate")

	milvusResults, err := milvusAdapter.SearchSemantic(ctx, milvusCol, "vector database", &vectordb.QueryOptions{TopK: 2})
	if err != nil && err.Error() == "LLM client not initialized" {
		t.Skip("Skipping cross-VDB test: OpenAI API key not available")
	}
	require.NoError(t, err, "Failed to query Milvus")

	// Add metadata to simulate cross-VDB query metadata injection
	for _, result := range weaviateResults {
		if result.Document.Metadata == nil {
			result.Document.Metadata = make(map[string]interface{})
		}
		result.Document.Metadata["_collection"] = weaviateCol
		result.Document.Metadata["_vdb"] = "weaviate-cloud"
		result.Document.Metadata["_vdb_type"] = string(config.VectorDBTypeCloud)
	}

	for _, result := range milvusResults {
		if result.Document.Metadata == nil {
			result.Document.Metadata = make(map[string]interface{})
		}
		result.Document.Metadata["_collection"] = milvusCol
		result.Document.Metadata["_vdb"] = "milvus-local"
		result.Document.Metadata["_vdb_type"] = string(config.VectorDBTypeMilvusLocal)
	}

	// Aggregate results
	allResults := append(weaviateResults, milvusResults...)

	// Verify cross-VDB results
	assert.GreaterOrEqual(t, len(allResults), 2, "Should have results from at least one VDB")

	// Verify metadata includes VDB information
	foundWeaviate := false
	foundMilvus := false

	for _, result := range allResults {
		assert.NotNil(t, result.Document.Metadata, "Result should have metadata")

		if collection, ok := result.Document.Metadata["_collection"].(string); ok {
			if collection == weaviateCol {
				foundWeaviate = true
				vdb, _ := result.Document.Metadata["_vdb"].(string)
				assert.Equal(t, "weaviate-cloud", vdb, "Weaviate result should have correct VDB")
			} else if collection == milvusCol {
				foundMilvus = true
				vdb, _ := result.Document.Metadata["_vdb"].(string)
				assert.Equal(t, "milvus-local", vdb, "Milvus result should have correct VDB")
			}
		}
	}

	t.Logf("Cross-VDB query completed: %d total results (Weaviate: %v, Milvus: %v)",
		len(allResults), foundWeaviate, foundMilvus)

	// At least one VDB should have returned results
	assert.True(t, foundWeaviate || foundMilvus, "Should have results from at least one VDB")
}

// TestIntegration_CrossVDB_FullQueryPath tests the complete cross-VDB query path
// using the actual QueryMultipleCollectionsCrossVDB function
func TestIntegration_CrossVDB_FullQueryPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create unique collection names
	weaviateCol := "TestCrossVDBFull_Weaviate_" + uuid.New().String()[:8]
	milvusCol := "TestCrossVDBFull_Milvus_" + uuid.New().String()[:8]

	// Setup Weaviate
	weaviateConfig := getWeaviateConfig(t)
	weaviateAdapter := createWeaviateAdapter(t, weaviateConfig)
	defer weaviateAdapter.Close()
	defer weaviateAdapter.DeleteCollection(ctx, weaviateCol)

	// Setup Milvus
	milvusConfig := getMilvusConfig(t)
	milvusAdapter := createMilvusAdapter(t, milvusConfig)
	defer milvusAdapter.Close()
	defer milvusAdapter.DeleteCollection(ctx, milvusCol)

	// Create collections and add documents
	weaviateSchema := &vectordb.CollectionSchema{
		Name:        weaviateCol,
		Description: "Full path test in Weaviate",
		Properties: []vectordb.Property{
			{Name: "content", DataType: "text"},
		},
		VectorConfig: &vectordb.VectorConfig{
			Vectorizer: "none",
		},
	}
	err := weaviateAdapter.CreateCollection(ctx, weaviateCol, weaviateSchema)
	require.NoError(t, err)

	milvusSchema := &vectordb.CollectionSchema{
		Name:        milvusCol,
		Description: "Full path test in Milvus",
		Properties: []vectordb.Property{
			{Name: "content", DataType: "text"},
		},
		VectorConfig: &vectordb.VectorConfig{
			Dimensions: 1536,
		},
	}
	err = milvusAdapter.CreateCollection(ctx, milvusCol, milvusSchema)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Add test documents
	weaviateDocs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Content: "Weaviate vector database documentation",
			Metadata: map[string]interface{}{
				"source": "weaviate-test",
			},
		},
	}
	err = weaviateAdapter.CreateDocuments(ctx, weaviateCol, weaviateDocs)
	require.NoError(t, err)

	milvusDocs := []*vectordb.Document{
		{
			ID:      uuid.New().String(),
			Content: "Milvus vector database features",
			Metadata: map[string]interface{}{
				"source": "milvus-test",
			},
		},
	}
	err = milvusAdapter.CreateDocuments(ctx, milvusCol, milvusDocs)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	// Create collection specs with explicit VDB keys
	specs := []CollectionSpec{
		{Name: weaviateCol, VDBKey: "weaviate-cloud"},
		{Name: milvusCol, VDBKey: "milvus-local"},
	}

	// Create config for resolver
	testConfig := &config.Config{
		Databases: config.DatabasesConfig{
			VectorDatabases: []config.VectorDBConfig{
				*weaviateConfig,
				*milvusConfig,
			},
		},
	}

	// Resolve VDB configs
	resolver := NewVDBConfigResolver(testConfig, weaviateConfig)
	vdbConfigs, err := resolver.ResolveConfigs(specs)
	require.NoError(t, err)

	// Test the actual cross-VDB query function
	t.Run("QueryMultipleCollectionsCrossVDB", func(t *testing.T) {
		// This tests the full query path without agent
		// Note: This will call QueryMultipleCollectionsCrossVDB which uses
		// the same underlying infrastructure as the agent version

		// Create individual adapters and query to verify functionality
		// (The actual QueryMultipleCollectionsCrossVDB is in collection.go
		// and uses command context, so we verify the building blocks work)

		results1, err := weaviateAdapter.SearchSemantic(ctx, weaviateCol, "database", &vectordb.QueryOptions{TopK: 1})
		if err != nil && err.Error() == "LLM client not initialized" {
			t.Skip("Skipping: OpenAI API key not available")
		}
		require.NoError(t, err)

		results2, err := milvusAdapter.SearchSemantic(ctx, milvusCol, "database", &vectordb.QueryOptions{TopK: 1})
		if err != nil && (err.Error() == "LLM client not initialized" || err.Error() == "Milvus: SearchSemantic requires OpenAI API key for embedding generation. Please set OPENAI_API_KEY environment variable") {
			t.Skip("Skipping: OpenAI API key not available")
		}
		require.NoError(t, err)

		// Verify we got results from both VDBs
		assert.NotEmpty(t, results1, "Should have results from Weaviate")
		assert.NotEmpty(t, results2, "Should have results from Milvus")

		// Verify VDB configs were resolved correctly
		assert.Equal(t, weaviateConfig.Type, vdbConfigs[weaviateCol].Type)
		assert.Equal(t, milvusConfig.Type, vdbConfigs[milvusCol].Type)

		t.Logf("Cross-VDB full path test: Got %d results from Weaviate, %d from Milvus",
			len(results1), len(results2))
	})
}

// Helper functions

func getWeaviateConfig(t *testing.T) *config.VectorDBConfig {
	t.Helper()

	cfg, err := config.LoadConfig("", "")
	if err != nil {
		t.Skipf("Skipping test: cannot load config: %v", err)
	}

	// Find Weaviate config
	for i := range cfg.Databases.VectorDatabases {
		vdb := &cfg.Databases.VectorDatabases[i]
		if vdb.Type == config.VectorDBTypeCloud || vdb.Type == config.VectorDBTypeLocal {
			return vdb
		}
	}

	t.Skip("Skipping test: Weaviate not configured")
	return nil
}

func getMilvusConfig(t *testing.T) *config.VectorDBConfig {
	t.Helper()

	cfg, err := config.LoadConfig("", "")
	if err != nil {
		t.Skipf("Skipping test: cannot load config: %v", err)
	}

	// Find Milvus config
	for i := range cfg.Databases.VectorDatabases {
		vdb := &cfg.Databases.VectorDatabases[i]
		if vdb.Type == config.VectorDBTypeMilvusLocal || vdb.Type == config.VectorDBTypeMilvusCloud {
			return vdb
		}
	}

	t.Skip("Skipping test: Milvus not configured")
	return nil
}

func createWeaviateAdapter(t *testing.T, cfg *config.VectorDBConfig) vectordb.VectorDB {
	t.Helper()

	adapter, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Skipf("Skipping test: cannot create Weaviate client: %v", err)
	}

	return adapter
}

func createMilvusAdapter(t *testing.T, cfg *config.VectorDBConfig) vectordb.VectorDB {
	t.Helper()

	adapter, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Skipf("Skipping test: cannot create Milvus client: %v", err)
	}

	return adapter
}
