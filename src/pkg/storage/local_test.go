// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLocalStorage(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "images")
	store, err := NewLocalStorage(Config{LocalPath: basePath})
	if err != nil {
		t.Fatalf("NewLocalStorage() error: %v", err)
	}
	if store.basePath != basePath {
		t.Fatalf("base path = %q, want %q", store.basePath, basePath)
	}
	if store.presignDuration != 24*time.Hour {
		t.Fatalf("presign duration = %s, want 24h", store.presignDuration)
	}
	if info, err := os.Stat(basePath); err != nil || !info.IsDir() {
		t.Fatalf("storage directory was not created: %v", err)
	}

	custom, err := NewLocalStorage(Config{LocalPath: t.TempDir(), PresignDuration: time.Hour})
	if err != nil {
		t.Fatalf("NewLocalStorage(custom) error: %v", err)
	}
	if custom.presignDuration != time.Hour {
		t.Fatalf("custom presign duration = %s, want 1h", custom.presignDuration)
	}

	parentFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(parentFile, []byte("x"), 0600); err != nil {
		t.Fatalf("write parent fixture: %v", err)
	}
	if _, err := NewLocalStorage(Config{LocalPath: filepath.Join(parentFile, "child")}); err == nil {
		t.Fatal("NewLocalStorage() succeeded below a regular file")
	}
}

