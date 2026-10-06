//go:build (darwin && amd64) || (darwin && arm64)

package chroma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestNoopEmbeddingQuery(t *testing.T) {
	factory := NewFactory()
	if _, err := factory.CreateClient(&vectordb.Config{URL: "http://127.0.0.1:8000"}); err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(&Config{URL: "http://127.0.0.1:1", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("expected Chroma health error")
	}
	fn := &noopEmbeddingFunction{dimensions: 4}
	embedding, err := fn.EmbedQuery(context.Background(), "query")
	if err != nil || embedding.Len() != 4 {
		t.Fatalf("embedding = (%v, %v)", embedding, err)
	}
	if err := fn.EmbedRecords(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestChromaHealthAuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("expected authentication health error")
	}
}
