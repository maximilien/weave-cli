// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
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

func TestWeaviateTextIngestionAndMetadataModes(t *testing.T) {
	var objects int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"WeaveDocs","properties":[{"name":"is_chunked","dataType":["boolean"]},{"name":"total_chunks","dataType":["int"]},{"name":"chunk_index","dataType":["int"]}]},{"class":"RagMeDocs","properties":[{"name":"metadata","dataType":["text"]}]}]}`)
		case "/v1/schema/WeaveDocs":
			_, _ = io.WriteString(w, `{"class":"WeaveDocs","properties":[{"name":"is_chunked","dataType":["boolean"]},{"name":"total_chunks","dataType":["int"]},{"name":"chunk_index","dataType":["int"]}]}`)
		case "/v1/schema/RagMeDocs":
			_, _ = io.WriteString(w, `{"class":"RagMeDocs","properties":[{"name":"metadata","dataType":["text"]}]}`)
		case "/v1/objects":
			objects++
			_, _ = io.WriteString(w, `{"id":"created"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "weave.txt")
	if err := os.WriteFile(path, []byte("alpha\nbeta\ngamma\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(t.TempDir(), "report.json")
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}
	if err := CreateWeaviateDocument(context.Background(), cfg, "WeaveDocs", path, 7, "", false, 0, 1, 1000, report, "json", "", 1, false); err != nil {
		t.Fatalf("CreateWeaviateDocument() = %v", err)
	}
	if objects != 3 {
		t.Fatalf("created objects = %d, want 3", objects)
	}
	if err := CreateWeaviateDocument(context.Background(), cfg, "RagMeDocs", path, 100, "", false, 0, 1, 1000, "", "", "", 1, false); err != nil {
		t.Fatalf("legacy CreateWeaviateDocument() = %v", err)
	}
	if err := CreateWeaviateDocument(context.Background(), cfg, "WeaveDocs", filepath.Join(t.TempDir(), "missing.txt"), 10, "", false, 0, 1, 1000, "", "", "", 1, false); err == nil {
		t.Fatal("missing Weaviate input file accepted")
	}
	if err := CreateWeaviateDocument(context.Background(), cfg, "WeaveDocs", path, 100, "", false, 0, 1, 1000, "", "", "", 1, true); err != nil {
		t.Fatalf("skip-existing ingestion = %v", err)
	}
	imagePath := filepath.Join(t.TempDir(), "fixture.png")
	if err := os.WriteFile(imagePath, []byte("not-a-real-image"), 0o600); err != nil {
		t.Fatal(err)
	}
	imageClient := mustWeaviateClient(t, server.URL)
	if err := processImageFile(context.Background(), imageClient, "WeaveImages", imagePath); err != nil {
		t.Fatalf("flat image processing = %v", err)
	}
	if err := processImageFile(context.Background(), imageClient, "RagMeImages", imagePath); err != nil {
		t.Fatalf("legacy image processing = %v", err)
	}
	if err := processImageFile(context.Background(), imageClient, "WeaveImages", filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("missing image accepted")
	}
	if err := processTextFile(context.Background(), imageClient, "WeaveDocs", filepath.Join(t.TempDir(), "missing.txt"), 10); err == nil {
		t.Fatal("missing text accepted")
	}
	if chunks := chunkText("", 10); len(chunks) != 1 || chunks[0] != "" {
		t.Fatalf("empty text chunks = %#v", chunks)
	}
	if data, err := os.ReadFile(report); err != nil || len(data) == 0 {
		t.Fatalf("processing report = %q, %v", data, err)
	}
	if got := getCollectionMetadataMode(context.Background(), mustWeaviateClient(t, server.URL), "Missing"); got != "flat" {
		t.Fatalf("missing schema metadata mode = %q", got)
	}
	if got := getCollectionMetadataMode(context.Background(), mustWeaviateClient(t, server.URL), "RagMeDocs"); got != "json" {
		t.Fatalf("JSON metadata mode = %q", got)
	}
}

