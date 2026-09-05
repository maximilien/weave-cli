// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package storage

import (
	"path"
	"strings"
	"testing"
	"time"
)

func TestNewMinioStorage(t *testing.T) {
	store, err := NewMinioStorage(Config{
		Endpoint:        "localhost:9000",
		AccessKey:       "access",
		SecretKey:       "secret",
		Region:          "us-west-2",
		Bucket:          "images",
		PathPrefix:      "prefix/",
		PresignDuration: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewMinioStorage() error: %v", err)
	}
	if store.bucket != "images" || store.pathPrefix != "prefix" || store.presignDuration != time.Hour {
		t.Fatalf("unexpected MinIO storage: %#v", store)
	}

	defaults, err := NewMinioStorage(Config{Endpoint: "localhost:9000"})
	if err != nil {
		t.Fatalf("NewMinioStorage(defaults) error: %v", err)
	}
	if defaults.presignDuration != 24*time.Hour {
		t.Fatalf("default presign duration = %s, want 24h", defaults.presignDuration)
	}
	if _, err := NewMinioStorage(Config{Endpoint: "://invalid"}); err == nil {
		t.Fatal("NewMinioStorage() accepted an invalid endpoint")
	}
}

func TestNewS3Storage(t *testing.T) {
	store, err := NewS3Storage(Config{Region: "us-west-2", Bucket: "images"})
	if err != nil {
		t.Fatalf("NewS3Storage() error: %v", err)
	}
	if store.endpoint != "s3.us-west-2.amazonaws.com" || !store.useSSL {
		t.Fatalf("unexpected S3 endpoint/SSL: %q, %v", store.endpoint, store.useSSL)
	}

	custom, err := NewS3Storage(Config{Endpoint: "s3.example.test", Region: "test"})
	if err != nil {
		t.Fatalf("NewS3Storage(custom) error: %v", err)
	}
	if custom.endpoint != "s3.example.test" || !custom.useSSL {
		t.Fatalf("custom S3 endpoint/SSL = %q, %v", custom.endpoint, custom.useSSL)
	}
}

func TestMinioStorageKeyAndURLHelpers(t *testing.T) {
	store := &MinioStorage{bucket: "images", endpoint: "localhost:9000", pathPrefix: "prefix"}
	tests := []struct {
		name        string
		metadata    ImageMetadata
		wantSuffix  string
		wantSegment string
	}{
		{name: "filename extension", metadata: ImageMetadata{Filename: "photo.jpeg", CollectionName: "animals"}, wantSuffix: ".jpeg", wantSegment: "prefix/animals/"},
		{name: "JPEG content type", metadata: ImageMetadata{ContentType: "image/jpeg"}, wantSuffix: ".jpg"},
		{name: "PNG content type", metadata: ImageMetadata{ContentType: "image/png"}, wantSuffix: ".png"},
		{name: "GIF content type", metadata: ImageMetadata{ContentType: "image/gif"}, wantSuffix: ".gif"},
		{name: "WebP content type", metadata: ImageMetadata{ContentType: "image/webp"}, wantSuffix: ".webp"},
		{name: "unknown content type", metadata: ImageMetadata{ContentType: "application/octet-stream"}, wantSuffix: ".bin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := store.generateKey(tt.metadata)
			if path.Ext(key) != tt.wantSuffix {
				t.Fatalf("key %q suffix = %q, want %q", key, path.Ext(key), tt.wantSuffix)
			}
			if tt.wantSegment != "" && !strings.Contains(key, tt.wantSegment) {
				t.Fatalf("key %q does not contain %q", key, tt.wantSegment)
			}
		})
	}

	key := "prefix/animals/2026/09/id.png"
	if got := store.buildURL(key); got != "http://localhost:9000/images/"+key {
		t.Fatalf("HTTP URL = %q", got)
	}
	store.useSSL = true
	if got := store.buildURL(key); got != "https://localhost:9000/images/"+key {
		t.Fatalf("HTTPS URL = %q", got)
	}
	extracted, err := store.extractKey(store.buildURL(key))
	if err != nil || extracted != key {
		t.Fatalf("extractKey() = %q, %v; want %q", extracted, err, key)
	}
	if _, err := store.extractKey("http://example.test/%zz"); err == nil {
		t.Fatal("extractKey() accepted an invalid URL")
	}
}

func TestNewImageStorage(t *testing.T) {
	local, err := NewImageStorage(Config{Type: StorageTypeLocal, LocalPath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewImageStorage(local) error: %v", err)
	}
	if _, ok := local.(*LocalStorage); !ok {
		t.Fatalf("local storage type = %T", local)
	}

	minioStore, err := NewImageStorage(Config{Type: StorageTypeMinio, Endpoint: "localhost:9000"})
	if err != nil {
		t.Fatalf("NewImageStorage(minio) error: %v", err)
	}
	if _, ok := minioStore.(*MinioStorage); !ok {
		t.Fatalf("MinIO storage type = %T", minioStore)
	}

	s3Store, err := NewImageStorage(Config{Type: StorageTypeS3, Region: "us-west-2"})
	if err != nil {
		t.Fatalf("NewImageStorage(S3) error: %v", err)
	}
	if _, ok := s3Store.(*MinioStorage); !ok {
		t.Fatalf("S3 storage type = %T", s3Store)
	}

	if _, err := NewImageStorage(Config{Type: StorageType("ftp")}); err == nil {
		t.Fatal("NewImageStorage() accepted unsupported storage")
	}
}
