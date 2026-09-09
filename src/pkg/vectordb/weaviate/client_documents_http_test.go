// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package weaviate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestDocumentRESTLifecycleWithFakeHTTPServer(t *testing.T) {
	var (
		mu      sync.Mutex
		created []map[string]interface{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer api-key" {
			t.Errorf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/objects") {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create request: %v", err)
			}
			if body["class"] == "Broken" {
				http.Error(w, "create failed", http.StatusInternalServerError)
				return
			}
			mu.Lock()
			created = append(created, body)
			mu.Unlock()
			_, _ = io.WriteString(w, `{"id":"created"}`)
			return
		}

		prefix := "/v1/objects/Docs/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, prefix)
		if r.Method == http.MethodDelete {
			switch id {
			case "ok", "doc-1", "text-only", "empty":
				w.WriteHeader(http.StatusNoContent)
			case "also-ok":
				_, _ = io.WriteString(w, `{}`)
			case "missing":
				http.Error(w, "missing", http.StatusNotFound)
			default:
				http.Error(w, "delete failed", http.StatusBadGateway)
			}
			return
		}

		switch id {
		case "doc-1":
			_, _ = io.WriteString(w, `{
				"id":"doc-1","class":"Docs","properties":{
					"text":"text fallback","content":"document body",
					"url":"https://example.test/doc","image":"data:image/png;base64,AA",
					"image_data":"AA","metadata":"{\"author\":\"Ada\"}",
					"page":7
				}}`)
		case "text-only":
			_, _ = io.WriteString(w, `{"id":"text-only","properties":{"text":"text body","metadata":{"kind":"object"}}}`)
		case "empty":
			_, _ = io.WriteString(w, `{"id":"empty","properties":{"metadata":"not-json"}}`)
		case "malformed":
			_, _ = io.WriteString(w, `{`)
		case "missing":
			http.Error(w, "missing", http.StatusNotFound)
		default:
			http.Error(w, "read failed", http.StatusBadGateway)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(&Config{URL: server.URL, APIKey: "api-key", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	metadata := map[string]interface{}{
		"pdf_title": "Title", "pdf_creator": "Creator", "pdf_producer": "Producer",
		"pdf_creation_date": "created", "pdf_mod_date": "modified", "ai_summary": "summary",
		"chunk_sizes": []int{100}, "original_filename": "doc.pdf", "type": "pdf",
		"filename": "doc.pdf", "storage_path": "/docs/doc.pdf", "date_added": "today",
		"exif_make": "Camera", "exif_model": "Model", "exif_datetime": "timestamp",
		"exif_orientation": 1, "exif_width": 640, "exif_height": 480,
		"exif_f_number": 2.8, "exif_exposure_time": "1/100", "exif_iso": 200,
		"exif_focal_length": "50mm", "exif_gps_latitude": 1.2,
		"exif_gps_longitude": 3.4, "exif_gps_altitude": 5.6,
		"ocr_text": "recognized", "ocr_confidence": 0.9, "ocr_language": "eng",
		"ocr_word_count": 1, "ocr_text_summary": "recognized",
		"storage_path_relative": "docs/doc.pdf", "processing_timestamp": "now",
		"processing_duration_ms": 12, "source_pdf": "source.pdf",
		"pdf_filename": "source.pdf", "pdf_page": 2, "pdf_image_index": 3,
	}
	doc := Document{ID: "doc-1", Content: "body", URL: "url", Image: "image", ImageData: "raw", Metadata: metadata}
	if err := client.CreateDocument(ctx, "Docs", doc); err != nil {
		t.Fatalf("CreateDocument(text) error = %v", err)
	}
	if err := client.CreateDocument(ctx, "ProductImages", doc); err != nil {
		t.Fatalf("CreateDocument(image) error = %v", err)
	}
	if err := client.CreateDocument(ctx, "Docs", Document{Metadata: map[string]interface{}{"bad": func() {}}}); err == nil || !strings.Contains(err.Error(), "marshal metadata") {
		t.Fatalf("CreateDocument(unmarshalable) error = %v", err)
	}
	if err := client.CreateDocument(ctx, "Broken", Document{}); err == nil || !strings.Contains(err.Error(), "failed to create document") {
		t.Fatalf("CreateDocument(server error) = %v", err)
	}

	mu.Lock()
	if len(created) != 2 {
		t.Fatalf("created requests = %d", len(created))
	}
	textProperties := created[0]["properties"].(map[string]interface{})
	imageProperties := created[1]["properties"].(map[string]interface{})
	mu.Unlock()
	if _, ok := textProperties["metadata"].(string); !ok {
		t.Fatalf("text metadata type = %T", textProperties["metadata"])
	}
	if _, ok := imageProperties["metadata"].(map[string]interface{}); !ok {
		t.Fatalf("image metadata type = %T", imageProperties["metadata"])
	}
	for _, key := range []string{"Title", "Creator", "Producer", "CreationDate", "ModDate", "exif_make", "ocr_text", "source_pdf"} {
		if _, ok := textProperties[key]; !ok {
			t.Errorf("create properties missing %q", key)
		}
	}

	got, err := client.GetDocument(ctx, "Docs", "doc-1")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if got.ID != "doc-1" || got.Content != "document body" || got.Text != "text fallback" || got.Metadata["author"] != "Ada" || got.Metadata["page"] != float64(7) {
		t.Fatalf("GetDocument() = %#v", got)
	}
	textOnly, err := client.GetDocument(ctx, "Docs", "text-only")
	if err != nil || textOnly.Content != "text body" {
		t.Fatalf("GetDocument(text-only) = (%#v, %v)", textOnly, err)
	}
	empty, err := client.GetDocument(ctx, "Docs", "empty")
	if err != nil || empty.Content != "Document ID: empty" || empty.Metadata["metadata"] != "not-json" {
		t.Fatalf("GetDocument(empty) = (%#v, %v)", empty, err)
	}
	for id, want := range map[string]string{
		"missing": "not found", "malformed": "failed to parse", "failed": "HTTP 502",
	} {
		if _, err := client.GetDocument(ctx, "Docs", id); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("GetDocument(%s) error = %v, want %q", id, err, want)
		}
	}

	for _, id := range []string{"ok", "also-ok"} {
		if err := client.DeleteDocument(ctx, "Docs", id); err != nil {
			t.Errorf("DeleteDocument(%s) error = %v", id, err)
		}
	}
	for id, want := range map[string]string{"missing": "document not found", "failed": "HTTP 502"} {
		if err := client.DeleteDocument(ctx, "Docs", id); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("DeleteDocument(%s) error = %v, want %q", id, err, want)
		}
	}
	if count, err := client.DeleteDocumentsBulk(ctx, "Docs", nil); err != nil || count != 0 {
		t.Fatalf("DeleteDocumentsBulk(empty) = (%d, %v)", count, err)
	}
	if count, err := client.DeleteDocumentsBulk(ctx, "Docs", []string{"ok", "missing", "also-ok"}); err != nil || count != 2 {
		t.Fatalf("DeleteDocumentsBulk() = (%d, %v)", count, err)
	}
}

func TestCountDocumentsWithFakeGraphQLServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(body), "Docs"):
			_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":3}}]}}}`)
		case strings.Contains(string(body), "Missing"):
			_, _ = io.WriteString(w, `{"errors":[{"message":"class Missing not found"}]}`)
		case strings.Contains(string(body), "Unknown"):
			_, _ = io.WriteString(w, `{"errors":[{"message":"Unknown class Unknown"}]}`)
		case strings.Contains(string(body), "Suggested"):
			_, _ = io.WriteString(w, `{"errors":[{"message":"bad query. Did you mean Docs?"}]}`)
		case strings.Contains(string(body), "Generic"):
			_, _ = io.WriteString(w, `{"errors":[{"message":"boom"}]}`)
		default:
			_, _ = io.WriteString(w, `{"data":{}}`)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}

	if count, err := client.CountDocuments(context.Background(), "Docs"); err != nil || count != 3 {
		t.Fatalf("CountDocuments() = (%d, %v)", count, err)
	}
	for collection, want := range map[string]string{
		"Missing": "does not exist", "Unknown": "does not exist",
		"Suggested": "Did you mean Docs?", "Generic": "graphql error: boom",
		"Malformed": "failed to parse count result",
	} {
		if _, err := client.CountDocuments(context.Background(), collection); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("CountDocuments(%s) error = %v, want %q", collection, err, want)
		}
	}
}

func TestBuildMetadataQueryWithFakeSchemaServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Openai-Api-Key"); got != "openai-key" {
			t.Errorf("X-Openai-Api-Key = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		switch strings.TrimPrefix(r.URL.Path, "/v1/schema/") {
		case "Docs":
			_, _ = io.WriteString(w, `{"properties":[{"name":"metadata","dataType":["object"],"nestedProperties":[{"name":"author"},{"name":"page"}]}]}`)
		case "ProductImages":
			_, _ = io.WriteString(w, `{"properties":[{"name":"metadata","dataType":["object"],"nestedProperties":[{"name":"image_index"}]}]}`)
		case "Legacy":
			_, _ = io.WriteString(w, `{"properties":[{"name":"metadata","dataType":["text"]}]}`)
		case "Absent":
			_, _ = io.WriteString(w, `{"properties":[{"name":"content","dataType":["text"]}]}`)
		case "Malformed":
			_, _ = io.WriteString(w, `{`)
		default:
			http.Error(w, "failed", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(&Config{URL: server.URL, APIKey: "api-key", OpenAIAPIKey: "openai-key", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string][]string{
		"Docs":          {"metadata {", "author", "page"},
		"ProductImages": {"metadata {", "filename", "image_index"},
		"Legacy":        {"metadata"},
		"Absent":        {"metadata"},
		"Malformed":     {"metadata"},
		"Failed":        {"metadata"},
	}
	for collection, expected := range tests {
		query, err := client.buildMetadataQuery(context.Background(), collection)
		if err != nil {
			t.Errorf("buildMetadataQuery(%s) error = %v", collection, err)
		}
		for _, fragment := range expected {
			if !strings.Contains(query, fragment) {
				t.Errorf("buildMetadataQuery(%s) = %q, missing %q", collection, query, fragment)
			}
		}
	}
}
