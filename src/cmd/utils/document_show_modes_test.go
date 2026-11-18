// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
)

type showModesFactory struct {
	vectordb.ClientFactory
	client *showModesClient
}

func (*showModesFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *showModesFactory) CreateClient(*vectordb.Config) (vectordb.VectorDBClient, error) {
	return f.client, nil
}

type showModesClient struct {
	vectordb.VectorDBClient
	mode      string
	requested []string
	listed    bool
}

func (*showModesClient) GetSchema(context.Context, string) (*vectordb.CollectionSchema, error) {
	return &vectordb.CollectionSchema{Vectorizer: "fixture-model"}, nil
}
func (c *showModesClient) GetDocument(_ context.Context, collection, id string) (*vectordb.Document, error) {
	c.requested = append(c.requested, collection+"/"+id)
	if id == "missing" {
		return nil, errors.New("fixture document missing")
	}
	return &vectordb.Document{ID: id, Content: "first line\nsecond line\nlast line", Metadata: map[string]interface{}{"author": "Ada"}}, nil
}
func (c *showModesClient) ListDocuments(_ context.Context, collection string, limit, offset int) ([]*vectordb.Document, error) {
	if collection != "Docs" || limit != 50 || offset != 0 {
		return nil, errors.New("unexpected pagination")
	}
	c.listed = true
	if c.mode == "list failure" {
		return nil, errors.New("fixture list unavailable")
	}
	if c.mode == "empty" {
		return nil, nil
	}
	docs := []*vectordb.Document{{ID: "first", Content: "first line\nsecond line\nlast line", Metadata: map[string]interface{}{"author": "Ada"}}, {ID: "second", Content: "another document"}}
	if c.mode == "invalid JSON" {
		docs[0].Metadata["invalid"] = make(chan int)
	}
	return docs, nil
}

func TestShowDocumentMultipleResultsAndFailures(t *testing.T) {
	for _, mode := range []string{"text preview", "text full", "JSON", "explicit IDs", "list failure", "empty", "invalid JSON"} {
		t.Run(mode, func(t *testing.T) {
			client := &showModesClient{mode: mode}
			vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, &showModesFactory{client: client})
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeSupabase, supabase.NewFactory()) })
			var ids []string
			if mode == "explicit IDs" {
				ids = []string{"missing", "first", "second"}
			}
			out, diagnostics := captureCollectionListing(t, func() {
				previous := color.Output
				color.Output = os.Stdout
				defer func() { color.Output = previous }()
				ShowDocument(context.Background(), &config.VectorDBConfig{Type: config.VectorDBTypeSupabase}, "Docs", ids, mode == "text full", 1, nil, "", false, false, strings.Contains(mode, "JSON"))
			})
			if client.listed != (len(ids) == 0) {
				t.Fatalf("list called=%v IDs=%v", client.listed, ids)
			}
			if len(ids) > 0 && !reflect.DeepEqual(client.requested, []string{"Docs/missing", "Docs/first", "Docs/second"}) {
				t.Fatalf("requested=%v", client.requested)
			}
			switch mode {
			case "list failure", "empty", "invalid JSON":
				want := map[string]string{"list failure": "Failed to list documents", "empty": "No documents found", "invalid JSON": "Failed to marshal JSON"}[mode]
				if !strings.Contains(out+diagnostics, want) {
					t.Fatalf("missing %q in %s %s", want, out, diagnostics)
				}
			case "JSON":
				var response struct {
					Count      int
					Vectorizer string
					Documents  []vectordb.Document
				}
				if err := json.Unmarshal([]byte(out), &response); err != nil {
					t.Fatal(err)
				}
				if response.Count != 2 || response.Vectorizer != "fixture-model" || len(response.Documents) != 2 || response.Documents[0].ID != "first" || response.Documents[1].ID != "second" || response.Documents[0].Metadata["author"] != "Ada" {
					t.Fatalf("response=%+v", response)
				}
			default:
				for _, want := range []string{"Document ID: first", "Document ID: second", "Collection: Docs", "fixture-model", "author: Ada", strings.Repeat("=", 80)} {
					if !strings.Contains(out, want) {
						t.Fatalf("missing %q in %s", want, out)
					}
				}
				if strings.Contains(out, "last line") != (mode == "text full") {
					t.Fatalf("incorrect truncation: %s", out)
				}
				if mode == "explicit IDs" && !strings.Contains(diagnostics, "fixture document missing") {
					t.Fatalf("missing lookup diagnostic: %s", diagnostics)
				}
			}
		})
	}
}
