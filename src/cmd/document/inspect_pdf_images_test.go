// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectPDFImagePreviewLimit(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "pdftotext"), "#!/bin/sh\nprintf 'A fixture document containing six images for inspection.' > \"$3\"\n")
	t.Setenv("PATH", bin)
	path := filepath.Join(root, "images.pdf")
	if err := os.WriteFile(path, inspectionImagesPDF(t), 0600); err != nil {
		t.Fatal(err)
	}
	for _, summary := range []bool{false, true} {
		t.Run(fmt.Sprintf("summary=%t", summary), func(t *testing.T) {
			previous := inspectSummary
			inspectSummary = summary
			t.Cleanup(func() { inspectSummary = previous })
			output, err := os.CreateTemp(t.TempDir(), "preview-*.log")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = output.Close() })
			old := os.Stdout
			os.Stdout = output
			t.Cleanup(func() { os.Stdout = old })
			err = inspectPDF(path)
			os.Stdout = old
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			for _, want := range []string{"Images:      6", "Avg Image:", "--image-collection <image-col>"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %q: %s", want, text)
				}
			}
			if summary {
				if strings.Contains(text, "Images Preview") || strings.Contains(text, "Image 1:") {
					t.Fatalf("summary includes detailed images: %s", text)
				}
				return
			}
			for _, want := range []string{"showing first 5", "Image 5:", "Format:", "... and 1 more images"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %q: %s", want, text)
				}
			}
			if strings.Contains(text, "Image 6:") {
				t.Fatalf("preview exceeded five images: %s", text)
			}
		})
	}
}

func inspectionImagesPDF(t *testing.T) []byte {
	t.Helper()
	var resources, stream strings.Builder
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&resources, "/Im%d %d 0 R ", i, i+5)
		fmt.Fprintf(&stream, "q 16 0 0 16 0 0 cm /Im%d Do Q\n", i)
	}
	objects := [][]byte{[]byte(`<< /Type /Catalog /Pages 2 0 R >>`), []byte(`<< /Type /Pages /Kids [3 0 R] /Count 1 >>`), []byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << %s >> >> /Contents 4 0 R >>", resources.String())), []byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", stream.Len(), stream.String()))}
	for i := 0; i < 6; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				img.Set(x, y, color.RGBA{R: uint8(20 + i*30), G: uint8(x * 10), B: uint8(y * 10), A: 255})
			}
		}
		var encoded bytes.Buffer
		if err := jpeg.Encode(&encoded, img, nil); err != nil {
			t.Fatal(err)
		}
		object := []byte(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 16 /Height 16 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n", encoded.Len()))
		object = append(object, encoded.Bytes()...)
		object = append(object, []byte("\nendstream")...)
		objects = append(objects, object)
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
