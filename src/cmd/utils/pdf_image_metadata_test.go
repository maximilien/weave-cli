// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/pdf"
)

// Minimal TIFF metadata keeps extraction deterministic without an OCR service.
func extractedImageEXIFFixture() []byte {
	data := make([]byte, 220)
	copy(data, "II")
	binary.LittleEndian.PutUint16(data[2:], 42)
	binary.LittleEndian.PutUint32(data[4:], 8)
	type entry struct {
		tag, kind    uint16
		count, value uint32
	}
	entries := []entry{{0x010f, 2, 6, 86}, {0x0110, 2, 7, 92}, {0x0112, 3, 1, 1}, {0x0132, 2, 20, 100}, {0x8769, 4, 1, 128}}
	write := func(offset int, entries []entry) {
		binary.LittleEndian.PutUint16(data[offset:], uint16(len(entries)))
		for i, e := range entries {
			p := offset + 2 + i*12
			binary.LittleEndian.PutUint16(data[p:], e.tag)
			binary.LittleEndian.PutUint16(data[p+2:], e.kind)
			binary.LittleEndian.PutUint32(data[p+4:], e.count)
			binary.LittleEndian.PutUint32(data[p+8:], e.value)
		}
	}
	write(8, entries)
	copy(data[86:], "Canon\x00")
	copy(data[92:], "EOS 80\x00")
	copy(data[100:], "2026:10:26 10:00:00\x00")
	write(128, []entry{{0xa002, 4, 1, 6000}, {0xa003, 4, 1, 4000}})
	return data
}

func TestExtractedPDFImageMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, collection, format string
		exif, context            bool
	}{
		{name: "default format", collection: "WeaveImages"},
		{name: "explicit format and context", collection: "WeaveImages", format: "jpg", exif: true, context: true},
		{name: "dotted format", collection: "WeaveImages", format: ".tiff"},
		{name: "legacy", collection: "RagMeImages", format: "png", context: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte("image without EXIF")
			if tc.exif {
				raw = extractedImageEXIFFixture()
			}
			source := filepath.Join(t.TempDir(), "catalogue.pdf")
			metadata := map[string]interface{}{"source_pdf": source, "pdf_page": 3}
			if tc.format != "" {
				metadata["image_format"] = tc.format
			}
			original := map[string]interface{}{}
			for k, v := range metadata {
				original[k] = v
			}
			input := pdf.PDFImageData{ID: "image-id", ImageData: base64.StdEncoding.EncodeToString(raw), Image: "data:image/png;base64,fixture", URL: "legacy://image", Metadata: metadata}
			if tc.context {
				input.Caption = "A camera"
				input.SurroundingText = "Auction catalogue"
				input.SectionHeading = "Cameras"
			}
			doc, warnings, err := processExtractedPDFImage(context.Background(), input, tc.collection, source, 7)
			if err != nil {
				t.Fatal(err)
			}
			if doc.ID != input.ID || doc.ImageData != input.ImageData || doc.Image != input.Image {
				t.Fatalf("image payload changed: %#v", doc)
			}
			if !reflect.DeepEqual(metadata, original) {
				t.Fatalf("input metadata mutated: %#v", metadata)
			}
			if len(warnings) != 1 || !strings.Contains(warnings[0], "OCR extraction failed") {
				t.Fatalf("OCR unavailable warning = %#v", warnings)
			}
			if tc.collection == "RagMeImages" {
				if doc.URL != input.URL || !reflect.DeepEqual(doc.Metadata, metadata) {
					t.Fatalf("legacy metadata changed: %#v", doc)
				}
				return
			}
			ext := tc.format
			if ext == "" {
				ext = "png"
			}
			ext = "." + strings.TrimPrefix(ext, ".")
			wantFilename := "catalogue_image_7" + ext
			for key, want := range map[string]interface{}{"filename": wantFilename, "original_filename": wantFilename, "source_pdf": source, "storage_path": source, "pdf_filename": "catalogue.pdf", "pdf_page": 3, "pdf_image_index": 7, "file_size": len(raw), "image_size": len(raw), "type": "image", "total_chunks": 1, "chunk_index": 0, "is_chunked": false} {
				if doc.Metadata[key] != want {
					t.Errorf("%s = %#v; want %#v", key, doc.Metadata[key], want)
				}
			}
			if doc.URL != "pdf://catalogue.pdf/image_7" {
				t.Fatalf("URL = %s", doc.URL)
			}
			if !reflect.DeepEqual(doc.Metadata["chunk_sizes"], []int{len(raw)}) {
				t.Fatalf("chunk_sizes = %#v", doc.Metadata["chunk_sizes"])
			}
			for _, key := range []string{"date_added", "creation_date", "modified_date", "processing_timestamp"} {
				value, ok := doc.Metadata[key].(string)
				if !ok {
					t.Fatalf("%s missing", key)
				}
				if _, err := time.Parse(time.RFC3339, value); err != nil {
					t.Errorf("%s invalid: %v", key, err)
				}
			}
			for key, want := range map[string]string{"caption": input.Caption, "surrounding_text": input.SurroundingText, "section_heading": input.SectionHeading} {
				got, exists := doc.Metadata[key]
				if tc.context {
					if got != want {
						t.Errorf("%s = %#v", key, got)
					}
				} else if exists {
					t.Errorf("unexpected empty metadata %s", key)
				}
			}
			if tc.exif {
				for key, want := range map[string]interface{}{"exif_make": "Canon", "exif_model": "EOS 80", "exif_orientation": 1, "exif_width": 6000, "exif_height": 4000} {
					if doc.Metadata[key] != want {
						t.Errorf("%s = %#v; want %#v", key, doc.Metadata[key], want)
					}
				}
			}
		})
	}
}

func TestExtractedPDFImageRejectsMalformedBase64(t *testing.T) {
	doc, warnings, err := processExtractedPDFImage(context.Background(), pdf.PDFImageData{ImageData: "not-base64!"}, "WeaveImages", "source.pdf", 1)
	if err == nil || !strings.Contains(err.Error(), "failed to decode image data") || doc.ID != "" || warnings != nil {
		t.Fatalf("result = %#v, %#v, %v", doc, warnings, err)
	}
}
