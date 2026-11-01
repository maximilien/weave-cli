// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWeaviateImageIngestionPreservesEXIF(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/meta" {
			_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/v1/objects" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Error(err)
		}
		_, _ = io.WriteString(w, `{"id":"created"}`)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "fixture.jpg")
	if err := os.WriteFile(path, extractedImageEXIFFixture(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := processImageFile(context.Background(), mustWeaviateClient(t, server.URL), "WeaveImages", path); err != nil {
		t.Fatal(err)
	}
	props, ok := captured["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("properties = %#v", captured)
	}
	metadata, ok := props["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata = %#v", props)
	}
	for key, want := range map[string]interface{}{"exif_make": "Canon", "exif_model": "EOS 80", "exif_orientation": float64(1), "exif_width": float64(6000), "exif_height": float64(4000), "filename": "fixture.jpg"} {
		if metadata[key] != want {
			t.Errorf("%s = %#v; want %#v", key, metadata[key], want)
		}
	}
	if captured["class"] != "WeaveImages" || props["image_data"] == "" {
		t.Fatalf("upload = %#v", captured)
	}
}
