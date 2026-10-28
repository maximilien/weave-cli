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
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"gopkg.in/yaml.v3"
)

func inspectionFixture(t *testing.T, failSample bool) *config.VectorDBConfig {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/meta":
			_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","vectorizer":"none","properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}]}`)
		case "/v1/schema/Docs":
			_, _ = io.WriteString(w, `{"class":"Docs","vectorizer":"none","properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}`)
		case "/v1/graphql":
			var request struct{ Query string }
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			if strings.Contains(request.Query, "Aggregate") {
				_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":1}}]}}}`)
				return
			}
			if failSample {
				http.Error(w, "sample unavailable", 500)
				return
			}
			_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-id"},"content":"first line\nsecond line\nthird line","metadata":"{\"author\":\"Ada\",\"topic\":\"search\"}"}]}}}`)
		case "/v1/objects/Docs/doc-id":
			_, _ = io.WriteString(w, `{"id":"doc-id","class":"Docs","properties":{"content":"first line\nsecond line\nthird line","author":"Ada","topic":"search"}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return &config.VectorDBConfig{Name: "fixture", Type: config.VectorDBTypeLocal, URL: srv.URL, Timeout: 1}
}

func TestCollectionInspectionSampleAndFailure(t *testing.T) {
	for _, tc := range []struct {
		name                                    string
		generic, noTruncate, failSample, schema bool
		want                                    string
	}{
		{name: "weaviate truncated", want: "more lines"},
		{name: "weaviate full", noTruncate: true, want: "third line"},
		{name: "generic truncated", generic: true, want: "more lines"},
		{name: "generic full", generic: true, noTruncate: true, want: "third line"},
		{name: "weaviate sample failure", failSample: true, want: "Could not retrieve sample document"},
		{name: "generic sample failure", generic: true, failSample: true, want: "Could not retrieve sample document"},
		{name: "generic schema notice", generic: true, schema: true, want: "Schema display not yet implemented"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := inspectionFixture(t, tc.failSample)
			out, diagnostics := captureCollectionListing(t, func() {
				if tc.generic {
					ShowGenericCollection(context.Background(), cfg, "Docs", 1, tc.noTruncate, false, tc.schema, false, false, false, false, "", "", false)
				} else {
					ShowWeaviateCollection(context.Background(), cfg, "Docs", 1, tc.noTruncate, false, false, false, false, false, false, "", "", false)
				}
			})
			if !strings.Contains(out+diagnostics, tc.want) {
				t.Fatalf("output = %q %q; want %q", out, diagnostics, tc.want)
			}
			if !strings.Contains(out+diagnostics, "summary retrieved successfully") {
				t.Fatalf("missing summary: %q %q", out, diagnostics)
			}
		})
	}
}

func TestWeaviateInspectionExportsSchemaFiles(t *testing.T) {
	cfg := inspectionFixture(t, false)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "schema.yaml")
	jsonPath := filepath.Join(dir, "schema.json")
	_, diagnostics := captureCollectionListing(t, func() {
		ShowWeaviateCollection(context.Background(), cfg, "Docs", 3, false, false, false, true, true, true, true, yamlPath, jsonPath, false)
	})
	if !strings.Contains(diagnostics, "exported to YAML file") || !strings.Contains(diagnostics, "exported to JSON file") {
		t.Fatalf("diagnostics = %s", diagnostics)
	}
	for _, tc := range []struct {
		path   string
		decode func([]byte, interface{}) error
	}{{yamlPath, yaml.Unmarshal}, {jsonPath, json.Unmarshal}} {
		data, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]interface{}
		if err := tc.decode(data, &got); err != nil {
			t.Fatalf("decode %s: %v", tc.path, err)
		}
		if got["name"] != "Docs" || got["schema"] == nil {
			t.Fatalf("export %s = %#v", tc.path, got)
		}
	}
	for _, tc := range []struct{ name, yamlPath, jsonPath, want string }{
		{"yaml write failure", filepath.Join(dir, "missing", "schema.yaml"), "", "Failed to write YAML file"},
		{"json write failure", "", filepath.Join(dir, "missing", "schema.json"), "Failed to write JSON file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, diagnostics := captureCollectionListing(t, func() {
				ShowWeaviateCollection(context.Background(), cfg, "Docs", 3, false, false, false, false, false, false, false, tc.yamlPath, tc.jsonPath, false)
			})
			if !strings.Contains(diagnostics, tc.want) {
				t.Fatalf("diagnostics = %s", diagnostics)
			}
		})
	}
}
