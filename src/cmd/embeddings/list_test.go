// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package embeddings

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

func TestEmbeddingCatalogAndAPIKeys(t *testing.T) {
	models := GetAllEmbeddingModels()
	if len(models) == 0 {
		t.Fatal("GetAllEmbeddingModels() returned no models")
	}

	seen := make(map[string]bool, len(models))
	for _, model := range models {
		if model.Name == "" || model.Provider == "" || model.Type == "" {
			t.Fatalf("incomplete embedding model: %+v", model)
		}
		if seen[model.Name] {
			t.Fatalf("duplicate embedding model %q", model.Name)
		}
		seen[model.Name] = true
	}

	if !seen["text-embedding-3-small"] || !seen["clip-vit-base-patch32"] {
		t.Fatalf("catalog is missing expected models: %#v", seen)
	}

	if !IsAPIKeySet("") {
		t.Fatal("IsAPIKeySet() should accept models without credentials")
	}
	t.Setenv("EMBEDDINGS_TEST_KEY", "")
	if IsAPIKeySet("EMBEDDINGS_TEST_KEY") {
		t.Fatal("IsAPIKeySet() accepted an unset credential")
	}
	t.Setenv("EMBEDDINGS_TEST_KEY", "secret")
	t.Setenv("EMBEDDINGS_SECOND_KEY", "secret")
	if !IsAPIKeySet(" EMBEDDINGS_TEST_KEY, EMBEDDINGS_SECOND_KEY ") {
		t.Fatal("IsAPIKeySet() rejected configured credentials")
	}
}

func TestShowAllEmbeddings(t *testing.T) {
	for _, test := range []struct {
		name          string
		verbose       bool
		database      string
		compatibility bool
	}{
		{name: "default"},
		{name: "filtered verbose", verbose: true, database: "supabase", compatibility: true},
		{name: "unsupported database", database: "not-a-database"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := showAllEmbeddings(test.verbose, test.database, test.compatibility); err != nil {
				t.Fatalf("showAllEmbeddings() error = %v", err)
			}
		})
	}
}

func TestPrintEmbeddingModelVariants(t *testing.T) {
	for _, model := range []EmbeddingModel{
		{Name: "text", Provider: "test", Type: "text"},
		{Name: "image", Provider: "test", Type: "image", Dimensions: 8, APIKeyEnv: "TEST_KEY", Description: "image model", Module: "img2vec-test", SupportedDatabases: []string{"unknown"}},
		{Name: "multimodal", Provider: "test", Type: "multimodal", SupportedDatabases: []string{"mock"}},
	} {
		printEmbeddingModel(model, true, true)
	}
}

func TestShowCollectionEmbeddings(t *testing.T) {
	for _, test := range []struct {
		name      string
		dbType    config.VectorDBType
		wantError bool
	}{
		{name: "Weaviate cloud", dbType: config.VectorDBTypeCloud},
		{name: "Weaviate local", dbType: config.VectorDBTypeLocal},
		{name: "Supabase", dbType: config.VectorDBTypeSupabase},
		{name: "mock", dbType: config.VectorDBTypeMock},
		{name: "known database without models", dbType: config.VectorDBTypeQdrantLocal},
		{name: "unknown database", dbType: config.VectorDBType("unknown"), wantError: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := showCollectionEmbeddings(context.Background(), &config.VectorDBConfig{Type: test.dbType}, "Documents", true)
			if (err != nil) != test.wantError {
				t.Fatalf("showCollectionEmbeddings() error = %v, wantError %v", err, test.wantError)
			}
		})
	}

	// A supported model forces the switch's default branch for an otherwise
	// unknown type, which is distinct from the no-model early return above.
	if err := showCollectionEmbeddings(context.Background(), &config.VectorDBConfig{Type: config.VectorDBTypeMongoDB}, "Documents", false); err == nil {
		t.Fatal("showCollectionEmbeddings() accepted an unsupported database type")
	}
}

func TestRunListWithoutConfiguration(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().StringP("database", "d", "mock", "")
	cmd.Flags().BoolP("show-compatibility", "c", true, "")
	if err := runList(cmd, nil); err != nil {
		t.Fatalf("runList() error = %v", err)
	}
}
