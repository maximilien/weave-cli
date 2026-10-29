// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalPDF creates a valid one-page document without external fixtures.
func minimalPDF() []byte {
	objects := []string{
		`<< /Type /Catalog /Pages 2 0 R >>`,
		`<< /Type /Pages /Kids [3 0 R] /Count 1 >>`,
		`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << >> /Contents 4 0 R >>`,
		"<< /Length 0 >>\nstream\n\nendstream",
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	start := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), start)
	return out.Bytes()
}

func TestPDFIngestionChunkReporting(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	toolsDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(toolsDir, 0700); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf 'Alpha line for fixture\\n\\nBeta line for fixture\\n\\nGamma line for fixture\\n' > \"$3\"\n"
	if err := os.WriteFile(filepath.Join(toolsDir, "pdftotext"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", toolsDir)
	pdfPath := filepath.Join(root, "catalogue.pdf")
	if err := os.WriteFile(pdfPath, minimalPDF(), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                   string
		failSecond, withReport bool
	}{
		{name: "all chunks", withReport: true}, {name: "partial failure", withReport: true, failSecond: true}, {name: "without report"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var objects []map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/meta" {
					_, _ = io.WriteString(w, `{ "version":"1.27.0", "modules":{} }`)
					return
				}
				if r.Method != http.MethodPost || r.URL.Path != "/v1/objects" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
					return
				}
				var obj map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
					t.Error(err)
				}
				objects = append(objects, obj)
				if tc.failSecond && len(objects) == 2 {
					http.Error(w, "fixture rejected chunk", 500)
					return
				}
				_, _ = io.WriteString(w, `{"id":"created"}`)
			}))
			defer server.Close()
			client := mustWeaviateClient(t, server.URL)
			var report *ProcessingReport
			if tc.withReport {
				report = &ProcessingReport{}
			}
			out, diagnostics := captureCollectionListing(t, func() {
				if err := processPDFFile(context.Background(), client, "Docs", pdfPath, 24, "", false, 0, 10, 1000, report); err != nil {
					t.Errorf("processPDFFile: %v", err)
				}
			})
			if len(objects) < 2 {
				t.Fatalf("created only %d chunks: %s %s", len(objects), out, diagnostics)
			}
			if !strings.Contains(out+diagnostics, "PDF processed") {
				t.Fatalf("missing completion: %s %s", out, diagnostics)
			}
			for _, obj := range objects {
				if obj["class"] != "Docs" {
					t.Fatalf("wrong collection: %#v", obj)
				}
				props, ok := obj["properties"].(map[string]interface{})
				if !ok || props["content"] == "" {
					t.Fatalf("missing chunk content: %#v", obj)
				}
			}
			if tc.withReport {
				if len(report.TextChunks) != len(objects) {
					t.Fatalf("report count=%d, objects=%d", len(report.TextChunks), len(objects))
				}
				for i, chunk := range report.TextChunks {
					want := !(tc.failSecond && i == 1)
					if chunk.ChunkNumber != i+1 || chunk.Success != want || chunk.SizeBytes <= 0 || (want && chunk.Error != "") || (!want && chunk.Error == "") {
						t.Errorf("chunk %d = %#v", i, chunk)
					}
				}
			}
			if tc.failSecond && !strings.Contains(diagnostics, "Failed to create PDF text document") {
				t.Fatalf("missing failure diagnostic: %s", diagnostics)
			}
		})
	}
}

func TestPDFIngestionRejectsMissingInput(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("PATH", root)
	client := mustWeaviateClient(t, "http://127.0.0.1:1")
	report := &ProcessingReport{}
	err := processPDFFile(context.Background(), client, "Docs", filepath.Join(root, "missing.pdf"), 100, "", false, 0, 10, 1000, report)
	if err == nil || len(report.TextChunks) != 0 || len(report.Images) != 0 {
		t.Fatalf("missing file result = %v, %#v", err, report)
	}
}
