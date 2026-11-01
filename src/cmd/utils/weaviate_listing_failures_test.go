// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestWeaviateListingEmptyAndUnavailable(t *testing.T) {
	emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/meta":
			_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[]}`)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer emptyServer.Close()
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: emptyServer.URL, Timeout: 1}
	for _, asJSON := range []bool{false, true} {
		out, diagnostics := captureCollectionListing(t, func() { ListWeaviateCollections(context.Background(), cfg, 0, false, asJSON) })
		if asJSON {
			var result struct {
				Collections []interface{} `json:"collections"`
				Total       int           `json:"total"`
			}
			if err := json.Unmarshal([]byte(out), &result); err != nil || len(result.Collections) != 0 || result.Total != 0 {
				t.Fatalf("empty JSON = %q, %v", out, err)
			}
		} else if !strings.Contains(diagnostics, "No collections found") {
			t.Fatalf("empty diagnostic: %s", diagnostics)
		}
	}
	// Reserve and close a local port to reliably exercise connection refusal.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	unavailable := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: "http://" + address, Timeout: 1}
	for _, asJSON := range []bool{false, true} {
		out, diagnostics := captureCollectionListing(t, func() { ListWeaviateCollections(context.Background(), unavailable, 0, false, asJSON) })
		if !strings.Contains(diagnostics, "Weaviate local instance not available") {
			t.Fatalf("unavailable diagnostic: %s", diagnostics)
		}
		hasTips := strings.Contains(out+diagnostics, "To start Weaviate locally")
		if hasTips == asJSON {
			t.Fatalf("asJSON=%v tips=%v output=%s %s", asJSON, hasTips, out, diagnostics)
		}
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, diagnostics := captureCollectionListing(t, func() { ListWeaviateCollections(ctx, cfg, 0, false, false) })
	if !strings.Contains(diagnostics, "connection timeout after 30 seconds") {
		t.Fatalf("timeout diagnostic: %s", diagnostics)
	}
}
