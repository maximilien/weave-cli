// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/mock"
	"github.com/spf13/pflag"
)

type reembedFactory struct {
	vectordb.ClientFactory
	client     *reembedClient
	dimensions []int
}

func (f *reembedFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *reembedFactory) CreateClient(cfg *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.dimensions = append(f.dimensions, cfg.VectorDimensions)
	return f.client, nil
}

type reembedClient struct {
	vectordb.VectorDBClient
	mode    string
	reads   int
	created []string
	schema  *vectordb.CollectionSchema
	stored  []*vectordb.Document
}

func (c *reembedClient) CollectionExists(_ context.Context, name string) (bool, error) {
	return name == "Docs", nil
}
func (c *reembedClient) GetCollectionCount(_ context.Context, name string) (int64, error) {
	if name == "Docs" {
		return 1, nil
	}
	return int64(len(c.stored)), nil
}
func (c *reembedClient) GetDefaultSchema(_ vectordb.SchemaType, name string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{Class: name}
}
func (c *reembedClient) CreateCollection(_ context.Context, name string, schema *vectordb.CollectionSchema) error {
	c.created = append(c.created, name)
	c.schema = schema
	return nil
}
func (c *reembedClient) ListDocuments(_ context.Context, _ string, limit, offset int) ([]*vectordb.Document, error) {
	c.reads++
	if c.mode == "read retry" && c.reads == 1 {
		return nil, errors.New("fixture read failure")
	}
	if c.mode == "empty batch" {
		return nil, nil
	}
	if limit != 1 || offset != 0 {
		return nil, errors.New("unexpected pagination")
	}
	return []*vectordb.Document{{ID: "source-1", Content: "Fixture text", Metadata: map[string]interface{}{"owner": "Ada"}}}, nil
}
func (c *reembedClient) CreateDocuments(_ context.Context, _ string, docs []*vectordb.Document) error {
	if c.mode == "write failure" {
		return errors.New("fixture write failure")
	}
	c.stored = append(c.stored, docs...)
	return nil
}

func TestReembedCommandProcessesBatchesAndFailures(t *testing.T) {
	for _, mode := range []string{"success", "read retry", "empty batch", "embedding failure", "write failure"} {
		t.Run(mode, func(t *testing.T) {
			setupMockCollectionConfig(t)
			ReEmbedCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				value, changed := flag.Value.String(), flag.Changed
				t.Cleanup(func() { _ = flag.Value.Set(value); flag.Changed = changed })
			})
			for name, value := range map[string]string{"new-embedding": "text-embedding-3-small", "output": "Output", "batch-size": "1", "skip-existing": "false"} {
				if err := ReEmbedCmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
			}
			client := &reembedClient{mode: mode}
			factory := &reembedFactory{client: client}
			vectordb.RegisterFactory(vectordb.VectorDBTypeMock, factory)
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeMock, mock.NewFactory()) })
			embedding := make([]float64, 1536)
			embedding[0] = 0.25
			requests := make(chan struct{}, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var request struct {
					Input string `json:"input"`
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				if request.Input != "Fixture text" || request.Model != "text-embedding-3-small" {
					t.Errorf("embedding request=%+v", request)
				}
				requests <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				if mode == "embedding failure" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"message": "fixture embedding failed", "type": "invalid_request_error"}})
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "model": "text-embedding-3-small", "data": []interface{}{map[string]interface{}{"object": "embedding", "index": 0, "embedding": embedding}}})
			}))
			t.Cleanup(server.Close)
			t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
			runReEmbed(ReEmbedCmd, []string{"Docs"})
			if len(factory.dimensions) != 2 || factory.dimensions[1] != 1536 {
				t.Fatalf("client dimensions=%v", factory.dimensions)
			}
			if !reflect.DeepEqual(client.created, []string{"Output"}) || client.schema.Class != "Output" || client.schema.Vectorizer != "text-embedding-3-small" {
				t.Fatalf("output creation=%v schema=%+v", client.created, client.schema)
			}
			wantCalls := 1
			if mode == "empty batch" {
				wantCalls = 0
			}
			if len(requests) != wantCalls {
				t.Fatalf("embedding calls=%d want=%d", len(requests), wantCalls)
			}
			wantStored := 0
			if mode == "success" || mode == "read retry" {
				wantStored = 1
			}
			if len(client.stored) != wantStored {
				t.Fatalf("stored documents=%d want=%d", len(client.stored), wantStored)
			}
			if wantStored == 1 {
				doc := client.stored[0]
				if doc.ID != "source-1" || doc.Metadata["owner"] != "Ada" || !reflect.DeepEqual(doc.Embedding, embedding) {
					t.Fatalf("stored document id=%q metadata=%v vector size=%d", doc.ID, doc.Metadata, len(doc.Embedding))
				}
			}
			if mode == "read retry" && client.reads != 2 {
				t.Fatalf("read attempts=%d", client.reads)
			}
		})
	}
}
