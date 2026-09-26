package providers

import (
	"context"
	"testing"
)

func TestProviderPurePaths(t *testing.T) {
	ctx := context.Background()
	t.Setenv("OPENAI_API_KEY", "")
	if _, err := CreateProvider(ctx, ""); err == nil {
		t.Fatal("expected empty model error")
	}
	if _, err := CreateProvider(ctx, "unknown/model"); err == nil {
		t.Fatal("expected unknown model error")
	}

	sentence, err := NewSentenceTransformersProvider("sentence-transformers/test-model")
	if err != nil {
		t.Fatal(err)
	}
	if sentence.GetModelName() != "sentence-transformers/test-model" || sentence.GetProvider() != "sentence-transformers" {
		t.Fatal("sentence provider metadata mismatch")
	}
	embeddings, err := sentence.GenerateEmbeddings(ctx, nil)
	if err != nil || len(embeddings) != 0 {
		t.Fatalf("empty sentence batch: %#v %v", embeddings, err)
	}
	openai := &OpenAIProvider{modelName: "text-embedding"}
	if openai.GetModelName() != "text-embedding" || openai.GetProvider() != "openai" {
		t.Fatal("openai provider metadata mismatch")
	}
	if err := openai.IsAvailable(ctx); err == nil {
		t.Fatal("expected missing OpenAI key error")
	}
}
