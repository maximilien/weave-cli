// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pdfWithJPEG(t *testing.T) []byte {
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

func TestPDFIngestionImageReporting(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "pdftotext"), []byte("#!/bin/sh\nprintf 'A catalogue page with image context' > \"$3\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	path := filepath.Join(root, "catalogue.pdf")
	if err := os.WriteFile(path, pdfWithJPEG(t), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, fail         string
		withReport, images bool
	}{
		{name: "success", withReport: true, images: true},
		{name: "without report", images: true},
		{name: "ordinary failure", fail: "rejected", withReport: true, images: true},
		{name: "memory failure", fail: "not enough memory", withReport: true, images: true},
		{name: "text only", withReport: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var classes []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/meta" {
					_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
					return
				}
				if r.Method != http.MethodPost || r.URL.Path != "/v1/objects" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
					return
				}
				var obj struct {
					Class      string
					Properties map[string]interface{}
				}
				if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
					t.Error(err)
				}
				classes = append(classes, obj.Class)
				if obj.Class == "WeaveImages" {
					if obj.Properties["image_data"] == "" {
						t.Error("missing image data")
					}
					if tc.fail != "" {
						http.Error(w, tc.fail, 500)
						return
					}
				}
				_, _ = io.WriteString(w, `{"id":"created"}`)
			}))
			defer server.Close()
			client := mustWeaviateClient(t, server.URL)
			var report *ProcessingReport
			if tc.withReport {
				report = &ProcessingReport{}
			}
			imageCollection := ""
			if tc.images {
				imageCollection = "WeaveImages"
			}
			out, diagnostics := captureCollectionListing(t, func() {
				if err := processPDFFile(context.Background(), client, "Docs", path, 1000, imageCollection, false, 0, 10, 1000, report); err != nil {
					t.Errorf("process PDF: %v", err)
				}
			})
			wantImages := 0
			if tc.images {
				wantImages = 1
			}
			if len(classes) != 1+wantImages || classes[0] != "Docs" || (tc.images && classes[1] != "WeaveImages") {
				t.Fatalf("created classes = %#v; output %s %s", classes, out, diagnostics)
			}
			if report != nil {
				if len(report.TextChunks) != 1 || len(report.Images) != wantImages {
					t.Fatalf("report = %#v", report)
				}
				if tc.images {
					img := report.Images[0]
					if img.ImageNumber != 1 || img.Filename != "catalogue_image_1.jpg" || img.SizeBytes <= 0 || img.Success != (tc.fail == "") || ((img.Error != "") != (tc.fail != "")) {
						t.Fatalf("image report = %#v", img)
					}
				}
			}
			if tc.fail != "" && !strings.Contains(diagnostics, "Failed to create PDF image document") {
				t.Fatalf("missing image failure: %s", diagnostics)
			}
			if tc.fail == "not enough memory" && !strings.Contains(diagnostics, "Try reducing batch size") {
				t.Fatalf("missing memory advice: %s", diagnostics)
			}
			if !strings.Contains(out+diagnostics, "PDF processed") {
				t.Fatalf("missing completion: %s %s", out, diagnostics)
			}
		})
	}
}
