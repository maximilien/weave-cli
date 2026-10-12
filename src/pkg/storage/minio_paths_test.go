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
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := backend.Upload(cancelled, []byte("data"), ImageMetadata{CollectionName: "docs", Filename: "x.png", ContentType: "image/png"}); err == nil {
		t.Fatal("Upload unexpectedly succeeded with cancelled context")
	}
	validURL := "http://127.0.0.1:1/docs/key.png"
	if _, err := backend.Download(cancelled, validURL); err == nil {
		t.Fatal("Download unexpectedly succeeded with cancelled context")
	}
	if err := backend.Delete(cancelled, validURL); err == nil {
		t.Fatal("Delete unexpectedly succeeded with cancelled context")
	}
	if _, err := backend.Exists(cancelled, validURL); err == nil {
		t.Fatal("Exists unexpectedly succeeded with cancelled context")
	}
	if _, err := backend.GetPresignedURL(cancelled, validURL, time.Minute); err == nil {
		t.Fatal("GetPresignedURL unexpectedly succeeded with cancelled context")
	}
	if _, err := backend.GetInfo(cancelled, validURL); err == nil {
		t.Fatal("GetInfo unexpectedly succeeded with cancelled context")
	}
}
