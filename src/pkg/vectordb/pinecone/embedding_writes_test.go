// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pinecone

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestEmbeddingBackedWritesPreserveContentAndMetadata(t *testing.T) {
	for _, batch := range []bool{false, true} {
		for _, mode := range []string{"success", "embedding failure", "write failure", "invalid metadata"} {
			t.Run(map[bool]string{false: "single", true: "batch"}[batch]+"/"+mode, func(t *testing.T) {
				adapter, index, inputs := embeddingDataFixture(t, mode)
				doc := &vectordb.Document{ID: "one", Content: "body", Text: "ignored", Image: "https://fixture/image.png", ImageData: "binary-data-must-not-be-stored", URL: "https://fixture/source", Metadata: map[string]interface{}{"owner": "Ada"}}
				if mode == "invalid metadata" {
					doc.Metadata["bad"] = make(chan int)
				}
				docs := []*vectordb.Document{doc}
				if batch {
					docs = append(docs, &vectordb.Document{ID: "two", Text: "fallback text"})
				}
				var err error
				if batch {
					err = adapter.CreateDocuments(context.Background(), "docs", docs)
				} else {
					err = adapter.CreateDocument(context.Background(), "docs", doc)
				}
				wantError := map[string]string{"embedding failure": "failed to generate embedding", "write failure": "failed to upsert", "invalid metadata": "failed to create metadata"}[mode]
				if wantError != "" {
					if err == nil || !strings.Contains(err.Error(), wantError) {
						t.Fatalf("write error=%v want %q", err, wantError)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if index.closed != 1 {
					t.Fatalf("index closes=%d", index.closed)
				}
				if mode == "success" || mode == "write failure" {
					if len(index.upserted) != len(docs) {
						t.Fatalf("upserted vectors=%d want=%d", len(index.upserted), len(docs))
					}
					for i, vector := range index.upserted {
						if vector.Id != docs[i].ID || !reflect.DeepEqual(vector.Values, []float32{0.25, 0.5, 0.75}) {
							t.Fatalf("vector=%+v", vector)
						}
					}
					metadata := index.upserted[0].Metadata.AsMap()
					if metadata["content"] != "body" || metadata["owner"] != "Ada" || metadata["image"] != doc.Image || metadata["url"] != doc.URL {
						t.Fatalf("stored metadata=%v", metadata)
					}
					for _, value := range metadata {
						if value == doc.ImageData {
							t.Fatal("binary image payload leaked into metadata")
						}
					}
					want := []string{"body"}
					if batch {
						want = append(want, "fallback text")
						if index.upserted[1].Metadata.AsMap()["content"] != "fallback text" {
							t.Fatal("text fallback missing")
						}
					}
					if !reflect.DeepEqual(*inputs, want) {
						t.Fatalf("embedding inputs=%v", *inputs)
					}
				} else if len(index.upserted) != 0 {
					t.Fatal("invalid batch was upserted")
				}
			})
		}
	}
}
