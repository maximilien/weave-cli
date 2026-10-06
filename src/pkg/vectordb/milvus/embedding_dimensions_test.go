package milvus

import (
	"context"
	"testing"
)

func TestEmbeddingDimensionMappings(t *testing.T) {
	adapter := &Adapter{}
	if _, err := adapter.createEmbeddingProvider(context.Background(), ""); err == nil {
		t.Fatal("expected empty model error")
	}
	if _, err := adapter.createEmbeddingProvider(context.Background(), "unknown-model"); err == nil {
		t.Fatal("expected unknown model error")
	}
	if _, err := NewClient(&Config{}); err == nil {
		t.Fatal("expected missing Milvus address error")
	}
	cases := map[int]string{
		768: "sentence-transformers/all-mpnet-base-v2", 384: "sentence-transformers/all-MiniLM-L6-v2",
		1536: "text-embedding-3-small", 3072: "text-embedding-3-large", 1024: "nomic-embed-text", 999: "text-embedding-3-small",
	}
	for dims, want := range cases {
		if got := inferEmbeddingModelFromDimensions(dims); got != want {
			t.Errorf("dims %d = %q, want %q", dims, got, want)
		}
	}
	models := map[string]int{
		"sentence-transformers/all-mpnet-base-v2": 768, "sentence-transformers/all-MiniLM-L12-v2": 768,
		"sentence-transformers/all-MiniLM-L6-v2": 384, "text-embedding-3-small": 1536,
		"text-embedding-ada-002": 1536, "text-embedding-3-large": 3072, "nomic-embed-text": 1024, "unknown": 1536,
	}
	for model, want := range models {
		if got := getVectorDimensionsFromModel(model); got != want {
			t.Errorf("model %s = %d, want %d", model, got, want)
		}
	}
}
