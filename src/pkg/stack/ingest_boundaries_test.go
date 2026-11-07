// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
)

type ingestFixtureFactory struct {
	*milvus.Factory
	client vectordb.VectorDBClient
	err    error
	config *vectordb.Config
}

func (f *ingestFixtureFactory) CreateClient(config *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.config = config
	return f.client, f.err
}

type ingestFixtureClient struct {
	vectordb.VectorDBClient
	existsErr error
	createErr error
	checks    []string
	created   []string
	schema    *vectordb.CollectionSchema
}

func (c *ingestFixtureClient) CollectionExists(_ context.Context, name string) (bool, error) {
	c.checks = append(c.checks, name)
	return false, c.existsErr
}

func (c *ingestFixtureClient) CreateCollection(_ context.Context, name string, schema *vectordb.CollectionSchema) error {
	c.created = append(c.created, name)
	c.schema = schema
	return c.createErr
}

func TestIngestToStackValidatesInputsAndResumesBeforeCreatingCollections(t *testing.T) {
	for _, tt := range []struct {
		name        string
		wantError   string
		wantChecks  int
		wantCreates int
	}{
		{name: "client unavailable", wantError: "failed to create VDB client: failed to create vector database client: fixture unavailable"},
		{name: "missing API key", wantError: "OPENAI_API_KEY environment variable is required"},
		{name: "missing source", wantError: "failed to scan files:"},
		{name: "empty source", wantError: "no files found in:"},
		{name: "corrupt checkpoint", wantError: "failed to load checkpoint: failed to parse checkpoint:"},
		{name: "missing resume file", wantError: "failed to filter files: resume file not found: absent.txt"},
		{name: "completed checkpoint"},
		{name: "collection lookup failed", wantError: "failed to check collection existence: fixture lookup failed", wantChecks: 1},
		{name: "collection creation failed", wantError: "failed to create collection: fixture creation failed", wantChecks: 1, wantCreates: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			t.Setenv("OPENAI_BASE_URL", "http://127.0.0.1:1/v1/")
			source := filepath.Join(root, "source")
			if err := os.Mkdir(source, 0o700); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(source, "notes.txt")
			if tt.name != "empty source" {
				if err := os.WriteFile(file, []byte("fixture source text"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := IngestConfig{
				CollectionName: "FixtureDocs", DataPath: source,
				EmbeddingModel:  "sentence-transformers/all-MiniLM-L6-v2",
				MilvusLocalPort: 19531, BatchSize: 10, ParallelWorkers: 1,
				CheckpointFile: filepath.Join(root, "checkpoint.json"),
			}
			client := &ingestFixtureClient{}
			factory := &ingestFixtureFactory{Factory: milvus.NewFactory(), client: client}
			// These tests run serially and restore the registered production factory.
			vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, factory)
			t.Cleanup(func() {
				vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory())
			})
			switch tt.name {
			case "client unavailable":
				factory.err = errors.New("fixture unavailable")
			case "missing API key":
				t.Setenv("OPENAI_API_KEY", "")
			case "missing source":
				cfg.DataPath = filepath.Join(root, "absent")
			case "corrupt checkpoint":
				cfg.CheckpointEvery = 1
				if err := os.WriteFile(cfg.CheckpointFile, []byte("invalid JSON"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "missing resume file":
				cfg.ResumeFrom = "absent.txt"
			case "completed checkpoint":
				cfg.CheckpointEvery = 1
				checkpoint := InitCheckpoint(cfg.CollectionName, []string{file}, nil)
				checkpoint.MarkFileCompleted(file)
				checkpoint.MarkFileFailed("previous.txt", 1, "previous failure")
				if err := SaveCheckpoint(checkpoint, cfg.CheckpointFile); err != nil {
					t.Fatal(err)
				}
			case "collection lookup failed":
				client.existsErr = errors.New("fixture lookup failed")
			case "collection creation failed":
				client.createErr = errors.New("fixture creation failed")
			}

			err := IngestToStack(cfg)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("IngestToStack() = %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("IngestToStack() = %v, want %q", err, tt.wantError)
			}
			if factory.config == nil || factory.config.Address != "localhost:19531" || factory.config.VectorDimensions != 384 || factory.config.Timeout != 30 {
				t.Fatalf("database configuration = %+v", factory.config)
			}
			if len(client.checks) != tt.wantChecks || len(client.created) != tt.wantCreates {
				t.Fatalf("collection lookups = %v, creations = %v; want %d, %d", client.checks, client.created, tt.wantChecks, tt.wantCreates)
			}
			for _, name := range append(client.checks, client.created...) {
				if name != cfg.CollectionName {
					t.Fatalf("database operation targeted %q, want %q", name, cfg.CollectionName)
				}
			}
			if tt.wantCreates > 0 && (client.schema.Class != cfg.CollectionName || client.schema.Vectorizer != cfg.EmbeddingModel) {
				t.Fatalf("collection schema = %+v", client.schema)
			}
		})
	}
}
