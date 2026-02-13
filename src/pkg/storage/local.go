// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// LocalStorage implements ImageStorage using local filesystem
// NOTE: For testing only, not recommended for production
type LocalStorage struct {
	basePath        string
	presignDuration time.Duration
}

// NewLocalStorage creates a new local filesystem storage backend
func NewLocalStorage(config Config) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.LocalPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	presignDuration := config.PresignDuration
	if presignDuration == 0 {
		presignDuration = 24 * time.Hour
	}

	return &LocalStorage{
		basePath:        config.LocalPath,
		presignDuration: presignDuration,
	}, nil
}

// Upload stores an image and returns its file path
func (s *LocalStorage) Upload(ctx context.Context, image []byte, metadata ImageMetadata) (string, error) {
	// Generate unique filename
	id := uuid.New().String()
	ext := filepath.Ext(metadata.Filename)
	if ext == "" {
		ext = ".bin"
	}

	// Create collection-specific subdirectory
	collectionDir := filepath.Join(s.basePath, metadata.CollectionName)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		return "", &ErrUploadFailed{URL: collectionDir, Err: err}
	}

	// Build full path
	filename := id + ext
	fullPath := filepath.Join(collectionDir, filename)

	// Write file
	if err := os.WriteFile(fullPath, image, 0644); err != nil {
		return "", &ErrUploadFailed{URL: fullPath, Err: err}
	}

	// Return file:// URL
	return fmt.Sprintf("file://%s", fullPath), nil
}

// Download retrieves an image by its file path
func (s *LocalStorage) Download(ctx context.Context, imageURL string) ([]byte, error) {
	// Extract file path from URL
	filePath := extractFilePath(imageURL)

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrImageNotFound{URL: imageURL}
		}
		return nil, &ErrDownloadFailed{URL: imageURL, Err: err}
	}

	return data, nil
}

// Delete removes an image by its file path
func (s *LocalStorage) Delete(ctx context.Context, imageURL string) error {
	// Extract file path from URL
	filePath := extractFilePath(imageURL)

	// Delete file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return &ErrImageNotFound{URL: imageURL}
		}
		return &ErrDeleteFailed{URL: imageURL, Err: err}
	}

	return nil
}

// Exists checks if an image exists at the given path
func (s *LocalStorage) Exists(ctx context.Context, imageURL string) (bool, error) {
	// Extract file path from URL
	filePath := extractFilePath(imageURL)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetPresignedURL returns the original file:// URL (no presigning for local files)
func (s *LocalStorage) GetPresignedURL(ctx context.Context, imageURL string, duration time.Duration) (string, error) {
	// For local storage, just return the original URL
	return imageURL, nil
}

// GetInfo returns metadata about a stored image
func (s *LocalStorage) GetInfo(ctx context.Context, imageURL string) (*ImageMetadata, error) {
	// Extract file path from URL
	filePath := extractFilePath(imageURL)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ErrImageNotFound{URL: imageURL}
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract collection name from path
	dir := filepath.Dir(filePath)
	collectionName := filepath.Base(dir)

	return &ImageMetadata{
		Filename:       filepath.Base(filePath),
		Size:           info.Size(),
		CollectionName: collectionName,
		Custom:         make(map[string]string),
	}, nil
}

// extractFilePath extracts the file system path from a file:// URL
func extractFilePath(imageURL string) string {
	// Remove file:// prefix if present
	if len(imageURL) > 7 && imageURL[:7] == "file://" {
		return imageURL[7:]
	}
	return imageURL
}
