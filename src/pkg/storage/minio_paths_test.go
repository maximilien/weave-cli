package storage

import (
	"context"
	"testing"
	"time"
)

func TestMinioOperationsRejectMalformedURLs(t *testing.T) {
	backend, err := NewMinioStorage(Config{Endpoint: "127.0.0.1:1", AccessKey: "key", SecretKey: "secret", Bucket: "docs"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	badURL := "://malformed"
	if _, err := backend.Download(ctx, badURL); err == nil {
		t.Fatal("Download accepted malformed URL")
	}
	if err := backend.Delete(ctx, badURL); err == nil {
		t.Fatal("Delete accepted malformed URL")
	}
	if _, err := backend.Exists(ctx, badURL); err == nil {
		t.Fatal("Exists accepted malformed URL")
	}
	if _, err := backend.GetPresignedURL(ctx, badURL, time.Minute); err == nil {
		t.Fatal("GetPresignedURL accepted malformed URL")
	}
	if _, err := backend.GetInfo(ctx, badURL); err == nil {
		t.Fatal("GetInfo accepted malformed URL")
	}
}
