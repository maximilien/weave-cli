// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package image

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoOpOCRExtractorContract(t *testing.T) {
	extractor := &NoOpOCRExtractor{}
	for name, run := range map[string]func() (*OCRData, error){
		"bytes":    func() (*OCRData, error) { return extractor.ExtractFromBytes([]byte("image")) },
		"file":     func() (*OCRData, error) { return extractor.ExtractFromFile("image.png") },
		"language": func() (*OCRData, error) { return extractor.ExtractWithLanguage("image.png", "eng") },
	} {
		t.Run(name, func(t *testing.T) {
			data, err := run()
			if data != nil || err == nil || !strings.Contains(err.Error(), "Tesseract not installed") {
				t.Fatalf("unexpected no-op result: data=%#v err=%v", data, err)
			}
		})
	}
}

func TestOCRExtractorFallback(t *testing.T) {
	if extractor := GetOCRExtractor(); extractor == nil {
		t.Fatal("expected fallback extractor")
	} else if _, ok := extractor.(*NoOpOCRExtractor); !ok {
		t.Fatalf("expected no-op fallback, got %T", extractor)
	}
	if extractor, err := NewTesseractExtractor(); extractor != nil || err == nil {
		t.Fatalf("expected unavailable Tesseract result, extractor=%T err=%v", extractor, err)
	}
}

func TestOCRDataHelpers(t *testing.T) {
	empty := &OCRData{}
	if empty.GetTextSummary(10) != "" || empty.WordCount() != 0 || !empty.IsEmpty() {
		t.Fatalf("unexpected empty OCR behavior: %#v", empty)
	}

	data := &OCRData{Text: "  alpha\n beta\r\n gamma   delta  ", HasText: true, Language: "eng", Confidence: 98.5}
	if got := data.GetTextSummary(50); got != "alpha beta gamma delta" {
		t.Fatalf("unexpected full summary: %q", got)
	}
	if got := data.GetTextSummary(10); got != "alpha beta..." {
		t.Fatalf("unexpected truncated summary: %q", got)
	}
	if got := data.WordCount(); got != 4 {
		t.Fatalf("unexpected word count: %d", got)
	}
	if data.IsEmpty() {
		t.Fatal("expected OCR data to be non-empty")
	}
	data.Text = ""
	if !data.IsEmpty() {
		t.Fatal("expected blank OCR text to be empty")
	}
}

func TestExtractEXIFWithoutMetadata(t *testing.T) {
	data, err := ExtractEXIFFromBytes([]byte("not an image"))
	if err != nil || data == nil || !data.IsEmpty() {
		t.Fatalf("expected empty EXIF for undecodable bytes, data=%#v err=%v", data, err)
	}

	path := filepath.Join(t.TempDir(), "plain.bin")
	if err := os.WriteFile(path, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = ExtractEXIF(path)
	if err != nil || data == nil || !data.IsEmpty() {
		t.Fatalf("expected empty EXIF for undecodable file, data=%#v err=%v", data, err)
	}
	if _, err := ExtractEXIF(filepath.Join(t.TempDir(), "missing.jpg")); err == nil || !strings.Contains(err.Error(), "failed to open") {
		t.Fatalf("expected missing file error, got %v", err)
	}
}
