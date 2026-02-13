// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package storage

import (
	"context"
	"time"
)

// ImageMetadata contains metadata about an uploaded image
type ImageMetadata struct {
	// Original filename
	Filename string

	// MIME type (e.g., "image/jpeg", "image/png")
	ContentType string

	// Size in bytes
	Size int64

	// Image dimensions
	Width  int
	Height int

	// Collection and document IDs for organization
	CollectionName string
	DocumentID     string

	// Additional custom metadata
	Custom map[string]string
}

// ImageStorage defines the interface for external image storage backends
type ImageStorage interface {
	// Upload stores an image and returns its URL
	// The URL format depends on the storage backend:
	// - S3: https://bucket.s3.region.amazonaws.com/path/to/image.jpg
	// - MinIO: http://localhost:9000/bucket/path/to/image.jpg
	Upload(ctx context.Context, image []byte, metadata ImageMetadata) (string, error)

	// Download retrieves an image by its URL
	Download(ctx context.Context, url string) ([]byte, error)

	// Delete removes an image by its URL
	Delete(ctx context.Context, url string) error

	// Exists checks if an image exists at the given URL
	Exists(ctx context.Context, url string) (bool, error)

	// GetURL returns a pre-signed URL for temporary access (if supported)
	// Duration specifies how long the URL should be valid
	// Returns the original URL if pre-signing is not supported
	GetPresignedURL(ctx context.Context, url string, duration time.Duration) (string, error)

	// GetInfo returns metadata about a stored image
	GetInfo(ctx context.Context, url string) (*ImageMetadata, error)
}

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeMinio StorageType = "minio"
	StorageTypeLocal StorageType = "local" // For testing only
)

// Config contains configuration for storage backends
type Config struct {
	// Type of storage (s3, minio, local)
	Type StorageType

	// S3/MinIO configuration
	Endpoint        string // MinIO endpoint (e.g., "localhost:9000")
	AccessKey       string
	SecretKey       string
	Region          string
	Bucket          string
	UseSSL          bool
	PathPrefix      string // Optional prefix for all keys (e.g., "images/")
	PresignDuration time.Duration

	// Local storage configuration (for testing)
	LocalPath string // Base directory for local storage
}

// NewImageStorage creates a new ImageStorage instance based on the config
func NewImageStorage(config Config) (ImageStorage, error) {
	switch config.Type {
	case StorageTypeS3:
		return NewS3Storage(config)
	case StorageTypeMinio:
		return NewMinioStorage(config)
	case StorageTypeLocal:
		return NewLocalStorage(config)
	default:
		return nil, &ErrUnsupportedStorageType{Type: string(config.Type)}
	}
}
