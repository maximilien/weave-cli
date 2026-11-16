// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package milvus

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/storage"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type imageUploadFixture struct {
	storage.ImageStorage
	data     []byte
	metadata storage.ImageMetadata
	fail     bool
}

func (s *imageUploadFixture) Upload(_ context.Context, data []byte, metadata storage.ImageMetadata) (string, error) {
	s.data = append([]byte(nil), data...)
	s.metadata = metadata
	if s.fail {
		return "", errors.New("fixture upload rejected")
	}
	return "https://storage.fixture/full-image", nil
}

func TestMilvusImageAndEmbeddingWrites(t *testing.T) {
	random := rand.New(rand.NewSource(42))
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{R: uint8(random.Intn(256)), G: uint8(random.Intn(256)), B: uint8(random.Intn(256)), A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	imageBytes := encoded.Bytes()
	if !storage.IsThumbnailNeeded(imageBytes) {
		t.Fatal("fixture image does not exercise external storage")
	}
	for _, batch := range []bool{false, true} {
		for _, mode := range []string{"success", "upload failure", "embedding failure", "invalid image"} {
			t.Run(map[bool]string{false: "single", true: "batch"}[batch]+"/"+mode, func(t *testing.T) {
				fixture := &milvusFixture{described: fixtureCollection("docs", "text-embedding-3-small", 1536)}
				adapter := newMilvusAdapter(fixture)
				upload := &imageUploadFixture{fail: mode == "upload failure"}
				adapter.imageStorage = upload
				source := imageBytes
				if mode == "invalid image" {
					source = bytes.Repeat([]byte("x"), 70000)
				}
				doc := &vectordb.Document{ID: "image-doc", Content: "Fixture document", Text: "ignored", ImageData: base64.StdEncoding.EncodeToString(source), Image: strings.Repeat("x", 3000), URL: strings.Repeat("u", 2300), Metadata: map[string]interface{}{"owner": "Ada"}}
				requests := make(chan struct{}, 2)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" {
						t.Errorf("unexpected embedding request %s %s", r.Method, r.URL.Path)
					}
					var request struct {
						Input string `json:"input"`
						Model string `json:"model"`
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
					}
					if request.Input != "Fixture document" || request.Model != "text-embedding-3-small" {
						t.Errorf("embedding request=%+v", request)
					}
					requests <- struct{}{}
					w.Header().Set("Content-Type", "application/json")
					if mode == "embedding failure" {
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"message": "fixture embedding rejected", "type": "invalid_request_error"}})
						return
					}
					vector := make([]float64, 1536)
					vector[0] = 0.25
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{map[string]interface{}{"index": 0, "embedding": vector}}})
				}))
				t.Cleanup(server.Close)
				t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
				client, err := llm.NewOpenAIClient("fixture-key")
				if err != nil {
					t.Fatal(err)
				}
				adapter.llmClient = client
				if batch {
					err = adapter.CreateDocuments(context.Background(), "docs", []*vectordb.Document{doc})
				} else {
					err = adapter.CreateDocument(context.Background(), "docs", doc)
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(requests) != 1 || len(fixture.insertedColumns) != 13 {
					t.Fatalf("embedding requests=%d inserted columns=%d", len(requests), len(fixture.insertedColumns))
				}
				columns := fixture.insertedColumns
				vector := columns[6].(*entity.ColumnFloatVector).Data()[0]
				wantFirst := float32(0.25)
				if mode == "embedding failure" {
					wantFirst = 0
				}
				if len(vector) != 1536 || vector[0] != wantFirst {
					t.Fatalf("inserted vector dimensions=%d first=%v", len(vector), vector[0])
				}
				if value := columns[3].(*entity.ColumnVarChar).Data()[0]; value != strings.Repeat("u", 2048) {
					t.Fatalf("image reference size=%d", len(value))
				}
				var inline map[string]string
				if err := json.Unmarshal(columns[4].(*entity.ColumnJSONBytes).Data()[0], &inline); err != nil {
					t.Fatal(err)
				}
				if inline["data"] != "" {
					t.Fatal("oversized image was stored inline")
				}
				if mode == "invalid image" {
					if upload.data != nil {
						t.Fatal("invalid image was uploaded")
					}
				} else if !bytes.Equal(upload.data, imageBytes) || upload.metadata.DocumentID != doc.ID || upload.metadata.CollectionName != "docs" || upload.metadata.Size != int64(len(imageBytes)) {
					t.Fatalf("upload metadata=%+v bytes=%d", upload.metadata, len(upload.data))
				}
				if mode == "success" || mode == "embedding failure" {
					if columns[11].(*entity.ColumnVarChar).Data()[0] != "https://storage.fixture/full-image" {
						t.Fatal("external image URL missing")
					}
					thumbnail := columns[10].(*entity.ColumnVarChar).Data()[0]
					data, err := base64.StdEncoding.DecodeString(thumbnail)
					if err != nil {
						t.Fatal(err)
					}
					if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
						t.Fatalf("invalid thumbnail: %v", err)
					}
				} else if columns[11].(*entity.ColumnVarChar).Data()[0] != "" {
					t.Fatal("failed upload produced an external URL")
				}
				if columns[1].(*entity.ColumnVarChar).Data()[0] != doc.Text || columns[2].(*entity.ColumnVarChar).Data()[0] != doc.Content {
					t.Fatal("document content changed")
				}
			})
		}
	}
}
