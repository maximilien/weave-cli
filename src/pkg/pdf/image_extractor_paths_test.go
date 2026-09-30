package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageExtractionErrorAndMetadataPaths(t *testing.T) {
	fallbackDir := t.TempDir()
	if _, err := extractImagesWithFallback(filepath.Join(fallbackDir, "missing.pdf"), fallbackDir, false, 0, true); err != nil {
		t.Fatalf("fallback extraction: %v", err)
	}
	binDir := t.TempDir()
	fakePDFImages := filepath.Join(binDir, "pdfimages")
	if err := os.WriteFile(fakePDFImages, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	if _, err := extractImagesWithFallback(filepath.Join(fallbackDir, "missing.pdf"), fallbackDir, false, 0, true); err != nil {
		t.Fatalf("empty fallback extraction: %v", err)
	}
	t.Setenv("PATH", filepath.Join(fallbackDir, "no-tools"))
	if _, err := extractImagesWithFallback(filepath.Join(fallbackDir, "missing.pdf"), fallbackDir, false, 0, false); err != nil {
		t.Fatalf("missing-tool fallback extraction: %v", err)
	}
	if _, err := extractPDFImages(filepath.Join(t.TempDir(), "missing.pdf"), false, 0, true); err == nil {
		t.Fatal("expected missing PDF extraction error")
	}
	if got := extractEXIFData([]byte("not-an-image")); got["format"] != "unknown" || got["width"] != 0 {
		t.Fatalf("unexpected placeholder EXIF data: %#v", got)
	}
	imagePath := filepath.Join(t.TempDir(), "catalogue_003_Im0.jpg")
	if err := os.WriteFile(imagePath, []byte("image bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := processExtractedImage(imagePath, "/tmp/catalogue.pdf", 0)
	if err != nil {
		t.Fatal(err)
	}
	if data.PageNumber != 3 || data.ImageIndex != 0 || data.Metadata["page_number"] != 3 {
		t.Fatalf("unexpected extracted metadata: %#v", data.Metadata)
	}
}
