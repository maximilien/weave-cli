// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package config

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type discoveryTransport func(*http.Request) (*http.Response, error)

func (f discoveryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestOllamaDiscoveryPresentation(t *testing.T) {
	for _, mode := range []string{"models", "empty", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			previous := http.DefaultTransport
			http.DefaultTransport = discoveryTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/api/tags" {
					t.Errorf("path=%s", r.URL.Path)
				}
				body := `{"models":[{"name":"nomic-embed-text:latest"}]}`
				if mode == "empty" {
					body = `{"models":[]}`
				}
				if mode == "invalid" {
					body = "invalid"
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			defer func() { http.DefaultTransport = previous }()
			file, err := os.CreateTemp(t.TempDir(), "output")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			old := os.Stdout
			os.Stdout = file
			defer func() { os.Stdout = old }()
			discoverOllamaModels()
			if _, err = file.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "models" && !strings.Contains(string(data), "nomic-embed-text") {
				t.Fatalf("model output=%s", data)
			}
		})
	}
}
