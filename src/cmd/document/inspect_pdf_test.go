// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectPDFSummaryAndTextPreview(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(bin, "pdftotext"), "#!/bin/sh\nprintf '%s' \"$INSPECTION_TEXT\" > \"$3\"\n")
	t.Setenv("PATH", bin)
	paragraph := strings.Repeat("preview text ", 17) // 221 characters, 220 after trimming.
	t.Setenv("INSPECTION_TEXT", strings.Join([]string{paragraph, paragraph, paragraph, paragraph}, "\n\n"))
	path := filepath.Join(root, "fixture.pdf")
	if err := os.WriteFile(path, inspectionPDFFixture(), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, summary := range []bool{false, true} {
		t.Run(fmt.Sprintf("summary=%t", summary), func(t *testing.T) {
			previousSize, previousSummary := inspectChunkSize, inspectSummary
			inspectChunkSize, inspectSummary = 250, summary
			t.Cleanup(func() { inspectChunkSize, inspectSummary = previousSize, previousSummary })
			output, err := os.CreateTemp(t.TempDir(), "inspection-*.log")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = output.Close() })
			previousOutput := os.Stdout
			os.Stdout = output
			t.Cleanup(func() { os.Stdout = previousOutput })
			err = runInspect(InspectCmd, []string{path})
			os.Stdout = previousOutput
			if err != nil {
				t.Fatalf("runInspect(PDF) = %v", err)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			for _, want := range []string{"Text Chunks: 4", "Images:      0", "Avg Chunk:   220 characters", "Inspection complete!", "weave docs create <collection> fixture.pdf"} {
				if !strings.Contains(text, want) {
					t.Fatalf("inspection output missing %q:\n%s", want, text)
				}
			}
			if summary {
				if strings.Contains(text, "Text Chunks Preview") || strings.Contains(text, "Chunk 1 (page") {
					t.Fatalf("summary includes a detailed preview:\n%s", text)
				}
				return
			}
			for _, want := range []string{"showing first 3", "Chunk 3 (page 3)", strings.TrimSpace(paragraph)[:200] + "...", "... and 1 more text chunks"} {
				if !strings.Contains(text, want) {
					t.Fatalf("detailed output missing %q:\n%s", want, text)
				}
			}
			if strings.Contains(text, "Chunk 4 (page") || strings.Contains(text, strings.TrimSpace(paragraph)+"\n") {
				t.Fatalf("preview limit was not applied:\n%s", text)
			}
		})
	}
}

// A valid image-free PDF keeps inspection independent of external PDF fixtures.
func inspectionPDFFixture() []byte {
	objects := []string{
		`<< /Type /Catalog /Pages 2 0 R >>`,
		`<< /Type /Pages /Kids [3 0 R] /Count 1 >>`,
		`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << >> /Contents 4 0 R >>`,
		"<< /Length 0 >>\nstream\n\nendstream",
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, object := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	start := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), start)
	return out.Bytes()
}
