// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package storage

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage implements ImageStorage using MinIO (S3-compatible)
type MinioStorage struct {
	client          *minio.Client
	bucket          string
	pathPrefix      string
	presignDuration time.Duration
	endpoint        string
	useSSL          bool
}

// NewMinioStorage creates a new MinIO storage backend
func NewMinioStorage(config Config) (*MinioStorage, error) {
	// Create MinIO client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Set default presign duration if not specified
	presignDuration := config.PresignDuration
	if presignDuration == 0 {
		presignDuration = 24 * time.Hour // Default: 24 hours
	}

	return &MinioStorage{
		client:          client,
		bucket:          config.Bucket,
		pathPrefix:      strings.TrimSuffix(config.PathPrefix, "/"),
		presignDuration: presignDuration,
		endpoint:        config.Endpoint,
		useSSL:          config.UseSSL,
	}, nil
}

// NewS3Storage creates a new AWS S3 storage backend
// Uses the same implementation as MinIO (S3-compatible)
func NewS3Storage(config Config) (*MinioStorage, error) {
	// S3 endpoint is region-specific
	if config.Endpoint == "" {
		config.Endpoint = fmt.Sprintf("s3.%s.amazonaws.com", config.Region)
	}

	// S3 always uses SSL
	config.UseSSL = true

	return NewMinioStorage(config)
}

// Upload stores an image and returns its URL
func (s *MinioStorage) Upload(ctx context.Context, image []byte, metadata ImageMetadata) (string, error) {
	// Generate unique key for the image
	key := s.generateKey(metadata)

	// Prepare upload options
	opts := minio.PutObjectOptions{
		ContentType: metadata.ContentType,
		UserMetadata: map[string]string{
			"collection": metadata.CollectionName,
			"document":   metadata.DocumentID,
			"filename":   metadata.Filename,
		},
	}

	// Add custom metadata
	for k, v := range metadata.Custom {
		opts.UserMetadata[k] = v
	}

	// Upload to MinIO/S3
	reader := strings.NewReader(string(image))
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, int64(len(image)), opts)
	if err != nil {
		return "", &ErrUploadFailed{URL: key, Err: err}
	}

	// Return the full URL
	return s.buildURL(key), nil
}

// Download retrieves an image by its URL
func (s *MinioStorage) Download(ctx context.Context, imageURL string) ([]byte, error) {
	// Extract key from URL
	key, err := s.extractKey(imageURL)
	if err != nil {
		return nil, err
	}

	// Download from MinIO/S3
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, &ErrDownloadFailed{URL: imageURL, Err: err}
	}
	defer object.Close()

	// Read object data
	buf := make([]byte, 0, 1024*1024) // Start with 1MB buffer
	tmp := make([]byte, 4096)
	for {
		n, err := object.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, &ErrDownloadFailed{URL: imageURL, Err: err}
		}
	}

	return buf, nil
}

// Delete removes an image by its URL
func (s *MinioStorage) Delete(ctx context.Context, imageURL string) error {
	// Extract key from URL
	key, err := s.extractKey(imageURL)
	if err != nil {
		return err
	}

	// Delete from MinIO/S3
	err = s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return &ErrDeleteFailed{URL: imageURL, Err: err}
	}

	return nil
}

// Exists checks if an image exists at the given URL
func (s *MinioStorage) Exists(ctx context.Context, imageURL string) (bool, error) {
	// Extract key from URL
	key, err := s.extractKey(imageURL)
	if err != nil {
		return false, err
	}

	// Check if object exists
	_, err = s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetPresignedURL returns a pre-signed URL for temporary access
func (s *MinioStorage) GetPresignedURL(ctx context.Context, imageURL string, duration time.Duration) (string, error) {
	// Extract key from URL
	key, err := s.extractKey(imageURL)
	if err != nil {
		return "", err
	}

	// Use provided duration or default
	if duration == 0 {
		duration = s.presignDuration
	}

	// Generate presigned URL
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, duration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GetInfo returns metadata about a stored image
func (s *MinioStorage) GetInfo(ctx context.Context, imageURL string) (*ImageMetadata, error) {
	// Extract key from URL
	key, err := s.extractKey(imageURL)
	if err != nil {
		return nil, err
	}

	// Get object info
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return nil, &ErrImageNotFound{URL: imageURL}
		}
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	// Extract metadata
	metadata := &ImageMetadata{
		ContentType:    info.ContentType,
		Size:           info.Size,
		CollectionName: info.UserMetadata["X-Amz-Meta-Collection"],
		DocumentID:     info.UserMetadata["X-Amz-Meta-Document"],
		Filename:       info.UserMetadata["X-Amz-Meta-Filename"],
		Custom:         make(map[string]string),
	}

	// Extract custom metadata (skip standard MinIO headers)
	for k, v := range info.UserMetadata {
		if !strings.HasPrefix(k, "X-Amz-") && k != "collection" && k != "document" && k != "filename" {
			metadata.Custom[k] = v
		}
	}

	return metadata, nil
}

// generateKey creates a unique key for storing an image
func (s *MinioStorage) generateKey(metadata ImageMetadata) string {
	// Generate unique ID
	id := uuid.New().String()

	// Build path: <prefix>/<collection>/<year>/<month>/<uuid>.<ext>
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// Extract extension from filename
	ext := path.Ext(metadata.Filename)
	if ext == "" {
		// Determine extension from content type
		switch metadata.ContentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".bin"
		}
	}

	// Build key
	parts := []string{}
	if s.pathPrefix != "" {
		parts = append(parts, s.pathPrefix)
	}
	if metadata.CollectionName != "" {
		parts = append(parts, metadata.CollectionName)
	}
	parts = append(parts, year, month, id+ext)

	return path.Join(parts...)
}

// buildURL constructs the full URL for a key
func (s *MinioStorage) buildURL(key string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, key)
}

// extractKey extracts the storage key from a full URL
func (s *MinioStorage) extractKey(imageURL string) (string, error) {
	// Parse URL
	u, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("invalid image URL: %w", err)
	}

	// Extract path (remove leading slash and bucket name)
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimPrefix(path, s.bucket+"/")

	return path, nil
}
