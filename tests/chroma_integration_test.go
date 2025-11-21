// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)

func getChromaTestClient(t *testing.T) *chroma.Client {
	// Check for Chroma URL - default to localhost for local testing
	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		chromaURL = "http://localhost:8000"
	}

	config := &chroma.Config{
		URL:              chromaURL,
		APIKey:           os.Getenv("CHROMA_API_KEY"), // Optional for cloud
		Tenant:           "default_tenant",
		Database:         "default_database",
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Timeout:          10,
	}

	client, err := chroma.NewClient(config)
	if err != nil {
		t.Skipf("Failed to create Chroma client (is Chroma running?): %v", err)
	}

	return client
}

func TestChromaHealth(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	t.Log("Health check passed")
}

func TestChromaCollectionOperations(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Test collection name with timestamp to avoid conflicts
	collectionName := fmt.Sprintf("test_collection_%d", time.Now().UnixNano())

	// Clean up after test
	defer func() {
		_ = client.DeleteCollection(ctx, collectionName)
	}()

	// Test CreateCollection
	t.Run("CreateCollection", func(t *testing.T) {
		err := client.CreateCollection(ctx, collectionName, nil)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		t.Logf("Created collection: %s", collectionName)
	})

	// Test CollectionExists
	t.Run("CollectionExists", func(t *testing.T) {
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to check collection exists: %v", err)
		}
		if !exists {
			t.Fatal("Collection should exist")
		}
		t.Log("Collection exists check passed")
	})

	// Test ListCollections
	t.Run("ListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Fatalf("Failed to list collections: %v", err)
		}

		found := false
		for _, col := range collections {
			if col.Name == collectionName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Collection %s not found in list", collectionName)
		}
		t.Logf("Found %d collections, including test collection", len(collections))
	})

	// Test GetCollectionCount
	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to get collection count: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected count 0, got %d", count)
		}
		t.Log("Collection count check passed")
	})

	// Test DeleteCollection
	t.Run("DeleteCollection", func(t *testing.T) {
		err := client.DeleteCollection(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to delete collection: %v", err)
		}

		// Verify deletion
		exists, err := client.CollectionExists(ctx, collectionName)
		if err != nil {
			t.Fatalf("Failed to check collection exists after delete: %v", err)
		}
		if exists {
			t.Fatal("Collection should not exist after deletion")
		}
		t.Log("Collection deleted successfully")
	})
}

func TestChromaCollectionNotFound(t *testing.T) {
	client := getChromaTestClient(t)
	ctx := context.Background()

	// Test getting count for non-existent collection
	_, err := client.GetCollectionCount(ctx, "non_existent_collection_xyz")
	if err == nil {
		t.Fatal("Expected error for non-existent collection")
	}
	t.Logf("Got expected error for non-existent collection: %v", err)
}