func TestLocalStorageLifecycle(t *testing.T) {
	store, err := NewLocalStorage(Config{LocalPath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLocalStorage() error: %v", err)
	}
	ctx := context.Background()
	content := []byte("image bytes")
	metadata := ImageMetadata{Filename: "photo.png", CollectionName: "animals"}

	imageURL, err := store.Upload(ctx, content, metadata)
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if !strings.HasPrefix(imageURL, "file://") || filepath.Ext(imageURL) != ".png" {
		t.Fatalf("unexpected image URL %q", imageURL)
	}

	exists, err := store.Exists(ctx, imageURL)
	if err != nil || !exists {
		t.Fatalf("Exists() = %v, %v; want true, nil", exists, err)
	}
	downloaded, err := store.Download(ctx, imageURL)
	if err != nil || string(downloaded) != string(content) {
		t.Fatalf("Download() = %q, %v", downloaded, err)
	}
	info, err := store.GetInfo(ctx, imageURL)
	if err != nil {
		t.Fatalf("GetInfo() error: %v", err)
	}
	if info.CollectionName != "animals" || info.Size != int64(len(content)) || filepath.Ext(info.Filename) != ".png" {
		t.Fatalf("unexpected image info: %#v", info)
	}
	if info.Custom == nil {
		t.Fatal("GetInfo() returned nil custom metadata")
	}
	if got, err := store.GetPresignedURL(ctx, imageURL, time.Minute); err != nil || got != imageURL {
		t.Fatalf("GetPresignedURL() = %q, %v", got, err)
	}

	if err := store.Delete(ctx, imageURL); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	exists, err = store.Exists(ctx, imageURL)
	if err != nil || exists {
		t.Fatalf("Exists() after delete = %v, %v; want false, nil", exists, err)
	}
}

func TestLocalStorageDefaultsExtensionAndAcceptsRawPath(t *testing.T) {
	store, err := NewLocalStorage(Config{LocalPath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLocalStorage() error: %v", err)
	}
	imageURL, err := store.Upload(context.Background(), []byte("data"), ImageMetadata{
		Filename:       "no-extension",
		CollectionName: "default",
	})
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}
	if filepath.Ext(imageURL) != ".bin" {
		t.Fatalf("extension = %q, want .bin", filepath.Ext(imageURL))
	}
	rawPath := strings.TrimPrefix(imageURL, "file://")
	data, err := store.Download(context.Background(), rawPath)
	if err != nil || string(data) != "data" {
		t.Fatalf("raw-path Download() = %q, %v", data, err)
	}
	if extractFilePath("short") != "short" || extractFilePath(imageURL) != rawPath {
		t.Fatal("extractFilePath() did not preserve/remove prefixes as expected")
	}
}

func TestLocalStorageErrors(t *testing.T) {
	store, err := NewLocalStorage(Config{LocalPath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLocalStorage() error: %v", err)
	}
	missing := "file://" + filepath.Join(t.TempDir(), "missing.png")

	if _, err := store.Download(context.Background(), missing); !isImageNotFound(err) {
		t.Fatalf("Download(missing) error = %T %v", err, err)
	}
	if err := store.Delete(context.Background(), missing); !isImageNotFound(err) {
		t.Fatalf("Delete(missing) error = %T %v", err, err)
	}
	if _, err := store.GetInfo(context.Background(), missing); !isImageNotFound(err) {
		t.Fatalf("GetInfo(missing) error = %T %v", err, err)
	}

	baseFile := filepath.Join(t.TempDir(), "base-file")
	if err := os.WriteFile(baseFile, []byte("x"), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	broken := &LocalStorage{basePath: baseFile}
	_, err = broken.Upload(context.Background(), []byte("x"), ImageMetadata{
		Filename:       "image.png",
		CollectionName: "collection",
	})
	var uploadErr *ErrUploadFailed
	if !errors.As(err, &uploadErr) {
		t.Fatalf("Upload() error = %T %v, want ErrUploadFailed", err, err)
	}

	if _, err := store.Download(context.Background(), "file://"+t.TempDir()); err == nil {
		t.Fatal("Download(directory) succeeded")
	} else {
		var downloadErr *ErrDownloadFailed
		if !errors.As(err, &downloadErr) {
			t.Fatalf("Download(directory) error = %T, want ErrDownloadFailed", err)
		}
	}

	nonEmptyDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "child"), []byte("x"), 0600); err != nil {
		t.Fatalf("write non-empty directory fixture: %v", err)
	}
	if err := store.Delete(context.Background(), "file://"+nonEmptyDir); err == nil {
		t.Fatal("Delete(non-empty directory) succeeded")
	} else {
		var deleteErr *ErrDeleteFailed
		if !errors.As(err, &deleteErr) {
			t.Fatalf("Delete(non-empty directory) error = %T, want ErrDeleteFailed", err)
		}
	}

	invalidPath := "file://\x00"
	if exists, err := store.Exists(context.Background(), invalidPath); err == nil || exists {
		t.Fatalf("Exists(invalid path) = %v, %v; want false and an error", exists, err)
	}
	if _, err := store.GetInfo(context.Background(), invalidPath); err == nil || !strings.Contains(err.Error(), "failed to get file info") {
		t.Fatalf("GetInfo(invalid path) error = %v", err)
	}
}

func TestStorageErrors(t *testing.T) {
	cause := errors.New("cause")
	tests := []struct {
		err      error
		contains string
		unwrap   bool
	}{
		{err: &ErrUnsupportedStorageType{Type: "ftp"}, contains: "unsupported storage type"},
		{err: &ErrImageNotFound{URL: "image"}, contains: "image not found"},
		{err: &ErrUploadFailed{URL: "image", Err: cause}, contains: "failed to upload", unwrap: true},
		{err: &ErrDownloadFailed{URL: "image", Err: cause}, contains: "failed to download", unwrap: true},
		{err: &ErrDeleteFailed{URL: "image", Err: cause}, contains: "failed to delete", unwrap: true},
	}
	for _, tt := range tests {
		if !strings.Contains(tt.err.Error(), tt.contains) {
			t.Errorf("error %q does not contain %q", tt.err, tt.contains)
		}
		if tt.unwrap && !errors.Is(tt.err, cause) {
			t.Errorf("error %q does not unwrap its cause", tt.err)
		}
	}
}

func isImageNotFound(err error) bool {
	var notFound *ErrImageNotFound
	return errors.As(err, &notFound)
}
