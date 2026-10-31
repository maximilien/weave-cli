// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pdfImageFixture(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 12), G: uint8(y * 12), B: 40, A: 255})
		}
	}
	var jpegData bytes.Buffer
	if err := jpeg.Encode(&jpegData, img, nil); err != nil {
		t.Fatal(err)
	}
	stream := "q 16 0 0 16 0 0 cm /Im0 Do Q\n"
	objects := [][]byte{
		[]byte(`<< /Type /Catalog /Pages 2 0 R >>`),
		[]byte(`<< /Type /Pages /Kids [3 0 R] /Count 1 >>`),
		[]byte(`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >>`),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream)),
		append(append([]byte(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 16 /Height 16 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n", jpegData.Len())), jpegData.Bytes()...), []byte("\nendstream")...),
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n", i+1)
		out.Write(obj)
		out.WriteString("\nendobj\n")
	}
	start := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), start)
	return out.Bytes()
}

func TestExtractPDFContentTextImagesAndFiltering(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "pdftotext"), []byte("#!/bin/sh\nprintf 'A catalogue page containing an image and descriptive context' > \"$3\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	path := filepath.Join(root, "catalogue.pdf")
	if err := os.WriteFile(path, pdfImageFixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		filter     bool
		min        int
		wantImages int
	}{
		{name: "image included", wantImages: 1}, {name: "small image filtered", filter: true, min: 1000000, wantImages: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			texts, images, err := ExtractPDFContent(path, 1000, tc.filter, tc.min, 2000, true)
			if err != nil || len(texts) != 1 || len(images) != tc.wantImages {
				t.Fatalf("extraction texts=%#v images=%#v err=%v", texts, images, err)
			}
			if texts[0].SourcePDF != path || !strings.Contains(texts[0].Content, "catalogue page") || texts[0].Metadata["filename"] != "catalogue.pdf" {
				t.Fatalf("text = %#v", texts[0])
			}
			if tc.wantImages > 0 {
				img := images[0]
				data, err := base64.StdEncoding.DecodeString(img.ImageData)
				if err != nil || len(data) == 0 {
					t.Fatalf("image data error=%v", err)
				}
				if img.Metadata["source_pdf"] != path || img.PageNumber != 1 || !strings.Contains(img.SurroundingText, "catalogue page") {
					t.Fatalf("image = %#v", img)
				}
			}
		})
	}
	if _, _, err := ExtractPDFContent(filepath.Join(root, "missing.pdf"), 1000, false, 0, 2000, true); err == nil {
		t.Fatal("missing PDF accepted")
	}
}

func TestFallbackPDFImagesExtractionAndFiltering(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, nil); err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(root, "fixture.jpg")
	if err := os.WriteFile(fixture, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PDF_FIXTURE_IMAGE", fixture)
	script := "#!/bin/sh\n/bin/cp \"$PDF_FIXTURE_IMAGE\" \"$4-001.jpg\"\n"
	if err := os.WriteFile(filepath.Join(bin, "pdfimages"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	for _, tc := range []struct {
		name      string
		filter    bool
		min, want int
	}{
		{name: "extract", want: 1}, {name: "filter small", filter: true, min: 1000000, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := extractImagesWithFallback(filepath.Join(root, "fixture.pdf"), t.TempDir(), tc.filter, tc.min, true)
			if err != nil || len(out) != tc.want {
				t.Fatalf("images=%#v err=%v", out, err)
			}
			if tc.want == 1 {
				if out[0].Metadata["source_pdf"] != filepath.Join(root, "fixture.pdf") || out[0].ImageData == "" || out[0].Metadata["filename"] != "fixture_image_1.jpg" {
					t.Fatalf("image=%#v", out[0])
				}
			}
		})
	}
	if err := os.WriteFile(filepath.Join(bin, "pdfimages"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if out, err := extractImagesWithFallback(filepath.Join(root, "fixture.pdf"), t.TempDir(), false, 0, true); err != nil || len(out) != 0 {
		t.Fatalf("empty extraction images=%#v err=%v", out, err)
	}
	if _, err := extractImagesWithFallback(filepath.Join(root, "fixture.pdf"), filepath.Join(root, "missing-directory"), false, 0, true); err == nil || !strings.Contains(err.Error(), "failed to walk") {
		t.Fatalf("missing extraction directory error=%v", err)
	}
	if err := os.WriteFile(filepath.Join(bin, "pdfimages"), []byte("#!/bin/sh\nexit 2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	out, err := extractImagesWithFallback(filepath.Join(root, "fixture.pdf"), t.TempDir(), false, 0, true)
	if err != nil || len(out) != 0 {
		t.Fatalf("failed extractor images=%#v err=%v", out, err)
	}
}
