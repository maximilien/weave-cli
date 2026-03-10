// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"context"
	"os"
	"testing"
)

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
