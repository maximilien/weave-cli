// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"
)

func TestNeo4jCollectionOperations(t *testing.T) {
	// Create client
	config := &neo4j.Config{
		URI:      "bolt://localhost:7687",
		Username: "neo4j",
		Password: "testpassword",
		Database: "neo4j",
	}

	client, err := neo4j.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Neo4j client: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()
	testCollection := "TestCollection"

	// Clean up any existing test collection
	exists, _ := client.CollectionExists(ctx, testCollection)
	if exists {
		client.DeleteCollection(ctx, testCollection, true)
	}

	// Test: Create Collection
	t.Run("CreateCollection", func(t *testing.T) {
		err := client.CreateCollection(ctx, testCollection, 1536, "cosine")
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}
		t.Log("✓ Collection created")
	})

	// Test: Collection Exists
	t.Run("CollectionExists", func(t *testing.T) {
		exists, err := client.CollectionExists(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to check collection existence: %v", err)
		}
		if !exists {
			t.Fatal("Collection should exist but doesn't")
		}
		t.Log("✓ Collection exists")
	})

	// Test: List Collections
	t.Run("ListCollections", func(t *testing.T) {
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Fatalf("Failed to list collections: %v", err)
		}
		found := false
		for _, name := range collections {
			if name == testCollection {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Test collection not found in list: %v", collections)
		}
		t.Logf("✓ Found collection in list (%d total)", len(collections))
	})

	// Test: Get Collection Count (should be 0 initially)
	t.Run("GetCollectionCount", func(t *testing.T) {
		count, err := client.GetCollectionCount(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to get collection count: %v", err)
		}
		if count != 0 {
			t.Fatalf("Expected count 0, got %d", count)
		}
		t.Log("✓ Collection count is 0")
	})

	// Test: Get Collection Info
	t.Run("GetCollectionInfo", func(t *testing.T) {
		info, err := client.GetCollectionInfo(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to get collection info: %v", err)
		}
		if info["type"] != "VECTOR" {
			t.Fatalf("Expected type VECTOR, got %v", info["type"])
		}
		t.Logf("✓ Collection info retrieved: %v", info)
	})

	// Test: Delete Collection
	t.Run("DeleteCollection", func(t *testing.T) {
		err := client.DeleteCollection(ctx, testCollection, true)
		if err != nil {
			t.Fatalf("Failed to delete collection: %v", err)
		}

		// Verify it's deleted
		exists, err := client.CollectionExists(ctx, testCollection)
		if err != nil {
			t.Fatalf("Failed to check collection existence after delete: %v", err)
		}
		if exists {
			t.Fatal("Collection should not exist after deletion")
		}
		t.Log("✓ Collection deleted")
	})
}
