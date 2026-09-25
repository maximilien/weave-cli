// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
)

func TestGenericTextIngestionSequentialAndParallel(t *testing.T) {
	ctx := context.Background()
	cfg := &config.VectorDBConfig{
		Type:        config.VectorDBTypeMock,
		Enabled:     true,
		Collections: []config.Collection{{Name: "WeaveDocs", Type: "text"}},
	}
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "guide.txt")
	if err := os.WriteFile(path, []byte("first line\nsecond line\nthird line\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := client.GetCollectionCount(ctx, "WeaveDocs")
	if err != nil {
		t.Fatal(err)
	}
	if err := processTextFileGeneric(ctx, client, "WeaveDocs", path, 15, 1, ""); err != nil {
		t.Fatalf("sequential ingestion: %v", err)
	}
	if err := processTextFileGeneric(ctx, client, "WeaveDocs", path, 15, 2, ""); err != nil {
		t.Fatalf("parallel ingestion: %v", err)
	}
	docs, err := client.ListDocuments(ctx, "WeaveDocs", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != int(before)+6 {
		t.Fatalf("document count after both ingestion modes = %d, want %d", len(docs), before+6)
	}
	for _, doc := range docs[int(before):] {
		if doc.Metadata["original_filename"] != "guide.txt" || doc.Metadata["total_chunks"] != 3 || doc.Metadata["is_chunked"] != true {
			t.Fatalf("ingested document metadata = %+v", doc.Metadata)
		}
		if !strings.Contains(doc.URL, "guide.txt#chunk-") {
			t.Fatalf("document URL = %q", doc.URL)
		}
	}
	if err := processTextFileGeneric(ctx, client, "Missing", path, 15, 1, ""); err == nil {
		t.Fatal("missing collection accepted by sequential ingestion")
	}
	if err := processTextFileGeneric(ctx, client, "Missing", path, 15, 2, ""); err == nil {
		t.Fatal("missing collection accepted by parallel ingestion")
	}
	if err := processTextFileGeneric(ctx, client, "WeaveDocs", filepath.Join(t.TempDir(), "missing.txt"), 15, 1, ""); err == nil {
		t.Fatal("missing input file accepted")
	}
}

func TestLegacyMockTextAndImageIngestion(t *testing.T) {
	ctx := context.Background()
	client := mock.NewClient(&config.MockConfig{Collections: []config.MockCollection{
		{Name: "NewDocs"}, {Name: "RagMeDocs"}, {Name: "NewImages"}, {Name: "RagMeImages"},
	}})
	root := t.TempDir()
	textPath := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(textPath, []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, collection := range []string{"NewDocs", "RagMeDocs"} {
		processTextFileMock(ctx, client, collection, textPath, 7)
		docs, err := client.ListDocuments(ctx, collection, 10)
		if err != nil || len(docs) != 2 {
			t.Fatalf("%s text documents = %+v, %v", collection, docs, err)
		}
		if docs[0].Content != "alpha" || docs[1].Content != "beta" {
			t.Fatalf("%s text chunks = %+v", collection, docs)
		}
		if docs[0].Metadata["filename"] != "notes.txt" {
			t.Fatalf("%s metadata = %+v", collection, docs[0].Metadata)
		}
		if collection == "NewDocs" && docs[0].Metadata["total_chunks"] != 2 {
			t.Fatalf("new schema metadata = %+v", docs[0].Metadata)
		}
		if collection == "RagMeDocs" && docs[0].Metadata["file_size"] != int64(11) {
			t.Fatalf("legacy schema metadata = %+v", docs[0].Metadata)
		}
	}

	imagePath := filepath.Join(root, "image.png")
	imageBytes := []byte("image fixture")
	if err := os.WriteFile(imagePath, imageBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, collection := range []string{"NewImages", "RagMeImages"} {
		processImageFileMock(ctx, client, collection, imagePath)
		docs, err := client.ListDocuments(ctx, collection, 10)
		if err != nil || len(docs) != 1 {
			t.Fatalf("%s image documents = %+v, %v", collection, docs, err)
		}
		if docs[0].ImageData != base64.StdEncoding.EncodeToString(imageBytes) || docs[0].Metadata["image_size"] != len(imageBytes) {
			t.Fatalf("%s image document = %+v", collection, docs[0])
		}
	}

	processTextFileMock(ctx, client, "Missing", textPath, 7)
	processImageFileMock(ctx, client, "Missing", imagePath)
	processTextFileMock(ctx, client, "NewDocs", filepath.Join(root, "absent.txt"), 7)
	processImageFileMock(ctx, client, "NewImages", filepath.Join(root, "absent.png"))
}

func TestGenericMockDocumentRead(t *testing.T) {
	ctx := context.Background()
	cfg := &config.VectorDBConfig{
		Type:        config.VectorDBTypeMock,
		Enabled:     true,
		Collections: []config.Collection{{Name: "WeaveDocs", Type: "text"}},
	}
	count, err := CountDocuments(ctx, cfg, "WeaveDocs")
	if err != nil || count != 6 {
		t.Fatalf("document count = %d, %v", count, err)
	}

	output := captureUtilsOutput(t, func() {
		ShowDocument(ctx, cfg, "WeaveDocs", []string{"doc2-single"}, false, 2, nil, "", false, false, true)
	})
	var result struct {
		Count     int `json:"count"`
		Documents []struct {
			ID       string                 `json:"id"`
			Content  string                 `json:"content"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"documents"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid document JSON: %v: %q", err, output)
	}
	if result.Count != 1 || len(result.Documents) != 1 || result.Documents[0].ID != "doc2-single" || result.Documents[0].Metadata["author"] != "Test Author" {
		t.Fatalf("document JSON = %+v", result)
	}

	textOutput := captureUtilsOutput(t, func() {
		ShowDocument(ctx, cfg, "WeaveDocs", []string{"doc2-single", "missing"}, false, 2, nil, "", false, false, false)
	})
	if !strings.Contains(textOutput, "data preprocessing") || !strings.Contains(textOutput, "Collection: WeaveDocs") {
		t.Fatalf("document text output = %q", textOutput)
	}
}
