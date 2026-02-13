// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

//go:build integration
// +build integration

package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_MinioAutoBucketCreation tests automatic bucket creation
// when the bucket doesn't exist during upload operations
func TestIntegration_MinioAutoBucketCreation(t *testing.T) {
	// Check if MinIO is available
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	// Create a unique bucket name for this test
	bucketName := "test-auto-create-" + time.Now().Format("20060102-150405")

	// Clean up bucket at the end
	defer func() {
		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})
		if err == nil {
			ctx := context.Background()
			// Remove all objects first
			objectsCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})
			for object := range objectsCh {
				if object.Err != nil {
					continue
				}
				_ = client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
			}
			// Then remove the bucket
			_ = client.RemoveBucket(ctx, bucketName)
		}
	}()

	t.Run("AutoCreateBucket", func(t *testing.T) {
		// Create MinIO storage with a non-existent bucket
		config := Config{
			Type:       StorageTypeMinio,
			Endpoint:   endpoint,
			AccessKey:  accessKey,
			SecretKey:  secretKey,
			Bucket:     bucketName,
			PathPrefix: "test",
			UseSSL:     false,
		}

		storage, err := NewMinioStorage(config)
		require.NoError(t, err, "Failed to create MinIO storage")

		ctx := context.Background()

		// Verify bucket doesn't exist yet
		exists, err := storage.client.BucketExists(ctx, bucketName)
		require.NoError(t, err)
		assert.False(t, exists, "Bucket should not exist before upload")

		// Upload an image - this should auto-create the bucket
		testImage := []byte("fake-image-data-for-testing")
		metadata := ImageMetadata{
			Filename:       "test-image.jpg",
			ContentType:    "image/jpeg",
			Size:           int64(len(testImage)),
			CollectionName: "TestCollection",
			DocumentID:     "test-doc-1",
			Custom:         make(map[string]string),
		}

		imageURL, err := storage.Upload(ctx, testImage, metadata)
		require.NoError(t, err, "Upload should succeed and auto-create bucket")
		assert.NotEmpty(t, imageURL, "Image URL should be returned")

		// Verify bucket was created
		exists, err = storage.client.BucketExists(ctx, bucketName)
		require.NoError(t, err)
		assert.True(t, exists, "Bucket should exist after upload")

		// Verify the image was uploaded correctly
		downloadedImage, err := storage.Download(ctx, imageURL)
		require.NoError(t, err, "Should be able to download the uploaded image")
		assert.Equal(t, testImage, downloadedImage, "Downloaded image should match uploaded image")
	})

	t.Run("SecondUploadReusesBucket", func(t *testing.T) {
		// Create storage again with the same bucket (which now exists)
		config := Config{
			Type:       StorageTypeMinio,
			Endpoint:   endpoint,
			AccessKey:  accessKey,
			SecretKey:  secretKey,
			Bucket:     bucketName,
			PathPrefix: "test",
			UseSSL:     false,
		}

		storage, err := NewMinioStorage(config)
		require.NoError(t, err)

		ctx := context.Background()

		// Verify bucket exists from previous test
		exists, err := storage.client.BucketExists(ctx, bucketName)
		require.NoError(t, err)
		assert.True(t, exists, "Bucket should still exist from first upload")

		// Upload another image - should not fail
		testImage := []byte("another-test-image-data")
		metadata := ImageMetadata{
			Filename:       "test-image-2.jpg",
			ContentType:    "image/jpeg",
			Size:           int64(len(testImage)),
			CollectionName: "TestCollection",
			DocumentID:     "test-doc-2",
			Custom:         make(map[string]string),
		}

		imageURL, err := storage.Upload(ctx, testImage, metadata)
		require.NoError(t, err, "Second upload should succeed")
		assert.NotEmpty(t, imageURL, "Image URL should be returned")

		// Verify the second image was uploaded
		downloadedImage, err := storage.Download(ctx, imageURL)
		require.NoError(t, err)
		assert.Equal(t, testImage, downloadedImage)
	})
}

// TestIntegration_MinioBasicOperations tests basic MinIO operations
// This is a sanity check for the storage implementation
func TestIntegration_MinioBasicOperations(t *testing.T) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := "test-basic-ops-" + time.Now().Format("20060102-150405")

	config := Config{
		Type:       StorageTypeMinio,
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		Bucket:     bucketName,
		PathPrefix: "test",
		UseSSL:     false,
	}

	storage, err := NewMinioStorage(config)
	require.NoError(t, err)

	ctx := context.Background()

	// Clean up
	defer func() {
		client, _ := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})
		if client != nil {
			objectsCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})
			for object := range objectsCh {
				if object.Err != nil {
					continue
				}
				_ = client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
			}
			_ = client.RemoveBucket(ctx, bucketName)
		}
	}()

	t.Run("UploadDownloadDelete", func(t *testing.T) {
		testImage := []byte("test-image-data")
		metadata := ImageMetadata{
			Filename:       "test.jpg",
			ContentType:    "image/jpeg",
			Size:           int64(len(testImage)),
			CollectionName: "Test",
			DocumentID:     "doc1",
			Custom:         map[string]string{"test": "value"},
		}

		// Upload
		url, err := storage.Upload(ctx, testImage, metadata)
		require.NoError(t, err)
		assert.NotEmpty(t, url)

		// Exists
		exists, err := storage.Exists(ctx, url)
		require.NoError(t, err)
		assert.True(t, exists)

		// Download
		downloaded, err := storage.Download(ctx, url)
		require.NoError(t, err)
		assert.Equal(t, testImage, downloaded)

		// Get Info
		info, err := storage.GetInfo(ctx, url)
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", info.ContentType)
		assert.Equal(t, "Test", info.CollectionName)
		assert.Equal(t, "doc1", info.DocumentID)

		// Delete
		err = storage.Delete(ctx, url)
		require.NoError(t, err)

		// Verify deleted
		exists, err = storage.Exists(ctx, url)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
