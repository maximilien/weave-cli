// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package storage

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/draw"
)

// ThumbnailConfig contains configuration for thumbnail generation
type ThumbnailConfig struct {
	// Maximum width in pixels (default: 256)
	MaxWidth int

	// Maximum height in pixels (default: 256)
	MaxHeight int

	// JPEG quality (1-100, default: 85)
	Quality int

	// Output format: "jpeg" or "png" (default: "jpeg")
	Format string
}

// DefaultThumbnailConfig returns sensible defaults for thumbnail generation
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		MaxWidth:  256,
		MaxHeight: 256,
		Quality:   85,
		Format:    "jpeg",
	}
}

// GenerateThumbnail creates a thumbnail from an image
// Returns the thumbnail as bytes and its dimensions
func GenerateThumbnail(imageData []byte, config ThumbnailConfig) ([]byte, int, int, error) {
	// Decode original image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate thumbnail dimensions (maintain aspect ratio)
	thumbWidth, thumbHeight := calculateThumbnailSize(origWidth, origHeight, config.MaxWidth, config.MaxHeight)

	// Skip resizing if image is already smaller than thumbnail size
	if origWidth <= thumbWidth && origHeight <= thumbHeight {
		// Return original image if it's already small enough
		return imageData, origWidth, origHeight, nil
	}

	// Create thumbnail image
	thumbnail := image.NewRGBA(image.Rect(0, 0, thumbWidth, thumbHeight))

	// Resize using high-quality Lanczos algorithm
	draw.CatmullRom.Scale(thumbnail, thumbnail.Bounds(), img, bounds, draw.Over, nil)

	// Encode thumbnail
	var buf bytes.Buffer
	outputFormat := config.Format
	if outputFormat == "" {
		outputFormat = format // Use original format if not specified
	}

	switch outputFormat {
	case "jpeg", "jpg":
		quality := config.Quality
		if quality == 0 {
			quality = 85
		}
		if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: quality}); err != nil {
			return nil, 0, 0, fmt.Errorf("failed to encode JPEG thumbnail: %w", err)
		}
	case "png":
		if err := png.Encode(&buf, thumbnail); err != nil {
			return nil, 0, 0, fmt.Errorf("failed to encode PNG thumbnail: %w", err)
		}
	default:
		return nil, 0, 0, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return buf.Bytes(), thumbWidth, thumbHeight, nil
}

// GenerateThumbnailBase64 creates a base64-encoded thumbnail
func GenerateThumbnailBase64(imageData []byte, config ThumbnailConfig) (string, int, int, error) {
	thumbnail, width, height, err := GenerateThumbnail(imageData, config)
	if err != nil {
		return "", 0, 0, err
	}

	encoded := base64.StdEncoding.EncodeToString(thumbnail)
	return encoded, width, height, nil
}

// calculateThumbnailSize calculates thumbnail dimensions while maintaining aspect ratio
func calculateThumbnailSize(origWidth, origHeight, maxWidth, maxHeight int) (int, int) {
	// Calculate aspect ratio
	aspectRatio := float64(origWidth) / float64(origHeight)

	thumbWidth := origWidth
	thumbHeight := origHeight

	// Scale down to fit within max dimensions
	if thumbWidth > maxWidth {
		thumbWidth = maxWidth
		thumbHeight = int(float64(thumbWidth) / aspectRatio)
	}

	if thumbHeight > maxHeight {
		thumbHeight = maxHeight
		thumbWidth = int(float64(thumbHeight) * aspectRatio)
	}

	return thumbWidth, thumbHeight
}

// EstimateThumbnailSize estimates the size of a thumbnail in bytes
// This is useful for checking if it will fit within Milvus VARCHAR limits (65KB)
func EstimateThumbnailSize(origWidth, origHeight int, config ThumbnailConfig) int {
	thumbWidth, thumbHeight := calculateThumbnailSize(origWidth, origHeight, config.MaxWidth, config.MaxHeight)

	// Rough estimate based on format and quality
	pixels := thumbWidth * thumbHeight

	switch config.Format {
	case "jpeg", "jpg":
		// JPEG: ~0.5-1.5 bytes per pixel depending on quality
		quality := config.Quality
		if quality == 0 {
			quality = 85
		}
		bytesPerPixel := 0.5 + (float64(quality) / 100.0)
		return int(float64(pixels) * bytesPerPixel)

	case "png":
		// PNG: ~2-4 bytes per pixel (depends on compression)
		return pixels * 3

	default:
		// Conservative estimate
		return pixels * 2
	}
}

// IsThumbnailNeeded checks if a thumbnail is needed to fit within Milvus limits
// Milvus VARCHAR limit: 65,535 bytes
// Base64 overhead: 1.37x
// Safe limit for original image: ~47KB (~48,000 bytes)
func IsThumbnailNeeded(imageData []byte) bool {
	const MaxSafeSize = 48000 // ~47KB in bytes

	return len(imageData) > MaxSafeSize
}

// GenerateSafeThumbnail generates a thumbnail that fits within Milvus VARCHAR limits
// Returns base64-encoded thumbnail that is guaranteed to be <65KB
func GenerateSafeThumbnail(imageData []byte) (string, error) {
	const MilvusVarcharLimit = 65535
	const Base64Overhead = 1.37
	// 90% of limit for safety: 65535 / 1.37 * 0.9 ≈ 43052
	const TargetSize = 43052

	// Try different thumbnail sizes until we find one that fits
	configs := []ThumbnailConfig{
		{MaxWidth: 256, MaxHeight: 256, Quality: 85, Format: "jpeg"},
		{MaxWidth: 200, MaxHeight: 200, Quality: 80, Format: "jpeg"},
		{MaxWidth: 150, MaxHeight: 150, Quality: 75, Format: "jpeg"},
		{MaxWidth: 128, MaxHeight: 128, Quality: 70, Format: "jpeg"},
		{MaxWidth: 100, MaxHeight: 100, Quality: 65, Format: "jpeg"},
	}

	for _, config := range configs {
		thumbnail, _, _, err := GenerateThumbnailBase64(imageData, config)
		if err != nil {
			continue // Try next size
		}

		// Check if it fits
		if len(thumbnail) <= TargetSize {
			return thumbnail, nil
		}
	}

	return "", fmt.Errorf("failed to generate thumbnail small enough to fit in Milvus VARCHAR limit (65KB)")
}
