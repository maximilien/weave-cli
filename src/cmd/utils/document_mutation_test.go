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
)

func TestUpdateWeaviateDocumentSelectionAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, id, filename, content, file, wantErr string
		filters, metadata                          []string
		listFailure, updateFailure                 bool
	}{
		{name: "id and metadata", id: "chosen", content: "new", metadata: []string{"status=ready=now"}},
		{name: "filename", filename: "chosen.txt", content: "new"},
		{name: "original filename", filename: "original.txt", content: "new"},
		{name: "metadata match", filters: []string{"kind=target", "revision=2"}, content: "new"},
		{name: "file overrides content", id: "chosen", content: "ignored", file: "content.txt"},
		{name: "metadata only", id: "chosen", metadata: []string{"status=ready=now"}},
		{name: "invalid metadata", id: "chosen", metadata: []string{"invalid"}, wantErr: "invalid metadata format"},
		{name: "missing file", id: "chosen", file: "missing", wantErr: "failed to read file"},
		{name: "missing name", filename: "absent", wantErr: "not found"},
		{name: "missing metadata", filters: []string{"kind=absent"}, wantErr: "no document found"},
		{name: "invalid filter", filters: []string{"invalid"}, wantErr: "invalid metadata filter"},
		{name: "name list failure", filename: "chosen.txt", listFailure: true, wantErr: "failed to list documents"},
		{name: "filter list failure", filters: []string{"kind=target"}, listFailure: true, wantErr: "failed to list documents"},
		{name: "update failure", id: "chosen", content: "new", updateFailure: true, wantErr: "failed to update document"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutations := make(chan map[string]interface{}, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/v1/meta":
					_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
				case r.URL.Path == "/v1/schema":
					_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}]}`)
				case r.URL.Path == "/v1/schema/Docs":
					_, _ = io.WriteString(w, `{"class":"Docs","properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}`)
				case r.URL.Path == "/v1/graphql":
					if tc.listFailure {
						http.Error(w, "unavailable", 500)
						return
					}
					_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"other"},"metadata":"{\"kind\":\"other\"}"},{"_additional":{"id":"chosen"},"metadata":"{\"filename\":\"chosen.txt\",\"original_filename\":\"original.txt\",\"kind\":\"target\",\"revision\":2}"}]}}}`)
				case r.Method == http.MethodGet && r.URL.Path == "/v1/objects/Docs/chosen":
					_, _ = io.WriteString(w, `{"id":"chosen","properties":{"content":"old","metadata":"{\"owner\":\"Ada\"}"}}`)
				case r.Method == http.MethodPatch && r.URL.Path == "/v1/objects/Docs/chosen":
					var body map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					mutations <- body
					if tc.updateFailure {
						http.Error(w, "update rejected", 500)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			file := ""
			if tc.file != "" {
				file = filepath.Join(t.TempDir(), tc.file)
				if tc.file != "missing" {
					if err := os.WriteFile(file, []byte("from file"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			err := UpdateWeaviateDocument(context.Background(), &config.VectorDBConfig{URL: server.URL, Timeout: 1}, "Docs", tc.id, tc.filename, tc.filters, tc.content, file, tc.metadata)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v; want %q", err, tc.wantErr)
				}
				if !tc.updateFailure && len(mutations) != 0 {
					t.Fatal("invalid selection mutated a document")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(mutations) != 1 {
				t.Fatalf("mutation count = %d", len(mutations))
			}
			body := <-mutations
			props, ok := body["properties"].(map[string]interface{})
			if !ok {
				t.Fatalf("missing properties: %#v", body)
			}
			want := tc.content
			if file != "" {
				want = "from file"
			}
			if want != "" && props["content"] != want {
				t.Fatalf("content = %#v", props)
			}
			if len(tc.metadata) > 0 {
				if props["owner"] != "Ada" || props["status"] != "ready=now" {
					t.Fatalf("merged metadata properties = %#v", props)
				}
			}
		})
	}
}

func TestDeleteWeaviateDocumentsSelection(t *testing.T) {
	for _, tc := range []struct {
		name, pattern, want, filename string
		ids                           []string
		fail                          bool
	}{
		{name: "ids", ids: []string{"chosen"}, want: "Successfully deleted 1 document(s) by ID"},
		{name: "failed id", ids: []string{"chosen"}, fail: true, want: "Failed to delete any documents"},
		{name: "pattern", pattern: "*.txt", want: "Successfully deleted 1 document(s) matching pattern"},
		{name: "no match", pattern: "*.pdf", want: "No documents found matching pattern"},
		{name: "missing filename", filename: "absent.txt", want: "No documents found with filename"},
		{name: "filename query failure", filename: "absent.txt", fail: true, want: "Failed to delete document by name"},
		{name: "no criteria", want: "No deletion criteria specified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deleted := make(chan string, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/v1/meta":
					_, _ = io.WriteString(w, `{"version":"1.27.0","modules":{}}`)
				case r.URL.Path == "/v1/schema":
					_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","properties":[{"name":"content","dataType":["text"]},{"name":"metadata","dataType":["text"]}]}]}`)
				case r.URL.Path == "/v1/schema/Docs":
					_, _ = io.WriteString(w, `{"properties":[{"name":"metadata","dataType":["text"]}]}`)
				case r.URL.Path == "/v1/graphql":
					if tc.filename != "" {
						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Error(err)
						}
						if !strings.Contains(string(body), tc.filename) {
							t.Errorf("filename filter absent: %s", body)
						}
						if tc.fail {
							http.Error(w, "query rejected", 500)
							return
						}
						_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[]}}}`)
						return
					}
					_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"other"},"metadata":"{\"filename\":\"keep.md\"}"},{"_additional":{"id":"chosen"},"metadata":"{\"filename\":\"chosen.txt\"}"}]}}}`)
				case r.Method == http.MethodDelete:
					deleted <- r.URL.Path
					if tc.fail {
						http.Error(w, "rejected", 500)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			out, diagnostics := captureCollectionListing(t, func() {
				DeleteWeaviateDocuments(context.Background(), &config.VectorDBConfig{URL: server.URL, Timeout: 1}, "Docs", tc.ids, nil, false, tc.pattern, tc.filename)
			})
			if !strings.Contains(out+diagnostics, tc.want) {
				t.Fatalf("output = %s %s; want %s", out, diagnostics, tc.want)
			}
			expected := 0
			if len(tc.ids) > 0 || tc.pattern == "*.txt" {
				expected = 1
			}
			if len(deleted) != expected {
				t.Fatalf("delete count = %d; want %d", len(deleted), expected)
			}
			if expected > 0 {
				if path := <-deleted; !strings.HasSuffix(path, "/chosen") {
					t.Fatalf("deleted wrong document: %s", path)
				}
			}
		})
	}
}
