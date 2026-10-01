package weaviate

import (
	"context"
	"testing"
)

func TestWeaveClientDelegationCancellationPaths(t *testing.T) {
	client, err := NewWeaveClient(&Config{URL: "http://127.0.0.1:1", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.ListCollections(ctx); err == nil {
		t.Fatal("expected canceled list error")
	}
	if _, err := client.ListDocuments(ctx, "Docs", 1); err == nil {
		t.Fatal("expected canceled documents error")
	}
	if _, err := client.CountDocuments(ctx, "Docs"); err == nil {
		t.Fatal("expected canceled count error")
	}
	if _, err := client.GetDocument(ctx, "Docs", "id"); err == nil {
		t.Fatal("expected canceled get error")
	}
	if err := client.Health(ctx); err == nil {
		t.Fatal("expected canceled health error")
	}
	if _, err := client.GetCollectionSchema(ctx, "Docs"); err == nil {
		t.Fatal("expected canceled schema error")
	}
}
