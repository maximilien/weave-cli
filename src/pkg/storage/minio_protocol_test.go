// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMinioUploadAndReadProtocol(t *testing.T) {
	for _, mode := range []string{"existing bucket", "new bucket", "bucket denied", "create denied", "upload denied"} {
		t.Run(mode, func(t *testing.T) {
			image := []byte("fixture image bytes\x00\xff")
			type request struct {
				method, path string
				headers      http.Header
				body         []byte
			}
			requests := make(chan request, 16)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				requests <- request{r.Method, r.URL.Path, r.Header.Clone(), body}
				if r.URL.Path == "/images/" || r.URL.Path == "/images" {
					if r.Method == http.MethodHead {
						switch mode {
						case "new bucket", "create denied":
							w.WriteHeader(http.StatusNotFound)
						case "bucket denied":
							w.WriteHeader(http.StatusForbidden)
						default:
							w.WriteHeader(http.StatusOK)
						}
						return
					}
					if r.Method == http.MethodPut {
						if mode == "create denied" {
							w.WriteHeader(http.StatusForbidden)
							_, _ = io.WriteString(w, "<Error><Code>AccessDenied</Code><Message>fixture denied</Message></Error>")
							return
						}
						w.WriteHeader(http.StatusOK)
						return
					}
				}
				w.Header().Set("ETag", `"fixture-etag"`)
				w.Header().Set("Last-Modified", time.Unix(1700000000, 0).UTC().Format(http.TimeFormat))
				w.Header().Set("Content-Type", "image/png")
				switch r.Method {
				case http.MethodPut:
					if mode == "upload denied" {
						w.WriteHeader(http.StatusForbidden)
						_, _ = io.WriteString(w, "<Error><Code>AccessDenied</Code><Message>fixture denied</Message></Error>")
						return
					}
					w.WriteHeader(http.StatusOK)
				case http.MethodHead:
					w.Header().Set("Content-Length", "21")
					w.WriteHeader(http.StatusOK)
				case http.MethodDelete:
					w.WriteHeader(http.StatusNoContent)
				case http.MethodGet:
					w.Header().Set("Content-Length", "21")
					_, _ = w.Write(image)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			backend, err := NewMinioStorage(Config{Endpoint: strings.TrimPrefix(server.URL, "http://"), Bucket: "images", Region: "us-east-1", PathPrefix: "assets"})
			if err != nil {
				t.Fatal(err)
			}
			metadata := ImageMetadata{CollectionName: "docs", DocumentID: "doc-1", Filename: "photo.png", ContentType: "image/png", Custom: map[string]string{"author": "Ada"}}
			imageURL, err := backend.Upload(context.Background(), image, metadata)
			if mode == "bucket denied" || mode == "create denied" || mode == "upload denied" {
				if err == nil || imageURL != "" {
					t.Fatalf("upload=%q error=%v", imageURL, err)
				}
				if mode == "upload denied" {
					var failure *ErrUploadFailed
					if !errors.As(err, &failure) {
						t.Fatalf("upload failure type=%T", err)
					}
				} else if !strings.Contains(err.Error(), "failed to ensure bucket exists") {
					t.Fatalf("bucket failure=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(imageURL)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(imageURL, server.URL+"/images/assets/docs/") || !strings.HasSuffix(parsed.Path, ".png") {
				t.Fatalf("image URL=%s", imageURL)
			}
			var uploaded *request
			created := false
			for len(requests) > 0 {
				r := <-requests
				if r.method == http.MethodPut {
					if r.path == "/images/" || r.path == "/images" {
						created = true
					} else {
						uploaded = &r
					}
				}
			}
			if created != (mode == "new bucket") {
				t.Fatalf("bucket created=%v mode=%s", created, mode)
			}
			if uploaded == nil || uploaded.path != parsed.Path || !bytes.Equal(uploaded.body, image) {
				t.Fatalf("uploaded request=%+v", uploaded)
			}
			for key, want := range map[string]string{"Content-Type": "image/png", "X-Amz-Meta-Collection": "docs", "X-Amz-Meta-Document": "doc-1", "X-Amz-Meta-Filename": "photo.png", "X-Amz-Meta-Author": "Ada"} {
				if got := uploaded.headers.Get(key); got != want {
					t.Errorf("%s=%q want=%q", key, got, want)
				}
			}
			got, err := backend.Download(context.Background(), imageURL)
			if err != nil || !bytes.Equal(got, image) {
				t.Fatalf("download=%q error=%v", got, err)
			}
			info, infoErr := backend.GetInfo(context.Background(), imageURL)
			if infoErr != nil || info.Size != int64(len(image)) || info.ContentType != "image/png" {
				t.Fatalf("info=%+v error=%v", info, infoErr)
			}
			exists, existsErr := backend.Exists(context.Background(), imageURL)
			if existsErr != nil || !exists {
				t.Fatalf("exists=%v error=%v", exists, existsErr)
			}
			if err := backend.Delete(context.Background(), imageURL); err != nil {
				t.Fatal(err)
			}
			last := <-requests
			if last.method != http.MethodGet || last.path != parsed.Path {
				t.Fatalf("download request=%+v", last)
			}
		})
	}
}
