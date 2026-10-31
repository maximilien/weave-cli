// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pipelineMinimalPDF() []byte {
	objects := []string{`<< /Type /Catalog /Pages 2 0 R >>`, `<< /Type /Pages /Kids [3 0 R] /Count 1 >>`, `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << >> /Contents 4 0 R >>`, "<< /Length 0 >>\nstream\n\nendstream"}
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

func TestProcessorPDFExtractionAndEmbedding(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "pdftotext"), []byte("#!/bin/sh\nprintf 'A catalogue page with text for embedding' > \"$3\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	path := filepath.Join(root, "catalogue.pdf")
	if err := os.WriteFile(path, pipelineMinimalPDF(), 0600); err != nil {
		t.Fatal(err)
	}
	file := FileInfo{Path: path, Type: FileTypePDF, Hash: "fixture-hash"}
	for _, tc := range []struct {
		name    string
		status  int
		missing bool
		wantErr string
	}{
		{name: "success", status: http.StatusOK},
		{name: "embedding failure", status: http.StatusInternalServerError, wantErr: "failed to generate embedding"},
		{name: "missing PDF", status: http.StatusOK, missing: true, wantErr: "failed to extract PDF content"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := file
			if tc.missing {
				input.Path = filepath.Join(root, "missing.pdf")
			}
			processor := NewProcessor(&recordingPipelineVDB{}, newPipelineLLM(t, tc.status), &IngestOptions{Metadata: map[string]string{"campaign": "fixture"}}, NewProgressTracker(false, false))
			docs, err := processor.processPDF(context.Background(), input)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) || docs != nil {
					t.Fatalf("result = %#v, %v", docs, err)
				}
				return
			}
			if err != nil || len(docs) != 1 {
				t.Fatalf("result = %#v, %v", docs, err)
			}
			doc := docs[0]
			if doc.ID == "" || doc.Text != "A catalogue page with text for embedding" || doc.Content != doc.Text || len(doc.Embedding) != 3 {
				t.Fatalf("document = %#v", doc)
			}
			for key, want := range map[string]interface{}{"campaign": "fixture", "source_file": path, "file_type": "pdf", "file_hash": "fixture-hash", "page_number": 1} {
				if doc.Metadata[key] != want {
					t.Errorf("metadata[%s] = %#v; want %#v", key, doc.Metadata[key], want)
				}
			}
		})
	}
}
