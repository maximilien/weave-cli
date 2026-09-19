// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFallbackPDFContentVariants(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		pages    int
		contains string
	}{
		{name: "known pages", size: 1234, pages: 3, contains: "Pages: 3"},
		{name: "small", size: 999, contains: "Small PDF file"},
		{name: "moderate", size: 50_000, contains: "moderate content"},
		{name: "large", size: 100_000, contains: "Large PDF Document"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := generateFallbackPDFContent("fixture.pdf", test.size, test.pages)
			if !strings.Contains(got, test.contains) || !strings.Contains(got, "fixture.pdf") {
				t.Fatalf("fallback content = %q", got)
			}
		})
	}
	if max(7, 3) != 7 || max(2, 5) != 5 {
		t.Fatal("max returned an incorrect value")
	}
}

func TestRenderingCommandDetection(t *testing.T) {
	if looksLikeRenderingCommands("short text") {
		t.Fatal("short text detected as rendering commands")
	}
	rendering := strings.Repeat("prefix ", 10) + " cm\nvalue RG value rg\nmore\nq\n"
	if !looksLikeRenderingCommands(rendering) {
		t.Fatal("rendering operators were not detected")
	}
	if looksLikeRenderingCommands(strings.Repeat("ordinary prose ", 20)) {
		t.Fatal("ordinary prose detected as rendering commands")
	}
}

func TestImageContextFromPageText(t *testing.T) {
	pageText := "2. RESULTS\nFigure 1: Accuracy by model\nAdditional caption detail\n" + strings.Repeat("context ", 50)
	image := &PDFImageData{PageNumber: 2, ImageIndex: 0, OCRText: strings.Repeat("ocr ", 30)}
	enrichImageWithContext(image, map[int]string{2: pageText}, 80)

	if !strings.Contains(image.Caption, "Figure 1") {
		t.Fatalf("caption = %q", image.Caption)
	}
	if image.SectionHeading != "2. RESULTS" {
		t.Fatalf("section heading = %q", image.SectionHeading)
	}
	if len(image.SurroundingText) != 80 || !strings.HasSuffix(image.SurroundingText, "...") {
		t.Fatalf("surrounding text = %q", image.SurroundingText)
	}
	for _, key := range []string{"caption", "surrounding_text", "section_heading", "ocr_content"} {
		if _, ok := image.Metadata[key]; !ok {
			t.Errorf("metadata missing %q: %#v", key, image.Metadata)
		}
	}

	withoutCaption := &PDFImageData{PageNumber: 1, Metadata: map[string]interface{}{}}
	enrichImageWithContext(withoutCaption, map[int]string{1: "A meaningful introductory sentence for this section."}, 0)
	if withoutCaption.Caption != "" || withoutCaption.SectionHeading == "" {
		t.Fatalf("context without caption = %#v", withoutCaption)
	}

	enrichImageWithContext(nil, nil, 10)
}

func TestImageContextFallsBackToOCR(t *testing.T) {
	ocr := "tiny\nThis is the first meaningful OCR caption line\nremaining OCR content"
	image := &PDFImageData{PageNumber: 4, OCRText: ocr}
	enrichImageWithContext(image, map[int]string{}, 35)
	if image.Caption != "This is the first meaningful OCR caption line" {
		t.Fatalf("OCR caption = %q", image.Caption)
	}
	if len(image.SurroundingText) != 35 || image.Metadata["ocr_content"] != image.SurroundingText {
		t.Fatalf("OCR context = %#v", image)
	}

	empty := &PDFImageData{PageNumber: 5}
	enrichImageWithContext(empty, nil, 20)
	if empty.Metadata != nil {
		t.Fatalf("empty OCR metadata = %#v", empty.Metadata)
	}
}

func TestCaptionHeadingAndTruncationHelpers(t *testing.T) {
	if got := extractCaptionFromText("preface\nChart: Revenue\nby quarter", 8); got != "Chart: Revenue by quarter" {
		t.Fatalf("caption = %q", got)
	}
	if got := extractCaptionFromText("no caption here", 0); got != "" {
		t.Fatalf("missing caption = %q", got)
	}
	if got := extractCaptionFromOCR("short\nlong enough caption"); got != "long enough caption" {
		t.Fatalf("OCR caption = %q", got)
	}
	if got := extractCaptionFromOCR("a\nb"); got != "" {
		t.Fatalf("short OCR caption = %q", got)
	}

	if got := extractSectionHeading("\nINTRODUCTION\nbody"); got != "INTRODUCTION" {
		t.Fatalf("all-caps heading = %q", got)
	}
	if got := extractSectionHeading("\nA substantial opening sentence for the section\nbody"); got != "A substantial opening sentence for the section" {
		t.Fatalf("fallback heading = %q", got)
	}
	if got := extractSectionHeading("tiny\ntext"); got != "" {
		t.Fatalf("missing heading = %q", got)
	}

	if got := truncateText("  short text  ", 20); got != "short text" {
		t.Fatalf("short truncation = %q", got)
	}
	if got := truncateText("abcdefghij", 7); got != "abcd..." {
		t.Fatalf("long truncation = %q", got)
	}
}

func TestPdftotextProtocolAndPageSplitting(t *testing.T) {
	binDir := t.TempDir()
	script := filepath.Join(binDir, "pdftotext")
	contents := "#!/bin/sh\nprintf 'First page content\\fSecond page content' > \"$3\"\n"
	if err := os.WriteFile(script, []byte(contents), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	input := filepath.Join(t.TempDir(), "fixture.pdf")
	if err := os.WriteFile(input, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	text, pages, err := extractTextWithPdftotext(input)
	if err != nil || pages != 2 || !strings.Contains(text, "Second page") {
		t.Fatalf("extractTextWithPdftotext() = (%q, %d, %v)", text, pages, err)
	}
	pageText, err := extractTextByPage(input)
	if err != nil || pageText[1] != "First page content" || pageText[2] != "Second page content" {
		t.Fatalf("extractTextByPage() = (%#v, %v)", pageText, err)
	}
	text, pages, err = extractTextFromPDF(input)
	if err != nil || pages != 2 || text == "" {
		t.Fatalf("extractTextFromPDF() = (%q, %d, %v)", text, pages, err)
	}
}

func TestPDFExtractionFailurePaths(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.pdf")
	metadata, err := extractPDFMetadata(missing)
	if err == nil || metadata["filename"] != "missing.pdf" || metadata["type"] != "pdf" {
		t.Fatalf("extractPDFMetadata(missing) = (%#v, %v)", metadata, err)
	}
	if _, _, err := extractTextWithPdfcpu(missing); err == nil {
		t.Fatal("extractTextWithPdfcpu(missing) expected error")
	}
	if _, _, err := ExtractPDFContent(missing, 100, true, 100, 100, true); err == nil {
		t.Fatal("ExtractPDFContent(missing) expected error")
	}
	if _, err := extractTextByPage(missing); err == nil {
		t.Fatal("extractTextByPage(missing) expected error")
	}
}
