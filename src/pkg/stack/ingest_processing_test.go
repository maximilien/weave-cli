// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
)

type recordingIngestClient struct {
	*ingestFixtureClient
	alreadyExists bool
	documents     []*vectordb.Document
	counted       []string
}

func (c *recordingIngestClient) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := c.ingestFixtureClient.CollectionExists(ctx, name)
	return c.alreadyExists, err
}

func (c *recordingIngestClient) CreateDocuments(_ context.Context, collection string, documents []*vectordb.Document) error {
	if collection != "FixtureDocs" {
		return vectordb.ErrInvalidConfig("unexpected collection: " + collection)
	}
	c.documents = append(c.documents, documents...)
	return nil
}

func (c *recordingIngestClient) GetCollectionCount(_ context.Context, collection string) (int64, error) {
	c.counted = append(c.counted, collection)
	return int64(len(c.documents)), nil
}

func TestIngestToStackProcessesAndVerifiesDocuments(t *testing.T) {
	for _, mode := range []string{"normal", "scheduled", "automatic with failed file"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			source := filepath.Join(root, "source")
			if err := os.Mkdir(source, 0o700); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(source, "notes.txt")
			const content = "Fixture content for stack ingestion."
			if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			embedding := make([]float64, 1536)
			embedding[0] = 0.25
			requests := make(chan struct{}, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer fixture-key" {
					t.Error("missing fixture authentication")
				}
				var request struct {
					Input string `json:"input"`
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode embedding request: %v", err)
				}
				if request.Input != content || request.Model != "text-embedding-3-small" {
					t.Errorf("embedding request = %+v", request)
				}
				select {
				case requests <- struct{}{}:
				default:
					t.Error("unexpected extra embedding request")
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{
					"object": "list", "model": "text-embedding-3-small",
					"data":  []interface{}{map[string]interface{}{"object": "embedding", "index": 0, "embedding": embedding}},
					"usage": map[string]int{"prompt_tokens": 7, "total_tokens": 7},
				}); err != nil {
					t.Errorf("write embedding response: %v", err)
				}
			}))
			t.Cleanup(server.Close)
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
			client := &recordingIngestClient{ingestFixtureClient: &ingestFixtureClient{}, alreadyExists: mode == "normal"}
			factory := &ingestFixtureFactory{Factory: milvus.NewFactory(), client: client}
			vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, factory)
			t.Cleanup(func() {
				vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory())
			})
			cfg := IngestConfig{
				CollectionName: "FixtureDocs", DataPath: source, EmbeddingModel: "text-embedding-3-small",
				MilvusLocalPort: 19531, ParallelWorkers: 1, BatchSize: 1,
				CheckpointFile: filepath.Join(root, "checkpoint.json"),
			}
			if mode == "scheduled" {
				cfg.RestartEvery = 2 // One file completes without a scheduled restart.
			}
			failedFile := filepath.Join(source, "unsupported.fixture")
			if mode == "automatic with failed file" {
				cfg.AutoRestart, cfg.SkipFailed, cfg.CheckpointEvery = true, true, 1
				if err := os.WriteFile(failedFile, []byte("unsupported input"), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			if err := IngestToStack(cfg); err != nil {
				t.Fatalf("IngestToStack() = %v", err)
			}
			if len(requests) != 1 || len(client.documents) != 1 {
				t.Fatalf("embedding requests = %d, stored documents = %d", len(requests), len(client.documents))
			}
			document := client.documents[0]
			if document.ID == "" || document.Content != content || document.Text != content || !reflect.DeepEqual(document.Embedding, embedding) {
				t.Fatalf("stored document content or embedding is incorrect: id=%q content=%q dimensions=%d", document.ID, document.Content, len(document.Embedding))
			}
			if document.Metadata["source_file"] != file || document.Metadata["file_type"] != "txt" || document.Metadata["file_hash"] == "" {
				t.Fatalf("document metadata = %#v", document.Metadata)
			}
			if !slices.Equal(client.counted, []string{cfg.CollectionName}) {
				t.Fatalf("persistence checks = %v", client.counted)
			}
			if mode == "normal" {
				if len(client.created) != 0 {
					t.Fatal("existing collection was recreated")
				}
			} else if !slices.Equal(client.created, []string{cfg.CollectionName}) || client.schema.Vectorizer != cfg.EmbeddingModel {
				t.Fatalf("created collections = %v, schema = %+v", client.created, client.schema)
			}
			if cfg.AutoRestart {
				checkpoint, err := LoadCheckpoint(cfg.CheckpointFile)
				if err != nil || checkpoint == nil {
					t.Fatalf("checkpoint = %+v, %v", checkpoint, err)
				}
				if !slices.Equal(checkpoint.CompletedFiles, []string{file}) || len(checkpoint.PendingFiles) != 0 {
					t.Fatalf("checkpoint progress = %+v", checkpoint)
				}
				if len(checkpoint.FailedFiles) != 1 || checkpoint.FailedFiles[0].File != failedFile || checkpoint.FailedFiles[0].Attempts != 1 {
					t.Fatalf("checkpoint failures = %+v", checkpoint.FailedFiles)
				}
			}
		})
	}
}
