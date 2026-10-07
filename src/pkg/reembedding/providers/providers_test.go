package providers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

type providerRoundTripFunc func(*http.Request) (*http.Response, error)

func (f providerRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

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
	if _, err := NewOpenAIProvider("text-embedding"); err == nil {
		t.Fatal("expected OpenAI constructor to require an API key")
	}
	t.Setenv("OPENAI_API_KEY", "fixture")
	if err := openai.IsAvailable(ctx); err != nil {
		t.Fatalf("OpenAI availability with key = %v", err)
	}
	t.Setenv("OPENAI_API_KEY", "")
	if _, err := sentence.GenerateEmbedding(ctx, ""); err == nil {
		t.Fatal("expected empty sentence embedding to fail")
	}
	if _, err := sentence.GenerateEmbeddings(ctx, []string{"test"}); err == nil {
		t.Fatal("expected unavailable sentence-transformers command to fail")
	}
	if err := sentence.IsAvailable(ctx); err == nil {
		t.Fatal("expected sentence-transformers availability check to fail")
	}
	if _, err := CreateProvider(ctx, "openai/text-embedding-3-small"); err == nil {
		t.Fatal("expected OpenAI provider creation to fail without an API key")
	}
	if _, err := CreateProvider(ctx, "ollama/nomic-embed-text"); err == nil {
		t.Fatal("expected unavailable Ollama provider to fail")
	}
}

func TestOpenAIProviderEmbeddingPaths(t *testing.T) {
	client, err := llm.NewOpenAIClientWithHTTP("fixture", &http.Client{Transport: providerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"object":"list","data":[{"object":"embedding","embedding":[0.1,0.2],"index":0}],"model":"fixture","usage":{"prompt_tokens":1,"total_tokens":1}}`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	provider := &OpenAIProvider{modelName: "fixture", client: client}
	one, err := provider.GenerateEmbedding(context.Background(), "hello")
	if err != nil || len(one) != 2 {
		t.Fatalf("GenerateEmbedding() = %#v, %v", one, err)
	}
	many, err := provider.GenerateEmbeddings(context.Background(), []string{"one", "two"})
	if err != nil || len(many) != 2 {
		t.Fatalf("GenerateEmbeddings() = %#v, %v", many, err)
	}
}
