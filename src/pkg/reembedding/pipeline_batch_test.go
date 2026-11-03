// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package reembedding

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type batchFixtureProvider struct {
	texts   []string
	vectors [][]float64
	err     error
}

func (p *batchFixtureProvider) GenerateEmbedding(context.Context, string) ([]float64, error) {
	return nil, errors.New("unexpected single request")
}
func (p *batchFixtureProvider) GenerateEmbeddings(_ context.Context, texts []string) ([][]float64, error) {
	p.texts = append([]string(nil), texts...)
	return p.vectors, p.err
}
func (*batchFixtureProvider) GetModelName() string                { return "fixture" }
func (*batchFixtureProvider) GetProvider() string                 { return "fixture" }
func (p *batchFixtureProvider) IsAvailable(context.Context) error { return p.err }

func TestBatchReembeddingUsesTextFallbackAndZeroVectors(t *testing.T) {
	provider := &batchFixtureProvider{vectors: [][]float64{{1, 2}, nil, {3, 4}, nil}}
	pipeline := &EmbeddingPipeline{dimensions: 2, embProvider: provider}
	docs := []*vectordb.Document{{ID: "text", Text: "first"}, nil, {ID: "content", Content: "second"}, {ID: "empty"}}
	if err := pipeline.ProcessBatch(context.Background(), docs); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(provider.texts, []string{"first", "", "second", ""}) {
		t.Fatalf("provider texts: %v", provider.texts)
	}
	for i, want := range map[int][]float64{0: {1, 2}, 2: {3, 4}, 3: {0, 0}} {
		if !reflect.DeepEqual(docs[i].Embedding, want) {
			t.Fatalf("document %d embedding = %v, want %v", i, docs[i].Embedding, want)
		}
	}
}

func TestBatchReembeddingReportsProviderAndDimensionErrors(t *testing.T) {
	provider := &batchFixtureProvider{err: errors.New("fixture failure")}
	pipeline := &EmbeddingPipeline{dimensions: 2, embProvider: provider}
	doc := &vectordb.Document{ID: "fixture", Text: "text"}
	if err := pipeline.ProcessBatch(context.Background(), []*vectordb.Document{doc}); err == nil || !strings.Contains(err.Error(), "fixture failure") {
		t.Fatalf("provider error = %v", err)
	}
	provider.err = nil
	provider.vectors = [][]float64{{1}}
	if err := pipeline.ProcessBatch(context.Background(), []*vectordb.Document{doc}); err == nil || !strings.Contains(err.Error(), "expected 2, got 1") {
		t.Fatalf("dimension error = %v", err)
	}
}
