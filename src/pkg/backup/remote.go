// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// RemoteStorageConfig holds configuration for remote storage backends
type RemoteStorageConfig struct {
	// Type: "s3" or "minio"
	Type string

	// S3/MinIO endpoint (MinIO only, S3 uses region-based endpoints)
	Endpoint string

	// AWS Region (required for S3, optional for MinIO)
	Region string

	// S3 Bucket name
	Bucket string

	// Access credentials
	AccessKeyID     string
	SecretAccessKey string

	// Use SSL/TLS (default: true for S3, false for MinIO)
	UseSSL bool

	// Optional path prefix within bucket (e.g., "backups/")
	PathPrefix string
}

// RemoteStorage handles backup uploads/downloads to S3/MinIO
type RemoteStorage struct {
	config *RemoteStorageConfig
	client *s3.S3
}

// NewRemoteStorage creates a new remote storage client
func NewRemoteStorage(cfg *RemoteStorageConfig) (*RemoteStorage, error) {
	// Validate config
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("access key and secret key are required")
	}

	// Create AWS session config
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		Region: aws.String(cfg.Region),
	}

	// Configure for MinIO if specified
	if cfg.Type == "minio" {
		if cfg.Endpoint == "" {
			return nil, fmt.Errorf("endpoint is required for MinIO")
		}

		awsConfig.Endpoint = aws.String(cfg.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true) // MinIO requires path-style addressing
		awsConfig.DisableSSL = aws.Bool(!cfg.UseSSL)
	} else {
		// S3 defaults
		if cfg.Region == "" {
			cfg.Region = "us-east-1" // Default region
		}
		awsConfig.Region = aws.String(cfg.Region)
	}

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 client
	client := s3.New(sess)

	return &RemoteStorage{
		config: cfg,
		client: client,
	}, nil
}

// Upload uploads a backup to remote storage
func (r *RemoteStorage) Upload(ctx context.Context, key string, data []byte) error {
	// Add path prefix if configured
	fullKey := r.buildKey(key)

	// Upload to S3/MinIO
	_, err := r.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(fullKey),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to %s: %w", r.config.Type, err)
	}

	return nil
}

// UploadStream uploads a backup from a reader
func (r *RemoteStorage) UploadStream(ctx context.Context, key string, reader io.Reader) error {
	// Read all data (S3 PutObject requires knowing content length)
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read backup data: %w", err)
	}

	// Upload (buildKey is called in Upload method)
	return r.Upload(ctx, key, data)
}

// Download downloads a backup from remote storage
func (r *RemoteStorage) Download(ctx context.Context, key string) ([]byte, error) {
	// Add path prefix if configured
	fullKey := r.buildKey(key)

	// Download from S3/MinIO
	result, err := r.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download from %s: %w", r.config.Type, err)
	}
	defer result.Body.Close()

	// Read all data
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup data: %w", err)
	}

	return data, nil
}

// DownloadStream downloads a backup and returns a reader
func (r *RemoteStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Add path prefix if configured
	fullKey := r.buildKey(key)

	// Download from S3/MinIO
	result, err := r.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download from %s: %w", r.config.Type, err)
	}

	return result.Body, nil
}

// List lists backups in remote storage with optional prefix
func (r *RemoteStorage) List(ctx context.Context, prefix string) ([]string, error) {
	// Build full prefix
	fullPrefix := r.buildKey(prefix)

	// List objects
	result, err := r.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.config.Bucket),
		Prefix: aws.String(fullPrefix),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// Extract keys
	keys := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		// Remove path prefix to get relative key
		key := aws.StringValue(obj.Key)
		if r.config.PathPrefix != "" {
			key = strings.TrimPrefix(key, r.config.PathPrefix+"/")
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// Delete deletes a backup from remote storage
func (r *RemoteStorage) Delete(ctx context.Context, key string) error {
	// Add path prefix if configured
	fullKey := r.buildKey(key)

	// Delete from S3/MinIO
	_, err := r.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from %s: %w", r.config.Type, err)
	}

	return nil
}

// Exists checks if a backup exists in remote storage
func (r *RemoteStorage) Exists(ctx context.Context, key string) (bool, error) {
	// Add path prefix if configured
	fullKey := r.buildKey(key)

	// Check if object exists
	_, err := r.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// buildKey constructs the full S3 key with path prefix
func (r *RemoteStorage) buildKey(key string) string {
	if r.config.PathPrefix == "" {
		return key
	}
	return r.config.PathPrefix + "/" + key
}
