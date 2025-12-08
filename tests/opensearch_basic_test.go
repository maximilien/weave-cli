// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/opensearch"
)

func TestOpenSearchBasicConnectivity(t *testing.T) {
	// Create client with local OpenSearch
	config := &vectordb.Config{
		Type:             vectordb.VectorDBTypeOpenSearchLocal,
		URL:              "http://localhost:9200",
		VectorDimensions: 384,
		SimilarityMetric: "l2",
		Timeout:          30,
	}

	adapter, err := opensearch.NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create OpenSearch adapter: %v", err)
	}
	defer adapter.Close()

	// Test health check
	ctx := context.Background()
	err = adapter.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	t.Log("✓ OpenSearch connectivity test passed")
}
