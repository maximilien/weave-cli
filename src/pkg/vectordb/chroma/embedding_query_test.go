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
	if err := (&Client{}).Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(&Config{}); err == nil {
		t.Fatal("expected local URL validation error")
	}
	if _, err := NewClient(&Config{APIKey: "test-key", Tenant: "tenant", Database: "database"}); err != nil {
		t.Fatal(err)
	}
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

func TestChromaHealthGenericErrorAndSuccess(t *testing.T) {
	failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("backend error"))
	}))
	t.Cleanup(failed.Close)
	client, err := NewClient(&Config{URL: failed.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("expected generic health error")
	}
	success := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"nanosecond heartbeat":1}`))
	}))
	t.Cleanup(success.Close)
	client, err = NewClient(&Config{URL: success.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestChromaCollectionErrorPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend error", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteCollection(context.Background(), "Docs"); err == nil {
		t.Fatal("expected delete collection error")
	}
	if _, err := client.ListCollections(context.Background()); err == nil {
		t.Fatal("expected list collections error")
	}
	if _, err := client.CollectionExists(context.Background(), "Docs"); err == nil {
		t.Fatal("expected collection exists error")
	}
	if _, err := client.GetCollectionCount(context.Background(), "Docs"); err == nil {
		t.Fatal("expected collection count error")
	}
}
