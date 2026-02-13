// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package storage

import "fmt"

// ErrUnsupportedStorageType is returned when an unsupported storage type is requested
type ErrUnsupportedStorageType struct {
	Type string
}

func (e *ErrUnsupportedStorageType) Error() string {
	return fmt.Sprintf("unsupported storage type: %s (supported: s3, minio, local)", e.Type)
}

// ErrImageNotFound is returned when an image doesn't exist
type ErrImageNotFound struct {
	URL string
}

func (e *ErrImageNotFound) Error() string {
	return fmt.Sprintf("image not found: %s", e.URL)
}

// ErrUploadFailed is returned when image upload fails
type ErrUploadFailed struct {
	URL string
	Err error
}

func (e *ErrUploadFailed) Error() string {
	return fmt.Sprintf("failed to upload image to %s: %v", e.URL, e.Err)
}

func (e *ErrUploadFailed) Unwrap() error {
	return e.Err
}

// ErrDownloadFailed is returned when image download fails
type ErrDownloadFailed struct {
	URL string
	Err error
}

func (e *ErrDownloadFailed) Error() string {
	return fmt.Sprintf("failed to download image from %s: %v", e.URL, e.Err)
}

func (e *ErrDownloadFailed) Unwrap() error {
	return e.Err
}

// ErrDeleteFailed is returned when image deletion fails
type ErrDeleteFailed struct {
	URL string
	Err error
}

func (e *ErrDeleteFailed) Error() string {
	return fmt.Sprintf("failed to delete image at %s: %v", e.URL, e.Err)
}

func (e *ErrDeleteFailed) Unwrap() error {
	return e.Err
}
