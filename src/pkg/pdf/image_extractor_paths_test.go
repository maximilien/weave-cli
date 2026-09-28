package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageExtractionErrorAndMetadataPaths(t *testing.T) {
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
