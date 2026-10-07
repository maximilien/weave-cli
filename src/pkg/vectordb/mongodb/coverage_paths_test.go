// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mongodb

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestClientValidationAndConnectionFailurePaths(t *testing.T) {
	if _, err := NewClient(&Config{Database: "docs"}); err == nil {
		t.Fatal("NewClient accepted a missing URI")
	}
	if _, err := NewClient(&Config{URI: "mongodb://127.0.0.1:1"}); err == nil {
		t.Fatal("NewClient accepted a missing database")
	}
	if _, err := NewClient(&Config{URI: "mongodb://127.0.0.1:1", Database: "docs", Timeout: 1}); err == nil {
		t.Fatal("NewClient unexpectedly connected to an unavailable MongoDB")
	}
	if _, err := NewClient(&Config{URI: "mongodb://[::1", Database: "docs", Timeout: 1}); err == nil {
		t.Fatal("NewClient accepted a malformed URI")
	}
}

func TestAdapterNilCloseAndEmbeddingProviderFailure(t *testing.T) {
	adapter := &Adapter{}
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("Close(nil client) = %v", err)
	}
	if _, err := adapter.createEmbeddingProvider(context.Background(), "unknown-model"); err == nil {
		t.Fatal("createEmbeddingProvider accepted an unknown model")
	}
}

func TestClientHealthAndCloseWithoutMongoServer(t *testing.T) {
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:1").SetServerSelectionTimeout(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{client: mongoClient, config: &Config{Timeout: 1}}
	if err := client.Health(ctx); err == nil {
		t.Fatal("Health unexpectedly succeeded without a MongoDB server")
	}
	if err := client.Close(ctx); err != nil {
		t.Fatalf("Close() = %v", err)
	}
}