func mustWeaviateClient(t *testing.T, endpoint string) *weaviate.Client {
	t.Helper()
	client, err := weaviate.NewClient(&weaviate.Config{URL: endpoint, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	return client
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

func TestDocumentUtilityMockMutationPaths(t *testing.T) {
	ctx := context.Background()
	cfg := &config.VectorDBConfig{
		Type: config.VectorDBTypeMock, Enabled: true,
		Collections: []config.Collection{{Name: "WeaveDocs", Type: "text"}},
	}
	path := filepath.Join(t.TempDir(), "utility.txt")
	if err := os.WriteFile(path, []byte("utility document content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CreateDocument(ctx, cfg, "WeaveDocs", path, 100, "", false, 0, 10, 2000, "", "", "", 1, false); err != nil {
		t.Fatalf("CreateDocument() = %v", err)
	}
	if err := CreateDocument(ctx, cfg, "WeaveDocs", path, 100, "", false, 0, 10, 2000, "", "", "", 1, true); err != nil {
		t.Fatalf("CreateDocument(skip existing) = %v", err)
	}
	CreateMockDocument(ctx, cfg, "WeaveDocs", path, 100, "", false, 0, 10, 2000, "", "")
	imagePath := filepath.Join(t.TempDir(), "utility.png")
	if err := os.WriteFile(imagePath, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	CreateMockDocument(ctx, cfg, "WeaveImages", imagePath, 100, "", false, 0, 10, 2000, "", "")
	CreateMockDocument(ctx, cfg, "WeaveDocs", filepath.Join(t.TempDir(), "missing.txt"), 100, "", false, 0, 10, 2000, "", "")
	if err := CreateDocument(ctx, cfg, "WeaveDocs", imagePath, 100, "", false, 0, 10, 2000, "", "", "", 1, false); err == nil {
		t.Fatal("CreateDocument accepted unsupported generic image processing")
	}
	CreateMockDocument(ctx, cfg, "WeaveDocs", filepath.Join(t.TempDir(), "invalid.pdf"), 100, "", false, 0, 10, 2000, "", "")
	vdbClient, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if exists, err := DocumentExistsByFilename(ctx, vdbClient, "WeaveDocs", "utility.txt"); err != nil {
		t.Fatalf("DocumentExistsByFilename() = %t, %v", exists, err)
	}
	DeleteMockDocuments(ctx, cfg, "WeaveDocs", []string{"doc1"}, nil, false, "", "")
	DeleteMockDocuments(ctx, cfg, "WeaveDocs", nil, []string{"filename=doc1.txt"}, false, "", "")
	DeleteMockDocuments(ctx, cfg, "WeaveDocs", nil, nil, false, "doc*", "")
	DeleteAllMockDocuments(ctx, cfg, "WeaveDocs")
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "doc1-single", "", nil, "updated", "", []string{"source=test"})
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "doc2-single", "", nil, "updated body", "", []string{"source=test", "reviewed=true"})
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "", "doc2.txt", nil, "named update", "", nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "", "", []string{"filename=doc2.txt"}, "filtered update", "", nil)
	updateFile := filepath.Join(t.TempDir(), "updated.txt")
	if err := os.WriteFile(updateFile, []byte("content from file"), 0o600); err != nil {
		t.Fatal(err)
	}
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "doc2-single", "", nil, "", updateFile, nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "", "missing.txt", nil, "", "", nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "", "", []string{"broken-filter"}, "", "", nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "missing-id", "", nil, "", "", nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "doc2-single", "", nil, "", filepath.Join(t.TempDir(), "missing.txt"), nil)
	UpdateMockDocument(ctx, cfg, "WeaveDocs", "", "", nil, "", "", []string{"invalid"})
}

func TestProcessPDFFileMockCreatesTextDocuments(t *testing.T) {
	ctx := context.Background()
	client := mock.NewClient(&config.MockConfig{Collections: []config.MockCollection{
		{Name: "WeaveDocs", Type: "text"},
		{Name: "WeaveImages", Type: "image"},
	}})
	fixture := filepath.Join("..", "..", "..", "tests", "fixtures", "ragme-io.pdf")
	processPDFFileMock(ctx, client, "WeaveDocs", fixture, 300, "WeaveImages", true, 100, 2000)

	documents, err := client.ListDocuments(ctx, "WeaveDocs", 100)
	if err != nil {
		t.Fatalf("ListDocuments() after PDF ingestion = %v", err)
	}
	if len(documents) == 0 {
		t.Fatal("PDF ingestion did not create text documents")
	}
	images, err := client.ListDocuments(ctx, "WeaveImages", 100)
	if err != nil {
		t.Fatalf("ListDocuments() for extracted images = %v", err)
	}
	if len(images) == 0 {
		t.Fatal("PDF ingestion did not create image documents")
	}
}

func TestProcessPDFFileGenericCreatesTextAndImageDocuments(t *testing.T) {
	ctx := context.Background()
	cfg := &config.VectorDBConfig{
		Type:        config.VectorDBTypeMock,
		Enabled:     true,
		Collections: []config.Collection{{Name: "WeaveDocs", Type: "text"}, {Name: "WeaveImages", Type: "image"}},
	}
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join("..", "..", "..", "tests", "fixtures", "ragme-io.pdf")
	if err := processPDFFileGeneric(ctx, cfg, client, "WeaveDocs", fixture, 300, "WeaveImages", true, 100, 2000); err != nil {
		t.Fatalf("processPDFFileGeneric() = %v", err)
	}
	textDocs, err := client.ListDocuments(ctx, "WeaveDocs", 100, 0)
	if err != nil || len(textDocs) == 0 {
		t.Fatalf("generic PDF text documents = %d, %v", len(textDocs), err)
	}
	imageDocs, err := client.ListDocuments(ctx, "WeaveImages", 100, 0)
	if err != nil || len(imageDocs) == 0 {
		t.Fatalf("generic PDF image documents = %d, %v", len(imageDocs), err)
	}
	if err := processPDFFileGeneric(ctx, cfg, client, "WeaveDocs", filepath.Join(t.TempDir(), "missing.pdf"), 300, "", true, 100, 2000); err == nil {
		t.Fatal("processPDFFileGeneric() accepted missing PDF")
	}
	if err := processPDFFileGeneric(ctx, cfg, client, "Missing", fixture, 300, "", true, 100, 2000); err == nil {
		t.Fatal("processPDFFileGeneric() accepted missing target collection")
	}
}

func TestGenericDocumentDeletionModes(t *testing.T) {
	ctx := context.Background()
	cfg := mockOperationConfig()
	// Exercise direct ID deletion, including the fallback path for a missing ID.
	DeleteDocuments(ctx, cfg, "AlphaDocs", []string{"doc1-chunk1"}, nil, false, "", "")
	DeleteDocuments(ctx, cfg, "AlphaDocs", []string{"missing-id"}, nil, false, "", "")
	// Exercise filename and metadata-filter deletion modes.
	DeleteDocuments(ctx, cfg, "AlphaDocs", nil, nil, false, "", "doc2.txt")
	DeleteDocuments(ctx, cfg, "AlphaDocs", nil, []string{"author=Test Author", "ignored"}, false, "", "")
	// No criteria should be a safe warning rather than a mutation.
	DeleteDocuments(ctx, cfg, "AlphaDocs", nil, nil, true, "unused*", "")
	DeleteMockDocuments(ctx, cfg, "BetaImages", []string{"missing-id"}, nil, false, "", "")
	DeleteMockDocuments(ctx, cfg, "BetaImages", nil, []string{"filename=missing.txt"}, false, "", "")
	DeleteMockDocuments(ctx, cfg, "BetaImages", nil, nil, false, "", "missing.txt")
	DeleteMockDocuments(ctx, cfg, "BetaImages", nil, nil, false, "*.png", "")
	DeleteMockDocuments(ctx, cfg, "BetaImages", nil, nil, false, "", "")
}

func TestGenericCollectionQueryWrapperPaths(t *testing.T) {
	ctx := context.Background()
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeMock, Enabled: true, Collections: []config.Collection{{Name: "WeaveDocs", Type: "text"}}}
	options := weaviate.QueryOptions{TopK: 2, IncludeImages: true, NoTruncate: true, JSONOutput: true}
	captureUtilsOutput(t, func() { QueryCollection(ctx, cfg, "WeaveDocs", "hello", options) })
	options.UseBM25 = true
	captureUtilsOutput(t, func() { QueryCollection(ctx, cfg, "WeaveDocs", "hello", options) })
	options.UseBM25 = false
	captureUtilsOutput(t, func() { QueryMultipleCollections(ctx, cfg, []string{"WeaveDocs"}, "hello", options) })
	captureUtilsOutput(t, func() {
		QueryMultipleCollectionsCrossVDB(ctx, []CollectionSpec{{Name: "WeaveDocs", VDBKey: "fixture"}}, map[string]*config.VectorDBConfig{"WeaveDocs": cfg}, "hello", options)
	})
}
