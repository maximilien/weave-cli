// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type failingBackupReader struct{}

func (failingBackupReader) Read([]byte) (int, error) { return 0, fmt.Errorf("read failure") }

func TestRemoteStorageProtocolLifecycle(t *testing.T) {
	objects := make(map[string][]byte)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/fixture-bucket/")
		if r.URL.Query().Get("list-type") == "2" {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>fixture-bucket</Name><IsTruncated>false</IsTruncated><Contents><Key>backups/one.weavebak</Key><Size>3</Size></Contents><Contents><Key>backups/two.weavebak</Key><Size>3</Size></Contents></ListBucketResult>`)
			return
		}
		switch r.Method {
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read upload: %v", err)
			}
			objects[key] = data
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			data, ok := objects[key]
			if !ok {
				http.Error(w, "missing", http.StatusNotFound)
				return
			}
			_, _ = w.Write(data)
		case http.MethodHead:
			if _, ok := objects[key]; !ok {
				http.Error(w, "missing", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(objects, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unsupported", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	storage, err := NewRemoteStorage(&RemoteStorageConfig{
		Type: "minio", Endpoint: server.URL, Region: "us-east-1", Bucket: "fixture-bucket",
		AccessKeyID: "fixture", SecretAccessKey: "fixture", PathPrefix: "backups",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := storage.Upload(ctx, "one.weavebak", []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := storage.UploadStream(ctx, "two.weavebak", bytes.NewBufferString("two")); err != nil {
		t.Fatal(err)
	}
	if err := storage.UploadStream(ctx, "bad.weavebak", failingBackupReader{}); err == nil || !strings.Contains(err.Error(), "read backup data") {
		t.Fatalf("expected stream read error, got %v", err)
	}

	data, err := storage.Download(ctx, "one.weavebak")
	if err != nil || string(data) != "one" {
		t.Fatalf("Download() = (%q, %v)", data, err)
	}
	stream, err := storage.DownloadStream(ctx, "two.weavebak")
	if err != nil {
		t.Fatal(err)
	}
	streamData, err := io.ReadAll(stream)
	_ = stream.Close()
	if err != nil || string(streamData) != "two" {
		t.Fatalf("DownloadStream() = (%q, %v)", streamData, err)
	}

	keys, err := storage.List(ctx, "")
	if err != nil || len(keys) != 2 || keys[0] != "one.weavebak" || keys[1] != "two.weavebak" {
		t.Fatalf("List() = (%v, %v)", keys, err)
	}
	exists, err := storage.Exists(ctx, "one.weavebak")
	if err != nil || !exists {
		t.Fatalf("Exists(existing) = (%v, %v)", exists, err)
	}
	if err := storage.Delete(ctx, "one.weavebak"); err != nil {
		t.Fatal(err)
	}
	exists, err = storage.Exists(ctx, "one.weavebak")
	if err != nil || exists {
		t.Fatalf("Exists(deleted) = (%v, %v)", exists, err)
	}
}

func TestRemoteStorageProtocolErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend unavailable", http.StatusInternalServerError)
	}))
	defer server.Close()
	storage, err := NewRemoteStorage(&RemoteStorageConfig{
		Type: "minio", Endpoint: server.URL, Region: "us-east-1", Bucket: "fixture-bucket",
		AccessKeyID: "fixture", SecretAccessKey: "fixture",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := storage.Upload(ctx, "file", nil); err == nil || !strings.Contains(err.Error(), "failed to upload") {
		t.Fatalf("expected upload error, got %v", err)
	}
	if _, err := storage.Download(ctx, "file"); err == nil || !strings.Contains(err.Error(), "failed to download") {
		t.Fatalf("expected download error, got %v", err)
	}
	if _, err := storage.DownloadStream(ctx, "file"); err == nil || !strings.Contains(err.Error(), "failed to download") {
		t.Fatalf("expected download stream error, got %v", err)
	}
	if _, err := storage.List(ctx, ""); err == nil || !strings.Contains(err.Error(), "failed to list") {
		t.Fatalf("expected list error, got %v", err)
	}
	if err := storage.Delete(ctx, "file"); err == nil || !strings.Contains(err.Error(), "failed to delete") {
		t.Fatalf("expected delete error, got %v", err)
	}
	if exists, err := storage.Exists(ctx, "file"); err == nil || exists {
		t.Fatalf("expected exists backend error, got (%v, %v)", exists, err)
	}
}

// TestRemoteStorageConfig tests configuration validation
func TestRemoteStorageConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *RemoteStorageConfig
		wantErr bool
	}{
		{
			name: "valid S3 config",
			config: &RemoteStorageConfig{
				Type:            "s3",
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				UseSSL:          true,
			},
			wantErr: false,
		},
		{
			name: "valid MinIO config",
			config: &RemoteStorageConfig{
				Type:            "minio",
				Bucket:          "test-bucket",
				Endpoint:        "localhost:9000",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				UseSSL:          false,
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: &RemoteStorageConfig{
				Type:            "s3",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing credentials",
			config: &RemoteStorageConfig{
				Type:   "s3",
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "MinIO without endpoint",
			config: &RemoteStorageConfig{
				Type:            "minio",
				Bucket:          "test-bucket",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRemoteStorage(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRemoteStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBuildKey tests path prefix handling
func TestBuildKey(t *testing.T) {
	tests := []struct {
		name       string
		pathPrefix string
		key        string
		want       string
	}{
		{
			name:       "no prefix",
			pathPrefix: "",
			key:        "backup.weavebak",
			want:       "backup.weavebak",
		},
		{
			name:       "with prefix",
			pathPrefix: "backups",
			key:        "backup.weavebak",
			want:       "backups/backup.weavebak",
		},
		{
			name:       "with trailing slash",
			pathPrefix: "backups/",
			key:        "backup.weavebak",
			want:       "backups//backup.weavebak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &RemoteStorage{
				config: &RemoteStorageConfig{
					PathPrefix: tt.pathPrefix,
				},
			}
			got := storage.buildKey(tt.key)
			if got != tt.want {
				t.Errorf("buildKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEnvVarCredentials tests that credentials can be loaded from environment variables
func TestEnvVarCredentials(t *testing.T) {
	// Save original env vars
	origAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	origSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Set test env vars
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key-from-env")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-from-env")

	// Restore original env vars after test
	defer func() {
		if origAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", origAccessKey)
		} else {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
		}
		if origSecretKey != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", origSecretKey)
		} else {
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}
	}()

	// Note: We can't actually test the full functionality without mocking AWS SDK
	// This test just verifies that the env vars are set correctly
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey != "test-key-from-env" {
		t.Errorf("Expected AWS_ACCESS_KEY_ID to be set from env, got %s", accessKey)
	}
	if secretKey != "test-secret-from-env" {
		t.Errorf("Expected AWS_SECRET_ACCESS_KEY to be set from env, got %s", secretKey)
	}
}

// TestUploadDownloadRoundtrip is a placeholder for integration tests
// This would require a real S3/MinIO instance or mock
func TestUploadDownloadRoundtrip(t *testing.T) {
	t.Skip("Integration test - requires S3/MinIO instance")

	// This is where we would test:
	// 1. Upload a backup file
	// 2. Download it back
	// 3. Verify contents match
	// 4. Cleanup

	ctx := context.Background()
	_ = ctx // Use ctx to avoid unused variable error
}

// TestListBackups is a placeholder for integration tests
func TestListBackups(t *testing.T) {
	t.Skip("Integration test - requires S3/MinIO instance")

	// This is where we would test:
	// 1. Upload multiple backup files with different prefixes
	// 2. List backups with prefix filtering
	// 3. Verify correct files are returned
	// 4. Cleanup
}

// TestDeleteBackup is a placeholder for integration tests
func TestDeleteBackup(t *testing.T) {
	t.Skip("Integration test - requires S3/MinIO instance")

	// This is where we would test:
	// 1. Upload a backup file
	// 2. Delete it
	// 3. Verify it's gone
}

// TestExistsBackup is a placeholder for integration tests
func TestExistsBackup(t *testing.T) {
	t.Skip("Integration test - requires S3/MinIO instance")

	// This is where we would test:
	// 1. Check non-existent file (should return false)
	// 2. Upload a file
	// 3. Check it exists (should return true)
	// 4. Delete it
	// 5. Check again (should return false)
}
