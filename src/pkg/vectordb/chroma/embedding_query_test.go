//go:build (darwin && amd64) || (darwin && arm64)

package chroma

import (
	"context"
	"testing"
)

func TestNoopEmbeddingQuery(t *testing.T) {
	fn := &noopEmbeddingFunction{dimensions: 4}
	embedding, err := fn.EmbedQuery(context.Background(), "query")
	if err != nil || embedding.Len() != 4 {
		t.Fatalf("embedding = (%v, %v)", embedding, err)
	}
	if err := fn.EmbedRecords(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
}
