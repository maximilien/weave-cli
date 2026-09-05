// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package storage

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestDefaultThumbnailConfig(t *testing.T) {
	config := DefaultThumbnailConfig()
	if config.MaxWidth != 256 || config.MaxHeight != 256 || config.Quality != 85 || config.Format != "jpeg" {
		t.Fatalf("unexpected defaults: %#v", config)
	}
}

func TestCalculateThumbnailSize(t *testing.T) {
	tests := []struct {
		name                  string
		width, height         int
		maxWidth, maxHeight   int
		wantWidth, wantHeight int
	}{
		{name: "unchanged", width: 20, height: 10, maxWidth: 100, maxHeight: 100, wantWidth: 20, wantHeight: 10},
		{name: "landscape", width: 200, height: 100, maxWidth: 50, maxHeight: 50, wantWidth: 50, wantHeight: 25},
		{name: "portrait", width: 100, height: 200, maxWidth: 50, maxHeight: 50, wantWidth: 25, wantHeight: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := calculateThumbnailSize(tt.width, tt.height, tt.maxWidth, tt.maxHeight)
			if width != tt.wantWidth || height != tt.wantHeight {
				t.Fatalf("size = %dx%d, want %dx%d", width, height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestGenerateThumbnail(t *testing.T) {
	small := encodePNG(t, 20, 10)
	thumbnail, width, height, err := GenerateThumbnail(small, ThumbnailConfig{
		MaxWidth: 50, MaxHeight: 50, Format: "jpeg",
	})
	if err != nil || !bytes.Equal(thumbnail, small) || width != 20 || height != 10 {
		t.Fatalf("small thumbnail = %dx%d, %v (same=%v)", width, height, err, bytes.Equal(thumbnail, small))
	}

	large := encodePNG(t, 200, 100)
	tests := []struct {
		name   string
		config ThumbnailConfig
	}{
		{name: "jpeg", config: ThumbnailConfig{MaxWidth: 50, MaxHeight: 50, Quality: 75, Format: "jpeg"}},
		{name: "jpeg defaults quality", config: ThumbnailConfig{MaxWidth: 50, MaxHeight: 50, Format: "jpg"}},
		{name: "png", config: ThumbnailConfig{MaxWidth: 50, MaxHeight: 50, Format: "png"}},
		{name: "original format", config: ThumbnailConfig{MaxWidth: 50, MaxHeight: 50}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, gotWidth, gotHeight, err := GenerateThumbnail(large, tt.config)
			if err != nil {
				t.Fatalf("GenerateThumbnail() error: %v", err)
			}
			if gotWidth != 50 || gotHeight != 25 {
				t.Fatalf("thumbnail size = %dx%d, want 50x25", gotWidth, gotHeight)
			}
			decoded, _, err := image.Decode(bytes.NewReader(data))
			if err != nil || decoded.Bounds().Dx() != 50 || decoded.Bounds().Dy() != 25 {
				t.Fatalf("decode thumbnail: %v", err)
			}
		})
	}

	if _, _, _, err := GenerateThumbnail([]byte("not an image"), DefaultThumbnailConfig()); err == nil {
		t.Fatal("GenerateThumbnail() accepted invalid image data")
	}
	if _, _, _, err := GenerateThumbnail(large, ThumbnailConfig{MaxWidth: 50, MaxHeight: 50, Format: "gif"}); err == nil {
		t.Fatal("GenerateThumbnail() accepted unsupported output format")
	}
}

func TestGenerateThumbnailBase64(t *testing.T) {
	encoded, width, height, err := GenerateThumbnailBase64(encodePNG(t, 100, 50), ThumbnailConfig{
		MaxWidth: 20, MaxHeight: 20, Format: "png",
	})
	if err != nil || width != 20 || height != 10 {
		t.Fatalf("GenerateThumbnailBase64() = %dx%d, %v", width, height, err)
	}
	if _, err := base64.StdEncoding.DecodeString(encoded); err != nil {
		t.Fatalf("thumbnail is not base64: %v", err)
	}
	if _, _, _, err := GenerateThumbnailBase64([]byte("invalid"), DefaultThumbnailConfig()); err == nil {
		t.Fatal("GenerateThumbnailBase64() accepted invalid image data")
	}
}

func TestThumbnailSizeEstimates(t *testing.T) {
	jpegConfig := ThumbnailConfig{MaxWidth: 100, MaxHeight: 50, Quality: 50, Format: "jpeg"}
	if got := EstimateThumbnailSize(200, 100, jpegConfig); got != 5000 {
		t.Fatalf("JPEG estimate = %d, want 5000", got)
	}
	jpegConfig.Quality = 0
	if got := EstimateThumbnailSize(200, 100, jpegConfig); got != 6750 {
		t.Fatalf("default-quality JPEG estimate = %d, want 6750", got)
	}
	pngConfig := ThumbnailConfig{MaxWidth: 100, MaxHeight: 50, Format: "png"}
	if got := EstimateThumbnailSize(200, 100, pngConfig); got != 15000 {
		t.Fatalf("PNG estimate = %d, want 15000", got)
	}
	pngConfig.Format = "webp"
	if got := EstimateThumbnailSize(200, 100, pngConfig); got != 10000 {
		t.Fatalf("fallback estimate = %d, want 10000", got)
	}

	if IsThumbnailNeeded(make([]byte, 48000)) {
		t.Fatal("48,000-byte image should fit")
	}
	if !IsThumbnailNeeded(make([]byte, 48001)) {
		t.Fatal("48,001-byte image should need a thumbnail")
	}
}

func TestGenerateSafeThumbnail(t *testing.T) {
	encoded, err := GenerateSafeThumbnail(encodeJPEG(t, 600, 600))
	if err != nil {
		t.Fatalf("GenerateSafeThumbnail() error: %v", err)
	}
	if encoded == "" || len(encoded) > 43052 {
		t.Fatalf("safe thumbnail length = %d", len(encoded))
	}
	if _, err := GenerateSafeThumbnail([]byte("invalid")); err == nil {
		t.Fatal("GenerateSafeThumbnail() accepted invalid image data")
	}
}

func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fillImage(img)
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode PNG fixture: %v", err)
	}
	return buffer.Bytes()
}

func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fillImage(img)
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode JPEG fixture: %v", err)
	}
	return buffer.Bytes()
}

func fillImage(img *image.RGBA) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 100, A: 255})
		}
	}
}
