// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestShowCollectionSchemaNestedAndEmptyProperties(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(fmt.Sprintf("empty=%t", empty), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/v1/meta" {
					fmt.Fprint(w, `{"version":"1.27.0","modules":{}}`)
					return
				}
				schema := `{"class":"Docs","vectorizer":"none","properties":[{"name":"metadata","dataType":["object"],"description":"Document attributes","nestedProperties":[{"name":"author","dataType":["text"]},{"name":"revision","dataType":["int"]}]}]}`
				if empty {
					schema = `{"class":"Docs","properties":[]}`
				}
				if r.URL.Path == "/v1/schema" {
					fmt.Fprintf(w, `{"classes":[%s]}`, schema)
					return
				}
				if r.URL.Path == "/v1/schema/Docs" {
					fmt.Fprint(w, schema)
					return
				}
				t.Errorf("unexpected schema request %s", r.URL.Path)
				http.NotFound(w, r)
			}))
			t.Cleanup(server.Close)
			client, err := weaviate.NewClient(&weaviate.Config{URL: server.URL, Timeout: 1})
			if err != nil {
				t.Fatal(err)
			}
			out, diag := captureCollectionListing(t, func() {
				previous := color.Output
				color.Output = os.Stdout
				defer func() { color.Output = previous }()
				ShowCollectionSchema(context.Background(), client, "Docs")
			})
			wants := []string{"Docs", "metadata", "Document attributes", "author", "text", "revision", "int"}
			if empty {
				wants = []string{"Docs", "No properties defined"}
			}
			for _, want := range wants {
				if !strings.Contains(out+diag, want) {
					t.Fatalf("missing %q in schema output: %s %s", want, out, diag)
				}
			}
			if empty && strings.Contains(out+diag, "1.1.") {
				t.Fatal("empty schema rendered nested properties")
			}
		})
	}
}
